type AfterSalesStatusClass = 'status-gray' | 'status-green' | 'status-amber' | 'status-blue' | 'status-coral'

interface AfterSalesOption {
  label: string
  value: string
}

export const afterSalesStatusOptions: AfterSalesOption[] = [
  { value: 'requested', label: '申请中' },
  { value: 'reviewing', label: '审核中' },
  { value: 'approved', label: '已批准' },
  { value: 'awaiting_return', label: '等待寄回' },
  { value: 'return_in_transit', label: '退回运输中' },
  { value: 'received', label: '已收货' },
  { value: 'inspecting', label: '检验中' },
  { value: 'resolving', label: '处理中' },
  { value: 'completed', label: '已完成' },
  { value: 'rejected', label: '已拒绝' },
  { value: 'cancelled', label: '已取消' },
  { value: 'exception', label: '异常' },
]

export const afterSalesStatusFilterOptions: AfterSalesOption[] = [
  { value: 'all', label: '全部状态' },
  ...afterSalesStatusOptions,
]

export const afterSalesTypeOptions: AfterSalesOption[] = [
  { value: 'return_refund', label: '退货退款' },
  { value: 'exchange', label: '换货' },
  { value: 'refund_only', label: '仅退款' },
  { value: 'reshipment', label: '补发' },
  { value: 'customer_request', label: '客户申请' },
]

export const afterSalesTypeFilterOptions: AfterSalesOption[] = [
  { value: 'all', label: '全部类型' },
  ...afterSalesTypeOptions,
]

const afterSalesTransitionMap: Record<string, string[]> = {
  requested: ['reviewing', 'cancelled'],
  reviewing: ['approved', 'rejected', 'cancelled'],
  approved: ['awaiting_return', 'resolving', 'cancelled'],
  awaiting_return: ['return_in_transit', 'cancelled'],
  return_in_transit: ['received', 'exception', 'cancelled'],
  received: ['inspecting', 'exception'],
  inspecting: ['resolving', 'exception'],
  resolving: ['completed', 'exception'],
  exception: ['reviewing', 'cancelled'],
}

export const getAfterSalesStatusName = (status?: string | null): string =>
  afterSalesStatusOptions.find((option) => option.value === status)?.label || status || '-'

export const getAfterSalesTypeName = (type?: string | null): string =>
  afterSalesTypeOptions.find((option) => option.value === type)?.label || type || '-'

export const getAfterSalesNextStatuses = (status?: string | null): string[] => [
  ...(afterSalesTransitionMap[status || ''] || []),
]

export const afterSalesStatusClass = (status?: string | null): AfterSalesStatusClass =>
  ({
    requested: 'status-gray',
    reviewing: 'status-amber',
    approved: 'status-green',
    awaiting_return: 'status-blue',
    return_in_transit: 'status-blue',
    received: 'status-green',
    inspecting: 'status-amber',
    resolving: 'status-amber',
    completed: 'status-green',
    rejected: 'status-coral',
    cancelled: 'status-gray',
    exception: 'status-coral',
  } as Record<string, AfterSalesStatusClass>)[status || ''] || 'status-gray'

export const isAfterSalesRefundType = (type?: string | null): boolean =>
  type === 'return_refund' || type === 'refund_only'
