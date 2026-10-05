import { computed, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { customsClassificationApi } from '@/api/customsClassifications'
import type {
  CustomsClassificationForm,
  CustomsClassificationRecord,
  LookupCandidate,
} from '@/modules/customs/customsClassificationTypes'

const createEmptyTemplateForm = (): CustomsClassificationForm => ({
  id: undefined,
  name: '',
  slug: '',
  component_kind: '',
  material: '',
  hs_code: '',
  cn_code: '',
  country_of_origin: '',
  customs_description: '',
  source: '',
  source_code: '',
  source_url: '',
  source_url_us: '',
  source_url_eu: '',
  source_url_uk: '',
  notes: '',
  verified_at: null,
  review_due_at: null,
  trade_remedy_risk_level: 'none',
  trade_remedy_risk_tags: [],
  trade_remedy_declaration_advice: '',
  status: 'active',
})

const slugifyTemplateValue = (value: string): string => String(value || '')
  .trim()
  .toLowerCase()
  .replace(/[^a-z0-9]+/g, '-')
  .replace(/^-+|-+$/g, '')
  .replace(/-{2,}/g, '-')

export const useCustomsTemplates = () => {
  const templates = ref<CustomsClassificationRecord[]>([])
  const templateLoading = ref(false)
  const templateDialogOpen = ref(false)
  const templateSaving = ref(false)
  const templateForm = reactive<CustomsClassificationForm>(createEmptyTemplateForm())
  const activeTemplates = computed(() => templates.value.filter((template) => template.status === 'active'))

  const fetchTemplates = async () => {
    templateLoading.value = true
    try {
      templates.value = await customsClassificationApi.list({ include_paused: true })
    } catch (error) {
      console.error('Failed to fetch customs classification templates:', error)
      toast.error('清关模板加载失败')
    } finally {
      templateLoading.value = false
    }
  }

  const resetTemplateForm = () => {
    Object.assign(templateForm, createEmptyTemplateForm())
  }

  const openTemplateCreate = () => {
    resetTemplateForm()
    templateDialogOpen.value = true
  }

  const openTemplateEdit = (template: CustomsClassificationRecord) => {
    Object.assign(templateForm, {
      id: template.id,
      name: template.name || '',
      slug: template.slug || '',
      component_kind: template.component_kind || '',
      material: template.material || '',
      hs_code: template.hs_code || '',
      cn_code: template.cn_code || '',
      country_of_origin: template.country_of_origin || '',
      customs_description: template.customs_description || '',
      source: template.source || '',
      source_code: template.source_code || '',
      source_url: template.source_url || '',
      source_url_us: template.source_url_us || template.source_url || '',
      source_url_eu: template.source_url_eu || '',
      source_url_uk: template.source_url_uk || '',
      notes: template.notes || '',
      verified_at: template.verified_at || null,
      review_due_at: template.review_due_at || null,
      trade_remedy_risk_level: template.trade_remedy_risk_level || 'none',
      trade_remedy_risk_tags: Array.isArray(template.trade_remedy_risk_tags) ? [...template.trade_remedy_risk_tags] : [],
      trade_remedy_declaration_advice: template.trade_remedy_declaration_advice || '',
      status: template.status || 'active',
    })
    templateDialogOpen.value = true
  }

  const saveTemplate = async () => {
    if (!templateForm.name.trim() || !templateForm.slug.trim() || !templateForm.hs_code.trim()) {
      toast.error('请填写模板名称、Slug 和 HS Code')
      return
    }
    templateSaving.value = true
    try {
      const payload = {
        name: templateForm.name.trim(),
        slug: templateForm.slug.trim(),
        component_kind: templateForm.component_kind.trim(),
        material: templateForm.material.trim(),
        hs_code: templateForm.hs_code.trim(),
        cn_code: templateForm.cn_code.trim(),
        country_of_origin: templateForm.country_of_origin.trim().toUpperCase(),
        customs_description: templateForm.customs_description.trim(),
        source: templateForm.source.trim(),
        source_code: templateForm.source_code.trim(),
        source_url: templateForm.source_url_us.trim() || templateForm.source_url.trim(),
        source_url_us: templateForm.source_url_us.trim() || templateForm.source_url.trim(),
        source_url_eu: templateForm.source_url_eu.trim(),
        source_url_uk: templateForm.source_url_uk.trim(),
        notes: templateForm.notes.trim(),
        trade_remedy_risk_level: templateForm.trade_remedy_risk_level,
        trade_remedy_risk_tags: [...templateForm.trade_remedy_risk_tags],
        trade_remedy_declaration_advice: templateForm.trade_remedy_declaration_advice.trim(),
        status: templateForm.status,
      }
      if (templateForm.id) {
        await customsClassificationApi.update(templateForm.id, payload)
      } else {
        await customsClassificationApi.create(payload)
      }
      templateDialogOpen.value = false
      toast.success('清关模板已保存')
      await fetchTemplates()
    } catch (error) {
      console.error('Failed to save customs classification template:', error)
      toast.error('清关模板保存失败')
    } finally {
      templateSaving.value = false
    }
  }

  const removeTemplate = async (template: CustomsClassificationRecord) => {
    if (!template.id || !window.confirm(`确定删除清关模板“${template.name}”？`)) return
    try {
      await customsClassificationApi.remove(template.id)
      toast.success('清关模板已删除')
      await fetchTemplates()
    } catch (error) {
      console.error('Failed to delete customs classification template:', error)
      toast.error('清关模板删除失败')
    }
  }

  const openTemplateFromCandidate = (candidate: LookupCandidate) => {
    resetTemplateForm()
    const candidateName = candidate.customs_description || candidate.description || candidate.source_code
    templateForm.name = candidateName
    templateForm.slug = slugifyTemplateValue(`${candidateName}-${candidate.source_code}`) || `${candidate.provider}-${candidate.source_code}`
    templateForm.hs_code = candidate.hs_code
    templateForm.cn_code = candidate.cn_code || ''
    templateForm.customs_description = candidate.customs_description || candidate.description
    templateForm.source = candidate.provider
    templateForm.source_code = candidate.source_code
    if (candidate.provider === 'uk_trade_tariff') {
      templateForm.source_url_uk = candidate.source_url
    } else {
      templateForm.source_url = candidate.source_url
      templateForm.source_url_us = candidate.source_url
    }
    templateDialogOpen.value = true
  }

  return {
    templates,
    templateLoading,
    templateDialogOpen,
    templateSaving,
    templateForm,
    activeTemplates,
    fetchTemplates,
    openTemplateCreate,
    openTemplateEdit,
    saveTemplate,
    removeTemplate,
    openTemplateFromCandidate,
  }
}

export default useCustomsTemplates
