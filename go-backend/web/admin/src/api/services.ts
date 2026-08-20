import axios from '@/utils/axios'
import {
  requireApiArrayField,
  requireApiBooleanField,
  requireApiNumberField,
  requireApiObject,
  requireApiStringField,
  unwrapApiPayload,
} from '@/utils/apiResponse'
import type {
  OpsConnector,
  OpsConnectorOAuthStartResult,
  OpsDomain,
  OpsEnvironment,
  OpsNetworkSummary,
  OpsProject,
  OpsVPS,
} from '@/api/ops'

export interface ServiceCenterProvider {
  id: string
  label: string
  route?: string
  connection_count: number
  active_connection_count: number
  resource_count: number
  status: 'active' | 'attention' | 'pending' | 'not_connected' | 'not_configured'
}

export interface ServiceCenterOverview {
  environment: OpsEnvironment | ''
  generated_at: string
  providers: ServiceCenterProvider[]
  assets: {
    vps: OpsVPS[]
    projects: OpsProject[]
    domains: OpsDomain[]
  }
  network: OpsNetworkSummary
}

export interface ServiceCenterCloudflareZone {
  name: string
  environment: string
  connector_id?: number
  connector_name?: string
  domain_count: number
  domains: OpsDomain[]
}

export interface ServiceCenterCloudflare {
  environment: string
  generated_at: string
  connection_count: number
  active_connection_count: number
  credential_configured_count: number
  domain_count: number
  zone_count: number
  attention_count: number
  connections: OpsConnector[]
  domains: OpsDomain[]
  zones: ServiceCenterCloudflareZone[]
}

export interface CloudflareCacheRule {
  id: string
  version?: string
  action: string
  description?: string
  expression: string
  enabled: boolean
  last_updated?: string
  edge_ttl_mode?: string
  browser_ttl?: string
  origin_cache_control_status: 'overridden' | 'respected' | 'not_applicable'
}

export interface CloudflareCacheRules {
  connector_id: number
  zone: string
  zone_id: string
  ruleset_id?: string
  ruleset_name?: string
  ruleset_configured: boolean
  origin_cache_control_status: 'overridden' | 'respected' | 'no_rules'
  rules: CloudflareCacheRule[]
}

const readPayload = (response: unknown, endpoint: string) => unwrapApiPayload(response, endpoint)

const readObjectPayload = (response: unknown, endpoint: string) => (
  requireApiObject(readPayload(response, endpoint), endpoint)
)

const readOverviewPayload = (response: unknown, endpoint: string): ServiceCenterOverview => {
  const payload = readObjectPayload(response, endpoint)
  requireApiStringField(payload, 'environment', endpoint)
  requireApiStringField(payload, 'generated_at', endpoint)
  requireApiArrayField(payload, 'providers', endpoint)
  const assets = requireApiObject(payload.assets, endpoint)
  requireApiArrayField(assets, 'vps', endpoint)
  requireApiArrayField(assets, 'projects', endpoint)
  requireApiArrayField(assets, 'domains', endpoint)
  const network = requireApiObject(payload.network, endpoint)
  requireApiStringField(network, 'environment', endpoint)
  requireApiStringField(network, 'generated_at', endpoint)
  const networkSummary = requireApiObject(network.summary, endpoint)
  requireApiNumberField(networkSummary, 'total', endpoint)
  requireApiNumberField(networkSummary, 'explicit_rule_count', endpoint)
  requireApiNumberField(networkSummary, 'inferred_item_count', endpoint)
  requireApiArrayField(network, 'items', endpoint)
  return payload as ServiceCenterOverview
}

const readCloudflarePayload = (response: unknown, endpoint: string): ServiceCenterCloudflare => {
  const payload = readObjectPayload(response, endpoint)
  requireApiStringField(payload, 'generated_at', endpoint)
  requireApiNumberField(payload, 'connection_count', endpoint)
  requireApiNumberField(payload, 'domain_count', endpoint)
  requireApiNumberField(payload, 'zone_count', endpoint)
  requireApiArrayField(payload, 'connections', endpoint)
  requireApiArrayField(payload, 'domains', endpoint)
  requireApiArrayField(payload, 'zones', endpoint)
  return payload as ServiceCenterCloudflare
}

const readOAuthStartPayload = (response: unknown, endpoint: string): OpsConnectorOAuthStartResult => {
  const payload = readObjectPayload(response, endpoint)
  requireApiStringField(payload, 'authorization_url', endpoint)
  requireApiStringField(payload, 'provider', endpoint)
  requireApiNumberField(payload, 'connector_id', endpoint)
  requireApiStringField(payload, 'connector_name', endpoint)
  return payload as OpsConnectorOAuthStartResult
}

const readCloudflareCacheRulesPayload = (response: unknown, endpoint: string): CloudflareCacheRules => {
  const payload = readObjectPayload(response, endpoint)
  requireApiNumberField(payload, 'connector_id', endpoint)
  requireApiStringField(payload, 'zone', endpoint)
  requireApiStringField(payload, 'zone_id', endpoint)
  requireApiBooleanField(payload, 'ruleset_configured', endpoint)
  requireApiStringField(payload, 'origin_cache_control_status', endpoint)
  requireApiArrayField(payload, 'rules', endpoint)
  return payload as CloudflareCacheRules
}

export default {
  async getOverview(environment?: OpsEnvironment): Promise<ServiceCenterOverview> {
    const endpoint = '/api/admin/services/overview'
    return readOverviewPayload(
      await axios.get(endpoint, { params: environment ? { environment } : undefined }),
      endpoint,
    )
  },
  async getCloudflare(environment?: OpsEnvironment): Promise<ServiceCenterCloudflare> {
    const endpoint = '/api/admin/services/cloudflare'
    return readCloudflarePayload(
      await axios.get(endpoint, { params: environment ? { environment } : undefined }),
      endpoint,
    )
  },
  async startCloudflareOAuth(
    connectorID?: number,
    returnPath = '/services/cloudflare',
    environment?: OpsEnvironment,
  ): Promise<OpsConnectorOAuthStartResult> {
    const endpoint = '/api/admin/services/cloudflare/oauth/start'
    const payload: Record<string, unknown> = { return_path: returnPath }
    if (connectorID) payload.connector_id = connectorID
    if (environment) payload.environment = environment
    return readOAuthStartPayload(await axios.post(endpoint, payload), endpoint)
  },
  async testCloudflareConnection(id: number) {
    const endpoint = `/api/admin/services/cloudflare/connectors/${id}/test`
    const payload = readObjectPayload(await axios.post(endpoint), endpoint)
    requireApiNumberField(payload, 'connector_id', endpoint)
    requireApiStringField(payload, 'message', endpoint)
    return payload
  },
  async getCloudflareCacheRules(connectorID: number, zone: string): Promise<CloudflareCacheRules> {
    const endpoint = '/api/admin/services/cloudflare/cache-rules'
    return readCloudflareCacheRulesPayload(
      await axios.get(endpoint, { params: { connector_id: connectorID, zone } }),
      endpoint,
    )
  },
  async setCloudflareCacheRuleEnabled(
    connectorID: number,
    zone: string,
    ruleID: string,
    enabled: boolean,
  ): Promise<CloudflareCacheRules> {
    const endpoint = `/api/admin/services/cloudflare/cache-rules/${encodeURIComponent(ruleID)}/enabled`
    return readCloudflareCacheRulesPayload(
      await axios.patch(endpoint, { enabled }, { params: { connector_id: connectorID, zone } }),
      endpoint,
    )
  },
}
