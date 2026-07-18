<template>
  <div class="space-y-4">
    <header class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        <h1 class="text-2xl font-semibold tracking-normal">仪表板</h1>
        <p class="mt-1 text-sm text-muted-foreground">今日经营与服务概览</p>
      </div>
      <div class="flex items-center gap-1.5 text-xs text-muted-foreground">
        <CalendarDays class="size-3.5" />
        <span>{{ currentDate }}</span>
      </div>
    </header>

    <section class="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4" aria-label="关键指标">
      <button
        v-for="metric in metricCards"
        :key="metric.key"
        type="button"
        class="group flex min-h-31 flex-col justify-between rounded-lg border bg-card p-4 text-left text-card-foreground shadow-xs transition-[border-color,box-shadow,transform] hover:-translate-y-px hover:border-foreground/15 hover:shadow-sm focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/30"
        @click="navigateTo(metric.path)"
      >
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0">
            <span class="block text-xs font-medium text-muted-foreground">{{ metric.label }}</span>
            <strong class="mt-1.5 block truncate text-2xl font-semibold tabular-nums">{{ metric.value }}</strong>
          </div>
          <span class="flex size-9 shrink-0 items-center justify-center rounded-lg" :class="metricToneClass(metric.tone)">
            <component :is="metric.icon" class="size-4" />
          </span>
        </div>
        <div class="flex items-center justify-between gap-2 text-xs text-muted-foreground">
          <span>{{ metric.detailLabel }}</span>
          <strong class="font-medium tabular-nums text-foreground/75">{{ metric.detailValue }}</strong>
        </div>
      </button>
    </section>

    <section class="grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(280px,0.8fr)]">
      <Card class="min-w-0 gap-0 py-0 shadow-none">
        <CardHeader class="flex flex-row items-center justify-between border-b py-4">
          <div>
            <CardTitle class="text-sm font-semibold">销售趋势</CardTitle>
            <CardDescription class="mt-1 text-xs">最近 30 天</CardDescription>
          </div>
          <Tooltip>
            <TooltipTrigger as-child>
              <Button
                variant="outline"
                size="icon"
                aria-label="刷新销售趋势"
                :disabled="chartLoading"
                @click="fetchSalesChart"
              >
                <RefreshCw class="size-3.5" :class="chartLoading ? 'animate-spin' : ''" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>刷新销售趋势</TooltipContent>
          </Tooltip>
        </CardHeader>
        <CardContent class="flex h-80 items-center justify-center p-4">
          <div v-if="chartLoading" class="w-full space-y-4">
            <Skeleton class="h-4 w-36" />
            <Skeleton class="h-56 w-full" />
          </div>
          <v-chart v-else-if="chartOption" class="h-full w-full" :option="chartOption" autoresize />
          <div v-else class="flex flex-col items-center text-center text-muted-foreground">
            <ChartNoAxesCombined class="mb-3 size-8 opacity-55" />
            <p class="text-sm font-medium text-foreground/75">暂无销售数据</p>
          </div>
        </CardContent>
      </Card>

      <Card class="min-w-0 gap-0 py-0 shadow-none">
        <CardHeader class="border-b py-4">
          <CardTitle class="text-sm font-semibold">快速操作</CardTitle>
          <CardDescription class="text-xs">常用管理入口</CardDescription>
        </CardHeader>
        <CardContent class="grid grid-cols-1 gap-1 p-2 sm:grid-cols-2 xl:grid-cols-1">
          <Button
            v-for="action in visibleQuickActions"
            :key="action.path"
            variant="ghost"
            class="h-10 w-full justify-start gap-2.5 px-2.5"
            @click="navigateTo(action.path)"
          >
            <span class="flex size-7 shrink-0 items-center justify-center rounded-md" :class="metricToneClass(action.tone)">
              <component :is="action.icon" class="size-3.5" />
            </span>
            <span class="truncate">{{ action.label }}</span>
            <ArrowRight class="ml-auto size-3.5 text-muted-foreground" />
          </Button>
        </CardContent>
      </Card>
    </section>

    <Card class="gap-0 py-0 shadow-none">
      <Tabs v-model="activeActivity">
        <CardHeader class="flex flex-col gap-3 border-b py-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <CardTitle class="text-sm font-semibold">最近活动</CardTitle>
            <CardDescription class="mt-1 text-xs">最新业务记录</CardDescription>
          </div>
          <TabsList variant="line">
            <TabsTrigger value="orders">订单</TabsTrigger>
            <TabsTrigger value="users">用户</TabsTrigger>
            <TabsTrigger value="tickets">工单</TabsTrigger>
          </TabsList>
        </CardHeader>

        <CardContent class="pb-4 pt-2">
          <TabsContent value="orders" class="mt-0">
            <div class="flex min-h-9 items-center justify-between">
              <strong class="text-xs font-medium text-foreground/75">最近订单</strong>
              <Button variant="link" size="sm" class="px-0" @click="navigateTo('/orders')">
                查看全部
                <ArrowRight class="size-3.5" />
              </Button>
            </div>
            <EmptyActivity v-if="recentOrders.length === 0" label="暂无订单" />
            <div v-else class="grid grid-cols-1 gap-x-6 md:grid-cols-2">
              <div v-for="order in recentOrders" :key="order.id" class="flex min-w-0 items-center justify-between gap-4 border-b py-3">
                <div class="min-w-0">
                  <strong class="block truncate text-xs font-medium">#{{ order.order_number }}</strong>
                  <span class="mt-1 block truncate text-xs text-muted-foreground">¥{{ formatNumber(order.total_amount) }}</span>
                </div>
                <Badge variant="outline" :class="orderStatusClass(order.status)">
                  {{ getOrderStatusName(order.status) }}
                </Badge>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="users" class="mt-0">
            <div class="flex min-h-9 items-center justify-between">
              <strong class="text-xs font-medium text-foreground/75">最近用户</strong>
              <Button variant="link" size="sm" class="px-0" @click="navigateTo('/users')">
                查看全部
                <ArrowRight class="size-3.5" />
              </Button>
            </div>
            <EmptyActivity v-if="recentUsers.length === 0" label="暂无用户" />
            <div v-else class="grid grid-cols-1 gap-x-6 md:grid-cols-2">
              <div v-for="recentUser in recentUsers" :key="recentUser.id" class="flex min-w-0 items-center justify-between gap-4 border-b py-3">
                <div class="min-w-0">
                  <strong class="block truncate text-xs font-medium">{{ recentUser.username }}</strong>
                  <span class="mt-1 block truncate text-xs text-muted-foreground">{{ recentUser.email }}</span>
                </div>
                <Badge variant="outline" :class="roleStatusClass(recentUser.role)">
                  {{ getRoleName(recentUser.role) }}
                </Badge>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="tickets" class="mt-0">
            <div class="flex min-h-9 items-center justify-between">
              <strong class="text-xs font-medium text-foreground/75">最近工单</strong>
              <Button variant="link" size="sm" class="px-0" @click="navigateTo('/tickets')">
                查看全部
                <ArrowRight class="size-3.5" />
              </Button>
            </div>
            <EmptyActivity v-if="recentTickets.length === 0" label="暂无工单" />
            <div v-else class="grid grid-cols-1 gap-x-6 md:grid-cols-2">
              <div v-for="ticket in recentTickets" :key="ticket.id" class="flex min-w-0 items-center justify-between gap-4 border-b py-3">
                <div class="min-w-0">
                  <strong class="block truncate text-xs font-medium">{{ ticket.subject }}</strong>
                  <span class="mt-1 block truncate text-xs text-muted-foreground">{{ ticket.category }}</span>
                </div>
                <Badge variant="outline" :class="ticketStatusClass(ticket.status)">
                  {{ getTicketStatusName(ticket.status) }}
                </Badge>
              </div>
            </div>
          </TabsContent>
        </CardContent>
      </Tabs>
    </Card>
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import {
  ArrowRight,
  CalendarDays,
  ChartNoAxesCombined,
  FileText,
  Inbox,
  MessagesSquare,
  PackagePlus,
  RefreshCw,
  Settings,
  ShoppingCart,
  Users,
  WalletCards
} from '@lucide/vue'
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent])

const EmptyActivity = defineComponent({
  props: {
    label: {
      type: String,
      required: true
    }
  },
  setup(props) {
    return () => h('div', { class: 'flex min-h-36 flex-col items-center justify-center text-muted-foreground' }, [
      h(Inbox, { class: 'mb-2 size-6 opacity-55' }),
      h('span', { class: 'text-xs' }, props.label)
    ])
  }
})

const router = useRouter()
const authStore = useAuthStore()

const stats = ref({})
const chartLoading = ref(false)
const chartOption = ref(null)
const recentOrders = ref([])
const recentUsers = ref([])
const recentTickets = ref([])
const activeActivity = ref('orders')

const currentDate = new Intl.DateTimeFormat('zh-CN', {
  year: 'numeric',
  month: 'long',
  day: 'numeric',
  weekday: 'long'
}).format(new Date())

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
  { label: '工单管理', path: '/tickets', permission: 'ticket:view', icon: MessagesSquare, tone: 'coral' },
  { label: '内容管理', path: '/content', permission: 'content:view', icon: FileText, tone: 'gray' },
  { label: '系统设置', path: '/settings', permission: 'settings:view', icon: Settings, tone: 'gray' }
]

const visibleQuickActions = computed(() =>
  quickActions.filter((action) => authStore.hasPermission(action.permission))
)

const metricToneClass = (tone) => {
  const classes = {
    blue: 'bg-blue-50 text-blue-700',
    green: 'bg-emerald-50 text-emerald-700',
    amber: 'bg-amber-50 text-amber-700',
    coral: 'bg-rose-50 text-rose-700',
    gray: 'bg-muted text-muted-foreground'
  }
  return classes[tone] || classes.gray
}

function formatNumber(value) {
  return Number(value).toLocaleString('zh-CN', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  })
}

const navigateTo = (path) => router.push(path)

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

const roleStatusClass = (role) => {
  const classes = {
    admin: 'border-rose-200 bg-rose-50 text-rose-700',
    manager: 'border-amber-200 bg-amber-50 text-amber-700',
    editor: 'border-emerald-200 bg-emerald-50 text-emerald-700',
    support: 'border-blue-200 bg-blue-50 text-blue-700',
    viewer: 'border-border bg-muted text-muted-foreground'
  }
  return classes[role] || classes.viewer
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

const orderStatusClass = (status) => {
  const classes = {
    pending: 'border-amber-200 bg-amber-50 text-amber-700',
    paid: 'border-emerald-200 bg-emerald-50 text-emerald-700',
    shipped: 'border-blue-200 bg-blue-50 text-blue-700',
    completed: 'border-border bg-muted text-muted-foreground',
    cancelled: 'border-rose-200 bg-rose-50 text-rose-700'
  }
  return classes[status] || classes.completed
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

const ticketStatusClass = (status) => {
  const classes = {
    open: 'border-rose-200 bg-rose-50 text-rose-700',
    pending: 'border-amber-200 bg-amber-50 text-amber-700',
    resolved: 'border-emerald-200 bg-emerald-50 text-emerald-700',
    closed: 'border-border bg-muted text-muted-foreground'
  }
  return classes[status] || classes.closed
}

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

    if (data.length === 0) {
      chartOption.value = null
      return
    }

    chartOption.value = {
      color: ['#2563eb', '#16803c'],
      tooltip: {
        trigger: 'axis',
        backgroundColor: '#182230',
        borderWidth: 0,
        textStyle: { color: '#ffffff' }
      },
      legend: {
        top: 0,
        right: 0,
        itemWidth: 10,
        itemHeight: 10,
        textStyle: { color: '#667085' },
        data: ['订单数', '销售额']
      },
      grid: {
        top: 44,
        right: 24,
        bottom: 16,
        left: 12,
        containLabel: true
      },
      xAxis: {
        type: 'category',
        boundaryGap: false,
        data: data.map((item) => item.date),
        axisLine: { lineStyle: { color: '#e4e7ec' } },
        axisTick: { show: false },
        axisLabel: { color: '#667085' }
      },
      yAxis: [
        {
          type: 'value',
          name: '订单数',
          nameTextStyle: { color: '#667085' },
          splitLine: { lineStyle: { color: '#eaecf0' } },
          axisLabel: { color: '#667085' }
        },
        {
          type: 'value',
          name: '销售额',
          nameTextStyle: { color: '#667085' },
          splitLine: { show: false },
          axisLabel: { color: '#667085' }
        }
      ],
      series: [
        {
          name: '订单数',
          type: 'line',
          data: data.map((item) => item.count),
          smooth: true,
          symbolSize: 7,
          lineStyle: { width: 3 }
        },
        {
          name: '销售额',
          type: 'line',
          yAxisIndex: 1,
          data: data.map((item) => item.amount),
          smooth: true,
          symbolSize: 7,
          lineStyle: { width: 3 }
        }
      ]
    }
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
