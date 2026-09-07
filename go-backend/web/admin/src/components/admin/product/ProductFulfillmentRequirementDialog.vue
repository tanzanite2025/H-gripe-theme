<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg">
      <DialogHeader>
        <DialogTitle>产品履约要求</DialogTitle>
        <DialogDescription>
          仅维护显式的张力表规则。订单创建时按商品或 Variant 快照读取，不根据分类、名称或 QUICKBUY 推断。
        </DialogDescription>
      </DialogHeader>

      <div v-if="product" class="space-y-4">
        <div class="flex min-w-0 items-start justify-between gap-4 rounded-lg border bg-muted/20 px-3 py-3">
          <div class="min-w-0">
            <p class="truncate text-sm font-black">{{ product.name || `商品 #${product.id}` }}</p>
            <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground">
              ID {{ product.id }} · {{ primarySku(product) }}
            </p>
          </div>
          <AdminStatusBadge tone="blue">spoke_tension_qc</AdminStatusBadge>
        </div>

        <div v-if="loading" class="flex min-h-40 items-center justify-center text-muted-foreground">
          <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载履约要求" />
        </div>

        <template v-else>
          <section class="space-y-2">
            <div class="flex items-end justify-between gap-3">
              <div>
                <h2 class="text-sm font-black tracking-tight">商品默认规则</h2>
                <p class="mt-1 text-[10px] text-muted-foreground">所有没有单独 Variant 规则的 SKU 都按这里解析。</p>
              </div>
              <Button
                v-if="canEdit"
                variant="outline"
                size="sm"
                :disabled="saving"
                @click="selectScope(null)"
              >
                <Pencil class="size-3.5" />
                编辑
              </Button>
            </div>

            <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
              <div class="min-w-0">
                <p class="text-xs font-bold">{{ ruleSummary(productRule) }}</p>
                <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground">
                  {{ productRule ? `${productRule.rule_version || '无版本'} · ${productRule.reason || '未填写原因'}` : '未配置商品默认规则' }}
                </p>
              </div>
              <AdminStatusBadge :tone="ruleTone(productRule)">
                {{ ruleStatusName(productStatus(productRule)) }}
              </AdminStatusBadge>
            </div>
          </section>

          <section class="space-y-2">
            <div>
              <h2 class="text-sm font-black tracking-tight">Variant 规则</h2>
              <p class="mt-1 text-[10px] text-muted-foreground">只有这里明确保存的 SKU 规则才会覆盖商品默认规则。</p>
            </div>

            <div v-if="variantRows.length" class="divide-y rounded-lg border">
              <div
                v-for="row in variantRows"
                :key="String(row.variant.id)"
                class="flex items-center justify-between gap-3 px-3 py-3"
              >
                <div class="min-w-0">
                  <p class="truncate text-xs font-bold">{{ row.variant.title || '未命名 Variant' }}</p>
                  <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground">
                    {{ row.variant.sku || `Variant #${row.variant.id}` }}
                  </p>
                  <p class="mt-1 text-[10px] text-muted-foreground/80">{{ row.summary }}</p>
                </div>
                <div class="flex shrink-0 items-center gap-2">
                  <AdminStatusBadge :tone="ruleTone(row.rule)">
                    {{ ruleStatusName(row.status) }}
                  </AdminStatusBadge>
                  <Button
                    v-if="canEdit"
                    variant="ghost"
                    size="icon"
                    :aria-label="`编辑 ${row.variant.sku || `Variant ${row.variant.id}`} 履约规则`"
                    title="编辑 Variant 履约规则"
                    :disabled="saving"
                    @click="selectScope(Number(row.variant.id))"
                  >
                    <Pencil class="size-3.5" />
                  </Button>
                </div>
              </div>
            </div>
            <div v-else class="rounded-lg border border-dashed px-3 py-4 text-center text-xs text-muted-foreground">
              当前商品没有可维护的 Variant。
            </div>
          </section>

          <section class="space-y-3 border-t pt-4">
            <div class="flex items-start gap-2">
              <Info class="mt-0.5 size-4 shrink-0 text-primary" />
              <div>
                <h2 class="text-sm font-black tracking-tight">{{ editingScopeTitle }}</h2>
                <p class="mt-1 text-[10px] leading-5 text-muted-foreground">
                  停用规则仍保留为历史配置，但不会参与新订单的规则解析；规则版本和原因会随订单履约快照保存。
                </p>
              </div>
            </div>

            <div class="grid gap-3 md:grid-cols-2">
              <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
                <div>
                  <p class="text-xs font-bold">需要编轮质检张力表</p>
                  <p class="mt-0.5 text-[10px] text-muted-foreground">仅表示是否追加张力表证据项。</p>
                </div>
                <Switch v-model="ruleForm.spoke_tension_qc_required" :disabled="!canEdit || saving" aria-label="需要编轮质检张力表" />
              </div>

              <AdminFormField label="规则状态" required>
                <Select v-model="ruleForm.status" :disabled="!canEdit || saving">
                  <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="active">启用</SelectItem>
                    <SelectItem value="inactive">停用</SelectItem>
                  </SelectContent>
                </Select>
              </AdminFormField>

              <AdminFormField label="规则版本" required description="例如 spoke-qc-2026-09">
                <Input v-model="ruleForm.rule_version" :disabled="!canEdit || saving" class="font-mono" placeholder="请输入规则版本" />
              </AdminFormField>

              <AdminFormField label="维护原因" class="md:col-span-2">
                <Textarea v-model="ruleForm.reason" :disabled="!canEdit || saving" placeholder="说明为什么需要或不需要张力表" />
              </AdminFormField>
            </div>

            <div class="flex justify-end">
              <Button :disabled="!canEdit || saving" @click="saveRule">
                <LoaderCircle v-if="saving" class="size-4 animate-spin" />
                <Save v-else class="size-4" />
                保存规则
              </Button>
            </div>
          </section>
        </template>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { Info, LoaderCircle, Pencil, Save } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import productFulfillmentRequirementApi from '@/api/productFulfillmentRequirements'
import type {
  ProductFulfillmentRequirementRule,
  ProductFulfillmentRuleStatus,
} from '@/modules/product/productFulfillmentRequirementTypes'
import type { ProductRecord, ProductVariantRecord } from '@/modules/product/productEditorTypes'

interface RuleForm {
  spoke_tension_qc_required: boolean
  status: 'active' | 'inactive'
  rule_version: string
  reason: string
}

interface VariantRuleRow {
  variant: ProductVariantRecord
  rule: ProductFulfillmentRequirementRule | null
  status: ProductFulfillmentRuleStatus
  summary: string
}

const props = withDefaults(defineProps<{
  open?: boolean
  product?: ProductRecord | null
  canEdit?: boolean
}>(), {
  open: false,
  product: null,
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
}>()

const loading = ref(false)
const saving = ref(false)
const rules = ref<ProductFulfillmentRequirementRule[]>([])
const requestSequence = ref(0)
const selectedVariantID = ref<number | null>(null)
const ruleForm = reactive<RuleForm>({
  spoke_tension_qc_required: false,
  status: 'active',
  rule_version: '',
  reason: '',
})

const productRule = computed(() => rules.value.find((rule) => rule.variant_id == null) || null)
const productVariants = computed<ProductVariantRecord[]>(() => (
  Array.isArray(props.product?.variants) ? props.product.variants.filter((variant) => variant?.id != null) : []
))
const variantRows = computed<VariantRuleRow[]>(() => productVariants.value.map((variant) => {
  const rule = explicitRule(Number(variant.id))
  const activeProductRule = productRule.value?.status === 'active' ? productRule.value : null
  const summary = rule
    ? ruleSummary(rule)
    : activeProductRule
      ? `继承商品默认：${activeProductRule.spoke_tension_qc_required ? '需要张力表' : '不需要张力表'}`
      : '未单独配置，当前不追加张力表'
  return {
    variant,
    rule,
    status: rule?.status || (activeProductRule ? 'inherited' : 'unconfigured'),
    summary,
  }
}))
const editingScopeTitle = computed(() => {
  if (selectedVariantID.value == null) return '编辑商品默认规则'
  const variant = productVariants.value.find((item) => Number(item.id) === selectedVariantID.value)
  return `编辑 Variant 规则 · ${variant?.sku || `#${selectedVariantID.value}`}`
})

const primarySku = (product: ProductRecord): string => {
  const variants = Array.isArray(product.variants) ? product.variants : []
  const defaultVariant = variants.find((variant) => variant?.is_default)
  return String(defaultVariant?.sku || product.sku || '未设置 SKU')
}

const explicitRule = (variantID: number | null): ProductFulfillmentRequirementRule | null => (
  rules.value.find((rule) => (
    variantID == null ? rule.variant_id == null : Number(rule.variant_id) === variantID
  )) || null
)

const ruleSummary = (rule: ProductFulfillmentRequirementRule | null): string => {
  if (!rule) return '未配置张力表要求'
  return rule.spoke_tension_qc_required ? '需要编轮质检张力表' : '不需要编轮质检张力表'
}

const productStatus = (rule: ProductFulfillmentRequirementRule | null): ProductFulfillmentRuleStatus => (
  rule?.status || 'unconfigured'
)

const ruleStatusName = (status: ProductFulfillmentRuleStatus): string => ({
  active: '启用',
  inactive: '已停用',
  inherited: '继承默认',
  unconfigured: '未配置',
}[status] || status || '未配置')

const ruleTone = (rule: ProductFulfillmentRequirementRule | null): AdminStatusTone => {
  if (!rule) return 'gray'
  if (rule.status !== 'active') return 'gray'
  return rule.spoke_tension_qc_required ? 'blue' : 'green'
}

const resetRuleForm = (variantID: number | null): void => {
  selectedVariantID.value = variantID
  const rule = explicitRule(variantID)
  const inheritedRule = variantID == null ? null : productRule.value
  Object.assign(ruleForm, {
    spoke_tension_qc_required: rule?.spoke_tension_qc_required ?? inheritedRule?.spoke_tension_qc_required ?? false,
    status: rule?.status === 'inactive' ? 'inactive' : 'active',
    rule_version: rule?.rule_version || '',
    reason: rule?.reason || '',
  })
}

const selectScope = (variantID: number | null): void => {
  resetRuleForm(variantID)
}

const errorMessage = (error: unknown, fallback: string): string => {
  const responseError = (error as { response?: { data?: { error?: unknown } } })?.response?.data?.error
  return typeof responseError === 'string' && responseError.trim() ? responseError : fallback
}

const loadRules = async (): Promise<void> => {
  if (!props.product?.id) return
  const sequence = requestSequence.value + 1
  requestSequence.value = sequence
  loading.value = true
  try {
    const result = await productFulfillmentRequirementApi.list(props.product.id)
    if (sequence !== requestSequence.value) return
    if (result.product_id !== Number(props.product.id)) {
      throw new Error('产品履约要求返回了不匹配的商品')
    }
    rules.value = result.rules
    resetRuleForm(selectedVariantID.value)
  } catch (error) {
    if (sequence !== requestSequence.value) return
    rules.value = []
    toast.error(errorMessage(error, '产品履约要求加载失败'))
  } finally {
    if (sequence === requestSequence.value) loading.value = false
  }
}

const saveRule = async (): Promise<void> => {
  if (!props.product?.id || !props.canEdit || saving.value) return
  const ruleVersion = ruleForm.rule_version.trim()
  if (!ruleVersion) {
    toast.error('请填写规则版本')
    return
  }

  saving.value = true
  try {
    await productFulfillmentRequirementApi.upsert(props.product.id, {
      ...(selectedVariantID.value == null ? {} : { variant_id: selectedVariantID.value }),
      spoke_tension_qc_required: ruleForm.spoke_tension_qc_required,
      status: ruleForm.status,
      rule_version: ruleVersion,
      reason: ruleForm.reason.trim(),
    })
    toast.success('产品履约规则已保存')
    await loadRules()
  } catch (error) {
    toast.error(errorMessage(error, '产品履约规则保存失败'))
  } finally {
    saving.value = false
  }
}

watch(() => props.open, (open) => {
  if (open) {
    selectedVariantID.value = null
    void loadRules()
  } else {
    requestSequence.value += 1
  }
})

watch(() => props.product?.id, (productID, previousProductID) => {
  if (props.open && productID && productID !== previousProductID) {
    selectedVariantID.value = null
    void loadRules()
  }
})
</script>
