import axios from '@/utils/axios'
import {
  requireApiArrayField,
  requireApiBooleanField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiPagination,
  requireApiStringField,
  unwrapApiPayload,
} from '@/utils/apiResponse'
import type { MediaAsset, MediaID, MediaPagination } from '@/api/media'

export type SiteQualityStrategy = 'mobile' | 'desktop'

export interface SiteQualityRemediation {
  label: string
  route: string
}

export interface SiteQualityIssue {
  id: string
  kind?: string
  rule_version?: string
  title: string
  description?: string
  score?: number
  display_value?: string
  numeric_value?: number
  savings_ms?: number
  savings_bytes?: number
  severity: 'low' | 'medium' | 'high' | 'critical'
  resources?: SiteQualityFindingResource[]
  headings?: SiteQualityHeadingEvidence[]
  structured_data?: SiteQualityStructuredDataEvidence[]
  remediation?: SiteQualityRemediation
}

export interface SiteQualityFindingResource {
  url: string
  total_bytes?: number
  wasted_ms?: number
}

export interface SiteQualityFindingEvidence {
  audit_id: string
  title: string
  description?: string
  score?: number
  display_value?: string
  numeric_value?: number
  savings_ms?: number
  savings_bytes?: number
  resources?: SiteQualityFindingResource[]
  headings?: SiteQualityHeadingEvidence[]
  structured_data?: SiteQualityStructuredDataEvidence[]
}

export interface SiteQualityHeadingEvidence {
  level?: number
  text?: string
  snippet?: string
  selector?: string
  explanation?: string
}

export interface SiteQualityStructuredDataEvidence {
  format?: string
  type?: string
  id?: string
  name?: string
  url?: string
  selector?: string
  snippet?: string
  property?: string
  explanation?: string
}

export type SiteQualityFindingState = 'open' | 'acknowledged' | 'resolved' | 'verified'
export type SiteQualityFindingStateFilter = SiteQualityFindingState | 'active' | 'all'
export type SiteQualityFindingKind = 'opportunity' | 'headings' | 'schema'

export interface SiteQualityFinding {
  id: number
  target_id?: number
  target_url: string
  strategy: SiteQualityStrategy
  audit_id: string
  finding_kind?: SiteQualityFindingKind
  rule_version?: string
  confidence: number
  sample_count: number
  confirmations: number
  consecutive_clean: number
  title: string
  description?: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  state: SiteQualityFindingState
  first_detected_at: string
  last_detected_at: string
  latest_run_id: number
  latest_savings_ms?: number
  latest_savings_bytes?: number
  resource_count: number
  latest_evidence: string
  resolution_note?: string
  resolved_at?: string
  verified_at?: string
  created_at: string
  updated_at: string
}

export interface SiteQualityFindingEvent {
  id: number
  finding_id: number
  run_id?: number
  event_type: string
  actor_user_id: number
  note?: string
  metadata?: string
  created_at: string
}

export interface SiteQualityFindingStats {
  active: number
  open: number
  acknowledged: number
  resolved: number
  verified: number
  critical: number
  high: number
}

export interface SiteQualityTargetStats {
  total: number
  enabled: number
  due: number
  critical: number
  standard: number
  background: number
}

export interface SiteQualityJobStats {
  total: number
  queued: number
  processing: number
  succeeded: number
  failed: number
  dead_letter: number
  claimable: number
  stale_leases: number
  oldest_queued_at?: string
  oldest_processing_at?: string
  latest_success_at?: string
  latest_failure_at?: string
  latest_dead_letter_at?: string
}

export interface SiteQualityProviderSlotStats {
  provider: string
  configured: number
  total: number
  available: number
  locked: number
  stale_locked: number
  next_available_at?: string
}

export interface SiteQualityOperationalSummary {
  generated_at: string
  status: 'healthy' | 'degraded' | 'not_configured' | 'unavailable'
  warnings?: string[]
  worker_enabled: boolean
  worker_interval_seconds: number
  runner_configured: boolean
  default_url?: string
  release_id?: string
  sample_count: number
  required_confirmations: number
  required_clean_evaluations: number
  worker_batch_limit: number
  lease_timeout_seconds: number
  provider_concurrency: number
  provider_request_interval_seconds: number
  run_count: number
  latest_run?: SiteQualityRun
  latest_success_at?: string
  targets: SiteQualityTargetStats
  jobs: SiteQualityJobStats
  provider_slots: SiteQualityProviderSlotStats
  findings: SiteQualityFindingStats
}

export interface SiteQualityFindingList {
  items: SiteQualityFinding[]
  stats: SiteQualityFindingStats
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export interface SiteQualityFindingEvents {
  items: SiteQualityFindingEvent[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export interface SiteQualityRun {
  id: number
  target_id?: number
  job_id?: number
  target_url: string
  canonical_url?: string
  final_url?: string
  strategy: SiteQualityStrategy
  status: 'success' | 'failed'
  initiated_by_user_id: number
  performance_score?: number
  accessibility_score?: number
  best_practices_score?: number
  seo_score?: number
  first_contentful_paint_ms?: number
  largest_contentful_paint_ms?: number
  interaction_to_next_paint_ms?: number
  cumulative_layout_shift?: number
  total_blocking_time_ms?: number
  speed_index_ms?: number
  issues: SiteQualityIssue[]
  error_message?: string
  created_at: string
}

export interface SiteQualityRunList {
  runner_configured: boolean
  default_url?: string
  summary: SiteQualityOperationalSummary
  items: SiteQualityRun[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export interface SiteQualityTargetOption {
  url: string
  path: string
  title: string
  locale: string
  source_type: string
  is_home: boolean
}

export interface SiteQualityTargetList {
  default_url: string
  items: SiteQualityTargetOption[]
}

export type SiteQualityJobStatus = 'queued' | 'processing' | 'succeeded' | 'failed' | 'dead_letter'

export interface SiteQualityJob {
  id: number
  target_id: number
  finding_id?: number
  strategy: SiteQualityStrategy
  kind: 'scheduled' | 'manual' | 'recheck'
  status: SiteQualityJobStatus
  idempotency_key: string
  sample_count: number
  required_confirmations: number
  attempts: number
  max_attempts: number
  available_at: string
  locked_at?: string
  locked_by?: string
  lease_generation: number
  lease_expires_at?: string
  heartbeat_at?: string
  started_at?: string
  finished_at?: string
  initiated_by_user_id: number
  release_id?: string
  last_error?: string
  created_at: string
  updated_at: string
}

export type PreflightImageDimensionState =
  | 'all'
  | 'attention'
  | 'missing_dimensions'
  | 'missing_variants'
  | 'ready'

export interface PreflightImageDimensionFinding {
  asset: MediaAsset
  state: 'missing_dimensions' | 'missing_variants' | 'missing_dimensions_and_variants' | 'ready'
  missing_presets?: string[]
}

export interface PreflightImageDimensionSummary {
  total: number
  attention: number
  ready: number
  missing_dimensions: number
  missing_variants: number
}

export interface PreflightImageDimensionPreset {
  name: string
  label: string
  max_width: number
  generation_version: number
  sort_order?: number
  is_system?: boolean
}

export interface PreflightImageDimensionListResult {
  items: PreflightImageDimensionFinding[]
  summary: PreflightImageDimensionSummary
  presets: PreflightImageDimensionPreset[]
  pagination: MediaPagination
}

export type PreflightContentLinkIssueState = 'open' | 'resolved' | 'verified' | 'ignored'
export type PreflightContentLinkIssueStateFilter = PreflightContentLinkIssueState | 'active' | 'all'
export type PreflightContentLinkFixStatus = 'not_fixable' | 'pending' | 'applied' | 'failed'

export interface PreflightContentLinkTargetOption {
  url: string
  path: string
  title: string
  locale: string
  source_type: string
  is_home: boolean
}

export interface PreflightContentLinkTargetList {
  default_url: string
  items: PreflightContentLinkTargetOption[]
}

export interface PreflightContentLinkRun {
  id: number
  target_url: string
  route_entry_id?: number
  status: 'success' | 'failed'
  checked_at: string
  issue_count: number
  fixable_count: number
  error_message?: string
  created_at: string
}

export interface PreflightContentLinkIssue {
  id: number
  route_entry_id?: number
  run_id: number
  target_url: string
  final_url: string
  link_url: string
  link_text: string
  selector: string
  snippet: string
  source_type: string
  source_id?: number
  source_key: string
  source_field: string
  issue_key: string
  severity: 'low' | 'medium' | 'high' | 'critical'
  state: PreflightContentLinkIssueState
  suggested_text: string
  fix_status: PreflightContentLinkFixStatus
  fix_error?: string
  latest_evidence: string
  first_detected_at: string
  last_detected_at: string
  resolved_at?: string
  verified_at?: string
  fixed_at?: string
  created_at: string
  updated_at: string
}

export interface PreflightContentLinkIssueEvent {
  id: number
  issue_id: number
  event_type: string
  actor_user_id: number
  note?: string
  metadata?: string
  created_at: string
}

export interface PreflightContentLinkStats {
  active: number
  open: number
  resolved: number
  verified: number
  ignored: number
  fixable: number
  needs_source: number
  applied: number
}

export interface PreflightContentLinkIssueList {
  items: PreflightContentLinkIssue[]
  stats: PreflightContentLinkStats
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export interface PreflightContentLinkRunResult {
  run: PreflightContentLinkRun
  issues: PreflightContentLinkIssue[]
  stats: PreflightContentLinkStats
}

export interface PreflightContentLinkEvents {
  items: PreflightContentLinkIssueEvent[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}

export type FontPreflightStatus = 'pass' | 'warning' | 'block'

export interface FontPreflightBaseline {
  id: string
  label: string
  font_display: 'swap'
  rules: string[]
}

export interface FontPreflightCheck {
  key: string
  label: string
  status: FontPreflightStatus
  message: string
  details: string[]
}

export interface FontPreflightStrategy {
  status: FontPreflightStatus
  label: string
  default_stack: string[]
  latin_bytes: number
  latin_budget_bytes: number
  complete_maple_ui_family: string
  cjk_unicode_range: string
  layout_parity_verified: boolean
  rationale: string
}

export interface FontPreflightFace {
  family: string
  role: string
  script: string
  filename: string
  bytes: number
  font_display: string
  unicode_range: string
  self_hosted: boolean
}

export interface FontPreflightLocaleCoverage {
  locale: string
  source_files: number
  checked_characters: number
  missing_characters: number
  missing_sample: string[]
  font_stack: string[]
  status: FontPreflightStatus
}

export interface FontPreflightCoverage {
  locale_count: number
  source_file_count: number
  checked_characters: number
  missing_characters: number
  locales: FontPreflightLocaleCoverage[]
}

export interface FontPreflightReport {
  schema_version: number
  project: string
  generated_at: string
  overall_status: FontPreflightStatus
  baseline: FontPreflightBaseline
  checks: FontPreflightCheck[]
  strategy: FontPreflightStrategy
  faces: FontPreflightFace[]
  coverage: FontPreflightCoverage
}

const readPayload = (response: unknown, endpoint: string) => unwrapApiPayload(response, endpoint)

const readObjectPayload = (response: unknown, endpoint: string) => (
  requireApiObject(readPayload(response, endpoint), endpoint)
)

const readSiteQualityRunPayload = (response: unknown, endpoint: string): SiteQualityRun => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'id', endpoint)
  requireApiStringField(payload, 'target_url', endpoint)
  requireApiStringField(payload, 'strategy', endpoint)
  requireApiStringField(payload, 'status', endpoint)
  requireApiArrayField(payload, 'issues', endpoint)
  return payload as SiteQualityRun
}

const readSiteQualityRunsPayload = (response: unknown, endpoint: string): SiteQualityRunList => {
  const payload = readObjectPayload(response, endpoint)
  requireApiBooleanField(payload, 'runner_configured', endpoint)
  requireApiArrayField(payload, 'items', endpoint)
  requireApiObjectField(payload, 'summary', endpoint)
  requireApiObject(payload.pagination, `${endpoint}.pagination`)
  return payload as SiteQualityRunList
}

const readSiteQualityTargetsPayload = (response: unknown, endpoint: string): SiteQualityTargetList => {
  const payload = readObjectPayload(response, endpoint)
  requireApiStringField(payload, 'default_url', endpoint)
  requireApiArrayField(payload, 'items', endpoint)
  return payload as SiteQualityTargetList
}

const readSiteQualityFindingPayload = (response: unknown, endpoint: string): SiteQualityFinding => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'id', endpoint)
  requireApiStringField(payload, 'target_url', endpoint)
  requireApiStringField(payload, 'audit_id', endpoint)
  requireApiStringField(payload, 'state', endpoint)
  return payload as SiteQualityFinding
}

const readSiteQualityFindingActionPayload = (response: unknown, endpoint: string): SiteQualityFinding => {
  const payload = readObjectPayload(response, endpoint)
  return requireApiObject(payload.finding, `${endpoint}.finding`) as SiteQualityFinding
}

const readSiteQualityFindingsPayload = (response: unknown, endpoint: string): SiteQualityFindingList => {
  const payload = readObjectPayload(response, endpoint)
  requireApiArrayField(payload, 'items', endpoint)
  requireApiObject(payload.stats, `${endpoint}.stats`)
  requireApiObject(payload.pagination, `${endpoint}.pagination`)
  return payload as SiteQualityFindingList
}

const readSiteQualityFindingEventsPayload = (response: unknown, endpoint: string): SiteQualityFindingEvents => {
  const payload = readObjectPayload(response, endpoint)
  requireApiArrayField(payload, 'items', endpoint)
  requireApiObject(payload.pagination, `${endpoint}.pagination`)
  return payload as SiteQualityFindingEvents
}

const readSiteQualityJobPayload = (response: unknown, endpoint: string): SiteQualityJob => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'id', endpoint)
  requireApiStringField(payload, 'status', endpoint)
  requireApiStringField(payload, 'strategy', endpoint)
  requireApiStringField(payload, 'kind', endpoint)
  return payload as SiteQualityJob
}

const readFontPreflightPayload = (response: unknown, endpoint: string): FontPreflightReport => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'schema_version', endpoint)
  requireApiStringField(payload, 'project', endpoint)
  requireApiStringField(payload, 'generated_at', endpoint)
  requireApiStringField(payload, 'overall_status', endpoint)
  requireApiObjectField(payload, 'baseline', endpoint)
  requireApiArrayField(payload, 'checks', endpoint)
  requireApiObjectField(payload, 'strategy', endpoint)
  requireApiArrayField(payload, 'faces', endpoint)
  requireApiObjectField(payload, 'coverage', endpoint)
  return payload as FontPreflightReport
}

const readContentLinkTargetsPayload = (response: unknown, endpoint: string): PreflightContentLinkTargetList => {
  const payload = readObjectPayload(response, endpoint)
  requireApiStringField(payload, 'default_url', endpoint)
  requireApiArrayField(payload, 'items', endpoint)
  return payload as PreflightContentLinkTargetList
}

const readContentLinkIssuePayload = (response: unknown, endpoint: string): PreflightContentLinkIssue => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'id', endpoint)
  requireApiStringField(payload, 'target_url', endpoint)
  requireApiStringField(payload, 'link_text', endpoint)
  requireApiStringField(payload, 'state', endpoint)
  return payload as PreflightContentLinkIssue
}

const readContentLinkIssueActionPayload = (response: unknown, endpoint: string): PreflightContentLinkIssue => {
  const payload = readObjectPayload(response, endpoint)
  return requireApiObjectField<PreflightContentLinkIssue>(payload, 'issue', endpoint)
}

const readContentLinkIssuesPayload = (response: unknown, endpoint: string): PreflightContentLinkIssueList => {
  const payload = readObjectPayload(response, endpoint)
  requireApiArrayField(payload, 'items', endpoint)
  requireApiObjectField(payload, 'stats', endpoint)
  requireApiObjectField(payload, 'pagination', endpoint)
  return payload as PreflightContentLinkIssueList
}

const readContentLinkStatsPayload = (response: unknown, endpoint: string): PreflightContentLinkStats => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'active', endpoint)
  requireApiNumberField(payload, 'fixable', endpoint)
  return payload as PreflightContentLinkStats
}

const readContentLinkRunResultPayload = (response: unknown, endpoint: string): PreflightContentLinkRunResult => {
  const payload = readObjectPayload(response, endpoint)
  requireApiObjectField(payload, 'run', endpoint)
  requireApiArrayField(payload, 'issues', endpoint)
  requireApiObjectField(payload, 'stats', endpoint)
  return payload as PreflightContentLinkRunResult
}

const readContentLinkEventsPayload = (response: unknown, endpoint: string): PreflightContentLinkEvents => {
  const payload = readObjectPayload(response, endpoint)
  requireApiArrayField(payload, 'items', endpoint)
  requireApiObjectField(payload, 'pagination', endpoint)
  return payload as PreflightContentLinkEvents
}

export const preflightApi = {
  async getFontPreflight(): Promise<FontPreflightReport> {
    const endpoint = '/api/admin/preflight/fonts'
    return readFontPreflightPayload(await axios.get(endpoint), endpoint)
  },

  async getSiteQualityRuns(params?: {
    page?: number
    pageSize?: number
    url?: string
    strategy?: SiteQualityStrategy
  }): Promise<SiteQualityRunList> {
    const endpoint = '/api/admin/preflight/site-quality'
    const query: Record<string, string | number> = {}
    if (params?.page) query.page = params.page
    if (params?.pageSize) query.page_size = params.pageSize
    if (params?.url) query.url = params.url
    if (params?.strategy) query.strategy = params.strategy
    return readSiteQualityRunsPayload(await axios.get(endpoint, { params: query }), endpoint)
  },

  async getSiteQualityTargets(): Promise<SiteQualityTargetList> {
    const endpoint = '/api/admin/preflight/site-quality/targets'
    return readSiteQualityTargetsPayload(await axios.get(endpoint), endpoint)
  },

  async createSiteQualityJob(url: string, strategy: SiteQualityStrategy): Promise<{ job_id: number; job: SiteQualityJob }> {
    const endpoint = '/api/admin/preflight/site-quality/jobs'
    const payload = readObjectPayload(await axios.post(endpoint, { url, strategy }), endpoint)
    requireApiNumberField(payload, 'job_id', endpoint)
    return {
      job_id: payload.job_id as number,
      job: readSiteQualityJobPayload({ data: payload.job }, `${endpoint}.job`),
    }
  },

  async getSiteQualityJob(id: number): Promise<SiteQualityJob> {
    const endpoint = `/api/admin/preflight/site-quality/jobs/${id}`
    return readSiteQualityJobPayload(await axios.get(endpoint), endpoint)
  },

  async waitForSiteQualityJob(id: number, options?: { intervalMs?: number; timeoutMs?: number }): Promise<SiteQualityJob> {
    const intervalMs = options?.intervalMs ?? 1500
    const timeoutMs = options?.timeoutMs ?? 10 * 60 * 1000
    const startedAt = Date.now()
    while (true) {
      const job = await this.getSiteQualityJob(id)
      if (
        job.status === 'succeeded'
        || job.status === 'failed'
        || job.status === 'dead_letter'
      ) {
        return job
      }
      if (Date.now() - startedAt >= timeoutMs) {
        throw new Error('页面质量任务轮询超时')
      }
      await new Promise((resolve) => window.setTimeout(resolve, intervalMs))
    }
  },

  async enqueueSiteQualityFindingRecheck(id: number): Promise<{ job_id: number; finding_id: number; job: SiteQualityJob }> {
    const endpoint = `/api/admin/preflight/site-quality/findings/${id}/recheck`
    const payload = readObjectPayload(await axios.post(endpoint), endpoint)
    requireApiNumberField(payload, 'job_id', endpoint)
    requireApiNumberField(payload, 'finding_id', endpoint)
    return {
      job_id: payload.job_id as number,
      finding_id: payload.finding_id as number,
      job: readSiteQualityJobPayload({ data: payload.job }, `${endpoint}.job`),
    }
  },

  async getSiteQualityFindings(params?: {
    page?: number
    pageSize?: number
    state?: SiteQualityFindingStateFilter
    severity?: SiteQualityFinding['severity']
    url?: string
    strategy?: SiteQualityStrategy
    kind?: SiteQualityFindingKind
  }): Promise<SiteQualityFindingList> {
    const endpoint = '/api/admin/preflight/site-quality/findings'
    const query: Record<string, string | number> = {}
    if (params?.page) query.page = params.page
    if (params?.pageSize) query.page_size = params.pageSize
    if (params?.state) query.state = params.state
    if (params?.severity) query.severity = params.severity
    if (params?.url) query.url = params.url
    if (params?.strategy) query.strategy = params.strategy
    if (params?.kind) query.kind = params.kind
    return readSiteQualityFindingsPayload(await axios.get(endpoint, { params: query }), endpoint)
  },

  async getSiteQualityFinding(id: number): Promise<SiteQualityFinding> {
    const endpoint = `/api/admin/preflight/site-quality/findings/${id}`
    return readSiteQualityFindingPayload(await axios.get(endpoint), endpoint)
  },

  async getSiteQualityFindingEvents(
    id: number,
    params?: { page?: number; pageSize?: number },
  ): Promise<SiteQualityFindingEvents> {
    const endpoint = `/api/admin/preflight/site-quality/findings/${id}/events`
    const query: Record<string, number> = {}
    if (params?.page) query.page = params.page
    if (params?.pageSize) query.page_size = params.pageSize
    return readSiteQualityFindingEventsPayload(await axios.get(endpoint, { params: query }), endpoint)
  },

  async acknowledgeSiteQualityFinding(id: number, note = ''): Promise<SiteQualityFinding> {
    const endpoint = `/api/admin/preflight/site-quality/findings/${id}/acknowledge`
    return readSiteQualityFindingActionPayload(await axios.post(endpoint, { note }), endpoint)
  },

  async resolveSiteQualityFinding(id: number, resolutionNote: string): Promise<SiteQualityFinding> {
    const endpoint = `/api/admin/preflight/site-quality/findings/${id}/resolve`
    return readSiteQualityFindingActionPayload(
      await axios.post(endpoint, { resolution_note: resolutionNote }),
      endpoint,
    )
  },

  async listImageDimensions(params: {
    page?: number
    page_size?: number
    search?: string
    state?: PreflightImageDimensionState
  } = {}): Promise<PreflightImageDimensionListResult> {
    const endpoint = '/api/admin/preflight/image-dimensions'
    const response = await axios.get(endpoint, { params })
    const body = requireApiObject(response.data, endpoint, 'response body')
    const payload = requireApiObject(unwrapApiPayload(response, endpoint), endpoint)

    return {
      items: requireApiArrayField<PreflightImageDimensionFinding>(payload, 'items', endpoint),
      summary: requireApiObjectField<PreflightImageDimensionSummary>(payload, 'summary', endpoint),
      presets: requireApiArrayField<PreflightImageDimensionPreset>(payload, 'presets', endpoint),
      pagination: requireApiPagination(body, payload, endpoint),
    }
  },

  async reconcileImageDimensions(id: MediaID): Promise<PreflightImageDimensionFinding> {
    const endpoint = `/api/admin/preflight/image-dimensions/${id}/reconcile`
    const payload = requireApiObject(unwrapApiPayload(await axios.post(endpoint), endpoint), endpoint)
    return requireApiObjectField<PreflightImageDimensionFinding>(payload, 'item', endpoint)
  },

  async getContentLinkTargets(): Promise<PreflightContentLinkTargetList> {
    const endpoint = '/api/admin/preflight/content-links/targets'
    return readContentLinkTargetsPayload(await axios.get(endpoint), endpoint)
  },

  async runContentLinkCheck(url: string): Promise<PreflightContentLinkRunResult> {
    const endpoint = '/api/admin/preflight/content-links/runs'
    return readContentLinkRunResultPayload(await axios.post(endpoint, { url }), endpoint)
  },

  async getContentLinkIssues(params?: {
    page?: number
    pageSize?: number
    state?: PreflightContentLinkIssueStateFilter
    url?: string
    search?: string
    fixable?: boolean
  }): Promise<PreflightContentLinkIssueList> {
    const endpoint = '/api/admin/preflight/content-links/issues'
    const query: Record<string, string | number | boolean> = {}
    if (params?.page) query.page = params.page
    if (params?.pageSize) query.page_size = params.pageSize
    if (params?.state) query.state = params.state
    if (params?.url) query.url = params.url
    if (params?.search) query.search = params.search
    if (typeof params?.fixable === 'boolean') query.fixable = params.fixable
    return readContentLinkIssuesPayload(await axios.get(endpoint, { params: query }), endpoint)
  },

  async getContentLinkStats(): Promise<PreflightContentLinkStats> {
    const endpoint = '/api/admin/preflight/content-links/stats'
    return readContentLinkStatsPayload(await axios.get(endpoint), endpoint)
  },

  async getContentLinkIssue(id: number): Promise<PreflightContentLinkIssue> {
    const endpoint = `/api/admin/preflight/content-links/issues/${id}`
    return readContentLinkIssuePayload(await axios.get(endpoint), endpoint)
  },

  async getContentLinkIssueEvents(
    id: number,
    params?: { page?: number; pageSize?: number },
  ): Promise<PreflightContentLinkEvents> {
    const endpoint = `/api/admin/preflight/content-links/issues/${id}/events`
    const query: Record<string, number> = {}
    if (params?.page) query.page = params.page
    if (params?.pageSize) query.page_size = params.pageSize
    return readContentLinkEventsPayload(await axios.get(endpoint, { params: query }), endpoint)
  },

  async applyContentLinkSuggestion(id: number): Promise<PreflightContentLinkIssue> {
    const endpoint = `/api/admin/preflight/content-links/issues/${id}/apply`
    return readContentLinkIssueActionPayload(await axios.post(endpoint), endpoint)
  },

  async resolveContentLinkIssue(id: number, note: string): Promise<PreflightContentLinkIssue> {
    const endpoint = `/api/admin/preflight/content-links/issues/${id}/resolve`
    return readContentLinkIssueActionPayload(await axios.post(endpoint, { note }), endpoint)
  },

  async recheckContentLinkIssue(id: number): Promise<PreflightContentLinkRunResult> {
    const endpoint = `/api/admin/preflight/content-links/issues/${id}/recheck`
    return readContentLinkRunResultPayload(await axios.post(endpoint), endpoint)
  },
}

export default preflightApi
