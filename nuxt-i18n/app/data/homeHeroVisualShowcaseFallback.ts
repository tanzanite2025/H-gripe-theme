import type { HomeHeroVisualShowcaseItem } from '~/types/homeHeroVisualShowcase'

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
  width: 900,
  height: 1200,
  desktopOrder,
})

export const homeHeroVisualShowcaseFallback: HomeHeroVisualShowcaseItem[] = [
  fallbackImage('fallback-factory-meeting', '/company/ourstory/ourstory/ourstory.webp', 'Factory meeting and engineering discussion', 'Factory-direct engineering', 'Engineering decisions made close to production.', 1),
  fallbackImage('fallback-pre-mold-workshop', '/company/ourstory/factory/factory-premoldlayupworkshop6.webp', 'Carbon pre-mold layup workshop', 'Carbon layup workshop', 'Controlled preparation for consistent carbon wheel parts.', 2),
  fallbackImage('fallback-hub-spoke-comparison', '/public/wheelsetbuyersguide/wheelcomponents/hubs/bicycle-hub-spoke-type-comparison-jbend-straightpull.webp', 'Hub spoke type comparison', 'Hub and spoke options', 'Compare the component choices behind a complete wheelset.', 3),
  fallbackImage('fallback-carbon-rim-finish', '/company/aboutus/appearance/carbon-rim-finish1.webp', 'Finished carbon rim detail', 'Finished carbon rim', 'A close look at the surface and finish of the final rim.', 4),
  fallbackImage('fallback-inspection-packing', '/company/ourstory/factory/factory-inspectionpacking18.webp', 'Final inspection and packing area', 'Final inspection', 'Every wheelset passes a final inspection before packing.', 5),
  fallbackImage('fallback-rim-profile', '/public/wheelsetbuyersguide/wheelcomponents/rim/carbon-rim-hooked-vs-hookless.webp', 'Carbon rim hooked and hookless profile comparison', 'Rim profile choices', 'Hooked and hookless profiles for different riding needs.', 6),
  fallbackImage('fallback-cnc-machining', '/company/ourstory/factory/factory-cncmachiningworkshop9.webp', 'CNC spoke and valve hole machining', 'CNC machining', 'Accurate drilling supports clean assembly and reliable tension.', 7),
  fallbackImage('fallback-wheel-building', '/testreport/wheelsetassembly/4/wheelsbuilding-and-check-spoke-tension.webp', 'Wheel building and spoke tension check', 'Wheel building and tension', 'Assembly and spoke tension are checked by the wheel builder.', 8),
  fallbackImage('fallback-prepreg-workshop', '/company/ourstory/factory/factory-carbonprepregsworkshop2.webp', 'Carbon prepreg workshop', 'Carbon prepreg preparation', 'Material preparation is part of the finished wheelset story.', 9),
]
