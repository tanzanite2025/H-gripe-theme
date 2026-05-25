<template>
  <div class="subscriptions-page">
    <div class="page-header">
      <h2>订阅管理</h2>
      <el-button
        v-if="hasPermission('subscription:export')"
        type="success"
        :icon="Download"
        @click="exportEmails"
      >
        导出邮箱
      </el-button>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">总订阅数</div>
            <div class="stat-value">{{ stats.total || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">活跃订阅</div>
            <div class="stat-value" style="color: #67c23a">{{ stats.active || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">已取消</div>
            <div class="stat-value" style="color: #f56c6c">{{ stats.cancelled || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div class="stat-item">
            <div class="stat-label">今日新增</div>
            <div class="stat-value" style="color: #409eff">{{ stats.today || 0 }}</div>
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
            placeholder="邮箱"
            clearable
            @clear="fetchSubscriptions"
            @keyup.enter="fetchSubscriptions"
          />
        </el-form-item>

        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable @change="fetchSubscriptions">
            <el-option label="活跃" value="active" />
            <el-option label="已取消" value="cancelled" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchSubscriptions">搜索</el-button>
          <el-button :icon="Refresh" @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 订阅列表 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="subscriptions"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        
        <el-table-column prop="email" label="邮箱" min-width="200" />
        
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'">
              {{ row.status === 'active' ? '活跃' : '已取消' }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column prop="source" label="来源" width="120">
          <template #default="{ row }">
            {{ getSourceName(row.source) }}
          </template>
        </el-table-column>
        
        <el-table-column prop="subscribed_at" label="订阅时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.subscribed_at) }}
          </template>
        </el-table-column>
        
        <el-table-column prop="unsubscribed_at" label="取消时间" width="180">
          <template #default="{ row }">
            {{ row.unsubscribed_at ? formatDate(row.unsubscribed_at) : '-' }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="hasPermission('subscription:edit')"
              :type="row.status === 'active' ? 'warning' : 'success'"
              size="small"
              link
              @click="toggleStatus(row)"
            >
              {{ row.status === 'active' ? '取消订阅' : '恢复订阅' }}
            </el-button>
            <el-button
              v-if="hasPermission('subscription:delete')"
              type="danger"
              size="small"
              link
              @click="deleteSubscription(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 批量操作 -->
      <div v-if="selectedSubscriptions.length > 0" class="batch-actions">
        <span>已选择 {{ selectedSubscriptions.length }} 项</span>
        <el-button
          v-if="hasPermission('subscription:delete')"
          type="danger"
          size="small"
          @click="batchDelete"
        >
          批量删除
        </el-button>
      </div>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchSubscriptions"
        @current-change="fetchSubscriptions"
      />
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Download, Search, Refresh } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()

const loading = ref(false)
const subscriptions = ref([])
const selectedSubscriptions = ref([])
const stats = ref({})

const filters = reactive({
  search: '',
  status: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const hasPermission = (permission) => {
  return authStore.hasPermission(permission)
}

const getSourceName = (source) => {
  const sourceMap = {
    website: '网站',
    popup: '弹窗',
    footer: '页脚',
    checkout: '结账页'
  }
  return sourceMap[source] || source
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

const fetchSubscriptions = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...filters
    }

    const response = await axios.get('/api/admin/subscriptions', { params })
    subscriptions.value = response.data.subscriptions
    pagination.total = response.data.total
  } catch (error) {
    ElMessage.error('获取订阅列表失败')
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/subscriptions/stats')
    stats.value = response.data
  } catch (error) {
    console.error('获取统计失败', error)
  }
}

const resetFilters = () => {
  filters.search = ''
  filters.status = ''
  pagination.page = 1
  fetchSubscriptions()
}

const toggleStatus = async (subscription) => {
  const newStatus = subscription.status === 'active' ? 'cancelled' : 'active'
  const action = newStatus === 'active' ? '恢复' : '取消'

  try {
    await ElMessageBox.confirm(`确定要${action}订阅 ${subscription.email} 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await axios.patch(`/api/admin/subscriptions/${subscription.email}/status`, { status: newStatus })
    ElMessage.success(`${action}成功`)
    fetchSubscriptions()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`${action}失败`)
    }
  }
}

const deleteSubscription = async (subscription) => {
  try {
    await ElMessageBox.confirm(`确定要删除订阅 ${subscription.email} 吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    await axios.delete(`/api/admin/subscriptions/${subscription.email}`)
    ElMessage.success('删除成功')
    fetchSubscriptions()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleSelectionChange = (selection) => {
  selectedSubscriptions.value = selection
}

const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedSubscriptions.value.length} 个订阅吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    const emails = selectedSubscriptions.value.map(s => s.email)
    await axios.post('/api/admin/subscriptions/batch-delete', { emails })
    ElMessage.success('批量删除成功')
    fetchSubscriptions()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

const exportEmails = async () => {
  try {
    const response = await axios.get('/api/admin/subscriptions/active-emails')
    const emails = response.data.emails.join('\n')
    
    const blob = new Blob([emails], { type: 'text/plain' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `subscriptions_${new Date().toISOString().split('T')[0]}.txt`
    link.click()
    window.URL.revokeObjectURL(url)
    
    ElMessage.success('导出成功')
  } catch (error) {
    ElMessage.error('导出失败')
  }
}

onMounted(() => {
  fetchSubscriptions()
  fetchStats()
})
</script>

<style scoped>
.subscriptions-page {
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

.batch-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-top: 1px solid #ebeef5;
  margin-top: 12px;
}

.el-pagination {
  margin-top: 20px;
  justify-content: flex-end;
}
</style>
