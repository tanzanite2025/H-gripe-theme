import { computed, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { faqAdminApi } from '@/api/faq'
import { buildStructurePageOptions } from '@/lib/faqAdminPresentation'

export function useFaqStructure({ onChanged }) {
  const structureLoading = ref(false)
  const faqStructures = reactive({ zh: [], en: [] })
  const activeStructureLocale = ref('zh')
  const pageDialogVisible = ref(false)
  const pageSubmitting = ref(false)
  const categoryDialogVisible = ref(false)
  const categoryDialogMode = ref('create')
  const categorySubmitting = ref(false)

  const pageForm = reactive({
    page_id: '',
    route_path: '',
    domain: '',
    locale: 'zh',
    title: '',
    subtitle: '',
    status: 'active',
    sort_order: 0
  })
  const categoryForm = reactive({
    id: null,
    page_id: '',
    category_key: '',
    name: '',
    icon: '',
    locale: 'zh',
    status: 'active',
    sort_order: 0
  })

  const faqStructure = computed(() => faqStructures[activeStructureLocale.value] || [])
  const allStructurePages = computed(() => Object.values(faqStructures).flat())
  const structurePageOptions = computed(() => (
    buildStructurePageOptions(faqStructure.value, allStructurePages.value)
  ))

  const fetchFAQStructure = async (locale = activeStructureLocale.value) => {
    if (locale === activeStructureLocale.value) structureLoading.value = true
    try {
      const payload = await faqAdminApi.listStructure(locale)
      faqStructures[locale] = payload.pages || []
    } catch (error) {
      console.error('Failed to fetch FAQ structure:', error)
    } finally {
      if (locale === activeStructureLocale.value) structureLoading.value = false
    }
  }

  const refreshFAQStructure = () => Promise.all([fetchFAQStructure('zh'), fetchFAQStructure('en')])

  const switchStructureLocale = async (locale) => {
    activeStructureLocale.value = locale
    await fetchFAQStructure(locale)
  }

  const showPageDialog = (page) => {
    Object.assign(pageForm, {
      page_id: page.page_id,
      route_path: page.route_path || '',
      domain: page.domain || '',
      locale: page.locale || activeStructureLocale.value,
      title: page.title || '',
      subtitle: page.subtitle || '',
      status: page.status || 'active',
      sort_order: page.sort_order || 0
    })
    pageDialogVisible.value = true
  }

  const submitPageForm = async () => {
    if (!pageForm.page_id || !pageForm.title.trim()) {
      toast.error('页面标识和页面标题不能为空')
      return
    }

    pageSubmitting.value = true
    try {
      await faqAdminApi.updatePage(pageForm.page_id, {
        route_path: pageForm.route_path.trim(),
        domain: pageForm.domain.trim(),
        locale: pageForm.locale,
        title: pageForm.title.trim(),
        subtitle: pageForm.subtitle.trim(),
        status: pageForm.status,
        sort_order: Math.max(0, Number(pageForm.sort_order || 0))
      })
      toast.success('FAQ 页面已保存')
      pageDialogVisible.value = false
      await onChanged()
    } catch (error) {
      console.error('Failed to save FAQ page:', error)
    } finally {
      pageSubmitting.value = false
    }
  }

  const slugifyKey = (value) => String(value || '')
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')

  const showCategoryDialog = (mode, page, category = null) => {
    categoryDialogMode.value = mode

    if (mode === 'edit' && category) {
      Object.assign(categoryForm, {
        id: category.id,
        page_id: category.page_id,
        category_key: category.category_key || '',
        name: category.name || '',
        icon: category.icon || '',
        locale: category.locale || page.locale || activeStructureLocale.value,
        status: category.status || 'active',
        sort_order: category.sort_order || 0
      })
    } else {
      Object.assign(categoryForm, {
        id: null,
        page_id: page.page_id,
        category_key: '',
        name: '',
        icon: '',
        locale: page.locale || activeStructureLocale.value,
        status: 'active',
        sort_order: ((page.categories || []).length + 1) * 10
      })
    }

    categoryDialogVisible.value = true
  }

  const submitCategoryForm = async () => {
    const payload = {
      page_id: categoryForm.page_id,
      category_key: (categoryForm.category_key || slugifyKey(categoryForm.name)).trim(),
      name: categoryForm.name.trim(),
      icon: categoryForm.icon.trim(),
      locale: categoryForm.locale,
      status: categoryForm.status,
      sort_order: Math.max(0, Number(categoryForm.sort_order || 0))
    }

    if (!payload.page_id || !payload.category_key || !payload.name) {
      toast.error('页面、分类标识和分类名称不能为空')
      return
    }

    categorySubmitting.value = true
    try {
      if (categoryDialogMode.value === 'create') {
        await faqAdminApi.createCategory(payload)
        toast.success('FAQ 分类已创建')
      } else {
        await faqAdminApi.updateCategory(categoryForm.id, payload)
        toast.success('FAQ 分类已保存')
      }
      categoryDialogVisible.value = false
      await onChanged()
    } catch (error) {
      console.error('Failed to save FAQ category:', error)
    } finally {
      categorySubmitting.value = false
    }
  }

  return {
    structureLoading,
    faqStructures,
    activeStructureLocale,
    faqStructure,
    allStructurePages,
    structurePageOptions,
    pageDialogVisible,
    pageSubmitting,
    pageForm,
    categoryDialogVisible,
    categoryDialogMode,
    categorySubmitting,
    categoryForm,
    fetchFAQStructure,
    refreshFAQStructure,
    switchStructureLocale,
    showPageDialog,
    submitPageForm,
    showCategoryDialog,
    submitCategoryForm
  }
}
