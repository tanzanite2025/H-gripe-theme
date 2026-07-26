import { computed, reactive, ref, watch } from 'vue'
import { faqAdminApi } from '@/api/faq'
import {
  buildCategoryFilterOptions,
  buildPageFilterOptions
} from '@/lib/faqAdminPresentation'

export function useFaqList({ faqStructures, activeStructureLocale }) {
  const loading = ref(false)
  const faqs = ref([])
  const faqGroups = ref([])
  const selectedFAQs = ref([])
  const filters = reactive({
    search: '',
    page_id: 'all',
    category: 'all',
    status: 'all',
    locale: activeStructureLocale?.value || 'all'
  })
  const pagination = reactive({ total: 0 })

  const pageFilterOptions = computed(() => buildPageFilterOptions(faqStructures, filters.locale))
  const categoryFilterOptions = computed(() => (
    buildCategoryFilterOptions(faqStructures, filters.locale, filters.page_id)
  ))
  const categoryOptionsForFilter = () => (
    buildCategoryFilterOptions(faqStructures, filters.locale, filters.page_id)
      .filter((option) => option.value !== 'all')
  )

  const buildFilterParams = () => ({
    ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
    ...(filters.page_id !== 'all' ? { page_id: filters.page_id } : {}),
    ...(filters.category !== 'all' ? { category: filters.category } : {}),
    ...(filters.status !== 'all' ? { status: filters.status } : {}),
    ...(filters.locale !== 'all' ? { locale: filters.locale } : {})
  })

  const fetchFAQs = async () => {
    loading.value = true
    try {
      const payload = await faqAdminApi.listFAQGroups(buildFilterParams())
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

  const applyFilters = () => {
    fetchFAQs()
  }

  const resetFilters = () => {
    Object.assign(filters, {
      search: '',
      page_id: 'all',
      category: 'all',
      status: 'all',
      locale: filters.locale
    })
    fetchFAQs()
  }

  const setLocale = (locale, { fetch = true } = {}) => {
    if (!locale || filters.locale === locale) {
      return fetch ? fetchFAQs() : Promise.resolve()
    }
    filters.locale = locale
    filters.page_id = 'all'
    filters.category = 'all'
    return fetch ? fetchFAQs() : Promise.resolve()
  }

  const isSelected = (faqID) => selectedFAQs.value.some((faq) => faq.id === faqID)

  const toggleFAQ = (faq, checked) => {
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
