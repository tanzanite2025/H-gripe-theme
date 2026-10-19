import { z } from 'zod'
import axios from '@/utils/axios'
import { requireApiObjectField, unwrapApiPayload } from '@/utils/apiResponse'

const capacityCount = z.number().int().nonnegative().max(Number.MAX_SAFE_INTEGER)
const capacityMetric = z.object({ objects: capacityCount, bytes: z.string().regex(/^\d+$/) })
const capacityReport = z.object({
  started_at: z.string().datetime(), completed_at: z.string().datetime(), physical_scan_complete: z.boolean(),
  locations_scanned: capacityCount, locations_incomplete: capacityCount,
  observed: capacityMetric, registered_originals: capacityMetric, registered_derivatives: capacityMetric,
  cleanup_receipt_only: capacityMetric, unmatched_registration: capacityMetric,
  with_cleanup_receipt: capacityMetric, with_upload_receipt: capacityMetric, with_soft_deleted_registration: capacityMetric,
  registered_not_observed: capacityMetric, registered_absence_unverified: capacityCount,
  registered_size_mismatches: capacityCount, conflicting_recorded_sizes: capacityCount,
  registered_unknown_recorded_size_objects: capacityCount,
  unobserved_upload_reservations: capacityMetric, unknown_upload_reservation_objects: capacityCount,
  unresolved_record_locations: capacityCount, issues: z.array(z.string()),
})
export type MediaPhysicalStorageCapacityAuditReport = z.infer<typeof capacityReport>

export async function requestReadOnlyMediaPhysicalStorageCapacityAudit(signal: AbortSignal): Promise<MediaPhysicalStorageCapacityAuditReport> {
  const path = '/api/admin/media/storage-capacity-audits'
  const payload = unwrapApiPayload(await axios.post(path, {}, { signal, timeout: 55000, suppressGlobalErrorToast: true }), path)
  return capacityReport.parse(requireApiObjectField(payload, 'report', path))
}
