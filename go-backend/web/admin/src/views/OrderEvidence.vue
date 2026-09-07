<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="履约证据"
      description="所有订单的配置、身份、包装、签收与条件质检记录"
    >
      <template #actions>
        <Button
          v-if="returnRoute"
          variant="outline"
          size="sm"
          class="rounded-full"
          @click="returnToDispute"
        >
          <ArrowLeft class="size-3.5" />
          返回拒付工作台
        </Button>
        <Button
          variant="outline"
          size="icon"
          aria-label="刷新履约证据"
          title="刷新履约证据"
          :disabled="loading"
          @click="refreshCurrent"
        >
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <AdminFilterPanel>
      <form class="grid grid-cols-1 gap-3 md:grid-cols-[minmax(220px,1fr)_180px_180px_180px_auto]" @submit.prevent="applyFilters">
        <label class="block space-y-1">
          <span class="filter-label">SEARCH / 搜索</span>
          <Input v-model="filters.search" placeholder="订单号、客户或 Email" />
        </label>
        <label class="block space-y-1">
          <span class="filter-label">PACKAGE / 证据包</span>
          <select v-model="filters.package_status" class="filter-select">
            <option value="">全部状态</option>
            <option value="incomplete">待补全</option>
            <option value="ready">可锁定</option>
            <option value="locked">已锁定</option>
            <option value="superseded">已修订</option>
          </select>
        </label>
        <label class="block space-y-1">
          <span class="filter-label">HIGH VALUE / 高价</span>
          <select v-model="filters.high_value" class="filter-select">
            <option value="all">全部订单</option>
            <option value="true">订单 ≥ 750 USD</option>
            <option value="false">低于 750 USD</option>
          </select>
        </label>
        <label class="block space-y-1">
          <span class="filter-label">SPOKE QC / 张力表</span>
          <select v-model="filters.spoke_tension_qc" class="filter-select">
            <option value="all">全部订单</option>
            <option value="true">需要张力表</option>
            <option value="false">不需要张力表</option>
          </select>
        </label>
        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9 px-4 text-xs font-black uppercase tracking-wider">
            查询
          </Button>
          <Button type="button" variant="outline" class="h-9 px-4 text-xs font-black uppercase tracking-wider" @click="resetFilters">
            重置
          </Button>
        </div>
      </form>
    </AdminFilterPanel>

    <OrderEvidenceTablePanel
      :orders="orders"
      :loading="loading"
      :selected-order-id="selectedOrder?.order_id"
      :pagination="pagination"
      @select-order="openOrder"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />

    <OrderEvidenceDetailDialog
      v-model:open="detailOpen"
      :loading="detailLoading"
      :result="detailResult"
      :order="selectedOrder"
      :can-edit="canEdit"
      :saving-item-id="savingItemId"
      :locking="locking"
      :revising="revising"
      :exporting="exporting"
      @edit-item="openItemEditor"
      @lock="requestLock"
      @revision="requestRevision"
      @export="exportSnapshot"
    />

    <OrderEvidenceItemDialog
      v-model:open="itemDialogOpen"
      :item="editingItem"
      :can-edit="canEdit"
      :saving="savingItemId !== null"
      :uploading-attachment="uploadingAttachmentItemId === editingItem?.id"
      :deleting-attachment-id="deletingAttachmentID"
      :attachment-url="attachmentUrlForEditingItem"
      @upload-attachment="uploadAttachment"
      @delete-attachment="requestDeleteAttachment"
      @submit="saveItem"
    />

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      :destructive="false"
      @confirm="executeConfirmation"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ArrowLeft, FileCheck2, ListChecks, RefreshCw, ShieldCheck, TriangleAlert } from '@lucide/vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import OrderEvidenceDetailDialog from '@/components/admin/order/OrderEvidenceDetailDialog.vue'
import OrderEvidenceItemDialog from '@/components/admin/order/OrderEvidenceItemDialog.vue'
import OrderEvidenceTablePanel from '@/components/admin/order/OrderEvidenceTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { orderEvidenceApi, orderEvidenceAttachmentUrl } from '@/api/orderEvidence'
import { orderEvidenceItemTypeName } from '@/lib/orderEvidencePresentation'
import { useAuthStore } from '@/stores/auth'
import type {
  OrderEvidenceItem,
  OrderEvidenceItemUpdateInput,
  OrderEvidenceListParams,
  OrderEvidenceOrderSummary,
  OrderEvidencePackageResult,
} from '@/modules/order/orderEvidenceTypes'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const loading = ref(false)
const detailLoading = ref(false)
const locking = ref(false)
const revising = ref(false)
const exporting = ref(false)
const orders = ref<OrderEvidenceOrderSummary[]>([])
const selectedOrder = ref<OrderEvidenceOrderSummary | null>(null)
const detailResult = ref<OrderEvidencePackageResult | null>(null)
const detailOpen = ref(false)
const itemDialogOpen = ref(false)
const editingItem = ref<OrderEvidenceItem | null>(null)
const savingItemId = ref<number | string | null>(null)
const uploadingAttachmentItemId = ref<number | string | null>(null)
const deletingAttachmentID = ref<number | string | null>(null)
const confirmation = reactive<{
  open: boolean
  action: 'lock' | 'revision' | 'delete-attachment' | null
  title: string
  description: string
  confirmLabel: string
  attachmentID: number | string | null
}>({
  open: false,
  action: null,
  title: '',
  description: '',
  confirmLabel: '确定',
  attachmentID: null,
})

const filters = reactive({
  search: '',
  package_status: '',
  high_value: 'all',
  spoke_tension_qc: 'all',
})
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const queryValue = (value: unknown): string => {
  if (Array.isArray(value)) return String(value[0] || '').trim()
  return String(value || '').trim()
}

const routeOrderID = computed(() => queryValue(route.query.order_id))
const returnRoute = computed(() => {
  const disputeID = queryValue(route.query.return_dispute_id)
  if (!disputeID) return null
  const provider = queryValue(route.query.return_provider).toLowerCase() === 'paypal' ? 'paypal' : 'stripe'
  return {
    name: provider === 'paypal' ? 'PaymentPayPalDisputes' : 'PaymentStripeDisputes',
    query: {
      provider,
      dispute_id: disputeID,
    },
  }
})

const canEdit = computed(() => authStore.hasPermission('order:edit'))
const statItems = computed(() => {
  const highValue = orders.value.filter((item) => item.is_high_value).length
  const tension = orders.value.filter((item) => item.has_spoke_tension_qc).length
  const pending = orders.value.reduce((total, item) => total + Number(item.pending_evidence_items || 0), 0)
  const notGenerated = orders.value.filter((item) => !item.package_id).length
  return [
    { key: 'total', label: '当前结果', value: pagination.total, icon: FileCheck2, tone: 'gray' as const },
    { key: 'high-value', label: '当前页高价', value: highValue, icon: ShieldCheck, tone: highValue ? 'amber' as const : 'green' as const },
    { key: 'pending', label: '待处理证据', value: pending, icon: TriangleAlert, tone: pending ? 'coral' as const : 'green' as const },
    { key: 'spoke-qc', label: '当前页张力表', value: tension, icon: ListChecks, tone: tension ? 'blue' as const : 'gray' as const },
    { key: 'not-generated', label: '未生成包', value: notGenerated, icon: FileCheck2, tone: notGenerated ? 'coral' as const : 'green' as const },
  ]
})

const errorMessage = (error: unknown, fallback: string): string => {
  const responseError = (error as { response?: { data?: { error?: unknown } } })?.response?.data?.error
  return typeof responseError === 'string' && responseError.trim() ? responseError : fallback
}

const listParams = (): OrderEvidenceListParams => ({
  page: pagination.page,
  page_size: pagination.pageSize,
  ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
  ...(filters.package_status ? { package_status: filters.package_status } : {}),
  ...(filters.high_value !== 'all' ? { high_value: filters.high_value } : {}),
  ...(filters.spoke_tension_qc !== 'all' ? { spoke_tension_qc: filters.spoke_tension_qc } : {}),
})

const fetchOrders = async (): Promise<void> => {
  loading.value = true
  try {
    const result = await orderEvidenceApi.listOrders(listParams())
    orders.value = result.data
    pagination.page = result.pagination.page
    pagination.pageSize = result.pagination.page_size
    pagination.total = result.pagination.total
  } catch (error) {
    toast.error(errorMessage(error, '履约证据列表加载失败'))
  } finally {
    loading.value = false
  }
}

const fetchSelectedPackage = async (): Promise<void> => {
  const orderID = selectedOrder.value?.order_id
  if (!orderID || !selectedOrder.value?.package_id) {
    detailResult.value = null
    return
  }

  detailLoading.value = true
  try {
    detailResult.value = await orderEvidenceApi.getPackage(orderID)
  } catch (error) {
    detailResult.value = null
    toast.error(errorMessage(error, '证据包详情加载失败'))
  } finally {
    detailLoading.value = false
  }
}

const openOrder = async (order: OrderEvidenceOrderSummary): Promise<void> => {
  selectedOrder.value = order
  detailResult.value = null
  detailOpen.value = true
  await fetchSelectedPackage()
}

const openItemEditor = (item: OrderEvidenceItem): void => {
  editingItem.value = item
  itemDialogOpen.value = true
}

const attachmentUrlForEditingItem = (attachmentID: number | string): string => {
  const orderID = selectedOrder.value?.order_id
  const itemID = editingItem.value?.id
  if (!orderID || !itemID) return '#'
  return orderEvidenceAttachmentUrl(orderID, itemID, attachmentID)
}

const uploadAttachment = async (file: File): Promise<void> => {
  const orderID = selectedOrder.value?.order_id
  const itemID = editingItem.value?.id
  if (!orderID || !itemID) return

  uploadingAttachmentItemId.value = itemID
  try {
    const attachment = await orderEvidenceApi.uploadAttachment(orderID, itemID, file)
    const currentItem = editingItem.value
    if (currentItem) {
      editingItem.value = {
        ...currentItem,
        attachments: [...(currentItem.attachments || []), attachment],
      }
    }
    await fetchSelectedPackage()
    const refreshedItem = detailResult.value?.package.items?.find((item) => item.id === itemID)
    if (refreshedItem) editingItem.value = refreshedItem
    await fetchOrders()
    toast.success('证据附件已上传')
  } catch (error) {
    toast.error(errorMessage(error, '证据附件上传失败'))
  } finally {
    uploadingAttachmentItemId.value = null
  }
}

const requestDeleteAttachment = (attachmentID: number | string): void => {
  confirmation.action = 'delete-attachment'
  confirmation.attachmentID = attachmentID
  confirmation.title = '移除这个证据附件？'
  confirmation.description = '只从当前未锁定证据包移除附件引用；底层文件对象不会物理删除，便于后续审计追溯。'
  confirmation.confirmLabel = '移除附件引用'
  confirmation.open = true
}

const deleteAttachment = async (attachmentID: number | string): Promise<void> => {
  const orderID = selectedOrder.value?.order_id
  const itemID = editingItem.value?.id
  if (!orderID || !itemID) return

  deletingAttachmentID.value = attachmentID
  try {
    await orderEvidenceApi.deleteAttachment(orderID, itemID, attachmentID)
    await fetchSelectedPackage()
    const refreshedItem = detailResult.value?.package?.items?.find((item) => String(item.id) === String(itemID))
    if (refreshedItem) editingItem.value = refreshedItem
    await fetchOrders()
    toast.success('错误附件已从当前证据包移除')
  } catch (error) {
    toast.error(errorMessage(error, '证据附件移除失败'))
  } finally {
    deletingAttachmentID.value = null
  }
}

const saveItem = async (input: OrderEvidenceItemUpdateInput): Promise<void> => {
  const orderID = selectedOrder.value?.order_id
  const itemID = editingItem.value?.id
  if (!orderID || !itemID) return

  savingItemId.value = itemID
  try {
    detailResult.value = await orderEvidenceApi.updateItem(orderID, itemID, input)
    itemDialogOpen.value = false
    await fetchOrders()
    toast.success(`${orderEvidenceItemTypeName(editingItem.value?.item_type)}已保存`)
  } catch (error) {
    toast.error(errorMessage(error, '证据项保存失败'))
  } finally {
    savingItemId.value = null
  }
}

const requestLock = (): void => {
  confirmation.action = 'lock'
  confirmation.title = '锁定当前证据包？'
  confirmation.description = '锁定后当前版本不可覆盖修改；后续更正必须创建新的修订版本。'
  confirmation.confirmLabel = '锁定证据包'
  confirmation.open = true
}

const requestRevision = (): void => {
  confirmation.action = 'revision'
  confirmation.title = '创建证据包修订版？'
  confirmation.description = '旧版本将保持已锁定状态，新版本会复制当前结构化记录和附件引用。'
  confirmation.confirmLabel = '创建修订版'
  confirmation.open = true
}

const exportSnapshot = async (): Promise<void> => {
  const orderID = selectedOrder.value?.order_id
  const packageVersion = detailResult.value?.package?.package_version
  if (!orderID || !packageVersion) return

  exporting.value = true
  try {
    const blob = await orderEvidenceApi.exportSnapshot(orderID)
    const objectURL = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = objectURL
    anchor.download = `order-evidence-${String(orderID)}-v${String(packageVersion)}.json`
    document.body.appendChild(anchor)
    anchor.click()
    anchor.remove()
    URL.revokeObjectURL(objectURL)
    toast.success('锁定证据清单已导出')
  } catch (error) {
    toast.error(errorMessage(error, '证据清单导出失败'))
  } finally {
    exporting.value = false
  }
}

const executeConfirmation = async (): Promise<void> => {
  const orderID = selectedOrder.value?.order_id
  const action = confirmation.action
  const attachmentID = confirmation.attachmentID
  if (!orderID || !action) return

  confirmation.open = false
  confirmation.attachmentID = null
  if (action === 'delete-attachment') {
    if (attachmentID) await deleteAttachment(attachmentID)
    return
  }
  if (action === 'lock') {
    locking.value = true
    try {
      detailResult.value = await orderEvidenceApi.lockPackage(orderID)
      await fetchOrders()
      toast.success('证据包已锁定')
    } catch (error) {
      toast.error(errorMessage(error, '证据包锁定失败'))
    } finally {
      locking.value = false
    }
    return
  }

  revising.value = true
  try {
    detailResult.value = await orderEvidenceApi.createRevision(orderID)
    await fetchOrders()
    toast.success('证据包修订版已创建')
  } catch (error) {
    toast.error(errorMessage(error, '证据包修订版创建失败'))
  } finally {
    revising.value = false
  }
}

const applyFilters = (): void => {
  pagination.page = 1
  void fetchOrders()
}

const resetFilters = (): void => {
  Object.assign(filters, {
    search: '',
    package_status: '',
    high_value: 'all',
    spoke_tension_qc: 'all',
  })
  pagination.page = 1
  void fetchOrders()
}

const updatePage = (page: number): void => {
  pagination.page = page
  void fetchOrders()
}

const updatePageSize = (pageSize: number): void => {
  pagination.pageSize = pageSize
  pagination.page = 1
  void fetchOrders()
}

const refreshCurrent = async (): Promise<void> => {
  await fetchOrders()
  if (detailOpen.value) await fetchSelectedPackage()
}

const openOrderFromRoute = async (): Promise<void> => {
  const orderID = routeOrderID.value
  if (!orderID) {
    await fetchOrders()
    return
  }

  filters.search = orderID
  pagination.page = 1
  await fetchOrders()
  const target = orders.value.find((order) => String(order.order_id) === orderID)
  if (!target) {
    toast.error(`未找到订单 #${orderID} 的履约证据记录`)
    return
  }
  await openOrder(target)
}

const returnToDispute = async (): Promise<void> => {
  if (returnRoute.value) await router.push(returnRoute.value)
}

onMounted(() => {
  void openOrderFromRoute()
})
</script>

<style scoped>
.filter-label {
  display: block;
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: hsl(var(--muted-foreground) / 0.75);
}

.filter-select {
  height: 2.25rem;
  width: 100%;
  border: 1px dashed hsl(var(--border));
  border-radius: 0.75rem;
  background: hsl(var(--background));
  padding: 0 0.75rem;
  font-size: 0.875rem;
}
</style>
