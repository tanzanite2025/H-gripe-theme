import { reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import shippingApi from '@/api/shipping'
import {
  clearErrors,
  defaultShippingPackagingForm,
  resetReactive,
} from '@/lib/shippingForms'

export const useShippingPackagingManager = (options: Record<string, any> = {}) => {
  const fetchPackagingRules = options.fetchPackagingRules || (() => Promise.resolve())

  const packagingDialogOpen = ref(false)
  const packagingDialogMode = ref<'create' | 'edit'>('create')
  const packagingSubmitting = ref(false)
  const packagingErrors = reactive<Record<string, string>>({})
  const packagingForm = reactive(defaultShippingPackagingForm())
  const packagingAppliesDialogOpen = ref(false)
  const packagingAppliesRule = ref<any>(null)

  const clearPackagingError = (field: string) => {
    delete packagingErrors[field]
  }

  const showCreatePackagingDialog = () => {
    packagingDialogMode.value = 'create'
    resetReactive(packagingForm, defaultShippingPackagingForm())
    clearErrors(packagingErrors)
    packagingDialogOpen.value = true
  }

  const showEditPackagingDialog = (rule: any) => {
    packagingDialogMode.value = 'edit'
    resetReactive(packagingForm, {
      ...defaultShippingPackagingForm(),
      ...rule,
      box_weight: Number(rule.box_weight || 0),
      box_length: Number(rule.box_length || 0),
      box_width: Number(rule.box_width || 0),
      box_height: Number(rule.box_height || 0),
      max_weight: Number(rule.max_weight || 0),
      is_active: rule.is_active !== false,
    })
    clearErrors(packagingErrors)
    packagingDialogOpen.value = true
  }

  const showPackagingAppliesDialog = (rule: any) => {
    packagingAppliesRule.value = rule
    packagingAppliesDialogOpen.value = true
  }

  const validatePackaging = () => {
    clearErrors(packagingErrors)
    if (!packagingForm.rule_name?.trim()) packagingErrors.rule_name = '请输入规则名称'
    if (Number(packagingForm.box_weight) < 0) packagingErrors.box_weight = '包装重量不能小于 0'
    if (Number(packagingForm.max_weight) < 0) packagingErrors.max_weight = '最大承重不能小于 0'
    return Object.keys(packagingErrors).length === 0
  }

  const savePackagingRule = async () => {
    if (!validatePackaging()) return

    packagingSubmitting.value = true
    try {
      const payload = {
        rule_name: packagingForm.rule_name.trim(),
        description: packagingForm.description || '',
        box_weight: Number(packagingForm.box_weight || 0),
        box_length: Number(packagingForm.box_length || 0),
        box_width: Number(packagingForm.box_width || 0),
        box_height: Number(packagingForm.box_height || 0),
        max_weight: Number(packagingForm.max_weight || 0),
        is_active: Boolean(packagingForm.is_active),
      }

      if (packagingDialogMode.value === 'create') {
        await shippingApi.createPackagingRule(payload)
        toast.success('包装规则已创建')
      } else {
        await shippingApi.updatePackagingRule(packagingForm.id, payload)
        toast.success('包装规则已更新')
      }

      packagingDialogOpen.value = false
      await fetchPackagingRules()
    } catch (error) {
      console.error('Failed to save packaging rule:', error)
    } finally {
      packagingSubmitting.value = false
    }
  }

  return {
    packagingDialogOpen,
    packagingDialogMode,
    packagingSubmitting,
    packagingErrors,
    packagingForm,
    packagingAppliesDialogOpen,
    packagingAppliesRule,
    clearPackagingError,
    showCreatePackagingDialog,
    showEditPackagingDialog,
    showPackagingAppliesDialog,
    savePackagingRule,
  }
}

export default useShippingPackagingManager
