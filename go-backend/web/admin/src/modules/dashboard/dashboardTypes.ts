import type { Component } from 'vue'
import type { DashboardTone, RevenueByCurrency, SalesChartPoint } from '@/lib/dashboardPresentation'

export type DashboardActivity = 'orders' | 'users'
export type DashboardMetricToneClass = (tone?: string | null) => string
export type DashboardMoneyFormatter = (
  amountMinor?: number | string | bigint | null,
  currency?: string | null,
) => string
export type DashboardLabelResolver = (value?: string | null) => string
export type DashboardToneResolver = (value?: string | null) => DashboardTone

export interface DashboardMetricCard {
  key: string
  label: string
  value: string | number
  detailLabel: string
  detailValue: string | number
  icon: Component
  tone: DashboardTone
  path: string
}

export interface DashboardQuickAction {
  label: string
  path: string
  permission: string
  icon: Component
  tone: DashboardTone
}

export interface DashboardRecentOrder {
  id: number | string
  order_number?: string
  total_amount_minor: number | string
  currency: string
  status?: string | null
}

export interface DashboardRecentUser {
  id: number | string
  username?: string
  email?: string
  role?: string | null
}

export interface DashboardStats {
  orders?: {
    total?: number
    today?: number
    revenue_by_currency?: RevenueByCurrency[] | null
    today_revenue_by_currency?: RevenueByCurrency[] | null
  }
  users?: {
    total?: number
    today?: number
  }
}

export interface DashboardSalesChartResponse {
  data?: SalesChartPoint[]
}
