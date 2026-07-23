import { computed, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { faqAdminApi } from '@/api/faq'
import { buildFAQPageOptions, findAvailableFAQCategories } from '@/lib/faqAdminPresentation'

export function useFaqEditor({ faqStructures, onChanged }) {
  const dialogVisible = ref(false)
  const dialogMode = ref('create')
  const submitting = ref(false)
  const formErrors = reactive({})
  const faqForm = reactive({
    id: null,
    question: '',
    answer: '',
    answer_image_url: '',
    answer_image_alt: '',
    answer_image_width: 0,
    answer_image_height: 0,
    category: '',
    page_id: '',
    locale: 'zh',
    status: 'published',
    order: 0
  })

  const faqPageOptions = computed(() => buildFAQPageOptions(faqStructures, faqForm.locale))
  const availableFAQCategories = computed(() => (
    findAvailableFAQCategories(faqStructures, faqForm.locale, faqForm.page_id)
  ))

  const clearFormErrors = () => Object.keys(formErrors).forEach((key) => delete formErrors[key])
  const clearFieldError = (field) => { delete formErrors[field] }
  const updateFAQAnswer = (value) => {
    faqForm.answer = value
    clearFieldError('answer')
  }

  const buildFAQPayload = () => ({
    question: faqForm.question.trim(),
    answer: faqForm.answer.trim(),
    answer_image_url: faqForm.answer_image_url.trim(),
    answer_image_alt: faqForm.answer_image_alt.trim(),
    answer_image_width: faqForm.answer_image_url ? 800 : 0,
    answer_image_height: faqForm.answer_image_url ? 800 : 0,
    category: faqForm.category.trim(),
    page_id: faqForm.page_id.trim(),
    locale: faqForm.locale,
    status: faqForm.status,
    order: Math.max(0, Number(faqForm.order || 0))
  })

  const validateForm = (payload) => {
    clearFormErrors()
    if (!payload.question) formErrors.question = '请输入问题'
    if (!payload.answer) formErrors.answer = '请输入答案'
    if (payload.answer_image_url && !payload.answer_image_alt) formErrors.answer = 'FAQ 图片需要填写替代文本'
    if (!payload.page_id) formErrors.page_id = '请选择页面'
    if (!payload.category) formErrors.category = '请输入分类'
    if (Object.keys(formErrors).length > 0) {
      toast.error('请检查 FAQ 表单中的必填项')
      return false
    }
    return true
  }

  const resetForm = () => {
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
      locale: 'zh',
      status: 'published',
      order: 0
    })
    clearFormErrors()
  }

  const ensureFAQPageSelection = () => {
    const pages = faqStructures[faqForm.locale] || []
    if (pages.length === 0) return
    if (!pages.some((page) => page.page_id === faqForm.page_id)) {
      faqForm.page_id = pages[0].page_id
    }
    clearFieldError('page_id')
  }

  const ensureFAQCategorySelection = () => {
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

  const ensureFAQSelection = () => {
    ensureFAQPageSelection()
    ensureFAQCategorySelection()
  }

  const showCreateDialog = () => {
    dialogMode.value = 'create'
    resetForm()
    ensureFAQSelection()
    dialogVisible.value = true
  }

  const showEditDialog = async (faq) => {
    dialogMode.value = 'edit'
    try {
      const payload = await faqAdminApi.getFAQ(faq.id)
      const detail = payload.faq || faq
      Object.assign(faqForm, {
        id: detail.id,
        question: detail.question || '',
        answer: detail.answer || '',
        answer_image_url: detail.answer_image_url || '',
        answer_image_alt: detail.answer_image_alt || '',
        answer_image_width: detail.answer_image_width || 0,
        answer_image_height: detail.answer_image_height || 0,
        category: detail.category || '',
        page_id: detail.page_id || '',
        locale: detail.locale || 'zh',
        status: detail.status || 'published',
        order: detail.order ?? detail.sort_order ?? 0
      })
      if (!faqForm.page_id || !faqForm.category) ensureFAQSelection()
      clearFormErrors()
      dialogVisible.value = true
    } catch (error) {
      console.error('Failed to fetch FAQ detail:', error)
    }
  }

  const submitForm = async () => {
    const payload = buildFAQPayload()
    if (!validateForm(payload)) return

    submitting.value = true
    try {
      if (dialogMode.value === 'create') {
        await faqAdminApi.createFAQ(payload)
        toast.success('FAQ 创建成功')
      } else {
        await faqAdminApi.updateFAQ(faqForm.id, payload)
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
    ensureFAQPageSelection()
    ensureFAQCategorySelection()
  })

  watch(() => faqForm.page_id, () => {
    ensureFAQCategorySelection()
  })

  return {
    dialogVisible,
    dialogMode,
    submitting,
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
