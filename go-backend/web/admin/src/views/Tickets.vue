<template>
  <div class="space-y-4">
    <AdminPageHeader title="工单管理" description="处理客户请求、分配负责人并跟进消息记录" />

    <AdminStatsGrid :items="statItems" />

    <AdminFilterPanel>
      <div class="flex flex-col gap-3 sm:flex-row sm:items-end">
        <FilterSelect v-model="filters.status" label="状态" :options="statusFilterOptions" />
        <FilterSelect v-model="filters.priority" label="优先级" :options="priorityFilterOptions" />
        <Button type="button" variant="outline" class="h-9" @click="resetFilters">
          <RotateCcw class="size-4" />
          重置
        </Button>
        <Button type="button" variant="ghost" class="h-9" @click="refreshTickets">
          <RefreshCw class="size-4" />
          刷新
        </Button>
      </div>
    </AdminFilterPanel>

    <AdminTablePanel :loading="loading">
      <Table class="min-w-[1080px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-44">工单号</TableHead>
            <TableHead>标题</TableHead>
            <TableHead class="w-28">分类</TableHead>
            <TableHead class="w-24">状态</TableHead>
            <TableHead class="w-24">优先级</TableHead>
            <TableHead class="w-40">用户</TableHead>
            <TableHead class="w-32">负责人</TableHead>
            <TableHead class="w-44">创建时间</TableHead>
            <TableHead class="w-16 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="tickets.length === 0" :colspan="9">
            <div class="flex flex-col items-center text-muted-foreground">
              <MessagesSquare class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无工单</span>
            </div>
          </TableEmpty>

          <TableRow v-for="ticket in tickets" :key="ticket.id">
            <TableCell class="font-mono text-xs font-medium">{{ ticket.ticket_number }}</TableCell>
            <TableCell class="max-w-80 truncate font-medium">{{ ticket.subject }}</TableCell>
            <TableCell>{{ categoryName(ticket.category) }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="statusTone(ticket.status)">{{ statusName(ticket.status) }}</AdminStatusBadge>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="priorityTone(ticket.priority)">{{ priorityName(ticket.priority) }}</AdminStatusBadge>
            </TableCell>
            <TableCell class="max-w-40 truncate">{{ customerName(ticket) }}</TableCell>
            <TableCell>{{ assigneeName(ticket.assigned_to) }}</TableCell>
            <TableCell class="text-xs text-muted-foreground">{{ formatDate(ticket.created_at) }}</TableCell>
            <TableCell class="text-right">
              <DropdownMenu>
                <DropdownMenuTrigger as-child>
                  <Button variant="ghost" size="icon" :aria-label="`管理工单 ${ticket.ticket_number}`">
                    <MoreHorizontal class="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end" class="w-40">
                  <DropdownMenuItem @select="viewTicket(ticket)">
                    <Eye class="size-4" />
                    查看详情
                  </DropdownMenuItem>
                  <DropdownMenuItem v-if="hasPermission('ticket:edit')" @select="showAssignDialog(ticket)">
                    <UserRoundCog class="size-4" />
                    分配工单
                  </DropdownMenuItem>
                  <DropdownMenuSeparator v-if="hasPermission('ticket:delete')" />
                  <DropdownMenuItem
                    v-if="hasPermission('ticket:delete')"
                    class="text-destructive focus:text-destructive"
                    @select="requestDelete(ticket)"
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
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </template>
    </AdminTablePanel>

    <Dialog v-model:open="detailDialogVisible">
      <DialogContent class="max-h-[92vh] overflow-y-auto p-0 sm:max-w-6xl" @open-auto-focus.prevent>
        <DialogHeader class="border-b px-5 py-4 pr-12">
          <DialogTitle>{{ currentTicket?.ticket_number || '工单详情' }}</DialogTitle>
          <DialogDescription>{{ currentTicket?.subject || '查看工单信息和消息记录' }}</DialogDescription>
        </DialogHeader>

        <div class="relative min-h-80">
          <div v-if="detailLoading" class="absolute inset-0 z-10 flex items-center justify-center bg-background/80">
            <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载工单详情" />
          </div>

          <div v-if="currentTicket" class="grid lg:grid-cols-[320px_minmax(0,1fr)]">
            <aside class="space-y-6 border-b p-5 lg:border-b-0 lg:border-r">
              <section class="space-y-3">
                <h3 class="text-sm font-semibold">工单信息</h3>
                <dl class="divide-y rounded-lg border">
                  <DetailItem label="状态">
                    <AdminStatusBadge :tone="statusTone(currentTicket.status)">{{ statusName(currentTicket.status) }}</AdminStatusBadge>
                  </DetailItem>
                  <DetailItem label="优先级">
                    <AdminStatusBadge :tone="priorityTone(currentTicket.priority)">{{ priorityName(currentTicket.priority) }}</AdminStatusBadge>
                  </DetailItem>
                  <DetailItem label="分类">{{ categoryName(currentTicket.category) }}</DetailItem>
                  <DetailItem label="客户">{{ customerName(currentTicket) }}</DetailItem>
                  <DetailItem label="负责人">{{ assigneeName(currentTicket.assigned_to) }}</DetailItem>
                  <DetailItem label="创建时间">{{ formatDate(currentTicket.created_at) }}</DetailItem>
                  <DetailItem label="更新时间">{{ formatDate(currentTicket.updated_at) }}</DetailItem>
                  <DetailItem v-if="currentTicket.tags" label="标签">{{ currentTicket.tags }}</DetailItem>
                </dl>
              </section>

              <section v-if="hasPermission('ticket:edit')" class="space-y-3 border-t pt-5">
                <h3 class="text-sm font-semibold">处理操作</h3>
                <label class="block space-y-1.5">
                  <span class="text-xs font-medium text-muted-foreground">状态</span>
                  <Select v-model="statusUpdate">
                    <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                    <SelectContent>
                      <SelectItem v-for="option in editableStatusOptions" :key="option.value" :value="option.value">
                        {{ option.label }}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                </label>
                <Button class="w-full" :disabled="statusUpdating || statusUpdate === currentTicket.status" @click="updateStatus">
                  <LoaderCircle v-if="statusUpdating" class="size-4 animate-spin" />
                  更新状态
                </Button>
                <Button variant="outline" class="w-full" @click="showAssignDialog(currentTicket)">
                  <UserRoundCog class="size-4" />
                  {{ currentTicket.assigned_to ? '更换负责人' : '分配负责人' }}
                </Button>
              </section>
            </aside>

            <section class="flex min-h-[620px] min-w-0 flex-col">
              <div class="flex items-center justify-between border-b px-5 py-3">
                <h3 class="text-sm font-semibold">消息记录</h3>
                <span class="text-xs text-muted-foreground">{{ messages.length }} 条</span>
              </div>

              <div class="relative min-h-64 flex-1 overflow-y-auto px-5 py-4">
                <div v-if="messagesLoading" class="absolute inset-0 flex items-center justify-center bg-background/75">
                  <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载消息" />
                </div>
                <div v-else-if="messages.length === 0" class="flex h-52 flex-col items-center justify-center text-muted-foreground">
                  <MessageCircleOff class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无消息记录</span>
                </div>
                <div v-else class="space-y-3">
                  <article
                    v-for="message in messages"
                    :key="message.id"
                    class="max-w-[88%] rounded-lg border px-3.5 py-3"
                    :class="message.is_staff ? 'ml-auto border-blue-200 bg-blue-50/70' : 'mr-auto bg-muted/40'"
                  >
                    <header class="flex flex-wrap items-center justify-between gap-x-4 gap-y-1">
                      <div class="flex items-center gap-2">
                        <span class="text-xs font-semibold">{{ messageSender(message) }}</span>
                        <AdminStatusBadge :tone="message.is_staff ? 'blue' : 'gray'">
                          {{ message.is_staff ? '客服' : '客户' }}
                        </AdminStatusBadge>
                      </div>
                      <time class="text-[11px] text-muted-foreground">{{ formatDate(message.created_at) }}</time>
                    </header>
                    <p class="mt-2 whitespace-pre-wrap break-words text-sm leading-6">{{ message.content || message.message }}</p>
                  </article>
                </div>
              </div>

              <form v-if="hasPermission('ticket:edit')" class="border-t p-4" @submit.prevent="sendReply">
                <Textarea v-model="replyMessage" class="min-h-24 resize-y" placeholder="输入回复内容" />
                <div class="mt-3 flex justify-end">
                  <Button type="submit" :disabled="replying || !replyMessage.trim()">
                    <LoaderCircle v-if="replying" class="size-4 animate-spin" />
                    <Send v-else class="size-4" />
                    发送回复
                  </Button>
                </div>
              </form>
            </section>
          </div>
        </div>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="assignDialogVisible">
      <DialogContent class="sm:max-w-md" @open-auto-focus.prevent>
        <form class="space-y-5" @submit.prevent="assignTicket">
          <DialogHeader>
            <DialogTitle>分配工单</DialogTitle>
            <DialogDescription>{{ currentTicket?.ticket_number }} · {{ currentTicket?.subject }}</DialogDescription>
          </DialogHeader>
          <label class="block space-y-1.5">
            <span class="text-xs font-medium">负责人</span>
            <Select v-model="assignTo">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择负责人" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="user in supportUsers" :key="user.id" :value="String(user.id)">
                  {{ supportUserName(user) }}
                </SelectItem>
              </SelectContent>
            </Select>
          </label>
          <DialogFooter>
            <Button type="button" variant="outline" @click="assignDialogVisible = false">取消</Button>
            <Button type="submit" :disabled="assigning || !assignTo">
              <LoaderCircle v-if="assigning" class="size-4 animate-spin" />
              确认分配
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      :title="confirmation.title"
      :description="confirmation.description"
      confirm-label="删除"
      destructive
      @confirm="executeDelete"
    />
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  CircleCheck,
  CircleDot,
  CirclePause,
  CircleX,
  Eye,
  LoaderCircle,
  MessageCircleOff,
  MessagesSquare,
  MoreHorizontal,
  RefreshCw,
  RotateCcw,
  Send,
  Trash2,
  UserRoundCog
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const FilterSelect = defineComponent({
  props: {
    modelValue: { type: String, required: true },
    label: { type: String, required: true },
    options: { type: Array, required: true }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('label', { class: 'w-full space-y-1.5 sm:w-52' }, [
      h('span', { class: 'text-xs font-medium text-muted-foreground' }, props.label),
      h(Select, {
        modelValue: props.modelValue,
        'onUpdate:modelValue': (value) => {
          emit('update:modelValue', value)
          applyFilters()
        }
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
  props: { label: { type: String, required: true } },
  setup(props, { slots }) {
    return () => h('div', { class: 'grid grid-cols-[86px_minmax(0,1fr)] gap-3 px-3 py-2.5' }, [
      h('dt', { class: 'text-xs font-medium text-muted-foreground' }, props.label),
      h('dd', { class: 'min-w-0 break-words text-xs' }, slots.default?.())
    ])
  }
})

const authStore = useAuthStore()
const loading = ref(false)
const tickets = ref([])
const stats = ref({})
const detailDialogVisible = ref(false)
const detailLoading = ref(false)
const currentTicket = ref(null)
const messages = ref([])
const messagesLoading = ref(false)
const replyMessage = ref('')
const replying = ref(false)
const statusUpdate = ref('open')
const statusUpdating = ref(false)

const assignDialogVisible = ref(false)
const assignTo = ref('')
const assigning = ref(false)
const supportUsers = ref([])

const filters = reactive({ status: 'all', priority: 'all' })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const confirmation = reactive({ open: false, target: null, title: '', description: '' })

const editableStatusOptions = [
  { label: '待处理', value: 'open' },
  { label: '处理中', value: 'in_progress' },
  { label: '已解决', value: 'resolved' },
  { label: '已关闭', value: 'closed' }
]
const statusFilterOptions = [{ label: '全部状态', value: 'all' }, ...editableStatusOptions]
const priorityFilterOptions = [
  { label: '全部优先级', value: 'all' },
  { label: '低', value: 'low' },
  { label: '中', value: 'medium' },
  { label: '高', value: 'high' },
  { label: '紧急', value: 'urgent' }
]

const statItems = computed(() => [
  { key: 'total', label: '总工单数', value: stats.value.total || 0, icon: MessagesSquare, tone: 'gray' },
  { key: 'open', label: '待处理', value: stats.value.open || 0, icon: CircleDot, tone: 'amber' },
  { key: 'progress', label: '处理中', value: stats.value.in_progress || 0, icon: CirclePause, tone: 'blue' },
  { key: 'closed', label: '已解决/关闭', value: Number(stats.value.resolved || 0) + Number(stats.value.closed || 0), icon: CircleCheck, tone: 'green' }
])

const apiData = (response) => response.data?.data ?? response.data ?? {}
const hasPermission = (permission) => authStore.hasPermission(permission)
const statusName = (status) => ({ open: '待处理', in_progress: '处理中', resolved: '已解决', closed: '已关闭' })[status] || status || '-'
const statusTone = (status) => ({ open: 'amber', in_progress: 'blue', resolved: 'green', closed: 'gray' })[status] || 'gray'
const priorityName = (priority) => ({ low: '低', medium: '中', high: '高', urgent: '紧急' })[priority] || priority || '-'
const priorityTone = (priority) => ({ low: 'gray', medium: 'blue', high: 'amber', urgent: 'coral' })[priority] || 'gray'
const categoryName = (category) => ({ order: '订单', product: '商品', shipping: '物流', customer_service: '在线客服', other: '其他' })[category] || category || '-'
const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
const supportUserName = (user) => user.username || user.email || `用户 ${user.id}`
const customerName = (ticket) => ticket.user_name || ticket.user?.username || ticket.user?.email || `用户 ${ticket.user_id}`
const assigneeName = (assignedTo) => {
  if (!assignedTo) return '未分配'
  const user = supportUsers.value.find((item) => Number(item.id) === Number(assignedTo))
  return user ? supportUserName(user) : `用户 ${assignedTo}`
}
const messageSender = (message) => message.sender_name || message.user?.username || message.user?.email || (message.is_staff ? '客服' : '客户')

const fetchTickets = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/tickets', {
      params: {
        page: pagination.page,
        page_size: pagination.pageSize,
        ...(filters.status !== 'all' ? { status: filters.status } : {}),
        ...(filters.priority !== 'all' ? { priority: filters.priority } : {})
      }
    })
    const data = apiData(response)
    tickets.value = data.tickets || []
    pagination.total = data.pagination?.total ?? 0
  } catch (error) {
    console.error('Failed to fetch tickets:', error)
  } finally {
    loading.value = false
  }
}
const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/tickets/stats')
    stats.value = apiData(response) || {}
  } catch (error) {
    console.error('Failed to fetch ticket stats:', error)
  }
}
const fetchSupportUsers = async () => {
  try {
    const response = await axios.get('/api/admin/users', { params: { role: 'support', page_size: 100 } })
    supportUsers.value = response.data.users || []
  } catch (error) {
    console.error('Failed to fetch support users:', error)
  }
}
const refreshTickets = () => Promise.all([fetchTickets(), fetchStats()])
const applyFilters = () => { pagination.page = 1; fetchTickets() }
const resetFilters = () => { Object.assign(filters, { status: 'all', priority: 'all' }); pagination.page = 1; fetchTickets() }
const updatePage = (page) => { pagination.page = page; fetchTickets() }
const updatePageSize = (pageSize) => { pagination.pageSize = pageSize; pagination.page = 1; fetchTickets() }

const viewTicket = async (ticket) => {
  currentTicket.value = ticket
  messages.value = []
  replyMessage.value = ''
  statusUpdate.value = ticket.status
  detailDialogVisible.value = true
  detailLoading.value = true
  messagesLoading.value = true
  try {
    const [detailResponse, messagesResponse] = await Promise.all([
      axios.get(`/api/admin/tickets/${ticket.id}`),
      axios.get(`/api/admin/tickets/${ticket.id}/messages`)
    ])
    const detailData = apiData(detailResponse)
    const messageData = apiData(messagesResponse)
    currentTicket.value = detailData.ticket || ticket
    statusUpdate.value = currentTicket.value.status
    messages.value = messageData.messages || currentTicket.value.messages || []
    await axios.post(`/api/admin/tickets/${ticket.id}/messages/mark-read`)
  } catch (error) {
    console.error('Failed to fetch ticket detail:', error)
  } finally {
    detailLoading.value = false
    messagesLoading.value = false
  }
}
const fetchMessages = async (ticketId) => {
  messagesLoading.value = true
  try {
    const response = await axios.get(`/api/admin/tickets/${ticketId}/messages`)
    messages.value = apiData(response).messages || []
    await axios.post(`/api/admin/tickets/${ticketId}/messages/mark-read`)
  } catch (error) {
    console.error('Failed to fetch ticket messages:', error)
  } finally {
    messagesLoading.value = false
  }
}
const updateStatus = async () => {
  statusUpdating.value = true
  try {
    await axios.patch(`/api/admin/tickets/${currentTicket.value.id}/status`, { status: statusUpdate.value })
    currentTicket.value.status = statusUpdate.value
    toast.success('工单状态已更新')
    await refreshTickets()
  } catch (error) {
    console.error('Failed to update ticket status:', error)
  } finally {
    statusUpdating.value = false
  }
}
const sendReply = async () => {
  const message = replyMessage.value.trim()
  if (!message) return
  replying.value = true
  try {
    await axios.post(`/api/admin/tickets/${currentTicket.value.id}/messages`, { message })
    replyMessage.value = ''
    toast.success('回复已发送')
    await Promise.all([fetchMessages(currentTicket.value.id), fetchTickets()])
  } catch (error) {
    console.error('Failed to send ticket reply:', error)
  } finally {
    replying.value = false
  }
}

const showAssignDialog = (ticket) => {
  currentTicket.value = ticket
  assignTo.value = ticket.assigned_to ? String(ticket.assigned_to) : ''
  assignDialogVisible.value = true
}
const assignTicket = async () => {
  if (!assignTo.value) return
  assigning.value = true
  try {
    await axios.patch(`/api/admin/tickets/${currentTicket.value.id}/assign`, { assigned_to: Number(assignTo.value) })
    currentTicket.value.assigned_to = Number(assignTo.value)
    currentTicket.value.status = 'in_progress'
    statusUpdate.value = 'in_progress'
    assignDialogVisible.value = false
    toast.success('工单已分配')
    await refreshTickets()
  } catch (error) {
    console.error('Failed to assign ticket:', error)
  } finally {
    assigning.value = false
  }
}

const requestDelete = (ticket) => Object.assign(confirmation, {
  open: true,
  target: ticket,
  title: '删除工单？',
  description: `工单 ${ticket.ticket_number} 及全部消息将被永久删除，此操作不可恢复。`
})
const executeDelete = async () => {
  const ticket = confirmation.target
  confirmation.open = false
  try {
    await axios.delete(`/api/admin/tickets/${ticket.id}`)
    if (currentTicket.value?.id === ticket.id) detailDialogVisible.value = false
    toast.success('工单已删除')
    await refreshTickets()
  } catch (error) {
    console.error('Failed to delete ticket:', error)
  }
}

onMounted(() => Promise.all([fetchTickets(), fetchStats(), fetchSupportUsers()]))
</script>
