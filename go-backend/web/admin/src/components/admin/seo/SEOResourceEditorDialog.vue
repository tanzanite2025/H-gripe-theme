<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[92dvh] overflow-y-auto p-0" @open-auto-focus.prevent>
      <form @submit.prevent="submit">
        <DialogHeader class="border-b px-5 py-4 pr-12">
          <DialogTitle>编辑{{ resourceKindLabel }} SEO</DialogTitle>
          <DialogDescription>这里只更新当前资源的搜索元数据。</DialogDescription>
        </DialogHeader>

        <div v-if="resource" class="space-y-5 px-5 py-5">
          <section class="grid gap-4 rounded-xl border bg-muted/20 p-4 sm:grid-cols-[150px_minmax(0,1fr)] sm:items-center">
            <div>
              <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">页面资源</p>
              <p class="mt-1 truncate text-sm font-black">{{ resource.title }}</p>
            </div>
            <div class="flex min-w-0 items-center gap-2">
              <LockKeyhole class="size-4 shrink-0 text-muted-foreground" />
              <code class="truncate text-xs text-muted-foreground">{{ resource.routePath }}</code>
            </div>
          </section>

          <div class="grid gap-4">
            <AdminFormField label="Meta 标题">
              <Input v-model="form.meta_title" maxlength="160" placeholder="搜索结果标题" :disabled="!canEdit || saving" />
            </AdminFormField>
            <AdminFormField label="Meta 描述">
              <Textarea
                v-model="form.meta_description"
                maxlength="320"
                class="min-h-28"
                placeholder="搜索结果摘要"
                :disabled="!canEdit || saving"
              />
            </AdminFormField>
            <AdminFormField v-if="kind === 'article'" label="规范 URL">
              <Input
                v-model="form.canonical_url"
                maxlength="2048"
                type="url"
                placeholder="留空则使用当前页面路由"
                :disabled="!canEdit || saving"
              />
            </AdminFormField>
            <AdminFormField v-if="kind === 'category'" label="分类导语">
              <ProductDescriptionEditor
                v-model="form.intro"
                :disabled="!canEdit || saving"
              />
            </AdminFormField>
          </div>

          <section v-if="kind === 'product'" class="rounded-xl border bg-muted/20 p-4">
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Product JSON-LD</p>
                <p class="mt-1 text-sm font-bold">
                  {{ productDiagnostics ? productDiagnostics.structured_data_type : '等待诊断' }}
                </p>
              </div>
              <AdminStatusBadge :tone="productDiagnostics?.ready ? 'green' : 'amber'">
                {{ productDiagnostics?.ready ? '可生成' : '需检查' }}
              </AdminStatusBadge>
            </div>
            <dl v-if="productDiagnostics" class="mt-4 grid gap-3 text-xs sm:grid-cols-2">
              <div>
                <dt class="text-muted-foreground">Product.name</dt>
                <dd class="mt-1 font-medium">{{ productDiagnostics.product_name || resource.title }}</dd>
              </div>
              <div>
                <dt class="text-muted-foreground">brand</dt>
 <dd class="mt-1 font-medium">{{ productDiagnostics.brand || '未配置'}}</dd>
              </div>
              <div>
                <dt class="text-muted-foreground">SKU</dt>
 <dd class="mt-1 font-medium">{{ productDiagnostics.sku || '未配置'}}</dd>
              </div>
              <div>
                <dt class="text-muted-foreground">价格 / 币种</dt>
                <dd class="mt-1 font-medium">
                  {{ formatPrice(productDiagnostics.price, productDiagnostics.currency) }}
                </dd>
              </div>
              <div>
                <dt class="text-muted-foreground">库存状态</dt>
                <dd class="mt-1 font-medium">{{ availabilityLabel(productDiagnostics.availability) }}</dd>
              </div>
              <div>
                <dt class="text-muted-foreground">图片</dt>
                <dd class="mt-1 font-medium">{{ productDiagnostics.image_count }} 张可见图片</dd>
              </div>
              <div>
                <dt class="text-muted-foreground">Meta 标题来源</dt>
                <dd class="mt-1 font-medium">{{ metaStateLabel(productDiagnostics.meta_title) }}</dd>
              </div>
              <div>
                <dt class="text-muted-foreground">Meta 描述来源</dt>
                <dd class="mt-1 font-medium">{{ metaStateLabel(productDiagnostics.meta_description) }}</dd>
              </div>
              <div v-if="productDiagnostics.missing.length" class="sm:col-span-2">
                <dt class="text-muted-foreground">缺少或待确认</dt>
                <dd class="mt-1 font-medium text-amber-700 dark:text-amber-300">
                  {{ productDiagnostics.missing.map(issueLabel).join('、') }}
                </dd>
              </div>
              <div v-if="productDiagnostics.blocking_issues.length" class="sm:col-span-2">
                <dt class="text-muted-foreground">阻塞项</dt>
                <dd class="mt-1 font-medium text-destructive">
                  {{ productDiagnostics.blocking_issues.map(issueLabel).join('、') }}
                </dd>
              </div>
              <div v-if="productDiagnostics.warnings.length" class="sm:col-span-2">
                <dt class="text-muted-foreground">提醒</dt>
                <dd class="mt-1 font-medium text-amber-700 dark:text-amber-300">
                  {{ productDiagnostics.warnings.map(issueLabel).join('、') }}
                </dd>
              </div>
              <div class="sm:col-span-2">
                <dt class="text-muted-foreground">Google Merchant 字段</dt>
                <dd class="mt-1 font-medium">GTIN、MPN、identifier status 仅在 Merchant 渠道维护</dd>
              </div>
            </dl>
            <p v-else class="mt-4 text-xs text-muted-foreground">暂时无法读取产品源数据诊断。</p>

            <details v-if="productDiagnostics" class="mt-4 rounded-lg border bg-background/70 p-3">
              <summary class="cursor-pointer text-xs font-bold">查看结构化数据投影</summary>
              <pre class="mt-3 max-h-72 overflow-auto whitespace-pre-wrap break-all text-[11px] leading-5 text-muted-foreground">{{ structuredDataPreview }}</pre>
            </details>
          </section>

          <section v-if="kind === 'product'" class="rounded-xl border bg-muted/20 p-4">
            <div class="flex items-start justify-between gap-3">
              <div>
                <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">SEO Breadcrumb</p>
                <p class="mt-1 text-sm font-bold">分类仍使用 /shop 层级，商品末项使用 /products</p>
              </div>
              <AdminStatusBadge :tone="breadcrumbPathComplete ? 'green' : 'amber'">
                {{ breadcrumbPathComplete ? '路径完整' : '需检查' }}
              </AdminStatusBadge>
            </div>
            <dl class="mt-4 grid gap-3 text-xs sm:grid-cols-2">
              <div>
                <dt class="text-muted-foreground">主分类</dt>
                <dd class="mt-1 font-medium">
                  <template v-if="productDiagnostics?.primary_category">
                    {{ productDiagnostics.primary_category.name }}
                    <code class="ml-1 text-[11px] text-muted-foreground">
                      {{ productDiagnostics.primary_category.path }}
                    </code>
                  </template>
                  <span v-else>未解析</span>
                </dd>
              </div>
              <div>
                <dt class="text-muted-foreground">SSR BreadcrumbList</dt>
                <dd class="mt-1 font-medium">{{ breadcrumbSSRStatusLabel }}</dd>
              </div>
              <div class="sm:col-span-2">
                <dt class="text-muted-foreground">面包屑预览</dt>
                <dd class="mt-1 break-words font-medium">{{ breadcrumbPreview }}</dd>
              </div>
              <div v-if="breadcrumbReason" class="sm:col-span-2">
                <dt class="text-muted-foreground">异常原因</dt>
                <dd class="mt-1 font-medium text-amber-700 dark:text-amber-300">{{ breadcrumbReasonLabel }}</dd>
              </div>
            </dl>
          </section>
        </div>

        <DialogFooter class="sticky bottom-0 mx-0 mb-0 rounded-b-lg border-t bg-background px-5 py-4">
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="!canEdit || saving || !resource">
            <LoaderCircle v-if="saving" class="size-4 animate-spin" />
            <Save v-else class="size-4" />
            {{ saving ? '保存中' : '保存 SEO' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { LockKeyhole, LoaderCircle, Save } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import ProductDescriptionEditor from '@/components/admin/product/ProductDescriptionEditor.vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import type {
  SEOProductMetaFieldState,
  SEOResourceEditorValues,
  SEOResourceItem,
} from '@/modules/seo/types'

const props = withDefaults(defineProps<{
  open?: boolean
  kind: 'article' | 'product' | 'category'
  resource?: SEOResourceItem | null
  saving?: boolean
  canEdit?: boolean
}>(), {
  open: false,
  resource: null,
  saving: false,
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'save', values: SEOResourceEditorValues): void
}>()

const form = reactive<SEOResourceEditorValues>({
  meta_title: '',
  meta_description: '',
  canonical_url: '',
  intro: '',
})

const resetForm = (resource: SEOResourceItem | null): void => {
  Object.assign(form, {
    meta_title: resource?.metaTitle || '',
    meta_description: resource?.metaDescription || '',
    canonical_url: resource?.canonicalUrl || '',
    intro: resource?.intro || '',
  })
}

watch(() => props.resource, resetForm, { immediate: true })

const resourceKindLabel = computed(() => {
  if (props.kind === 'article') return '文章'
  if (props.kind === 'category') return '分类'
  return '产品'
})

const productDiagnostics = computed(() => (
  props.kind === 'product' ? props.resource?.productDiagnostics || null : null
))

const structuredDataPreview = computed(() => (
  productDiagnostics.value
    ? JSON.stringify(productDiagnostics.value.structured_data, null, 2)
    : ''
))

const breadcrumbPathComplete = computed(() => (
  productDiagnostics.value?.breadcrumb_path_complete === true
))

const breadcrumbPreview = computed(() => {
  const items = productDiagnostics.value?.breadcrumb?.items || []
  if (!items.length) return '未生成'
  return items.map((item) => item.name).filter(Boolean).join(' > ') || '未生成'
})

const breadcrumbSSRStatusLabel = computed(() => {
  switch (productDiagnostics.value?.breadcrumb_ssr_status) {
    case 'ready':
      return '已就绪，初始 HTML 可输出'
    case 'blocked':
      return '未就绪，初始 HTML 不输出'
    default:
      return '未检查'
  }
})

const breadcrumbReason = computed(() => (
  productDiagnostics.value?.breadcrumb_reason
    || productDiagnostics.value?.breadcrumb?.reason
    || ''
))

const breadcrumbReasonLabel = computed(() => {
  const labels: Record<string, string> = {
    category_service_unavailable: '分类服务不可用',
    missing_category: '商品没有主分类',
    category_cycle: '分类父级存在循环',
    category_path_incomplete: '分类层级不完整或包含未启用分类',
    breadcrumb_unavailable: '面包屑数据不可用',
  }
  return labels[breadcrumbReason.value] || breadcrumbReason.value
})

const metaStateLabel = (state: SEOProductMetaFieldState): string => {
  if (state.is_custom) {
    return state.soft_length_warning ? `自定义，${state.length} 字符（偏长提醒）` : `自定义，${state.length} 字符`
  }
  if (state.source) {
    return state.soft_length_warning
      ? `Fallback：${state.source}，${state.length} 字符（偏长提醒）`
      : `Fallback：${state.source}，${state.length} 字符`
  }
  return '无可用值'
}

const formatPrice = (price: number | null | undefined, currency: string): string => {
  if (typeof price !== 'number' || !Number.isFinite(price) || !currency) return '未配置'
  return `${price.toFixed(2)} ${currency}`
}

const availabilityLabel = (availability: string): string => {
  switch (availability) {
    case 'in_stock':
      return 'In Stock'
    case 'out_of_stock':
      return 'Out of Stock'
    default:
      return '无法确认'
  }
}

const issueLabel = (issue: string): string => {
  const labels: Record<string, string> = {
    product_name: '产品名称',
    image: '可见产品图片',
    price_or_currency: '有效价格和币种',
    active_variant: '可用变体',
    product_not_public: '产品未发布',
    sku: 'SKU',
    brand: '品牌',
    meta_title_fallback: 'Meta 标题使用 fallback',
    meta_description_fallback: 'Meta 描述使用 fallback',
    meta_title_length: 'Meta 标题偏长',
    meta_description_length: 'Meta 描述偏长',
  }
  return labels[issue] || issue
}

const submit = (): void => {
  if (!props.resource || !props.canEdit || props.saving) return
  emit('save', { ...form })
}
</script>
