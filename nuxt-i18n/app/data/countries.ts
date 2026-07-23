/**
 * 国家列表数据
 * ISO 3166-1 alpha-2 国家代码
 */

export interface Country {
  code: string
  name: string
  nameZh?: string
}

/**
 * 常用国家列表（按字母排序）
 * 包含英文名和中文名
 */
export const COUNTRIES: Country[] = [
  { code: 'AF', name: 'Afghanistan', nameZh: '阿富汗' },
  { code: 'AL', name: 'Albania', nameZh: '阿尔巴尼亚' },
  { code: 'DZ', name: 'Algeria', nameZh: '阿尔及利亚' },
  { code: 'AR', name: 'Argentina', nameZh: '阿根廷' },
  { code: 'AM', name: 'Armenia', nameZh: '亚美尼亚' },
  { code: 'AU', name: 'Australia', nameZh: '澳大利亚' },
  { code: 'AT', name: 'Austria', nameZh: '奥地利' },
  { code: 'AZ', name: 'Azerbaijan', nameZh: '阿塞拜疆' },
  { code: 'BH', name: 'Bahrain', nameZh: '巴林' },
  { code: 'BD', name: 'Bangladesh', nameZh: '孟加拉国' },
  { code: 'BY', name: 'Belarus', nameZh: '白俄罗斯' },
  { code: 'BE', name: 'Belgium', nameZh: '比利时' },
  { code: 'BR', name: 'Brazil', nameZh: '巴西' },
  { code: 'BG', name: 'Bulgaria', nameZh: '保加利亚' },
  { code: 'KH', name: 'Cambodia', nameZh: '柬埔寨' },
  { code: 'CA', name: 'Canada', nameZh: '加拿大' },
  { code: 'CL', name: 'Chile', nameZh: '智利' },
  { code: 'CN', name: 'China', nameZh: '中国' },
  { code: 'CO', name: 'Colombia', nameZh: '哥伦比亚' },
  { code: 'HR', name: 'Croatia', nameZh: '克罗地亚' },
  { code: 'CY', name: 'Cyprus', nameZh: '塞浦路斯' },
  { code: 'CZ', name: 'Czech Republic', nameZh: '捷克' },
  { code: 'DK', name: 'Denmark', nameZh: '丹麦' },
  { code: 'EG', name: 'Egypt', nameZh: '埃及' },
  { code: 'EE', name: 'Estonia', nameZh: '爱沙尼亚' },
  { code: 'FI', name: 'Finland', nameZh: '芬兰' },
  { code: 'FR', name: 'France', nameZh: '法国' },
  { code: 'GE', name: 'Georgia', nameZh: '格鲁吉亚' },
  { code: 'DE', name: 'Germany', nameZh: '德国' },
  { code: 'GR', name: 'Greece', nameZh: '希腊' },
  { code: 'HK', name: 'Hong Kong', nameZh: '香港' },
  { code: 'HU', name: 'Hungary', nameZh: '匈牙利' },
  { code: 'IS', name: 'Iceland', nameZh: '冰岛' },
  { code: 'IN', name: 'India', nameZh: '印度' },
  { code: 'ID', name: 'Indonesia', nameZh: '印度尼西亚' },
  { code: 'IR', name: 'Iran', nameZh: '伊朗' },
  { code: 'IQ', name: 'Iraq', nameZh: '伊拉克' },
  { code: 'IE', name: 'Ireland', nameZh: '爱尔兰' },
  { code: 'IL', name: 'Israel', nameZh: '以色列' },
  { code: 'IT', name: 'Italy', nameZh: '意大利' },
  { code: 'JP', name: 'Japan', nameZh: '日本' },
  { code: 'JO', name: 'Jordan', nameZh: '约旦' },
  { code: 'KZ', name: 'Kazakhstan', nameZh: '哈萨克斯坦' },
  { code: 'KE', name: 'Kenya', nameZh: '肯尼亚' },
  { code: 'KR', name: 'South Korea', nameZh: '韩国' },
  { code: 'KW', name: 'Kuwait', nameZh: '科威特' },
  { code: 'LV', name: 'Latvia', nameZh: '拉脱维亚' },
  { code: 'LB', name: 'Lebanon', nameZh: '黎巴嫩' },
  { code: 'LT', name: 'Lithuania', nameZh: '立陶宛' },
  { code: 'LU', name: 'Luxembourg', nameZh: '卢森堡' },
  { code: 'MO', name: 'Macau', nameZh: '澳门' },
  { code: 'MY', name: 'Malaysia', nameZh: '马来西亚' },
  { code: 'MX', name: 'Mexico', nameZh: '墨西哥' },
  { code: 'MA', name: 'Morocco', nameZh: '摩洛哥' },
  { code: 'NL', name: 'Netherlands', nameZh: '荷兰' },
  { code: 'NZ', name: 'New Zealand', nameZh: '新西兰' },
  { code: 'NG', name: 'Nigeria', nameZh: '尼日利亚' },
  { code: 'NO', name: 'Norway', nameZh: '挪威' },
  { code: 'OM', name: 'Oman', nameZh: '阿曼' },
  { code: 'PK', name: 'Pakistan', nameZh: '巴基斯坦' },
  { code: 'PA', name: 'Panama', nameZh: '巴拿马' },
  { code: 'PE', name: 'Peru', nameZh: '秘鲁' },
  { code: 'PH', name: 'Philippines', nameZh: '菲律宾' },
  { code: 'PL', name: 'Poland', nameZh: '波兰' },
  { code: 'PT', name: 'Portugal', nameZh: '葡萄牙' },
  { code: 'QA', name: 'Qatar', nameZh: '卡塔尔' },
  { code: 'RO', name: 'Romania', nameZh: '罗马尼亚' },
  { code: 'RU', name: 'Russia', nameZh: '俄罗斯' },
  { code: 'SA', name: 'Saudi Arabia', nameZh: '沙特阿拉伯' },
  { code: 'RS', name: 'Serbia', nameZh: '塞尔维亚' },
  { code: 'SG', name: 'Singapore', nameZh: '新加坡' },
  { code: 'SK', name: 'Slovakia', nameZh: '斯洛伐克' },
  { code: 'SI', name: 'Slovenia', nameZh: '斯洛文尼亚' },
  { code: 'ZA', name: 'South Africa', nameZh: '南非' },
  { code: 'ES', name: 'Spain', nameZh: '西班牙' },
  { code: 'LK', name: 'Sri Lanka', nameZh: '斯里兰卡' },
  { code: 'SE', name: 'Sweden', nameZh: '瑞典' },
  { code: 'CH', name: 'Switzerland', nameZh: '瑞士' },
  { code: 'TW', name: 'Taiwan', nameZh: '台湾' },
  { code: 'TH', name: 'Thailand', nameZh: '泰国' },
  { code: 'TR', name: 'Turkey', nameZh: '土耳其' },
  { code: 'UA', name: 'Ukraine', nameZh: '乌克兰' },
  { code: 'AE', name: 'United Arab Emirates', nameZh: '阿联酋' },
  { code: 'GB', name: 'United Kingdom', nameZh: '英国' },
  { code: 'US', name: 'United States', nameZh: '美国' },
  { code: 'VN', name: 'Vietnam', nameZh: '越南' },
]

/**
 * 根据国家代码获取国家信息
 */
export function getCountryByCode(code: string): Country | undefined {
  return COUNTRIES.find(c => c.code.toUpperCase() === code.toUpperCase())
}

/**
 * 获取国家名称（优先返回指定语言）
 */
export function getCountryName(code: string, locale: string = 'en'): string {
  const country = getCountryByCode(code)
  if (!country) return code
  
  const normalizedLocale = locale.toLowerCase().replace('_', '-')
  if (normalizedLocale === 'zh' || normalizedLocale.startsWith('zh-')) {
    return country.nameZh || country.name
  }
  return country.name
}

/**
 * 邮编格式配置
 */
export interface ZipFormatHint {
  placeholder: string
  pattern: string
  hint: string
}

export const ZIP_FORMAT_HINTS: Record<string, ZipFormatHint> = {
  US: { placeholder: '10001', pattern: '^\\d{5}(-\\d{4})?$', hint: '5 digits (e.g., 10001)' },
  GB: { placeholder: 'SW1A 1AA', pattern: '^[A-Z]{1,2}\\d[A-Z\\d]? ?\\d[A-Z]{2}$', hint: 'e.g., SW1A 1AA' },
  DE: { placeholder: '10115', pattern: '^\\d{5}$', hint: '5 digits (e.g., 10115)' },
  JP: { placeholder: '100-0001', pattern: '^\\d{3}-?\\d{4}$', hint: '7 digits (e.g., 100-0001)' },
  CA: { placeholder: 'K1A 0B1', pattern: '^[A-Z]\\d[A-Z] ?\\d[A-Z]\\d$', hint: 'e.g., K1A 0B1' },
  AU: { placeholder: '2000', pattern: '^\\d{4}$', hint: '4 digits (e.g., 2000)' },
  CN: { placeholder: '100000', pattern: '^\\d{6}$', hint: '6 digits (e.g., 100000)' },
  FR: { placeholder: '75001', pattern: '^\\d{5}$', hint: '5 digits (e.g., 75001)' },
  IT: { placeholder: '00100', pattern: '^\\d{5}$', hint: '5 digits (e.g., 00100)' },
  ES: { placeholder: '28001', pattern: '^\\d{5}$', hint: '5 digits (e.g., 28001)' },
  NL: { placeholder: '1012 AB', pattern: '^\\d{4} ?[A-Z]{2}$', hint: 'e.g., 1012 AB' },
  BE: { placeholder: '1000', pattern: '^\\d{4}$', hint: '4 digits (e.g., 1000)' },
  CH: { placeholder: '8001', pattern: '^\\d{4}$', hint: '4 digits (e.g., 8001)' },
  AT: { placeholder: '1010', pattern: '^\\d{4}$', hint: '4 digits (e.g., 1010)' },
  SE: { placeholder: '111 22', pattern: '^\\d{3} ?\\d{2}$', hint: 'e.g., 111 22' },
  NO: { placeholder: '0001', pattern: '^\\d{4}$', hint: '4 digits (e.g., 0001)' },
  DK: { placeholder: '1000', pattern: '^\\d{4}$', hint: '4 digits (e.g., 1000)' },
  FI: { placeholder: '00100', pattern: '^\\d{5}$', hint: '5 digits (e.g., 00100)' },
  PL: { placeholder: '00-001', pattern: '^\\d{2}-\\d{3}$', hint: 'e.g., 00-001' },
  CZ: { placeholder: '100 00', pattern: '^\\d{3} ?\\d{2}$', hint: 'e.g., 100 00' },
  HU: { placeholder: '1011', pattern: '^\\d{4}$', hint: '4 digits (e.g., 1011)' },
  PT: { placeholder: '1000-001', pattern: '^\\d{4}-\\d{3}$', hint: 'e.g., 1000-001' },
  GR: { placeholder: '104 31', pattern: '^\\d{3} ?\\d{2}$', hint: 'e.g., 104 31' },
  RU: { placeholder: '101000', pattern: '^\\d{6}$', hint: '6 digits (e.g., 101000)' },
  BR: { placeholder: '01310-100', pattern: '^\\d{5}-?\\d{3}$', hint: 'e.g., 01310-100' },
  MX: { placeholder: '01000', pattern: '^\\d{5}$', hint: '5 digits (e.g., 01000)' },
  IN: { placeholder: '110001', pattern: '^\\d{6}$', hint: '6 digits (e.g., 110001)' },
  KR: { placeholder: '03141', pattern: '^\\d{5}$', hint: '5 digits (e.g., 03141)' },
  SG: { placeholder: '018956', pattern: '^\\d{6}$', hint: '6 digits (e.g., 018956)' },
  MY: { placeholder: '50000', pattern: '^\\d{5}$', hint: '5 digits (e.g., 50000)' },
  TH: { placeholder: '10100', pattern: '^\\d{5}$', hint: '5 digits (e.g., 10100)' },
  VN: { placeholder: '100000', pattern: '^\\d{6}$', hint: '6 digits (e.g., 100000)' },
  PH: { placeholder: '1000', pattern: '^\\d{4}$', hint: '4 digits (e.g., 1000)' },
  ID: { placeholder: '10110', pattern: '^\\d{5}$', hint: '5 digits (e.g., 10110)' },
  TW: { placeholder: '100', pattern: '^\\d{3,5}$', hint: '3-5 digits (e.g., 100)' },
  HK: { placeholder: '', pattern: '', hint: 'No postal code required' },
  AE: { placeholder: '', pattern: '', hint: 'No postal code required' },
  SA: { placeholder: '11564', pattern: '^\\d{5}$', hint: '5 digits (e.g., 11564)' },
  NZ: { placeholder: '1010', pattern: '^\\d{4}$', hint: '4 digits (e.g., 1010)' },
  ZA: { placeholder: '0001', pattern: '^\\d{4}$', hint: '4 digits (e.g., 0001)' },
}

/**
 * 获取邮编格式提示
 */
export function getZipFormatHint(countryCode: string): ZipFormatHint {
  return ZIP_FORMAT_HINTS[countryCode.toUpperCase()] || {
    placeholder: '',
    pattern: '',
    hint: 'Enter postal/zip code'
  }
}

/**
 * 验证邮编格式
 */
export function validateZipFormat(countryCode: string, zipCode: string): boolean {
  const hint = ZIP_FORMAT_HINTS[countryCode.toUpperCase()]
  if (!hint || !hint.pattern) {
    return true // 没有格式要求的国家，任何值都有效
  }
  
  const regex = new RegExp(hint.pattern, 'i')
  return regex.test(zipCode.trim())
}
