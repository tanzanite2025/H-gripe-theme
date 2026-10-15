import { z } from 'zod'
import axios from '@/utils/axios'

const outboxFailureCountSchema = z.number().int().nonnegative().safe()
const dashboardBusinessFollowUpItemSchema = z.object({
  id: z.number().int().positive().safe(),
  title: z.string().min(1),
  category: z.string().min(1),
  business_impact: z.string().min(1),
  responsible_team: z.string().min(1),
  sales_next_steps: z.array(z.string().min(1)).min(1),
  reason_summary: z.string().min(1),
  processing_state: z.enum(['automatic_retry_scheduled', 'manual_attention']),
  attempts: outboxFailureCountSchema,
  max_attempts: z.number().int().positive().safe(),
  created_at: z.string().datetime({ offset: true }),
  updated_at: z.string().datetime({ offset: true }),
  last_attempt_at: z.string().datetime({ offset: true }).nullable(),
  next_retry_at: z.string().datetime({ offset: true }).nullable(),
  related_record: z.object({
    label: z.string().min(1),
    path: z.string().refine((path) => path === '' || (path.startsWith('/') && !path.startsWith('//')), 'Related record must use a local route'),
    required_permission: z.string(),
  }).nullable(),
  related_record_notice: z.string(),
}).refine((item) => item.processing_state === 'automatic_retry_scheduled'
  ? item.next_retry_at !== null && item.attempts < item.max_attempts
  : item.next_retry_at === null, 'Retry state and schedule must be consistent')

const dashboardOutboxFailuresSchema = z.discriminatedUnion('status', [
  z.object({
    status: z.literal('ready'),
    total_count: outboxFailureCountSchema,
    failed_count: outboxFailureCountSchema,
    dead_letter_count: outboxFailureCountSchema,
    manual_attention_count: outboxFailureCountSchema,
    automatic_retry_count: outboxFailureCountSchema,
    preview_limit: z.union([z.literal(5), z.literal(100)]),
    items: z.array(dashboardBusinessFollowUpItemSchema).max(100),
  }),
  z.object({
    status: z.literal('unavailable'),
    total_count: z.null(),
    failed_count: z.null(),
    dead_letter_count: z.null(),
    manual_attention_count: z.null(),
    automatic_retry_count: z.null(),
    preview_limit: z.union([z.literal(5), z.literal(100)]),
    items: z.array(dashboardBusinessFollowUpItemSchema).max(0),
  }),
])

const dashboardCommandCenterResponseSchema = z.object({
  timestamp: z.string().datetime({ offset: true }),
  action_center: z.object({ outbox_failures: dashboardOutboxFailuresSchema }),
  system_degraded_nodes: z.array(z.string()),
}).refine((response) => {
  const failures = response.action_center.outbox_failures
  if (failures.status === 'unavailable') {
    return response.system_degraded_nodes.includes('outbox_failures')
  }
  return !response.system_degraded_nodes.includes('outbox_failures')
    && failures.total_count === failures.failed_count + failures.dead_letter_count
    && failures.total_count === failures.manual_attention_count + failures.automatic_retry_count
    && failures.manual_attention_count >= failures.dead_letter_count
    && failures.items.length === Math.min(failures.total_count, failures.preview_limit)
    && failures.items.filter((item) => item.processing_state === 'manual_attention').length <= failures.manual_attention_count
    && failures.items.filter((item) => item.processing_state === 'automatic_retry_scheduled').length <= failures.automatic_retry_count
    && new Set(failures.items.map((item) => item.id)).size === failures.items.length
}, 'Outbox availability, counts and preview must be consistent')

export type DashboardCommandCenterOutboxFailureResponse = z.infer<typeof dashboardCommandCenterResponseSchema>
export type DashboardCommandCenterOutboxFailures = z.infer<typeof dashboardOutboxFailuresSchema>
export type DashboardBusinessFollowUpItem = z.infer<typeof dashboardBusinessFollowUpItemSchema>

export function parseDashboardCommandCenterOutboxFailureResponse(payload: unknown): DashboardCommandCenterOutboxFailureResponse {
  return dashboardCommandCenterResponseSchema.parse(payload)
}

export async function fetchDashboardBusinessFollowUpList(): Promise<DashboardCommandCenterOutboxFailureResponse> {
  const response = await axios.get<unknown>('/api/admin/dashboard/business-follow-ups', { suppressGlobalErrorToast: true })
  return parseDashboardCommandCenterOutboxFailureResponse(response.data)
}

export async function fetchDashboardBusinessFollowUpDetail(eventId: number): Promise<DashboardBusinessFollowUpItem> {
  const response = await axios.get<unknown>(`/api/admin/dashboard/business-follow-ups/${eventId}`, { suppressGlobalErrorToast: true })
  const item = dashboardBusinessFollowUpItemSchema.parse(response.data)
  if (item.id !== eventId) throw new Error('业务跟进详情返回的编号与请求不匹配')
  return item
}

export async function fetchDashboardCommandCenterOutboxFailures(): Promise<DashboardCommandCenterOutboxFailureResponse> {
  const response = await axios.get<unknown>('/api/admin/dashboard/command-center', { suppressGlobalErrorToast: true })
  return parseDashboardCommandCenterOutboxFailureResponse(response.data)
}
