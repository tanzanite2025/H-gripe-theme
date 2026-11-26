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

  it('validates optional dedicated-product GTIN length and GS1 check digit while preserving leading zeros', () => {
    for (const gtin of ['96385074', '036000291452', '4006381333931', '00012345678905']) {
      expect(validateProductDeclarationIdentifiers({
        product_type: 'wheelset',
        merchant_product_identifier: 'SHOP-WHEELSET-1',
        manufacturer_product_identifier: 'FACTORY-WHEELSET-1',
        standard_product_identifier: gtin,
      })).toEqual({})
    }

    expect(validateProductDeclarationIdentifiers({
      product_type: 'wheelset',
      merchant_product_identifier: 'SHOP-WHEELSET-1',
      manufacturer_product_identifier: 'FACTORY-WHEELSET-1',
      standard_product_identifier: '4006381333932',
    }).standard_product_identifier).toContain('校验位正确')
    expect(validateProductDeclarationIdentifiers({
      product_type: 'wheelset',
      merchant_product_identifier: 'SHOP-WHEELSET-1',
      manufacturer_product_identifier: 'FACTORY-WHEELSET-1',
      standard_product_identifier: '400-6381333931',
    }).standard_product_identifier).toContain('8、12、13 或 14 位数字')
    expect(validateProductDeclarationIdentifiers({
      product_type: 'spoke',
      merchant_product_identifier: 'SHOP-SPOKE-1',
      manufacturer_product_identifier: 'FACTORY-SPOKE-1',
      standard_product_identifier: '4006381333931',
    })).toEqual({})
    expect(validateProductDeclarationIdentifiers({
      product_type: 'spoke',
      merchant_product_identifier: 'SHOP-SPOKE-1',
      manufacturer_product_identifier: 'FACTORY-SPOKE-1',
      standard_product_identifier: '4006381333932',
    }).standard_product_identifier).toContain('校验位正确')
    expect(validateProductDeclarationIdentifiers({
      product_type: 'wheelset',
      merchant_product_identifier: 'SHOP-WHEELSET-1',
      manufacturer_product_identifier: 'FACTORY-WHEELSET-1',
    })).toEqual({})
  })
})
