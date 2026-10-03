import {
  HOME_HERO_VISUAL_SHOWCASE_IMAGE_DIMENSION,
  type HomeHeroVisualShowcaseItem,
} from '~/types/homeHeroVisualShowcase'

const fallbackImage = (
  id: string,
  src: string,
  altText: string,
  title: string,
  caption: string,
  desktopOrder: number,
): HomeHeroVisualShowcaseItem => ({
  id,
  showcaseKey: 'home-hero',
  locale: 'en',
  src,
  altText,
  title,
  caption,
  width: HOME_HERO_VISUAL_SHOWCASE_IMAGE_DIMENSION,
  height: HOME_HERO_VISUAL_SHOWCASE_IMAGE_DIMENSION,
  desktopOrder,
})

export const homeHeroVisualShowcaseFallback: HomeHeroVisualShowcaseItem[] = [
  fallbackImage('fallback-factory-meeting', '/company/ourstory/ourstory/ourstory.webp', 'Factory meeting and engineering discussion', 'Factory-direct engineering', 'Engineering decisions made close to production.', 1),
  fallbackImage('fallback-pre-mold-workshop', '/company/ourstory/factory/factory-premoldlayupworkshop6.webp', 'Carbon pre-mold layup workshop', 'Carbon layup workshop', 'Controlled preparation for consistent carbon wheel parts.', 2),
  fallbackImage('fallback-carbon-rim-finish', '/company/aboutus/appearance/carbon-rim-finish1.webp', 'Finished carbon rim detail', 'Finished carbon rim', 'A close look at the surface and finish of the final rim.', 3),
  fallbackImage('fallback-inspection-packing', '/company/ourstory/factory/factory-inspectionpacking18.webp', 'Final inspection and packing area', 'Final inspection', 'Every wheelset passes a final inspection before packing.', 4),
  fallbackImage('fallback-cnc-machining', '/company/ourstory/factory/factory-cncmachiningworkshop9.webp', 'CNC spoke and valve hole machining', 'CNC machining', 'Accurate drilling supports clean assembly and reliable tension.', 5),
  fallbackImage('fallback-wheel-building', '/testreport/wheelsetassembly/4/wheelsbuilding-and-check-spoke-tension.webp', 'Wheel building and spoke tension check', 'Wheel building and tension', 'Assembly and spoke tension are checked by the wheel builder.', 6),
  fallbackImage('fallback-prepreg-workshop', '/company/ourstory/factory/factory-carbonprepregsworkshop2.webp', 'Carbon prepreg workshop', 'Carbon prepreg preparation', 'Material preparation is part of the finished wheelset story.', 7),
  fallbackImage('fallback-cutting-workshop', '/company/ourstory/factory/factory-cuttingworkshop4.webp', 'Carbon material cutting workshop', 'Material cutting', 'Accurate material cutting prepares each carbon component for production.', 8),
  fallbackImage('fallback-grinding-workshop', '/company/ourstory/factory/factory-grinding12.webp', 'Carbon rim grinding workshop', 'Surface finishing', 'Controlled surface finishing prepares the rim for inspection and final coating.', 9),
]
