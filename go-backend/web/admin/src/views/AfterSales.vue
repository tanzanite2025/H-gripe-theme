<template>
  <div class="space-y-4">
    <AdminPageHeader title="退换货管理" description="独立售后单工作台，订单状态与售后状态分开管理">
      <template #actions>
        <Button
          variant="outline"
          size="icon"
          aria-label="刷新退换货"
          title="刷新退换货"
          :disabled="loading"
          @click="fetchCases"
        >
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <AdminFilterPanel>
      <form class="grid grid-cols-1 gap-3 md:grid-cols-[minmax(220px,1fr)_180px_180px_auto]" @submit.prevent="applyFilters">
        <label class="block space-y-1">
          <span class="filter-label">SEARCH / 搜索</span>
          <Input v-model="filters.search" placeholder="订单号或申请原因" />
        </label>
        <label class="block space-y-1">
          <span class="filter-label">STATUS / 状态</span>
          <select v-model="filters.status" class="filter-select">
            <option v-for="option in statusOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </label>
        <label class="block space-y-1">
          <span class="filter-label">TYPE / 类型</span>
          <select v-model="filters.type" class="filter-select">
            <option v-for="option in typeOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
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

    <AdminTablePanel :loading="loading">
      <table class="w-full min-w-[980px] border-collapse text-left">
        <thead class="border-b border-dashed border-border/70 bg-muted/20">
          <tr>
            <th class="table-heading">售后单</th>
            <th class="table-heading">订单</th>
            <th class="table-heading">类型</th>
            <th class="table-heading">申请原因</th>
            <th class="table-heading">商品</th>
            <th class="table-heading">状态</th>
            <th class="table-heading">创建时间</th>
            <th class="table-heading text-right">处理</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="record in records" :key="String(record.id)" class="border-b border-dashed border-border/60 last:border-0 hover:bg-muted/20">
            <td class="table-cell">
              <span class="font-mono font-bold text-foreground">#{{ record.id }}</span>
            </td>
            <td class="table-cell">
              <RouterLink
                :to="{ name: 'OrdersList', query: record.order_number ? { search: record.order_number } : undefined }"
                class="font-mono text-xs font-bold text-primary hover:underline"
              >
                {{ record.order_number || `订单 #${record.order_id || '-'}` }}
              </RouterLink>
            </td>
            <td class="table-cell">
              <span class="type-mark">{{ getTypeName(record.type) }}</span>
            </td>
            <td class="table-cell max-w-[260px]">
              <p class="truncate font-semibold text-foreground" :title="record.reason || ''">{{ record.reason || '-' }}</p>
              <p v-if="record.description" class="mt-0.5 truncate text-xs text-muted-foreground" :title="record.description">{{ record.description }}</p>
            </td>
            <td class="table-cell">
              <span class="font-mono text-xs text-muted-foreground">
                {{ itemSummary(record) }}
              </span>
            </td>
            <td class="table-cell">
              <span class="status-mark" :class="statusClass(record.status)">
                {{ getStatusName(record.status) }}
              </span>
            </td>
            <td class="table-cell whitespace-nowrap text-xs text-muted-foreground">
              {{ formatDate(record.created_at) }}
            </td>
            <td class="table-cell">
              <div class="flex items-center justify-end">
                <Button
                  size="icon"
                  variant="outline"
                  class="size-8"
                  :aria-label="`查看售后单 #${record.id} 详情`"
                  :title="nextStatuses(record.status).length ? '查看详情并处理' : '查看详情'"
                  @click="openCaseDetail(record)"
                >
                  <Eye class="size-3.5" />
                </Button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && records.length === 0">
            <td colspan="8" class="px-4 py-14 text-center text-sm text-muted-foreground">
              暂无符合条件的售后单
            </td>
          </tr>
        </tbody>
      </table>

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

    <AfterSalesCaseDetailDialog
      v-model:open="detailDialogVisible"
      :record="selectedCase"
      :loading="detailLoading"
      :submitting="submittingCaseID === selectedCase?.id"
      :refund-submitting="refundSubmittingCaseID === selectedCase?.id"
      @submit="updateStatus"
      @save-refund-review="saveRefundReview"
      @decide-refund-review="decideRefundReview"
      @create-pending-refund="createPendingRefund"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Check, Eye, RefreshCw } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AfterSalesCaseDetailDialog from '@/components/admin/order/AfterSalesCaseDetailDialog.vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { afterSalesApi } from '@/api/afterSales'
import type { AfterSalesCase } from '@/api/afterSales'
import {
  afterSalesStatusFilterOptions as statusOptions,
  afterSalesStatusClass as statusClass,
  afterSalesTypeFilterOptions as typeOptions,
  getAfterSalesNextStatuses as nextStatuses,
  getAfterSalesStatusName as getStatusName,
  getAfterSalesTypeName as getTypeName,
} from '@/lib/afterSalesPresentation'

const loading = ref(false)
const route = useRoute()
const records = ref<AfterSalesCase[]>([])
const submittingCaseID = ref<number | string | null>(null)
const refundSubmittingCaseID = ref<number | string | null>(null)
const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const selectedCase = ref<AfterSalesCase | null>(null)
const queryString = (value: unknown): string => {
  const firstValue = Array.isArray(value) ? value[0] : value
  return typeof firstValue === 'string' ? firstValue : ''
}
const filters = reactive({ search: queryString(route.query.search), status: 'all', type: 'all' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const itemSummary = (record: AfterSalesCase): string => {
  const items = record.items || []
  const quantity = items.reduce((total, item) => total + Number(item.quantity || 0), 0)
  return `${items.length} 个商品 / ${quantity} 件`
}
const formatDate = (value?: string | null): string => value ? new Date(value).toLocaleString('zh-CN') : '-'
const statItems = computed(() => [
  { key: 'total', label: '筛选结果', value: pagination.total, icon: RefreshCw, tone: 'blue' },
  { key: 'reviewing', label: '本页待处理', value: records.value.filter((item) => ['requested', 'reviewing'].includes(item.status || '')).length, icon: Check, tone: 'amber' },
  { key: 'exception', label: '本页异常', value: records.value.filter((item) => item.status === 'exception').length, icon: RefreshCw, tone: 'coral' },
])

const fetchCases = async (): Promise<void> => {
  loading.value = true
  try {
    const result = await afterSalesApi.list({
      page: pagination.page,
      page_size: pagination.pageSize,
      status: filters.status !== 'all' ? filters.status : undefined,
      type: filters.type !== 'all' ? filters.type : undefined,
      search: filters.search.trim() || undefined,
    })
    records.value = result.data
    pagination.page = result.pagination.page
    pagination.pageSize = result.pagination.page_size
    pagination.total = result.pagination.total
  } catch (error) {
    console.error('Failed to fetch after-sales cases:', error)
    toast.error('退换货列表加载失败')
  } finally {
    loading.value = false
  }
}

const applyFilters = (): void => {
  pagination.page = 1
  void fetchCases()
}

const resetFilters = (): void => {
  filters.search = ''
  filters.status = 'all'
  filters.type = 'all'
  pagination.page = 1
  void fetchCases()
}

const updatePage = (page: number): void => {
  pagination.page = page
  void fetchCases()
}

const updatePageSize = (pageSize: number): void => {
  pagination.pageSize = pageSize
  pagination.page = 1
  void fetchCases()
}

const openCaseDetail = async (record: AfterSalesCase): Promise<void> => {
  selectedCase.value = record
  detailDialogVisible.value = true
  detailLoading.value = true
  try {
    selectedCase.value = await afterSalesApi.get(record.id)
  } catch (error) {
    console.error('Failed to fetch after-sales case:', error)
    toast.error('售后单详情加载失败')
  } finally {
    detailLoading.value = false
  }
}

const updateStatus = async (nextStatus: string, resolution: string): Promise<void> => {
  const record = selectedCase.value
  if (!record || !nextStatus || nextStatus === record.status) return

  submittingCaseID.value = record.id
  try {
    const updated = await afterSalesApi.updateStatus(record.id, nextStatus, resolution)
    const index = records.value.findIndex((item) => item.id === record.id)
    if (index >= 0) records.value[index] = updated
    selectedCase.value = updated
    toast.success(`售后单 #${record.id} 已更新`)
  } catch (error) {
    console.error('Failed to update after-sales status:', error)
    toast.error('售后状态更新失败')
  } finally {
    submittingCaseID.value = null
  }
}

const saveRefundReview = async (
  proposedAmount: number,
  currency: string,
  requestNotes: string,
): Promise<void> => {
  const record = selectedCase.value
  if (!record) return

  refundSubmittingCaseID.value = record.id
  try {
    const review = await afterSalesApi.saveRefundReview(record.id, {
      proposed_amount: proposedAmount,
      currency,
      request_notes: requestNotes,
    })
    selectedCase.value = { ...record, refund_review: review }
    toast.success(`售后单 #${record.id} 的退款审批草稿已保存`)
  } catch (error) {
    console.error('Failed to save after-sales refund review:', error)
    toast.error('退款审批草稿保存失败')
  } finally {
    refundSubmittingCaseID.value = null
  }
}

const decideRefundReview = async (
  status: 'approved' | 'rejected' | 'cancelled',
  decisionNotes: string,
): Promise<void> => {
  const record = selectedCase.value
  if (!record) return

  refundSubmittingCaseID.value = record.id
  try {
    const review = await afterSalesApi.decideRefundReview(record.id, status, decisionNotes)
    selectedCase.value = { ...record, refund_review: review }
    toast.success(`售后单 #${record.id} 的退款审批已更新`)
  } catch (error) {
    console.error('Failed to decide after-sales refund review:', error)
    toast.error('退款审批操作失败')
  } finally {
    refundSubmittingCaseID.value = null
  }
}

const createPendingRefund = async (): Promise<void> => {
  const record = selectedCase.value
  if (!record) return

  refundSubmittingCaseID.value = record.id
  try {
    const result = await afterSalesApi.createPendingRefund(record.id)
    selectedCase.value = { ...record, refund_review: result.refund_review }
    toast.success(`已生成待处理退款 #${result.refund.id}，尚未调用支付渠道`)
  } catch (error) {
    console.error('Failed to create after-sales pending refund:', error)
    toast.error('待处理退款创建失败')
  } finally {
    refundSubmittingCaseID.value = null
  }
}

onMounted(() => {
  void fetchCases()
})

watch(
  () => route.query.search,
  (value) => {
    const search = queryString(value)
    if (filters.search === search) return
    filters.search = search
    pagination.page = 1
    void fetchCases()
  },
)
</script>

<style scoped>
.filter-label {
  display: block;
  color: hsl(var(--muted-foreground) / 0.7);
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.filter-select,
.status-select {
  height: 2.25rem;
  width: 100%;
  border: 1px dashed hsl(var(--border) / 0.8);
  border-radius: 0.375rem;
  background: hsl(var(--background));
  padding: 0 0.75rem;
  color: hsl(var(--foreground));
  font-size: 0.875rem;
}

.status-select {
  width: 7.75rem;
  height: 2rem;
  padding-inline: 0.5rem;
  font-size: 0.75rem;
}

.table-heading {
  padding: 0.75rem 1rem;
  color: hsl(var(--muted-foreground) / 0.65);
  font-size: 10px;
  font-weight: 900;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  white-space: nowrap;
}

.table-cell {
  padding: 0.9rem 1rem;
  vertical-align: middle;
}

.type-mark,
.status-mark {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.25rem 0.55rem;
  font-size: 0.6875rem;
  font-weight: 800;
  white-space: nowrap;
}

.type-mark {
  border: 1px dashed hsl(var(--border));
  color: hsl(var(--muted-foreground));
}

.status-gray {
  background: rgb(148 163 184 / 0.12);
  color: rgb(71 85 105);
}

.status-green {
  background: rgb(16 185 129 / 0.12);
  color: rgb(4 120 87);
}

.status-amber {
  background: rgb(245 158 11 / 0.14);
  color: rgb(180 83 9);
}

.status-blue {
  background: rgb(59 130 246 / 0.12);
  color: rgb(29 78 216);
}

.status-coral {
  background: rgb(244 63 94 / 0.12);
  color: rgb(190 24 93);
}
</style>
