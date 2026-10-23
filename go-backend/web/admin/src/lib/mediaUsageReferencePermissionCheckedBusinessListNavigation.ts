import type { MediaUsageSummaryReference } from '@/api/mediaUsageReadOnlyRebuildableSummaryWithValidatedCoverageAndReferenceStates'

// These existing screens are lists. Do not fabricate record-detail links or
// attach unsupported ID queries that would silently open the wrong record.
export function mediaUsageReferenceAuthorizedBusinessListDestination(reference: MediaUsageSummaryReference, hasPermission: (permission: string) => boolean): { name: string } | null {
  if (!/^[1-9]\d*$/.test(reference.resource_id) || ['missing_parent', 'retained_deleted_parent', 'retained_deleted_placement', 'immutable_history'].includes(reference.usage_state)) return null
  const destinations: Record<string, [string, string]> = {
    product_media: ['product:view', 'CatalogProducts'], product_rich_text: ['product:view', 'CatalogProducts'], product_option_swatch: ['product:view', 'CatalogProducts'],
    product_template_option_swatch: ['product:view', 'CatalogProductTemplates'], product_category: ['product:view', 'CatalogProductCategories'], product_brand_logo: ['product:view', 'CatalogProductBrands'],
    product_information_template_rich_text: ['product:view', 'CatalogProductInformationTemplates'], product_category_rich_text: ['product:view', 'CatalogProductCategories'], product_category_translation_rich_text: ['product:view', 'CatalogProductCategories'],
    post: ['content:view', 'ContentBlog'], home_visual_tile: ['content:view', 'ContentVisualShowcase'], faq: ['faq:view', 'ContentFAQs'],
    gallery_cover: ['gallery:view', 'ContentBrandGallery'], gallery_image: ['gallery:view', 'ContentBrandGallery'], showcase: ['gallery:view', 'ContentShowcase'],
    review: ['review:view', 'ContentReviews'], workbench_feed_media: ['workbench_feed:view', 'WorkbenchFeed'],
    website_profile_setting: ['settings:view', 'SiteIntroductionWebsiteProfile'], site_identity_setting: ['settings:view', 'SettingsSite'], site_logo: ['settings:view', 'SettingsSite'], site_favicon: ['settings:view', 'SettingsSite'],
  }
  const destination = destinations[reference.resource_type]
  return destination && hasPermission(destination[0]) ? { name: destination[1] } : null
}
