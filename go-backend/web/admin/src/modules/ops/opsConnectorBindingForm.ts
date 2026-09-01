import type { OpsConnector } from '@/api/ops'

export interface CredentialField {
  key: string
  label: string
  placeholder: string
  secret: boolean
}

export interface OpsConnectorForm {
  id: number
  name: string
  provider: string
  environment: string
  endpoint: string
  auth_type: string
  credential_ref: string
  credentials: Record<string, string>
  scopes: string
  status: string
  enabled: boolean
  notes: string
}

export const opsConnectorProviderOptions = [
  { value: 'cloudflare', label: 'Cloudflare' },
  { value: 'hostinger', label: 'Hostinger' },
  { value: 'github', label: 'GitHub' },
  { value: 'ghcr', label: 'GitHub / GHCR' },
  { value: 'other', label: '其他' },
]

export const opsConnectorEnvironmentOptions = [
  { value: 'production', label: '生产' },
  { value: 'staging', label: '预发布' },
  { value: 'test', label: '测试' },
  { value: 'local', label: '本地' },
]

export const opsConnectorAuthTypeOptions = [
  { value: 'api_token', label: 'API Token' },
  { value: 'api_key', label: 'API Key' },
  { value: 'bearer', label: 'Bearer Token' },
  { value: 'basic', label: 'Basic Auth' },
  { value: 'environment', label: '后端环境变量' },
  { value: 'manual', label: '手动登记' },
  { value: 'none', label: '无需认证' },
]

export const opsConnectorStatusOptions = [
  { value: 'active', label: '正常' },
  { value: 'pending', label: '待测试' },
  { value: 'disabled', label: '已停用' },
  { value: 'error', label: '测试失败' },
]

const emptyConnectorCredentials = (): Record<string, string> => ({
  token: '',
  api_key: '',
  username: '',
  password: '',
})

export const emptyOpsConnectorForm = (): OpsConnectorForm => ({
  id: 0,
  name: '',
  provider: 'cloudflare',
  environment: 'production',
  endpoint: '',
  auth_type: 'api_token',
  credential_ref: '',
  credentials: emptyConnectorCredentials(),
  scopes: '',
  status: 'pending',
  enabled: true,
  notes: '',
})

export const assignOpsConnectorForm = (form: OpsConnectorForm, connector?: Partial<OpsConnector>): void => {
  Object.assign(form, emptyOpsConnectorForm(), connector || {})
  form.credentials = emptyConnectorCredentials()
}

export const connectorCredentialFields = (authType: string): CredentialField[] => {
  if (authType === 'basic') {
    return [
      { key: 'username', label: '用户名', placeholder: '只读账号用户名', secret: false },
      { key: 'password', label: '密码', placeholder: '只读账号密码', secret: true },
    ]
  }
  if (authType === 'api_key') {
    return [{ key: 'api_key', label: 'API Key', placeholder: '粘贴 API Key', secret: true }]
  }
  return [{ key: 'token', label: 'Token', placeholder: '粘贴 Token，不会回显', secret: true }]
}
