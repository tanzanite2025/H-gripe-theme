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

export const useProductAttributes = () => {
  const config = useRuntimeConfig()
  const apiBase = computed(() => {
    const base = (config.public as { apiBase?: string }).apiBase || '/api/v1'
    return base.replace(/\/$/, '')
  })

  const colorAttributes = ref<AttributeWithValues[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // 本地开发环境下，当 WordPress REST 路由不可用时的兜底配置（用于让高级筛选 UI 先展示出来）
  // 与后台 Attributes 页面当前配置保持一致：Color / Diameter / Brake
  const fallbackAttributes: AttributeWithValues[] = [
    {
      id: 1,
      name: 'Color',
      slug: 'color',
      type: 'color',
      is_filterable: true,
      affects_sku: true,
      affects_stock: false,
      is_enabled: true,
      sort_order: 10,
      meta: {},
      values: [
        { id: 11, attribute_id: 1, name: 'Sliver', slug: 'sliver', value: '#C0C0C0', sort_order: 10 },
        { id: 12, attribute_id: 1, name: 'Black', slug: 'black', value: '#000000', sort_order: 20 },
        { id: 13, attribute_id: 1, name: 'Red', slug: 'red', value: '#FF0000', sort_order: 30 },
        { id: 14, attribute_id: 1, name: 'Blue', slug: 'blue', value: '#0000FF', sort_order: 40 },
        { id: 15, attribute_id: 1, name: 'Green', slug: 'green', value: '#00FF00', sort_order: 50 },
        { id: 16, attribute_id: 1, name: 'Yellow', slug: 'yellow', value: '#FFFF00', sort_order: 60 },
      ],
    },
    {
      id: 2,
      name: 'Diameter',
      slug: 'diameter',
      type: 'select',
      is_filterable: true,
      affects_sku: true,
      affects_stock: false,
      is_enabled: true,
      sort_order: 20,
      meta: {},
      values: [
        { id: 21, attribute_id: 2, name: '700C', slug: '700C', sort_order: 10 },
        { id: 22, attribute_id: 2, name: '12inch', slug: '12inch', sort_order: 20 },
        { id: 23, attribute_id: 2, name: '14inch', slug: '14inch', sort_order: 30 },
        { id: 24, attribute_id: 2, name: '20inch', slug: '20inch', sort_order: 40 },
        { id: 25, attribute_id: 2, name: '451/22inch', slug: '451/22inch', sort_order: 50 },
        { id: 26, attribute_id: 2, name: '26inch', slug: '26inch', sort_order: 60 },
        { id: 27, attribute_id: 2, name: '27.5inch/650B', slug: '27.5inch/650B', sort_order: 70 },
        { id: 28, attribute_id: 2, name: '29inch', slug: '29inch', sort_order: 80 },
      ],
    },
    {
      id: 3,
      name: 'Brake',
      slug: 'brake',
      type: 'select',
      is_filterable: true,
      affects_sku: true,
      affects_stock: false,
      is_enabled: true,
      sort_order: 30,
      meta: {},
      values: [
        { id: 31, attribute_id: 3, name: 'V-Brake', slug: 'v-brake', sort_order: 10 },
        { id: 32, attribute_id: 3, name: 'Disc-Brake', slug: 'disc-brake', sort_order: 20 },
      ],
    },
  ]

  const loadFilterableColorAttributes = async () => {
    if (loading.value) return

    loading.value = true
    error.value = null

    try {
      const response = await $fetch<{ data: AttributeWithValues[] }>(
        `${apiBase.value}/products/attributes/filterable`,
        {
          headers: { accept: 'application/json' },
        },
      )

      const items = Array.isArray(response?.data) ? response.data : []
      colorAttributes.value = items
    } catch (e: any) {
      // eslint-disable-next-line no-console
      console.error('Failed to load filterable color attributes:', e)
      error.value = e?.data?.message || e?.message || 'Failed to load product attributes.'
      // 当后端 REST 路由不可用（如 DEV 环境返回 rest_no_route）时，使用本地兜底配置，确保 UI 仍然能展示筛选块
      colorAttributes.value = fallbackAttributes
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
