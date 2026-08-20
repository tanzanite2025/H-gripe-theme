import type { EChartsOption } from 'echarts'

export type DashboardTone = 'blue' | 'green' | 'amber' | 'coral' | 'gray'

export const adminChartFontFamily = 'MapleUICJK'
const adminChartTextStyle = { fontFamily: adminChartFontFamily } as const

export interface SalesChartPoint {
  date: string
  count: number
  amount: number
}

export const currentDashboardDate = (): string => new Intl.DateTimeFormat('zh-CN', {
  year: 'numeric',
  month: 'long',
  day: 'numeric',
  weekday: 'long'
}).format(new Date())

export const metricToneClass = (tone?: string | null): string => {
  const classes: Record<DashboardTone, string> = {
    blue: 'bg-blue-50 text-blue-700',
    green: 'bg-emerald-50 text-emerald-700',
    amber: 'bg-amber-50 text-amber-700',
    coral: 'bg-rose-50 text-rose-700',
    gray: 'bg-muted text-muted-foreground'
  }
  return classes[tone as DashboardTone] || classes.gray
}

export const formatNumber = (value?: number | string | null): string => Number(value).toLocaleString('zh-CN', {
  minimumFractionDigits: 2,
  maximumFractionDigits: 2
})

export const getRoleName = (role?: string | null): string => ({
  admin: '管理员',
  manager: '经理',
  editor: '编辑',
  support: '客服',
  viewer: '查看者'
})[role || ''] || role || '-'

export const roleTone = (role?: string | null): DashboardTone => ({
  admin: 'coral',
  manager: 'amber',
  editor: 'green',
  support: 'blue',
  viewer: 'gray'
} as Record<string, DashboardTone>)[role || ''] || 'gray'

export const getOrderStatusName = (status?: string | null): string => ({
  pending: '待付款',
  paid: '已付款',
  processing: '处理中',
  shipped: '已发货',
  completed: '已完成',
  payment_expired: '支付超时',
  cancelled: '已取消'
})[status || ''] || status || '-'

export const orderStatusTone = (status?: string | null): DashboardTone => ({
  pending: 'amber',
  paid: 'green',
  processing: 'amber',
  shipped: 'blue',
  completed: 'green',
  payment_expired: 'amber',
  cancelled: 'coral'
} as Record<string, DashboardTone>)[status || ''] || 'gray'

export const buildSalesChartOption = (data?: SalesChartPoint[] | null): EChartsOption | null => {
  if (!Array.isArray(data) || data.length === 0) return null

  return {
    color: ['#2563eb', '#16803c'],
    textStyle: adminChartTextStyle,
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#182230',
      borderWidth: 0,
      textStyle: { ...adminChartTextStyle, color: '#ffffff' }
    },
    legend: {
      top: 0,
      right: 0,
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { ...adminChartTextStyle, color: '#667085' },
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
      axisLabel: { ...adminChartTextStyle, color: '#667085' }
    },
    yAxis: [
      {
        type: 'value',
        name: '订单数',
        nameTextStyle: { ...adminChartTextStyle, color: '#667085' },
        splitLine: { lineStyle: { color: '#eaecf0' } },
        axisLabel: { ...adminChartTextStyle, color: '#667085' }
      },
      {
        type: 'value',
        name: '销售额',
        nameTextStyle: { ...adminChartTextStyle, color: '#667085' },
        splitLine: { show: false },
        axisLabel: { ...adminChartTextStyle, color: '#667085' }
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
}
