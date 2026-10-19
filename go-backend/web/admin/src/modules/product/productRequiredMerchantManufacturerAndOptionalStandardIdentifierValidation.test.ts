import { describe, expect, it } from 'vitest'
import { validateProductDeclarationIdentifiers } from './productRequiredMerchantManufacturerAndOptionalStandardIdentifierValidation'

describe('product declaration identifiers', () => {
  it('reports both required fields and accepts an absent optional code', () => {
    expect(validateProductDeclarationIdentifiers({ merchant_product_identifier: ' \t', manufacturer_product_identifier: '\n' })).toEqual({
      merchant_product_identifier: '请填写商家产品识别码（MPID）',
      manufacturer_product_identifier: '请填写制造商产品识别码（NSPID）',
    })
    expect(validateProductDeclarationIdentifiers({ merchant_product_identifier: ' SHOP-1 ', manufacturer_product_identifier: ' FACTORY-2 ' })).toEqual({})
  })

  it('enforces a 70-character limit on all three fields including optional values', () => {
    const valid = { merchant_product_identifier: '商'.repeat(70), manufacturer_product_identifier: 'M'.repeat(70), standard_product_identifier: '0'.repeat(70) }
    expect(validateProductDeclarationIdentifiers(valid)).toEqual({})
    for (const field of Object.keys(valid)) {
      expect(validateProductDeclarationIdentifiers({ ...valid, [field]: '字'.repeat(71) })[field]).toContain('70 个字符')
    }
  })
})
