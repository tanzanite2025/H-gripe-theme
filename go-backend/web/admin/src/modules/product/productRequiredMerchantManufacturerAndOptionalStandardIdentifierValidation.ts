export interface ProductDeclarationIdentifierFields {
  merchant_product_identifier?: string
  manufacturer_product_identifier?: string
  standard_product_identifier?: string
  product_type?: string
}

export function isValidGS1GlobalTradeItemNumber(value: string): boolean {
  const normalized = String(value || '').trim()
  if (!/^(?:\d{8}|\d{12}|\d{13}|\d{14})$/.test(normalized)) return false

  let sum = 0
  let weight = 3
  for (let index = normalized.length - 2; index >= 0; index -= 1) {
    sum += Number(normalized[index]) * weight
    weight = weight === 3 ? 1 : 3
  }

  return Number(normalized[normalized.length - 1]) === (10 - (sum % 10)) % 10
}

export function validateProductDeclarationIdentifiers(fields: ProductDeclarationIdentifierFields): Record<string, string> {
  const errors: Record<string, string> = {}
  const definitions = [
    { key: 'merchant_product_identifier', label: '商家产品识别码（MPID）', required: true },
    { key: 'manufacturer_product_identifier', label: '制造商产品识别码（NSPID）', required: true },
    { key: 'standard_product_identifier', label: '国际标准商品编码（SPID）', required: false },
  ] as const
  for (const definition of definitions) {
    const value = (fields[definition.key] || '').trim()
    if (definition.required && !value) errors[definition.key] = `请填写${definition.label}`
    else if (Array.from(value).length > 70) errors[definition.key] = `${definition.label}不能超过 70 个字符`
  }
  const standardProductIdentifier = (fields.standard_product_identifier || '').trim()
  if (fields.product_type === 'wheelset' && standardProductIdentifier && !isValidGS1GlobalTradeItemNumber(standardProductIdentifier)) {
    errors.standard_product_identifier = 'EAN/GTIN 必须是校验位正确的 8、12、13 或 14 位数字'
  }
  return errors
}
