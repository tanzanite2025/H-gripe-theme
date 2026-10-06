export interface ShippingCountryOption {
  code: string
  name: string
  nameZh: string
}

// The admin country selector uses the same ISO alpha-2 contract as storefront
// shipping validation. Keep this catalog independent from carrier collections.
export const SHIPPING_COUNTRY_OPTIONS: ShippingCountryOption[] = [
  ['AF', 'Afghanistan', '阿富汗'], ['AL', 'Albania', '阿尔巴尼亚'], ['DZ', 'Algeria', '阿尔及利亚'],
  ['AR', 'Argentina', '阿根廷'], ['AM', 'Armenia', '亚美尼亚'], ['AU', 'Australia', '澳大利亚'],
  ['AT', 'Austria', '奥地利'], ['AZ', 'Azerbaijan', '阿塞拜疆'], ['BH', 'Bahrain', '巴林'],
  ['BD', 'Bangladesh', '孟加拉国'], ['BY', 'Belarus', '白俄罗斯'], ['BE', 'Belgium', '比利时'],
  ['BR', 'Brazil', '巴西'], ['BG', 'Bulgaria', '保加利亚'], ['KH', 'Cambodia', '柬埔寨'],
  ['CA', 'Canada', '加拿大'], ['CL', 'Chile', '智利'], ['CN', 'China', '中国'],
  ['CO', 'Colombia', '哥伦比亚'], ['HR', 'Croatia', '克罗地亚'], ['CY', 'Cyprus', '塞浦路斯'],
  ['CZ', 'Czech Republic', '捷克'], ['DK', 'Denmark', '丹麦'], ['EG', 'Egypt', '埃及'],
  ['EE', 'Estonia', '爱沙尼亚'], ['FI', 'Finland', '芬兰'], ['FR', 'France', '法国'],
  ['GE', 'Georgia', '格鲁吉亚'], ['DE', 'Germany', '德国'], ['GR', 'Greece', '希腊'],
  ['HK', 'Hong Kong', '香港'], ['HU', 'Hungary', '匈牙利'], ['IS', 'Iceland', '冰岛'],
  ['IN', 'India', '印度'], ['ID', 'Indonesia', '印度尼西亚'], ['IR', 'Iran', '伊朗'],
  ['IQ', 'Iraq', '伊拉克'], ['IE', 'Ireland', '爱尔兰'], ['IL', 'Israel', '以色列'],
  ['IT', 'Italy', '意大利'], ['JP', 'Japan', '日本'], ['JO', 'Jordan', '约旦'],
  ['KZ', 'Kazakhstan', '哈萨克斯坦'], ['KE', 'Kenya', '肯尼亚'], ['KR', 'South Korea', '韩国'],
  ['KW', 'Kuwait', '科威特'], ['LV', 'Latvia', '拉脱维亚'], ['LB', 'Lebanon', '黎巴嫩'],
  ['LT', 'Lithuania', '立陶宛'], ['LU', 'Luxembourg', '卢森堡'], ['MO', 'Macau', '澳门'],
  ['MY', 'Malaysia', '马来西亚'], ['MX', 'Mexico', '墨西哥'], ['MA', 'Morocco', '摩洛哥'],
  ['NL', 'Netherlands', '荷兰'], ['NZ', 'New Zealand', '新西兰'], ['NG', 'Nigeria', '尼日利亚'],
  ['NO', 'Norway', '挪威'], ['OM', 'Oman', '阿曼'], ['PK', 'Pakistan', '巴基斯坦'],
  ['PA', 'Panama', '巴拿马'], ['PE', 'Peru', '秘鲁'], ['PH', 'Philippines', '菲律宾'],
  ['PL', 'Poland', '波兰'], ['PT', 'Portugal', '葡萄牙'], ['QA', 'Qatar', '卡塔尔'],
  ['RO', 'Romania', '罗马尼亚'], ['RU', 'Russia', '俄罗斯'], ['SA', 'Saudi Arabia', '沙特阿拉伯'],
  ['RS', 'Serbia', '塞尔维亚'], ['SG', 'Singapore', '新加坡'], ['SK', 'Slovakia', '斯洛伐克'],
  ['SI', 'Slovenia', '斯洛文尼亚'], ['ZA', 'South Africa', '南非'], ['ES', 'Spain', '西班牙'],
  ['LK', 'Sri Lanka', '斯里兰卡'], ['SE', 'Sweden', '瑞典'], ['CH', 'Switzerland', '瑞士'],
  ['TW', 'Taiwan', '台湾'], ['TH', 'Thailand', '泰国'], ['TR', 'Turkey', '土耳其'],
  ['UA', 'Ukraine', '乌克兰'], ['AE', 'United Arab Emirates', '阿联酋'],
  ['GB', 'United Kingdom', '英国'], ['US', 'United States', '美国'], ['VN', 'Vietnam', '越南'],
].map(([code, name, nameZh]) => ({ code, name, nameZh }))
