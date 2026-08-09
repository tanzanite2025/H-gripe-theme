import { computed, onMounted, reactive } from 'vue'
import { toast } from 'vue-sonner'
import { faqAdminApi } from '@/api/faq'
import { useFaqEditor } from '@/composables/faq/useFaqEditor'
import { useFaqList } from '@/composables/faq/useFaqList'
import { useFaqStructure } from '@/composables/faq/useFaqStructure'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import {
  buildStructureLocaleOptions,
  domainName,
  FAQ_STATUS_FILTER_OPTIONS,
  formatDate,
  localeName,
  plainTextFromHTML,
  statusName,
  statusTone,
  visibilityName,
  visibilityTone
} from '@/lib/faqAdminPresentation'
import type { FAQCategory, FAQID, FAQItemLike } from '@/lib/faqAdminPresentation'
import { useAuthStore } from '@/stores/auth'

type ConfirmationType = '' | 'delete' | 'batch-delete' | 'category-delete'
type ConfirmationTarget = FAQItemLike | FAQItemLike[] | FAQCategory | null

interface ConfirmationState {
  open: boolean
  type: ConfirmationType
  target: ConfirmationTarget
  title: string
  description: string
  confirmLabel: string
}

export function useFaqAdmin() {
  const authStore = useAuthStore()
  const hasPermission = (permission: string): boolean => authStore.hasPermission(permission)
  const confirmation = reactive<ConfirmationState>({
    open: false,
    type: '',
    target: null,
    title: '',
    description: '',
    confirmLabel: '删除'
  })

  let refreshFAQs: () => Promise<unknown> = async () => {}

  const supportedLanguages = useSupportedLanguages()
  const structureLocales = computed(() => buildStructureLocaleOptions(supportedLanguages.enabledLanguages.value))
  const displayLocaleName = (locale?: string | null): string => localeName(locale, supportedLanguages.enabledLanguages.value)

  const structure = useFaqStructure({
    languages: supportedLanguages.enabledLanguages,
    defaultLocale: supportedLanguages.defaultLocale,
    onChanged: () => refreshFAQs()
  })
  const list = useFaqList({
    faqStructures: structure.faqStructures,
    activeStructureLocale: structure.activeStructureLocale
  })
  const editor = useFaqEditor({
    faqStructures: structure.faqStructures,
    activeStructureLocale: structure.activeStructureLocale,
    defaultLocale: supportedLanguages.defaultLocale,
    onChanged: () => refreshFAQs()
  })

  refreshFAQs = () => Promise.all([
    list.fetchFAQs(),
    structure.refreshFAQStructure()
  ])

  const switchStructureLocale = async (locale: string): Promise<void> => {
    await structure.switchStructureLocale(locale)
    await list.setLocale(locale)
  }

  const requestDelete = (faq: FAQItemLike): void => {
    Object.assign(confirmation, {
      open: true,
      type: 'delete',
      target: faq,
      title: '删除 FAQ？',
      description: `问题“${faq.question || ''}”将被永久删除，此操作不可恢复。`,
      confirmLabel: '删除'
    })
  }

  const requestBatchDelete = (): void => {
    Object.assign(confirmation, {
      open: true,
      type: 'batch-delete',
      target: [...list.selectedFAQs.value],
      title: '批量删除 FAQ？',
      description: `${list.selectedFAQs.value.length} 个 FAQ 将被永久删除，此操作不可恢复。`,
      confirmLabel: '批量删除'
    })
  }

  const requestDeleteCategory = (category: FAQCategory): void => {
    Object.assign(confirmation, {
      open: true,
      type: 'category-delete',
      target: category,
      title: '删除 FAQ 分类？',
      description: Number(category.faq_count || 0) > 0
        ? `分类“${category.name || ''}”下还有 ${category.faq_count} 条 FAQ。请先移动或删除内容，再删除分类。`
        : `分类“${category.name || ''}”将从前端 FAQ 分类结构中删除，此操作不可恢复。`,
      confirmLabel: '删除分类'
    })
  }

  const executeConfirmedAction = async (): Promise<void> => {
    const { type, target } = confirmation
    confirmation.open = false
    if (!target) return

    try {
      if (type === 'delete') {
        const faq = target as FAQItemLike
        await faqAdminApi.deleteFAQ(faq.id as FAQID)
        toast.success('FAQ 已删除')
      } else if (type === 'batch-delete') {
        const faqs = Array.isArray(target) ? target : []
        const payload = await faqAdminApi.deleteFAQs(faqs.map((faq) => faq.id as FAQID)) as { deleted?: number }
        toast.success(`已删除 ${payload.deleted ?? faqs.length} 个 FAQ`)
      } else if (type === 'category-delete') {
        const category = target as FAQCategory
        await faqAdminApi.deleteCategory(category.id as FAQID)
        toast.success('FAQ 分类已删除')
      }
      await refreshFAQs()
    } catch (error) {
      console.error('Failed to delete FAQs:', error)
    }
  }

  onMounted(async () => {
    await supportedLanguages.fetchLanguages()
    list.setLocale(structure.activeStructureLocale.value, { fetch: false })
    await refreshFAQs()
  })

  return {
    loading: list.loading,
    faqGroups: list.faqGroups,
    structureLoading: structure.structureLoading,
    activeStructureLocale: structure.activeStructureLocale,
    selectedFAQs: list.selectedFAQs,
    dialogVisible: editor.dialogVisible,
    dialogMode: editor.dialogMode,
    submitting: editor.submitting,
    placementLocked: editor.placementLocked,
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
    structureLocales,
    languageOptions: supportedLanguages.languageOptions,
    structurePageOptions: structure.structurePageOptions,
    faqPageOptions: editor.faqPageOptions,
    availableFAQCategories: editor.availableFAQCategories,
    pageFilterOptions: list.pageFilterOptions,
    categoryFilterOptions: list.categoryFilterOptions,
    hasPermission,
    localeName: displayLocaleName,
    statusName,
    statusTone,
    visibilityName,
    visibilityTone,
    domainName,
    formatDate,
    plainTextFromHTML,
    clearFieldError: editor.clearFieldError,
    updateFAQAnswer: editor.updateFAQAnswer,
    switchStructureLocale,
    applyFilters: list.applyFilters,
    resetFilters: list.resetFilters,
    showCreateDialog: editor.showCreateDialog,
    showEditDialog: editor.showEditDialog,
    submitForm: editor.submitForm,
    showPageDialog: structure.showPageDialog,
    submitPageForm: structure.submitPageForm,
    showCategoryDialog: structure.showCategoryDialog,
    submitCategoryForm: structure.submitCategoryForm,
    isSelected: list.isSelected,
    toggleFAQ: list.toggleFAQ,
    requestDelete,
    requestBatchDelete,
    requestDeleteCategory,
    executeConfirmedAction
  }
}
