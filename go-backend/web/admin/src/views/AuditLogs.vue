<template>
  <div class="space-y-4">
    <AdminPageHeader title="审计日志" description="追踪后台操作、资源变更和请求结果">
      <template #actions>
        <Button variant="outline" @click="refreshLogs">
          <RefreshCw class="size-4" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <AdminFilterPanel>
      <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-4" @submit.prevent="applyFilters">
        <label class="space-y-1 block xl:col-span-2">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">KEYWORD / 关键词</span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
            <Input v-model="filters.keyword" class="h-9 pl-9" placeholder="用户、操作、资源、路径或错误信息" />
          </div>
        </label>
        <FilterSelect v-model="filters.action" label="操作" :options="actionFilterOptions" />
        <FilterSelect v-model="filters.resource" label="资源" :options="resourceFilterOptions" />
        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">USER ID / 用户 ID</span>
          <Input v-model="filters.user_id" type="number" min="1" class="h-9" placeholder="全部用户" />
        </label>
        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">IP ADDRESS / IP 地址</span>
          <Input v-model="filters.ip_address" class="h-9 font-mono" placeholder="全部地址" />
        </label>
        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">START DATE / 开始日期</span>
          <Input v-model="filters.start_date" type="date" class="h-9" />
        </label>
        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">END DATE / 结束日期</span>
          <Input v-model="filters.end_date" type="date" class="h-9" />
        </label>
        <label class="space-y-1 block xl:col-span-4">
          <span class="block text-[10px] font-black uppercase tracking-widest text-transparent select-none">ACTION / 操作</span>
          <div class="flex items-center gap-2">
            <Button type="submit" class="h-9 rounded-full px-4 font-black text-xs uppercase tracking-wider">
              <Search class="size-3.5" />
              搜索
            </Button>
            <Button type="button" variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="resetFilters">
              <RotateCcw class="size-3.5" />
              重置
            </Button>
          </div>
        </label>
      </form>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[1440px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-16">ID</TableHead>
            <TableHead class="w-36">用户</TableHead>
            <TableHead class="w-24">操作</TableHead>
            <TableHead class="w-28">资源</TableHead>
            <TableHead class="w-24">资源 ID</TableHead>
            <TableHead class="w-20">方法</TableHead>
            <TableHead>路径</TableHead>
            <TableHead class="w-36">IP 地址</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-24 text-right">耗时</TableHead>
            <TableHead class="w-44">时间</TableHead>
            <TableHead class="w-16 text-right">详情</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="logs.length === 0" :colspan="12">
            <div class="flex flex-col items-center text-muted-foreground">
              <ScrollText class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无审计日志</span>
            </div>
          </TableEmpty>
          <TableRow v-for="log in logs" :key="log.id">
            <TableCell class="font-mono text-xs text-muted-foreground">{{ log.id }}</TableCell>
            <TableCell>
              <span class="block truncate font-medium">{{ log.username || '-' }}</span>
              <span class="block font-mono text-[11px] text-muted-foreground">ID {{ log.user_id || '-' }}</span>
            </TableCell>
            <TableCell><AdminStatusBadge :tone="actionTone(log.action)">{{ actionName(log.action) }}</AdminStatusBadge></TableCell>
            <TableCell>{{ resourceName(log.resource) }}</TableCell>
            <TableCell class="font-mono text-xs">{{ log.resource_id || '-' }}</TableCell>
            <TableCell><AdminStatusBadge :tone="methodTone(log.method)">{{ log.method || '-' }}</AdminStatusBadge></TableCell>
            <TableCell class="max-w-96 truncate font-mono text-xs text-muted-foreground">{{ log.path || '-' }}</TableCell>
            <TableCell class="font-mono text-xs">{{ log.ip_address || '-' }}</TableCell>
            <TableCell><AdminStatusBadge :tone="log.status === 'success' ? 'green' : 'coral'">{{ log.status === 'success' ? '成功' : '失败' }}</AdminStatusBadge></TableCell>
            <TableCell class="text-right tabular-nums" :class="durationClass(log.duration)">{{ log.duration || 0 }} ms</TableCell>
            <TableCell class="text-xs text-muted-foreground">{{ formatDate(log.created_at) }}</TableCell>
            <TableCell class="text-right">
              <Button variant="ghost" size="icon" :aria-label="`查看日志 ${log.id}`" @click="viewDetail(log)">
                <Eye class="size-4" />
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

    <Dialog v-model:open="detailDialogVisible">
      <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
        <DialogHeader>
          <DialogTitle>日志详情</DialogTitle>
          <DialogDescription v-if="currentLog">审计日志 #{{ currentLog.id }}</DialogDescription>
        </DialogHeader>

        <div v-if="detailLoading" class="flex h-52 items-center justify-center">
          <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载日志详情" />
        </div>
        <div v-else-if="currentLog" class="space-y-6">
          <dl class="grid overflow-hidden rounded-lg border sm:grid-cols-2">
            <DetailItem label="用户">{{ currentLog.username || '-' }}（ID {{ currentLog.user_id || '-' }}）</DetailItem>
            <DetailItem label="状态"><AdminStatusBadge :tone="currentLog.status === 'success' ? 'green' : 'coral'">{{ currentLog.status === 'success' ? '成功' : '失败' }}</AdminStatusBadge></DetailItem>
            <DetailItem label="操作"><AdminStatusBadge :tone="actionTone(currentLog.action)">{{ actionName(currentLog.action) }}</AdminStatusBadge></DetailItem>
            <DetailItem label="资源">{{ resourceName(currentLog.resource) }} / {{ currentLog.resource_id || '-' }}</DetailItem>
            <DetailItem label="请求方法"><AdminStatusBadge :tone="methodTone(currentLog.method)">{{ currentLog.method || '-' }}</AdminStatusBadge></DetailItem>
            <DetailItem label="耗时">{{ currentLog.duration || 0 }} ms</DetailItem>
            <DetailItem label="IP 地址"><span class="font-mono text-xs">{{ currentLog.ip_address || '-' }}</span></DetailItem>
            <DetailItem label="时间">{{ formatDate(currentLog.created_at) }}</DetailItem>
            <DetailItem label="请求路径" class="sm:col-span-2"><span class="break-all font-mono text-xs">{{ currentLog.path || '-' }}</span></DetailItem>
            <DetailItem label="User Agent" class="sm:col-span-2"><span class="break-all text-xs">{{ currentLog.user_agent || '-' }}</span></DetailItem>
          </dl>

          <Alert v-if="currentLog.error_message" variant="destructive">
            <CircleAlert class="size-4" />
            <AlertTitle>错误信息</AlertTitle>
            <AlertDescription class="whitespace-pre-wrap break-words">{{ currentLog.error_message }}</AlertDescription>
          </Alert>

          <JsonSection v-if="currentLog.changes" title="变更内容" :value="currentLog.changes" />
          <div v-if="currentLog.old_value || currentLog.new_value" class="grid gap-4 md:grid-cols-2">
            <JsonSection v-if="currentLog.old_value" title="变更前" :value="currentLog.old_value" />
            <JsonSection v-if="currentLog.new_value" title="变更后" :value="currentLog.new_value" />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { CircleAlert, Eye, LoaderCircle, RefreshCw, RotateCcw, ScrollText, Search, ShieldCheck, ShieldX } from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import axios from '@/utils/axios'

const FilterSelect = defineComponent({
  props: {
    modelValue: { type: String, required: true },
    label: { type: String, required: true },
    options: { type: Array, required: true }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('label', { class: 'space-y-1 block' }, [
      h('span', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block' }, props.label),
      h(Select, {
        modelValue: props.modelValue,
        'onUpdate:modelValue': (value) => emit('update:modelValue', value)
      }, {
        default: () => [
          h(SelectTrigger, { class: 'h-9 w-full' }, { default: () => h(SelectValue) }),
          h(SelectContent, {}, {
            default: () => props.options.map((option) => h(SelectItem, { value: option.value }, { default: () => option.label }))
          })
        ]
      })
    ])
  }
})

const DetailItem = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: [String, Number], default: '' }
  },
  setup(props, { slots }) {
    return () => h('div', { class: 'space-y-1' }, [
      h('dt', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block' }, props.label),
      h('dd', { class: 'text-xs font-bold' }, slots.default ? slots.default() : (props.value || '-'))
    ])
  }
})

const JsonSection = defineComponent({
  props: { title: { type: String, required: true }, value: { type: String, required: true } },
  setup(props) {
    return () => h('section', { class: 'min-w-0 space-y-2' }, [
      h('h3', { class: 'text-sm font-black tracking-tighter italic uppercase text-foreground' }, props.title),
      h('pre', { class: 'max-h-80 overflow-auto rounded-lg border border-dashed bg-muted/40 p-3 font-mono text-xs leading-5 whitespace-pre-wrap break-words' }, formatJSON(props.value))
    ])
  }
})

const loading = ref(false)
const logs = ref([])
const stats = ref({})
const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const currentLog = ref(null)
const filters = reactive({ keyword: '', action: 'all', resource: 'all', user_id: '', ip_address: '', start_date: '', end_date: '' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const actionFilterOptions = [
  { label: '全部操作', value: 'all' }, { label: '创建', value: 'create' }, { label: '更新', value: 'update' },
  { label: '删除', value: 'delete' }, { label: '查看', value: 'view' }
]
const resourceFilterOptions = [
  { label: '全部资源', value: 'all' }, { label: '用户', value: 'user' }, { label: '商品', value: 'product' },
  { label: '订单', value: 'order' }, { label: '文章', value: 'post' }, { label: '工单', value: 'ticket' },
  { label: 'FAQ', value: 'faq' }, { label: '图库', value: 'gallery' }, { label: '订阅', value: 'subscription' },
  { label: '营销', value: 'marketing' }, { label: '设置', value: 'setting' }
]

const statItems = computed(() => [
  { key: 'total', label: '总日志数', value: stats.value.total_count || 0, icon: ScrollText, tone: 'gray' },
  { key: 'today', label: '今日操作', value: stats.value.today_count || 0, icon: RefreshCw, tone: 'blue' },
  { key: 'success', label: '成功操作', value: stats.value.success_count || 0, icon: ShieldCheck, tone: 'green' },
  { key: 'failed', label: '失败操作', value: stats.value.failed_count || 0, icon: ShieldX, tone: 'coral' }
])

const actionName = (action) => ({ create: '创建', update: '更新', delete: '删除', view: '查看' })[action] || action || '-'
const actionTone = (action) => ({ create: 'green', update: 'amber', delete: 'coral', view: 'gray' })[action] || 'gray'
const resourceName = (resource) => ({
  user: '用户', product: '商品', order: '订单', post: '文章', ticket: '工单', faq: 'FAQ', gallery: '图库',
  subscription: '订阅', marketing: '营销', setting: '设置'
})[resource] || resource || '-'
const methodTone = (method) => ({ GET: 'gray', POST: 'green', PUT: 'amber', PATCH: 'blue', DELETE: 'coral' })[method] || 'gray'
const durationClass = (duration) => Number(duration || 0) >= 1000 ? 'font-medium text-destructive' : Number(duration || 0) >= 300 ? 'text-amber-700' : ''
const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
const formatJSON = (value) => {
  if (typeof value !== 'string') return JSON.stringify(value, null, 2)
  try { return JSON.stringify(JSON.parse(value), null, 2) } catch { return value }
}
const dateParams = () => ({
  ...(filters.start_date ? { start_date: filters.start_date } : {}),
  ...(filters.end_date ? { end_date: filters.end_date } : {})
})

const fetchLogs = async () => {
  loading.value = true
  try {
    const keyword = filters.keyword.trim()
    const endpoint = keyword ? '/api/admin/logs/search' : '/api/admin/logs'
    const params = keyword
      ? { keyword, page: pagination.page, page_size: pagination.pageSize }
      : {
          page: pagination.page,
          page_size: pagination.pageSize,
          ...(filters.action !== 'all' ? { action: filters.action } : {}),
          ...(filters.resource !== 'all' ? { resource: filters.resource } : {}),
          ...(filters.user_id ? { user_id: filters.user_id } : {}),
          ...(filters.ip_address.trim() ? { ip_address: filters.ip_address.trim() } : {}),
          ...dateParams()
        }
    const response = await axios.get(endpoint, { params })
    logs.value = response.data.logs || []
    pagination.total = response.data.total || 0
  } catch (error) {
    console.error('Failed to fetch audit logs:', error)
  } finally {
    loading.value = false
  }
}
const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/logs/stats', { params: dateParams() })
    stats.value = response.data || {}
  } catch (error) {
    console.error('Failed to fetch audit stats:', error)
  }
}
const refreshLogs = () => Promise.all([fetchLogs(), fetchStats()])
const applyFilters = () => { pagination.page = 1; refreshLogs() }
const resetFilters = () => {
  Object.assign(filters, { keyword: '', action: 'all', resource: 'all', user_id: '', ip_address: '', start_date: '', end_date: '' })
  pagination.page = 1
  refreshLogs()
}
const updatePage = (page) => { pagination.page = page; fetchLogs() }
const updatePageSize = (pageSize) => { pagination.pageSize = pageSize; pagination.page = 1; fetchLogs() }
const viewDetail = async (log) => {
  currentLog.value = log
  detailDialogVisible.value = true
  detailLoading.value = true
  try {
    const response = await axios.get(`/api/admin/logs/${log.id}`)
    currentLog.value = response.data.log || log
  } catch (error) {
    console.error('Failed to fetch audit log detail:', error)
  } finally {
    detailLoading.value = false
  }
}

onMounted(refreshLogs)
</script>
