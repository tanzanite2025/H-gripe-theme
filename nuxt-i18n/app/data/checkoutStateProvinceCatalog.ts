import {
  AF,
  AE,
  AL,
  AM,
  AR,
  AT,
  AU,
  AZ,
  BD,
  BE,
  BG,
  BH,
  BR,
  BY,
  CA,
  CH,
  CL,
  CN,
  CO,
  CY,
  CZ,
  DE,
  DK,
  DZ,
  EE,
  EG,
  ES,
  FI,
  FR,
  GB,
  GE,
  GR,
  HK,
  HR,
  HU,
  ID,
  IE,
  IL,
  IN,
  IQ,
  IR,
  IS,
  IT,
  JO,
  JP,
  KE,
  KH,
  KR,
  KW,
  KZ,
  LB,
  LK,
  LT,
  LU,
  LV,
  MA,
  MO,
  MX,
  MY,
  NG,
  NL,
  NO,
  NZ,
  OM,
  PA,
  PE,
  PH,
  PK,
  PL,
  PT,
  QA,
  RO,
  RS,
  RU,
  SA,
  SE,
  SG,
  SI,
  SK,
  TH,
  TR,
  TW,
  UA,
  US,
  VN,
  ZA,
} from 'country-region-data'
import { COUNTRIES } from '~/data/countries'

// Checkout and tax rules use the same country-specific region codes.
export interface CheckoutStateProvinceOption {
  code: string
  name: string
}

const checkoutStateProvinceOptionsByCountryCode = new Map<string, CheckoutStateProvinceOption[]>()
const supportedCheckoutCountryCodes = new Set(COUNTRIES.map(country => country.code.toUpperCase()))
const supportedCheckoutCountryRegionData = [
  AF, AL, DZ, AR, AM, AU, AT, AZ, BH, BD, BY, BE, BR, BG, KH, CA, CL, CN, CO,
  HR, CY, CZ, DK, EG, EE, FI, FR, GE, DE, GR, HK, HU, IS, IN, ID, IR, IQ, IE,
  IL, IT, JP, JO, KZ, KE, KR, KW, LV, LB, LT, LU, MO, MY, MX, MA, NL, NZ, NG,
  NO, OM, PK, PA, PE, PH, PL, PT, QA, RO, RU, SA, RS, SG, SK, SI, ZA, ES, LK,
  SE, CH, TW, TH, TR, UA, AE, GB, US, VN,
]

for (const [, countryCode, regions] of supportedCheckoutCountryRegionData) {
  if (!supportedCheckoutCountryCodes.has(countryCode)) continue
  checkoutStateProvinceOptionsByCountryCode.set(
    countryCode.toUpperCase(),
    regions
      .filter(([, regionCode]) => regionCode.trim() !== '')
      .map(([name, code]) => ({ code, name })),
  )
}

const normalizeCheckoutStateProvinceValueForComparison = (value: string): string => (
  value
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/[^a-z\d]/gi, '')
    .toUpperCase()
)

export const getCheckoutStateProvinceOptions = (countryCode: string): CheckoutStateProvinceOption[] => (
  checkoutStateProvinceOptionsByCountryCode.get(String(countryCode || '').trim().toUpperCase()) || []
)

export const normalizeCheckoutStateProvinceCode = (countryCode: string, stateProvinceValue: string): string => {
  const originalValue = String(stateProvinceValue || '').trim()
  if (!originalValue) return ''

  const options = getCheckoutStateProvinceOptions(countryCode)
  if (!options.length) return ''

  const comparableValue = normalizeCheckoutStateProvinceValueForComparison(originalValue)
  const matchingOption = options.find(option => (
    normalizeCheckoutStateProvinceValueForComparison(option.code) === comparableValue ||
    normalizeCheckoutStateProvinceValueForComparison(option.name) === comparableValue
  ))

  return matchingOption?.code || ''
}
