import { computed, reactive, ref, watch } from 'vue'
import type { ComputedRef, Ref } from 'vue'
import { toast } from 'vue-sonner'
import { faqAdminApi } from '@/api/faq'
import { buildStructurePageOptions } from '@/lib/faqAdminPresentation'
import type { AdminLanguage, FAQCategory, FAQID, FAQStructureMap, FAQStructurePage } from '@/lib/faqAdminPresentation'

interface UseFaqStructureOptions {
  languages?: Ref<AdminLanguage[]> | ComputedRef<AdminLanguage[]>
  defaultLocale?: Ref<string> | ComputedRef<string>
  onChanged: () => Promise<unknown> | unknown
}

export type FAQCategoryDialogMode = 'create' | 'edit'

export interface FAQPageForm {
  page_id: string
  route_path: string
  domain: string
  locale: string
  title: string
  subtitle: string
  status: string
  sort_order: number
}

export interface FAQCategoryForm {
  id: FAQID | null
  page_id: string
  category_key: string
  name: string
  icon: string
  locale: string
  status: string
  sort_order: number
}

interface FAQStructurePayload {
  pages?: FAQStructurePage[]
}

export function useFaqStructure({ languages, defaultLocale, onChanged }: UseFaqStructureOptions) {
  const structureLoading = ref(false)
  const faqStructures = reactive<FAQStructureMap>({})
  const localeCodes = computed(() => (languages?.value || [])
    .filter((language) => language.enabled !== false && language.code)
    .map((language) => language.code))
  const resolveDefaultLocale = (): string => defaultLocale?.value || localeCodes.value[0] || ''
  const activeStructureLocale = ref(resolveDefaultLocale())
  const pageDialogVisible = ref(false)
  const pageSubmitting = ref(false)
  const lockedPageRoutePath = ref('')
  const categoryDialogVisible = ref(false)
  const categoryDialogMode = ref<FAQCategoryDialogMode>('create')
  const categorySubmitting = ref(false)

  const pageForm = reactive<FAQPageForm>({
    page_id: '',
    route_path: '',
    domain: '',
    locale: '',
    title: '',
    subtitle: '',
    status: 'active',
    sort_order: 0
  })
  const categoryForm = reactive<FAQCategoryForm>({
    id: null,
    page_id: '',
    category_key: '',
    name: '',
    icon: '',
    locale: '',
    status: 'active',
    sort_order: 0
  })

  const faqStructure = computed(() => faqStructures[activeStructureLocale.value] || [])
  const allStructurePages = computed(() => Object.values(faqStructures).flat().filter(Boolean) as FAQStructurePage[])
  const structurePageOptions = computed(() => (
    buildStructurePageOptions(faqStructure.value, allStructurePages.value)
  ))

  const syncStructureLocales = (codes: string[]): void => {
    for (const locale of Object.keys(faqStructures)) {
      if (!codes.includes(locale)) delete faqStructures[locale]
    }
    for (const locale of codes) {
      if (!Object.prototype.hasOwnProperty.call(faqStructures, locale)) {
        faqStructures[locale] = []
      }
    }
    if (!codes.includes(activeStructureLocale.value)) {
      activeStructureLocale.value = resolveDefaultLocale()
    }
  }

  const fetchFAQStructure = async (locale = activeStructureLocale.value, { setLoading = true }: { setLoading?: boolean } = {}): Promise<void> => {
    if (!locale) return
    if (setLoading) structureLoading.value = true
    try {
      const payload = await faqAdminApi.listStructure(locale) as FAQStructurePayload
      faqStructures[locale] = payload.pages || []
    } catch (error) {
      console.error('Failed to fetch FAQ structure:', error)
    } finally {
      if (setLoading) structureLoading.value = false
    }
  }

  const refreshFAQStructure = async (): Promise<void[]> => {
    const codes = localeCodes.value
    if (codes.length === 0) return []

    structureLoading.value = true
    try {
      return await Promise.all(codes.map((locale) => fetchFAQStructure(locale, { setLoading: false })))
    } finally {
      structureLoading.value = false
    }
  }

  const switchStructureLocale = async (locale: string): Promise<void> => {
    if (!locale || !localeCodes.value.includes(locale)) return
    activeStructureLocale.value = locale
    await fetchFAQStructure(locale)
  }

  const showPageDialog = (page: FAQStructurePage): void => {
    lockedPageRoutePath.value = page.route_path || ''
    Object.assign(pageForm, {
      page_id: page.page_id,
      route_path: lockedPageRoutePath.value,
      domain: page.domain || '',
      locale: page.locale || activeStructureLocale.value,
      title: page.title || '',
      subtitle: page.subtitle || '',
      status: page.status || 'active',
      sort_order: Number(page.sort_order || 0)
    })
    pageDialogVisible.value = true
  }

  const submitPageForm = async (): Promise<void> => {
    if (!pageForm.page_id || !pageForm.locale || !pageForm.title.trim()) {
      toast.error('页面标识、语言和页面标题不能为空')
      return
    }

    pageSubmitting.value = true
    try {
      await faqAdminApi.updatePage(pageForm.page_id, {
        route_path: lockedPageRoutePath.value,
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

  const slugifyKey = (value?: string | null): string => String(value || '')
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')

  const showCategoryDialog = (mode: FAQCategoryDialogMode, page: FAQStructurePage, category: FAQCategory | null = null): void => {
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
        sort_order: Number(category.sort_order || 0)
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

  const submitCategoryForm = async (): Promise<void> => {
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
    if (!payload.locale) {
      toast.error('分类语言不能为空')
      return
    }

    categorySubmitting.value = true
    try {
      if (categoryDialogMode.value === 'create') {
        await faqAdminApi.createCategory(payload)
        toast.success('FAQ 分类已创建')
      } else {
        await faqAdminApi.updateCategory(categoryForm.id as FAQID, payload)
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

  watch(localeCodes, (codes) => {
    syncStructureLocales(codes)
  }, { immediate: true })

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
