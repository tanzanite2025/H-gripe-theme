import { describe, expect, it } from 'vitest'
import { createEmptyCarbonRimSpecificationForm, createEmptyCarbonRimVariantSpecificationForm, normalizeCarbonRimSpecificationForSubmission, normalizeCarbonRimVariantSpecificationForSubmission, validateCarbonRimDedicatedProductSubmission } from './carbonRimDedicatedSpecificationFormValuesAndValidation'

function createValidCarbonRimSubmission() {
  return {
    carbon_rim_specification: normalizeCarbonRimSpecificationForSubmission({ ...createEmptyCarbonRimSpecificationForm(), model: 'R45', wheel_size: '700c', riding_uses: ['road', 'gravel'], bead_seat_diameter_mm: '622', depth_mm: '45', internal_width_mm: '25', external_width_mm: '32', bead_structure: 'hookless', tubeless_ready: false }),
    variants: [{ price_minor: 19900, stock: 4, weight_grams: 450, carbon_rim_specification: normalizeCarbonRimVariantSpecificationForSubmission({ ...createEmptyCarbonRimVariantSpecificationForm(), spoke_hole_count: '24', drilling_pattern: 'uniform', finish_code: 'matte-black', net_weight_grams: '430' }) }],
  }
}

describe('dedicated carbon rim submission boundaries', () => {
  it('requires explicit riding uses and rejects a nominal wheel size that conflicts with BSD', () => {
    const payload=createValidCarbonRimSubmission()
    payload.carbon_rim_specification.wheel_size='650b'
    expect(validateCarbonRimDedicatedProductSubmission(payload).carbon_rim_specification).toContain('轮径')
    payload.carbon_rim_specification.wheel_size='29-inch'
    expect(validateCarbonRimDedicatedProductSubmission(payload)).toEqual({})
    payload.carbon_rim_specification.riding_uses=[]
    expect(validateCarbonRimDedicatedProductSubmission(payload).carbon_rim_specification).toContain('骑行场景')
    expect(payload.carbon_rim_specification.riding_uses).toEqual([])
  })
  it('rejects impossible cross-section geometry and nonpositive optional valve dimensions before sending', () => {
    const payload = createValidCarbonRimSubmission()
    payload.carbon_rim_specification.asymmetric_offset_mm = 16
    expect(validateCarbonRimDedicatedProductSubmission(payload).carbon_rim_specification).toContain('偏心量')
    payload.carbon_rim_specification.asymmetric_offset_mm = 0
    payload.carbon_rim_specification.depth_mm = 311
    expect(validateCarbonRimDedicatedProductSubmission(payload).carbon_rim_specification).toContain('框高')
    payload.carbon_rim_specification.depth_mm = 45
    payload.carbon_rim_specification.valve_hole_diameter_mm = 0
    expect(validateCarbonRimDedicatedProductSubmission(payload).carbon_rim_specification).toContain('气嘴孔')
  })
  it('preserves explicitly false tubeless and unknown optional facts without inventing zero measurements', () => {
    const payload = createValidCarbonRimSubmission()
    expect(payload.carbon_rim_specification.tubeless_ready).toBe(false)
    expect(payload.carbon_rim_specification.rim_tape_required).toBeNull()
    expect(payload.carbon_rim_specification.valve_hole_diameter_mm).toBeNull()
    expect(payload.variants[0].carbon_rim_specification.erd_mm).toBeNull()
    expect(validateCarbonRimDedicatedProductSubmission(payload)).toEqual({})
    payload.carbon_rim_specification.tubeless_ready = null
    expect(validateCarbonRimDedicatedProductSubmission(payload).carbon_rim_specification).toBeTruthy()
  })
  it('rejects duplicate actual configurations and requires a convention with ERD', () => {
    const payload = createValidCarbonRimSubmission()
    payload.variants.push(structuredClone(payload.variants[0]))
    expect(validateCarbonRimDedicatedProductSubmission(payload).variants).toContain('组合重复')
    payload.variants[1].carbon_rim_specification.spoke_hole_count = 28
    payload.variants[1].carbon_rim_specification.erd_mm = 532
    expect(validateCarbonRimDedicatedProductSubmission(payload).variants).toContain('测量约定')
    payload.variants[1].carbon_rim_specification.erd_measurement_notes = 'Measured at nipple seats'
    expect(validateCarbonRimDedicatedProductSubmission(payload)).toEqual({})
  })
})
