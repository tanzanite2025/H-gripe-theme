<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent
      size="full"
      class="flex h-[92dvh] max-h-[calc(100dvh-1rem)] gap-0 overflow-hidden p-0"
      @open-auto-focus.prevent
    >
      <form class="flex min-h-0 min-w-0 flex-1 flex-col" @submit.prevent="emit('submit')">
        <DialogHeader class="shrink-0 border-b px-5 py-3 pr-12">
          <DialogTitle>{{ mode === 'create' ? '添加商品规格模板' : '编辑商品规格模板' }}</DialogTitle>
          <DialogDescription>
            商品规格模板只定义字段结构；具体重量、价格、库存和每个 SKU 的实际取值在商品编辑里维护。
          </DialogDescription>
        </DialogHeader>

        <div class="min-h-0 flex-1 space-y-4 overflow-y-auto px-5 py-4">
          <section class="rounded-2xl border border-dashed border-border/80 bg-card/70 p-4">
            <div class="mb-3 flex flex-wrap items-center justify-between gap-3">
              <div>
                <h3 class="text-sm font-black tracking-tighter uppercase">基础信息</h3>
                <p class="mt-1 text-xs text-muted-foreground">两行内完成模板名称、标识、排序和启用状态。</p>
              </div>
              <label class="inline-flex items-center gap-2 rounded-full border border-dashed px-3 py-1.5 text-xs font-medium">
                <Switch v-model="form.is_enabled" aria-label="启用商品规格模板" />
                启用模板
              </label>
            </div>
            <div
              v-if="systemManaged"
              class="mb-3 rounded-xl border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-xs leading-5 text-amber-800 dark:text-amber-200"
            >
              这是平台系统模板。模板标识和字段骨架由系统维护；名称、分组、单位、必填、可见和排序仍可调整，具体规格值在商品编辑中维护。
            </div>
            <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-[minmax(220px,1fr)_minmax(220px,1fr)_120px]">
              <AdminFormField label="模板名称" required :error="errors.name">
                <Input v-model="form.name" @input="emit('clear-error', 'name')" />
              </AdminFormField>
              <AdminFormField label="模板标识" required :error="errors.slug">
                <Input v-model="form.slug" class="font-mono" placeholder="例如：rim / carbon_frame" :disabled="systemManaged" @input="emit('clear-error', 'slug')" />
              </AdminFormField>
              <AdminFormField label="排序">
                <Input v-model.number="form.sort_order" type="number" min="0" step="1" />
              </AdminFormField>
              <AdminFormField label="说明" class="md:col-span-2 xl:col-span-3">
                <Textarea v-model="form.description" class="min-h-14 resize-y" placeholder="可选，给后台识别用，不会作为具体商品参数" />
              </AdminFormField>
            </div>
          </section>

          <section class="rounded-2xl border border-dashed border-border/80 bg-card/70 p-4">
            <div class="mb-3 flex flex-wrap items-start justify-between gap-3">
              <div class="space-y-1">
                <h3 class="text-sm font-black tracking-tighter uppercase">字段模板</h3>
                <p class="text-xs leading-5 text-muted-foreground">
                  这里只定义“该类商品编辑时出现哪些字段”。不要在这里填写某个具体产品的重量、尺寸、库存或价格。
                </p>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <span class="rounded-full bg-muted px-2.5 py-1 text-xs text-muted-foreground">
                  {{ form.spec_definitions.length }} 个字段
                </span>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  :aria-pressed="showSpecAdvanced"
                  @click="emit('update:showSpecAdvanced', !showSpecAdvanced)"
                >
                  <SlidersHorizontal class="size-3.5" />
                  {{ showSpecAdvanced ? '隐藏属性' : '字段属性' }}
                </Button>
                <Button v-if="!systemManaged" type="button" variant="outline" size="sm" @click="emit('add-spec')">
                  <Plus class="size-3.5" />
                  添加字段
                </Button>
              </div>
            </div>

            <div class="grid min-w-0 gap-3 lg:grid-cols-2 xl:grid-cols-3">
              <div v-if="form.spec_definitions.length === 0" class="rounded-xl border border-dashed py-8 text-center text-xs text-muted-foreground lg:col-span-2 xl:col-span-3">
                暂无字段模板。添加后，这些字段会出现在商品编辑页。
              </div>

              <section
                v-for="(spec, index) in form.spec_definitions"
                :key="spec.clientKey"
                class="rounded-xl border bg-background/80 p-3"
              >
                <div class="mb-2 flex items-center justify-between gap-2">
                  <strong class="text-xs font-black uppercase tracking-wider text-muted-foreground">字段 {{ index + 1 }}</strong>
                  <Button v-if="!systemManaged" type="button" variant="ghost" size="icon" class="size-8" :aria-label="`删除字段 ${index + 1}`" @click="emit('remove-spec', index)">
                    <Trash2 class="size-4 text-destructive" />
                  </Button>
                </div>

                <div
                  v-if="isProductSpecificSelect(spec)"
                  class="mb-2 rounded-lg border border-amber-500/25 bg-amber-500/10 px-3 py-2 text-xs leading-5 text-amber-800 dark:text-amber-200"
                >
                  这个字段看起来像每个商品/SKU 自己决定的值。若不同商品的可选值不同，请把字段类型改成“文本/数字”，不要在商品规格模板里固定列出选项。
                </div>

                <div class="grid gap-2 sm:grid-cols-2">
                  <AdminFormField label="字段名称" required :error="errors[`spec:${index}:name`]">
                    <Input v-model="spec.name" placeholder="字段显示名" @input="emit('clear-error', `spec:${index}:name`)" />
                  </AdminFormField>
                  <AdminFormField label="字段标识" required :error="errors[`spec:${index}:slug`]">
                    <Input v-model="spec.slug" class="font-mono" placeholder="field_slug" @input="emit('clear-error', `spec:${index}:slug`)" />
                  </AdminFormField>
                  <AdminFormField label="字段类型" required>
                    <Select v-model="spec.field_type" :disabled="systemManaged">
                      <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="text">文本</SelectItem>
                        <SelectItem value="number">数字</SelectItem>
                        <SelectItem value="select">选项</SelectItem>
                        <SelectItem value="boolean">开关</SelectItem>
                      </SelectContent>
                    </Select>
                  </AdminFormField>
                  <AdminFormField label="字段角色" :error="errors[`spec:${index}:role`]">
                    <Select
                      v-model="spec.role"
                      :disabled="systemManaged"
                    >
                      <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="attribute">静态规格</SelectItem>
                        <SelectItem value="variant">核心变体</SelectItem>
                        <SelectItem value="custom_option">买家选配</SelectItem>
                      </SelectContent>
                    </Select>
                  </AdminFormField>
                  <AdminFormField v-if="spec.field_type === 'select' && (spec.role === 'variant' || spec.role === 'custom_option')" label="前台展示">
                    <Select v-model="spec.presentation" :disabled="systemManaged">
                      <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="text">文字按钮</SelectItem>
                        <SelectItem value="color">颜色色板</SelectItem>
                        <SelectItem value="image">图片选项</SelectItem>
                      </SelectContent>
                    </Select>
                  </AdminFormField>
                  <AdminFormField v-if="spec.role === 'custom_option'" label="选择方式">
                    <Select v-model="spec.selection_mode" :disabled="systemManaged">
                      <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="single">单选</SelectItem>
                        <SelectItem value="multiple">多选</SelectItem>
                      </SelectContent>
                    </Select>
                  </AdminFormField>
                  <AdminFormField v-if="spec.role === 'custom_option'" label="最少选择数">
                    <Input v-model.number="spec.min_selections" type="number" min="0" step="1" />
                  </AdminFormField>
                  <AdminFormField v-if="spec.role === 'custom_option'" label="最多选择数">
                    <Input v-model.number="spec.max_selections" type="number" min="0" step="1" placeholder="不限制" />
                  </AdminFormField>
                  <AdminFormField label="单位">
                    <Input v-model="spec.unit" placeholder="可选" />
                  </AdminFormField>
                  <AdminFormField label="排序">
                    <Input v-model.number="spec.sort_order" type="number" min="0" step="1" />
                  </AdminFormField>
                  <AdminFormField
                    v-if="spec.field_type === 'select'"
                    label="模板共享选项（可选）"
                    :required="false"
                    class="sm:col-span-2"
                    :error="errors[`spec:${index}:options`]"
                    description="可选：填写该类型产品常用的共享值，每行一个。留空表示商品或 SKU 编辑时动态录入实际值；这里不会固定具体规格。"
                  >
                    <Textarea
                      v-model="spec.optionsText"
                      class="min-h-12 font-mono text-xs"
                      :disabled="systemManaged"
                      placeholder="可选，每行一个常用共享值，例如：Black\nWhite"
                      @input="syncOptionItems(spec); emit('clear-error', `spec:${index}:options`)"
                    />
                  </AdminFormField>
                  <div
                    v-if="spec.role === 'variant' || spec.role === 'custom_option'"
                    class="sm:col-span-2"
                  >
                    <div class="mb-2 flex items-center justify-between gap-2">
                      <span class="text-xs font-semibold">候选值属性</span>
                      <span class="text-[11px] text-muted-foreground">新建商品时会复制这些默认值</span>
                    </div>
                    <div v-if="spec.option_items.length" class="grid gap-2 md:grid-cols-2">
                      <div v-for="item in spec.option_items" :key="item.value_key" class="space-y-2 rounded-lg border p-2">
                        <div class="grid gap-2 sm:grid-cols-2">
                          <Input v-model="item.value_key" class="font-mono text-xs" placeholder="稳定值" />
                          <Input v-model="item.default_label" class="text-xs" placeholder="默认显示名" />
                          <Input v-if="spec.presentation === 'color'" v-model="item.color_hex" class="font-mono text-xs uppercase" placeholder="#000000" />
                          <Input v-if="spec.role === 'custom_option'" v-model.number="item.default_price_delta_minor" type="number" min="0" step="1" placeholder="默认加价（最小货币单位）" />
                        </div>
                        <label v-if="spec.role === 'custom_option'" class="flex items-center justify-between gap-2 text-xs">
                          <span>默认选中</span>
                          <Switch v-model="item.is_default" :aria-label="`${item.default_label || item.value_key}默认选中`" />
                        </label>
                      </div>
                    </div>
                  </div>
                </div>

                <div v-if="showSpecAdvanced" class="mt-2 grid gap-2 border-t border-dashed pt-2 sm:grid-cols-2">
                  <label class="flex items-center justify-between gap-3 rounded-xl border border-dashed px-3 py-2 text-xs font-bold uppercase tracking-wider">
                    <span>必填</span>
                    <Switch v-model="spec.is_required" :aria-label="`${spec.name || '字段'}必填`" />
                  </label>
                  <label class="flex items-center justify-between gap-3 rounded-xl border border-dashed px-3 py-2 text-xs font-bold uppercase tracking-wider">
                    <span>可筛选</span>
                    <Switch v-model="spec.is_filterable" :disabled="systemManaged" :aria-label="`${spec.name || '字段'}可筛选`" />
                  </label>
                  <label class="flex items-center justify-between gap-3 rounded-xl border border-dashed px-3 py-2 text-xs font-bold uppercase tracking-wider">
                    <span>前台可见</span>
                    <Switch v-model="spec.is_visible" :aria-label="`${spec.name || '字段'}前台可见`" />
                  </label>
                  <label class="flex items-center justify-between gap-3 rounded-xl border border-dashed px-3 py-2 text-xs font-bold uppercase tracking-wider">
                    <span>运行时角色</span>
                    <span class="font-mono text-[10px] text-muted-foreground">{{ spec.role }}</span>
                  </label>
                </div>
              </section>
            </div>
          </section>
        </div>

        <DialogFooter class="mx-0 mb-0 shrink-0 rounded-none border-t bg-background/95 px-5 py-3 backdrop-blur sm:flex-row sm:flex-nowrap sm:justify-end">
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存模板' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { LoaderCircle, Plus, SlidersHorizontal, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import type {
  ProductSpecificSpecPredicate,
  ProductSpecTemplateDialogMode,
  ProductSpecTemplateForm,
  ProductSpecTemplateFormErrors,
  ProductSpecTemplateSpecForm
} from '@/modules/product/productSpecificationTemplateTypes'

const props = withDefaults(defineProps<{
  open?: boolean
  mode?: ProductSpecTemplateDialogMode
  form: ProductSpecTemplateForm
  errors?: ProductSpecTemplateFormErrors
  submitting?: boolean
  showSpecAdvanced?: boolean
  systemManaged?: boolean
  isProductSpecificSelect: ProductSpecificSpecPredicate
}>(), {
  open: false,
  mode: 'create',
  errors: () => ({}),
  submitting: false,
  showSpecAdvanced: false,
  systemManaged: false
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'update:showSpecAdvanced', value: boolean): void
  (event: 'submit'): void
  (event: 'clear-error', key: string): void
  (event: 'add-spec'): void
  (event: 'remove-spec', index: number): void
}>()

const syncOptionItems = (spec: ProductSpecTemplateSpecForm): void => {
  const keys = String(spec.optionsText || '')
    .split(/\r?\n/)
    .map((value) => value.trim())
    .filter(Boolean)
    .filter((value, index, values) => values.indexOf(value) === index)
  const existing = new Map((spec.option_items || []).map((item) => [String(item.value_key || ''), item]))
  spec.option_items = keys.map((valueKey, index) => ({
    ...(existing.get(valueKey) || {}),
    value_key: valueKey,
    default_label: existing.get(valueKey)?.default_label || valueKey,
    is_enabled_by_default: existing.get(valueKey)?.is_enabled_by_default !== false,
    sort_order: existing.get(valueKey)?.sort_order ?? index * 10,
    revision: existing.get(valueKey)?.revision || 1
  }))
}
</script>

