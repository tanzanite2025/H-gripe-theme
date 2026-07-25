<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent
      size="full"
      class="h-[94dvh] max-h-[calc(100dvh-1rem)] overflow-hidden p-0"
      @open-auto-focus.prevent
    >
      <form class="flex h-full min-h-0 min-w-0 flex-col" @submit.prevent="emit('submit')">
        <DialogHeader class="shrink-0 border-b px-5 py-4 pr-12">
          <DialogTitle>{{ mode === 'create' ? '添加商品' : '编辑商品' }}</DialogTitle>
          <DialogDescription>先录入商品基础识别信息，再绑定产品模板；模板决定下方商品字段和 SKU 选项列。</DialogDescription>
        </DialogHeader>

        <div class="min-h-0 min-w-0 flex-1 space-y-5 overflow-x-hidden overflow-y-auto overscroll-contain px-5 py-5 [scrollbar-gutter:stable]">
          <div class="grid gap-2 rounded-2xl border border-dashed bg-muted/20 p-3 text-xs leading-5 text-muted-foreground sm:grid-cols-2 xl:grid-cols-4">
            <div class="rounded-xl bg-background/70 px-3 py-2">
              <span class="font-mono text-[10px] font-black text-primary">01</span>
              <strong class="mt-0.5 block text-foreground">基础识别</strong>
              <span>名称、Slug、语言和描述，只负责识别这个商品。</span>
            </div>
            <div class="rounded-xl bg-background/70 px-3 py-2">
              <span class="font-mono text-[10px] font-black text-primary">02</span>
              <strong class="mt-0.5 block text-foreground">绑定模板</strong>
              <span>车圈、车架等模板决定后续字段结构。</span>
            </div>
            <div class="rounded-xl bg-background/70 px-3 py-2">
              <span class="font-mono text-[10px] font-black text-primary">03</span>
              <strong class="mt-0.5 block text-foreground">填写参数</strong>
              <span>商品参数来自模板，但具体值只属于当前商品。</span>
            </div>
            <div class="rounded-xl bg-background/70 px-3 py-2">
              <span class="font-mono text-[10px] font-black text-primary">04</span>
              <strong class="mt-0.5 block text-foreground">维护 SKU</strong>
              <span>价格、重量、库存和 SKU 选项按每行变体维护。</span>
            </div>
          </div>

          <AdminFormSection title="基础信息" description="这里不放规格字段；规格字段必须先通过产品模板统一定义，再在下方录入具体值。">
            <div class="grid gap-4 md:grid-cols-3">
              <AdminFormField label="商品名称" required :error="errors.name">
                <Input v-model="form.name" placeholder="请输入商品名称" @input="emit('clear-error', 'name')" />
              </AdminFormField>
              <AdminFormField label="Slug" required :error="errors.slug">
                <Input v-model="form.slug" placeholder="例如 crystal-bracelet" @input="emit('clear-error', 'slug')" />
              </AdminFormField>
              <AdminFormField label="语言" required>
                <Select v-model="form.locale">
                  <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="zh">中文</SelectItem>
                    <SelectItem value="en">English</SelectItem>
                  </SelectContent>
                </Select>
              </AdminFormField>
              <AdminFormField label="简短描述" class="md:col-span-3">
                <Textarea v-model="form.short_description" class="min-h-20" placeholder="用于列表和摘要展示" />
              </AdminFormField>
              <AdminFormField label="详细描述" class="md:col-span-3">
                <Textarea v-model="form.description" class="min-h-28" placeholder="请输入商品详细描述" />
              </AdminFormField>
            </div>
          </AdminFormSection>

          <AdminFormSection title="绑定产品模板" description="这是商品资料和模板字段之间的总开关。选择模板后，下方才会出现对应的商品参数字段和 SKU 选项列。">
            <div class="grid gap-4 xl:grid-cols-[minmax(260px,0.9fr)_minmax(0,1.1fr)]">
              <div class="space-y-2">
                <AdminFormField label="产品模板">
                  <Select :model-value="productTypeSelectValue" @update:model-value="emit('product-type-select', $event)">
                    <SelectTrigger class="w-full"><SelectValue placeholder="请选择产品模板" /></SelectTrigger>
                    <SelectContent>
                      <SelectItem value="__none__">未选择模板</SelectItem>
                      <SelectItem v-for="type in productTypes" :key="type.id" :value="String(type.id)">
                        {{ type.name }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </AdminFormField>
                <p class="text-xs leading-5 text-muted-foreground">
                  模板只定义“要填哪些字段”，不保存某个商品的具体重量、价格、库存或尺寸值。
                </p>
                <Button type="button" variant="outline" size="sm" as-child>
                  <RouterLink to="/product-types">
                    <Tags class="size-3.5" />
                    维护产品模板
                  </RouterLink>
                </Button>
              </div>

              <div class="rounded-2xl border border-dashed bg-muted/20 p-4">
                <div v-if="selectedProductType" class="space-y-3">
                  <div class="flex flex-wrap items-center gap-2">
                    <span class="text-sm font-bold">{{ selectedProductType.name }}</span>
                    <span class="rounded-full bg-background px-2 py-0.5 font-mono text-[10px] text-muted-foreground">
                      {{ selectedProductType.slug }}
                    </span>
                  </div>
                  <p v-if="selectedProductType.description" class="text-xs leading-5 text-muted-foreground">
                    {{ selectedProductType.description }}
                  </p>
                  <div class="grid gap-2 sm:grid-cols-2">
                    <div class="rounded-xl bg-background/70 px-3 py-2">
                      <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">商品字段</span>
                      <strong class="mt-1 block font-mono text-lg">{{ selectedSpecDefinitions.length }}</strong>
                    </div>
                    <div class="rounded-xl bg-background/70 px-3 py-2">
                      <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SKU 选项字段</span>
                      <strong class="mt-1 block font-mono text-lg">{{ variantSpecDefinitions.length }}</strong>
                    </div>
                  </div>
                  <div v-if="templateScopedValuesTouched" class="rounded-xl border border-amber-500/25 bg-amber-500/10 px-3 py-2 text-xs leading-5 text-amber-800 dark:text-amber-200">
                    如果切换模板，旧模板下的商品字段值和 SKU 选项值会清空；SKU 价格、重量、库存和商品媒体会保留。
                  </div>
                  <div class="grid gap-3 lg:grid-cols-2">
                    <div class="min-w-0 rounded-xl bg-background/70 p-3">
                      <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">商品参数字段</span>
                      <div v-if="selectedSpecDefinitions.length" class="mt-2 flex flex-wrap gap-1.5">
                        <span
                          v-for="spec in selectedSpecDefinitions"
                          :key="`product-${spec.id || spec.slug}`"
                          class="rounded-full bg-muted px-2 py-0.5 text-[11px] font-medium text-muted-foreground"
                        >
                          {{ getSpecLabel(spec) }}
                        </span>
                      </div>
                      <p v-else class="mt-2 text-xs text-muted-foreground">该模板没有商品级参数字段。</p>
                    </div>
                    <div class="min-w-0 rounded-xl bg-background/70 p-3">
                      <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SKU 选项字段</span>
                      <div v-if="variantSpecDefinitions.length" class="mt-2 flex flex-wrap gap-1.5">
                        <span
                          v-for="spec in variantSpecDefinitions"
                          :key="`variant-${spec.id || spec.slug}`"
                          class="rounded-full bg-primary/10 px-2 py-0.5 text-[11px] font-medium text-primary"
                        >
                          {{ getSpecLabel(spec) }}
                        </span>
                      </div>
                      <p v-else class="mt-2 text-xs text-muted-foreground">该模板没有 SKU 选项字段，仅维护默认 SKU 的价格、重量和库存。</p>
                    </div>
                  </div>
                </div>
                <p v-else class="text-xs leading-5 text-muted-foreground">
                  未选择模板时，下方不会出现商品参数字段，也不会给 SKU 生成选项列。建议先在“产品模板”页面创建车圈、车架等模板，再回到这里绑定。
                </p>
              </div>
            </div>
          </AdminFormSection>

          <AdminFormSection
            title="商品参数（来自模板）"
            :description="selectedSpecDefinitions.length ? '这里填写当前商品自己的参数值；字段来源于已绑定产品模板，但具体值不写回模板。' : '当前模板没有商品级参数字段；可以直接继续维护 SKU。'"
          >
            <div v-if="selectedSpecDefinitions.length" class="grid gap-4 md:grid-cols-2">
              <AdminFormField
                v-for="spec in selectedSpecDefinitions"
                :key="spec.id || spec.slug"
                :label="getSpecLabel(spec)"
                :required="spec.is_required"
                :error="errors[`spec:${spec.slug}`]"
              >
                <Input
                  v-if="spec.field_type === 'number'"
                  v-model.number="form.specs[spec.slug]"
                  type="number"
                  min="0"
                  @input="emit('clear-error', `spec:${spec.slug}`)"
                />
                <Select
                  v-else-if="spec.field_type === 'select'"
                  :model-value="specSelectValue(form.specs[spec.slug])"
                  @update:model-value="emit('set-spec-select-value', spec.slug, $event)"
                >
                  <SelectTrigger class="w-full"><SelectValue placeholder="请选择" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="__empty__">未设置</SelectItem>
                    <SelectItem v-for="option in parseSpecOptions(spec)" :key="String(option)" :value="String(option)">
                      {{ formatSpecOption(option) }}
                    </SelectItem>
                  </SelectContent>
                </Select>
                <div v-else-if="spec.field_type === 'boolean'" class="flex h-9 items-center gap-2">
                  <Switch v-model="form.specs[spec.slug]" :aria-label="spec.name" />
                  <span class="text-xs text-muted-foreground">{{ form.specs[spec.slug] ? '是' : '否' }}</span>
                </div>
                <Input
                  v-else
                  v-model="form.specs[spec.slug]"
                  :placeholder="`请输入${spec.name}`"
                  @input="emit('clear-error', `spec:${spec.slug}`)"
                />
              </AdminFormField>
            </div>
            <div v-else class="rounded-xl border border-dashed bg-muted/20 px-4 py-5 text-center text-xs text-muted-foreground">
              {{ selectedProductType ? '这个模板没有商品级字段。' : '选择产品模板后，商品级字段会显示在这里。' }}
            </div>
          </AdminFormSection>

          <AdminFormSection title="SKU 变体矩阵" description="SKU 选项列来自产品模板；价格、重量和库存永远按每一行 SKU 独立维护，前台按用户选中的 SKU 显示对应重量。">
            <div class="min-w-0 rounded-lg border">
              <ProductVariantEditor
                :variants="form.variants"
                :spec-definitions="variantSpecDefinitions"
                :default-index="defaultVariantIndex"
                class="min-w-0 p-3"
                @add="emit('add-variant')"
                @remove="emit('remove-variant', $event)"
                @set-default="emit('set-default-variant', $event)"
              />
            </div>
            <p v-if="errors.variants" class="mt-2 text-xs font-medium text-destructive">{{ errors.variants }}</p>
          </AdminFormSection>

          <ProductMediaSection
            :media-items="form.media"
            :uploading="uploadingMedia"
            :error="errors.media"
            @upload="(...args) => emit('upload-media', ...args)"
            @add-url="emit('add-media-url', $event)"
            @clear-error="emit('clear-error', 'media')"
            @set-primary="emit('set-primary-media', $event)"
            @move="(...args) => emit('move-media', ...args)"
            @remove="emit('remove-media', $event)"
          />

          <AdminFormSection title="发布设置" description="控制商品的公开状态和前台可见性。">
            <div class="grid gap-4 md:grid-cols-2">
              <div class="md:col-span-2 rounded-lg border bg-muted/30 px-3 py-2.5 text-xs text-muted-foreground">
                重量现在只在 SKU 变体里维护，前台会按当前选中的 SKU 显示对应重量。
              </div>
              <AdminFormField label="状态" required>
                <Select v-model="form.status">
                  <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                  <SelectContent>
                    <SelectItem value="active">在售</SelectItem>
                    <SelectItem value="inactive">下架</SelectItem>
                    <SelectItem value="out_of_stock">缺货</SelectItem>
                  </SelectContent>
                </Select>
              </AdminFormField>
              <div class="flex items-center justify-between gap-4 rounded-lg border px-3 py-2.5 md:col-span-2">
                <div>
                  <Label for="product-featured">精选商品</Label>
                  <p class="mt-0.5 text-xs text-muted-foreground">在前台精选区域优先展示该商品。</p>
                </div>
                <Switch id="product-featured" v-model="form.featured" />
              </div>
            </div>
          </AdminFormSection>

          <AdminFormSection title="SEO" description="可选的搜索结果标题和描述。">
            <div class="grid gap-4">
              <AdminFormField label="SEO 标题">
                <Input v-model="form.meta_title" placeholder="请输入 SEO 标题" />
              </AdminFormField>
              <AdminFormField label="SEO 描述">
                <Textarea v-model="form.meta_description" class="min-h-20" placeholder="请输入 SEO 描述" />
              </AdminFormField>
            </div>
          </AdminFormSection>
        </div>

        <DialogFooter class="mx-0 mb-0 shrink-0 rounded-b-lg border-t bg-background px-5 py-4">
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存商品' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { RouterLink } from 'vue-router'
import { LoaderCircle, Tags } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import ProductMediaSection from '@/components/admin/product/ProductMediaSection.vue'
import ProductVariantEditor from '@/components/admin/product/ProductVariantEditor.vue'
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
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  submitting: { type: Boolean, default: false },
  form: { type: Object, required: true },
  errors: { type: Object, default: () => ({}) },
  productTypes: { type: Array, default: () => [] },
  selectedProductType: { type: Object, default: null },
  selectedSpecDefinitions: { type: Array, default: () => [] },
  variantSpecDefinitions: { type: Array, default: () => [] },
  defaultVariantIndex: { type: Number, default: 0 },
  productTypeSelectValue: { type: String, default: '__none__' },
  templateScopedValuesTouched: { type: Boolean, default: false },
  uploadingMedia: { type: Boolean, default: false },
  parseSpecOptions: { type: Function, required: true },
  formatSpecOption: { type: Function, required: true },
  getSpecLabel: { type: Function, required: true },
  specSelectValue: { type: Function, required: true },
})

const emit = defineEmits([
  'update:open',
  'submit',
  'clear-error',
  'product-type-select',
  'set-spec-select-value',
  'add-variant',
  'remove-variant',
  'set-default-variant',
  'upload-media',
  'add-media-url',
  'set-primary-media',
  'move-media',
  'remove-media',
])
</script>
