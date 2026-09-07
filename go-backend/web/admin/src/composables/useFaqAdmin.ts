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
import type { FAQID, FAQItemLike } from '@/lib/faqAdminPresentation'
import { useAuthStore } from '@/stores/auth'

type ConfirmationType = '' | 'delete' | 'batch-delete'
type ConfirmationTarget = FAQItemLike | FAQItemLike[] | null

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
    formErrors: editor.formErrors,
    filters: list.filters,
    pagination: list.pagination,
    faqForm: editor.faqForm,
    pageForm: structure.pageForm,
    confirmation,
    statusFilterOptions: FAQ_STATUS_FILTER_OPTIONS,
    structureLocales,
    languageOptions: supportedLanguages.languageOptions,
    structurePageOptions: structure.structurePageOptions,
    faqPageOptions: editor.faqPageOptions,
    pageFilterOptions: list.pageFilterOptions,
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
    isSelected: list.isSelected,
    toggleFAQ: list.toggleFAQ,
    requestDelete,
    requestBatchDelete,
    executeConfirmedAction
  }
}
