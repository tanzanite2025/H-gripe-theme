import { countryTuples } from 'country-region-data'

export interface AddressRegionOption {
  code: string
  name: string
  englishName: string
  keywords: string
}

const normalizeRegionCode = (value: unknown) => String(value || '').trim().toUpperCase()

const regionDisplayName = (code: string, locale: string) => {
  try {
    const displayNames = new Intl.DisplayNames([locale], { type: 'region' })
    return displayNames.of(code) || code
  } catch {
    return code
  }
}

const supportedRegionCodes = countryTuples
  .map(([, code]) => normalizeRegionCode(code))
  .filter((code) => /^[A-Z]{2}$/.test(code))

const supportedRegionCodeSet = new Set(supportedRegionCodes)

export const addressRegionOptions: AddressRegionOption[] = countryTuples
  .map(([englishCountryName, countryCode]) => {
    const code = normalizeRegionCode(countryCode)
    const name = regionDisplayName(code, 'zh-CN')
    const englishName = englishCountryName || regionDisplayName(code, 'en')

    return {
      code,
      name,
      englishName,
      keywords: `${code} ${name} ${englishName}`.toLowerCase(),
    }
  })
  .sort((left, right) => left.name.localeCompare(right.name, 'zh-CN'))

const addressRegionByCode = new Map(addressRegionOptions.map((region) => [region.code, region]))

export const parseAddressRegionCodes = (value: unknown): string[] => {
  const rawItems = (() => {
    if (Array.isArray(value)) return value

    const text = String(value || '').trim()
    if (!text) return []

    try {
      const parsed = JSON.parse(text)
      if (Array.isArray(parsed)) return parsed
    } catch {
      // Keep legacy comma/newline input readable.
    }

    return text.split(/[\s,;|]+/g)
  })()

  const seen = new Set<string>()
  const codes: string[] = []

  for (const item of rawItems) {
    const code = normalizeRegionCode(item)
    if (!supportedRegionCodeSet.has(code) || seen.has(code)) continue
    seen.add(code)
    codes.push(code)
  }

  return codes
}

export const serializeAddressRegionCodes = (codes: unknown[]) => {
  return JSON.stringify(parseAddressRegionCodes(codes))
}

export const addressRegionName = (code: unknown) => {
  const normalized = normalizeRegionCode(code)
  return addressRegionByCode.get(normalized)?.name || normalized
}

export const addressRegionSummary = (value: unknown, limit = 4) => {
  const codes = parseAddressRegionCodes(value)
  if (!codes.length) return '-'

  const visible = codes.slice(0, limit).map((code) => addressRegionName(code))
  const remaining = codes.length - visible.length
  return remaining > 0 ? `${visible.join('、')} +${remaining}` : visible.join('、')
}
