<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="xl" class="max-h-[calc(100dvh-1rem)] overflow-y-auto p-4 sm:p-5" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '新增问题' : '编辑问题' }}</DialogTitle>
          <DialogDescription>单选题按固定顺序展示；内容按语言分别维护。</DialogDescription>
        </DialogHeader>

        <section class="grid gap-3 border-t border-dashed border-border/70 pt-4 sm:grid-cols-2">
          <AdminFormField label="问题 Key" required description="从选型配置 Key 管理中选择启用项">
            <Select v-model="form.question_key" :disabled="disabled || selectionConfigurationKeyOptionsLoading">
              <SelectTrigger class="w-full font-mono">
                <SelectValue placeholder="选择问题 Key" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="option in questionKeySelectionOptions" :key="option.code" :value="option.code">
                  {{ option.display_label }} · {{ option.code }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>
          <AdminFormField label="回答 Key" required description="从选型配置 Key 管理中选择启用项">
            <Select v-model="form.answer_key" :disabled="disabled || selectionConfigurationKeyOptionsLoading">
              <SelectTrigger class="w-full font-mono">
                <SelectValue placeholder="选择回答 Key" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="option in answerKeySelectionOptions" :key="option.code" :value="option.code">
                  {{ option.display_label }} · {{ option.code }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

 <div class="flex flex-wrap gap-x-6 gap-y-3 sm:col-span-2">
 <label class="inline-flex items-center gap-2 text-sm font-bold">
              <Switch v-model="form.is_required" :disabled="disabled" />
              必答
            </label>
 <label class="inline-flex items-center gap-2 text-sm font-bold">
              <Switch v-model="form.allow_unknown" :disabled="disabled" />
              允许不确定
            </label>
 <label class="inline-flex items-center gap-2 text-sm font-bold">
              <Switch v-model="form.is_enabled" :disabled="disabled" />
              启用问题
            </label>
          </div>
        </section>

 <section class="space-y-3 border-t border-dashed border-border/70 pt-4">
 <div class="flex flex-wrap items-center justify-between gap-3">
 <h3 class="text-sm font-black tracking-tight">问题文案</h3>
 <select v-model="activeLocale" :class="selectClass" :disabled="disabled">
              <option v-for="language in languageOptions" :key="language.value" :value="language.value">
                {{ language.label }}
              </option>
            </select>
          </div>

 <div class="grid gap-3">
            <div class="grid gap-3 sm:grid-cols-2">
              <AdminFormField label="问题标题" :required="activeLocale === sourceLocale">
                <Input v-model="currentQuestionTranslation.prompt" placeholder="请输入问题标题" :disabled="disabled" />
              </AdminFormField>
              <AdminFormField label="HELP 标题">
                <Input v-model="currentQuestionTranslation.help_title" placeholder="为什么需要这个？" :disabled="disabled" />
              </AdminFormField>
            </div>
            <AdminFormField label="HELP 内容">
 <Textarea v-model="currentQuestionTranslation.help_body" class="min-h-20 resize-y" placeholder="请输入帮助说明" :disabled="disabled" />
            </AdminFormField>
          </div>
        </section>

 <section class="space-y-3 border-t border-dashed border-border/70 pt-4">
 <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
 <h3 class="text-sm font-black tracking-tight">选项</h3>
 <p class="text-[11px] font-bold text-muted-foreground">{{ form.options.length }} 个选项</p>
            </div>
            <Button type="button" size="sm" variant="outline" :disabled="disabled" @click="addOption">
 <Plus class="size-3.5" />
              新增选项
            </Button>
          </div>

 <div v-if="form.options.length" class="divide-y divide-dashed divide-border/70 border-y border-dashed border-border/70">
 <section v-for="(option, index) in form.options" :key="option.client_id" class="space-y-3 py-4">
 <div class="flex flex-wrap items-center justify-between gap-3">
 <div class="flex min-w-0 items-center gap-2">
 <span class="grid size-6 shrink-0 place-items-center rounded-full bg-muted text-[11px] font-black">{{ index + 1 }}</span>
 <span class="truncate font-mono text-xs font-black">{{ option.option_key || 'new_option'}}</span>
                </div>
 <div class="flex items-center gap-1">
                  <Tooltip>
                    <TooltipTrigger as-child>
                      <Button type="button" size="icon-sm" variant="ghost" :disabled="disabled || index === 0" aria-label="上移选项" @click="moveOption(index, -1)">
 <ChevronUp class="size-4" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>上移选项</TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger as-child>
                      <Button type="button" size="icon-sm" variant="ghost" :disabled="disabled || index === form.options.length - 1" aria-label="下移选项" @click="moveOption(index, 1)">
 <ChevronDown class="size-4" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>下移选项</TooltipContent>
                  </Tooltip>
                  <Tooltip>
                    <TooltipTrigger as-child>
 <Button type="button" size="icon-sm" variant="ghost" class="text-destructive hover:text-destructive" :disabled="disabled" aria-label="删除选项" @click="removeOption(index)">
 <Trash2 class="size-4" />
                      </Button>
                    </TooltipTrigger>
                    <TooltipContent>删除选项</TooltipContent>
                  </Tooltip>
                </div>
              </div>

              <div class="space-y-3">
                <div class="rounded-lg border border-border/60 bg-muted/20 px-3 py-2">
                  <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
                    前台展示文案 · 当前语言
                  </p>
                  <p class="mt-1 text-xs font-bold text-muted-foreground">
                    用户在前台看到的是下面的名称和说明；切换右上角语言后，分别维护对应语言的文字。
                  </p>
                </div>

                <div class="grid gap-3 sm:grid-cols-2">
                  <AdminFormField
                    label="前台选项名称"
                    :required="activeLocale === sourceLocale"
                    description="用户实际看到的选项文字，例如：山地车"
                  >
                    <Input v-model="optionTranslation(option).label" placeholder="请输入前台显示名称" :disabled="disabled" />
                  </AdminFormField>
                  <AdminFormField
                    label="前台选项说明"
                    description="显示在选项名称下方，可选"
                  >
                    <Input v-model="optionTranslation(option).description" placeholder="可选" :disabled="disabled" />
                  </AdminFormField>
                </div>

                <details class="rounded-lg border border-dashed border-border/70 px-3 py-2">
                  <summary class="flex cursor-pointer list-none items-center justify-between gap-3 text-left [&::-webkit-details-marker]:hidden">
                    <span class="text-xs font-black text-muted-foreground/70">系统字段 · 不展示给用户</span>
                    <ChevronRight class="size-4 shrink-0 text-muted-foreground" />
                  </summary>
                  <div class="mt-3 space-y-3">
                    <p class="text-xs font-bold text-muted-foreground">
                      这两个值用于系统识别选项、生成答案和执行商品筛选；即使前台名称按语言变化，也应保持稳定。
                    </p>
                    <div class="grid gap-3 sm:grid-cols-2">
                      <AdminFormField
                        label="内部选项 Key"
                        required
                        description="系统唯一标识，使用小写 snake_case，例如 boost_148"
                      >
                        <Input v-model="option.option_key" class="font-mono" placeholder="boost_148" :disabled="disabled" />
                      </AdminFormField>
                      <AdminFormField
                        label="系统回答值"
                        required
                        description="提交到选型结果和客服数据的标准值，例如 mtb"
                      >
                        <Input v-model="option.answer_value" class="font-mono" placeholder="mtb" :disabled="disabled" />
                      </AdminFormField>
                    </div>
                  </div>
                </details>
              </div>

              <AdminFormField label="商品筛选规则" description="从轮组商品真实动态值里选择，自动生成规则。">
                <WheelsetFitProductFilterBuilder
                  v-model="option.product_filter_effects_json"
                  :filter-options="productFilterOptions"
                  :loading="productFilterOptionsLoading"
                  :disabled="disabled"
                />
              </AdminFormField>

 <div class="flex flex-wrap gap-x-6 gap-y-3">
 <label class="inline-flex items-center gap-2 text-sm font-bold">
                  <Switch v-model="option.is_unknown" :disabled="disabled || !form.allow_unknown" />
                  不确定选项
                </label>
 <label class="inline-flex items-center gap-2 text-sm font-bold">
                  <Switch v-model="option.is_enabled" :disabled="disabled" />
                  启用选项
                </label>
              </div>
            </section>
          </div>
 <div v-else class="border-y border-dashed border-border/70 py-8 text-center text-sm font-bold text-muted-foreground">
            暂无选项
          </div>
        </section>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="disabled || saving">
 <Save class="size-4" />
            {{ saving ? '保存中' : '保存问题' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ChevronDown, ChevronRight, ChevronUp, Plus, Save, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import WheelsetFitProductFilterBuilder from '@/components/admin/wheelset-fit/WheelsetFitProductFilterBuilder.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import type { SelectionConfigurationKeyOption } from '@/api/selectionConfigurationKeys'
import type { WheelsetFitProductFilterOption } from '@/api/wheelsetFitQuestionnaire'
import type { WheelsetFitQuestionForm, WheelsetFitQuestionOptionForm } from '@/modules/wheelset-fit/questionnaire'

interface LanguageOption {
  label: string
  value: string
}

const props = withDefaults(defineProps<{
  open?: boolean
  mode?: 'create' | 'edit'
  form: WheelsetFitQuestionForm
  languageOptions: LanguageOption[]
  sourceLocale: string
  productFilterOptions: WheelsetFitProductFilterOption[]
  productFilterOptionsLoading?: boolean
  questionKeyOptions?: SelectionConfigurationKeyOption[]
  answerKeyOptions?: SelectionConfigurationKeyOption[]
  selectionConfigurationKeyOptionsLoading?: boolean
  disabled?: boolean
  saving?: boolean
}>(), {
  open: false,
  mode: 'create',
  productFilterOptions: () => [],
  productFilterOptionsLoading: false,
  questionKeyOptions: () => [],
  answerKeyOptions: () => [],
  selectionConfigurationKeyOptionsLoading: false,
  disabled: false,
  saving: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit'): void
}>()

const activeLocale = ref(props.sourceLocale)
let nextClientID = 1

const selectClass = 'h-9 min-w-44 rounded-md border border-input bg-background px-3 text-xs font-bold outline-none transition focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 disabled:cursor-not-allowed disabled:opacity-50'

const currentQuestionTranslation = computed(() => {
  const existing = props.form.translations.find((translation) => translation.locale === activeLocale.value)
  if (existing) return existing
  return props.form.translations[0]
})

const buildSelectionConfigurationKeyOptions = (
  options: SelectionConfigurationKeyOption[],
  currentValue: string,
) => {
  const normalizedCurrentValue = currentValue.trim()
  if (!normalizedCurrentValue) return options
  if (options.some((option) => option.code === normalizedCurrentValue)) return options
  return [
    {
      id: 0,
      code: normalizedCurrentValue,
      display_label: `${normalizedCurrentValue}（当前值）`,
    },
    ...options,
  ]
}

const questionKeySelectionOptions = computed(() => (
  buildSelectionConfigurationKeyOptions(props.questionKeyOptions, props.form.question_key)
))

const answerKeySelectionOptions = computed(() => (
  buildSelectionConfigurationKeyOptions(props.answerKeyOptions, props.form.answer_key)
))

watch(
  () => [props.open, props.sourceLocale],
  () => {
    if (props.open) activeLocale.value = props.sourceLocale
  },
)

const optionTranslation = (option: WheelsetFitQuestionOptionForm) => (
  option.translations.find((translation) => translation.locale === activeLocale.value)
  || option.translations[0]
)

const addOption = () => {
  const optionNumber = props.form.options.length + 1
  props.form.options.push({
    client_id: Date.now() + nextClientID++,
    option_key: '',
    answer_value: '',
    sort_order: optionNumber * 10,
    is_unknown: false,
    is_enabled: true,
    product_filter_effects_json: '{}',
    translations: props.form.translations.map((translation) => ({
      locale: translation.locale,
      label: '',
      description: '',
    })),
  })
}

const removeOption = (index: number) => {
  props.form.options.splice(index, 1)
  props.form.options.forEach((option, optionIndex) => {
    option.sort_order = (optionIndex + 1) * 10
  })
}

const moveOption = (index: number, direction: number) => {
  const target = index + direction
  if (target < 0 || target >= props.form.options.length) return
  const [option] = props.form.options.splice(index, 1)
  props.form.options.splice(target, 0, option)
  props.form.options.forEach((entry, optionIndex) => {
    entry.sort_order = (optionIndex + 1) * 10
  })
}
</script>
