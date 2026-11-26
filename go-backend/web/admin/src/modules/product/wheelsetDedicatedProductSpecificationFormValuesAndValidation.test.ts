import { describe, expect, it } from 'vitest'
import {
  createEmptyWheelsetSpecificationForm,
  createWheelsetSpecificationFormFromProduct,
  normalizeWheelsetSpecificationForSubmission,
  validateWheelsetDedicatedProductSubmission,
} from './wheelsetDedicatedProductSpecificationFormValuesAndValidation'

describe('WHEELSET dedicated product specification form values', () => {
  it('keeps front and rear product facts independent and omits empty optional fields', () => {
    const form = createEmptyWheelsetSpecificationForm()
    form.front_wheel.wheel_size = '700C'
    form.front_wheel.spoke_hole_counts = '24, 28'
    form.front_wheel.lacing_patterns = '2x, 3x'
    form.rear_wheel.wheel_size = '650B'
    form.rear_wheel.spoke_hole_counts = '28, 32'
    form.rear_wheel.lacing_patterns = '3x'

    expect(normalizeWheelsetSpecificationForSubmission(form)).toEqual({
      schema_version: 1,
      front_wheel: {
        wheel_size: '700C',
        spoke_hole_counts: [24, 28],
        lacing_patterns: ['2x', '3x'],
      },
      rear_wheel: {
        wheel_size: '650B',
        spoke_hole_counts: [28, 32],
        lacing_patterns: ['3x'],
      },
      certifications: [],
    })
  })

  it('hydrates arrays for editing and preserves entered optional measurements', () => {
    const form = createWheelsetSpecificationFormFromProduct({
      front_wheel: { wheel_size: '700C', spoke_hole_counts: [24, 28], lacing_patterns: ['2x'], rim_depth_mm: 45 },
      rear_wheel: { wheel_size: '650B', spoke_hole_counts: [28], lacing_patterns: ['3x'] },
      max_system_weight_kg: 120,
      certifications: ['ENVE tested', 'UCI approved'],
    })

    expect(form.front_wheel.spoke_hole_counts).toBe('24, 28')
    expect(form.rear_wheel.spoke_hole_counts).toBe('28')
    expect(normalizeWheelsetSpecificationForSubmission(form).front_wheel).toMatchObject({ rim_depth_mm: 45 })
    expect(normalizeWheelsetSpecificationForSubmission(form).max_system_weight_kg).toBe(120)
    expect(normalizeWheelsetSpecificationForSubmission(form).certifications).toEqual(['ENVE tested', 'UCI approved'])
  })

  it('requires both wheels sizes, spoke hole counts, and lacing patterns', () => {
    const payload = { wheelset_specification: normalizeWheelsetSpecificationForSubmission(createEmptyWheelsetSpecificationForm()) }
    expect(validateWheelsetDedicatedProductSubmission(payload).wheelset_specification).toContain('前轮必须填写轮径')

    const form = createEmptyWheelsetSpecificationForm()
    form.front_wheel.wheel_size = '700C'
    form.front_wheel.spoke_hole_counts = '24'
    form.front_wheel.lacing_patterns = '2x'
    form.rear_wheel.wheel_size = '700C'
    form.rear_wheel.spoke_hole_counts = '28'
    form.rear_wheel.lacing_patterns = '3x'
    expect(validateWheelsetDedicatedProductSubmission({ wheelset_specification: normalizeWheelsetSpecificationForSubmission(form) })).toEqual({})
  })

  it('rejects invalid counts, duplicate supported values, and impossible width relationships', () => {
    const form = createEmptyWheelsetSpecificationForm()
    form.front_wheel.wheel_size = '700C'
    form.front_wheel.spoke_hole_counts = '24, 24'
    form.front_wheel.lacing_patterns = '2x'
    form.rear_wheel.wheel_size = '700C'
    form.rear_wheel.spoke_hole_counts = '28'
    form.rear_wheel.lacing_patterns = '3x'

    const duplicateValueError = validateWheelsetDedicatedProductSubmission({ wheelset_specification: normalizeWheelsetSpecificationForSubmission(form) }).wheelset_specification
    expect(duplicateValueError).toContain('不能重复')

    form.front_wheel.spoke_hole_counts = '24'
    form.rear_wheel.spoke_hole_counts = '0'
    const invalidCountError = validateWheelsetDedicatedProductSubmission({ wheelset_specification: normalizeWheelsetSpecificationForSubmission(form) }).wheelset_specification
    expect(invalidCountError).toContain('孔数必须是 1 到 128 之间的整数')

    form.rear_wheel.spoke_hole_counts = '28'
    form.front_wheel.rim_internal_width_mm = 25
    form.front_wheel.rim_external_width_mm = 20
    const invalidWidthError = validateWheelsetDedicatedProductSubmission({ wheelset_specification: normalizeWheelsetSpecificationForSubmission(form) }).wheelset_specification
    expect(invalidWidthError).toContain('外宽必须大于内宽')
  })
})
