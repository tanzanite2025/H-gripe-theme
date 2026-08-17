export type ProductSeoAvailability = 'in_stock' | 'out_of_stock'

export interface ProductSeoOfferInput {
  price?: number | null
  currency?: string | null
  availability?: ProductSeoAvailability | null
  sku?: string | null
}

export interface ProductSeoAggregateRatingInput {
  ratingValue?: number | null
  reviewCount?: number | null
}

export interface ProductSeoShippingDetailsInput {
  country?: string | null
  amount?: number | null
  currency?: string | null
  freeShipping?: boolean | null
  etaMinDays?: number | null
  etaMaxDays?: number | null
}

export interface ProductSeoVariantInput {
  id?: number | string | null
  name?: string | null
  sku?: string | null
  price?: number | null
  currency?: string | null
  availability?: ProductSeoAvailability | null
  localizedPath?: string | null
  imageUrls?: Array<string | null | undefined> | null
  shippingDetails?: ProductSeoShippingDetailsInput | null
}

export interface ProductSeoInput {
  name?: string | null
  brand?: string | null
  metaTitle?: string | null
  metaDescription?: string | null
  shortDescription?: string | null
  description?: string | null
  sku?: string | null
  imageUrls?: Array<string | null | undefined> | null
  offer?: ProductSeoOfferInput | null
  aggregateRating?: ProductSeoAggregateRatingInput | null
  shippingDetails?: ProductSeoShippingDetailsInput | null
  productGroupId?: string | null
  variesBy?: string[] | null
  variants?: ProductSeoVariantInput[] | null
}

export interface ProductSeoContext {
  siteOrigin: string
  localizedPath: string
  mediaOrigins?: string[]
}

export interface ProductSeoOffer {
  '@type': 'Offer'
  price: number
  priceCurrency: string
  availability: `https://schema.org/${'InStock' | 'OutOfStock'}`
  url: string
  shippingDetails?: ProductSeoShippingDetails
}

export interface ProductSeoAggregateRating {
  '@type': 'AggregateRating'
  ratingValue: number
  reviewCount: number
  ratingCount: number
  bestRating: number
  worstRating: number
}

export interface ProductSeoShippingDetails {
  '@type': 'OfferShippingDetails'
  shippingRate: {
    '@type': 'MonetaryAmount'
    value: number
    currency: string
  }
  shippingDestination: {
    '@type': 'DefinedRegion'
    addressCountry: string
  }
  deliveryTime: {
    '@type': 'ShippingDeliveryTime'
    transitTime: {
      '@type': 'QuantitativeValue'
      minValue: number
      maxValue: number
      unitCode: 'DAY'
    }
  }
}

export interface ProductSeoSchema {
  '@context': 'https://schema.org'
  '@type': 'Product'
  name: string
  brand?: {
    '@type': 'Brand'
    name: string
  }
  description?: string
  image: string[]
  sku?: string
  url: string
  offers?: ProductSeoOffer
  aggregateRating?: ProductSeoAggregateRating
}

export interface ProductSeoProductGroupSchema {
  '@context': 'https://schema.org'
  '@type': 'ProductGroup'
  name: string
  url: string
  productGroupID: string
  brand?: {
    '@type': 'Brand'
    name: string
  }
  description?: string
  image: string[]
  variesBy?: string[]
  hasVariant: ProductSeoSchema[]
  aggregateRating?: ProductSeoAggregateRating
}

export type ProductSeoStructuredData = ProductSeoSchema | ProductSeoProductGroupSchema

export interface ProductSeoDocument {
  title: string
  description: string
  canonicalUrl: string
  images: string[]
  schema: ProductSeoStructuredData | null
}
