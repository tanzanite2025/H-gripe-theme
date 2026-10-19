export interface ProductDeclarationIdentifierFields {
  merchant_product_identifier?: string
  manufacturer_product_identifier?: string
  standard_product_identifier?: string
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
  return errors
}
