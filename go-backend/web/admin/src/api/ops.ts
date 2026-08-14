import axios from '@/utils/axios'
import {
  requireApiArrayField,
  requireApiBooleanField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiStringField,
  unwrapApiPayload,
} from '@/utils/apiResponse'

export interface OpsConnectorPayload {
  name: string
  provider: string
  environment: string
  endpoint: string
  auth_type: string
  credential_ref: string
  credentials?: Record<string, string>
  scopes: string
  status: string
  enabled: boolean
  notes: string
}

export interface OpsVPSPayload {
  name: string
  provider: string
  environment: string
  connector_id?: number | null
  provider_resource_id: string
  hostname: string
  ipv4: string
  region: string
  operating_system: string
  status: string
  enabled: boolean
  notes: string
}

export interface OpsProjectPayload {
  name: string
  vps_binding_id: number
  connector_id?: number | null
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
  quick_buy_rate_limit_policy: string
  notes: string
}

export interface OpsDomainSyncResult {
  domain_id: number
  domain: string
  connector_id: number
  connector_name: string
  zone_id?: string
  observed_status: string
  observed_target: string
  observed_proxy_mode: string
  observed_tls_mode: string
  observed_source: string
  last_observed_at: string
  observed_error?: string
  dns_record_count: number
  message: string
}

export interface OpsDomainDiff {
  domain_id: number
  domain: string
  environment: string
  generated_at: string
  status: string
  summary: string
  observed_source?: string
  last_observed_at?: string
  observed_error?: string
  items: Array<{
    key: string
    label: string
    desired: string
    observed: string
    status: string
    message?: string
  }>
}

export interface OpsDomainPreview {
  domain_id: number
  domain: string
  environment: string
  generated_at: string
  warnings: string[]
  dns: {
    provider: string
    zone: string
    record_type: string
    name: string
    content: string
    proxy_mode: string
    tls_mode: string
    redirect: boolean
    redirect_target?: string
  }
  caddy: {
    filename: string
    content: string
  }
  nginx: {
    filename: string
    content: string
  }
}

export interface OpsVPSSyncResult {
  vps_id: number
  name: string
  connector_id: number
  connector_name: string
  provider_resource_id: string
  hostname?: string
  ipv4?: string
  operating_system?: string
  remote_state?: string
  observed_plan?: string
  observed_region?: string
  observed_status: string
  observed_source: string
  last_observed_at: string
  observed_error?: string
  message: string
}

export interface OpsProjectSyncResult {
  project_id: number
  name: string
  vps_id: number
  vps_name: string
  connector_id: number
  connector_name: string
  compose_project_name: string
  remote_state?: string
  health_status: string
  container_count: number
  running_container_count: number
  healthy_container_count: number
  observed_source: string
  last_checked_at: string
  observed_error?: string
  message: string
}

export interface OpsDeploymentPreflightCheck {
  key: string
  category?: string
  label: string
  status: 'pass' | 'warning' | 'block' | 'info'
  message: string
  detail?: string
}

export interface OpsDeploymentPreflightGroup {
  category: string
  label: string
  total_count: number
  blocking_count: number
  warning_count: number
  pass_count: number
  info_count: number
}

export interface OpsDeploymentPreflight {
  project_id: number
  project: string
  environment: string
  generated_at: string
  ready: boolean
  status_level: 'ready' | 'review' | 'blocked'
  blocking_count: number
  warning_count: number
  pass_count: number
  info_count: number
  summary: string
  next_actions: string[]
  categories: OpsDeploymentPreflightGroup[]
  checks: OpsDeploymentPreflightCheck[]
}

export interface OpsDeploymentPreflightSummary {
  project_id: number
  project: string
  environment: string
  ready: boolean
  status_level: 'ready' | 'review' | 'blocked'
  blocking_count: number
  warning_count: number
  pass_count: number
  info_count: number
  summary: string
  block_reasons: string[]
  warn_reasons: string[]
  next_actions: string[]
  categories: OpsDeploymentPreflightGroup[]
  generated_at: string
}

export interface OpsDeploymentPreflightOverview {
  environment: string
  generated_at: string
  project_count: number
  ready_count: number
  review_count: number
  blocked_count: number
  warning_count: number
  categories: OpsDeploymentPreflightGroup[]
  projects: OpsDeploymentPreflightSummary[]
}

export interface OpsOverview {
  environment: string
  generated_at: string
  summary: Record<string, {
    total: number
    enabled: number
    attention: number
    unknown: number
    healthy: number
    configured: number
  }>
  topology: {
    vps: Array<Record<string, any>>
    projects: Array<Record<string, any>>
    domains: Array<Record<string, any>>
  }
  attention: Array<{
    kind: string
    id: number
    name: string
    environment: string
    status: string
    observed_status?: string
    health_status?: string
    message: string
    target?: string
    updated_at: string
  }>
  recent_audit: Array<Record<string, any>>
}

const readPayload = (response: unknown, endpoint: string) => (
  unwrapApiPayload(response, endpoint)
)

const readObjectPayload = (response: unknown, endpoint: string) => (
  requireApiObject(readPayload(response, endpoint), endpoint)
)

const readListPayload = (response: unknown, endpoint: string, field: string) => {
  const payload = readObjectPayload(response, endpoint)
  requireApiArrayField(payload, field, endpoint)
  return payload
}

const readEntityPayload = (response: unknown, endpoint: string): any => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'id', endpoint)
  return payload
}

const readDomainSyncResult = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'domain_id', endpoint)
  requireApiStringField(payload, 'observed_status', endpoint)
  requireApiStringField(payload, 'message', endpoint)
  return payload
}

const readVPSSyncResult = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'vps_id', endpoint)
  requireApiStringField(payload, 'observed_status', endpoint)
  requireApiStringField(payload, 'message', endpoint)
  return payload
}

const readProjectSyncResult = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'project_id', endpoint)
  requireApiStringField(payload, 'health_status', endpoint)
  requireApiStringField(payload, 'message', endpoint)
  return payload
}

const readOverviewPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint)
  requireApiStringField(payload, 'environment', endpoint)
  requireApiStringField(payload, 'generated_at', endpoint)
  const topology = requireApiObjectField(payload, 'topology', endpoint)
  requireApiObjectField(payload, 'summary', endpoint)
  requireApiArrayField(topology, 'vps', endpoint)
  requireApiArrayField(topology, 'projects', endpoint)
  requireApiArrayField(topology, 'domains', endpoint)
  requireApiArrayField(payload, 'attention', endpoint)
  requireApiArrayField(payload, 'recent_audit', endpoint)
  return payload
}

const readDomainDiffPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'domain_id', endpoint)
  requireApiStringField(payload, 'status', endpoint)
  requireApiStringField(payload, 'summary', endpoint)
  requireApiArrayField(payload, 'items', endpoint)
  return payload
}

const readDomainPreviewPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'domain_id', endpoint)
  requireApiArrayField(payload, 'warnings', endpoint)
  requireApiObjectField(payload, 'dns', endpoint)
  requireApiObjectField(payload, 'caddy', endpoint)
  requireApiObjectField(payload, 'nginx', endpoint)
  return payload
}

const readConnectorTestPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'connector_id', endpoint)
  requireApiBooleanField(payload, 'success', endpoint)
  requireApiStringField(payload, 'message', endpoint)
  return payload
}

const readPreflightPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'project_id', endpoint)
  requireApiBooleanField(payload, 'ready', endpoint)
  requireApiStringField(payload, 'status_level', endpoint)
  requireApiArrayField(payload, 'next_actions', endpoint)
  requireApiArrayField(payload, 'categories', endpoint)
  requireApiArrayField(payload, 'checks', endpoint)
  return payload
}

const readPreflightOverviewPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint)
  requireApiStringField(payload, 'environment', endpoint)
  requireApiStringField(payload, 'generated_at', endpoint)
  requireApiNumberField(payload, 'project_count', endpoint)
  requireApiArrayField(payload, 'categories', endpoint)
  requireApiArrayField(payload, 'projects', endpoint)
  return payload
}

export default {
  async getOverview(): Promise<OpsOverview> {
    const endpoint = '/api/admin/ops/overview'
    return readOverviewPayload(await axios.get(endpoint), endpoint) as OpsOverview
  },
  async listDomains() {
    const endpoint = '/api/admin/ops/domains'
    return readListPayload(await axios.get(endpoint), endpoint, 'domains')
  },
  async getDomain(id: number) {
    const endpoint = `/api/admin/ops/domains/${id}`
    return readEntityPayload(await axios.get(endpoint), endpoint)
  },
  async diffDomain(id: number): Promise<OpsDomainDiff> {
    const endpoint = `/api/admin/ops/domains/${id}/diff`
    return readDomainDiffPayload(await axios.get(endpoint), endpoint) as OpsDomainDiff
  },
  async previewDomain(id: number): Promise<OpsDomainPreview> {
    const endpoint = `/api/admin/ops/domains/${id}/preview`
    return readDomainPreviewPayload(await axios.get(endpoint), endpoint) as OpsDomainPreview
  },
  async createDomain(payload: Record<string, unknown>) {
    const endpoint = '/api/admin/ops/domains'
    return readEntityPayload(await axios.post(endpoint, payload), endpoint)
  },
  async updateDomain(id: number, payload: Record<string, unknown>) {
    const endpoint = `/api/admin/ops/domains/${id}`
    return readEntityPayload(await axios.put(endpoint, payload), endpoint)
  },
  async setDomainEnabled(id: number, enabled: boolean) {
    const endpoint = `/api/admin/ops/domains/${id}/enabled`
    return readEntityPayload(await axios.patch(endpoint, { enabled }), endpoint)
  },
  async syncDomain(id: number): Promise<OpsDomainSyncResult> {
    const endpoint = `/api/admin/ops/domains/${id}/sync`
    return readDomainSyncResult(await axios.post(endpoint), endpoint) as OpsDomainSyncResult
  },
  async listConnectors() {
    const endpoint = '/api/admin/ops/connectors'
    return readListPayload(await axios.get(endpoint), endpoint, 'connectors')
  },
  async getConnector(id: number) {
    const endpoint = `/api/admin/ops/connectors/${id}`
    return readEntityPayload(await axios.get(endpoint), endpoint)
  },
  async createConnector(payload: OpsConnectorPayload) {
    const endpoint = '/api/admin/ops/connectors'
    return readEntityPayload(await axios.post(endpoint, payload), endpoint)
  },
  async updateConnector(id: number, payload: OpsConnectorPayload) {
    const endpoint = `/api/admin/ops/connectors/${id}`
    return readEntityPayload(await axios.put(endpoint, payload), endpoint)
  },
  async setConnectorEnabled(id: number, enabled: boolean) {
    const endpoint = `/api/admin/ops/connectors/${id}/enabled`
    return readEntityPayload(await axios.patch(endpoint, { enabled }), endpoint)
  },
  async testConnector(id: number) {
    const endpoint = `/api/admin/ops/connectors/${id}/test`
    return readConnectorTestPayload(await axios.post(endpoint), endpoint)
  },
  async listVPS() {
    const endpoint = '/api/admin/ops/vps'
    return readListPayload(await axios.get(endpoint), endpoint, 'vps')
  },
  async getVPS(id: number) {
    const endpoint = `/api/admin/ops/vps/${id}`
    return readEntityPayload(await axios.get(endpoint), endpoint)
  },
  async createVPS(payload: OpsVPSPayload) {
    const endpoint = '/api/admin/ops/vps'
    return readEntityPayload(await axios.post(endpoint, payload), endpoint)
  },
  async updateVPS(id: number, payload: OpsVPSPayload) {
    const endpoint = `/api/admin/ops/vps/${id}`
    return readEntityPayload(await axios.put(endpoint, payload), endpoint)
  },
  async setVPSEnabled(id: number, enabled: boolean) {
    const endpoint = `/api/admin/ops/vps/${id}/enabled`
    return readEntityPayload(await axios.patch(endpoint, { enabled }), endpoint)
  },
  async syncVPS(id: number): Promise<OpsVPSSyncResult> {
    const endpoint = `/api/admin/ops/vps/${id}/sync`
    return readVPSSyncResult(await axios.post(endpoint), endpoint) as OpsVPSSyncResult
  },
  async listProjects() {
    const endpoint = '/api/admin/ops/projects'
    return readListPayload(await axios.get(endpoint), endpoint, 'projects')
  },
  async getProject(id: number) {
    const endpoint = `/api/admin/ops/projects/${id}`
    return readEntityPayload(await axios.get(endpoint), endpoint)
  },
  async createProject(payload: OpsProjectPayload) {
    const endpoint = '/api/admin/ops/projects'
    return readEntityPayload(await axios.post(endpoint, payload), endpoint)
  },
  async updateProject(id: number, payload: OpsProjectPayload) {
    const endpoint = `/api/admin/ops/projects/${id}`
    return readEntityPayload(await axios.put(endpoint, payload), endpoint)
  },
  async setProjectEnabled(id: number, enabled: boolean) {
    const endpoint = `/api/admin/ops/projects/${id}/enabled`
    return readEntityPayload(await axios.patch(endpoint, { enabled }), endpoint)
  },
  async syncProject(id: number): Promise<OpsProjectSyncResult> {
    const endpoint = `/api/admin/ops/projects/${id}/sync`
    return readProjectSyncResult(await axios.post(endpoint), endpoint) as OpsProjectSyncResult
  },
  async getProjectPreflight(projectID: number): Promise<OpsDeploymentPreflight> {
    const endpoint = `/api/admin/ops/projects/${projectID}/preflight`
    return readPreflightPayload(await axios.get(endpoint), endpoint) as OpsDeploymentPreflight
  },
  async getDeploymentPreflightOverview(): Promise<OpsDeploymentPreflightOverview> {
    const endpoint = '/api/admin/ops/deployments/preflight-overview'
    return readPreflightOverviewPayload(await axios.get(endpoint), endpoint) as OpsDeploymentPreflightOverview
  },
}
