<template>
  <div class="flex h-full min-h-0 flex-col gap-4 overflow-y-auto">
    <AdminPageHeader title="发送记录" description="查看事务邮件的发送结果和重试状态，不保存邮件正文。">
      <template #actions>
        <Button
          variant="outline"
          size="icon"
          aria-label="刷新发送记录"
          title="刷新发送记录"
          :disabled="loading"
          @click="loadEmailDeliveryRecords"
        >
          <RefreshCw class="size-4" :class="loading ? 'animate-spin' : ''" />
        </Button>
      </template>
    </AdminPageHeader>

    <AdminFilterPanel>
      <form class="grid grid-cols-1 gap-3 md:grid-cols-[minmax(240px,1fr)_180px_auto]" @submit.prevent="applyEmailDeliveryRecordFilters">
        <label class="block space-y-1">
          <span class="filter-label">SEARCH / 搜索</span>
          <Input v-model="filters.search" placeholder="主题、订单号或模板编码" />
        </label>
        <label class="block space-y-1">
          <span class="filter-label">STATUS / 状态</span>
          <select v-model="filters.status" class="filter-select">
            <option value="">全部状态</option>
            <option value="sending">发送中</option>
            <option value="sent">已发送</option>
            <option value="failed">失败</option>
            <option value="unknown">未知</option>
          </select>
        </label>
        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9 px-4 text-xs font-black uppercase tracking-wider">查询</Button>
          <Button type="button" variant="outline" class="h-9 px-4 text-xs font-black uppercase tracking-wider" @click="resetEmailDeliveryRecordFilters">重置</Button>
        </div>
      </form>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading" class="min-h-0 flex-1">
      <table class="w-full min-w-[1080px] border-collapse text-left">
        <thead class="border-b border-dashed border-border/70 bg-muted/20">
          <tr>
            <th class="table-heading">状态</th>
            <th class="table-heading">模板</th>
            <th class="table-heading">收件人</th>
            <th class="table-heading">邮件标题</th>
            <th class="table-heading">关联单号</th>
            <th class="table-heading">通道</th>
            <th class="table-heading">尝试次数</th>
            <th class="table-heading">最后尝试</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="record in records"
            :key="record.id"
            class="border-b border-dashed border-border/60 last:border-0 hover:bg-muted/20"
          >
            <td class="table-cell">
              <span class="status-mark" :class="getEmailDeliveryRecordStatusClass(record.status)">{{ getEmailDeliveryRecordStatusName(record.status) }}</span>
              <p v-if="record.last_error" class="mt-1 max-w-44 truncate text-[10px] text-rose-600" :title="record.last_error">
                {{ record.last_error }}
              </p>
            </td>
            <td class="table-cell">
              <p class="font-semibold text-foreground">{{ getEmailDeliveryRecordTemplateName(record.template_code) }}</p>
              <p class="mt-0.5 text-[10px] text-muted-foreground">v{{ record.template_version }} · {{ getEmailDeliveryRecordLocaleName(record.locale) }}</p>
            </td>
            <td class="table-cell whitespace-nowrap text-xs text-muted-foreground">{{ record.recipient_email }}</td>
            <td class="table-cell max-w-[300px]">
              <p class="truncate font-semibold text-foreground" :title="record.subject || ''">{{ record.subject || '-' }}</p>
            </td>
            <td class="table-cell">
              <span v-if="record.reference_number" class="text-xs font-bold">{{ record.reference_number }}</span>
              <span v-else class="text-muted-foreground">-</span>
            </td>
            <td class="table-cell text-xs text-muted-foreground">{{ record.provider_code || '-' }}</td>
            <td class="table-cell text-xs text-muted-foreground">{{ record.attempt_count || 0 }}</td>
            <td class="table-cell whitespace-nowrap text-xs text-muted-foreground">{{ formatEmailDeliveryRecordDate(record.last_attempt_at) }}</td>
          </tr>
          <tr v-if="!loading && records.length === 0">
            <td colspan="8" class="px-4 py-14 text-center text-sm text-muted-foreground">暂无发送记录</td>
          </tr>
        </tbody>
      </table>

      <template #footer>
        <AdminPagination
          :page="pagination.page"
          :page-size="pagination.page_size"
          :total="pagination.total"
          @update:page="updateEmailDeliveryRecordPage"
          @update:page-size="updateEmailDeliveryRecordPageSize"
        />
      </template>
    </AdminTablePanel>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { RefreshCw } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import emailDeliveryRecordsApi, {
  type EmailDeliveryRecord,
  type EmailDeliveryRecordStatus,
} from '@/api/emailDeliveryRecordsApi'

const records = ref<EmailDeliveryRecord[]>([])
const loading = ref(false)
const filters = reactive({ search: '', status: '' })
const pagination = reactive({ page: 1, page_size: 20, total: 0, total_pages: 0 })

const templateNames: Record<string, string> = {
  order_confirmation: '订单确认',
  order_payment_expired: '付款超时',
  order_cancelled: '订单取消',
  order_shipping_notification: '发货通知',
  order_delivered: '订单送达',
  order_completed: '订单完成',
  order_refunded: '退款完成',
  after_sales_requested: '售后申请',
  after_sales_approved: '售后批准',
  after_sales_awaiting_return: '售后寄回说明',
  after_sales_return_in_transit: '售后运输中',
  after_sales_received: '售后已收到',
  after_sales_resolving: '售后处理中',
  after_sales_completed: '售后完成',
  after_sales_rejected: '售后拒绝',
}

const statusNames: Record<EmailDeliveryRecordStatus, string> = {
  sending: '发送中',
  sent: '已发送',
  failed: '失败',
  unknown: '未知',
}

const getEmailDeliveryRecordTemplateName = (code: string): string => templateNames[code] || code || '-'
const getEmailDeliveryRecordLocaleName = (locale: string): string => ({
  en: '英语',
  zh_cn: '中文',
  ja: '日语',
  ko: '韩语',
  fr: '法语',
  de: '德语',
  es: '西班牙语',
}[locale] || locale || '-')
const getEmailDeliveryRecordStatusName = (status: EmailDeliveryRecordStatus): string => statusNames[status] || status
const getEmailDeliveryRecordStatusClass = (status: EmailDeliveryRecordStatus): string => ({
  sending: 'status-pending',
  sent: 'status-success',
  failed: 'status-error',
  unknown: 'status-warning',
}[status] || 'status-pending')
const formatEmailDeliveryRecordDate = (value?: string): string => value ? new Date(value).toLocaleString('zh-CN') : '-'

const loadEmailDeliveryRecords = async (): Promise<void> => {
  loading.value = true
  try {
    const result = await emailDeliveryRecordsApi.listEmailDeliveryRecords({
      page: pagination.page,
      page_size: pagination.page_size,
      ...(filters.status ? { status: filters.status } : {}),
      ...(filters.search.trim() ? { search: filters.search.trim() } : {}),
    })
    records.value = result.records
    Object.assign(pagination, result.pagination)
  } catch (error) {
    toast.error(error instanceof Error ? error.message : '发送记录加载失败')
  } finally {
    loading.value = false
  }
}

const applyEmailDeliveryRecordFilters = (): void => {
  pagination.page = 1
  void loadEmailDeliveryRecords()
}

const resetEmailDeliveryRecordFilters = (): void => {
  filters.search = ''
  filters.status = ''
  pagination.page = 1
  void loadEmailDeliveryRecords()
}

const updateEmailDeliveryRecordPage = (page: number): void => {
  pagination.page = page
  void loadEmailDeliveryRecords()
}

const updateEmailDeliveryRecordPageSize = (pageSize: number): void => {
  pagination.page_size = pageSize
  pagination.page = 1
  void loadEmailDeliveryRecords()
}

onMounted(() => {
  void loadEmailDeliveryRecords()
})
</script>

<style scoped>
.filter-label {
  display: block;
  color: hsl(var(--muted-foreground) / 0.7);
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.filter-select {
  height: 2.25rem;
  width: 100%;
  border: 0;
  border-radius: 0.75rem;
  background: hsl(var(--muted) / 0.5);
  padding: 0 0.75rem;
  color: hsl(var(--foreground));
  font-size: 0.75rem;
  font-weight: 700;
  outline: none;
}

.table-heading {
  padding: 0.75rem 1rem;
  color: hsl(var(--muted-foreground) / 0.7);
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  white-space: nowrap;
}

.table-cell {
  padding: 0.8rem 1rem;
  vertical-align: top;
  font-size: 0.875rem;
}

.status-mark {
  display: inline-flex;
  border-radius: 999px;
  border: 1px solid transparent;
  padding: 0.125rem 0.5rem;
  font-size: 10px;
  font-weight: 800;
  white-space: nowrap;
}

.status-pending {
  border-color: rgb(245 158 11 / 0.3);
  background: rgb(245 158 11 / 0.1);
  color: rgb(180 83 9);
}

.status-success {
  border-color: rgb(16 185 129 / 0.3);
  background: rgb(16 185 129 / 0.1);
  color: rgb(4 120 87);
}

.status-error {
  border-color: rgb(244 63 94 / 0.3);
  background: rgb(244 63 94 / 0.1);
  color: rgb(190 24 93);
}

.status-warning {
  border-color: rgb(249 115 22 / 0.3);
  background: rgb(249 115 22 / 0.1);
  color: rgb(194 65 12);
}
</style>
