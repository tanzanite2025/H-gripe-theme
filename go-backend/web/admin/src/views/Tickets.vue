<template>
  <div class="tickets-page">
    <div class="page-header">
      <h2>工单管理</h2>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">总工单数</div>
            <div class="stat-value">{{ stats.total || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">待处理</div>
            <div class="stat-value" style="color: #e6a23c">{{ stats.open || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">处理中</div>
            <div class="stat-value" style="color: #409eff">{{ stats.in_progress || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">已关闭</div>
            <div class="stat-value" style="color: #67c23a">{{ stats.closed || 0 }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选栏 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters">
        <el-form-item label="搜索">
          <el-input
            v-model="filters.search"
            placeholder="标题/工单号"
            clearable
            @clear="fetchTickets"
            @keyup.enter="fetchTickets"
          />
        </el-form-item>

        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable @change="fetchTickets">
            <el-option label="待处理" value="open" />
            <el-option label="处理中" value="in_progress" />
            <el-option label="已解决" value="resolved" />
            <el-option label="已关闭" value="closed" />
          </el-select>
        </el-form-item>

        <el-form-item label="优先级">
          <el-select v-model="filters.priority" placeholder="全部" clearable @change="fetchTickets">
            <el-option label="低" value="low" />
            <el-option label="中" value="medium" />
            <el-option label="高" value="high" />
            <el-option label="紧急" value="urgent" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchTickets">搜索</el-button>
          <el-button :icon="Refresh" @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 工单列表 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="tickets"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="ticket_number" label="工单号" width="150" />
        
        <el-table-column prop="subject" label="标题" min-width="200" show-overflow-tooltip />
        
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="优先级" width="100">
          <template #default="{ row }">
            <el-tag :type="getPriorityType(row.priority)">
              {{ getPriorityName(row.priority) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="user_name" label="用户" width="120" />
        
        <el-table-column prop="assigned_to_name" label="负责人" width="120">
          <template #default="{ row }">
            {{ row.assigned_to_name || '未分配' }}
          </template>
        </el-table-column>
        
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              link
              @click="viewTicket(row)"
            >
              查看
            </el-button>
            <el-button
              v-if="hasPermission('ticket:edit')"
              type="warning"
              size="small"
              link
              @click="showAssignDialog(row)"
            >
              分配
            </el-button>
            <el-button
              v-if="hasPermission('ticket:delete')"
              type="danger"
              size="small"
              link
              @click="deleteTicket(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchTickets"
        @current-change="fetchTickets"
      />
    </el-card>

    <!-- 工单详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      :title="`工单详情 - ${currentTicket?.ticket_number}`"
      width="900px"
      top="5vh"
    >
      <div v-if="currentTicket" class="ticket-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="工单号">{{ currentTicket.ticket_number }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(currentTicket.status)">
              {{ getStatusName(currentTicket.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="优先级">
            <el-tag :type="getPriorityType(currentTicket.priority)">
              {{ getPriorityName(currentTicket.priority) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="用户">{{ currentTicket.user_name }}</el-descriptions-item>
          <el-descriptions-item label="负责人">{{ currentTicket.assigned_to_name || '未分配' }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDate(currentTicket.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="标题" :span="2">{{ currentTicket.subject }}</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">
            <div style="white-space: pre-wrap;">{{ currentTicket.description }}</div>
          </el-descriptions-item>
        </el-descriptions>

        <!-- 状态更新 -->
        <div v-if="hasPermission('ticket:edit')" class="status-update">
          <el-select v-model="statusUpdate" placeholder="更新状态" style="width: 200px; margin-right: 10px;">
            <el-option label="待处理" value="open" />
            <el-option label="处理中" value="in_progress" />
            <el-option label="已解决" value="resolved" />
            <el-option label="已关闭" value="closed" />
          </el-select>
          <el-button type="primary" @click="updateStatus">更新状态</el-button>
        </div>

        <!-- 消息列表 -->
        <div class="messages-section">
          <h3>消息记录</h3>
          <div v-loading="messagesLoading" class="messages-list">
            <div v-for="msg in messages" :key="msg.id" class="message-item">
              <div class="message-header">
                <span class="message-sender">{{ msg.sender_name }}</span>
                <span class="message-time">{{ formatDate(msg.created_at) }}</span>
              </div>
              <div class="message-content">{{ msg.message }}</div>
            </div>
          </div>

          <!-- 回复表单 -->
          <div v-if="hasPermission('ticket:edit')" class="reply-form">
            <el-input
              v-model="replyMessage"
              type="textarea"
              :rows="3"
              placeholder="输入回复内容..."
            />
            <el-button
              type="primary"
              :loading="replying"
              style="margin-top: 10px;"
              @click="sendReply"
            >
              发送回复
            </el-button>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 分配工单对话框 -->
    <el-dialog
      v-model="assignDialogVisible"
      title="分配工单"
      width="400px"
    >
      <el-form label-width="80px">
        <el-form-item label="负责人">
          <el-select v-model="assignTo" placeholder="请选择负责人" style="width: 100%">
            <el-option
              v-for="user in supportUsers"
              :key="user.id"
              :label="user.username"
              :value="user.id"
            />
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="assignDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="assigning" @click="assignTicket">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()

const loading = ref(false)
const tickets = ref([])
const stats = ref({})
const detailDialogVisible = ref(false)
const currentTicket = ref(null)
const messages = ref([])
const messagesLoading = ref(false)
const replyMessage = ref('')
const replying = ref(false)
const statusUpdate = ref('')

const assignDialogVisible = ref(false)
const assignTo = ref(null)
const assigning = ref(false)
const supportUsers = ref([])

const filters = reactive({
  search: '',
  status: '',
  priority: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const hasPermission = (permission) => {
  return authStore.hasPermission(permission)
}

const getStatusName = (status) => {
  const map = {
    open: '待处理',
    in_progress: '处理中',
    resolved: '已解决',
    closed: '已关闭'
  }
  return map[status] || status
}

const getStatusType = (status) => {
  const map = {
    open: 'warning',
    in_progress: 'primary',
    resolved: 'success',
    closed: 'info'
  }
  return map[status] || ''
}

const getPriorityName = (priority) => {
  const map = {
    low: '低',
    medium: '中',
    high: '高',
    urgent: '紧急'
  }
  return map[priority] || priority
}

const getPriorityType = (priority) => {
  const map = {
    low: 'info',
    medium: '',
    high: 'warning',
    urgent: 'danger'
  }
  return map[priority] || ''
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

const fetchTickets = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...filters
    }

    const response = await axios.get('/api/admin/tickets', { params })
    tickets.value = response.data.tickets
    pagination.total = response.data.total
  } catch (error) {
    ElMessage.error('获取工单列表失败')
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/tickets/stats')
    stats.value = response.data
  } catch (error) {
    console.error('获取统计失败', error)
  }
}

const resetFilters = () => {
  filters.search = ''
  filters.status = ''
  filters.priority = ''
  pagination.page = 1
  fetchTickets()
}

const viewTicket = async (ticket) => {
  currentTicket.value = ticket
  statusUpdate.value = ticket.status
  detailDialogVisible.value = true
  await fetchMessages(ticket.id)
}

const fetchMessages = async (ticketId) => {
  messagesLoading.value = true
  try {
    const response = await axios.get(`/api/admin/tickets/${ticketId}/messages`)
    messages.value = response.data.messages
  } catch (error) {
    ElMessage.error('获取消息失败')
  } finally {
    messagesLoading.value = false
  }
}

const updateStatus = async () => {
  try {
    await axios.patch(`/api/admin/tickets/${currentTicket.value.id}/status`, {
      status: statusUpdate.value
    })
    ElMessage.success('状态更新成功')
    currentTicket.value.status = statusUpdate.value
    fetchTickets()
    fetchStats()
  } catch (error) {
    ElMessage.error('状态更新失败')
  }
}

const sendReply = async () => {
  if (!replyMessage.value.trim()) {
    ElMessage.warning('请输入回复内容')
    return
  }

  replying.value = true
  try {
    await axios.post(`/api/admin/tickets/${currentTicket.value.id}/messages`, {
      message: replyMessage.value
    })
    ElMessage.success('回复成功')
    replyMessage.value = ''
    fetchMessages(currentTicket.value.id)
  } catch (error) {
    ElMessage.error('回复失败')
  } finally {
    replying.value = false
  }
}

const showAssignDialog = (ticket) => {
  currentTicket.value = ticket
  assignTo.value = ticket.assigned_to
  assignDialogVisible.value = true
}

const assignTicket = async () => {
  if (!assignTo.value) {
    ElMessage.warning('请选择负责人')
    return
  }

  assigning.value = true
  try {
    await axios.patch(`/api/admin/tickets/${currentTicket.value.id}/assign`, {
      assigned_to: assignTo.value
    })
    ElMessage.success('分配成功')
    assignDialogVisible.value = false
    fetchTickets()
  } catch (error) {
    ElMessage.error('分配失败')
  } finally {
    assigning.value = false
  }
}

const deleteTicket = async (ticket) => {
  try {
    await ElMessageBox.confirm(`确定要删除工单 ${ticket.ticket_number} 吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    await axios.delete(`/api/admin/tickets/${ticket.id}`)
    ElMessage.success('删除成功')
    fetchTickets()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const fetchSupportUsers = async () => {
  try {
    const response = await axios.get('/api/admin/users', {
      params: { role: 'support', page_size: 100 }
    })
    supportUsers.value = response.data.users
  } catch (error) {
    console.error('获取客服列表失败', error)
  }
}

onMounted(() => {
  fetchTickets()
  fetchStats()
  fetchSupportUsers()
})
</script>

<style scoped>
.tickets-page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-item {
  text-align: center;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
}

.filter-card {
  margin-bottom: 20px;
}

.table-card {
  margin-bottom: 20px;
}

.el-pagination {
  margin-top: 20px;
  justify-content: flex-end;
}

.ticket-detail {
  padding: 10px 0;
}

.status-update {
  margin: 20px 0;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 4px;
}

.messages-section {
  margin-top: 20px;
}

.messages-section h3 {
  margin-bottom: 15px;
  font-size: 16px;
  color: #303133;
}

.messages-list {
  max-height: 400px;
  overflow-y: auto;
  margin-bottom: 15px;
}

.message-item {
  padding: 12px;
  margin-bottom: 10px;
  background: #f5f7fa;
  border-radius: 4px;
}

.message-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 8px;
}

.message-sender {
  font-weight: bold;
  color: #303133;
}

.message-time {
  font-size: 12px;
  color: #909399;
}

.message-content {
  color: #606266;
  white-space: pre-wrap;
}

.reply-form {
  padding: 15px;
  background: #f5f7fa;
  border-radius: 4px;
}
</style>
