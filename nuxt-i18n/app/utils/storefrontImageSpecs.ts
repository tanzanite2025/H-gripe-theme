export type StorefrontImagePreset =
  | 'avatar'
  | 'badge'
  | 'card'
  | 'content'
  | 'detail'
  | 'gallery'
  | 'hero'
  | 'history'
  | 'logo'
  | 'swatch'
  | 'thumbnail'

export interface StorefrontImageSpec {
  /**
   * CSS display slot sizes. @nuxt/image turns these into request widths.
   * The source asset is never rewritten by this contract.
   */
  sizes: string
  /**
   * Device-pixel-ratio variants requested for the slot.
   * 2x is important for small raster assets on high-density phones.
   */
  densities: string
}

export const STOREFRONT_IMAGE_SPECS: Record<StorefrontImagePreset, StorefrontImageSpec> = {
  avatar: {
    sizes: 'xs:56px sm:64px',
    densities: '1x 2x',
  },
  badge: {
    sizes: 'xs:96px sm:96px',
    densities: '1x 2x',
  },
  card: {
    sizes: 'xs:50vw sm:33vw md:280px lg:320px',
    densities: '1x 2x',
  },
  content: {
    sizes: 'xs:100vw sm:100vw md:768px lg:1024px',
    densities: '1x 2x',
  },
  detail: {
    sizes: 'xs:100vw sm:100vw md:50vw lg:640px xl:800px',
    densities: '1x 2x',
  },
  gallery: {
    sizes: 'xs:100vw sm:100vw md:50vw lg:640px xl:960px',
    densities: '1x 2x',
  },
  hero: {
    sizes: 'xs:100vw sm:100vw md:100vw lg:1280px',
    densities: '1x 2x',
  },
  history: {
    sizes: 'xs:160px sm:160px md:160px',
    densities: '1x 2x',
  },
  // The current header renders the square SVG mark at roughly 48px on
  // mobile and 55px on the desktop header. These are display slots, not
  // upload dimensions.
  logo: {
    sizes: 'xs:48px sm:48px md:48px lg:48px xl:55px',
    densities: '1x 2x',
  },
  swatch: {
    sizes: 'xs:40px sm:40px',
    densities: '1x 2x',
  },
  thumbnail: {
    sizes: 'xs:72px sm:96px',
    densities: '1x 2x',
  },
}
