import { describe, expect, it } from 'vitest'
import {
  createBicycleTireProductFormValuesFromSchwalbeCatalogItem,
  createEmptyBicycleTireSpecificationForm,
  normalizeBicycleTireSpecificationForSubmission,
  validateBicycleTireDedicatedProductSubmission,
} from './bicycleTireDedicatedSpecificationFormValuesAndValidation'

describe('bicycle tire dedicated product optional specifications', () => {
  it('submits an empty optional specification document without zero or empty values', () => {
    const payload = {
      brand_id: null,
      bicycle_tire_specification: normalizeBicycleTireSpecificationForSubmission(createEmptyBicycleTireSpecificationForm()),
    }

    expect(payload.bicycle_tire_specification).toEqual({ schema_version: 1 })
    expect(validateBicycleTireDedicatedProductSubmission(payload)).toEqual({})
  })

  it('maps Schwalbe catalog values into the dedicated tire specification and product EAN field', () => {
    const values = createBicycleTireProductFormValuesFromSchwalbeCatalogItem({
      article_no: '11100568',
      ean: '4026495667913',
      model_name: 'Fat Frank',
      etrto: '50-622',
      weight_g: 860,
      source_checked_at: '2026-11-24',
      product_exists: false,
    })

    expect(values.bicycle_tire_specification).toMatchObject({
      schema_version: 1,
      article_no: '11100568',
      model_name: 'Fat Frank',
      etrto: '50-622',
      weight_g: 860,
      inch_designation: '',
      max_pressure_bar: '',
    })
    expect(values.standard_product_identifier).toBe('4026495667913')
  })

  it('rejects an Article No. without a brand', () => {
    const specification = createEmptyBicycleTireSpecificationForm()
    specification.article_no = 'TIRE-ARTICLE-001'

    const errors = validateBicycleTireDedicatedProductSubmission({
      brand_id: null,
      bicycle_tire_specification: normalizeBicycleTireSpecificationForSubmission(specification),
    })

    expect(errors.bicycle_tire_specification).toContain('填写 Article No. 时必须选择品牌')
  })

  it('rejects reversed pressure bounds', () => {
    const specification = createEmptyBicycleTireSpecificationForm()
    specification.min_pressure_bar = 5
    specification.max_pressure_bar = 3

    const errors = validateBicycleTireDedicatedProductSubmission({
      brand_id: null,
      bicycle_tire_specification: normalizeBicycleTireSpecificationForSubmission(specification),
    })

    expect(errors.bicycle_tire_specification).toContain('最小 Bar 胎压不能大于最大 Bar 胎压')
  })

  it('matches backend UTF-8 byte limits for Chinese specification text', () => {
    const specification = createEmptyBicycleTireSpecificationForm()
    specification.model_name = '车'.repeat(61)

    const errors = validateBicycleTireDedicatedProductSubmission({
      brand_id: null,
      bicycle_tire_specification: normalizeBicycleTireSpecificationForSubmission(specification),
    })

    expect(errors.bicycle_tire_specification).toContain('model_name 不能超过 180 个字节')
  })
})
