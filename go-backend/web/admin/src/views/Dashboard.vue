<template>
  <div class="dashboard">
    <h2 class="page-title">仪表板</h2>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" @click="$router.push('/orders')">
          <div class="stat-content">
            <div class="stat-icon orders">
              <el-icon><ShoppingCart /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.orders?.total || 0 }}</div>
              <div class="stat-label">总订单数</div>
              <div class="stat-sub">今日: {{ stats.orders?.today || 0 }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" @click="$router.push('/users')">
          <div class="stat-content">
            <div class="stat-icon users">
              <el-icon><User /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.users?.total || 0 }}</div>
              <div class="stat-label">总用户数</div>
              <div class="stat-sub">今日: {{ stats.users?.today || 0 }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-icon revenue">
              <el-icon><Money /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">¥{{ formatNumber(stats.orders?.revenue || 0) }}</div>
              <div class="stat-label">总销售额</div>
              <div class="stat-sub">今日: ¥{{ formatNumber(stats.orders?.today_revenue || 0) }}</div>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" @click="$router.push('/tickets')">
          <div class="stat-content">
            <div class="stat-icon tickets">
              <el-icon><ChatDotRound /></el-icon>
            </div>
            <div class="stat-info">
              <div class="stat-value">{{ stats.tickets?.open || 0 }}</div>
              <div class="stat-label">待处理工单</div>
              <div class="stat-sub">总计: {{ stats.tickets?.total || 0 }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 销售图表和快速操作 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :xs="24" :lg="16">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>销售趋势（最近30天）</span>
              <el-button size="small" :icon="Refresh" @click="fetchSalesChart">刷新</el-button>
            </div>
          </template>
          <div v-loading="chartLoading" style="height: 350px">
            <v-chart v-if="chartOption" :option="chartOption" autoresize />
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="8">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>快速操作</span>
            </div>
          </template>
          <div class="quick-actions-grid">
            <div
              v-if="hasPermission('product:create')"
              class="quick-action-item"
              @click="$router.push('/products')"
            >
              <el-icon class="action-icon" color="#409eff"><Goods /></el-icon>
              <span>添加商品</span>
            </div>
            <div
              v-if="hasPermission('order:view')"
              class="quick-action-item"
              @click="$router.push('/orders')"
            >
              <el-icon class="action-icon" color="#67c23a"><ShoppingCart /></el-icon>
              <span>查看订单</span>
            </div>
            <div
              v-if="hasPermission('user:view')"
              class="quick-action-item"
              @click="$router.push('/users')"
            >
              <el-icon class="action-icon" color="#e6a23c"><User /></el-icon>
              <span>用户管理</span>
            </div>
            <div
              v-if="hasPermission('ticket:view')"
              class="quick-action-item"
              @click="$router.push('/tickets')"
            >
              <el-icon class="action-icon" color="#f56c6c"><ChatDotRound /></el-icon>
              <span>工单管理</span>
            </div>
            <div
              v-if="hasPermission('content:view')"
              class="quick-action-item"
              @click="$router.push('/content')"
            >
              <el-icon class="action-icon" color="#909399"><Document /></el-icon>
              <span>内容管理</span>
            </div>
            <div
              v-if="hasPermission('settings:view')"
              class="quick-action-item"
              @click="$router.push('/settings')"
            >
              <el-icon class="action-icon" color="#606266"><Setting /></el-icon>
              <span>系统设置</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近活动 -->
    <el-row :gutter="20" style="margin-top: 20px">
      <el-col :xs="24" :lg="8">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近订单</span>
              <el-button size="small" link @click="$router.push('/orders')">查看全部</el-button>
            </div>
          </template>
          <el-empty v-if="recentOrders.length === 0" description="暂无订单" :image-size="80" />
          <div v-else class="recent-list">
            <div v-for="order in recentOrders" :key="order.id" class="recent-item">
              <div class="recent-item-content">
                <div class="recent-item-title">#{{ order.order_number }}</div>
                <div class="recent-item-desc">¥{{ order.total_amount }}</div>
              </div>
              <el-tag :type="getOrderStatusType(order.status)" size="small">
                {{ getOrderStatusName(order.status) }}
              </el-tag>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="8">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近用户</span>
              <el-button size="small" link @click="$router.push('/users')">查看全部</el-button>
            </div>
          </template>
          <el-empty v-if="recentUsers.length === 0" description="暂无用户" :image-size="80" />
          <div v-else class="recent-list">
            <div v-for="user in recentUsers" :key="user.id" class="recent-item">
              <div class="recent-item-content">
                <div class="recent-item-title">{{ user.username }}</div>
                <div class="recent-item-desc">{{ user.email }}</div>
              </div>
              <el-tag :type="getRoleType(user.role)" size="small">
                {{ getRoleName(user.role) }}
              </el-tag>
            </div>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="8">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近工单</span>
              <el-button size="small" link @click="$router.push('/tickets')">查看全部</el-button>
            </div>
          </template>
          <el-empty v-if="recentTickets.length === 0" description="暂无工单" :image-size="80" />
          <div v-else class="recent-list">
            <div v-for="ticket in recentTickets" :key="ticket.id" class="recent-item">
              <div class="recent-item-content">
                <div class="recent-item-title">{{ ticket.subject }}</div>
                <div class="recent-item-desc">{{ ticket.category }}</div>
              </div>
              <el-tag :type="getTicketStatusType(ticket.status)" size="small">
                {{ getTicketStatusName(ticket.status) }}
              </el-tag>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ShoppingCart, User, Money, ChatDotRound, Goods, Document, Setting, Refresh } from '@element-plus/icons-vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

// 注册 ECharts 组件
use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent])

const authStore = useAuthStore()
const user = computed(() => authStore.user)

const stats = ref({})
const chartLoading = ref(false)
const chartOption = ref(null)
const recentOrders = ref([])
const recentUsers = ref([])
const recentTickets = ref([])

const hasPermission = (permission) => {
  return authStore.hasPermission(permission)
}

const formatNumber = (num) => {
  return Number(num).toLocaleString('zh-CN', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
}

const getRoleName = (role) => {
  const roleMap = {
    admin: '管理员',
    manager: '经理',
    editor: '编辑',
    support: '客服',
    viewer: '查看者'
  }
  return roleMap[role] || role
}

const getRoleType = (role) => {
  const typeMap = {
    admin: 'danger',
    manager: 'warning',
    editor: 'success',
    support: 'info',
    viewer: ''
  }
  return typeMap[role] || ''
}

const getOrderStatusName = (status) => {
  const statusMap = {
    pending: '待付款',
    paid: '已付款',
    shipped: '已发货',
    completed: '已完成',
    cancelled: '已取消'
  }
  return statusMap[status] || status
}

const getOrderStatusType = (status) => {
  const typeMap = {
    pending: 'warning',
    paid: 'success',
    shipped: 'primary',
    completed: 'info',
    cancelled: 'danger'
  }
  return typeMap[status] || ''
}

const getTicketStatusName = (status) => {
  const statusMap = {
    open: '待处理',
    pending: '处理中',
    resolved: '已解决',
    closed: '已关闭'
  }
  return statusMap[status] || status
}

const getTicketStatusType = (status) => {
  const typeMap = {
    open: 'danger',
    pending: 'warning',
    resolved: 'success',
    closed: 'info'
  }
  return typeMap[status] || ''
}

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/dashboard/stats')
    stats.value = response.data
  } catch (error) {
    console.error('Failed to fetch stats:', error)
  }
}

const fetchSalesChart = async () => {
  chartLoading.value = true
  try {
    const response = await axios.get('/api/admin/dashboard/sales-chart')
    const data = response.data.data || []

    const dates = data.map(item => item.date)
    const counts = data.map(item => item.count)
    const amounts = data.map(item => item.amount)

    chartOption.value = {
      tooltip: {
        trigger: 'axis'
      },
      legend: {
        data: ['订单数', '销售额']
      },
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: dates
      },
      yAxis: [
        {
          type: 'value',
          name: '订单数',
          position: 'left'
        },
        {
          type: 'value',
          name: '销售额',
          position: 'right'
        }
      ],
      series: [
        {
          name: '订单数',
          type: 'line',
          data: counts,
          smooth: true,
          itemStyle: { color: '#409eff' }
        },
        {
          name: '销售额',
          type: 'line',
          yAxisIndex: 1,
          data: amounts,
          smooth: true,
          itemStyle: { color: '#67c23a' }
        }
      ]
    }
  } catch (error) {
    console.error('Failed to fetch sales chart:', error)
  } finally {
    chartLoading.value = false
  }
}

const fetchRecentOrders = async () => {
  try {
    const response = await axios.get('/api/admin/dashboard/recent-orders')
    recentOrders.value = response.data.orders || []
  } catch (error) {
    console.error('Failed to fetch recent orders:', error)
  }
}

const fetchRecentUsers = async () => {
  try {
    const response = await axios.get('/api/admin/dashboard/recent-users')
    recentUsers.value = response.data.users || []
  } catch (error) {
    console.error('Failed to fetch recent users:', error)
  }
}

const fetchRecentTickets = async () => {
  try {
    const response = await axios.get('/api/admin/dashboard/recent-tickets')
    recentTickets.value = response.data.tickets || []
  } catch (error) {
    console.error('Failed to fetch recent tickets:', error)
  }
}

onMounted(() => {
  fetchStats()
  fetchSalesChart()
  fetchRecentOrders()
  fetchRecentUsers()
  fetchRecentTickets()
})
</script>

<style scoped>
.dashboard {
  padding: 0;
}

.page-title {
  margin: 0 0 20px 0;
  font-size: 24px;
  color: #303133;
}

.stats-row {
  margin-bottom: 20px;
}

.stat-card {
  cursor: pointer;
  transition: transform 0.3s, box-shadow 0.3s;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  color: #fff;
}

.stat-icon.orders {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.stat-icon.users {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
}

.stat-icon.revenue {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
}

.stat-icon.tickets {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 4px;
}

.stat-sub {
  font-size: 12px;
  color: #c0c4cc;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: bold;
  font-size: 16px;
}

.quick-actions-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.quick-action-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 20px;
  border: 1px solid #ebeef5;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
}

.quick-action-item:hover {
  border-color: #409eff;
  background-color: #f5f7fa;
  transform: translateY(-2px);
}

.action-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.quick-action-item span {
  font-size: 14px;
  color: #606266;
}

.recent-list {
  max-height: 350px;
  overflow-y: auto;
}

.recent-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #ebeef5;
}

.recent-item:last-child {
  border-bottom: none;
}

.recent-item-content {
  flex: 1;
  min-width: 0;
}

.recent-item-title {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.recent-item-desc {
  font-size: 12px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
