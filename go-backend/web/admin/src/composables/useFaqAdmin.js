import { onMounted, reactive } from 'vue'
import { toast } from 'vue-sonner'
import { faqAdminApi } from '@/api/faq'
import { useFaqEditor } from '@/composables/faq/useFaqEditor'
import { useFaqList } from '@/composables/faq/useFaqList'
import { useFaqStructure } from '@/composables/faq/useFaqStructure'
import {
  domainName,
  FAQ_LOCALE_FILTER_OPTIONS,
  FAQ_STATUS_FILTER_OPTIONS,
  FAQ_STRUCTURE_LOCALES,
  formatDate,
  localeName,
  plainTextFromHTML,
  statusName,
  statusTone,
  visibilityName,
  visibilityTone
} from '@/lib/faqAdminPresentation'
import { useAuthStore } from '@/stores/auth'

export function useFaqAdmin() {
  const authStore = useAuthStore()
  const hasPermission = (permission) => authStore.hasPermission(permission)
  const confirmation = reactive({
    open: false,
    type: '',
    target: null,
    title: '',
    description: '',
    confirmLabel: '删除'
  })

  let refreshFAQs = async () => {}

  const structure = useFaqStructure({ onChanged: () => refreshFAQs() })
  const list = useFaqList({
    faqStructures: structure.faqStructures,
    allStructurePages: structure.allStructurePages
  })
  const editor = useFaqEditor({
    faqStructures: structure.faqStructures,
    onChanged: () => refreshFAQs()
  })

  refreshFAQs = () => Promise.all([
    list.fetchFAQs(),
    structure.refreshFAQStructure()
  ])

  const requestDelete = (faq) => Object.assign(confirmation, {
    open: true,
    type: 'delete',
    target: faq,
    title: '删除 FAQ？',
    description: `问题“${faq.question}”将被永久删除，此操作不可恢复。`,
    confirmLabel: '删除'
  })

  const requestBatchDelete = () => Object.assign(confirmation, {
    open: true,
    type: 'batch-delete',
    target: [...list.selectedFAQs.value],
    title: '批量删除 FAQ？',
    description: `${list.selectedFAQs.value.length} 个 FAQ 将被永久删除，此操作不可恢复。`,
    confirmLabel: '批量删除'
  })

  const requestDeleteCategory = (category) => Object.assign(confirmation, {
    open: true,
    type: 'category-delete',
    target: category,
    title: '删除 FAQ 分类？',
    description: category.faq_count > 0
      ? `分类“${category.name}”下还有 ${category.faq_count} 条 FAQ。请先移动或删除内容，再删除分类。`
      : `分类“${category.name}”将从前端 FAQ 分类结构中删除，此操作不可恢复。`,
    confirmLabel: '删除分类'
  })

  const executeConfirmedAction = async () => {
    const { type, target } = confirmation
    confirmation.open = false

    try {
      if (type === 'delete') {
        await faqAdminApi.deleteFAQ(target.id)
        toast.success('FAQ 已删除')
      } else if (type === 'batch-delete') {
        const payload = await faqAdminApi.deleteFAQs(target.map((faq) => faq.id))
        toast.success(`已删除 ${payload.deleted ?? target.length} 个 FAQ`)
      } else if (type === 'category-delete') {
        await faqAdminApi.deleteCategory(target.id)
        toast.success('FAQ 分类已删除')
      }
      await refreshFAQs()
    } catch (error) {
      console.error('Failed to delete FAQs:', error)
    }
  }

  onMounted(refreshFAQs)

  return {
    loading: list.loading,
    structureLoading: structure.structureLoading,
    faqs: list.faqs,
    activeStructureLocale: structure.activeStructureLocale,
    selectedFAQs: list.selectedFAQs,
    dialogVisible: editor.dialogVisible,
    dialogMode: editor.dialogMode,
    submitting: editor.submitting,
    pageDialogVisible: structure.pageDialogVisible,
    pageSubmitting: structure.pageSubmitting,
    categoryDialogVisible: structure.categoryDialogVisible,
    categoryDialogMode: structure.categoryDialogMode,
    categorySubmitting: structure.categorySubmitting,
    formErrors: editor.formErrors,
    filters: list.filters,
    pagination: list.pagination,
    faqForm: editor.faqForm,
    pageForm: structure.pageForm,
    categoryForm: structure.categoryForm,
    confirmation,
    statusFilterOptions: FAQ_STATUS_FILTER_OPTIONS,
    localeFilterOptions: FAQ_LOCALE_FILTER_OPTIONS,
    structureLocales: FAQ_STRUCTURE_LOCALES,
    faqStructure: structure.faqStructure,
    structurePageOptions: structure.structurePageOptions,
    faqPageOptions: editor.faqPageOptions,
    availableFAQCategories: editor.availableFAQCategories,
    pageFilterOptions: list.pageFilterOptions,
    categoryFilterOptions: list.categoryFilterOptions,
    selectionState: list.selectionState,
    hasPermission,
    localeName,
    statusName,
    statusTone,
    visibilityName,
    visibilityTone,
    domainName,
    formatDate,
    plainTextFromHTML,
    pageTitleForFAQ: list.pageTitle,
    categoryLabelForFAQ: list.categoryLabel,
    clearFieldError: editor.clearFieldError,
    updateFAQAnswer: editor.updateFAQAnswer,
    switchStructureLocale: structure.switchStructureLocale,
    applyFilters: list.applyFilters,
    resetFilters: list.resetFilters,
    updatePage: list.updatePage,
    updatePageSize: list.updatePageSize,
    showCreateDialog: editor.showCreateDialog,
    showEditDialog: editor.showEditDialog,
    submitForm: editor.submitForm,
    showPageDialog: structure.showPageDialog,
    submitPageForm: structure.submitPageForm,
    showCategoryDialog: structure.showCategoryDialog,
    submitCategoryForm: structure.submitCategoryForm,
    isSelected: list.isSelected,
    toggleAllFAQs: list.toggleAllFAQs,
    toggleFAQ: list.toggleFAQ,
    requestDelete,
    requestBatchDelete,
    requestDeleteCategory,
    executeConfirmedAction
  }
}
