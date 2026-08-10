<template>
  <div class="space-y-4">
    <AdminPageHeader title="QUICK 选配流程" description="维护 Dock QUICK 弹窗的流程版本、步骤顺序和每一步可选产品类型">
      <template #actions>
        <Button variant="outline" :disabled="loading" @click="reload">
          <RefreshCw class="size-4" />
          刷新
        </Button>
        <Button variant="outline" :disabled="!canEdit" @click="startCreate">
          <Plus class="size-4" />
          新建
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

    <div class="grid gap-4 xl:grid-cols-[320px_minmax(0,1fr)]">
      <AdminTablePanel :loading="loading">
        <template #header>
          <div class="flex items-center justify-between gap-3">
            <div>
              <h2 class="text-sm font-black uppercase tracking-tight">流程列表</h2>
              <p class="text-[11px] font-bold text-muted-foreground">{{ flows.length }} flows</p>
            </div>
            <Badge variant="outline">{{ productTypes.length }} product types</Badge>
          </div>
        </template>

        <div v-if="flows.length" class="divide-y divide-dashed divide-border/70">
          <button
            v-for="flow in flows"
            :key="flow.id"
            type="button"
            class="flex w-full items-start justify-between gap-3 px-4 py-3 text-left transition hover:bg-muted/50"
            :class="activeFlowId === flow.id ? 'bg-muted/60' : ''"
            @click="selectFlow(flow.id)"
          >
            <span class="min-w-0">
              <span class="block truncate text-sm font-black">{{ flow.name }}</span>
              <span class="mt-1 block truncate font-mono text-[11px] font-bold text-muted-foreground">{{ flow.slug }}</span>
              <span class="mt-2 flex flex-wrap gap-1">
                <Badge :variant="flow.is_enabled ? 'default' : 'outline'">{{ flow.is_enabled ? 'enabled' : 'disabled' }}</Badge>
                <Badge v-if="publishedVersion(flow)" variant="secondary">published</Badge>
                <Badge v-if="draftVersion(flow)" variant="outline">draft</Badge>
              </span>
            </span>
            <ChevronRight class="mt-1 size-4 shrink-0 text-muted-foreground" />
          </button>
        </div>
        <div v-else class="px-4 py-10 text-center text-sm font-bold text-muted-foreground">
          暂无 QUICK flow
        </div>
      </AdminTablePanel>

      <AdminTablePanel :loading="loadingFlow">
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
                步骤
              </Button>
            </div>
          </div>
        </template>

        <div class="space-y-5 p-4">
          <section class="grid gap-3 lg:grid-cols-4">
            <label class="space-y-1.5">
              <span class="field-label">Flow slug</span>
              <Input v-model="flowForm.slug" class="font-mono" placeholder="wheelset-build" :disabled="formDisabled" />
            </label>
            <label class="space-y-1.5 lg:col-span-2">
              <span class="field-label">名称</span>
              <Input v-model="flowForm.name" placeholder="Wheelset Build" :disabled="formDisabled" />
            </label>
            <label class="space-y-1.5">
              <span class="field-label">入口</span>
              <Input v-model="flowForm.entry_surface" class="font-mono" placeholder="dock" :disabled="formDisabled" />
            </label>
            <label class="space-y-1.5">
              <span class="field-label">排序</span>
              <Input v-model="flowForm.sort_order" inputmode="numeric" :disabled="formDisabled" />
            </label>
            <div class="flex items-center justify-between gap-3 rounded-lg border border-dashed border-border/80 px-3 py-2">
              <span class="text-xs font-black uppercase">启用</span>
              <Switch v-model="flowForm.is_enabled" :disabled="formDisabled" />
            </div>
            <label class="space-y-1.5 lg:col-span-4">
              <span class="field-label">说明</span>
              <Textarea v-model="flowForm.description" rows="2" :disabled="formDisabled" />
            </label>
          </section>

          <Alert v-if="!canEdit" class="rounded-lg">
            <Lock class="size-4" />
            <AlertTitle>只读模式</AlertTitle>
            <AlertDescription>当前账号只有 product:view 权限，可以查看和校验 QUICK 流程，但不能修改或发布。</AlertDescription>
          </Alert>

          <Alert v-if="selectedFlow?.version?.status === 'published'" class="rounded-lg">
            <Info class="size-4" />
            <AlertTitle>已发布版本不会原地修改</AlertTitle>
            <AlertDescription>保存时会基于当前发布版本创建新的 draft，确认无误后再发布，已开始的前台选配会话仍绑定旧版本。</AlertDescription>
          </Alert>

          <Alert v-if="validationResult" :variant="validationResult.valid ? 'default' : 'destructive'" class="rounded-lg">
            <ShieldCheck v-if="validationResult.valid" class="size-4" />
            <CircleAlert v-else class="size-4" />
            <AlertTitle>{{ validationTitle }}</AlertTitle>
            <AlertDescription>
              <div v-if="validationIssues.length" class="mt-1 space-y-1">
                <p
                  v-for="issue in validationIssues"
                  :key="`${issue.severity}-${issue.code}-${issue.step_key || issue.rule_key || issue.product_type_id || issue.message}`"
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
                <p class="text-[11px] font-bold text-muted-foreground">每一步引用商品中心的产品类型，不重新定义分类。</p>
              </div>
              <Button size="sm" variant="outline" :disabled="formDisabled" @click="addStep">
                <Plus class="size-3.5" />
                添加步骤
              </Button>
            </div>

            <div v-if="stepForms.length" class="space-y-3">
              <div
                v-for="(step, index) in stepForms"
                :key="step.client_id"
                class="rounded-lg border border-dashed border-border/80 bg-muted/20 p-3"
              >
                <div class="grid gap-3 lg:grid-cols-[80px_1fr_1fr_150px_160px]">
                  <label class="space-y-1.5">
                    <span class="field-label">排序</span>
                    <Input v-model="step.sort_order" inputmode="numeric" :disabled="formDisabled" />
                  </label>
                  <label class="space-y-1.5">
                    <span class="field-label">Step key</span>
                    <Input v-model="step.step_key" class="font-mono" placeholder="rim" :disabled="formDisabled" />
                  </label>
                  <label class="space-y-1.5">
                    <span class="field-label">名称</span>
                    <Input v-model="step.name" placeholder="Rims" :disabled="formDisabled" />
                  </label>
                  <label class="space-y-1.5">
                    <span class="field-label">选择模式</span>
                    <select v-model="step.selection_mode" :class="nativeSelectClass" :disabled="formDisabled">
                      <option value="single">single</option>
                      <option value="multiple">multiple</option>
                      <option value="quantity">quantity</option>
                      <option value="auto">auto</option>
                    </select>
                  </label>
                  <div class="grid grid-cols-3 gap-2">
                    <label class="space-y-1.5">
                      <span class="field-label">Min</span>
                      <Input v-model="step.min_select" inputmode="numeric" :disabled="formDisabled" />
                    </label>
                    <label class="space-y-1.5">
                      <span class="field-label">Max</span>
                      <Input v-model="step.max_select" inputmode="numeric" :disabled="formDisabled" />
                    </label>
                    <label class="space-y-1.5">
                      <span class="field-label">Qty</span>
                      <Input v-model="step.default_quantity" inputmode="numeric" :disabled="formDisabled" />
                    </label>
                  </div>
                </div>

                <div class="mt-3 grid gap-3 lg:grid-cols-[1fr_220px]">
                  <label class="space-y-1.5">
                    <span class="field-label">帮助文案</span>
                    <Input v-model="step.help_text" :disabled="formDisabled" />
                  </label>
                  <div class="grid grid-cols-2 gap-2">
                    <label class="flex items-center justify-between gap-2 rounded-lg border border-dashed border-border/80 px-3 py-2">
                      <span class="text-[11px] font-black uppercase">必选</span>
                      <Switch v-model="step.is_required" :disabled="formDisabled" />
                    </label>
                    <label class="flex items-center justify-between gap-2 rounded-lg border border-dashed border-border/80 px-3 py-2">
                      <span class="text-[11px] font-black uppercase">可跳过</span>
                      <Switch v-model="step.allow_skip" :disabled="formDisabled" />
                    </label>
                  </div>
                </div>

                <div class="mt-3">
                  <span class="field-label">可选产品类型</span>
                  <div class="mt-2 grid gap-2 sm:grid-cols-2 xl:grid-cols-4">
                    <label
                      v-for="productType in productTypes"
                      :key="productType.id"
                      class="flex min-w-0 items-center gap-2 rounded-lg border border-dashed border-border/70 bg-background/70 px-2.5 py-2 text-xs font-bold"
                    >
                      <input
                        type="checkbox"
                        class="size-4 shrink-0 accent-primary"
                        :disabled="formDisabled"
                        :checked="step.product_type_ids.includes(productTypeId(productType))"
                        @change="toggleStepProductType(step, productTypeId(productType))"
                      >
                      <span class="min-w-0 flex-1 truncate">{{ productType.name || productType.slug }}</span>
                      <Badge v-if="productType.is_enabled === false" variant="outline">off</Badge>
                    </label>
                  </div>
                </div>

                <div class="mt-3 flex items-center justify-between gap-2">
                  <p class="font-mono text-[10px] font-bold text-muted-foreground">
                    {{ step.product_type_ids.length }} product types
                  </p>
                  <div class="flex gap-1">
                    <Button size="icon-xs" variant="ghost" :disabled="formDisabled || index === 0" aria-label="上移步骤" @click="moveStep(index, -1)">
                      <ArrowUp class="size-3" />
                    </Button>
                    <Button size="icon-xs" variant="ghost" :disabled="formDisabled || index === stepForms.length - 1" aria-label="下移步骤" @click="moveStep(index, 1)">
                      <ArrowDown class="size-3" />
                    </Button>
                    <Button size="icon-xs" variant="destructive" :disabled="formDisabled" aria-label="删除步骤" @click="removeStep(index)">
                      <Trash2 class="size-3" />
                    </Button>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="rounded-lg border border-dashed border-border/80 px-4 py-10 text-center text-sm font-bold text-muted-foreground">
              暂无步骤
            </div>
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
                    <option v-for="step in previewableSteps" :key="step.client_id" :value="normalizeKey(step.step_key)">
                      {{ step.name || step.step_key }}
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
                <span>{{ previewResult.step?.name || previewStepKey }} / {{ previewResult.total }} products</span>
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
                      <Badge variant="outline">{{ product.product_type?.name || product.product_type?.slug || 'type' }}</Badge>
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
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  ArrowDown,
  ArrowUp,
  ChevronRight,
  CircleAlert,
  Info,
  Layers3,
  ListChecks,
  Lock,
  Plus,
  RefreshCw,
  Rocket,
  Save,
  Search,
  ShieldCheck,
  Trash2,
  Zap,
} from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import quickBuyApi, {
  type QuickBuyFlow,
  type QuickBuyFlowPayload,
  type QuickBuyFlowSummary,
  type QuickBuyPreviewProduct,
  type QuickBuyPreviewResult,
  type QuickBuySelectionMode,
  type QuickBuyStep,
  type QuickBuyValidationResult,
  type QuickBuyVersionPayload,
} from '@/api/quickBuy'
import productTypeApi from '@/api/productTypes'
import type { ProductTypeRecord } from '@/components/admin/product/productTypeTypes'
import { useAuthStore } from '@/stores/auth'

interface StepForm {
  client_id: number
  step_key: string
  name: string
  description: string
  help_text: string
  sort_order: number | string
  selection_mode: QuickBuySelectionMode
  is_required: boolean
  min_select: number | string
  max_select: number | string
  default_quantity: number | string
  allow_skip: boolean
  product_type_ids: number[]
}

const flows = ref<QuickBuyFlowSummary[]>([])
const selectedFlow = ref<QuickBuyFlow | null>(null)
const productTypes = ref<ProductTypeRecord[]>([])
const activeFlowId = ref<number | null>(null)
const loading = ref(false)
const loadingFlow = ref(false)
const saving = ref(false)
const publishing = ref(false)
const validating = ref(false)
const validationResult = ref<QuickBuyValidationResult | null>(null)
const stepForms = ref<StepForm[]>([])
const previewing = ref(false)
const previewStepKey = ref('')
const previewKeyword = ref('')
const previewResult = ref<QuickBuyPreviewResult | null>(null)
const authStore = useAuthStore()
let nextStepClientID = 1

const flowForm = reactive({
  slug: '',
  name: '',
  description: '',
  entry_surface: 'dock',
  is_enabled: true,
  sort_order: 100 as number | string,
})

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
const canSave = computed(() => flowForm.slug.trim() !== '' && flowForm.name.trim() !== '')
const canEdit = computed(() => authStore.hasPermission('product:edit'))
const formDisabled = computed(() => !canEdit.value || loadingFlow.value || saving.value || publishing.value)
const validationIssues = computed(() => validationResult.value?.issues || [])
const previewableSteps = computed(() => stepForms.value.filter((step) => step.selection_mode !== 'auto' && normalizeKey(String(step.step_key || ''))))
const previewProducts = computed(() => previewResult.value?.products || [])
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
  { key: 'product-types', label: 'Product types', value: productTypes.value.length, icon: Layers3, tone: 'gray' },
])

const productTypeId = (productType: ProductTypeRecord) => Number(productType.id)

const normalizeKey = (value: string) => value
  .trim()
  .toLowerCase()
  .replace(/\s+/g, '-')
  .replace(/[^a-z0-9_-]/g, '')

const toPositiveNumber = (value: number | string, fallback: number) => {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) && numberValue > 0 ? numberValue : fallback
}

const toNonNegativeNumber = (value: number | string, fallback: number) => {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) && numberValue >= 0 ? numberValue : fallback
}

const draftVersion = (flow: QuickBuyFlowSummary) => (flow.versions || []).find((version) => version.status === 'draft')
const publishedVersion = (flow: QuickBuyFlowSummary) => (flow.versions || []).find((version) => version.status === 'published')

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

const emptyStep = (index = stepForms.value.length): StepForm => ({
  client_id: nextStepClientID++,
  step_key: `step-${index + 1}`,
  name: `Step ${index + 1}`,
  description: '',
  help_text: '',
  sort_order: (index + 1) * 10,
  selection_mode: 'single',
  is_required: true,
  min_select: 0,
  max_select: 1,
  default_quantity: 1,
  allow_skip: false,
  product_type_ids: [],
})

const resetFlowForm = () => {
  Object.assign(flowForm, {
    slug: 'quick-build',
    name: 'QUICK Build',
    description: '',
    entry_surface: 'dock',
    is_enabled: true,
    sort_order: 100,
  })
  stepForms.value = []
  previewResult.value = null
  syncPreviewStepKey()
}

const hydrateFlowForm = (flow: QuickBuyFlow) => {
  Object.assign(flowForm, {
    slug: flow.slug || '',
    name: flow.name || '',
    description: flow.description || '',
    entry_surface: flow.entry_surface || 'dock',
    is_enabled: flow.is_enabled !== false,
    sort_order: flow.sort_order || 100,
  })
  stepForms.value = (flow.steps || []).map((step, index) => stepToForm(step, index))
  previewResult.value = null
  syncPreviewStepKey()
}

const stepToForm = (step: QuickBuyStep, index: number): StepForm => ({
  client_id: nextStepClientID++,
  step_key: step.step_key || step.slug || `step-${index + 1}`,
  name: step.name || `Step ${index + 1}`,
  description: step.description || '',
  help_text: step.help_text || '',
  sort_order: step.sort_order || (index + 1) * 10,
  selection_mode: step.selection_mode || 'single',
  is_required: step.is_required !== false,
  min_select: step.min_select || 0,
  max_select: step.max_select || 1,
  default_quantity: step.default_quantity || 1,
  allow_skip: Boolean(step.allow_skip),
  product_type_ids: (step.product_types || []).map((productType) => Number(productType.id)).filter(Boolean),
})

const buildVersionPayload = (): QuickBuyVersionPayload => ({
  starts_at: null,
  ends_at: null,
  steps: stepForms.value.map((step, index) => ({
    step_key: normalizeKey(step.step_key) || `step-${index + 1}`,
    name: step.name.trim() || `Step ${index + 1}`,
    description: step.description.trim(),
    help_text: step.help_text.trim(),
    sort_order: toPositiveNumber(step.sort_order, (index + 1) * 10),
    selection_mode: step.selection_mode || 'single',
    is_required: step.is_required,
    min_select: toNonNegativeNumber(step.min_select, 0),
    max_select: step.selection_mode === 'auto' ? 0 : toNonNegativeNumber(step.max_select, 1),
    default_quantity: toPositiveNumber(step.default_quantity, 1),
    allow_skip: step.allow_skip,
    product_type_ids: step.product_type_ids,
  })),
})

const buildFlowPayload = (): QuickBuyFlowPayload => ({
  slug: normalizeKey(flowForm.slug),
  name: flowForm.name.trim(),
  description: flowForm.description.trim(),
  entry_surface: normalizeKey(flowForm.entry_surface) || 'dock',
  is_enabled: flowForm.is_enabled,
  sort_order: toPositiveNumber(flowForm.sort_order, 100),
  version: buildVersionPayload(),
})

const loadProductTypes = async () => {
  productTypes.value = await productTypeApi.list({ include_disabled: true })
}

const loadFlows = async () => {
  flows.value = await quickBuyApi.listFlows()
}

const reload = async () => {
  loading.value = true
  try {
    await Promise.all([loadProductTypes(), loadFlows()])
    if (activeFlowId.value) {
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
  if (!canSave.value) {
    toast.error('请填写 flow slug 和名称')
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
      await quickBuyApi.updateFlow(selectedFlow.value.id, payload)
      if (selectedFlow.value.version?.status === 'draft' && selectedFlow.value.version.id) {
        flow = await quickBuyApi.updateDraftVersion(selectedFlow.value.version.id, payload.version)
      } else {
        flow = await quickBuyApi.createDraftVersion(selectedFlow.value.id, payload.version)
      }
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

const removeStep = (index: number) => {
  if (!canEdit.value) return
  validationResult.value = null
  previewResult.value = null
  stepForms.value.splice(index, 1)
  syncPreviewStepKey()
}

const moveStep = (index: number, direction: -1 | 1) => {
  if (!canEdit.value) return
  const target = index + direction
  if (target < 0 || target >= stepForms.value.length) return
  const [item] = stepForms.value.splice(index, 1)
  if (!item) return
  validationResult.value = null
  previewResult.value = null
  stepForms.value.splice(target, 0, item)
  stepForms.value.forEach((step, stepIndex) => {
    step.sort_order = (stepIndex + 1) * 10
  })
}

const toggleStepProductType = (step: StepForm, productTypeID: number) => {
  if (!canEdit.value) return
  if (!productTypeID) return
  validationResult.value = null
  previewResult.value = null
  const index = step.product_type_ids.indexOf(productTypeID)
  if (index >= 0) {
    step.product_type_ids.splice(index, 1)
  } else {
    step.product_type_ids.push(productTypeID)
  }
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
