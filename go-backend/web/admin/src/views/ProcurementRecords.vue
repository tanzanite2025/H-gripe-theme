<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="商品成本"
      description="独立维护 SKU 的采购价、供应商资料、到货周期、起订量和附加成本，不写入商品目录。"
    >
      <template #actions>
        <Button v-if="canCreate" @click="openCreate">
          <Plus class="size-4" />
          新增商品成本
        </Button>
      </template>
    </AdminPageHeader>

    <AdminFilterPanel>
      <form
        class="grid gap-3 md:grid-cols-[minmax(260px,1.5fr)_auto]"
        @submit.prevent="applyFilters"
      >
        <label class="block space-y-1">
            <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
            SEARCH / 搜索
          </span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
            <Input v-model="filters.search" class="h-9 pl-9" placeholder="产品编码、名称、供应商或联系人" />
          </div>
        </label>

        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9 rounded-full px-4 text-xs font-black uppercase tracking-wider">
            <Search class="size-3.5" />
            搜索
          </Button>
          <Button
            type="button"
            variant="outline"
            class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider"
            @click="resetFilters"
          >
            <RotateCcw class="size-3.5" />
            重置
          </Button>
        </div>
      </form>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[1120px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-64">产品</TableHead>
            <TableHead class="w-52">来源与联系人</TableHead>
            <TableHead class="w-40">采购价</TableHead>
            <TableHead class="w-40">附加成本</TableHead>
            <TableHead class="w-28">到货周期</TableHead>
            <TableHead class="w-28">起订量</TableHead>
            <TableHead class="w-36">更新于</TableHead>
            <TableHead class="w-20 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="records.length === 0" :colspan="8">
            <div class="flex flex-col items-center text-muted-foreground">
              <ClipboardList class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无商品成本资料</span>
            </div>
          </TableEmpty>

          <TableRow v-for="record in records" :key="record.id">
            <TableCell>
              <div class="min-w-0">
                <p class="truncate text-xs font-black">{{ record.product_name }}</p>
                <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground">{{ record.product_code }}</p>
              </div>
            </TableCell>
            <TableCell>
              <div class="min-w-0">
                <p class="truncate text-xs font-bold">{{ record.supplier_name || '未填写来源' }}</p>
                <p v-if="record.supplier_contact_name" class="mt-1 truncate text-[10px] text-muted-foreground">
                  {{ record.supplier_contact_name }}
                </p>
                <p v-if="record.supplier_phone || record.supplier_email" class="mt-1 truncate text-[10px] text-muted-foreground/70">
                  {{ record.supplier_phone || record.supplier_email }}
                </p>
              </div>
            </TableCell>
            <TableCell>
              <span class="font-mono text-xs font-black tabular-nums">
                {{ formatMoney(record.purchase_price, record.currency) }}
              </span>
            </TableCell>
            <TableCell class="font-mono text-[10px] text-muted-foreground">
              <div class="space-y-0.5">
                <p>运费 {{ formatMoney(record.inbound_shipping_unit_cost, record.currency) }}</p>
                <p>包装 {{ formatMoney(record.packaging_unit_cost, record.currency) }}</p>
                <p>其他 {{ formatMoney(record.other_unit_cost, record.currency) }}</p>
              </div>
            </TableCell>
            <TableCell class="font-mono text-xs font-bold tabular-nums">
              {{ record.lead_time_days > 0 ? `${record.lead_time_days} 天` : '待确认' }}
            </TableCell>
            <TableCell class="font-mono text-xs font-bold tabular-nums">
              {{ record.minimum_order_quantity }}
            </TableCell>
            <TableCell class="font-mono text-[10px] text-muted-foreground/80">
              {{ formatDate(record.updated_at || record.created_at) }}
            </TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon" :aria-label="`管理商品成本 ${record.product_name}`">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" class="w-36">
                  <DropdownMenuItem v-if="canEdit" @select="openEdit(record)">
                    <Pencil class="size-4" />
                    编辑
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="canDelete" />
                  <DropdownMenuItem
                    v-if="canDelete"
                    class="text-destructive focus:text-destructive"
                    @select="removeRecord(record)"
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
          :page-size="pagination.page_size"
          :total="pagination.total"
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </template>
    </AdminTablePanel>

    <Dialog v-model:open="dialogOpen">
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>{{ form.id ? '编辑商品成本' : '新增商品成本' }}</DialogTitle>
          <DialogDescription>
            新增时从真实商品变体中选择 SKU；名称和变体标题由商品目录只读回填，采购币种与成本资料仍由本附加域独立维护。
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-5" @submit.prevent="save">
          <section class="space-y-3">
            <div>
              <h2 class="text-sm font-black tracking-tight">绑定真实商品</h2>
              <p class="mt-1 text-[10px] text-muted-foreground">成本资料按真实商品变体 SKU 保存，不加载库存、媒体或商品详情。</p>
            </div>
            <div class="flex min-h-20 items-center justify-between gap-4 rounded-lg border bg-muted/20 px-3 py-3">
              <div class="min-w-0">
                <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">当前商品变体</p>
                <template v-if="selectedProductOption">
                  <p class="mt-1 truncate text-xs font-black">{{ selectedProductOption.product_name }}</p>
                  <p v-if="selectedProductOption.variant_title" class="mt-1 truncate text-[10px] text-muted-foreground">
                    {{ selectedProductOption.variant_title }}
                  </p>
                  <p class="mt-1 font-mono text-[10px] text-muted-foreground">{{ selectedProductOption.sku }}</p>
                </template>
                <template v-else-if="form.id && historicalProductUnavailable">
                  <p class="mt-1 text-xs font-bold text-amber-700 dark:text-amber-300">原商品当前不可用</p>
                  <p class="mt-1 break-words font-mono text-[10px] text-muted-foreground">{{ form.sku }} · {{ form.product_name }}</p>
                </template>
                <template v-else-if="form.id">
                  <p class="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                    <LoaderCircle class="size-3.5 animate-spin" />
                    正在确认历史商品 SKU...
                  </p>
                </template>
                <template v-else>
                  <p class="mt-1 text-xs font-bold text-muted-foreground">尚未选择商品变体 SKU</p>
                  <p class="mt-1 text-[10px] text-muted-foreground/80">点击右侧按钮，从真实商品目录中搜索并选择。</p>
                </template>
              </div>
              <div class="flex shrink-0 items-center gap-2">
                <span
                  v-if="selectedProductOption"
                  class="hidden text-[10px] font-bold text-muted-foreground sm:inline"
                >
                  {{ form.id
                    ? (selectedProductOption.available ? '当前可用' : '原商品当前不可用')
                    : '已选择' }}
                </span>
                <Button
                  v-if="!form.id"
                  type="button"
                  variant="outline"
                  size="sm"
                  :disabled="saving"
                  @click="openProductPicker"
                >
                  <Search class="size-3.5" />
                  {{ selectedProductOption ? '更换商品' : '选择商品' }}
                </Button>
              </div>
            </div>
            <div v-if="form.id && historicalProductUnavailable" class="rounded-lg border border-amber-500/30 bg-amber-500/10 px-3 py-2.5 text-xs leading-5 text-amber-800 dark:text-amber-200">
              <div class="flex items-start gap-2">
                <TriangleAlert class="mt-0.5 size-4 shrink-0" />
                <div class="min-w-0">
                  <p class="font-bold">原商品当前不可用</p>
                  <p class="mt-1 break-words">保留历史成本记录的 SKU 和名称快照，不会静默改绑其他商品。</p>
                </div>
              </div>
            </div>
          </section>

          <section class="space-y-3">
            <div>
              <h2 class="text-sm font-black tracking-tight">成本与来源</h2>
              <p class="mt-1 text-[10px] text-muted-foreground">采购价与供应商资料只在这个独立附加域中维护。</p>
            </div>
            <div class="grid gap-3 md:grid-cols-2">
              <AdminFormField label="采购价" required>
                <Input v-model="form.purchase_price" :disabled="saving" type="number" min="0" step="0.01" />
              </AdminFormField>
              <AdminFormField label="币种">
                <Select v-model="form.currency" :disabled="saving">
                  <SelectTrigger>
                    <SelectValue placeholder="选择币种" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="currency in currencies" :key="currency" :value="currency">
                      {{ currency }}
                    </SelectItem>
                  </SelectContent>
                </Select>
              </AdminFormField>
              <AdminFormField label="供应商名称">
                <Input v-model="form.supplier_name" :disabled="saving" placeholder="例如 Xiamen Wheel Factory" />
              </AdminFormField>
              <AdminFormField label="联系人">
                <Input v-model="form.supplier_contact_name" :disabled="saving" />
              </AdminFormField>
              <AdminFormField label="联系电话">
                <Input v-model="form.supplier_phone" :disabled="saving" type="tel" />
              </AdminFormField>
              <AdminFormField label="联系邮箱">
                <Input v-model="form.supplier_email" :disabled="saving" type="email" />
              </AdminFormField>
            </div>
          </section>

          <section class="space-y-3">
            <div>
              <h2 class="text-sm font-black tracking-tight">到货与起订量</h2>
              <p class="mt-1 text-[10px] text-muted-foreground">只记录产品资料，不生成采购单或库存数量。</p>
            </div>
            <div class="grid gap-3 md:grid-cols-2">
              <AdminFormField label="到货周期">
                <div class="relative">
                  <Input
                    v-model="form.lead_time_days"
                    :disabled="saving"
                    class="pr-12"
                    type="number"
                    min="0"
                    step="1"
                  />
                  <span class="pointer-events-none absolute inset-y-0 right-8 flex items-center text-xs font-bold text-muted-foreground" aria-hidden="true">
                    天
                  </span>
                </div>
              </AdminFormField>
              <AdminFormField label="起订量">
                <Input v-model="form.minimum_order_quantity" :disabled="saving" type="number" min="1" step="1" />
              </AdminFormField>
            </div>
            <div class="grid gap-3 border-t border-border/60 pt-3 md:grid-cols-3">
              <AdminFormField label="入库运费 / 件">
                <Input v-model="form.inbound_shipping_unit_cost" :disabled="saving" type="number" min="0" step="0.01" />
              </AdminFormField>
              <AdminFormField label="包装 / 件">
                <Input v-model="form.packaging_unit_cost" :disabled="saving" type="number" min="0" step="0.01" />
              </AdminFormField>
              <AdminFormField label="其他 / 件">
                <Input v-model="form.other_unit_cost" :disabled="saving" type="number" min="0" step="0.01" />
              </AdminFormField>
            </div>
          </section>

          <DialogFooter>
            <Button type="button" variant="outline" @click="dialogOpen = false">取消</Button>
            <Button type="submit" :disabled="saving || (form.id ? !canEdit : !canCreate)">
              <LoaderCircle v-if="saving" class="size-4 animate-spin" />
              保存商品成本
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="productPickerOpen">
      <DialogContent size="md">
        <DialogHeader>
          <DialogTitle>选择商品变体 SKU</DialogTitle>
          <DialogDescription>
            搜索真实商品目录，选择后会回填到商品成本表单。
          </DialogDescription>
        </DialogHeader>

        <div class="space-y-3">
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground/60" />
            <Input
              v-model="productSearch"
              autofocus
              :disabled="saving"
              class="h-10 pl-9"
              placeholder="搜索 SKU、商品名称或变体标题"
              @update:model-value="scheduleProductOptionsSearch"
            />
          </div>
          <div class="max-h-[min(52vh,28rem)] overflow-y-auto rounded-lg border bg-background">
            <div v-if="productOptionsLoading" class="flex items-center gap-2 px-3 py-5 text-xs text-muted-foreground">
              <LoaderCircle class="size-4 animate-spin" />
              正在读取商品选项...
            </div>
            <div v-else-if="productOptions.length === 0" class="px-3 py-5 text-xs text-muted-foreground">
              {{ productSearch.trim() ? '没有找到可用商品变体' : '输入关键词开始搜索商品' }}
            </div>
            <div v-else class="divide-y">
              <button
                v-for="option in productOptions"
                :key="option.sku"
                type="button"
                class="block w-full px-3 py-3 text-left transition-colors hover:bg-muted/60"
                :disabled="saving"
                @click="selectProductOption(option)"
              >
                <span class="flex items-start justify-between gap-3">
                  <span class="min-w-0">
                    <span class="block truncate text-xs font-black">{{ option.product_name }}</span>
                    <span v-if="option.variant_title" class="mt-0.5 block truncate text-[10px] text-muted-foreground">
                      {{ option.variant_title }}
                    </span>
                  </span>
                  <span class="shrink-0 font-mono text-[10px] font-bold text-foreground">{{ option.sku }}</span>
                </span>
              </button>
            </div>
          </div>
          <Button
            v-if="productOptionPagination.page < productOptionPagination.total_pages && productOptions.length"
            type="button"
            variant="outline"
            class="w-full"
            :disabled="productOptionsLoading || saving"
            @click="loadMoreProductOptions"
          >
            <LoaderCircle v-if="productOptionsLoading" class="size-4 animate-spin" />
            继续加载商品
          </Button>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="productPickerOpen = false">取消</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  ClipboardList,
  LoaderCircle,
  MoreHorizontal,
  Pencil,
  Plus,
  RotateCcw,
  Search,
  Trash2,
  TriangleAlert,
} from '@lucide/vue'
import procurementApi, {
  type ProcurementProductOption,
  type ProcurementRecord,
  type ProcurementRecordCreatePayload,
  type ProcurementRecordDetailsPayload,
} from '@/api/procurement'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { useAuthStore } from '@/stores/auth'

interface ProcurementForm extends Omit<ProcurementRecordDetailsPayload, 'purchase_price'> {
  id?: number
  sku: string
  product_name: string
  purchase_price: string | number
}

const authStore = useAuthStore()
const canCreate = computed(() => authStore.hasPermission('procurement:create'))
const canEdit = computed(() => authStore.hasPermission('procurement:edit'))
const canDelete = computed(() => authStore.hasPermission('procurement:delete'))
const currencies = ['USD', 'CNY', 'EUR', 'GBP', 'JPY']

const records = ref<ProcurementRecord[]>([])
const loading = ref(false)
const saving = ref(false)
const dialogOpen = ref(false)
const productPickerOpen = ref(false)
const filters = reactive({ search: '' })
const pagination = reactive({ page: 1, page_size: 20, total: 0, total_pages: 0 })
const form = reactive<ProcurementForm>(emptyForm())
const productSearch = ref('')
const productOptions = ref<ProcurementProductOption[]>([])
const selectedProductOption = ref<ProcurementProductOption | null>(null)
const productOptionsLoading = ref(false)
const productOptionPagination = reactive({ page: 1, page_size: 20, total: 0, total_pages: 0 })
const productOptionRequestVersion = ref(0)
const historicalProductRequestVersion = ref(0)
let productOptionSearchTimer: ReturnType<typeof setTimeout> | undefined
const historicalProductUnavailable = ref(false)

function emptyForm(): ProcurementForm {
  return {
    sku: '',
    product_name: '',
    purchase_price: '',
    currency: 'USD',
    supplier_name: '',
    supplier_contact_name: '',
    supplier_phone: '',
    supplier_email: '',
    lead_time_days: 0,
    minimum_order_quantity: 1,
    inbound_shipping_unit_cost: 0,
    packaging_unit_cost: 0,
    other_unit_cost: 0,
  }
}

const assignForm = (record?: ProcurementRecord): void => {
  Object.assign(form, emptyForm())
  delete form.id
  if (!record) return

  Object.assign(form, {
    id: record.id,
    sku: record.product_code,
    product_name: record.product_name,
    purchase_price: record.purchase_price,
    currency: record.currency,
    supplier_name: record.supplier_name,
    supplier_contact_name: record.supplier_contact_name,
    supplier_phone: record.supplier_phone,
    supplier_email: record.supplier_email,
    lead_time_days: record.lead_time_days,
    minimum_order_quantity: record.minimum_order_quantity,
    inbound_shipping_unit_cost: record.inbound_shipping_unit_cost,
    packaging_unit_cost: record.packaging_unit_cost,
    other_unit_cost: record.other_unit_cost,
  })
}

const loadProductOptions = async (reset = true): Promise<void> => {
  if (reset) {
    productOptionPagination.page = 1
    productOptions.value = []
  }
  const requestVersion = productOptionRequestVersion.value + 1
  productOptionRequestVersion.value = requestVersion
  productOptionsLoading.value = true
  try {
    const payload = await procurementApi.listProductOptions({
      page: productOptionPagination.page,
      page_size: productOptionPagination.page_size,
      search: productSearch.value.trim(),
    })
    if (requestVersion !== productOptionRequestVersion.value) return
    productOptions.value = reset ? payload.options : [...productOptions.value, ...payload.options]
    Object.assign(productOptionPagination, payload.pagination)
  } catch (error) {
    if (requestVersion !== productOptionRequestVersion.value) return
    console.error('Failed to fetch procurement product options:', error)
    toast.error('商品选项加载失败')
  } finally {
    if (requestVersion === productOptionRequestVersion.value) productOptionsLoading.value = false
  }
}

const loadMoreProductOptions = async (): Promise<void> => {
  if (productOptionPagination.page >= productOptionPagination.total_pages) return
  productOptionPagination.page += 1
  await loadProductOptions(false)
}

const openProductPicker = (): void => {
  if (form.id) return
  if (productOptionSearchTimer) {
    clearTimeout(productOptionSearchTimer)
    productOptionSearchTimer = undefined
  }
  productPickerOpen.value = true
  productSearch.value = ''
  productOptions.value = []
  productOptionPagination.page = 1
  productOptionPagination.total = 0
  productOptionPagination.total_pages = 0
  void loadProductOptions(true)
}

const scheduleProductOptionsSearch = (): void => {
  if (productOptionSearchTimer) clearTimeout(productOptionSearchTimer)
  productOptionSearchTimer = setTimeout(() => {
    void loadProductOptions(true)
  }, 250)
}

const selectProductOption = (option: ProcurementProductOption): void => {
  if (productOptionSearchTimer) {
    clearTimeout(productOptionSearchTimer)
    productOptionSearchTimer = undefined
  }
  productOptionRequestVersion.value += 1
  productOptionsLoading.value = false
  selectedProductOption.value = option
  historicalProductUnavailable.value = false
  form.sku = option.sku
  form.product_name = option.product_name
  productSearch.value = ''
  productOptions.value = []
  productPickerOpen.value = false
}

const loadHistoricalProductOption = async (record: ProcurementRecord): Promise<void> => {
  const requestVersion = historicalProductRequestVersion.value + 1
  historicalProductRequestVersion.value = requestVersion
  try {
    const payload = await procurementApi.listProductOptions({ sku: record.product_code, page: 1, page_size: 1 })
    if (requestVersion !== historicalProductRequestVersion.value) return
    selectedProductOption.value = payload.options[0] || null
    historicalProductUnavailable.value = !selectedProductOption.value
  } catch (error) {
    if (requestVersion !== historicalProductRequestVersion.value) return
    console.error('Failed to resolve historical procurement product option:', error)
    selectedProductOption.value = null
    historicalProductUnavailable.value = true
  }
}

const fetchRecords = async (): Promise<void> => {
  loading.value = true
  try {
    const payload = await procurementApi.list({
      page: pagination.page,
      page_size: pagination.page_size,
      search: filters.search.trim(),
    })
    records.value = payload.records
    Object.assign(pagination, payload.pagination)
  } catch (error) {
    console.error('Failed to fetch procurement records:', error)
    toast.error('商品成本加载失败')
  } finally {
    loading.value = false
  }
}

const applyFilters = (): void => {
  pagination.page = 1
  void fetchRecords()
}

const resetFilters = (): void => {
  filters.search = ''
  applyFilters()
}

const updatePage = (page: number): void => {
  pagination.page = page
  void fetchRecords()
}

const updatePageSize = (pageSize: number): void => {
  pagination.page_size = pageSize
  pagination.page = 1
  void fetchRecords()
}

const openCreate = (): void => {
  historicalProductRequestVersion.value += 1
  productPickerOpen.value = false
  assignForm()
  productSearch.value = ''
  productOptions.value = []
  selectedProductOption.value = null
  historicalProductUnavailable.value = false
  dialogOpen.value = true
}

const openEdit = (record: ProcurementRecord): void => {
  historicalProductRequestVersion.value += 1
  productPickerOpen.value = false
  assignForm(record)
  productSearch.value = ''
  productOptions.value = []
  selectedProductOption.value = null
  historicalProductUnavailable.value = false
  dialogOpen.value = true
  void loadHistoricalProductOption(record)
}

const save = async (): Promise<void> => {
  if (!canEdit.value && !(!form.id && canCreate.value)) return
  const rawPurchasePrice = String(form.purchase_price).trim()
  const purchasePrice = rawPurchasePrice === ''
    ? null
    : Number(rawPurchasePrice)
  if ((!form.id && !selectedProductOption.value) || !form.sku.trim() || purchasePrice == null || Number.isNaN(purchasePrice)) {
    toast.error(form.id ? '请确认商品快照并填写采购价' : '请选择真实商品变体 SKU 并填写采购价')
    return
  }
  saving.value = true
  try {
    const detailsPayload: ProcurementRecordDetailsPayload = {
      purchase_price: purchasePrice,
      currency: form.currency,
      supplier_name: form.supplier_name.trim(),
      supplier_contact_name: form.supplier_contact_name.trim(),
      supplier_phone: form.supplier_phone.trim(),
      supplier_email: form.supplier_email.trim(),
      lead_time_days: Number(form.lead_time_days || 0),
      minimum_order_quantity: Number(form.minimum_order_quantity || 1),
      inbound_shipping_unit_cost: Number(form.inbound_shipping_unit_cost || 0),
      packaging_unit_cost: Number(form.packaging_unit_cost || 0),
      other_unit_cost: Number(form.other_unit_cost || 0),
    }
    if (form.id) {
      await procurementApi.update(form.id, detailsPayload)
    } else {
      const payload: ProcurementRecordCreatePayload = {
        sku: form.sku.trim(),
        ...detailsPayload,
      }
      await procurementApi.create(payload)
    }
    dialogOpen.value = false
    toast.success('商品成本已保存')
    await fetchRecords()
  } catch (error) {
    console.error('Failed to save procurement record:', error)
    toast.error('商品成本保存失败')
  } finally {
    saving.value = false
  }
}

const removeRecord = async (record: ProcurementRecord): Promise<void> => {
  if (!window.confirm(`确定删除“${record.product_name}”的商品成本？`)) return
  try {
    await procurementApi.remove(record.id)
    toast.success('商品成本已删除')
    if (records.value.length === 1 && pagination.page > 1) pagination.page -= 1
    await fetchRecords()
  } catch (error) {
    console.error('Failed to delete procurement record:', error)
    toast.error('商品成本删除失败')
  }
}

const formatMoney = (amount: number, currency: string): string => {
  try {
    return new Intl.NumberFormat('zh-CN', { style: 'currency', currency }).format(Number(amount || 0))
  } catch {
    return `${currency} ${Number(amount || 0).toFixed(2)}`
  }
}

const formatDate = (value?: string): string => value ? new Date(value).toLocaleString('zh-CN') : '-'

onMounted(() => {
  void fetchRecords()
})

onBeforeUnmount(() => {
  if (productOptionSearchTimer) clearTimeout(productOptionSearchTimer)
})
</script>
