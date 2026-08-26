import { computed, ref, watch, type Ref } from 'vue'
import { useI18n, useRoute } from '#imports'
import {
  PRODUCT_DETAIL_HIDDEN_SPEC_SLUGS,
  displayPriceSnapshotForCurrency,
  humanizeProductSpecSlug,
  normalizeProductCurrencyCode,
  parseProductVariantOptions,
  validProductDisplayPrice,
} from '~/utils/productDetail'
import type {
  GoProduct,
  ProductVariant,
  ProductVariantOptionGroup,
} from '~/types/productDetail'

export function useProductDetailVariants(
  product: Ref<GoProduct | null | undefined>,
  displayCurrency: Ref<string>,
) {
  const route = useRoute()
  const { locale } = useI18n()
  const selectedVariantId = ref<number | null>(null)

  const activeVariants = computed(() => {
    return (product.value?.variants || []).filter((variant) => variant.is_active !== false)
  })

  const isVariantInStock = (variant: ProductVariant) => variant.availability === 'in_stock'
  const requestedVariantId = computed(() => {
    const value = Number(route.query.variant || 0)
    return Number.isFinite(value) && value > 0 ? value : 0
  })

  watch([product, requestedVariantId], ([currentProduct, variantId]) => {
    const variants = (currentProduct?.variants || []).filter((variant) => variant.is_active !== false)
    if (!variants.length) {
      selectedVariantId.value = null
      return
    }

    const requestedVariant = variants.find((variant) => variant.id === variantId)
    if (requestedVariant) {
      selectedVariantId.value = requestedVariant.id
      return
    }

    const defaultVariant = variants.find((variant) => variant.is_default && isVariantInStock(variant))
      || variants.find(isVariantInStock)
      || variants.find((variant) => variant.is_default)
      || variants[0]
    if (defaultVariant) selectedVariantId.value = defaultVariant.id
  }, { immediate: true })

  const selectedVariant = computed(() => {
    if (!selectedVariantId.value) return null
    return activeVariants.value.find((variant) => variant.id === selectedVariantId.value) || null
  })

  const variantOptionDefinitions = computed(() => {
    return (product.value?.product_specification_template?.spec_definitions || [])
      .filter((definition) => (
        definition.is_visible !== false
        && definition.is_variant_option
        && !PRODUCT_DETAIL_HIDDEN_SPEC_SLUGS.has(String(definition.slug || '').trim().toLowerCase())
      ))
      .sort((left, right) => {
        const leftOrder = Number(left.sort_order || 0)
        const rightOrder = Number(right.sort_order || 0)
        if (leftOrder !== rightOrder) return leftOrder - rightOrder
        return String(left.name || left.slug).localeCompare(String(right.name || right.slug))
      })
  })

  const specDefinitionsBySlug = computed(() => {
    const entries = (product.value?.product_specification_template?.spec_definitions || [])
      .filter((definition) => definition.slug)
      .map((definition) => [definition.slug, definition] as const)
    return new Map(entries)
  })

  const variantOptionSlugs = computed(() => {
    const slugs = variantOptionDefinitions.value.map((definition) => definition.slug)
    const seen = new Set(slugs)

    activeVariants.value.forEach((variant) => {
      Object.keys(parseProductVariantOptions(variant)).forEach((slug) => {
        if (!slug || seen.has(slug)) return
        if (PRODUCT_DETAIL_HIDDEN_SPEC_SLUGS.has(String(slug).trim().toLowerCase())) return
        const definition = specDefinitionsBySlug.value.get(slug)
        if (definition?.is_visible === false) return
        seen.add(slug)
        slugs.push(slug)
      })
    })

    return slugs
  })

  const currentVariantOptions = computed(() => {
    return selectedVariant.value ? parseProductVariantOptions(selectedVariant.value) : {}
  })

  const variantOptionMetadata = (slug: string, value: string) => {
    return (product.value?.variant_option_values || []).find((option) => (
      option.is_enabled !== false
      && option.spec_slug === slug
      && option.value_key === value
    ))
  }

  const variantOptionGroups = computed<ProductVariantOptionGroup[]>(() => {
    return variantOptionSlugs.value
      .map((slug) => {
        const definition = specDefinitionsBySlug.value.get(slug)
        const presentation: ProductVariantOptionGroup['presentation'] = definition?.presentation === 'color'
          ? 'color'
          : definition?.presentation === 'image'
            ? 'image'
            : 'text'
        const optionsByValue = new Map<string, ProductVariantOptionGroup['options'][number]>()

        activeVariants.value.forEach((variant) => {
          const value = String(parseProductVariantOptions(variant)[slug] || '').trim()
          if (!value) return

          const existing = optionsByValue.get(value)
          const available = isVariantInStock(variant)
          const metadata = variantOptionMetadata(slug, value)
          if (existing) {
            existing.available = existing.available || available
            return
          }

          optionsByValue.set(value, {
            value,
            label: metadata?.label || value,
            colorHex: metadata?.color_hex || '',
            swatchUrl: metadata?.swatch_url || '',
            selected: currentVariantOptions.value[slug] === value,
            available,
          })
        })

        return {
          slug,
          name: definition?.name || humanizeProductSpecSlug(slug),
          presentation,
          options: [...optionsByValue.values()],
        }
      })
      .filter((group) => group.options.length > 0)
  })

  const selectVariantOption = (slug: string, value: string) => {
    const requestedOptions = {
      ...currentVariantOptions.value,
      [slug]: value,
    }

    const isExactMatch = (variant: ProductVariant) => {
      const options = parseProductVariantOptions(variant)
      return Object.entries(requestedOptions).every(([key, expectedValue]) => (
        !expectedValue || options[key] === expectedValue
      ))
    }

    const isFallbackMatch = (variant: ProductVariant) => (
      parseProductVariantOptions(variant)[slug] === value
    )

    const exactVariant = activeVariants.value.find((variant) => (
      isExactMatch(variant) && isVariantInStock(variant)
    )) || activeVariants.value.find(isExactMatch)
    const fallbackVariant = activeVariants.value.find((variant) => (
      isFallbackMatch(variant) && isVariantInStock(variant)
    )) || activeVariants.value.find(isFallbackMatch)
    const nextVariant = exactVariant || fallbackVariant
    if (nextVariant) selectedVariantId.value = nextVariant.id
  }

  const variantLabel = (variant: ProductVariant) => {
    const options = Object.values(parseProductVariantOptions(variant)).filter(Boolean)
    const optionText = options.join(' / ')
    const title = variant.title || optionText || 'Option'
    const optionLabel = optionText && title !== optionText ? ` · ${optionText}` : ''
    const weightLabel = variant.weight_grams ? ` · ${variant.weight_grams}g` : ''
    return `${title}${optionLabel}${weightLabel}`
  }

  const variantChoices = computed(() => activeVariants.value.map((variant) => ({
    id: variant.id,
    label: variantLabel(variant),
  })))

  const selectedVariantWeight = computed(() => {
    const value = Number(selectedVariant.value?.weight_grams || 0)
    return Number.isFinite(value) && value > 0 ? Math.round(value) : null
  })

  const selectedCartTitle = computed(() => {
    const productName = product.value?.name || ''
    const variant = selectedVariant.value
    if (!variant) return productName

    const optionText = Object.values(parseProductVariantOptions(variant)).filter(Boolean).join(' / ')
    if (optionText) return `${productName} - ${optionText}`

    const variantTitle = String(variant.title || '').trim()
    if (variantTitle && variantTitle.toLowerCase() !== 'default') {
      return `${productName} - ${variantTitle}`
    }

    return productName
  })

  const effectivePrice = computed(() => {
    return selectedVariant.value?.sale_price
      ?? selectedVariant.value?.price
      ?? product.value?.sale_price
      ?? product.value?.price
      ?? 0
  })

  const currentCurrency = computed(() => {
    return normalizeProductCurrencyCode(
      selectedVariant.value?.currency || product.value?.currency,
    ) || 'USD'
  })

  const currentDisplayPrice = computed(() => {
    const selectedVariantDisplayPrice =
      validProductDisplayPrice(selectedVariant.value?.display_price)
      || displayPriceSnapshotForCurrency(selectedVariant.value?.display_prices, displayCurrency.value)
    if (selectedVariantDisplayPrice) return selectedVariantDisplayPrice

    const productDisplayPrice =
      validProductDisplayPrice(product.value?.display_price)
      || displayPriceSnapshotForCurrency(product.value?.display_prices, displayCurrency.value)
    if (productDisplayPrice) return productDisplayPrice

    return { amount: Number(effectivePrice.value || 0), currency: currentCurrency.value }
  })

  const selectedAvailability = computed(() => {
    if (selectedVariant.value) return selectedVariant.value.availability || 'out_of_stock'
    if (product.value && activeVariants.value.length === 0) {
      return product.value.availability || 'out_of_stock'
    }
    return 'out_of_stock'
  })

  const canAddToCart = computed(() => Boolean(
    product.value
    && Number(effectivePrice.value) > 0
    && selectedAvailability.value === 'in_stock',
  ))

  const formattedPrice = computed(() => {
    const raw = currentDisplayPrice.value.amount
    const numeric = Number(raw)
    if (!Number.isFinite(numeric)) return ''
    const currencyCode = currentDisplayPrice.value.currency
    if (!currencyCode) return numeric.toFixed(2)

    try {
      return new Intl.NumberFormat(locale.value.replace('_', '-'), {
        style: 'currency',
        currency: currencyCode,
        minimumFractionDigits: 2,
        maximumFractionDigits: 2,
      }).format(numeric)
    } catch {
      return numeric.toFixed(2)
    }
  })

  return {
    selectedVariantId,
    activeVariants,
    selectedVariant,
    selectedVariantWeight,
    selectedCartTitle,
    variantOptionDefinitions,
    variantOptionGroups,
    variantChoices,
    currentVariantOptions,
    parseVariantOptions: parseProductVariantOptions,
    variantLabel,
    selectVariantOption,
    effectivePrice,
    currentCurrency,
    currentDisplayPrice,
    selectedAvailability,
    canAddToCart,
    formattedPrice,
  }
}
