<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="产品履约要求"
      description="按真实商品和 Variant 显式维护编轮质检张力表规则"
    >
      <template #actions>
        <Button variant="outline" size="icon" aria-label="刷新产品履约要求" title="刷新产品履约要求" :disabled="loading" @click="fetchProducts">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
        </Button>
        <Button variant="outline" as-child>
          <RouterLink to="/catalog/products">
            <Package class="size-4" />
            商品管理
          </RouterLink>
        </Button>
      </template>
    </AdminPageHeader>

    <div class="rounded-lg border border-blue-500/20 bg-blue-500/5 px-4 py-3">
      <div class="flex items-start gap-2">
        <ShieldCheck class="mt-0.5 size-4 shrink-0 text-blue-600 dark:text-blue-400" />
        <div class="min-w-0 text-xs leading-5">
          <p class="font-bold">规则入口独立于订单金额和 QUICKBUY</p>
          <p class="mt-1 text-muted-foreground">
            订单金额只负责高价触发；张力表只由这里保存的商品级或 Variant 级规则触发。未配置规则不会因为商品分类或名称自动加入。
          </p>
        </div>
      </div>
    </div>

    <AdminFilterPanel>
      <form class="grid gap-3 md:grid-cols-[minmax(260px,1.5fr)_180px_auto]" @submit.prevent="applyFilters">
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SEARCH / 搜索</span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
            <Input v-model="filters.search" class="h-9 pl-9" placeholder="商品名称、SKU 或描述" />
          </div>
        </label>

        <AdminFilterSelect v-model="filters.status" label="商品状态" :options="statusOptions" />

        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9 rounded-full px-4 text-xs font-black uppercase tracking-wider">
            <Search class="size-3.5" />
            搜索
          </Button>
          <Button type="button" variant="outline" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" @click="resetFilters">
            <RotateCcw class="size-3.5" />
            重置
          </Button>
        </div>
      </form>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[980px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-16">ID</TableHead>
            <TableHead class="w-72">商品</TableHead>
            <TableHead class="w-64">SKU / Variant</TableHead>
            <TableHead class="w-28">商品状态</TableHead>
            <TableHead class="w-32">维护范围</TableHead>
            <TableHead class="w-24 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="products.length === 0" :colspan="6">
            <div class="flex flex-col items-center text-muted-foreground">
              <ShieldCheck class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无匹配商品</span>
            </div>
          </TableEmpty>

          <TableRow v-for="product in products" :key="String(product.id)">
            <TableCell class="font-mono text-[10px] font-bold text-muted-foreground">{{ product.id }}</TableCell>
            <TableCell>
              <div class="min-w-0">
                <p class="truncate text-xs font-black">{{ product.name || `商品 #${product.id}` }}</p>
                <p class="mt-1 truncate text-[10px] text-muted-foreground">{{ product.short_description || '未填写简短描述' }}</p>
              </div>
            </TableCell>
            <TableCell>
              <div class="min-w-0">
                <p class="truncate font-mono text-xs font-bold">{{ primarySku(product) }}</p>
                <p class="mt-1 text-[10px] text-muted-foreground">{{ variantCount(product) }} 个 Variant 可单独配置</p>
              </div>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="statusTone(product.status)">{{ statusName(product.status) }}</AdminStatusBadge>
            </TableCell>
            <TableCell>
              <span class="text-xs font-bold">商品级 + Variant 级</span>
              <span class="mt-1 block text-[10px] text-muted-foreground">打开后读取当前显式规则</span>
            </TableCell>
            <TableCell class="text-right">
              <Button
                v-if="canView"
                variant="outline"
                size="icon"
                class="size-8 rounded-full"
                :aria-label="`配置 ${product.name || `商品 ${product.id}`} 履约要求`"
                title="配置履约要求"
                @click="openProduct(product)"
              >
                <Settings2 class="size-3.5" />
              </Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <template #footer>
        <AdminPagination
          :page="pagination.page"
          :page-size="pagination.pageSize"
          :total="pagination.total"
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </template>
    </AdminTablePanel>

    <ProductFulfillmentRequirementDialog
      v-model:open="dialogOpen"
      :product="selectedProduct"
      :can-edit="canEdit"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { toast } from 'vue-sonner'
import { Package, RefreshCw, RotateCcw, Search, Settings2, ShieldCheck } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminFilterSelect from '@/components/admin/AdminFilterSelect.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import ProductFulfillmentRequirementDialog from '@/components/admin/product/ProductFulfillmentRequirementDialog.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import productApi from '@/api/products'
import { useAuthStore } from '@/stores/auth'
import type { ProductRecord } from '@/modules/product/productEditorTypes'

const authStore = useAuthStore()
const loading = ref(false)
const dialogOpen = ref(false)
const selectedProduct = ref<ProductRecord | null>(null)
const products = ref<ProductRecord[]>([])
const requestSequence = ref(0)
const filters = reactive({
  search: '',
  status: 'all',
})
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const statusOptions = [
  { label: '全部状态', value: 'all' },
  { label: '在售', value: 'active' },
  { label: '下架', value: 'inactive' },
  { label: '缺货', value: 'out_of_stock' },
]
const canView = computed(() => authStore.hasPermission('product:view'))
const canEdit = computed(() => authStore.hasPermission('product:edit'))

const errorMessage = (error: unknown, fallback: string): string => {
  const responseError = (error as { response?: { data?: { error?: unknown } } })?.response?.data?.error
  return typeof responseError === 'string' && responseError.trim() ? responseError : fallback
}

const primarySku = (product: ProductRecord): string => {
  const variants = Array.isArray(product.variants) ? product.variants : []
  const defaultVariant = variants.find((variant) => variant?.is_default)
  return String(defaultVariant?.sku || product.sku || '未设置 SKU')
}

const variantCount = (product: ProductRecord): number => (
  Array.isArray(product.variants) ? product.variants.filter((variant) => variant?.id != null).length : 0
)

const statusName = (status?: string): string => ({
  active: '在售',
  inactive: '下架',
  out_of_stock: '缺货',
}[status || ''] || status || '未知')

const statusTone = (status?: string): AdminStatusTone => ({
  active: 'green',
  inactive: 'gray',
  out_of_stock: 'coral',
} as Record<string, AdminStatusTone>)[status || ''] || 'gray'

const fetchProducts = async (): Promise<void> => {
  const sequence = requestSequence.value + 1
  requestSequence.value = sequence
  loading.value = true
  try {
    const payload = await productApi.list({
      page: pagination.page,
      page_size: pagination.pageSize,
      ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
      ...(filters.status !== 'all' ? { status: filters.status } : {}),
    })
    if (sequence !== requestSequence.value) return
    products.value = payload.products as ProductRecord[]
    const page = payload.pagination as { page: number; page_size: number; total: number }
    pagination.page = page.page
    pagination.pageSize = page.page_size
    pagination.total = page.total
  } catch (error) {
    if (sequence === requestSequence.value) toast.error(errorMessage(error, '商品列表加载失败'))
  } finally {
    if (sequence === requestSequence.value) loading.value = false
  }
}

const applyFilters = (): void => {
  pagination.page = 1
  void fetchProducts()
}

const resetFilters = (): void => {
  Object.assign(filters, { search: '', status: 'all' })
  pagination.page = 1
  void fetchProducts()
}

const updatePage = (page: number): void => {
  pagination.page = page
  void fetchProducts()
}

const updatePageSize = (pageSize: number): void => {
  pagination.pageSize = pageSize
  pagination.page = 1
  void fetchProducts()
}

const openProduct = (product: ProductRecord): void => {
  if (!product.id) return
  selectedProduct.value = product
  dialogOpen.value = true
}

onMounted(() => {
  void fetchProducts()
})
</script>
