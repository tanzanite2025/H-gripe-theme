export interface CarbonRimDrillingAndLacingPhotoWallImage {
  id: string
  src: string
  alt: { en: string; zh_cn: string }
  caption: { en: string; zh_cn: string }
  mimeType: 'image/webp'
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
    caption: {
      en: 'Chris King\n32 holes · 3-cross',
      zh_cn: 'Chris King\n32 孔 · 三交叉',
    },
    mimeType: 'image/webp',
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
    caption: {
      en: 'Chris King\n24 holes · 2-cross',
      zh_cn: 'Chris King\n24 孔 · 二交叉',
    },
    mimeType: 'image/webp',
    width: 750,
    height: 750,
  },
  {
    id: 'white-industries-24-hole-two-cross-carbon-road-wheelset',
    src: '/media/company/about/hole-patterns/drilled-rim-lacing-photos/whiteindustries-30mm-24hole-2x-lacing-roadwheels.webp',
    alt: {
      en: 'Carbon road wheelset with White Industries hubs, 30 mm rims and 24-hole two-cross spoke lacing',
      zh_cn: '搭配 White Industries 花鼓及 30 mm 轮圈的 24 孔二交叉编法碳纤维公路轮组实拍',
    },
    caption: {
      en: 'White Industries\n24 holes · 2-cross',
      zh_cn: 'White Industries\n24 孔 · 二交叉',
    },
    mimeType: 'image/webp',
    width: 750,
    height: 750,
  },
  {
    id: 'bora-ultra-wto-35-21-hole-g3-lacing-carbon-wheel',
    src: '/media/company/about/hole-patterns/drilled-rim-lacing-photos/boraultra-wto-35-21hole-g3-lacing.webp',
    alt: {
      en: 'Bora Ultra WTO 35 carbon wheel with 21 spoke holes and grouped G3 spoke lacing',
      zh_cn: 'Bora Ultra WTO 35 碳纤维车轮的 21 孔 G3 分组辐条编法实拍',
    },
    caption: {
      en: 'Bora Ultra WTO 35\n21 holes · G3',
      zh_cn: 'Bora Ultra WTO 35\n21 孔 · G3',
    },
    mimeType: 'image/webp',
    width: 750,
    height: 750,
  },
  {
    id: 'carbon-ti-28-hole-two-cross-carbon-wheelset',
    src: '/media/company/about/hole-patterns/drilled-rim-lacing-photos/carbonti-28hole-2x-lacing.webp',
    alt: {
      en: 'Carbon wheelset with Carbon-Ti hubs and 28-hole two-cross spoke lacing',
      zh_cn: '搭配 Carbon-Ti 花鼓的 28 孔二交叉编法碳纤维轮组实拍',
    },
    caption: {
      en: 'Carbon-Ti\n28 holes · 2-cross',
      zh_cn: 'Carbon-Ti\n28 孔 · 二交叉',
    },
    mimeType: 'image/webp',
    width: 750,
    height: 750,
  },
  {
    id: 'hope-32-hole-three-cross-carbon-wheelset',
    src: '/media/company/about/hole-patterns/drilled-rim-lacing-photos/hope-32hole-3x-lacing.webp',
    alt: {
      en: 'Carbon wheelset with red Hope hubs and 32-hole three-cross spoke lacing',
      zh_cn: '搭配红色 Hope 花鼓的 32 孔三交叉编法碳纤维轮组实拍',
    },
    caption: {
      en: 'Hope\n32 holes · 3-cross',
      zh_cn: 'Hope\n32 孔 · 三交叉',
    },
    mimeType: 'image/webp',
    width: 750,
    height: 750,
  },
]
