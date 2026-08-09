<template>
  <AdminTablePanel :loading="loading" :batch-visible="selectedProducts.length > 0">
    <template #batch>
      <div class="flex flex-wrap items-center justify-between gap-2">
        <span class="text-xs font-medium">已选择 {{ selectedProducts.length }} 个商品</span>
        <div class="flex flex-wrap gap-2">
          <Button v-if="canEdit" size="sm" @click="emit('batch-status', 'active')">
            <CircleCheck class="size-3.5" />
            批量上架
          </Button>
          <Button v-if="canEdit" variant="outline" size="sm" @click="emit('batch-status', 'inactive')">
            <CircleOff class="size-3.5" />
            批量下架
          </Button>
          <Button v-if="canDelete" variant="destructive" size="sm" @click="emit('batch-delete')">
            <Trash2 class="size-3.5" />
            批量删除
          </Button>
        </div>
      </div>
    </template>

    <Table class="min-w-[1220px]">
      <TableHeader>
        <TableRow>
          <TableHead class="w-11">
            <Checkbox
              :model-value="selectionState"
              aria-label="选择当前页商品"
              @update:model-value="emit('toggle-all-products', $event)"
            />
          </TableHead>
          <TableHead class="w-16">ID</TableHead>
          <TableHead class="w-56">SKU</TableHead>
          <TableHead class="w-20">图片</TableHead>
          <TableHead>商品名称</TableHead>
          <TableHead class="w-32">价格</TableHead>
          <TableHead class="w-24">库存</TableHead>
          <TableHead class="w-24">状态</TableHead>
          <TableHead class="w-20 text-center">精选</TableHead>
          <TableHead class="w-20">语言</TableHead>
          <TableHead class="w-44">创建时间</TableHead>
          <TableHead class="w-16 text-right">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableEmpty v-if="products.length === 0" :colspan="12">
          <div class="flex flex-col items-center text-muted-foreground">
            <PackageOpen class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无商品</span>
          </div>
        </TableEmpty>

        <TableRow v-for="product in products" :key="product.id">
          <TableCell>
            <Checkbox
              :model-value="isProductSelected(product.id)"
              :aria-label="`选择商品 ${product.name}`"
              @update:model-value="emit('toggle-product', product, $event)"
            />
          </TableCell>
          <TableCell class="font-mono text-[10px] font-bold text-muted-foreground">{{ product.id }}</TableCell>
          <TableCell class="max-w-56">
            <div v-if="skuEntries(product).length" class="flex max-w-56 flex-wrap gap-1" :title="skuTitle(product)">
              <span
                v-for="entry in skuEntries(product)"
                :key="entry.sku"
                :class="[
                  'max-w-full break-all rounded-md border px-1.5 py-0.5 font-mono text-[10px] font-bold leading-tight',
                  entry.isDefault
                    ? 'border-admin-selected-border bg-admin-selected-soft text-admin-selected'
                    : entry.isInactive
                      ? 'border-border/50 bg-muted/30 text-muted-foreground/60'
                      : 'border-border bg-muted/40 text-muted-foreground',
                ]"
              >
                {{ entry.sku }}
                <span v-if="entry.isDefault" class="ml-1 text-[9px] font-black">默认</span>
                <span v-else-if="entry.isInactive" class="ml-1 text-[9px] font-black">停用</span>
              </span>
            </div>
            <span v-else class="font-mono text-[11px] font-bold text-muted-foreground/80">-</span>
          </TableCell>
          <TableCell>
            <ProductThumbnail :product="product" />
          </TableCell>
          <TableCell class="max-w-72 truncate font-bold text-xs">{{ product.name }}</TableCell>
          <TableCell>
            <div class="flex items-baseline gap-1.5 tabular-nums">
              <span v-if="product.sale_price" class="font-mono text-xs font-bold text-destructive">{{ formatMoney(product.sale_price, product.currency) }}</span>
              <span :class="product.sale_price ? 'font-mono text-[10px] text-muted-foreground/70 line-through' : 'font-mono text-xs font-bold'">
                {{ formatMoney(product.price, product.currency) }}
              </span>
            </div>
          </TableCell>
          <TableCell>
            <AdminStatusBadge v-if="Number(product.stock) === 0" tone="coral">缺货</AdminStatusBadge>
            <AdminStatusBadge v-else-if="Number(product.stock) < 10" tone="amber">{{ product.stock }}</AdminStatusBadge>
            <span v-else class="font-mono text-xs font-bold tabular-nums">{{ product.stock }}</span>
          </TableCell>
          <TableCell>
            <AdminStatusBadge :tone="statusTone(product.status)">{{ getStatusName(product.status) }}</AdminStatusBadge>
          </TableCell>
          <TableCell class="text-center">
            <Star v-if="product.featured" class="mx-auto size-4 fill-amber-400 text-amber-500" aria-label="精选商品" />
            <span v-else class="text-muted-foreground/50">-</span>
          </TableCell>
          <TableCell class="font-bold text-xs">{{ localeName(product.locale) }}</TableCell>
          <TableCell class="font-mono text-[10px] text-muted-foreground/80">{{ formatDate(product.created_at) }}</TableCell>
          <TableCell class="text-right">
            <DropdownMenu>
              <DropdownMenuTrigger as-child>
                <Button variant="ghost" size="icon" :aria-label="`管理商品 ${product.name}`">
                  <MoreHorizontal class="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" class="w-40">
                <DropdownMenuItem v-if="canEdit" @select="emit('edit', product)">
                  <Pencil class="size-4" />
                  编辑
                </DropdownMenuItem>
                <DropdownMenuItem v-if="canSyncGoogle" @select="emit('sync-google', product)">
                  <Globe2 class="size-4" />
                  同步到 Google
                </DropdownMenuItem>
                <DropdownMenuItem v-if="canEdit" @select="emit('toggle-status', product)">
                  <CircleCheck v-if="product.status !== 'active'" class="size-4" />
                  <CircleOff v-else class="size-4" />
                  {{ product.status === 'active' ? '下架' : '上架' }}
                </DropdownMenuItem>
                <DropdownMenuSeparator v-if="canDelete" />
                <DropdownMenuItem
                  v-if="canDelete"
                  class="text-destructive focus:text-destructive"
                  @select="emit('delete', product)"
                >
                  <Trash2 class="size-4" />
                  删除
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <template #footer>
      <AdminPagination
        :page="pagination.page"
        :page-size="pagination.pageSize"
        :total="pagination.total"
        @update:page="emit('update-page', $event)"
        @update:page-size="emit('update-page-size', $event)"
      />
    </template>
  </AdminTablePanel>
</template>

<script setup lang="ts">
import { CircleCheck, CircleOff, Globe2, MoreHorizontal, PackageOpen, Pencil, Star, Trash2 } from '@lucide/vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import ProductThumbnail from '@/components/admin/product/ProductThumbnail.vue'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import type { ProductPagination, ProductRecord } from './productEditorTypes'

type SelectionState = boolean | 'indeterminate'
type ProductStatusAction = 'active' | 'inactive'
interface ProductSkuEntry {
  sku: string
  isDefault: boolean
  isInactive: boolean
}

const props = withDefaults(defineProps<{
  loading?: boolean
  products?: ProductRecord[]
  selectedProducts?: ProductRecord[]
  pagination: ProductPagination
  selectionState?: SelectionState
  canEdit?: boolean
  canDelete?: boolean
  canSyncGoogle?: boolean
  localeName: (locale?: string) => string
}>(), {
  loading: false,
  products: () => [],
  selectedProducts: () => [],
  selectionState: false,
  canEdit: false,
  canDelete: false,
  canSyncGoogle: false,
})

const emit = defineEmits<{
  (event: 'batch-status', status: ProductStatusAction): void
  (event: 'batch-delete'): void
  (event: 'toggle-all-products', value: SelectionState): void
  (event: 'toggle-product', product: ProductRecord, value: SelectionState): void
  (event: 'edit', product: ProductRecord): void
  (event: 'sync-google', product: ProductRecord): void
  (event: 'toggle-status', product: ProductRecord): void
  (event: 'delete', product: ProductRecord): void
  (event: 'update-page', page: number): void
  (event: 'update-page-size', pageSize: number): void
}>()

const isProductSelected = (productId: ProductRecord['id']): boolean => props.selectedProducts.some((product) => product.id === productId)
const getStatusName = (status?: string): string => ({ active: '在售', inactive: '下架', out_of_stock: '缺货' } as Record<string, string>)[status || ''] || status || '-'
const statusTone = (status?: string): AdminStatusTone => ({ active: 'green', inactive: 'gray', out_of_stock: 'coral' } as Record<string, AdminStatusTone>)[status || ''] || 'gray'
const skuEntries = (product: ProductRecord): ProductSkuEntry[] => {
  const variants = Array.isArray(product.variants) ? product.variants : []
  const entries: ProductSkuEntry[] = []
  const seen = new Set<string>()

  for (const variant of variants) {
    const sku = String(variant?.sku || '').trim()
    if (!sku || seen.has(sku)) continue
    seen.add(sku)
    entries.push({
      sku,
      isDefault: variant?.is_default === true || sku === product.sku,
      isInactive: variant?.is_active === false,
    })
  }

  const fallbackSku = String(product.sku || '').trim()
  if (fallbackSku && !seen.has(fallbackSku)) {
    entries.unshift({ sku: fallbackSku, isDefault: true, isInactive: false })
  }

  return entries
}
const skuTitle = (product: ProductRecord): string => {
  return skuEntries(product)
    .map((entry) => `${entry.sku}${entry.isDefault ? ' (默认)' : entry.isInactive ? ' (停用)' : ''}`)
    .join('\n')
}
const formatDate = (dateString: unknown): string => dateString ? new Date(dateString as string | number | Date).toLocaleString('zh-CN') : '-'
const formatMoney = (amount: unknown, currency = 'USD'): string => {
  const normalizedCurrency = String(currency || 'USD').trim().toUpperCase()
  const value = Number(amount || 0)
  try {
    return new Intl.NumberFormat('zh-CN', { style: 'currency', currency: normalizedCurrency }).format(value)
  } catch {
    return `${normalizedCurrency} ${value.toFixed(2)}`
  }
}
</script>
