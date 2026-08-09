import { computed, reactive, ref, watch } from 'vue'
import type { ComputedRef, Ref } from 'vue'
import { toast } from 'vue-sonner'
import { faqAdminApi } from '@/api/faq'
import { buildFAQPageOptions, findAvailableFAQCategories } from '@/lib/faqAdminPresentation'
import type { FAQCategory, FAQID, FAQItemLike, FAQStructureMap, FAQStructurePage } from '@/lib/faqAdminPresentation'

interface UseFaqEditorOptions {
  faqStructures: FAQStructureMap
  activeStructureLocale?: Ref<string> | ComputedRef<string>
  defaultLocale?: Ref<string> | ComputedRef<string>
  onChanged: () => Promise<unknown> | unknown
}

export type FAQDialogMode = 'create' | 'edit'

export interface FAQPlacement {
  page?: FAQStructurePage | null
  category?: FAQCategory | null
}

export interface FAQForm {
  id: FAQID | null
  question: string
  answer: string
  answer_image_url: string
  answer_image_alt: string
  answer_image_width: number
  answer_image_height: number
  category: string
  page_id: string
  locale: string
  status: string
  order: number
}

export interface FAQPayload {
  question: string
  answer: string
  answer_image_url: string
  answer_image_alt: string
  answer_image_width: number
  answer_image_height: number
  category: string
  page_id: string
  locale: string
  status: string
  order: number
}

export type FAQFormErrors = Partial<Record<keyof FAQForm, string>>

export function useFaqEditor({ faqStructures, activeStructureLocale, defaultLocale, onChanged }: UseFaqEditorOptions) {
  const dialogVisible = ref(false)
  const dialogMode = ref<FAQDialogMode>('create')
  const submitting = ref(false)
  const lockedEditLocale = ref('')
  const placementLocked = ref(false)
  const formErrors = reactive<FAQFormErrors>({})
  const resolveDefaultLocale = (): string => activeStructureLocale?.value || defaultLocale?.value || ''
  const faqForm = reactive<FAQForm>({
    id: null,
    question: '',
    answer: '',
    answer_image_url: '',
    answer_image_alt: '',
    answer_image_width: 0,
    answer_image_height: 0,
    category: '',
    page_id: '',
    locale: resolveDefaultLocale(),
    status: 'published',
    order: 0
  })

  const faqPageOptions = computed(() => buildFAQPageOptions(faqStructures, faqForm.locale))
  const availableFAQCategories = computed(() => (
    findAvailableFAQCategories(faqStructures, faqForm.locale, faqForm.page_id)
  ))

  const clearFormErrors = (): void => Object.keys(formErrors).forEach((key) => delete formErrors[key as keyof FAQForm])
  const clearFieldError = (field: keyof FAQForm): void => { delete formErrors[field] }
  const updateFAQAnswer = (value: string): void => {
    faqForm.answer = value
    clearFieldError('answer')
  }

  const buildFAQPayload = (): FAQPayload => ({
    question: faqForm.question.trim(),
    answer: faqForm.answer.trim(),
    answer_image_url: faqForm.answer_image_url.trim(),
    answer_image_alt: faqForm.answer_image_alt.trim(),
    answer_image_width: faqForm.answer_image_url ? 800 : 0,
    answer_image_height: faqForm.answer_image_url ? 800 : 0,
    category: faqForm.category.trim(),
    page_id: faqForm.page_id.trim(),
    locale: dialogMode.value === 'edit' ? lockedEditLocale.value : faqForm.locale,
    status: faqForm.status,
    order: Math.max(0, Number(faqForm.order || 0))
  })

  const validateForm = (payload: FAQPayload): boolean => {
    clearFormErrors()
    if (!payload.question) formErrors.question = '请输入问题'
    if (!payload.answer) formErrors.answer = '请输入答案'
    if (payload.answer_image_url && !payload.answer_image_alt) formErrors.answer = 'FAQ 图片需要填写替代文本'
    if (!payload.locale) formErrors.locale = '请选择语言'
    if (!payload.page_id) formErrors.page_id = '请选择页面'
    if (!payload.category) formErrors.category = '请输入分类'
    if (Object.keys(formErrors).length > 0) {
      toast.error('请检查 FAQ 表单中的必填项')
      return false
    }
    return true
  }

  const resetForm = (): void => {
    Object.assign(faqForm, {
      id: null,
      question: '',
      answer: '',
      answer_image_url: '',
      answer_image_alt: '',
      answer_image_width: 0,
      answer_image_height: 0,
      category: '',
      page_id: '',
      locale: resolveDefaultLocale(),
      status: 'published',
      order: 0
    })
    lockedEditLocale.value = ''
    placementLocked.value = false
    clearFormErrors()
  }

  const ensureFAQPageSelection = (): void => {
    const pages = faqStructures[faqForm.locale] || []
    if (pages.length === 0) return
    if (!pages.some((page) => page.page_id === faqForm.page_id)) {
      faqForm.page_id = pages[0].page_id
    }
    clearFieldError('page_id')
  }

  const ensureFAQCategorySelection = (): void => {
    const categoriesForPage = availableFAQCategories.value
    if (categoriesForPage.length === 0) {
      faqForm.category = ''
      return
    }
    if (!categoriesForPage.some((category) => category.category_key === faqForm.category)) {
      faqForm.category = categoriesForPage[0].category_key
    }
    clearFieldError('category')
  }

  const ensureFAQSelection = (): void => {
    ensureFAQPageSelection()
    ensureFAQCategorySelection()
  }

  const showCreateDialog = (placement: FAQPlacement | null = null): void => {
    dialogMode.value = 'create'
    resetForm()
    if (placement?.page && placement?.category) {
      faqForm.locale = placement.category.locale || placement.page.locale || resolveDefaultLocale()
      faqForm.page_id = placement.category.page_id || placement.page.page_id || ''
      faqForm.category = placement.category.category_key || ''
      faqForm.order = ((placement.category.faqs || []).length + 1) * 10
      placementLocked.value = true
    } else {
      ensureFAQSelection()
    }
    dialogVisible.value = true
  }

  const showEditDialog = async (faq: FAQItemLike): Promise<void> => {
    dialogMode.value = 'edit'
    lockedEditLocale.value = ''
    placementLocked.value = false
    try {
      const payload = await faqAdminApi.getFAQ(faq.id as FAQID) as { faq?: FAQItemLike }
      const detail = payload.faq || faq
      const locale = detail.locale || resolveDefaultLocale()
      lockedEditLocale.value = locale
      Object.assign(faqForm, {
        id: detail.id,
        question: detail.question || '',
        answer: detail.answer || '',
        answer_image_url: detail.answer_image_url || '',
        answer_image_alt: detail.answer_image_alt || '',
        answer_image_width: Number(detail.answer_image_width || 0),
        answer_image_height: Number(detail.answer_image_height || 0),
        category: detail.category || '',
        page_id: detail.page_id || '',
        locale,
        status: detail.status || 'published',
        order: Number(detail.order ?? detail.sort_order ?? 0)
      })
      if (!faqForm.page_id || !faqForm.category) ensureFAQSelection()
      clearFormErrors()
      dialogVisible.value = true
    } catch (error) {
      console.error('Failed to fetch FAQ detail:', error)
    }
  }

  const submitForm = async (): Promise<void> => {
    const payload = buildFAQPayload()
    if (!validateForm(payload)) return

    submitting.value = true
    try {
      if (dialogMode.value === 'create') {
        await faqAdminApi.createFAQ(payload)
        toast.success('FAQ 创建成功')
      } else {
        await faqAdminApi.updateFAQ(faqForm.id as FAQID, payload)
        toast.success('FAQ 更新成功')
      }
      dialogVisible.value = false
      await onChanged()
    } catch (error) {
      console.error('Failed to save FAQ:', error)
    } finally {
      submitting.value = false
    }
  }

  watch(() => faqForm.locale, () => {
    if (dialogMode.value !== 'create' || placementLocked.value) return
    ensureFAQPageSelection()
    ensureFAQCategorySelection()
  })

  watch(() => faqForm.page_id, () => {
    if (placementLocked.value) return
    ensureFAQCategorySelection()
  })

  watch(resolveDefaultLocale, (locale) => {
    if (!faqForm.locale && locale) {
      faqForm.locale = locale
      ensureFAQSelection()
    }
  })

  return {
    dialogVisible,
    dialogMode,
    submitting,
    placementLocked,
    formErrors,
    faqForm,
    faqPageOptions,
    availableFAQCategories,
    clearFieldError,
    updateFAQAnswer,
    showCreateDialog,
    showEditDialog,
    submitForm
  }
}
