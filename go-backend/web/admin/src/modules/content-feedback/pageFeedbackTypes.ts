import type {
  PageFeedbackItem,
  PageFeedbackPagination,
  PageFeedbackRiskLevel,
  PageFeedbackStatus,
  PageFeedbackStatusFilter,
} from '@/api/pageFeedback'

export interface PageFeedbackFiltersState {
  status: PageFeedbackStatusFilter
  pagePath: string
  threadKey: string
  search: string
}

export const createDefaultPageFeedbackFilters = (): PageFeedbackFiltersState => ({
  status: 'pending',
  pagePath: '',
  threadKey: '',
  search: '',
})

export const displayPageTitle = (item: PageFeedbackItem): string => (
  item.page_title || item.page_path || item.thread_key || '未知页面'
)

export const displayPagePath = (item: PageFeedbackItem): string => (
  item.page_path || `thread:${item.thread_key}`
)

export const displaySourceHashPreview = (value?: string | null): string => {
  const normalized = value?.trim()
  return normalized ? `#${normalized}` : '未记录'
}

export const formatPageFeedbackDate = (value?: string | null): string =>
  value ? new Date(value).toLocaleString('zh-CN') : '-'

export const pageFeedbackStatusLabel = (value: PageFeedbackStatus): string => ({
  pending: '待处理',
  approved: '已发布',
  rejected: '已拒绝',
  hidden: '已隐藏',
}[value] || value)

export const pageFeedbackStatusBadgeVariant = (
  value: PageFeedbackStatus,
): 'default' | 'secondary' | 'destructive' | 'outline' => {
  if (value === 'approved') return 'default'
  if (value === 'rejected') return 'destructive'
  if (value === 'hidden') return 'outline'
  return 'secondary'
}

export const pageFeedbackRiskLabel = (value: PageFeedbackRiskLevel): string => ({
  normal: '正常',
  warning: '需要关注',
  critical: '高风险',
}[value] || value)

export const pageFeedbackRiskTone = (
  value: PageFeedbackRiskLevel,
): 'green' | 'amber' | 'coral' => {
  if (value === 'critical') return 'coral'
  if (value === 'warning') return 'amber'
  return 'green'
}

export type {
  PageFeedbackItem,
  PageFeedbackPagination,
  PageFeedbackRiskLevel,
  PageFeedbackStatus,
  PageFeedbackStatusFilter,
}
