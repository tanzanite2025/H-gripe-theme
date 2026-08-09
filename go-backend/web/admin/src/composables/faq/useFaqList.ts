import { computed, reactive, ref, watch } from 'vue'
import type { ComputedRef, Ref } from 'vue'
import { faqAdminApi } from '@/api/faq'
import {
  buildCategoryFilterOptions,
  buildPageFilterOptions
} from '@/lib/faqAdminPresentation'
import type { FAQID, FAQItemLike, FAQStructureMap, FAQStructurePage } from '@/lib/faqAdminPresentation'

interface UseFaqListOptions {
  faqStructures: FAQStructureMap
  activeStructureLocale?: Ref<string> | ComputedRef<string>
}

export interface FAQFilters {
  search: string
  page_id: string
  category: string
  status: string
  locale: string
}

export interface FAQPagination {
  total: number
}

interface FAQGroupsPayload {
  pages?: FAQStructurePage[]
  total?: number
}

export function useFaqList({ faqStructures, activeStructureLocale }: UseFaqListOptions) {
  const loading = ref(false)
  const faqs = ref<FAQItemLike[]>([])
  const faqGroups = ref<FAQStructurePage[]>([])
  const selectedFAQs = ref<FAQItemLike[]>([])
  const filters = reactive<FAQFilters>({
    search: '',
    page_id: 'all',
    category: 'all',
    status: 'all',
    locale: activeStructureLocale?.value || 'all'
  })
  const pagination = reactive<FAQPagination>({ total: 0 })

  const pageFilterOptions = computed(() => buildPageFilterOptions(faqStructures, filters.locale))
  const categoryFilterOptions = computed(() => (
    buildCategoryFilterOptions(faqStructures, filters.locale, filters.page_id)
  ))
  const categoryOptionsForFilter = () => (
    buildCategoryFilterOptions(faqStructures, filters.locale, filters.page_id)
      .filter((option) => option.value !== 'all')
  )

  const buildFilterParams = (): Record<string, string> => ({
    ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
    ...(filters.page_id !== 'all' ? { page_id: filters.page_id } : {}),
    ...(filters.category !== 'all' ? { category: filters.category } : {}),
    ...(filters.status !== 'all' ? { status: filters.status } : {}),
    ...(filters.locale !== 'all' ? { locale: filters.locale } : {})
  })

  const fetchFAQs = async (): Promise<void> => {
    loading.value = true
    try {
      const payload = await faqAdminApi.listFAQGroups(buildFilterParams()) as FAQGroupsPayload
      faqGroups.value = payload.pages || []
      faqs.value = faqGroups.value.flatMap((page) => (
        (page.categories || []).flatMap((category) => category.faqs || [])
      ))
      pagination.total = payload.total ?? faqs.value.length
      selectedFAQs.value = []
    } catch (error) {
      console.error('Failed to fetch FAQs:', error)
    } finally {
      loading.value = false
    }
  }

  const applyFilters = (): void => {
    fetchFAQs()
  }

  const resetFilters = (): void => {
    Object.assign(filters, {
      search: '',
      page_id: 'all',
      category: 'all',
      status: 'all',
      locale: filters.locale
    })
    fetchFAQs()
  }

  const setLocale = (locale: string, { fetch = true }: { fetch?: boolean } = {}): Promise<void> => {
    if (!locale || filters.locale === locale) {
      return fetch ? fetchFAQs() : Promise.resolve()
    }
    filters.locale = locale
    filters.page_id = 'all'
    filters.category = 'all'
    return fetch ? fetchFAQs() : Promise.resolve()
  }

  const isSelected = (faqID?: FAQID | null): boolean => selectedFAQs.value.some((faq) => faq.id === faqID)

  const toggleFAQ = (faq: FAQItemLike, checked: boolean | string): void => {
    if (checked === true && !isSelected(faq.id)) {
      selectedFAQs.value = [...selectedFAQs.value, faq]
    } else if (checked !== true) {
      selectedFAQs.value = selectedFAQs.value.filter((selected) => selected.id !== faq.id)
    }
  }

  watch(() => [filters.page_id, filters.locale], () => {
    const validValues = categoryOptionsForFilter().map((category) => category.value)
    if (filters.category !== 'all' && !validValues.includes(filters.category)) {
      filters.category = 'all'
    }
  })

  watch(activeStructureLocale, (locale) => {
    if (locale && filters.locale === 'all') {
      filters.locale = locale
    }
  })

  return {
    loading,
    faqs,
    faqGroups,
    selectedFAQs,
    filters,
    pagination,
    pageFilterOptions,
    categoryFilterOptions,
    fetchFAQs,
    setLocale,
    applyFilters,
    resetFilters,
    isSelected,
    toggleFAQ
  }
}
