import { computed, reactive, ref, watch } from 'vue'
import { faqAdminApi } from '@/api/faq'
import {
  buildCategoryFilterOptions,
  buildFAQLabelMaps,
  buildPageFilterOptions,
  categoryLabelForFAQ,
  pageTitleForFAQ
} from '@/lib/faqAdminPresentation'

export function useFaqList({ faqStructures, allStructurePages }) {
  const loading = ref(false)
  const faqs = ref([])
  const selectedFAQs = ref([])
  const filters = reactive({ search: '', page_id: 'all', category: 'all', status: 'all', locale: 'all' })
  const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

  const pageFilterOptions = computed(() => buildPageFilterOptions(faqStructures, filters.locale))
  const categoryFilterOptions = computed(() => (
    buildCategoryFilterOptions(faqStructures, filters.locale, filters.page_id)
  ))
  const selectionState = computed(() => {
    if (faqs.value.length === 0 || selectedFAQs.value.length === 0) return false
    return selectedFAQs.value.length === faqs.value.length ? true : 'indeterminate'
  })
  const faqLabelMaps = computed(() => buildFAQLabelMaps(allStructurePages.value))

  const categoryOptionsForFilter = () => (
    buildCategoryFilterOptions(faqStructures, filters.locale, filters.page_id)
      .filter((option) => option.value !== 'all')
  )

  const pageTitle = (faq) => pageTitleForFAQ(faqLabelMaps.value.pageTitles, faq)
  const categoryLabel = (faq) => categoryLabelForFAQ(faqLabelMaps.value.categoryLabels, faq)

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
      const payload = await faqAdminApi.listFAQs({
        page: pagination.page,
        page_size: pagination.pageSize,
        ...buildFilterParams()
      })
      faqs.value = payload.faqs || []
      pagination.total = payload.pagination?.total ?? payload.total ?? 0
      selectedFAQs.value = []
    } catch (error) {
      console.error('Failed to fetch FAQs:', error)
    } finally {
      loading.value = false
    }
  }

  const applyFilters = () => {
    pagination.page = 1
    fetchFAQs()
  }

  const resetFilters = () => {
    Object.assign(filters, { search: '', page_id: 'all', category: 'all', status: 'all', locale: 'all' })
    pagination.page = 1
    fetchFAQs()
  }

  const updatePage = (page) => {
    pagination.page = page
    fetchFAQs()
  }

  const updatePageSize = (pageSize) => {
    pagination.pageSize = pageSize
    pagination.page = 1
    fetchFAQs()
  }

  const isSelected = (faqID) => selectedFAQs.value.some((faq) => faq.id === faqID)

  const toggleAllFAQs = (checked) => {
    selectedFAQs.value = checked === true ? [...faqs.value] : []
  }

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

  return {
    loading,
    faqs,
    selectedFAQs,
    filters,
    pagination,
    pageFilterOptions,
    categoryFilterOptions,
    selectionState,
    pageTitle,
    categoryLabel,
    fetchFAQs,
    applyFilters,
    resetFilters,
    updatePage,
    updatePageSize,
    isSelected,
    toggleAllFAQs,
    toggleFAQ
  }
}
