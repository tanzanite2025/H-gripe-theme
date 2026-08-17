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
              <div>
                <h3 class="text-sm font-black tracking-tighter uppercase">分类图片</h3>
                <p class="mt-1 text-xs leading-5 text-muted-foreground">
                  仅支持 WEBP，固定为 {{ productSpecTemplateImageSize }} × {{ productSpecTemplateImageSize }} px（1:1）。
                </p>
              </div>
              <span
                v-if="imagePreviewSource"
                class="rounded-full bg-emerald-500/10 px-2.5 py-1 text-xs text-emerald-700 dark:text-emerald-300"
              >
                已选择
              </span>
            </div>

            <div class="flex flex-col gap-4 sm:flex-row sm:items-start">
              <div class="relative aspect-square w-36 shrink-0 overflow-hidden rounded-xl border bg-muted/50">
                <img
                  v-if="imagePreviewSource && !imageLoadFailed"
                  :src="imagePreviewSource"
                  alt="分类图片预览"
                  class="h-full w-full object-cover"
                  @error="imageLoadFailed = true"
                />
                <div v-else class="flex h-full w-full items-center justify-center text-muted-foreground/50">
                  <ImageOff class="size-8" />
                </div>
              </div>

              <div class="min-w-0 space-y-3">
                <div class="flex flex-wrap items-center gap-2">
                  <Button type="button" variant="outline" size="sm" @click="imageInput?.click()">
                    <UploadCloud class="size-3.5" />
                    选择 WEBP
                  </Button>
                  <Button
                    v-if="imagePreviewSource"
                    type="button"
                    variant="ghost"
                    size="icon"
                    class="size-8"
                    title="移除分类图片"
                    aria-label="移除分类图片"
                    @click="clearImage"
                  >
                    <Trash2 class="size-4 text-destructive" />
                  </Button>
                </div>
                <input
                  ref="imageInput"
                  type="file"
                  class="sr-only"
                  accept=".webp,image/webp"
                  @change="handleImageInput"
                />
                <p class="text-xs leading-5 text-muted-foreground">
                  首页分类卡片会按正方形显示；{{ form.id === null ? '保存模板后上传图片。' : '保存模板时同步替换图片。' }}
                </p>
              </div>
            </div>
          </section>

          <section class="rounded-2xl border border-dashed border-border/80 bg-card/70 p-4">
            <div class="mb-3 flex flex-wrap items-start justify-between gap-3">
              <div>
                <h3 class="text-sm font-black tracking-tighter uppercase">多语言名称</h3>
                <p class="mt-1 text-xs leading-5 text-muted-foreground">
                  为已启用后台语言维护分类名称和描述。空名称不会提交，前台会回退到基础名称。
                </p>
              </div>
              <span class="rounded-full bg-muted px-2.5 py-1 text-xs text-muted-foreground">
                {{ filledTranslationCount(form.translations) }} 个已填写
              </span>
            </div>

            <div v-if="form.translations.length" class="grid min-w-0 gap-3 md:grid-cols-2 xl:grid-cols-4">
              <section
                v-for="translation in form.translations"
                :key="translation.locale"
                class="min-w-0 rounded-xl border bg-background/80 p-3"
              >
                <div class="mb-2 flex items-center justify-between gap-2">
                  <div class="min-w-0">
                    <p class="truncate text-xs font-black">{{ languageLabel(translation.locale) }}</p>
                    <p class="font-mono text-[10px] text-muted-foreground">{{ translation.locale }}</p>
                  </div>
                  <span class="rounded-full bg-muted px-2 py-0.5 text-[10px] text-muted-foreground">
                    可选
                  </span>
                </div>
                <div class="space-y-2">
                  <AdminFormField label="名称">
                    <Input v-model="translation.name" :placeholder="`请输入${languageLabel(translation.locale)}名称`" />
                  </AdminFormField>
                  <AdminFormField label="描述">
                    <Textarea
                      v-model="translation.description"
                      class="min-h-16 resize-y"
                      :placeholder="`请输入${languageLabel(translation.locale)}描述（可选）`"
                    />
                  </AdminFormField>
                </div>
              </section>
            </div>
            <div v-else class="rounded-xl border border-dashed py-6 text-center text-xs text-muted-foreground">
              暂无已启用语言。请先在语言设置中启用语言。
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
                  <AdminFormField v-if="spec.field_type === 'select' && spec.is_variant_option" label="前台展示">
                    <Select v-model="spec.presentation" :disabled="systemManaged">
                      <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="text">文字按钮</SelectItem>
                        <SelectItem value="color">颜色色板</SelectItem>
                        <SelectItem value="image">图片选项</SelectItem>
                      </SelectContent>
                    </Select>
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
                      @input="emit('clear-error', `spec:${index}:options`)"
                    />
                  </AdminFormField>
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
                    <span>SKU 选项</span>
                    <Switch v-model="spec.is_variant_option" :disabled="systemManaged" :aria-label="`${spec.name || '字段'}作为变体选项`" />
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
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { ImageOff, LoaderCircle, Plus, SlidersHorizontal, Trash2, UploadCloud } from '@lucide/vue'
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
import type { LanguageOption } from '@/lib/languages'
import type {
  ProductSpecificSpecPredicate,
  ProductSpecTemplateDialogMode,
  ProductSpecTemplateForm,
  ProductSpecTemplateFormErrors,
  ProductSpecTemplateSpecForm,
  ProductSpecTemplateTranslationForm
} from './productSpecificationTemplateTypes'
import { PRODUCT_SPEC_TEMPLATE_IMAGE_SIZE } from './productSpecificationTemplateTypes'

const props = withDefaults(defineProps<{
  open?: boolean
  mode?: ProductSpecTemplateDialogMode
  form: ProductSpecTemplateForm
  errors?: ProductSpecTemplateFormErrors
  submitting?: boolean
  showSpecAdvanced?: boolean
  systemManaged?: boolean
  isProductSpecificSelect: ProductSpecificSpecPredicate
  languageOptions?: LanguageOption[]
}>(), {
  open: false,
  mode: 'create',
  errors: () => ({}),
  submitting: false,
  showSpecAdvanced: false,
  systemManaged: false,
  languageOptions: () => []
})

const languageLabel = (locale: string): string => (
  props.languageOptions.find((option) => option.value === locale)?.label || locale
)
const filledTranslationCount = (translations: ProductSpecTemplateTranslationForm[]): number => (
  translations.filter((translation) => translation.name.trim()).length
)
const imageInput = ref<HTMLInputElement | null>(null)
const pendingImagePreviewURL = ref('')
const imageLoadFailed = ref(false)
const productSpecTemplateImageSize = PRODUCT_SPEC_TEMPLATE_IMAGE_SIZE
let previewObjectURL = ''

const imagePreviewSource = computed(() => (
  pendingImagePreviewURL.value || String(props.form.image_url || '').trim()
))

const revokePreviewObjectURL = (): void => {
  if (!previewObjectURL) return
  URL.revokeObjectURL(previewObjectURL)
  previewObjectURL = ''
}

watch(() => props.form.pending_image_file, (file) => {
  revokePreviewObjectURL()
  pendingImagePreviewURL.value = ''
  imageLoadFailed.value = false
  if (!file) return
  previewObjectURL = URL.createObjectURL(file)
  pendingImagePreviewURL.value = previewObjectURL
}, { immediate: true })

watch(imagePreviewSource, () => {
  imageLoadFailed.value = false
})

onBeforeUnmount(() => {
  revokePreviewObjectURL()
})

const handleImageInput = (event: Event): void => {
  const input = event.target instanceof HTMLInputElement ? event.target : null
  const file = input?.files?.[0] || null
  if (input) input.value = ''
  if (!file) return

  const isWebP = file.name.toLowerCase().endsWith('.webp')
    && (!file.type || file.type === 'image/webp')
  if (!isWebP) {
    emit('image-error', '分类图片只能上传 WEBP 文件')
    return
  }

  const objectURL = URL.createObjectURL(file)
  const image = new window.Image()
  image.onload = () => {
    URL.revokeObjectURL(objectURL)
    if (image.naturalWidth !== productSpecTemplateImageSize || image.naturalHeight !== productSpecTemplateImageSize) {
      emit('image-error', `分类图片必须是 ${productSpecTemplateImageSize} × ${productSpecTemplateImageSize} px`)
      return
    }
    emit('image-selected', file)
  }
  image.onerror = () => {
    URL.revokeObjectURL(objectURL)
    emit('image-error', '无法读取分类图片，请重新选择有效的 WEBP 文件')
  }
  image.src = objectURL
}

const clearImage = (): void => {
  emit('image-cleared')
}

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'update:showSpecAdvanced', value: boolean): void
  (event: 'submit'): void
  (event: 'clear-error', key: string): void
  (event: 'add-spec'): void
  (event: 'remove-spec', index: number): void
  (event: 'image-selected', value: File): void
  (event: 'image-cleared'): void
  (event: 'image-error', message: string): void
}>()
</script>
