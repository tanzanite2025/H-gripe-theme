import axios from '@/utils/axios'
import { requireApiObject, requireApiObjectField, unwrapApiPayload } from '@/utils/apiResponse'
import type { MediaAssetListParams, MediaLibraryPartition } from '@/api/media'

export const mediaUsageSummaryFilterOptions = [
  { value: 'all', label: '全部核对资源' },
  { value: 'no_recognized_references', label: '未识别到引用' },
  { value: 'cross_business', label: '跨业务使用' },
  { value: 'retained', label: '保留与待绑定' },
  { value: 'anomalous', label: '异常关联' },
  { value: 'restricted', label: '含受限引用' },
] as const
export type MediaUsageSummaryFilter = typeof mediaUsageSummaryFilterOptions[number]['value']
export type MediaUsageSummaryParams = Omit<MediaAssetListParams, 'page' | 'page_size'>
export interface MediaUsageSummaryReference {
  category: string
  resource_type: string
  resource_id: string
  parent_resource_id: string
  label: string
  field: string
  usage_state: string
}
export interface MediaUsageSummaryRow {
  asset_id: string
  title: string
  library_partition: MediaLibraryPartition
  no_recognized_references: boolean
  has_restricted_references: boolean
  pending_business_draft: boolean
  visible_reference_total: number
  cross_business: boolean
  retained: boolean
  anomalous: boolean
  categories: string[]
  references: MediaUsageSummaryReference[]
  reference_preview_limited: boolean
}
export interface MediaUsageSummaryReport {
  coverage: 'recognized_business_sources_only'
  snapshot_started_at: string
  completed_at: string
  candidate_limit: number
  counts: Record<MediaUsageSummaryFilter, number>
  rows: MediaUsageSummaryRow[]
}

export function mediaUsageSummaryRowMatchesVisibleFilter(row: MediaUsageSummaryRow, filter: MediaUsageSummaryFilter): boolean {
  if (filter === 'all') return true
  if (filter === 'restricted') return row.has_restricted_references
  return row[filter]
}

export function validateReadOnlyMediaUsageSummaryReport(value: unknown): MediaUsageSummaryReport {
  const invalid = (): never => { throw new Error('Invalid media usage summary; no conclusion available') }
  const report = requireApiObject(value, 'media usage summary') as unknown as MediaUsageSummaryReport
  if (report.coverage !== 'recognized_business_sources_only' || report.candidate_limit !== 200 || !Array.isArray(report.rows) || report.rows.length > 200) invalid()
  if (typeof report.snapshot_started_at !== 'string' || typeof report.completed_at !== 'string' ||
    !Number.isFinite(Date.parse(report.snapshot_started_at)) || !Number.isFinite(Date.parse(report.completed_at)) ||
    Date.parse(report.completed_at) < Date.parse(report.snapshot_started_at)) invalid()
  const counts = requireApiObject(report.counts, 'media usage summary counts')
  const identities = new Set<string>()
  for (const row of report.rows) {
    if (!row || typeof row.asset_id !== 'string' || !/^[1-9]\d*$/.test(row.asset_id) || identities.has(row.asset_id)) invalid()
    identities.add(row.asset_id)
    if (typeof row.title !== 'string' || !['catalog', 'content', 'shared', 'customer_service', 'customer_submissions', 'unclassified'].includes(row.library_partition)) invalid()
    for (const key of ['no_recognized_references', 'has_restricted_references', 'pending_business_draft', 'cross_business', 'retained', 'anomalous', 'reference_preview_limited'] as const) {
      if (typeof row[key] !== 'boolean') invalid()
    }
    if (!Number.isSafeInteger(row.visible_reference_total) || row.visible_reference_total < 0 || !Array.isArray(row.categories) ||
      row.categories.some((category) => typeof category !== 'string') || !Array.isArray(row.references) || row.references.length > 20 ||
      row.references.length !== Math.min(row.visible_reference_total, 20) || row.reference_preview_limited !== (row.visible_reference_total > 20)) invalid()
    if (row.no_recognized_references !== (row.visible_reference_total === 0 && !row.has_restricted_references && !row.pending_business_draft)) invalid()
    if (row.pending_business_draft && !row.retained) invalid()
    if (row.no_recognized_references && (row.cross_business || row.retained || row.anomalous || row.categories.length)) invalid()
    for (const reference of row.references) {
      if (!reference || ['category', 'resource_type', 'label', 'field', 'usage_state'].some((key) => typeof reference[key as keyof MediaUsageSummaryReference] !== 'string') ||
        typeof reference.resource_id !== 'string' || !/^\d+$/.test(reference.resource_id) ||
        typeof reference.parent_resource_id !== 'string' || !/^\d+$/.test(reference.parent_resource_id)) invalid()
    }
  }
  for (const { value: filter } of mediaUsageSummaryFilterOptions) {
    if (!Number.isSafeInteger(counts[filter]) || counts[filter] !== report.rows.filter((row) => mediaUsageSummaryRowMatchesVisibleFilter(row, filter)).length) invalid()
  }
  return report
}

export async function requestReadOnlyMediaUsageSummary(params: MediaUsageSummaryParams, signal?: AbortSignal): Promise<MediaUsageSummaryReport> {
  const path = '/api/admin/media/usage-summaries'
  const payload = requireApiObject(unwrapApiPayload(await axios.post(path, null, { params, signal, suppressGlobalErrorToast: true }), path), path)
  return validateReadOnlyMediaUsageSummaryReport(requireApiObjectField(payload, 'report', path))
}
