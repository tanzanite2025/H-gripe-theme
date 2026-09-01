<template>
  <div class="space-y-4">
    <AdminPageHeader title="QUICK 配置" description="配置统一弹层说明和每步产品分类，入口与选择行为由系统固定">
      <template #actions>
        <Button variant="outline" :disabled="loading" @click="reload">
          <RefreshCw class="size-4" />
          刷新
        </Button>
        <Button variant="outline" :disabled="validating || (!selectedFlow?.version?.id && (!canEdit || !canSave))" @click="validateCurrentVersion({ saveFirst: canEdit })">
          <ShieldCheck class="size-4" />
          校验
        </Button>
        <Button :disabled="saving || !canEdit || !canSave" @click="saveDraft">
          <Save class="size-4" />
          保存草稿
        </Button>
        <Button :disabled="publishing || !canEdit || !canSave" @click="publishDraft">
          <Rocket class="size-4" />
          发布
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <AdminTablePanel :loading="loading || loadingFlow">
      <template #header>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 class="text-sm font-black uppercase tracking-tight">{{ editorTitle }}</h2>
            <p class="text-[11px] font-bold text-muted-foreground">{{ editorSubtitle }}</p>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <Badge :variant="statusBadgeVariant">{{ versionStatusLabel }}</Badge>
            <Button size="sm" variant="outline" :disabled="formDisabled" @click="addStep">
              <Plus class="size-3.5" />
              新增步骤
            </Button>
          </div>
        </div>
      </template>

      <div class="space-y-5 p-4">
          <section class="grid gap-3 sm:grid-cols-3">
            <div class="rounded-lg border border-dashed border-border/80 px-3 py-2">
              <span class="field-label">功能</span>
              <p class="mt-1 text-sm font-black">QUICK Build</p>
            </div>
            <div class="rounded-lg border border-dashed border-border/80 px-3 py-2">
              <span class="field-label">入口</span>
              <p class="mt-1 font-mono text-sm font-black uppercase">DOCK</p>
            </div>
            <div class="rounded-lg border border-dashed border-border/80 px-3 py-2">
              <span class="field-label">结构</span>
              <p class="mt-1 text-sm font-black">基础 3 步，可追加</p>
            </div>
          </section>

          <Alert v-if="!canEdit" class="rounded-lg">
            <Lock class="size-4" />
            <AlertTitle>只读模式</AlertTitle>
            <AlertDescription>当前账号只有 product:view 权限，可以查看和校验 QUICK 流程，但不能修改或发布。</AlertDescription>
          </Alert>

          <Alert v-if="validationResult" :variant="validationResult.valid ? 'default': 'destructive'" class="rounded-lg">
            <ShieldCheck v-if="validationResult.valid" class="size-4" />
            <CircleAlert v-else class="size-4" />
            <AlertTitle>{{ validationTitle }}</AlertTitle>
            <AlertDescription>
              <div v-if="validationIssues.length" class="mt-1 space-y-1">
                <p
                  v-for="issue in validationIssues"
                  :key="`${issue.severity}-${issue.code}-${issue.step_key || issue.rule_key || issue.product_category_id || issue.product_specification_template_id || issue.message}`"
                  class="text-xs font-bold"
                >
                  [{{ issue.severity }}] {{ issue.message }}
                </p>
              </div>
              <span v-else>当前保存版本可以发布。</span>
            </AlertDescription>
          </Alert>

          <section class="space-y-3">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h3 class="text-sm font-black uppercase tracking-tight">步骤配置</h3>
                <p class="text-[11px] font-bold text-muted-foreground">每一步只绑定产品分类，名称由系统按顺序生成。</p>
              </div>
            </div>

            <div v-if="stepForms.length" class="grid gap-3 lg:grid-cols-3">
              <div
                v-for="(step, index) in stepForms"
                :key="step.client_id"
                class="flex min-h-36 min-w-0 flex-col rounded-lg border border-dashed border-border/80 bg-muted/20 p-3"
              >
                <div class="flex items-center justify-between gap-3">
                  <p class="min-w-0 truncate text-sm font-black">第 {{ index + 1 }} 步</p>
                  <Badge v-if="index < 3" variant="secondary">基础步骤</Badge>
                </div>

                <label class="mt-3 block min-w-0 space-y-1.5">
                  <span class="field-label">产品分类</span>
                  <select
                    v-model.number="step.product_category_id"
                    :class="nativeSelectClass"
                    :disabled="formDisabled || !productCategoryOptions.length"
                  >
                    <option :value="0">请选择产品分类</option>
                    <option
                      v-for="category in productCategoryOptions"
                      :key="category.id"
                      :value="category.id"
                    >
                      {{ productCategoryOptionLabel(category) }}
                    </option>
                  </select>
                </label>

                <div class="mt-auto flex items-center justify-between gap-2 pt-3">
                  <p class="font-mono text-[10px] font-bold text-muted-foreground">
                    {{ normalizeKey(step.step_key) }}
                  </p>
                </div>
              </div>
            </div>
            <div v-else class="rounded-lg border border-dashed border-border/80 px-4 py-10 text-center text-sm font-bold text-muted-foreground">
              暂无步骤
            </div>
          </section>

          <section class="space-y-3 border-t border-dashed border-border/80 pt-4">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <div>
                <h3 class="text-sm font-black uppercase tracking-tight">QUICK 统一弹层说明</h3>
                <p class="text-[11px] font-bold text-muted-foreground">前台左右两列的问号都打开这同一份说明，未填写的语言回退基础文案。</p>
              </div>
              <Badge variant="outline">{{ filledFlowHelpTranslationCount }}/{{ languageOptions.length }} languages</Badge>
            </div>

            <label class="block space-y-1.5">
              <span class="field-label">基础说明</span>
              <Textarea
                v-model="flowHelpText"
                class="min-h-16 resize-y"
                placeholder="请输入 QUICK 弹层统一说明"
                :disabled="formDisabled"
              />
            </label>

            <details class="rounded-lg border border-dashed border-border/80 bg-background/50 px-3 py-2">
              <summary class="cursor-pointer list-none text-xs font-black">
                多语言说明
                <span class="ml-1 font-mono text-[10px] text-muted-foreground">
                  {{ filledFlowHelpTranslationCount }}/{{ languageOptions.length }}
                </span>
              </summary>
              <div v-if="flowHelpTranslations.length" class="mt-3 grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
                <section
                  v-for="translation in flowHelpTranslations"
                  :key="`flow-help-${translation.locale}`"
                  class="min-w-0 rounded-lg border border-border/70 bg-background/70 p-2.5"
                >
                  <div class="mb-2 flex items-center justify-between gap-2">
                    <span class="truncate text-xs font-black">{{ languageLabel(translation.locale) }}</span>
                    <span class="font-mono text-[10px] text-muted-foreground">{{ translation.locale }}</span>
                  </div>
                  <Textarea
                    v-model="translation.help_text"
                    class="min-h-14 resize-y text-xs"
                    :placeholder="`请输入${languageLabel(translation.locale)}QUICK 弹层说明`"
                    :disabled="formDisabled"
                  />
                </section>
              </div>
            </details>
          </section>

          <section class="space-y-3 border-t border-dashed border-border/80 pt-4">
            <div class="flex flex-wrap items-end justify-between gap-3">
              <div>
                <h3 class="text-sm font-black uppercase tracking-tight">候选预览</h3>
                <p class="text-[11px] font-bold text-muted-foreground">基于已保存 version 查看某一步实际可选商品。</p>
              </div>
              <div class="flex min-w-0 flex-1 flex-wrap items-end justify-end gap-2">
                <label class="min-w-40 space-y-1.5">
                  <span class="field-label">Step</span>
                  <select v-model="previewStepKey" :class="nativeSelectClass" :disabled="previewing || !selectedFlow?.version?.id">
                    <option v-for="(step, index) in previewableSteps" :key="step.client_id" :value="normalizeKey(step.step_key)">
                      第 {{ index + 1 }} 步
                    </option>
                  </select>
                </label>
                <label class="min-w-52 space-y-1.5">
                  <span class="field-label">搜索</span>
                  <Input
                    v-model="previewKeyword"
                    placeholder="name / SKU"
                    :disabled="previewing || !selectedFlow?.version?.id"
                    @keydown.enter.prevent="previewCandidates"
                  />
                </label>
                <Button variant="outline" :disabled="previewing || !selectedFlow?.version?.id || !previewStepKey" @click="previewCandidates">
                  <Search class="size-4" />
                  预览候选
                </Button>
              </div>
            </div>

            <div v-if="previewResult" class="rounded-lg border border-dashed border-border/80 bg-background/70 p-3">
              <div class="mb-3 flex flex-wrap items-center justify-between gap-2 text-xs font-bold text-muted-foreground">
                <span>{{ previewStepLabel }} / {{ previewResult.total }} products</span>
                <span class="font-mono">v{{ previewResult.flow_version_id || selectedFlow?.version?.id }}</span>
              </div>
              <div v-if="previewProducts.length" class="grid gap-2 md:grid-cols-2 xl:grid-cols-3">
                <div
                  v-for="product in previewProducts"
                  :key="product.id"
                  class="flex min-w-0 gap-2 rounded-lg border border-border/70 bg-muted/20 p-2"
                >
                  <img
                    v-if="previewProductImage(product)"
                    :src="previewProductImage(product)"
                    :alt="previewProductName(product)"
                    class="size-12 shrink-0 rounded-md object-cover"
                  >
                  <div v-else class="grid size-12 shrink-0 place-items-center rounded-md bg-muted text-[10px] font-black text-muted-foreground">
                    QUICK
                  </div>
                  <div class="min-w-0 flex-1">
                    <p class="truncate text-xs font-black">{{ previewProductName(product) }}</p>
                    <p class="truncate font-mono text-[10px] font-bold text-muted-foreground">{{ product.sku || product.slug }}</p>
                    <div class="mt-1 flex flex-wrap items-center gap-1">
                      <Badge variant="outline">{{ previewProductScopeLabel }}</Badge>
                      <Badge :variant="product.availability === 'in_stock' ? 'secondary' : 'outline'">{{ product.availability || 'unknown' }}</Badge>
                      <span class="text-[11px] font-black">{{ previewProductPrice(product) }}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div v-else class="px-3 py-8 text-center text-sm font-bold text-muted-foreground">
                当前步骤没有候选商品
              </div>
            </div>
          </section>
        </div>
    </AdminTablePanel>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  CircleAlert,
  Layers3,
  ListChecks,
  Lock,
  Plus,
  RefreshCw,
  Rocket,
  Save,
  Search,
  ShieldCheck,
  Zap,
} from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import quickBuyApi, {
  type QuickBuyFlow,
  type QuickBuyFlowPayload,
  type QuickBuyFlowSummary,
  type QuickBuyFlowTranslation,
  type QuickBuyPreviewProduct,
  type QuickBuyPreviewResult,
  type QuickBuyProductCategoryRef,
  type QuickBuyStep,
  type QuickBuyValidationResult,
  type QuickBuyVersionPayload,
} from '@/api/quickBuy'
import productCategoryApi, { type ProductCategoryRecord } from '@/api/productCategories'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import { normalizeLocaleCode } from '@/lib/languages'
import { useAuthStore } from '@/stores/auth'

interface FlowTranslationForm {
  id: number | string | null
  locale: string
  help_text: string
}

interface StepForm {
  client_id: number
  step_key: string
  product_category_id: number
}

const defaultQuickBuyFlowSlug = 'quick-build'
const defaultQuickBuyStepKeys = ['product-search', 'specifications', 'quantity'] as const
const defaultQuickBuyStepNames = ['Step 1', 'Step 2', 'Step 3'] as const

const flows = ref<QuickBuyFlowSummary[]>([])
const selectedFlow = ref<QuickBuyFlow | null>(null)
const supportedLanguages = useSupportedLanguages()
const languageOptions = supportedLanguages.languageOptions
const activeFlowId = ref<number | null>(null)
const loading = ref(false)
const loadingFlow = ref(false)
const saving = ref(false)
const publishing = ref(false)
const validating = ref(false)
const validationResult = ref<QuickBuyValidationResult | null>(null)
const stepForms = ref<StepForm[]>([])
const productCategories = ref<ProductCategoryRecord[]>([])
const flowHelpText = ref('')
const flowHelpTranslations = ref<FlowTranslationForm[]>([])
const previewing = ref(false)
const previewStepKey = ref('')
const previewKeyword = ref('')
const previewResult = ref<QuickBuyPreviewResult | null>(null)
const authStore = useAuthStore()
let nextStepClientID = 1

const nativeSelectClass = 'h-9 w-full rounded-lg border border-dashed border-border/80 bg-background px-3 text-xs font-bold outline-none transition focus:ring-2 focus:ring-ring/40 disabled:cursor-not-allowed disabled:opacity-50'

const editorTitle = computed(() => selectedFlow.value ? selectedFlow.value.name : '新建 QUICK Flow')
const editorSubtitle = computed(() => {
  if (!selectedFlow.value?.version?.id) return 'draft version will be created on save'
  return `version ${selectedFlow.value.version.version_number || 1} / ${selectedFlow.value.version.status || 'draft'}`
})
const versionStatusLabel = computed(() => selectedFlow.value?.version?.status || 'new draft')
const statusBadgeVariant = computed(() => {
  const status = selectedFlow.value?.version?.status
  if (status === 'published') return 'default'
  if (status === 'draft') return 'secondary'
  return 'outline'
})
const canSave = computed(() => true)
const canEdit = computed(() => authStore.hasPermission('product:edit'))
const formDisabled = computed(() => !canEdit.value || loadingFlow.value || saving.value || publishing.value)
const validationIssues = computed(() => validationResult.value?.issues || [])
const previewableSteps = computed(() => stepForms.value.filter((step) => normalizeKey(String(step.step_key || ''))))
const previewProducts = computed(() => previewResult.value?.products || [])
const previewProductScopeLabel = computed(() => {
  const category = previewResult.value?.step?.product_categories?.[0]
  return category?.name || category?.slug || '产品分类'
})
const previewStepLabel = computed(() => {
  const key = normalizeKey(previewStepKey.value)
  const index = previewableSteps.value.findIndex((step) => normalizeKey(step.step_key) === key)
  return index >= 0 ? `第 ${index + 1} 步` : key
})
const validationTitle = computed(() => {
  if (!validationResult.value) return ''
  const errorCount = validationIssues.value.filter((issue) => issue.severity === 'error').length
  const warningCount = validationIssues.value.filter((issue) => issue.severity === 'warning').length
  if (errorCount > 0) return `校验未通过：${errorCount} 个错误`
  if (warningCount > 0) return `校验通过，但有 ${warningCount} 个提醒`
  return '校验通过'
})

const statItems = computed(() => [
  { key: 'flows', label: 'Flows', value: flows.value.length, icon: Zap, tone: 'blue' },
  { key: 'enabled', label: 'Enabled', value: flows.value.filter((flow) => flow.is_enabled).length, icon: Layers3, tone: 'green' },
  { key: 'drafts', label: 'Draft versions', value: flows.value.filter((flow) => draftVersion(flow)).length, icon: ListChecks, tone: 'amber' },
  { key: 'steps', label: 'Steps', value: stepForms.value.length, icon: ListChecks, tone: 'gray' },
])

const productCategoryOptions = computed(() => productCategories.value)

const languageLabel = (locale: string) => supportedLanguages.localeName(locale)
const filledFlowHelpTranslationCount = computed(() => flowHelpTranslations.value.filter((translation) => (
  translation.help_text.trim()
)).length)

const normalizeKey = (value: string) => value
  .trim()
  .toLowerCase()
  .replace(/\s+/g, '-')
  .replace(/[^a-z0-9_-]/g, '')

const productCategoryOptionLabel = (category: ProductCategoryRecord) => {
  const indent = '　'.repeat(Math.max(Number(category.depth || 1) - 1, 0))
  const status = category.is_enabled === false ? '（已停用）' : ''
  return `${indent}${category.name || category.slug}${status}`
}

const defaultStepName = (index: number) => defaultQuickBuyStepNames[index] || `Step ${index + 1}`

const defaultProductCategoryID = () => {
  const enabled = productCategoryOptions.value.find((category) => category.is_enabled !== false)
  return Number(enabled?.id || 0)
}

const firstProductCategoryID = (categories: QuickBuyProductCategoryRef[] = []) => (
  Number(categories.find((category) => Number(category.id || 0) > 0)?.id || 0)
)

const draftVersion = (flow: QuickBuyFlowSummary) => (flow.versions || []).find((version) => version.status === 'draft')
const syncPreviewStepKey = () => {
  const firstStep = previewableSteps.value[0]
  if (!firstStep) {
    previewStepKey.value = ''
    return
  }
  const current = normalizeKey(previewStepKey.value)
  if (!previewableSteps.value.some((step) => normalizeKey(step.step_key) === current)) {
    previewStepKey.value = normalizeKey(firstStep.step_key)
  }
}

const flowTranslationRowsFor = (source: QuickBuyFlowTranslation[] = []): FlowTranslationForm[] => {
  const existing = new Map<string, QuickBuyFlowTranslation>()
  source.forEach((translation) => {
    const locale = normalizeLocaleCode(translation.locale)
    if (locale) existing.set(locale, translation)
  })

  const rows = languageOptions.value.map((option) => {
    const translation = existing.get(option.value)
    return {
      id: translation?.id ?? null,
      locale: option.value,
      help_text: String(translation?.help_text || ''),
    }
  })
  const displayedLocales = new Set(rows.map((translation) => translation.locale))

  for (const [locale, translation] of existing) {
    if (displayedLocales.has(locale)) continue
    rows.push({
      id: translation.id ?? null,
      locale,
      help_text: String(translation.help_text || ''),
    })
  }

  return rows
}

const emptyStep = (index = stepForms.value.length, overrides: Partial<StepForm> = {}): StepForm => ({
  client_id: nextStepClientID++,
  step_key: `step-${index + 1}`,
  product_category_id: defaultProductCategoryID(),
  ...overrides,
})

const defaultStepForms = (): StepForm[] => [
  emptyStep(0, {
    step_key: 'product-search',
  }),
  emptyStep(1, {
    step_key: 'specifications',
  }),
  emptyStep(2, {
    step_key: 'quantity',
  }),
]

const resetFlowForm = () => {
  stepForms.value = defaultStepForms()
  flowHelpText.value = ''
  flowHelpTranslations.value = flowTranslationRowsFor()
  previewResult.value = null
  syncPreviewStepKey()
}

const hydrateFlowForm = (flow: QuickBuyFlow) => {
  const steps = flow.steps || []
  stepForms.value = steps.length
    ? steps.map((step, index) => stepToForm(step, index))
    : normalizeKey(flow.slug) === defaultQuickBuyFlowSlug
      ? defaultStepForms()
      : []
  flowHelpText.value = flow.help_text || ''
  flowHelpTranslations.value = flowTranslationRowsFor(flow.translations || [])
  previewResult.value = null
  syncPreviewStepKey()
}

const stepToForm = (step: QuickBuyStep, index: number): StepForm => ({
  client_id: nextStepClientID++,
  step_key: normalizeKey(step.step_key || step.slug || defaultQuickBuyStepKeys[index] || `step-${index + 1}`),
  product_category_id: firstProductCategoryID(step.product_categories || []) || defaultProductCategoryID(),
})

const buildVersionPayload = (): QuickBuyVersionPayload => ({
  starts_at: null,
  ends_at: null,
  steps: stepForms.value.map((step, index) => ({
    step_key: normalizeKey(step.step_key) || defaultQuickBuyStepKeys[index] || `step-${index + 1}`,
    name: defaultStepName(index),
    product_category_ids: Number(step.product_category_id || 0) > 0 ? [Number(step.product_category_id)] : [],
  })),
})

const buildFlowPayload = (): QuickBuyFlowPayload => ({
  slug: defaultQuickBuyFlowSlug,
  name: 'QUICK Build',
  description: 'Default QUICK build flow',
  help_text: flowHelpText.value.trim(),
  translations: flowHelpTranslations.value
    .map((translation) => ({
      id: Number(translation.id || 0),
      locale: normalizeLocaleCode(translation.locale),
      help_text: translation.help_text.trim(),
    }))
    .filter((translation) => translation.locale && translation.help_text),
  entry_surface: 'dock',
  is_enabled: true,
  sort_order: 100,
  version: buildVersionPayload(),
})

const loadFlows = async () => {
  flows.value = (await quickBuyApi.listFlows()).filter((flow) => normalizeKey(flow.slug) === defaultQuickBuyFlowSlug)
}

const loadProductCategories = async () => {
  const payload = await productCategoryApi.list({ include_disabled: true })
  productCategories.value = payload.flat || []
}

const reload = async () => {
  loading.value = true
  try {
    await supportedLanguages.fetchLanguages()
    await loadProductCategories()
    await loadFlows()
    if (activeFlowId.value && flows.value.some((flow) => flow.id === activeFlowId.value)) {
      await selectFlow(activeFlowId.value)
    } else if (flows.value[0]?.id) {
      await selectFlow(flows.value[0].id)
    } else {
      startCreate()
    }
  } finally {
    loading.value = false
  }
}

const selectFlow = async (flowId: number) => {
  activeFlowId.value = flowId
  validationResult.value = null
  previewResult.value = null
  loadingFlow.value = true
  try {
    const flow = await quickBuyApi.getFlow(flowId)
    selectedFlow.value = flow
    hydrateFlowForm(flow)
  } finally {
    loadingFlow.value = false
  }
}

const startCreate = () => {
  activeFlowId.value = null
  selectedFlow.value = null
  validationResult.value = null
  previewResult.value = null
  resetFlowForm()
}

const saveDraft = async () => {
  if (!canEdit.value) {
    toast.error('当前账号没有 product:edit 权限')
    return null
  }
  saving.value = true
  try {
    const payload = buildFlowPayload()
    let flow: QuickBuyFlow
    if (!selectedFlow.value?.id) {
      flow = await quickBuyApi.createFlow(payload)
      toast.success('QUICK flow 已创建')
    } else {
      flow = await quickBuyApi.saveFlowConfiguration(selectedFlow.value.id, payload)
      toast.success('QUICK 草稿已保存')
    }
    selectedFlow.value = flow
    activeFlowId.value = flow.id
    validationResult.value = null
    hydrateFlowForm(flow)
    await loadFlows()
    return flow
  } catch (error) {
    console.error('Failed to save quick buy flow:', error)
    return null
  } finally {
    saving.value = false
  }
}

const publishDraft = async () => {
  if (!canEdit.value) {
    toast.error('当前账号没有 product:edit 权限')
    return
  }
  publishing.value = true
  try {
    const flow = await saveDraft()
    const versionID = flow?.version?.id
    if (!versionID) return
    const validation = await validateCurrentVersion({ flow, silent: true })
    if (!validation?.valid) {
      toast.error('校验未通过，请先处理错误后再发布')
      return
    }
    selectedFlow.value = await quickBuyApi.publishVersion(versionID)
    hydrateFlowForm(selectedFlow.value)
    await loadFlows()
    toast.success('QUICK flow 已发布')
  } catch (error) {
    console.error('Failed to publish quick buy flow:', error)
  } finally {
    publishing.value = false
  }
}

const validateCurrentVersion = async (options: { flow?: QuickBuyFlow, saveFirst?: boolean, silent?: boolean } = {}) => {
  let flow = options.flow || selectedFlow.value
  if (options.saveFirst && canEdit.value) {
    flow = await saveDraft() || flow
  }
  const versionID = flow?.version?.id
  if (!versionID) {
    toast.error('请先保存 QUICK 草稿')
    return null
  }

  validating.value = true
  try {
    const result = await quickBuyApi.validateVersion(versionID)
    validationResult.value = result
    if (!options.silent) {
      toast[result.valid ? 'success' : 'error'](result.valid ? '校验通过' : '校验未通过')
    }
    return result
  } catch (error) {
    console.error('Failed to validate quick buy flow:', error)
    return null
  } finally {
    validating.value = false
  }
}

const previewCandidates = async () => {
  const versionID = selectedFlow.value?.version?.id
  const stepKey = normalizeKey(previewStepKey.value)
  if (!versionID) {
    toast.error('请先保存 QUICK 草稿')
    return
  }
  if (!stepKey) {
    toast.error('请选择要预览的步骤')
    return
  }

  previewing.value = true
  try {
    previewResult.value = await quickBuyApi.previewVersion(versionID, {
      step_key: stepKey,
      keyword: previewKeyword.value.trim(),
      page: 1,
      page_size: 9,
    })
  } catch (error) {
    console.error('Failed to preview quick buy candidates:', error)
    toast.error('候选预览失败')
  } finally {
    previewing.value = false
  }
}

const previewProductImage = (product: QuickBuyPreviewProduct) => {
  const media = product.media || []
  const primary = media.find((item) => item.is_primary && item.media_type !== 'video') || media.find((item) => item.media_type !== 'video') || media[0]
  return primary?.thumbnail_url || primary?.url || ''
}

const previewProductName = (product: QuickBuyPreviewProduct) => product.name || product.title || product.sku || `#${product.id}`

const previewProductPrice = (product: QuickBuyPreviewProduct) => {
  const amount = Number(product.display_price?.amount ?? product.sale_price ?? product.price ?? 0)
  const currency = String(product.display_price?.currency || product.currency || 'USD')
  if (!amount) return ''
  try {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency }).format(amount)
  } catch {
    return `${currency} ${amount}`
  }
}

const addStep = () => {
  if (!canEdit.value) return
  validationResult.value = null
  previewResult.value = null
  stepForms.value.push(emptyStep())
  syncPreviewStepKey()
}

onMounted(() => {
  void reload()
})
</script>

<style scoped>
.field-label {
  display: block;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: hsl(var(--muted-foreground));
}
</style>
