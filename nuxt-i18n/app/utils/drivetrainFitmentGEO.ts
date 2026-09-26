import type { DrivetrainCassetteRule, DrivetrainFitmentOption } from '../types/drivetrainFitment.js'

// Page metadata is separate from the API's knowledge_as_of value. The former
// describes this page revision; the latter describes when the rules were last
// reviewed.
export const DRIVETRAIN_PAGE_LAST_MODIFIED = '2026-09-25'

export const DRIVETRAIN_BRAND_ALIASES = [
  { name: 'L-TWOO', alternateNames: ['LTWOO', '蓝图'] },
  { name: 'SENSAH', alternateNames: ['顺泰'] },
  { name: 'microSHIFT', alternateNames: ['Microshift', '微转'] },
  { name: 'SUNSHINE', alternateNames: ['日晖'] },
  { name: 'ZTTO', alternateNames: ['ZTTO'] },
  { name: 'Wheeltop', alternateNames: ['WheelTop'] },
] as const

const normalizeAnchorToken = (value: string) => value
  .normalize('NFKD')
  .replace(/[\u0300-\u036f]/g, '')
  .toLowerCase()
  .replace(/[^a-z0-9]+/g, '-')
  .replace(/^-+|-+$/g, '')

export const drivetrainRuleAnchor = (rule: DrivetrainCassetteRule, option: DrivetrainFitmentOption) => {
  const ruleToken = normalizeAnchorToken(rule.rule_id || `${rule.brand}-${rule.cassette_spec}`)
  const optionToken = normalizeAnchorToken(option.standard)
  return `rule-${ruleToken}-${optionToken}`
}

export const drivetrainDirectAnswer = ({
  rule,
  option,
  cassette,
  freehub,
  mechanical,
}: {
  rule: DrivetrainCassetteRule
  option: DrivetrainFitmentOption
  cassette: string
  freehub: string
  mechanical: string
}) => {
  const spacerAnswer = option.spacer.required
    ? `requires ${option.spacer.thickness_mm} mm spacer`
    : 'requires no spacer'
  return `${rule.brand} ${cassette} uses ${freehub}; ${spacerAnswer}. ${mechanical}`
}
