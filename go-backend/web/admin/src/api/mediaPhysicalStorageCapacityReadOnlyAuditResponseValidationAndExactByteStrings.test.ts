import { beforeEach, expect, it, vi } from 'vitest'
import axios from '@/utils/axios'
import { requestReadOnlyMediaPhysicalStorageCapacityAudit } from './mediaPhysicalStorageCapacityReadOnlyAudit'
vi.mock('@/utils/axios', () => ({ default: { post: vi.fn() } }))
function createValidPhysicalCapacityAuditResponse() {
  const metric = { objects: 1, bytes: '9007199254740993' }
  return { started_at: '2026-10-17T00:00:00Z', completed_at: '2026-10-17T00:00:01Z', physical_scan_complete: false,
    locations_scanned: 1, locations_incomplete: 1, observed: metric, registered_originals: metric, registered_derivatives: metric,
    cleanup_receipt_only: metric, unmatched_registration: metric, with_cleanup_receipt: metric, with_upload_receipt: metric, with_soft_deleted_registration: metric,
    registered_not_observed: metric, registered_absence_unverified: 0, registered_size_mismatches: 0, conflicting_recorded_sizes: 0, registered_unknown_recorded_size_objects: 0,
    unobserved_upload_reservations: metric, unknown_upload_reservation_objects: 0, unresolved_record_locations: 0, issues: [] }
}
beforeEach(() => vi.resetAllMocks())
it('keeps byte strings exact and forwards cancellation with a bounded request timeout', async () => {
  const report = createValidPhysicalCapacityAuditResponse()
  vi.mocked(axios.post).mockResolvedValue({ data: { data: { report } } })
  const signal = new AbortController().signal
  const result = await requestReadOnlyMediaPhysicalStorageCapacityAudit(signal)
  expect(result.observed.bytes).toBe('9007199254740993')
  expect(result.physical_scan_complete).toBe(false)
  expect(axios.post).toHaveBeenCalledWith('/api/admin/media/storage-capacity-audits', {}, expect.objectContaining({ signal, timeout: 55000 }))
})
it.each(['missing_status', 'numeric_bytes', 'missing_metric', 'negative_count'])('rejects %s without presenting missing data as zero', async (problem) => {
  const report: Record<string, unknown> = createValidPhysicalCapacityAuditResponse()
  if (problem === 'missing_status') delete report.physical_scan_complete
  if (problem === 'numeric_bytes') report.observed = { objects: 1, bytes: 13 }
  if (problem === 'missing_metric') delete report.registered_originals
  if (problem === 'negative_count') report.locations_scanned = -1
  vi.mocked(axios.post).mockResolvedValue({ data: { data: { report } } })
  await expect(requestReadOnlyMediaPhysicalStorageCapacityAudit(new AbortController().signal)).rejects.toThrow()
})
