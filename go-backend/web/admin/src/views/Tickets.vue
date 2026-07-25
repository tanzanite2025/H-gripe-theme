<template>
  <div class="space-y-4">
    <AdminPageHeader title="工单管理" description="处理客户请求、分配负责人并跟进消息记录" />

    <AdminStatsGrid :items="statItems" />

    <TicketFilterPanel
      :filters="filters"
      :status-filter-options="statusFilterOptions"
      :priority-filter-options="priorityFilterOptions"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <TicketTablePanel
      :loading="loading"
      :tickets="tickets"
      :pagination="pagination"
      :can-edit="hasPermission('ticket:edit')"
      :can-delete="hasPermission('ticket:delete')"
      :category-name="categoryName"
      :status-name="statusName"
      :status-tone="statusTone"
      :priority-name="priorityName"
      :priority-tone="priorityTone"
      :customer-name="customerName"
      :assignee-name="assigneeName"
      :format-date="formatDate"
      @view="viewTicket"
      @assign="showAssignDialog"
      @delete="requestDelete"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />

    <TicketDetailDialog
      v-model:open="detailDialogVisible"
      v-model:reply-message="replyMessage"
      v-model:status-update="statusUpdate"
      :current-ticket="currentTicket"
      :detail-loading="detailLoading"
      :messages="messages"
      :messages-loading="messagesLoading"
      :replying="replying"
      :status-updating="statusUpdating"
      :editable-status-options="editableStatusOptions"
      :can-edit="hasPermission('ticket:edit')"
      :category-name="categoryName"
      :status-name="statusName"
      :status-tone="statusTone"
      :priority-name="priorityName"
      :priority-tone="priorityTone"
      :customer-name="customerName"
      :assignee-name="assigneeName"
      :message-sender="messageSender"
      :format-date="formatDate"
      @update-status="updateStatus"
      @show-assign="showAssignDialog"
      @send-reply="sendReply"
    />

    <TicketAssignDialog
      v-model:open="assignDialogVisible"
      v-model:assign-to="assignTo"
      :current-ticket="currentTicket"
      :assigning="assigning"
      :support-users="supportUsers"
      :support-user-name="supportUserName"
      @submit="assignTicket"
    />

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
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  CircleCheck,
  CircleDot,
  CirclePause,
  MessagesSquare,
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import TicketAssignDialog from '@/components/admin/ticket/TicketAssignDialog.vue'
import TicketDetailDialog from '@/components/admin/ticket/TicketDetailDialog.vue'
import TicketFilterPanel from '@/components/admin/ticket/TicketFilterPanel.vue'
import TicketTablePanel from '@/components/admin/ticket/TicketTablePanel.vue'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

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
