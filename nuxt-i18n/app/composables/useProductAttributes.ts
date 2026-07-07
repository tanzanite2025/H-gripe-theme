import { ref, computed } from 'vue'

export interface AttributeValue {
  id: number
  attribute_id: number
  name: string
  slug: string
  value?: string | null
  is_enabled?: boolean
  sort_order?: number
  meta?: Record<string, any>
}

export interface AttributeWithValues {
  id: number
  name: string
  slug: string
  type: string
  is_filterable: boolean
  affects_sku: boolean
  affects_stock: boolean
  is_enabled: boolean
  sort_order: number
  meta?: Record<string, any>
  values: AttributeValue[]
}

interface ProductSpecDefinition {
  id: number
  name: string
  slug: string
  field_type: string
  is_filterable: boolean
  sort_order: number
  options?: string | null
}

interface ProductTypeSchema {
  id: number
  name: string
  slug: string
  spec_definitions?: ProductSpecDefinition[]
}

export const useProductAttributes = () => {
  const config = useRuntimeConfig()
  const apiBase = computed(() => {
    const base = (config.public as { apiBase?: string }).apiBase || '/api/v1'
    return base.replace(/\/$/, '')
  })

  const colorAttributes = ref<AttributeWithValues[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const parseOptions = (raw?: string | null): string[] => {
    if (!raw) return []
    try {
      const parsed = JSON.parse(raw)
      return Array.isArray(parsed) ? parsed.map(String) : []
    } catch {
      return []
    }
  }

  const formatLabel = (value: string) => value.replace(/_/g, ' ')

  const buildAttributesFromTypes = (productTypes: ProductTypeSchema[]): AttributeWithValues[] => {
    const bySlug = new Map<string, AttributeWithValues>()

    productTypes.forEach((productType) => {
      ;(productType.spec_definitions || [])
        .filter((spec) => spec.is_filterable)
        .forEach((spec) => {
          const options =
            spec.field_type === 'boolean'
              ? ['true', 'false']
              : spec.field_type === 'select'
                ? parseOptions(spec.options)
                : []

          if (options.length === 0) return

          const existing = bySlug.get(spec.slug)
          const attribute: AttributeWithValues = existing || {
            id: spec.id,
            name: spec.name,
            slug: spec.slug,
            type: 'select',
            is_filterable: true,
            affects_sku: false,
            affects_stock: false,
            is_enabled: true,
            sort_order: spec.sort_order || 0,
            meta: { source: 'product_type', product_type: productType.slug },
            values: [],
          }

          const existingValues = new Set(attribute.values.map((value) => value.slug))
          options.forEach((option, index) => {
            if (existingValues.has(option)) return
            attribute.values.push({
              id: spec.id * 1000 + index,
              attribute_id: spec.id,
              name: formatLabel(option),
              slug: option,
              value: option,
              sort_order: index + 1,
            })
          })

          bySlug.set(spec.slug, attribute)
        })
    })

    return [...bySlug.values()].sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0))
  }

  const loadFilterableColorAttributes = async () => {
    if (loading.value) return

    loading.value = true
    error.value = null

    try {
      const response = await $fetch<{ data: ProductTypeSchema[] }>(
        `${apiBase.value}/products/types`,
        {
          headers: { accept: 'application/json' },
        },
      )

      const items = Array.isArray(response?.data) ? response.data : []
      colorAttributes.value = buildAttributesFromTypes(items)
    } catch (e: any) {
      // eslint-disable-next-line no-console
      console.error('Failed to load product type filters:', e)
      error.value = e?.data?.message || e?.message || 'Failed to load product attributes.'
      colorAttributes.value = []
    } finally {
      loading.value = false
    }
  }

  return {
    colorAttributes,
    loading,
    error,
    loadFilterableColorAttributes,
  }
}
