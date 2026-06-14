import { ref } from 'vue'
import { useAuth } from '~/composables/useAuth'

interface WarrantyRemaining {
  months: number
  days: number
  total_days: number
  expired_days?: number
}

interface WarrantyResult {
  product_code: string
  product_type: {
    code: string
    name: string
    name_zh: string
  }
  product_name: string
  ship_date: string
  warranty_months: number
  warranty_end: string
  status: 'valid' | 'expired'
  remaining: WarrantyRemaining
  records: Array<{
    type: string
    type_name: string
    type_name_zh: string
    date: string
    description?: string
  }>
}

interface WarrantyCheckResponse {
  success: boolean
  data?: WarrantyResult
}

// Warranty check composable: shared logic for querying warranty status
export const useWarrantyCheck = () => {
  const { locale } = useI18n()
  const auth = useAuth()

  // 表单与结果状态
  const productCode = ref('')
  const searchedCode = ref('')
  const loading = ref(false)
  const error = ref(false)
  const result = ref<WarrantyResult | null>(null)

  const checkWarranty = async () => {
    if (!productCode.value.trim()) return

    loading.value = true
    error.value = false
    result.value = null
    searchedCode.value = productCode.value.trim()

    try {
      const response = await auth.request<WarrantyCheckResponse>(
        `/registrations/warranty/${encodeURIComponent(searchedCode.value)}`,
        {
          headers: { accept: 'application/json' },
        },
        'Warranty record not found'
      )

      if (response.success && response.data) {
        result.value = response.data
      } else {
        error.value = true
      }
    } catch (e: unknown) {
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
