import type { OpsDomain, OpsProject, OpsConnector } from '@/api/ops'

export interface OpsDomainForm {
  id: number
  domain: string
  connector_id: number | null
  project_binding_id: number | null
  role: string
  environment: string
  provider: string
  zone: string
  target: string
  proxy_mode: string
  tls_mode: string
  redirect_target: string
  status: string
  enabled: boolean
  notes: string
}

export const opsDomainRoleOptions = [
  { value: 'canonical', label: '主域名' },
  { value: 'alias', label: '别名' },
  { value: 'admin', label: '后台域' },
  { value: 'redirect', label: '跳转域' },
  { value: 'verification', label: '验证域' },
  { value: 'internal', label: '内部域' },
]

export const opsDomainEnvironmentOptions = [
  { value: 'production', label: '生产' },
  { value: 'staging', label: '预发布' },
  { value: 'test', label: '测试' },
  { value: 'local', label: '本地' },
]

export const opsDomainProviderOptions = [
  { value: 'cloudflare', label: 'Cloudflare' },
  { value: 'hostinger', label: 'Hostinger' },
  { value: 'other', label: '其他' },
]

export const opsDomainProxyOptions = [
  { value: 'proxied', label: '已代理' },
  { value: 'dns_only', label: 'DNS only' },
  { value: 'unknown', label: '未知' },
]

export const opsDomainTLSOptions = [
  { value: 'full_strict', label: 'Full (strict)' },
  { value: 'full', label: 'Full' },
  { value: 'flexible', label: 'Flexible' },
  { value: 'off', label: '关闭' },
  { value: 'unknown', label: '未知' },
]

export const opsDomainStatusOptions = [
  { value: 'active', label: '正常' },
  { value: 'pending', label: '待确认' },
  { value: 'disabled', label: '已停用' },
  { value: 'drifted', label: '配置漂移' },
  { value: 'error', label: '错误' },
]

export const opsDomainObservedStatusOptions = [
  { value: 'unknown', label: '未同步' },
  { value: 'matched', label: '已匹配' },
  { value: 'drifted', label: '漂移' },
  { value: 'error', label: '检查错误' },
]

export const opsDomainDiffStatusOptions = [
  { value: 'unknown', label: '未确认' },
  { value: 'matched', label: '已匹配' },
  { value: 'drifted', label: '有差异' },
  { value: 'error', label: '检查错误' },
]

export const emptyOpsDomainForm = (): OpsDomainForm => ({
  id: 0,
  domain: '',
  connector_id: null,
  project_binding_id: null,
  role: 'alias',
  environment: 'production',
  provider: 'cloudflare',
  zone: '',
  target: '',
  proxy_mode: 'unknown',
  tls_mode: 'unknown',
  redirect_target: '',
  status: 'pending',
  enabled: true,
  notes: '',
})

export const assignOpsDomainForm = (form: OpsDomainForm, domain?: Partial<OpsDomain>): void => {
  Object.assign(form, emptyOpsDomainForm(), domain || {})
  form.connector_id = domain?.connector_id ?? null
  form.project_binding_id = domain?.project_binding_id ?? null
}

export const domainRoleRequiresProject = (role: string): boolean => (
  ['canonical', 'alias', 'admin', 'redirect'].includes(role)
)

export const domainProjectLabel = (project: OpsProject): string => (
  `${project.name}${project.enabled ? '' : '（已停用）'}`
)

export const domainConnectorLabel = (connector: OpsConnector): string => (
  `${connector.name}${connector.enabled === false ? '（已停用）' : ''}`
)
