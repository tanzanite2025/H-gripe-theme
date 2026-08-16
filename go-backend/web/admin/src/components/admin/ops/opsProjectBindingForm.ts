import type { OpsProject, OpsVPS } from '@/api/ops'

export interface QuickBuyRateLimitForm {
  enabled: boolean
  ipRequestsPerMinute: number
  ipBurst: number
  sessionRequestsPerMinute: number
  sessionBurst: number
  edgeIPRequestsPerMinute: number
  edgeIPBurst: number
  caddyRateLimitEnabled: boolean
}

export interface OpsProjectForm {
  id: number
  name: string
  vps_binding_id: number
  connector_id: number | null
  provider_resource_id: string
  environment: string
  compose_source: string
  compose_project_name: string
  gateway_network: string
  gateway_alias: string
  services: string
  networks: string
  volumes: string
  current_image_tag: string
  current_commit_sha: string
  status: string
  enabled: boolean
  last_deployment_at: string
  backup_policy: string
  restore_notes: string
  quickBuyRateLimit: QuickBuyRateLimitForm
  notes: string
}

export const opsProjectEnvironmentOptions = [
  { value: 'production', label: '生产' },
  { value: 'staging', label: '预发布' },
  { value: 'test', label: '测试' },
  { value: 'local', label: '本地' },
]

export const opsProjectStatusOptions = [
  { value: 'active', label: '正常' },
  { value: 'pending', label: '待确认' },
  { value: 'disabled', label: '已停用' },
  { value: 'drifted', label: '配置漂移' },
  { value: 'error', label: '错误' },
]

const defaultQuickBuyRateLimit = (): QuickBuyRateLimitForm => ({
  enabled: true,
  ipRequestsPerMinute: 120,
  ipBurst: 40,
  sessionRequestsPerMinute: 60,
  sessionBurst: 20,
  edgeIPRequestsPerMinute: 240,
  edgeIPBurst: 80,
  caddyRateLimitEnabled: false,
})

const positiveInteger = (value: unknown, fallback: number): number => {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0) return fallback
  return Math.round(parsed)
}

const padDatePart = (value: number): string => String(value).padStart(2, '0')

export const toDateTimeLocalInput = (value?: string): string => {
  if (!value?.trim()) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value.slice(0, 16)
  return [
    `${date.getFullYear()}-${padDatePart(date.getMonth() + 1)}-${padDatePart(date.getDate())}`,
    `${padDatePart(date.getHours())}:${padDatePart(date.getMinutes())}`,
  ].join('T')
}

export const toOptionalISOTime = (value: string): string => {
  if (!value.trim()) return ''
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toISOString()
}

export const parseQuickBuyRateLimitPolicy = (raw?: string): QuickBuyRateLimitForm => {
  const defaults = defaultQuickBuyRateLimit()
  if (!raw?.trim()) return defaults
  try {
    const parsed = JSON.parse(raw) as Record<string, unknown>
    if (!parsed || typeof parsed !== 'object') return defaults
    return {
      enabled: Boolean(parsed.enabled ?? defaults.enabled),
      ipRequestsPerMinute: positiveInteger(parsed.ip_requests_per_minute, defaults.ipRequestsPerMinute),
      ipBurst: positiveInteger(parsed.ip_burst, defaults.ipBurst),
      sessionRequestsPerMinute: positiveInteger(parsed.session_requests_per_minute, defaults.sessionRequestsPerMinute),
      sessionBurst: positiveInteger(parsed.session_burst, defaults.sessionBurst),
      edgeIPRequestsPerMinute: positiveInteger(parsed.edge_ip_requests_per_minute, defaults.edgeIPRequestsPerMinute),
      edgeIPBurst: positiveInteger(parsed.edge_ip_burst, defaults.edgeIPBurst),
      caddyRateLimitEnabled: Boolean(parsed.caddy_rate_limit_enabled ?? defaults.caddyRateLimitEnabled),
    }
  } catch {
    return defaults
  }
}

export const serializeQuickBuyRateLimitPolicy = (policy: QuickBuyRateLimitForm): string => {
  const defaults = defaultQuickBuyRateLimit()
  return JSON.stringify({
    enabled: Boolean(policy.enabled),
    ip_requests_per_minute: positiveInteger(policy.ipRequestsPerMinute, defaults.ipRequestsPerMinute),
    ip_burst: positiveInteger(policy.ipBurst, defaults.ipBurst),
    session_requests_per_minute: positiveInteger(policy.sessionRequestsPerMinute, defaults.sessionRequestsPerMinute),
    session_burst: positiveInteger(policy.sessionBurst, defaults.sessionBurst),
    edge_ip_requests_per_minute: positiveInteger(policy.edgeIPRequestsPerMinute, defaults.edgeIPRequestsPerMinute),
    edge_ip_burst: positiveInteger(policy.edgeIPBurst, defaults.edgeIPBurst),
    caddy_rate_limit_enabled: Boolean(policy.caddyRateLimitEnabled),
  })
}

export const emptyOpsProjectForm = (): OpsProjectForm => ({
  id: 0,
  name: '',
  vps_binding_id: 0,
  connector_id: null,
  provider_resource_id: '',
  environment: 'production',
  compose_source: '',
  compose_project_name: '',
  gateway_network: '',
  gateway_alias: '',
  services: '',
  networks: '',
  volumes: '',
  current_image_tag: '',
  current_commit_sha: '',
  status: 'pending',
  enabled: true,
  last_deployment_at: '',
  backup_policy: '',
  restore_notes: '',
  quickBuyRateLimit: defaultQuickBuyRateLimit(),
  notes: '',
})

export const assignOpsProjectForm = (form: OpsProjectForm, project?: Partial<OpsProject>): void => {
  Object.assign(form, emptyOpsProjectForm(), project || {})
  form.vps_binding_id = project?.vps_binding_id ?? 0
  form.connector_id = project?.connector_id ?? null
  form.last_deployment_at = toDateTimeLocalInput(project?.last_deployment_at)
  form.quickBuyRateLimit = parseQuickBuyRateLimitPolicy(project?.quick_buy_rate_limit_policy)
}

export const projectQuickBuyRateLimitSummary = (project: OpsProject): string => {
  const policy = parseQuickBuyRateLimitPolicy(project.quick_buy_rate_limit_policy)
  if (!policy.enabled) return '关闭'
  return `${policy.ipRequestsPerMinute}/min IP · ${policy.sessionRequestsPerMinute}/min Session`
}

export const projectVPSLabel = (vps: OpsVPS): string => (
  `${vps.name} · ${vps.hostname || vps.ipv4 || '未登记地址'}`
)
