<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="xl" class="max-h-[calc(100dvh-1.5rem)]" @open-auto-focus.prevent>
      <form class="space-y-3" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '创建图库' : '编辑图库' }}</DialogTitle>
          <DialogDescription>图库信息、关联产品和图片在这里一次完成，首张排序图片自动作为封面。</DialogDescription>
        </DialogHeader>

        <div class="grid gap-3 lg:grid-cols-2">
          <AdminFormField label="标题" required :error="errors.title">
            <Input v-model="form.title" placeholder="请输入图库标题" @input="emit('clear-error', 'title')" />
          </AdminFormField>
          <AdminFormField label="Slug" required :error="errors.slug">
            <Input v-model="form.slug" placeholder="例如 customer-stories" @input="emit('clear-error', 'slug')" />
          </AdminFormField>
          <AdminFormField label="描述" class="lg:col-span-2">
            <Textarea v-model="form.description" class="min-h-16" placeholder="请输入图库描述" />
          </AdminFormField>
        </div>
        <div class="space-y-2 rounded-xl bg-muted/30 p-3">
          <div class="flex items-center justify-between gap-2">
            <div class="min-w-0">
              <p class="text-xs font-black text-foreground">关联产品</p>
              <p class="mt-0.5 text-[11px] text-muted-foreground">前台读取这些产品生成链接。</p>
            </div>
            <Button type="button" variant="outline" size="sm" @click="emit('open-product-picker')">
              <Plus class="size-3.5" />
              选择产品
            </Button>
          </div>
          <div v-if="form.product_links.length" class="flex flex-wrap gap-2">
            <span
              v-for="product in form.product_links"
              :key="String(product.product_id)"
              class="inline-flex max-w-full items-center gap-1.5 rounded-full bg-background px-2.5 py-1 text-[11px] font-bold text-foreground"
            >
              <span class="max-w-48 truncate">{{ product.name || product.slug || product.product_id }}</span>
              <button
                type="button"
                class="rounded-full text-muted-foreground hover:text-destructive focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                :aria-label="`移除产品 ${product.name || product.product_id}`"
                @click="emit('remove-product', product.product_id)"
              >
                <X class="size-3" />
              </button>
            </span>
          </div>
          <p v-else class="text-xs text-muted-foreground">暂无关联产品。</p>
        </div>

        <section class="space-y-3 rounded-xl bg-muted/30 p-3">
          <div class="flex items-center justify-between gap-2">
            <div class="min-w-0">
              <p class="text-xs font-black text-foreground">图库图片</p>
              <p class="mt-0.5 text-[11px] text-muted-foreground">从媒体仓库选择图片，可一次添加多张。</p>
            </div>
            <Button type="button" variant="outline" size="sm" @click="emit('add-image')">
              <Plus class="size-3.5" />
              添加图片
            </Button>
          </div>

          <div v-if="form.images.length" class="grid gap-3 xl:grid-cols-2">
            <div
              v-for="(image, index) in form.images"
              :key="image.id === null ? `new-${index}` : String(image.id)"
              class="min-w-0 space-y-2"
            >
              <p class="text-[11px] font-black uppercase tracking-wide text-muted-foreground">
                图片 {{ index + 1 }}
              </p>
              <GalleryImageFields
                :form="image"
                :errors="imageErrors[index]"
                :show-remove="form.images.length > 1 || mode === 'edit'"
                @pick-media="emit('pick-image', index)"
                @clear-error="emit('clear-image-error', index, $event)"
                @remove="emit('remove-image', index)"
              />
            </div>
          </div>
          <p v-else class="text-xs text-muted-foreground">暂无图片，点击“添加图片”开始配置。</p>
        </section>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存图库' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { LoaderCircle, Plus, X } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import GalleryImageFields from './GalleryImageFields.vue'
import type { GalleryDialogMode, GalleryForm, GalleryFormErrors } from './galleryTypes'

withDefaults(defineProps<{
  open?: boolean
  mode?: GalleryDialogMode
  form: GalleryForm
  errors?: GalleryFormErrors
  imageErrors?: GalleryFormErrors[]
  submitting?: boolean
}>(), {
  open: false,
  mode: 'create',
  errors: () => ({}),
  imageErrors: () => [],
  submitting: false
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit'): void
  (event: 'clear-error', key: string): void
  (event: 'open-product-picker'): void
  (event: 'remove-product', productId: string | number): void
  (event: 'add-image'): void
  (event: 'pick-image', index: number): void
  (event: 'clear-image-error', index: number, key: string): void
  (event: 'remove-image', index: number): void
}>()
</script>
