import {
  computed,
  toValue,
  type MaybeRefOrGetter,
} from 'vue'
import type {
  GoProduct,
  ProductBreadcrumbItem,
  ProductSpecificationGroup,
} from '~/types/productDetail'
import {
  formatProductSpecValue,
  PRODUCT_DETAIL_HIDDEN_SPEC_SLUGS,
} from '~/utils/productDetail'

export interface ProductDetailPresentationOptions {
  product: MaybeRefOrGetter<GoProduct | null | undefined>
}

export function useProductDetailPresentation(options: ProductDetailPresentationOptions) {
  const product = computed(() => toValue(options.product) || null)

  const specGroups = computed<ProductSpecificationGroup[]>(() => {
    const groups = new Map<string, Array<{
      slug: string
      name: string
      displayValue: string
    }>>()

    ;(product.value?.spec_values || []).forEach((item) => {
      const definition = item.definition
      if (!definition || definition.is_visible === false) return
      if (PRODUCT_DETAIL_HIDDEN_SPEC_SLUGS.has(String(definition.slug || '').trim().toLowerCase())) {
        return
      }

      const displayValue = formatProductSpecValue(item)
      if (!displayValue) return

      const groupName = definition.group || 'Specifications'
      const current = groups.get(groupName) || []
      current.push({
        slug: definition.slug,
        name: definition.name,
        displayValue,
      })
      groups.set(groupName, current)
    })

    return [...groups.entries()].map(([name, items]) => ({ name, items }))
  })

  const productBreadcrumbItems = computed<ProductBreadcrumbItem[]>(() => (
    (product.value?.breadcrumb?.items || [])
      .map((item) => ({
        type: String(item?.type || '').trim(),
        id: Number.isFinite(Number(item?.id)) ? Number(item.id) : undefined,
        name: String(item?.name || '').trim(),
        slug: String(item?.slug || '').trim() || undefined,
        path: String(item?.path || '').trim(),
      }))
      .filter((item) => item.name)
  ))

  return {
    specGroups,
    productBreadcrumbItems,
  }
}
