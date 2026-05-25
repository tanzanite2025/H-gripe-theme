<template>
  <div class="audit-logs-page">
    <div class="page-header">
      <h2>审计日志</h2>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">总日志数</div>
            <div class="stat-value">{{ stats.total_count || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">今日操作</div>
            <div class="stat-value" style="color: #409eff">{{ todayCount }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">成功操作</div>
            <div class="stat-value" style="color: #67c23a">{{ successCount }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">失败操作</div>
            <div class="stat-value" style="color: #f56c6c">{{ failedCount }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选栏 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters">
        <el-form-item label="搜索">
          <el-input
            v-model="filters.keyword"
            placeholder="关键词"
            clearable
            @clear="fetchLogs"
            @keyup.enter="fetchLogs"
          />
        </el-form-item>

        <el-form-item label="操作">
          <el-select v-model="filters.action" placeholder="全部" clearable @change="fetchLogs">
            <el-option label="创建" value="create" />
            <el-option label="更新" value="update" />
            <el-option label="删除" value="delete" />
            <el-option label="查看" value="view" />
          </el-select>
        </el-form-item>

        <el-form-item label="资源">
          <el-select v-model="filters.resource" placeholder="全部" clearable @change="fetchLogs">
            <el-option label="用户" value="user" />
            <el-option label="商品" value="product" />
            <el-option label="订单" value="order" />
            <el-option label="文章" value="post" />
            <el-option label="工单" value="ticket" />
          </el-select>
        </el-form-item>

        <el-form-item label="日期范围">
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            @change="fetchLogs"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchLogs">搜索</el-button>
          <el-button :icon="Refresh" @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 日志列表 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="logs"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="80" />
        
        <el-table-column prop="username" label="用户" width="120" />
        
        <el-table-column label="操作" width="100">
          <template #default="{ row }">
            <el-tag :type="getActionType(row.action)">
              {{ getActionName(row.action) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="resource" label="资源" width="100">
          <template #default="{ row }">
            {{ getResourceName(row.resource) }}
          </template>
        </el-table-column>
        
        <el-table-column prop="resource_id" label="资源ID" width="100" />
        
        <el-table-column prop="method" label="方法" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="getMethodType(row.method)">
              {{ row.method }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="path" label="路径" min-width="200" show-overflow-tooltip />
        
        <el-table-column prop="ip_address" label="IP 地址" width="140" />
        
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : 'danger'">
              {{ row.status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="duration" label="耗时(ms)" width="100" />
        
        <el-table-column prop="created_at" label="时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              link
              @click="viewDetail(row)"
            >
              详情
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
        @size-change="fetchLogs"
        @current-change="fetchLogs"
      />
    </el-card>

    <!-- 详情对话框 -->
    <el-dialog
      v-model="detailDialogVisible"
      title="日志详情"
      width="800px"
    >
      <div v-if="currentLog" class="log-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="ID">{{ currentLog.id }}</el-descriptions-item>
          <el-descriptions-item label="用户">{{ currentLog.username }}</el-descriptions-item>
          <el-descriptions-item label="操作">
            <el-tag :type="getActionType(currentLog.action)">
              {{ getActionName(currentLog.action) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="资源">{{ getResourceName(currentLog.resource) }}</el-descriptions-item>
          <el-descriptions-item label="资源ID">{{ currentLog.resource_id }}</el-descriptions-item>
          <el-descriptions-item label="方法">
            <el-tag size="small" :type="getMethodType(currentLog.method)">
              {{ currentLog.method }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="路径" :span="2">{{ currentLog.path }}</el-descriptions-item>
          <el-descriptions-item label="IP 地址">{{ currentLog.ip_address }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="currentLog.status === 'success' ? 'success' : 'danger'">
              {{ currentLog.status === 'success' ? '成功' : '失败' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="耗时">{{ currentLog.duration }} ms</el-descriptions-item>
          <el-descriptions-item label="时间">{{ formatDate(currentLog.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="User Agent" :span="2">
            <div style="word-break: break-all;">{{ currentLog.user_agent }}</div>
          </el-descriptions-item>
        </el-descriptions>

        <div v-if="currentLog.changes" class="changes-section">
          <h4>变更内容</h4>
          <pre>{{ formatJSON(currentLog.changes) }}</pre>
        </div>

        <div v-if="currentLog.error_message" class="error-section">
          <h4>错误信息</h4>
          <el-alert type="error" :closable="false">
            {{ currentLog.error_message }}
          </el-alert>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Refresh } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()

const loading = ref(false)
const logs = ref([])
const stats = ref({})
const dateRange = ref([])
const detailDialogVisible = ref(false)
const currentLog = ref(null)

const filters = reactive({
  keyword: '',
  action: '',
  resource: '',
  user_id: '',
  ip_address: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const todayCount = computed(() => {
  // 简化计算，实际应该从后端获取
  return 0
})

const successCount = computed(() => {
  // 简化计算，实际应该从后端获取
  return 0
})

const failedCount = computed(() => {
  // 简化计算，实际应该从后端获取
  return 0
})

const getActionName = (action) => {
  const map = {
    create: '创建',
    update: '更新',
    delete: '删除',
    view: '查看'
  }
  return map[action] || action
}

const getActionType = (action) => {
  const map = {
    create: 'success',
    update: 'warning',
    delete: 'danger',
    view: 'info'
  }
  return map[action] || ''
}

const getResourceName = (resource) => {
  const map = {
    user: '用户',
    product: '商品',
    order: '订单',
    post: '文章',
    ticket: '工单',
    faq: 'FAQ',
    gallery: '图库',
    subscription: '订阅'
  }
  return map[resource] || resource
}

const getMethodType = (method) => {
  const map = {
    GET: 'info',
    POST: 'success',
    PUT: 'warning',
    PATCH: 'warning',
    DELETE: 'danger'
  }
  return map[method] || ''
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

const formatJSON = (jsonString) => {
  try {
    return JSON.stringify(JSON.parse(jsonString), null, 2)
  } catch {
    return jsonString
  }
}

const fetchLogs = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...filters
    }

    if (dateRange.value && dateRange.value.length === 2) {
      params.start_date = dateRange.value[0].toISOString().split('T')[0]
      params.end_date = dateRange.value[1].toISOString().split('T')[0]
    }

    const response = await axios.get('/api/admin/logs', { params })
    logs.value = response.data.logs
    pagination.total = response.data.total
  } catch (error) {
    ElMessage.error('获取审计日志失败')
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/logs/stats')
    stats.value = response.data
  } catch (error) {
    console.error('获取统计失败', error)
  }
}

const resetFilters = () => {
  filters.keyword = ''
  filters.action = ''
  filters.resource = ''
  filters.user_id = ''
  filters.ip_address = ''
  dateRange.value = []
  pagination.page = 1
  fetchLogs()
}

const viewDetail = async (log) => {
  try {
    const response = await axios.get(`/api/admin/logs/${log.id}`)
    currentLog.value = response.data.log
    detailDialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取日志详情失败')
  }
}

onMounted(() => {
  fetchLogs()
  fetchStats()
})
</script>

<style scoped>
.audit-logs-page {
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

.log-detail {
  padding: 10px 0;
}

.changes-section,
.error-section {
  margin-top: 20px;
}

.changes-section h4,
.error-section h4 {
  margin-bottom: 10px;
  font-size: 14px;
  color: #303133;
}

.changes-section pre {
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.5;
}
</style>
