import type { Component } from 'vue'
import type { DashboardTone, SalesChartPoint } from '@/lib/dashboardPresentation'

export type DashboardActivity = 'orders' | 'users'
export type DashboardMetricToneClass = (tone?: string | null) => string
export type DashboardNumberFormatter = (value?: number | string | null) => string
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
  total_amount?: number | string | null
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
    revenue?: number | string | null
    today_revenue?: number | string | null
  }
  users?: {
    total?: number
    today?: number
  }
}

export interface DashboardSalesChartResponse {
  data?: SalesChartPoint[]
}
