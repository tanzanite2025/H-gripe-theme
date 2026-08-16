import { reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import productApi from '@/api/products'
import { customsSummaryApi } from '@/api/customsClassifications'
import type { CustomsProductFilters } from '@/components/admin/customs/customsClassificationTypes'
import { useCustomsEditorResources } from './useCustomsEditorResources'
import { useCustomsLookup } from './useCustomsLookup'
import { useCustomsTemplates } from './useCustomsTemplates'

export const useCustomsClassificationCenter = (options: Record<string, any> = {}) => {
  const resolveLocale = () => String(options.locale?.value || options.locale || 'en').trim() || 'en'

  const productRows = ref<Record<string, any>[]>([])
  const productLoading = ref(false)
  const productPage = ref(1)
  const productPageSize = ref(20)
  const productTotal = ref(0)
  const productFilters = reactive<CustomsProductFilters>({
    search: '',
    product_specification_template_id: 'all',
    customs_status: 'incomplete',
  })

  const catalogTotal = ref(0)
  const incompleteTotal = ref(0)
  const completeTotal = ref(0)
  const missingHSTotal = ref(0)

  const {
    templates,
    templateLoading,
    templateDialogOpen,
    templateSaving,
    templateForm,
    activeTemplates,
    fetchTemplates,
    openTemplateCreate,
    openTemplateEdit,
    templateProductSpecTemplateValue,
    setTemplateProductSpecTemplate,
    saveTemplate,
    removeTemplate,
    openTemplateFromCandidate,
  } = useCustomsTemplates()

  const {
    lookupProvider,
    lookupQuery,
    lookupLoading,
    lookupCompleted,
    lookupCandidates,
    runLookup,
  } = useCustomsLookup()

  const {
    brands,
    shippingTemplates,
    informationTemplates,
    fetchEditorResources,
  } = useCustomsEditorResources()

  const fetchProductRows = async () => {
    productLoading.value = true
    try {
      const params = {
        page: productPage.value,
        page_size: productPageSize.value,
        locale: resolveLocale(),
        ...(productFilters.search.trim() ? { search: productFilters.search.trim() } : {}),
        ...(productFilters.product_specification_template_id !== 'all' ? { product_specification_template_id: productFilters.product_specification_template_id } : {}),
        ...(productFilters.customs_status !== 'all' ? { customs_status: productFilters.customs_status } : {}),
      }
      const payload = await productApi.list(params)
      productRows.value = payload.products || []
      productTotal.value = payload.pagination?.total || 0
    } catch (error) {
      console.error('Failed to fetch customs product rows:', error)
      toast.error('商品清关状态加载失败')
    } finally {
      productLoading.value = false
    }
  }

  const fetchStats = async () => {
    try {
      const summary = await customsSummaryApi.get(resolveLocale())
      catalogTotal.value = summary.total || 0
      incompleteTotal.value = summary.incomplete || 0
      completeTotal.value = summary.complete || 0
      missingHSTotal.value = summary.missing_hs_code || 0
    } catch (error) {
      console.error('Failed to fetch customs stats:', error)
      toast.error('清关统计加载失败')
    }
  }

  const refreshWorkbench = async () => {
    await Promise.all([fetchProductRows(), fetchStats(), fetchTemplates()])
  }

  const applyProductFilters = () => {
    productPage.value = 1
    void fetchProductRows()
  }

  const updateProductPage = (page: number) => {
    productPage.value = page
    void fetchProductRows()
  }

  const updateProductPageSize = (pageSize: number) => {
    productPageSize.value = pageSize
    productPage.value = 1
    void fetchProductRows()
  }

  return {
    productRows,
    productLoading,
    productPage,
    productPageSize,
    productTotal,
    productFilters,
    templates,
    templateLoading,
    templateDialogOpen,
    templateSaving,
    templateForm,
    lookupProvider,
    lookupQuery,
    lookupLoading,
    lookupCompleted,
    lookupCandidates,
    catalogTotal,
    incompleteTotal,
    completeTotal,
    missingHSTotal,
    brands,
    shippingTemplates,
    informationTemplates,
    activeTemplates,
    fetchProductRows,
    fetchStats,
    fetchTemplates,
    fetchEditorResources,
    refreshWorkbench,
    applyProductFilters,
    updateProductPage,
    updateProductPageSize,
    openTemplateCreate,
    openTemplateEdit,
    templateProductSpecTemplateValue,
    setTemplateProductSpecTemplate,
    saveTemplate,
    removeTemplate,
    runLookup,
    openTemplateFromCandidate,
  }
}

export default useCustomsClassificationCenter
