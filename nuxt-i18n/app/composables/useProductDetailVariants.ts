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
  ProductAvailability,
  ProductVariant,
  ProductVariantOptionGroup,
  ProductCustomOptionGroup,
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

  const isVariantPurchasable = (variant: ProductVariant) => (
    variant.availability === 'in_stock' || variant.availability === 'made_to_order'
  )
  const definitionRole = (definition: { role?: string } | undefined): string => {
    const role = String(definition?.role || '').trim()
    if (role === 'attribute' || role === 'variant' || role === 'custom_option') return role
    return 'attribute'
  }
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

    const defaultVariant = variants.find((variant) => variant.is_default && isVariantPurchasable(variant))
      || variants.find(isVariantPurchasable)
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
        && definitionRole(definition) === 'variant'
        && !PRODUCT_DETAIL_HIDDEN_SPEC_SLUGS.has(String(definition.slug || '').trim().toLowerCase())
      ))
      .sort((left, right) => {
        const leftOrder = Number(left.sort_order || 0)
        const rightOrder = Number(right.sort_order || 0)
        if (leftOrder !== rightOrder) return leftOrder - rightOrder
        return String(left.name || left.slug).localeCompare(String(right.name || right.slug))
      })
  })

  const customOptionDefinitions = computed(() => {
    const currentVariant = selectedVariant.value
    const groupRules = new Map((currentVariant?.option_group_rules || []).map(rule => [Number(rule.spec_definition_id), rule]))
    return (product.value?.product_specification_template?.spec_definitions || [])
      .filter((definition) => definition.is_visible !== false && definitionRole(definition) === 'custom_option')
      .filter((definition) => groupRules.get(Number(definition.id))?.is_applicable !== false)
      .sort((left, right) => Number(left.sort_order || 0) - Number(right.sort_order || 0))
  })

  const selectedCustomOptions = ref<Record<string, string[]>>({})

  watch([product, selectedVariant], () => {
    const next: Record<string, string[]> = {}
    const valueRules = new Map((selectedVariant.value?.option_value_rules || []).map(rule => [Number(rule.product_variant_option_value_id), rule]))
    customOptionDefinitions.value.forEach((definition) => {
      const defaults = (product.value?.variant_option_values || [])
        .filter(option => option.spec_slug === definition.slug && option.is_enabled !== false)
        .filter(option => valueRules.get(Number(option.id))?.is_enabled !== false)
        .filter(option => option.is_default)
        .map(option => option.value_key)
      if (defaults.length) next[definition.slug] = defaults
    })
    selectedCustomOptions.value = next
  }, { immediate: true })

  const customOptionGroups = computed(() => customOptionDefinitions.value.map((definition) => {
    const currentVariant = selectedVariant.value
    const valueRules = new Map((currentVariant?.option_value_rules || []).map(rule => [Number(rule.product_variant_option_value_id), rule]))
    const groupRule = (currentVariant?.option_group_rules || [])
      .find(rule => Number(rule.spec_definition_id) === Number(definition.id))
    const minSelections = groupRule?.min_selections_override ?? definition.min_selections ?? 0
    const maxSelections = groupRule?.max_selections_override ?? definition.max_selections ?? null
    const values = (product.value?.variant_option_values || [])
      .filter(option => option.spec_slug === definition.slug && option.is_enabled !== false)
      .map(option => {
        const rule = valueRules.get(Number(option.id))
        return {
          value: option.value_key,
          label: option.label,
          colorHex: option.color_hex || '',
          swatchUrl: option.swatch_url || '',
          selected: (selectedCustomOptions.value[definition.slug] || []).includes(option.value_key),
          available: rule?.is_enabled !== false,
          unavailableReason: rule?.unavailable_reason || '',
          priceDeltaMinor: rule?.price_delta_minor_override ?? option.price_delta_minor ?? 0,
        }
      })
    return {
      slug: definition.slug,
      name: definition.name,
      selectionMode: definition.selection_mode || 'single',
      minSelections,
      maxSelections,
      presentation: definition.presentation || 'text',
      options: values,
      selectedCount: selectedCustomOptions.value[definition.slug]?.length || 0,
      isValid: (selectedCustomOptions.value[definition.slug]?.length || 0) >= minSelections
        && (maxSelections == null || (selectedCustomOptions.value[definition.slug]?.length || 0) <= maxSelections),
    }
  }))

  const selectCustomOption = (slug: string, value: string) => {
    const group = customOptionGroups.value.find(item => item.slug === slug)
    if (!group) return
    const current = selectedCustomOptions.value[slug] || []
    if (group.selectionMode === 'multiple') {
      selectedCustomOptions.value = { ...selectedCustomOptions.value, [slug]: current.includes(value) ? current.filter(item => item !== value) : [...current, value] }
    } else {
      selectedCustomOptions.value = { ...selectedCustomOptions.value, [slug]: [value] }
    }
  }

  const selectedOptions = computed(() => Object.entries(selectedCustomOptions.value).map(([groupSlug, valueKeys]) => ({ group_slug: groupSlug, value_keys: valueKeys })))
  const customOptionsValid = computed(() => customOptionGroups.value.every((group) => {
    const count = selectedCustomOptions.value[group.slug]?.length || 0
    return count >= group.minSelections && (group.maxSelections == null || count <= group.maxSelections)
  }))

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
          const available = isVariantPurchasable(variant)
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
      isExactMatch(variant) && isVariantPurchasable(variant)
    )) || activeVariants.value.find(isExactMatch)
    const fallbackVariant = activeVariants.value.find((variant) => (
      isFallbackMatch(variant) && isVariantPurchasable(variant)
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

  const selectedCustomOptionPriceDeltaMinor = computed(() => customOptionGroups.value.reduce((total, group) => (
    total + group.options
      .filter(option => option.selected)
      .reduce((subtotal, option) => subtotal + Number(option.priceDeltaMinor || 0), 0)
  ), 0))

  const minorUnitsForCurrency = (value: string) => {
    const code = normalizeProductCurrencyCode(value)
    if (['BHD', 'IQD', 'JOD', 'KWD', 'LYD', 'OMR', 'TND'].includes(code)) return 3
    if (['BIF', 'CLP', 'DJF', 'GNF', 'JPY', 'KMF', 'KRW', 'MGA', 'PYG', 'RWF', 'UGX', 'VND', 'VUV', 'XAF', 'XOF', 'XPF'].includes(code)) return 0
    return 2
  }

  const customOptionPriceDeltaMajor = computed(() => (
    selectedCustomOptionPriceDeltaMinor.value / (10 ** minorUnitsForCurrency(currentCurrency.value))
  ))

  const effectivePrice = computed(() => {
    const basePrice = selectedVariant.value?.sale_price
      ?? selectedVariant.value?.price
      ?? product.value?.sale_price
      ?? product.value?.price
      ?? 0
    return Number(basePrice || 0) + customOptionPriceDeltaMajor.value
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
    if (selectedVariantDisplayPrice) {
      const snapshotCurrency = normalizeProductCurrencyCode(selectedVariantDisplayPrice.currency)
      const baseCurrency = currentCurrency.value
      const delta = snapshotCurrency === baseCurrency
        ? customOptionPriceDeltaMajor.value
        : customOptionPriceDeltaMajor.value * Number(selectedVariantDisplayPrice.rate || 0)
      return { ...selectedVariantDisplayPrice, amount: Number(selectedVariantDisplayPrice.amount || 0) + delta }
    }

    const productDisplayPrice =
      validProductDisplayPrice(product.value?.display_price)
      || displayPriceSnapshotForCurrency(product.value?.display_prices, displayCurrency.value)
    if (productDisplayPrice) {
      const snapshotCurrency = normalizeProductCurrencyCode(productDisplayPrice.currency)
      const baseCurrency = currentCurrency.value
      const delta = snapshotCurrency === baseCurrency
        ? customOptionPriceDeltaMajor.value
        : customOptionPriceDeltaMajor.value * Number(productDisplayPrice.rate || 0)
      return { ...productDisplayPrice, amount: Number(productDisplayPrice.amount || 0) + delta }
    }

    return { amount: Number(effectivePrice.value || 0), currency: currentCurrency.value }
  })

  const selectedAvailability = computed<ProductAvailability>(() => {
    if (selectedVariant.value) return selectedVariant.value.availability || 'out_of_stock'
    if (product.value && activeVariants.value.length === 0) {
      return product.value.availability || 'out_of_stock'
    }
    return 'out_of_stock'
  })

  const canAddToCart = computed(() => Boolean(
    product.value
    && Number(effectivePrice.value) > 0
    && ['in_stock', 'made_to_order'].includes(selectedAvailability.value)
    && customOptionsValid.value,
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
    customOptionDefinitions,
    variantOptionGroups,
    customOptionGroups,
    customOptionsValid,
    selectedCustomOptions,
    selectedOptions,
    selectCustomOption,
    variantChoices,
    currentVariantOptions,
    parseVariantOptions: parseProductVariantOptions,
    variantLabel,
    selectVariantOption,
    effectivePrice,
    selectedCustomOptionPriceDeltaMinor,
    currentCurrency,
    currentDisplayPrice,
    selectedAvailability,
    canAddToCart,
    formattedPrice,
  }
}
