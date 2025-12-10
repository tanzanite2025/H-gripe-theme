import { ref } from 'vue'

// Warranty check composable: shared logic for querying warranty status
export const useWarrantyCheck = () => {
  const { locale } = useI18n()
  const config = useRuntimeConfig()

  const apiBase = config.public?.apiBase || ''

  // 表单与结果状态
  const productCode = ref('')
  const searchedCode = ref('')
  const loading = ref(false)
  const error = ref(false)
  const result = ref<any | null>(null)

  const getWpNonce = (): string => {
    const nonce = config.public?.wpNonce
    if (typeof nonce === 'string') {
      return nonce
    }
    return ''
  }

  const checkWarranty = async () => {
    if (!productCode.value.trim()) return

    if (!apiBase) {
      // 没有配置 API 基础地址时，直接标记为错误但不中断页面
      console.error('Missing runtimeConfig.public.apiBase for warranty check')
      error.value = true
      return
    }

    loading.value = true
    error.value = false
    result.value = null
    searchedCode.value = productCode.value.trim()

    try {
      const response = await $fetch(
        `${apiBase}/wp-json/tanzanite/v1/warranty/${encodeURIComponent(searchedCode.value)}`,
        {
          credentials: 'include',
          headers: {
            'X-WP-Nonce': getWpNonce(),
          },
        }
      )

      if (response && (response as any).success) {
        result.value = (response as any).data
      } else {
        error.value = true
      }
    } catch (e: any) {
      // eslint-disable-next-line no-console
      console.error('Warranty check error:', e)
      error.value = true
    } finally {
      loading.value = false
    }
  }

  const reset = () => {
    productCode.value = ''
    searchedCode.value = ''
    error.value = false
    result.value = null
  }

  const formatDate = (dateStr: string): string => {
    if (!dateStr) return '-'
    const [year, month] = dateStr.split('-')

    if (String(locale.value).startsWith('zh')) {
      return `${year}年${month}月`
    }

    const monthNames = [
      'January',
      'February',
      'March',
      'April',
      'May',
      'June',
      'July',
      'August',
      'September',
      'October',
      'November',
      'December',
    ]

    const index = parseInt(month, 10) - 1
    return `${monthNames[index] ?? month} ${year}`
  }

  return {
    productCode,
    searchedCode,
    loading,
    error,
    result,
    checkWarranty,
    reset,
    formatDate,
  }
}
