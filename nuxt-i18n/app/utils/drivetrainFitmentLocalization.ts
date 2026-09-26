import type {
  DrivetrainCassetteRule,
  DrivetrainFitmentOption,
  DrivetrainSpacerRequirement,
} from '~/types/drivetrainFitment'

const englishDisplayNames: Record<string, string> = {
  'sram-12s-10-52t': '12-speed 10-52T / 10-50T',
  'sram-12s-11-50t': '12-speed 11-50T',
  'sram-12s-10-36t': '12-speed 10-28T / 10-33T / 10-36T',
  'sram-11s-10-42t': '11-speed 10-42T',
  'sram-11s-11-36t': '11-speed 11-28T to 11-36T',
  'shimano-12s-road-11-34t': '12-speed road 11-30T / 11-34T',
  'shimano-12s-mtb-10-51t': '12-speed MTB 10-45T / 10-51T',
  'shimano-11s-road-11-34t': '11-speed road 11-25T to 11-34T',
  'shimano-11s-mtb-11-42t': '11-speed MTB 11-40T / 11-42T / 11-46T',
  'shimano-10s-road-11-30t': '10-speed road 11/12-28T / 11-30T',
  'shimano-10s-tiagra-11-30t': 'Tiagra 10-speed 11/12-28T / 11-30T',
  'campagnolo-13s-ekar': '13-speed 9-42T / 10-44T',
  'campagnolo-11-12s-11-34t': '11/12-speed 11-29T to 11-34T',
  'campagnolo-classic-9-12s': '9-12-speed, 11T start',
  'ltwoo-road-12s-11-34t': 'Road 12-speed 11-32T / 11-34T',
  'ltwoo-mtb-12s-11-52t': 'MTB 12-speed 11-50T / 11-52T',
  'sensah-road-12s-11-34t': 'Road 12-speed 11-32T / 11-34T',
  'sensah-gravel-11-12s-11-50t': 'Gravel 11/12-speed 11-42T / 11-50T',
  'microshift-sword-10s-11-48t': 'Sword gravel 10-speed 11-48T',
  'microshift-advent-x-10s-11-48t': 'Advent X MTB 10-speed 11-48T',
  'sunshine-road-12s-11-34t': 'Road conversion 12-speed 11-30T to 11-34T',
  'sunshine-mtb-12s-11-52t': 'MTB conversion 12-speed 11-50T / 11-52T',
  'ztto-hg-12s-11-34t': 'SLR HG 12-speed 11-32T / 11-34T / 11-50T',
  'ztto-xd-12s-10-52t': 'SLR XD 12-speed 9-50T / 10-52T',
  'ztto-ms-12s-10-52t': 'SLR Micro Spline 12-speed 10-51T / 10-52T',
  'wheeltop-11-12s-11-34t': 'EDS TX 11/12-speed 11-32T / 11-34T',
}

const englishHints: Record<string, string> = {
  'sram-12s-10-52t': 'GX / X01 / XX1 Eagle / Transmission',
  'sram-12s-11-50t': 'NX / SX Eagle and third-party MTB 12-speed',
  'sram-12s-10-36t': 'RED / Force / Rival AXS road',
  'sram-11s-10-42t': 'SRAM 11-speed MTB XD cassette',
  'sram-11s-11-36t': 'SRAM 11-speed road PG cassette',
  'shimano-12s-road-11-34t': 'Dura-Ace R9200 / Ultegra R8100 / 105 R7100',
  'shimano-12s-mtb-10-51t': 'XTR M9100 / XT M8100 / SLX M7100 / Deore M6100',
  'shimano-11s-road-11-34t': 'Dura-Ace R9100 / Ultegra R8000 / 105 R7000',
  'shimano-11s-mtb-11-42t': 'M8000 / M7000 mountain cassette',
  'shimano-10s-road-11-30t': 'CS-6700 / CS-5700',
  'shimano-10s-tiagra-11-30t': 'CS-4600 / CS-HG500-10',
  'campagnolo-13s-ekar': 'Ekar gravel / road',
  'campagnolo-11-12s-11-34t': 'Super Record / Record / Chorus',
  'campagnolo-classic-9-12s': 'Campagnolo classic cassette',
  'ltwoo-road-12s-11-34t': 'eRX / RX / R9 12S',
  'ltwoo-mtb-12s-11-52t': 'A12 12S',
  'sensah-road-12s-11-34t': 'Empire Pro',
  'sensah-gravel-11-12s-11-50t': 'SRX Pro',
  'microshift-sword-10s-11-48t': 'Sword',
  'microshift-advent-x-10s-11-48t': 'Advent X',
  'sunshine-road-12s-11-34t': 'SUNSHINE / VG Sports',
  'sunshine-mtb-12s-11-52t': 'SUNSHINE / VG Sports',
  'ztto-hg-12s-11-34t': 'ZTTO SLR HG',
  'ztto-xd-12s-10-52t': 'ZTTO SLR XD',
  'ztto-ms-12s-10-52t': 'ZTTO SLR Micro Spline',
  'wheeltop-11-12s-11-34t': 'Wheeltop EDS TX',
}

const englishFreehubNames: Record<string, string> = {
  'HG-11': 'Shimano HG-11 (road)',
  HG: 'Shimano HG (mountain)',
  'HG-L2': 'Shimano HG L2',
  XDR: 'SRAM XDR',
  XD: 'SRAM XD',
  MICRO_SPLINE: 'Shimano Micro Spline',
  N3W: 'Campagnolo N3W',
  CAMPY_CLASSIC: 'Campagnolo classic',
}

export const isEnglishDrivetrainLocale = (locale: string) => locale.toLowerCase().startsWith('en')
const isChineseDrivetrainLocale = (locale: string) => locale.toLowerCase().startsWith('zh')
const usesBackendChineseDrivetrainText = (locale: string) => (
  !isEnglishDrivetrainLocale(locale) && isChineseDrivetrainLocale(locale)
)

export const localizedRuleDisplayName = (rule: DrivetrainCassetteRule, locale: string) => {
  if (usesBackendChineseDrivetrainText(locale)) return rule.display_name
  return englishDisplayNames[rule.rule_id]
    || `${rule.speed}-speed cassette (${rule.min_cog_teeth}T-${rule.max_cog_teeth}T)`
}

export const localizedRuleHintGroupsets = (rule: DrivetrainCassetteRule, locale: string) => {
  if (usesBackendChineseDrivetrainText(locale)) return rule.hint_groupsets
  return englishHints[rule.rule_id] || rule.brand
}

export const localizedFreehubName = (option: DrivetrainFitmentOption, locale: string) => {
  if (usesBackendChineseDrivetrainText(locale)) return option.display_name
  return englishFreehubNames[option.standard] || option.standard
}

export const localizedSpacerDescription = (spacer: DrivetrainSpacerRequirement, locale: string) => {
  if (usesBackendChineseDrivetrainText(locale)) {
    if (!spacer.required) return spacer.description
    if (spacer.part_code) return `${spacer.description} (${spacer.part_code})`
    return spacer.description
  }
  if (!spacer.required) return 'Direct installation; no spacer or adapter required.'
  if (spacer.part_code === 'AC21-N3W') {
    return 'Install the AC21-N3W 4.4 mm adapter sleeve and extended lockring.'
  }
  if (spacer.parts?.length === 2 && spacer.thickness_mm === 2.85) {
    return 'Install a 1.85 mm freehub conversion spacer plus the cassette’s 1.0 mm spacer (2.85 mm total).'
  }
  return `Install a ${spacer.thickness_mm} mm spacer at the inner base of the freehub.`
}

export const localizedOptionNotes = (option: DrivetrainFitmentOption, locale: string) => {
  if (usesBackendChineseDrivetrainText(locale)) return option.notes || ''
  if (option.standard === 'XDR' && option.spacer.required) {
    return 'XDR is 1.85 mm longer than XD; the spacer belongs behind the cassette.'
  }
  if (option.standard === 'N3W' && option.spacer.part_code === 'AC21-N3W') {
    return 'Use the extended lockring supplied with the adapter kit.'
  }
  if (option.spacer.required) return 'Confirm the spacer is seated fully against the freehub shoulder.'
  return 'No additional spacer is required.'
}

export const localizedMechanicalNotes = (rule: DrivetrainCassetteRule, locale: string) => {
  if (usesBackendChineseDrivetrainText(locale)) return rule.mechanical_notes
  if (rule.min_cog_teeth <= 9) {
    return 'A 9T smallest cog cannot clear the 34.80 mm HG body; use a stepped body such as XD, Micro Spline, or N3W.'
  }
  if (rule.min_cog_teeth === 10) {
    return 'A 10T smallest cog has a root diameter below the 34.80 mm HG outer diameter and needs a stepped freehub.'
  }
  if (rule.brand === 'Shimano' && rule.speed === 12 && rule.min_cog_teeth === 11) {
    return 'Shimano 12-speed road cassettes start at 11T and fit HG-11 or HG L2 without a spacer.'
  }
  return `${rule.speed}-speed cassette with an ${rule.min_cog_teeth}T smallest cog; follow the selected freehub option’s spacer instructions.`
}
