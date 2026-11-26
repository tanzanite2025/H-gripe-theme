import { describe, expect, it } from 'vitest'
import {
  createEmptySpokesProductSpecificationForm,
  createEmptySpokeGaugeOptionForm,
  createSpokesProductSpecificationFormFromProduct,
  normalizeSpokesProductSpecificationForSubmission,
  validateSpokesDedicatedProductSubmission,
} from './spokesDedicatedProductSpecificationFormValuesAndValidation'

describe('dedicated spoke product specification form', () => {
  it('keeps optional specifications empty while requiring the model name', () => {
    const specification = createEmptySpokesProductSpecificationForm()
    specification.model_name = 'CX-Ray'
    const payload = { spoke_specification: normalizeSpokesProductSpecificationForSubmission(specification) }
    expect(payload.spoke_specification).toMatchObject({ schema_version: 1, model_name: 'CX-Ray' })
    expect(payload.spoke_specification.gauge_options).toEqual([])
    expect(validateSpokesDedicatedProductSubmission(payload)).toEqual({})
  })

  it('round trips gauge dimensions, reference weights and SAPIM middle strength units', () => {
    const product = createSpokesProductSpecificationFormFromProduct({
      model_name: 'CX-Ray',
      gauge_options: [{ gauge: '14G', hub_end_diameter_mm: 2, middle_width_mm: 2.2, middle_thickness_mm: 0.9, length_min_mm: 141, length_max_mm: 310, weight_reference: { spoke_count: 64, length_mm: 260, total_weight_g: 279 } }],
      middle_strength_n_per_mm2: { minimum_n_per_mm2: 1600, maximum_n_per_mm2: 1600 },
      color_options: ['Black oxidized', 'Silver'],
      source_urls: ['https://sapim.eu/spokes/'],
    })
    const normalized = normalizeSpokesProductSpecificationForSubmission(product)
    expect(normalized.gauge_options[0]).toMatchObject({ gauge: '14G', middle_thickness_mm: 0.9, weight_reference: { spoke_count: 64 } })
    expect(normalized.middle_strength_n_per_mm2).toEqual({ minimum_n_per_mm2: 1600, maximum_n_per_mm2: 1600 })
    expect(validateSpokesDedicatedProductSubmission({ spoke_specification: normalized })).toEqual({})
  })

  it('rejects invalid length ranges, partial weight references and non-HTTP sources', () => {
    const specification = createEmptySpokesProductSpecificationForm()
    specification.model_name = 'Race'
    specification.gauge_options = [createEmptySpokeGaugeOptionForm()]
    specification.gauge_options[0].gauge = '14G'
    specification.gauge_options[0].length_min_mm = 300
    specification.gauge_options[0].length_max_mm = 200
    specification.gauge_options[0].weight_reference.spoke_count = 64
    specification.source_urls = 'javascript:alert(1)'
    const errors = validateSpokesDedicatedProductSubmission({ spoke_specification: normalizeSpokesProductSpecificationForSubmission(specification) })
    expect(errors.spoke_specification).toContain('来源')
  })
})
