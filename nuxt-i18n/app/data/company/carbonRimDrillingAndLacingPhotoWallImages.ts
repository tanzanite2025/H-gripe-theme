export interface CarbonRimDrillingAndLacingPhotoWallImage {
  id: string
  src: string
  alt: { en: string; zh_cn: string }
  width: number
  height: number
}

// Add the actual photographs from /public/media/company/about/hole-patterns/drilled-rim-lacing-photos/.
// Keep their original aspect ratios and describe the visible drilling/lacing details in both languages.
export const carbonRimDrillingAndLacingPhotoWallImages: CarbonRimDrillingAndLacingPhotoWallImage[] = [
  {
    id: 'chris-king-32-hole-three-cross-carbon-wheelset',
    src: '/media/company/about/hole-patterns/drilled-rim-lacing-photos/chirsking-32hole-3x-lacing.webp',
    alt: {
      en: 'Carbon wheelset with Chris King hubs and 32-hole three-cross spoke lacing',
      zh_cn: '搭配 Chris King 花鼓的 32 孔三交叉编法碳纤维轮组实拍',
    },
    width: 720,
    height: 720,
  },
  {
    id: 'chris-king-24-hole-two-cross-hub-and-spoke-detail',
    src: '/media/company/about/hole-patterns/drilled-rim-lacing-photos/chirsking-enverime-24hole-2x-lacing.webp',
    alt: {
      en: 'Close-up of a Chris King hub and 24-hole two-cross spoke lacing',
      zh_cn: 'Chris King 花鼓与 24 孔二交叉辐条编法细节实拍',
    },
    width: 750,
    height: 750,
  },
  {
    id: 'white-industries-24-hole-two-cross-carbon-road-wheelset',
    src: '/media/company/about/hole-patterns/drilled-rim-lacing-photos/whiteindustries-30mm-24hole-2x-lacing-roadwheels.webp',
    alt: {
      en: 'Carbon road wheelset with White Industries hubs and 24-hole two-cross spoke lacing',
      zh_cn: '搭配 White Industries 花鼓的 24 孔二交叉编法碳纤维公路轮组实拍',
    },
    width: 750,
    height: 750,
  },
]
