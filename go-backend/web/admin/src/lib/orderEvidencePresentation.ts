import type { OrderEvidenceItem, OrderEvidenceItemStatus, OrderEvidenceItemType, OrderEvidencePackageStatus } from '@/modules/order/orderEvidenceTypes'
import type { OrderStatusTone } from '@/modules/order/orderTypes'

export const orderEvidencePackageStatusName = (status?: OrderEvidencePackageStatus | null): string => {
  switch (String(status || '').toLowerCase()) {
    case 'incomplete':
      return '待补全'
    case 'ready':
      return '可锁定'
    case 'locked':
      return '已锁定'
    case 'superseded':
      return '已修订'
    case '':
      return '未生成'
    default:
      return String(status)
  }
}

export const orderEvidencePackageStatusTone = (status?: OrderEvidencePackageStatus | null): OrderStatusTone => {
  switch (String(status || '').toLowerCase()) {
    case 'ready':
      return 'green'
    case 'locked':
      return 'blue'
    case 'incomplete':
      return 'amber'
    case 'superseded':
      return 'gray'
    default:
      return 'coral'
  }
}

export const orderEvidenceItemTypeName = (type?: OrderEvidenceItemType | null): string => {
  switch (String(type || '').toLowerCase()) {
    case 'configuration_confirmation':
      return '配置确认单'
    case 'product_identity':
      return '产品身份'
    case 'outbound_weight_packaging':
      return '出库称重 / 包装'
    case 'signed_pod':
      return '签收 / POD'
    case 'spoke_qc_tension':
      return '编轮张力表'
    default:
      return String(type || '证据项')
  }
}

export const orderEvidenceItemStatusName = (status?: OrderEvidenceItemStatus | null): string => {
  switch (String(status || '').toLowerCase()) {
    case 'missing':
      return '缺失'
    case 'draft':
      return '草稿'
    case 'complete':
      return '已完成'
    case 'waived':
      return '已豁免'
    default:
      return String(status || '未知')
  }
}

export const orderEvidenceItemStatusTone = (status?: OrderEvidenceItemStatus | null): OrderStatusTone => {
  switch (String(status || '').toLowerCase()) {
    case 'complete':
      return 'green'
    case 'waived':
      return 'gray'
    case 'draft':
      return 'amber'
    default:
      return 'coral'
  }
}

export const orderEvidenceTrackingAssociationName = (status?: string | null): string => {
  switch (String(status || '').toLowerCase()) {
    case 'matched':
      return '已匹配当前物流单'
    case 'mismatch':
      return '与当前物流单不一致'
    case 'not_recorded':
      return '人工 POD 未填写物流单号'
    case 'shipment_unavailable':
      return '当前暂无物流单'
    default:
      return String(status || '未关联')
  }
}

export const orderEvidenceTrackingAssociationTone = (status?: string | null): OrderStatusTone => {
  switch (String(status || '').toLowerCase()) {
    case 'matched':
      return 'green'
    case 'mismatch':
      return 'coral'
    case 'not_recorded':
    case 'shipment_unavailable':
      return 'amber'
    default:
      return 'gray'
  }
}

export const orderEvidenceRequiredReasonName = (reason?: string | null): string => {
  switch (String(reason || '').toLowerCase()) {
    case 'base':
      return '基础证据'
    case 'high_value':
      return '高价策略'
    case 'spoke_tension_qc':
      return '张力表规则'
    default:
      return String(reason || '规则')
  }
}

export const orderEvidenceItemIsReadOnly = (item?: OrderEvidenceItem | null): boolean => (
  item?.item_type === 'configuration_confirmation'
)

export const orderEvidenceItemScopeLabel = (item?: OrderEvidenceItem | null): string => (
  item?.order_item_id ? `订单行 #${item.order_item_id}` : '订单级'
)

export const orderEvidenceCustomerName = (summary: {
  customer_first_name?: string | null
  customer_last_name?: string | null
}): string => (
  [summary.customer_first_name, summary.customer_last_name]
    .map((value) => String(value || '').trim())
    .filter(Boolean)
    .join(' ') || '未填写姓名'
)

export const orderEvidencePercent = (summary: {
  total_evidence_items?: number
  complete_evidence_items?: number
  waived_evidence_items?: number
}): number => {
  const total = Number(summary.total_evidence_items || 0)
  if (total <= 0) return 0
  const satisfied = Number(summary.complete_evidence_items || 0) + Number(summary.waived_evidence_items || 0)
  return Math.round((satisfied / total) * 100)
}
