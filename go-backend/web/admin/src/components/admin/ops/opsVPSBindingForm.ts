import type { OpsVPS } from '@/api/ops'

export interface OpsVPSForm {
  id: number
  name: string
  provider: string
  environment: string
  connector_id: number | null
  provider_resource_id: string
  hostname: string
  ipv4: string
  region: string
  operating_system: string
  status: string
  enabled: boolean
  notes: string
}

export const opsVPSProviderOptions = [
  { value: 'hostinger', label: 'Hostinger' },
  { value: 'other', label: '其他' },
]

export const opsVPSEnvironmentOptions = [
  { value: 'production', label: '生产' },
  { value: 'staging', label: '预发布' },
  { value: 'test', label: '测试' },
  { value: 'local', label: '本地' },
]

export const opsVPSStatusOptions = [
  { value: 'active', label: '正常' },
  { value: 'pending', label: '待确认' },
  { value: 'disabled', label: '已停用' },
  { value: 'drifted', label: '配置漂移' },
  { value: 'error', label: '错误' },
]

export const opsVPSObservedOptions = [
  { value: 'healthy', label: '健康' },
  { value: 'degraded', label: '降级' },
  { value: 'unknown', label: '未同步' },
  { value: 'offline', label: '离线' },
]

export const emptyOpsVPSForm = (): OpsVPSForm => ({
  id: 0,
  name: '',
  provider: 'hostinger',
  environment: 'production',
  connector_id: null,
  provider_resource_id: '',
  hostname: '',
  ipv4: '',
  region: '',
  operating_system: '',
  status: 'pending',
  enabled: true,
  notes: '',
})

export const assignOpsVPSForm = (form: OpsVPSForm, vps?: Partial<OpsVPS>): void => {
  Object.assign(form, emptyOpsVPSForm(), vps || {})
  form.connector_id = vps?.connector_id ?? null
}
