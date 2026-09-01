import type {
  VisualShowcaseAdministrationImageUploadResponse,
  VisualShowcaseAdministrationItemApiRecord,
  VisualShowcaseAdministrationItemFormState,
  VisualShowcaseAdministrationItemSavePayload,
  VisualShowcaseLayoutVariant,
} from '@/modules/visual-showcase/visualShowcaseTypes'
import {
  HOME_HERO_VISUAL_SHOWCASE_REQUIRED_ITEM_COUNT,
  HOME_MAIN_PRODUCT_CATEGORIES_REQUIRED_ITEM_COUNT,
} from '@/modules/visual-showcase/visualShowcaseTypes'

const homeHeroFallbackLabels = [
  ['Factory-direct engineering', 'Engineering decisions made close to production.', 'Factory meeting and wheelset engineering discussion'],
  ['Carbon layup workshop', 'Controlled preparation for consistent carbon wheel parts.', 'Carbon layup workshop for wheelset production'],
  ['Hub and spoke options', 'Compare the component choices behind a complete wheelset.', 'Hub and spoke parts for carbon wheelsets'],
  ['Finished carbon rim', 'A close look at the surface and finish of the final rim.', 'Finished carbon rim detail'],
  ['Final inspection', 'Every wheelset passes final inspection before packing.', 'Wheelset inspection and packing process'],
  ['Rim profile choices', 'Hooked and hookless profiles for different riding needs.', 'Carbon rim profile choices'],
  ['CNC machining', 'Accurate drilling supports clean assembly and reliable tension.', 'CNC machining for carbon wheel components'],
  ['Wheel building and tension', 'Assembly and spoke tension are checked by the wheel builder.', 'Wheel building and spoke tension check'],
  ['Carbon prepreg preparation', 'Material preparation is part of the finished wheelset story.', 'Carbon prepreg preparation workshop'],
] as const

const homeMainProductCategoriesFallbackLabels = [
  ['Road wheelsets', 'Fast, balanced wheelsets for everyday road speed.', 'Road wheelset being built and checked for spoke tension'],
  ['Gravel wheelsets', 'Stable carbon builds for mixed surfaces and long rides.', 'Carbon rim hooked and hookless profile comparison'],
  ['MTB wheelsets', 'Confident handling with component choices built for the trail.', 'Bicycle hub and spoke type comparison'],
  ['Carbon rims', 'Start with the rim profile, depth, and finish that suit your build.', 'Finished carbon bicycle rim detail'],
  ['Hubs and spokes', 'Choose the component foundation behind a complete wheelset.', 'DT Swiss hub for a custom wheel build'],
  ['Custom builds', 'Need a specific setup? Build the right combination with support.', 'Wheel building tools used during final assembly'],
] as const

const toInteger = (value: unknown, fallback: number): number => {
  const parsed = Number(value)
  return Number.isFinite(parsed) ? Math.trunc(parsed) : fallback
}

const positiveInteger = (value: unknown, fallback: number): number => {
  const parsed = toInteger(value, fallback)
  return parsed > 0 ? parsed : fallback
}

const text = (value: unknown): string => String(value ?? '').trim()

const layoutVariant = (value: unknown): VisualShowcaseLayoutVariant => {
  const normalized = text(value)
  return normalized === 'offset' || normalized === 'wide' ? normalized : 'standard'
}

const mobilePairIndex = (value: unknown, fallback: number): number => {
  const parsed = toInteger(value, fallback)
  return Math.min(3, Math.max(0, parsed))
}

const fallbackForIndex = (
  index: number,
  labels: readonly (readonly [string, string, string])[],
) => labels[index % labels.length]

const hasValidAspectRatio = (
  width: number,
  height: number,
  ratioWidth: number,
  ratioHeight: number,
): boolean => (
  width > 0 && height > 0 && ratioWidth > 0 && ratioHeight > 0 && width * ratioHeight === height * ratioWidth
)

const imageURLFromSource = (source: VisualShowcaseAdministrationItemApiRecord): string => (
  text(source.image_url || source.thumbnail_url)
)

const createVisualShowcaseAdministrationItemFormState = (
  index: number,
  source: VisualShowcaseAdministrationItemApiRecord = {},
  labels: readonly (readonly [string, string, string])[] = homeHeroFallbackLabels,
  clientIdPrefix = 'home-hero-visual-showcase-item',
  defaultWidth = 900,
  defaultHeight = 1200,
): VisualShowcaseAdministrationItemFormState => {
  const [fallbackTitle, fallbackCaption, fallbackAltText] = fallbackForIndex(index, labels)
  const imageURL = imageURLFromSource(source)
  const thumbnailURL = text(source.thumbnail_url || source.image_url)

  return {
    client_id: source.id
      ? `visual-showcase-item-${source.id}`
      : `${clientIdPrefix}-${index + 1}`,
    id: toInteger(source.id, 0) > 0 ? toInteger(source.id, 0) : undefined,
    image_url: imageURL,
    thumbnail_url: thumbnailURL || imageURL,
    storage_key: text(source.storage_key),
    title: text(source.title) || fallbackTitle,
    caption: text(source.caption) || fallbackCaption,
    alt_text: text(source.alt_text) || fallbackAltText,
    desktop_order: positiveInteger(source.desktop_order, index + 1),
    mobile_pair_index: mobilePairIndex(source.mobile_pair_index, Math.floor(index / 2)),
    target_url: text(source.target_url),
    target_label: text(source.target_label),
    layout_variant: layoutVariant(source.layout_variant),
    is_published: source.is_published !== false,
    width: positiveInteger(source.width, defaultWidth),
    height: positiveInteger(source.height, defaultHeight),
  }
}

export const createVisualShowcaseHomeHeroAdministrationItemFormState = (
  index: number,
  source: VisualShowcaseAdministrationItemApiRecord = {},
): VisualShowcaseAdministrationItemFormState => createVisualShowcaseAdministrationItemFormState(
  index,
  source,
  homeHeroFallbackLabels,
  'home-hero-visual-showcase-item',
  900,
  1200,
)

export const createVisualShowcaseHomeMainProductCategoryAdministrationItemFormState = (
  index: number,
  source: VisualShowcaseAdministrationItemApiRecord = {},
): VisualShowcaseAdministrationItemFormState => createVisualShowcaseAdministrationItemFormState(
  index,
  source,
  homeMainProductCategoriesFallbackLabels,
  'home-main-product-category-item',
  1600,
  900,
)

const rowsFromApiItems = (
  items: VisualShowcaseAdministrationItemApiRecord[] = [],
  requiredItemCount: number,
  createItem: (
    index: number,
    source?: VisualShowcaseAdministrationItemApiRecord,
  ) => VisualShowcaseAdministrationItemFormState,
): VisualShowcaseAdministrationItemFormState[] => {
  const rows = [...items]
    .sort((left, right) => toInteger(left.desktop_order, 0) - toInteger(right.desktop_order, 0))
    .slice(0, requiredItemCount)
    .map((item, index) => createItem(index, item))

  while (rows.length < requiredItemCount) {
    rows.push(createItem(rows.length))
  }

  return rows
}

export const visualShowcaseHomeHeroAdministrationRowsFromApiItems = (
  items: VisualShowcaseAdministrationItemApiRecord[] = [],
): VisualShowcaseAdministrationItemFormState[] => rowsFromApiItems(
  items,
  HOME_HERO_VISUAL_SHOWCASE_REQUIRED_ITEM_COUNT,
  createVisualShowcaseHomeHeroAdministrationItemFormState,
)

export const visualShowcaseHomeMainProductCategoriesAdministrationRowsFromApiItems = (
  items: VisualShowcaseAdministrationItemApiRecord[] = [],
): VisualShowcaseAdministrationItemFormState[] => rowsFromApiItems(
  items,
  HOME_MAIN_PRODUCT_CATEGORIES_REQUIRED_ITEM_COUNT,
  createVisualShowcaseHomeMainProductCategoryAdministrationItemFormState,
)

export const visualShowcaseAdministrationSavePayloadFromFormRow = (
  row: VisualShowcaseAdministrationItemFormState,
  index: number,
  defaultWidth = 900,
  defaultHeight = 1200,
): VisualShowcaseAdministrationItemSavePayload => ({
  image_url: text(row.image_url),
  thumbnail_url: text(row.thumbnail_url || row.image_url),
  storage_key: text(row.storage_key),
  title: text(row.title),
  caption: text(row.caption),
  alt_text: text(row.alt_text),
  desktop_order: positiveInteger(row.desktop_order, index + 1),
  mobile_pair_index: mobilePairIndex(row.mobile_pair_index, Math.floor(index / 2)),
  target_url: text(row.target_url),
  target_label: text(row.target_label),
  layout_variant: layoutVariant(row.layout_variant),
  is_published: row.is_published,
  width: positiveInteger(row.width, defaultWidth),
  height: positiveInteger(row.height, defaultHeight),
})

export const visualShowcaseHomeHeroAdministrationSavePayloadFromFormRow = (
  row: VisualShowcaseAdministrationItemFormState,
  index: number,
): VisualShowcaseAdministrationItemSavePayload => visualShowcaseAdministrationSavePayloadFromFormRow(row, index)

export const visualShowcaseHomeMainProductCategoriesAdministrationSavePayloadFromFormRow = (
  row: VisualShowcaseAdministrationItemFormState,
  index: number,
): VisualShowcaseAdministrationItemSavePayload => visualShowcaseAdministrationSavePayloadFromFormRow(row, index, 1600, 900)

const visualShowcaseAdministrationValidationMessage = (
  rows: VisualShowcaseAdministrationItemFormState[],
  requiredItemCount: number,
  sectionLabel: string,
  ratioWidth: number,
  ratioHeight: number,
  ratioLabel: string,
): string => {
  if (rows.length < requiredItemCount) {
    return `${sectionLabel}至少需要 ${requiredItemCount} 张图片`
  }

  for (const [index, row] of rows.entries()) {
    const label = `第 ${index + 1} 张`
    if (!text(row.image_url) || !text(row.storage_key)) return `${label} 需要先上传图片`
    if (!text(row.title)) return `${label} 缺少标题`
    if (!text(row.alt_text)) return `${label} 缺少 ALT 文本`
    if (!hasValidAspectRatio(row.width, row.height, ratioWidth, ratioHeight)) return `${label} 图片必须为 ${ratioLabel} 比例，请重新上传`
  }

  return ''
}

export const visualShowcaseHomeHeroAdministrationValidationMessage = (
  rows: VisualShowcaseAdministrationItemFormState[],
): string => visualShowcaseAdministrationValidationMessage(
  rows,
  HOME_HERO_VISUAL_SHOWCASE_REQUIRED_ITEM_COUNT,
  '',
  3,
  4,
  '3:4',
)

export const visualShowcaseHomeMainProductCategoriesAdministrationValidationMessage = (
  rows: VisualShowcaseAdministrationItemFormState[],
): string => visualShowcaseAdministrationValidationMessage(
  rows,
  HOME_MAIN_PRODUCT_CATEGORIES_REQUIRED_ITEM_COUNT,
  '首页主力产品',
  16,
  9,
  '16:9',
)

const titleBaseFromUpload = (upload: VisualShowcaseAdministrationImageUploadResponse): string => (
  text(upload.original_filename || upload.filename).replace(/\.[^.]+$/, '') || 'Visual Showcase Image'
)

export const applyVisualShowcaseUploadToFormState = (
  row: VisualShowcaseAdministrationItemFormState,
  upload: VisualShowcaseAdministrationImageUploadResponse,
): VisualShowcaseAdministrationItemFormState => {
  const fallbackTitle = titleBaseFromUpload(upload)
  return {
    ...row,
    image_url: text(upload.image_url) || row.image_url,
    thumbnail_url: text(upload.thumbnail_url) || text(upload.image_url) || row.thumbnail_url,
    storage_key: text(upload.storage_key) || row.storage_key,
    title: text(row.title) || fallbackTitle,
    alt_text: text(row.alt_text) || fallbackTitle,
    width: positiveInteger(upload.width, row.width),
    height: positiveInteger(upload.height, row.height),
  }
}

