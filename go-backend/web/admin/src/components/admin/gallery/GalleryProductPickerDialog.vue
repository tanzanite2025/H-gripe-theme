<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="gap-0 p-0" @open-auto-focus.prevent>
      <DialogHeader class="border-b px-5 py-4 pr-12">
        <DialogTitle>选择关联产品</DialogTitle>
        <DialogDescription>从商品库选择，这里只保存产品 ID，前台链接由产品 Slug 生成。</DialogDescription>
      </DialogHeader>

      <div class="flex min-h-[28rem] max-h-[74dvh] min-w-0 flex-col overflow-hidden">
        <div class="flex shrink-0 flex-wrap items-center gap-3 border-b px-4 py-3">
          <div class="relative min-w-[16rem] flex-1">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <Input v-model="search" class="pl-8" placeholder="搜索商品名称、SKU 或 Slug" @keyup.enter="reload" />
          </div>
          <Button variant="outline" size="sm" :disabled="loading" @click="reload">
            <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
            搜索
          </Button>
        </div>

        <div v-if="loading" class="flex min-h-0 flex-1 items-center justify-center text-xs font-bold text-muted-foreground">
          正在加载商品
        </div>
        <div v-else-if="products.length === 0" class="flex min-h-0 flex-1 items-center justify-center text-xs font-bold text-muted-foreground">
          暂无可选商品
        </div>
        <div v-else class="min-h-0 flex-1 overflow-y-auto p-4">
          <div class="grid gap-2">
            <button
              v-for="product in products"
              :key="String(product.id)"
              type="button"
              class="flex min-w-0 items-center justify-between gap-3 rounded-xl bg-muted/35 px-3 py-2.5 text-left transition hover:bg-muted/60 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              @click="toggle(product)"
            >
              <span class="min-w-0">
 <span class="block truncate text-xs font-black text-foreground">{{ product.name || '-'}}</span>
                <span class="mt-0.5 block truncate font-mono text-[10px] font-bold text-muted-foreground">
                  #{{ product.id }} · {{ product.sku || '-' }} · {{ product.slug || '-' }}
                </span>
              </span>
              <span
                class="shrink-0 rounded-full px-2.5 py-1 text-[10px] font-black"
 :class="isSelected(product.id) ? 'bg-primary text-primary-foreground': 'bg-background text-muted-foreground'"
              >
                {{ isSelected(product.id) ? '已选择' : '选择' }}
              </span>
            </button>
          </div>
        </div>

        <div class="shrink-0 border-t px-4 py-3">
          <AdminPagination
            :page="pagination.page"
            :page-size="pagination.pageSize"
            :total="pagination.total"
            :page-sizes="[10, 20, 40]"
            @update:page="updatePage"
            @update:page-size="updatePageSize"
          />
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { RefreshCw, Search } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { productApi } from '@/api/products'
import type { ProductRecord } from '@/modules/product/productEditorTypes'
import type { GalleryId } from '@/modules/gallery/galleryTypes'

const props = withDefaults(defineProps<{
  open?: boolean
  selectedProductIds?: GalleryId[]
}>(), {
  open: false,
  selectedProductIds: () => [],
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'select', product: ProductRecord): void
  (event: 'remove', productId: GalleryId): void
}>()

const loading = ref(false)
const loadedOnce = ref(false)
const products = ref<ProductRecord[]>([])
const search = ref('')
const pagination = ref({ page: 1, pageSize: 10, total: 0 })

const isSelected = (id?: GalleryId | null): boolean => Boolean(id) && props.selectedProductIds.map(String).includes(String(id))

const loadProducts = async (): Promise<void> => {
  loading.value = true
  try {
    const payload = await productApi.list({
      page: pagination.value.page,
      page_size: pagination.value.pageSize,
      search: search.value.trim() || undefined,
      status: 'active',
    })
    products.value = Array.isArray(payload.products) ? payload.products : []
    pagination.value.total = Number(payload.pagination?.total ?? payload.total ?? 0)
    loadedOnce.value = true
  } catch (error) {
    console.error('Failed to load products:', error)
    products.value = []
    pagination.value.total = 0
  } finally {
    loading.value = false
  }
}

const reload = (): void => {
  pagination.value.page = 1
  void loadProducts()
}

const updatePage = (page: number): void => {
  pagination.value.page = page
  void loadProducts()
}

const updatePageSize = (pageSize: number): void => {
  pagination.value.pageSize = pageSize
  pagination.value.page = 1
  void loadProducts()
}

const toggle = (product: ProductRecord): void => {
  if (!product.id) return
  if (isSelected(product.id)) {
    emit('remove', product.id)
    return
  }
  emit('select', product)
}

watch(() => props.open, (isOpen) => {
  if (isOpen && !loadedOnce.value) {
    void loadProducts()
  }
})
</script>

