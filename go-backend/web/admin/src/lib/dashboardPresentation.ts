import type { EChartsOption } from 'echarts'

export type DashboardTone = 'blue' | 'green' | 'amber' | 'coral' | 'gray'

export const adminChartFontFamily = 'MapleUICJK'
const adminChartTextStyle = { fontFamily: adminChartFontFamily } as const

export interface SalesChartPoint {
  date: string
  count: number
  revenue_by_currency: RevenueByCurrency[]
}

export interface RevenueByCurrency {
  currency: string
  amount_minor: number | string
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

const currencyMinorUnits: Record<string, number> = {
  BHD: 3, IQD: 3, JOD: 3, KWD: 3, LYD: 3, OMR: 3, TND: 3,
  BIF: 0, CLP: 0, DJF: 0, GNF: 0, ISK: 0, JPY: 0, KMF: 0, KRW: 0,
  PYG: 0, RWF: 0, UGX: 0, UYI: 0, VND: 0, VUV: 0, XAF: 0, XOF: 0, XPF: 0
}

export const minorUnitsForCurrency = (currency?: string | null): number => (
  currencyMinorUnits[String(currency || '').trim().toUpperCase()] ?? 2
)

const parseMinorAmount = (value: number | string | bigint | null | undefined): bigint => {
  if (typeof value === 'bigint') return value
  if (typeof value === 'number') return Number.isFinite(value) ? BigInt(Math.trunc(value)) : 0n
  const text = String(value ?? '').trim()
  return /^[-+]?\d+$/.test(text) ? BigInt(text) : 0n
}

/** Format an amount represented in the currency's smallest unit. */
export const formatMinorMoney = (
  amountMinor?: number | string | bigint | null,
  currency?: string | null,
): string => {
  const code = String(currency || '').trim().toUpperCase()
  const minorUnits = minorUnitsForCurrency(code)
  const minor = parseMinorAmount(amountMinor)
  const negative = minor < 0n
  const absolute = negative ? -minor : minor
  const scale = 10n ** BigInt(minorUnits)
  const major = absolute / scale
  const fraction = absolute % scale
  const fractionText = minorUnits > 0
    ? fraction.toString().padStart(minorUnits, '0')
    : ''

  if (!code) return `${negative ? '-' : ''}${major.toString()}${minorUnits > 0 ? `.${fractionText}` : ''}`

  try {
    const formatter = new Intl.NumberFormat('zh-CN', {
      style: 'currency',
      currency: code,
      minimumFractionDigits: minorUnits,
      maximumFractionDigits: minorUnits,
    })
    // Intl can group a bigint without converting through an imprecise float.
    const groupedMajor = new Intl.NumberFormat('zh-CN', {
      useGrouping: true,
      maximumFractionDigits: 0,
    }).format(negative ? -major : major)
    const parts = formatter.formatToParts(negative ? -1 : 1)
    return parts.map((part) => {
      if (part.type === 'integer') return groupedMajor.replace(/^-/, '')
      if (part.type === 'fraction') return fractionText
      return part.value
    }).join('')
  } catch {
    return `${code} ${negative ? '-' : ''}${major.toString()}${minorUnits > 0 ? `.${fractionText}` : ''}`
  }
}

export const formatRevenueByCurrency = (values?: RevenueByCurrency[] | null): string =>
  (values || []).map((item) => formatMinorMoney(item.amount_minor, item.currency)).join(' · ') || '-'

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

  const currencies = Array.from(new Set(
    data.flatMap((item) => (item.revenue_by_currency || [])
      .map((revenue) => String(revenue.currency || '').trim().toUpperCase())
      .filter(Boolean)),
  ))
  const revenueSeriesName = (currency: string): string => `销售额 (${currency})`
  const series = [
    {
      name: '订单数',
      type: 'line' as const,
      data: data.map((item) => item.count),
      smooth: true,
      symbolSize: 7,
      lineStyle: { width: 3 }
    },
    ...currencies.map((currency) => ({
      name: revenueSeriesName(currency),
      type: 'line' as const,
      yAxisIndex: 1,
      data: data.map((item) => {
        const revenue = (item.revenue_by_currency || []).find((value) => String(value.currency || '').trim().toUpperCase() === currency)
        const minorUnits = minorUnitsForCurrency(currency)
        return Number(revenue?.amount_minor || 0) / (10 ** minorUnits)
      }),
      smooth: true,
      symbolSize: 7,
      lineStyle: { width: 3 }
    }))
  ]

  return {
    color: ['#2563eb', '#16803c', '#d97706', '#7c3aed', '#dc2626', '#0891b2'],
    textStyle: adminChartTextStyle,
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#182230',
      borderWidth: 0,
      textStyle: { ...adminChartTextStyle, color: '#ffffff' },
      formatter: (params: any) => {
        const entries = Array.isArray(params) ? params : [params]
        return entries.map((entry: any) => {
          const dataIndex = Number(entry?.dataIndex)
          const point = Number.isInteger(dataIndex) ? data[dataIndex] : undefined
          if (entry?.seriesName === '订单数') {
            return `${entry?.marker || ''}订单数: ${formatNumber(point?.count || 0)}`
          }
          const match = String(entry?.seriesName || '').match(/\(([^)]+)\)$/)
          const currency = match?.[1] || ''
          const revenue = point?.revenue_by_currency?.find((value) => (
            String(value.currency || '').trim().toUpperCase() === currency
          ))
          return `${entry?.marker || ''}${entry?.seriesName || '销售额'}: ${formatMinorMoney(revenue?.amount_minor, currency)}`
        }).join('<br/>')
      }
    },
    legend: {
      top: 0,
      right: 0,
      itemWidth: 10,
      itemHeight: 10,
      textStyle: { ...adminChartTextStyle, color: '#667085' },
      data: ['订单数', ...currencies.map(revenueSeriesName)]
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
    series
  }
}
