<template>
  <div class="space-y-4 animate-in fade-in duration-500">
    <DashboardHeader :current-date="currentDate" />

    <DashboardMetricGrid
      :metric-cards="metricCards"
      :metric-tone-class="metricToneClass"
      @navigate="navigateTo"
    />

    <section class="grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(280px,0.8fr)]">
      <DashboardSalesChartPanel
        :chart-loading="chartLoading"
        :chart-option="chartOption"
        @refresh="fetchSalesChart"
      />

      <DashboardQuickActionsPanel
        :actions="visibleQuickActions"
        :metric-tone-class="metricToneClass"
        @navigate="navigateTo"
      />
    </section>

    <DashboardActivityPanel
      v-model:active-activity="activeActivity"
      :recent-orders="recentOrders"
      :recent-users="recentUsers"
      :recent-tickets="recentTickets"
      :format-number="formatNumber"
      :get-order-status-name="getOrderStatusName"
      :order-status-tone="orderStatusTone"
      :get-role-name="getRoleName"
      :role-tone="roleTone"
      :get-ticket-status-name="getTicketStatusName"
      :ticket-status-tone="ticketStatusTone"
      @navigate="navigateTo"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  FileText,
  Headset,
  MessagesSquare,
  PackagePlus,
  Settings,
  ShoppingCart,
  Users,
  WalletCards
} from '@lucide/vue'
import DashboardActivityPanel from '@/components/admin/dashboard/DashboardActivityPanel.vue'
import DashboardHeader from '@/components/admin/dashboard/DashboardHeader.vue'
import DashboardMetricGrid from '@/components/admin/dashboard/DashboardMetricGrid.vue'
import DashboardQuickActionsPanel from '@/components/admin/dashboard/DashboardQuickActionsPanel.vue'
import DashboardSalesChartPanel from '@/components/admin/dashboard/DashboardSalesChartPanel.vue'
import {
  buildSalesChartOption,
  currentDashboardDate,
  formatNumber,
  getOrderStatusName,
  getRoleName,
  getTicketStatusName,
  metricToneClass,
  orderStatusTone,
  roleTone,
  ticketStatusTone
} from '@/lib/dashboardPresentation'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const router = useRouter()
const authStore = useAuthStore()

const stats = ref({})
const chartLoading = ref(false)
const chartOption = ref(null)
const recentOrders = ref([])
const recentUsers = ref([])
const recentTickets = ref([])
const activeActivity = ref('orders')

const currentDate = currentDashboardDate()

const metricCards = computed(() => [
  {
    key: 'orders',
    label: '总订单数',
    value: stats.value.orders?.total || 0,
    detailLabel: '今日新增',
    detailValue: stats.value.orders?.today || 0,
    icon: ShoppingCart,
    tone: 'blue',
    path: '/orders'
  },
  {
    key: 'users',
    label: '总用户数',
    value: stats.value.users?.total || 0,
    detailLabel: '今日新增',
    detailValue: stats.value.users?.today || 0,
    icon: Users,
    tone: 'green',
    path: '/users'
  },
  {
    key: 'revenue',
    label: '总销售额',
    value: '¥' + formatNumber(stats.value.orders?.revenue || 0),
    detailLabel: '今日销售',
    detailValue: '¥' + formatNumber(stats.value.orders?.today_revenue || 0),
    icon: WalletCards,
    tone: 'amber',
    path: '/orders'
  },
  {
    key: 'tickets',
    label: '待处理工单',
    value: stats.value.tickets?.open || 0,
    detailLabel: '工单总数',
    detailValue: stats.value.tickets?.total || 0,
    icon: MessagesSquare,
    tone: 'coral',
    path: '/tickets'
  }
])

const quickActions = [
  { label: '添加商品', path: '/products', permission: 'product:create', icon: PackagePlus, tone: 'blue' },
  { label: '查看订单', path: '/orders', permission: 'order:view', icon: ShoppingCart, tone: 'green' },
  { label: '用户管理', path: '/users', permission: 'user:view', icon: Users, tone: 'amber' },
  { label: '客服对话', path: '/customer-service', permission: 'ticket:view', icon: Headset, tone: 'blue' },
  { label: '工单管理', path: '/tickets', permission: 'ticket:view', icon: MessagesSquare, tone: 'coral' },
  { label: '内容管理', path: '/content', permission: 'content:view', icon: FileText, tone: 'gray' },
  { label: '系统设置', path: '/settings', permission: 'settings:view', icon: Settings, tone: 'gray' }
]

const visibleQuickActions = computed(() =>
  quickActions.filter((action) => authStore.hasPermission(action.permission))
)

const navigateTo = (path) => router.push(path)

const notifyLoadFailure = () => {
  toast.error('部分仪表盘数据加载失败', { id: 'dashboard-load-error' })
}

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/dashboard/stats')
    stats.value = response.data
  } catch (error) {
    console.error('Failed to fetch stats:', error)
    notifyLoadFailure()
  }
}

const fetchSalesChart = async () => {
  chartLoading.value = true
  try {
    const response = await axios.get('/api/admin/dashboard/sales-chart')
    const data = response.data.data || []
    chartOption.value = buildSalesChartOption(data)
  } catch (error) {
    console.error('Failed to fetch sales chart:', error)
    notifyLoadFailure()
  } finally {
    chartLoading.value = false
  }
}

const fetchRecentOrders = async () => {
  try {
    const response = await axios.get('/api/admin/dashboard/recent-orders')
    if (!response.data || !Array.isArray(response.data.orders)) {
      throw new Error('[CRITICAL] Missing orders array in response')
    }
    recentOrders.value = response.data.orders
  } catch (error) {
    console.error('Failed to fetch recent orders:', error)
    notifyLoadFailure()
  }
}

const fetchRecentUsers = async () => {
  try {
    const response = await axios.get('/api/admin/dashboard/recent-users')
    if (!response.data || !Array.isArray(response.data.users)) {
      throw new Error('[CRITICAL] Missing users array in response')
    }
    recentUsers.value = response.data.users
  } catch (error) {
    console.error('Failed to fetch recent users:', error)
    notifyLoadFailure()
  }
}

const fetchRecentTickets = async () => {
  try {
    const response = await axios.get('/api/admin/dashboard/recent-tickets')
    if (!response.data || !Array.isArray(response.data.tickets)) {
      throw new Error('[CRITICAL] Missing tickets array in response')
    }
    recentTickets.value = response.data.tickets
  } catch (error) {
    console.error('Failed to fetch recent tickets:', error)
    notifyLoadFailure()
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
