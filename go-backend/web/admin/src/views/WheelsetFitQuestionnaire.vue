<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="轮组选型问卷"
      description="维护固定顺序问题、多语言内容、帮助说明和商品筛选规则"
    >
      <template #actions>
        <Button variant="outline" :disabled="loading" @click="reload">
 <RefreshCw class="size-4" />
          刷新
        </Button>
        <Button v-if="showCreateDraft" variant="outline" :disabled="creatingDraft || !canEdit" @click="openCreateDialog">
 <FilePlus2 class="size-4" />
          {{ createDraftButtonLabel }}
        </Button>
        <Button variant="outline" :disabled="validating || !currentVersion" @click="validateCurrentVersion()">
 <ShieldCheck class="size-4" />
          校验
        </Button>
        <Button :disabled="!canEdit || !isDraft || publishing || !currentVersion" @click="requestPublish">
 <Rocket class="size-4" />
          发布
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

 <Alert v-if="!canEdit" class="rounded-lg">
 <Lock class="size-4" />
      <AlertTitle>只读模式</AlertTitle>
      <AlertDescription>当前账号可以查看和校验问卷，但没有修改或发布权限。</AlertDescription>
    </Alert>
 <Alert v-else-if="currentVersion && !isDraft" class="rounded-lg">
 <Info class="size-4" />
      <AlertTitle>当前显示已发布版本</AlertTitle>
      <AlertDescription>点击“开始编辑”或列表中的编辑操作会自动创建草稿，已发布内容会保持不变。</AlertDescription>
    </Alert>
 <Alert v-if="validationResult" :variant="validationResult.valid ? 'default': 'destructive'" class="rounded-lg">
 <ShieldCheck v-if="validationResult.valid" class="size-4" />
 <CircleAlert v-else class="size-4" />
      <AlertTitle>{{ validationResult.valid ? '校验通过' : '校验未通过' }}</AlertTitle>
      <AlertDescription>
 <div v-if="validationResult.issues.length" class="mt-1 space-y-1">
 <p v-for="issue in validationResult.issues" :key="validationIssueKey(issue)" class="text-xs font-bold">
            [{{ issue.severity }}] {{ issue.message }}
          </p>
        </div>
        <span v-else>当前版本可以发布。</span>
      </AlertDescription>
    </Alert>

    <AdminTablePanel :loading="loading">
      <template #header>
 <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
 <h2 class="text-sm font-black uppercase tracking-tight">问题列表</h2>
 <p class="text-[11px] font-bold text-muted-foreground">
              <template v-if="currentVersion">版本 {{ currentVersion.version_number }} · {{ versionStatusLabel }}</template>
              <template v-else>尚未创建问卷</template>
            </p>
          </div>
          <Button size="sm" variant="outline" :disabled="controlsDisabled" @click="openCreateDialog">
 <Plus class="size-3.5" />
            新增问题
          </Button>
        </div>
      </template>

 <div v-if="questions.length" class="divide-y divide-dashed divide-border/70">
 <article v-for="(question, index) in questions" :key="question.id" class="grid gap-3 px-4 py-4 sm:grid-cols-[auto_minmax(0,1fr)_auto] sm:items-center">
 <span class="grid size-8 place-items-center rounded-full bg-muted text-sm font-black tabular-nums">{{ index + 1 }}</span>
 <div class="min-w-0">
 <div class="flex flex-wrap items-center gap-2">
 <h3 class="truncate text-sm font-black">{{ questionPrompt(question) || question.question_key }}</h3>
              <Badge :variant="question.is_enabled ? 'secondary' : 'outline'">{{ question.is_enabled ? '启用' : '停用' }}</Badge>
              <Badge variant="outline">{{ question.is_required ? '必答' : '可跳过' }}</Badge>
              <Badge v-if="question.allow_unknown" variant="outline">允许不确定</Badge>
            </div>
 <p class="mt-1 truncate font-mono text-[11px] font-bold text-muted-foreground">
              {{ question.question_key }} / {{ question.answer_key }} · {{ question.options.length }} options
            </p>
          </div>
 <div class="flex items-center justify-end gap-1">
            <Tooltip>
              <TooltipTrigger as-child>
                <Button size="icon-sm" variant="ghost" :disabled="controlsDisabled || moving || index === 0" aria-label="上移问题" @click="moveQuestion(index, -1)">
 <ChevronUp class="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>上移问题</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger as-child>
                <Button size="icon-sm" variant="ghost" :disabled="controlsDisabled || moving || index === questions.length - 1" aria-label="下移问题" @click="moveQuestion(index, 1)">
 <ChevronDown class="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>下移问题</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger as-child>
                <Button size="icon-sm" variant="ghost" :disabled="controlsDisabled" aria-label="编辑问题" @click="openEditDialog(question)">
 <Pencil class="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>编辑问题</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger as-child>
 <Button size="icon-sm" variant="ghost" class="text-destructive hover:text-destructive" :disabled="controlsDisabled" aria-label="删除问题" @click="requestDelete(question)">
 <Trash2 class="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>删除问题</TooltipContent>
            </Tooltip>
          </div>
        </article>
      </div>
 <div v-else-if="!loading" class="px-4 py-14 text-center text-sm font-bold text-muted-foreground">
        <template v-if="currentVersion">当前版本还没有问题</template>
        <template v-else>还没有创建问卷，点击上方按钮开始初始化。</template>
      </div>
    </AdminTablePanel>

    <WheelsetFitQuestionEditorDialog
      v-if="editorForm"
      v-model:open="editorOpen"
      :mode="editorMode"
      :form="editorForm"
      :language-options="languageOptions"
      :source-locale="sourceLocale"
      :product-filter-options="productFilterOptions"
      :product-filter-options-loading="productFilterOptionsLoading"
      :question-key-options="questionKeyOptions"
      :answer-key-options="answerKeyOptions"
      :selection-configuration-key-options-loading="selectionConfigurationKeyOptionsLoading"
      :disabled="controlsDisabled"
      :saving="savingQuestion"
      @submit="saveQuestion"
    />

    <AdminConfirmDialog
      :open="!!deleteTarget"
      title="删除问题"
      :description="deleteTarget ? `将从当前草稿移除“${questionPrompt(deleteTarget) || deleteTarget.question_key}”。` : ''"
      confirm-label="删除"
      destructive
      @update:open="handleDeleteDialog"
      @confirm="deleteQuestion"
    />
    <AdminConfirmDialog
      :open="publishConfirmOpen"
      title="发布问卷"
      description="发布后将替换当前前台使用的问卷版本。"
      confirm-label="发布"
      @update:open="publishConfirmOpen = $event"
      @confirm="publishVersion"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  ChevronDown,
  ChevronUp,
  CircleAlert,
  FilePlus2,
  FileText,
  Info,
  Languages,
  ListChecks,
  Lock,
  Pencil,
  Plus,
  RefreshCw,
  Rocket,
  ShieldCheck,
  Trash2,
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import WheelsetFitQuestionEditorDialog from '@/components/admin/wheelset-fit/WheelsetFitQuestionEditorDialog.vue'
import selectionConfigurationKeyApi, { type SelectionConfigurationKeyOption } from '@/api/selectionConfigurationKeys'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import wheelsetFitQuestionnaireApi, {
  type WheelsetFitQuestion,
  type WheelsetFitProductFilterOption,
  type WheelsetFitQuestionnaireValidationIssue,
  type WheelsetFitQuestionnaireValidationResult,
  type WheelsetFitQuestionnaireVersion,
  type WheelsetFitQuestionPayload,
} from '@/api/wheelsetFitQuestionnaire'
import { useSupportedLanguages } from '@/composables/useSupportedLanguages'
import {
  selectionConfigurationKeyKindAnswerKey,
  selectionConfigurationKeyKindQuestionKey,
} from '@/modules/selection-configuration/selectionConfigurationKeys'
import type { WheelsetFitQuestionForm, WheelsetFitQuestionOptionForm } from '@/modules/wheelset-fit/questionnaire'
import { useAuthStore } from '@/stores/auth'

const supportedLanguages = useSupportedLanguages()
const languageOptions = supportedLanguages.languageOptions
const authStore = useAuthStore()

const currentVersion = ref<WheelsetFitQuestionnaireVersion | null>(null)
const validationResult = ref<WheelsetFitQuestionnaireValidationResult | null>(null)
const loading = ref(false)
const productFilterOptionsLoading = ref(false)
const productFilterOptions = ref<WheelsetFitProductFilterOption[]>([])
const selectionConfigurationKeyOptionsLoading = ref(false)
const questionKeyOptions = ref<SelectionConfigurationKeyOption[]>([])
const answerKeyOptions = ref<SelectionConfigurationKeyOption[]>([])
const creatingDraft = ref(false)
const savingQuestion = ref(false)
const moving = ref(false)
const validating = ref(false)
const publishing = ref(false)
const editorOpen = ref(false)
const editorMode = ref<'create' | 'edit'>('create')
const editorForm = ref<WheelsetFitQuestionForm | null>(null)
const deleteTarget = ref<WheelsetFitQuestion | null>(null)
const publishConfirmOpen = ref(false)
let nextClientID = 1

const canEdit = computed(() => authStore.hasPermission('product:edit'))
const sourceLocale = computed(() => currentVersion.value?.questionnaire?.source_locale || 'zh_cn')
const isDraft = computed(() => currentVersion.value?.status === 'draft')
const controlsDisabled = computed(() => !canEdit.value || creatingDraft.value)
const showCreateDraft = computed(() => !isDraft.value)
const createDraftButtonLabel = computed(() => (currentVersion.value ? '开始编辑' : '初始化问卷'))
const questions = computed(() => (
  [...(currentVersion.value?.questions || [])].sort((left, right) => (
    left.sort_order - right.sort_order || left.id - right.id
  ))
))
const versionStatusLabel = computed(() => {
  if (!currentVersion.value) return '尚未创建'
  if (currentVersion.value.status === 'draft') return '草稿'
  if (currentVersion.value.status === 'published') return '已发布'
  return currentVersion.value.status
})
const statItems = computed(() => [
  { key: 'questions', label: '问题', value: questions.value.length, icon: ListChecks, tone: 'blue' },
  {
    key: 'enabled',
    label: '启用',
    value: questions.value.filter((question) => question.is_enabled).length,
    icon: FileText,
    tone: 'green',
  },
  {
    key: 'options',
    label: '选项',
    value: questions.value.reduce((total, question) => total + question.options.length, 0),
    icon: ListChecks,
    tone: 'amber',
  },
  { key: 'languages', label: '语言', value: languageOptions.value.length, icon: Languages, tone: 'gray' },
])

const questionPrompt = (question: WheelsetFitQuestion): string => (
  question.translations.find((translation) => translation.locale === sourceLocale.value)?.prompt || ''
)

const validationIssueKey = (issue: WheelsetFitQuestionnaireValidationIssue) => (
  [issue.severity, issue.code, issue.question_id, issue.question_key, issue.option_key, issue.locale, issue.message].join(':')
)

const extractErrorMessage = (error: unknown, fallback: string): string => {
  const details = error as { response?: { data?: { error?: string, message?: string } }, message?: string }
  return details.response?.data?.error || details.response?.data?.message || details.message || fallback
}

const normalizedLocaleCodes = (): string[] => {
  const codes = [sourceLocale.value, ...languageOptions.value.map((language) => language.value)]
  return [...new Set(codes.filter(Boolean))]
}

const createTranslations = (
  existing: Array<{ locale: string, prompt?: string, help_title?: string, help_body?: string }>,
) => {
  const byLocale = new Map(existing.map((translation) => [translation.locale, translation]))
  return normalizedLocaleCodes().map((locale) => {
    const source = byLocale.get(locale)
    return {
      locale,
      prompt: source?.prompt || '',
      help_title: source?.help_title || '',
      help_body: source?.help_body || '',
    }
  })
}

const createOptionTranslations = (
  existing: Array<{ locale: string, label?: string, description?: string }>,
) => {
  const byLocale = new Map(existing.map((translation) => [translation.locale, translation]))
  return normalizedLocaleCodes().map((locale) => {
    const source = byLocale.get(locale)
    return {
      locale,
      label: source?.label || '',
      description: source?.description || '',
    }
  })
}

const filterEffectsJSON = (value: unknown): string => {
  try {
    return JSON.stringify(value && typeof value === 'object' ? value : {}, null, 2)
  } catch {
    return '{}'
  }
}

const createQuestionForm = (question?: WheelsetFitQuestion): WheelsetFitQuestionForm => ({
  id: question?.id,
  question_key: question?.question_key || '',
  answer_key: question?.answer_key || '',
  sort_order: question?.sort_order || (questions.value.length + 1) * 10,
  input_mode: 'single_choice',
  is_required: question?.is_required ?? true,
  allow_unknown: question?.allow_unknown ?? true,
  is_enabled: question?.is_enabled ?? true,
  translations: createTranslations(question?.translations || []),
  options: (question?.options || []).map((option, index) => ({
    client_id: option.id || Date.now() + nextClientID++,
    option_key: option.option_key,
    answer_value: option.answer_value,
    sort_order: option.sort_order || (index + 1) * 10,
    is_unknown: option.is_unknown,
    is_enabled: option.is_enabled,
    product_filter_effects_json: filterEffectsJSON(option.product_filter_effects),
    translations: createOptionTranslations(option.translations),
  })),
})

const parseProductFilterEffects = (raw: string, optionKey: string): Record<string, unknown> => {
  const value = raw.trim() || '{}'
  let parsed: unknown
  try {
    parsed = JSON.parse(value)
  } catch {
    throw new Error(`选项 ${optionKey || 'new_option'} 的商品筛选规则不是有效 JSON`)
  }
  if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
    throw new Error(`选项 ${optionKey || 'new_option'} 的商品筛选规则必须是 JSON 对象`)
  }
  return parsed as Record<string, unknown>
}

const buildQuestionPayload = (form: WheelsetFitQuestionForm): WheelsetFitQuestionPayload => ({
  question_key: form.question_key.trim(),
  answer_key: form.answer_key.trim(),
  sort_order: form.sort_order,
  input_mode: 'single_choice',
  is_required: form.is_required,
  allow_unknown: form.allow_unknown,
  is_enabled: form.is_enabled,
  translations: form.translations.map((translation) => ({
    locale: translation.locale,
    prompt: translation.prompt,
    help_title: translation.help_title,
    help_body: translation.help_body,
  })),
  options: form.options.map((option: WheelsetFitQuestionOptionForm) => ({
    option_key: option.option_key.trim(),
    answer_value: option.answer_value.trim(),
    sort_order: option.sort_order,
    is_unknown: option.is_unknown,
    is_enabled: option.is_enabled,
    product_filter_effects: parseProductFilterEffects(option.product_filter_effects_json, option.option_key),
    translations: option.translations.map((translation) => ({
      locale: translation.locale,
      label: translation.label,
      description: translation.description,
    })),
  })),
})

const assertQuestionForm = (form: WheelsetFitQuestionForm) => {
  const keyPattern = /^[a-z0-9]+(?:_[a-z0-9]+)*$/
  if (!keyPattern.test(form.question_key.trim())) {
    throw new Error('问题 Key 必须使用小写 snake_case')
  }
  if (!keyPattern.test(form.answer_key.trim())) {
    throw new Error('回答 Key 必须使用小写 snake_case')
  }
  const source = form.translations.find((translation) => translation.locale === sourceLocale.value)
  if (!source?.prompt.trim()) {
    throw new Error('基础语言的问题标题不能为空')
  }
  const optionKeys = new Set<string>()
  let unknownOptions = 0
  for (const option of form.options) {
    const optionKey = option.option_key.trim()
    if (!keyPattern.test(optionKey)) {
      throw new Error('选项 Key 必须使用小写 snake_case')
    }
    if (optionKeys.has(optionKey)) {
      throw new Error(`选项 Key ${optionKey} 重复`)
    }
    optionKeys.add(optionKey)
    if (!option.answer_value.trim()) {
      throw new Error(`选项 ${optionKey} 的回答值不能为空`)
    }
    if (option.is_unknown) unknownOptions++
    if (option.is_unknown && !form.allow_unknown) {
      throw new Error('未开启“允许不确定”时不能设置不确定选项')
    }
    const optionSource = option.translations.find((translation) => translation.locale === sourceLocale.value)
    if (option.is_enabled && !optionSource?.label.trim()) {
      throw new Error(`选项 ${optionKey} 的基础语言名称不能为空`)
    }
    parseProductFilterEffects(option.product_filter_effects_json, optionKey)
  }
  if (form.is_enabled && !form.options.some((option) => option.is_enabled)) {
    throw new Error('启用的问题至少需要一个启用选项')
  }
  if (unknownOptions > 1) {
    throw new Error('每个问题只能有一个不确定选项')
  }
}

const hydrateVersion = (version: WheelsetFitQuestionnaireVersion | null) => {
  currentVersion.value = version
  validationResult.value = null
}

const reload = async () => {
  loading.value = true
  try {
    hydrateVersion(await wheelsetFitQuestionnaireApi.getCurrentVersion())
  } catch (error) {
    toast.error(extractErrorMessage(error, '问卷加载失败'))
  } finally {
    loading.value = false
  }
}

const createDraft = async () => {
  if (!canEdit.value) return
  if (isDraft.value) return currentVersion.value
  const isInitializingEmptyQuestionnaire = !currentVersion.value
  creatingDraft.value = true
  try {
    hydrateVersion(await wheelsetFitQuestionnaireApi.createDraft())
    toast.success(isInitializingEmptyQuestionnaire ? '已初始化空白问卷' : '已创建可编辑草稿')
    return currentVersion.value
  } catch (error) {
    toast.error(extractErrorMessage(error, '创建草稿失败'))
    return null
  } finally {
    creatingDraft.value = false
  }
}

const loadProductFilterOptions = async () => {
  productFilterOptionsLoading.value = true
  try {
    productFilterOptions.value = await wheelsetFitQuestionnaireApi.getProductFilterOptions()
  } catch (error) {
    toast.error(extractErrorMessage(error, '轮组商品动态值加载失败'))
    productFilterOptions.value = []
  } finally {
    productFilterOptionsLoading.value = false
  }
}

const loadSelectionConfigurationKeyOptions = async () => {
  selectionConfigurationKeyOptionsLoading.value = true
  try {
    const [loadedQuestionKeyOptions, loadedAnswerKeyOptions] = await Promise.all([
      selectionConfigurationKeyApi.listEnabledKeyOptions(selectionConfigurationKeyKindQuestionKey),
      selectionConfigurationKeyApi.listEnabledKeyOptions(selectionConfigurationKeyKindAnswerKey),
    ])
    questionKeyOptions.value = loadedQuestionKeyOptions
    answerKeyOptions.value = loadedAnswerKeyOptions
  } catch (error) {
    questionKeyOptions.value = []
    answerKeyOptions.value = []
    toast.error(extractErrorMessage(error, '选型配置 Key 加载失败'))
  } finally {
    selectionConfigurationKeyOptionsLoading.value = false
  }
}

const ensureEditableDraft = async () => {
  if (!canEdit.value) {
    toast.error('当前账号没有 product:edit 权限')
    return null
  }
  if (isDraft.value) return currentVersion.value
  return createDraft()
}

const openCreateDialog = async () => {
  const version = await ensureEditableDraft()
  if (!version || controlsDisabled.value) return
  editorMode.value = 'create'
  editorForm.value = createQuestionForm()
  editorOpen.value = true
}

const openEditDialog = async (question: WheelsetFitQuestion) => {
  const version = await ensureEditableDraft()
  if (!version || controlsDisabled.value) return
  const editableQuestion = version.questions.find((item) => item.question_key === question.question_key)
    || version.questions.find((item) => item.id === question.id)
  if (!editableQuestion) {
    toast.error('无法找到要编辑的问题')
    return
  }
  editorMode.value = 'edit'
  editorForm.value = createQuestionForm(editableQuestion)
  editorOpen.value = true
}

const saveQuestion = async () => {
  const form = editorForm.value
  if (!form) return
  const version = await ensureEditableDraft()
  if (!version || controlsDisabled.value) return
  try {
    assertQuestionForm(form)
  } catch (error) {
    toast.error(extractErrorMessage(error, '请检查问题内容'))
    return
  }

  savingQuestion.value = true
  try {
    const payload = buildQuestionPayload(form)
    const version = form.id
      ? await wheelsetFitQuestionnaireApi.updateQuestion(form.id, payload)
      : await wheelsetFitQuestionnaireApi.createQuestion(payload)
    hydrateVersion(version)
    editorOpen.value = false
    toast.success(form.id ? '问题已保存' : '问题已添加')
  } catch (error) {
    toast.error(extractErrorMessage(error, '保存问题失败'))
  } finally {
    savingQuestion.value = false
  }
}

const moveQuestion = async (index: number, direction: number) => {
  if (controlsDisabled.value || moving.value) return
  const version = await ensureEditableDraft()
  if (!version || controlsDisabled.value) return
  const targetIndex = index + direction
  if (targetIndex < 0 || targetIndex >= questions.value.length) return
  const reordered = [...questions.value]
  const [question] = reordered.splice(index, 1)
  reordered.splice(targetIndex, 0, question)

  moving.value = true
  try {
    hydrateVersion(await wheelsetFitQuestionnaireApi.reorderQuestions(reordered.map((item) => item.id)))
  } catch (error) {
    toast.error(extractErrorMessage(error, '调整问题顺序失败'))
  } finally {
    moving.value = false
  }
}

const requestDelete = async (question: WheelsetFitQuestion) => {
  const version = await ensureEditableDraft()
  if (!version || controlsDisabled.value) return
  deleteTarget.value = version.questions.find((item) => item.question_key === question.question_key)
    || version.questions.find((item) => item.id === question.id)
    || null
}

const handleDeleteDialog = (open: boolean) => {
  if (!open) deleteTarget.value = null
}

const deleteQuestion = async () => {
  const target = deleteTarget.value
  if (!target || controlsDisabled.value) return
  try {
    hydrateVersion(await wheelsetFitQuestionnaireApi.deleteQuestion(target.id))
    toast.success('问题已删除')
  } catch (error) {
    toast.error(extractErrorMessage(error, '删除问题失败'))
  } finally {
    deleteTarget.value = null
  }
}

const validateCurrentVersion = async (options: { silent?: boolean } = {}) => {
  if (!currentVersion.value) return null
  validating.value = true
  try {
    const result = await wheelsetFitQuestionnaireApi.validateVersion(currentVersion.value.id)
    validationResult.value = result
    if (!options.silent) {
      toast[result.valid ? 'success' : 'error'](result.valid ? '校验通过' : '校验未通过')
    }
    return result
  } catch (error) {
    toast.error(extractErrorMessage(error, '问卷校验失败'))
    return null
  } finally {
    validating.value = false
  }
}

const requestPublish = async () => {
  if (!canEdit.value || !isDraft.value) return
  const result = await validateCurrentVersion({ silent: true })
  if (!result?.valid) {
    toast.error('请先处理校验错误')
    return
  }
  publishConfirmOpen.value = true
}

const publishVersion = async () => {
  if (!currentVersion.value || !canEdit.value || !isDraft.value) return
  publishing.value = true
  try {
    hydrateVersion(await wheelsetFitQuestionnaireApi.publishVersion(currentVersion.value.id))
    publishConfirmOpen.value = false
    toast.success('问卷已发布')
  } catch (error) {
    toast.error(extractErrorMessage(error, '发布问卷失败'))
  } finally {
    publishing.value = false
  }
}

onMounted(async () => {
  await supportedLanguages.fetchLanguages()
  await loadSelectionConfigurationKeyOptions()
  await loadProductFilterOptions()
  await reload()
})
</script>
