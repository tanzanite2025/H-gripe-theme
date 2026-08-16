import axios from "@/utils/axios";
import {
  requireApiArrayField,
  requireApiBooleanField,
  requireApiNumberField,
  requireApiObject,
  requireApiObjectField,
  requireApiStringField,
  unwrapApiPayload,
} from "@/utils/apiResponse";

export interface OpsConnectorPayload {
  name: string;
  provider: string;
  environment: string;
  endpoint: string;
  auth_type: string;
  credential_ref: string;
  credentials?: Record<string, string>;
  scopes: string;
  status: string;
  enabled: boolean;
  notes: string;
}

export type OpsEnvironment = "production" | "staging" | "test" | "local";

export interface OpsConnectorOAuthStartResult {
  authorization_url: string;
  provider: string;
  connector_id: number;
  connector_name: string;
}

export interface OpsVPSPayload {
  name: string;
  provider: string;
  environment: string;
  connector_id?: number | null;
  provider_resource_id: string;
  hostname: string;
  ipv4: string;
  region: string;
  operating_system: string;
  status: string;
  enabled: boolean;
  notes: string;
}

export interface OpsVPS {
  id: number;
  name: string;
  provider: string;
  environment: string;
  connector_id?: number | null;
  provider_resource_id: string;
  hostname: string;
  ipv4: string;
  region: string;
  operating_system: string;
  status: string;
  observed_status: string;
  observed_state: string;
  observed_source: string;
  observed_hostname: string;
  observed_ipv4: string;
  observed_operating_system: string;
  observed_plan: string;
  observed_region: string;
  enabled: boolean;
  last_observed_at?: string;
  last_error: string;
  notes: string;
}

export interface OpsConnector {
  id: number;
  name: string;
  provider: string;
  environment: string;
  endpoint: string;
  auth_type: string;
  credential_ref: string;
  credential_configured: boolean;
  credential_fields: string[];
  scopes: string;
  status: string;
  enabled: boolean;
  last_test_status: string;
  last_tested_at?: string;
  last_error: string;
  notes: string;
}

export interface OpsConnectorTestResult {
  connector_id: number;
  success: boolean;
  status_code?: number;
  message: string;
  checked_at: string;
  credential_configured: boolean;
}

export interface OpsProjectPayload {
  name: string;
  vps_binding_id: number;
  connector_id?: number | null;
  provider_resource_id: string;
  environment: string;
  compose_source: string;
  compose_project_name: string;
  gateway_network: string;
  gateway_alias: string;
  services: string;
  networks: string;
  volumes: string;
  current_image_tag: string;
  current_commit_sha: string;
  status: string;
  enabled: boolean;
  last_deployment_at: string;
  backup_policy: string;
  restore_notes: string;
  quick_buy_rate_limit_policy: string;
  notes: string;
}

export interface OpsProject {
  id: number;
  name: string;
  vps_binding_id: number;
  connector_id?: number | null;
  provider_resource_id: string;
  environment: string;
  compose_source: string;
  compose_project_name: string;
  gateway_network: string;
  gateway_alias: string;
  services: string;
  networks: string;
  volumes: string;
  current_image_tag: string;
  current_commit_sha: string;
  status: string;
  health_status: string;
  observed_state: string;
  observed_source: string;
  observed_container_count: number;
  observed_running_container_count: number;
  observed_healthy_container_count: number;
  enabled: boolean;
  last_deployment_at?: string;
  last_checked_at?: string;
  last_error: string;
  backup_policy: string;
  restore_notes: string;
  quick_buy_rate_limit_policy: string;
  notes: string;
  vps_name: string;
  vps_provider: string;
  vps_hostname: string;
  vps_ipv4: string;
  vps_connector_id?: number | null;
}

export interface OpsDomain {
  id: number;
  domain: string;
  connector_id?: number | null;
  project_binding_id?: number | null;
  role: string;
  environment: string;
  provider: string;
  zone: string;
  target: string;
  proxy_mode: string;
  tls_mode: string;
  redirect_target: string;
  status: string;
  observed_status: string;
  observed_target: string;
  observed_proxy_mode: string;
  observed_tls_mode: string;
  observed_source: string;
  last_observed_at?: string;
  observed_error?: string;
  enabled: boolean;
  notes: string;
}

export interface OpsAdminAccount {
  id: number;
  email: string;
  username: string;
  first_name?: string;
  last_name?: string;
  role: "admin" | "manager" | "editor" | "support";
  locale: string;
  status: "active" | "inactive" | "suspended";
  created_at: string;
  updated_at: string;
}

export interface OpsAdminAccountInput {
  email: string;
  username?: string;
  password: string;
  role?: "admin" | "manager" | "editor" | "support";
  first_name?: string;
  last_name?: string;
  locale?: string;
}

export interface OpsDomainSyncResult {
  domain_id: number;
  domain: string;
  connector_id: number;
  connector_name: string;
  zone_id?: string;
  observed_status: string;
  observed_target: string;
  observed_proxy_mode: string;
  observed_tls_mode: string;
  observed_source: string;
  last_observed_at: string;
  observed_error?: string;
  dns_record_count: number;
  message: string;
}

export interface OpsDomainDiff {
  domain_id: number;
  domain: string;
  environment: string;
  generated_at: string;
  status: string;
  summary: string;
  observed_source?: string;
  last_observed_at?: string;
  observed_error?: string;
  items: Array<{
    key: string;
    label: string;
    desired: string;
    observed: string;
    status: string;
    message?: string;
  }>;
}

export interface OpsDomainPreview {
  domain_id: number;
  domain: string;
  environment: string;
  generated_at: string;
  warnings: string[];
  dns: {
    provider: string;
    zone: string;
    record_type: string;
    name: string;
    content: string;
    proxy_mode: string;
    tls_mode: string;
    redirect: boolean;
    redirect_target?: string;
  };
  caddy: {
    filename: string;
    content: string;
  };
  nginx: {
    filename: string;
    content: string;
  };
}

export interface OpsVPSSyncResult {
  vps_id: number;
  name: string;
  connector_id: number;
  connector_name: string;
  provider_resource_id: string;
  hostname?: string;
  ipv4?: string;
  operating_system?: string;
  remote_state?: string;
  observed_plan?: string;
  observed_region?: string;
  observed_status: string;
  observed_source: string;
  last_observed_at: string;
  observed_error?: string;
  message: string;
}

export interface OpsProjectSyncResult {
  project_id: number;
  name: string;
  vps_id: number;
  vps_name: string;
  connector_id: number;
  connector_name: string;
  compose_project_name: string;
  remote_state?: string;
  health_status: string;
  container_count: number;
  running_container_count: number;
  healthy_container_count: number;
  observed_source: string;
  last_checked_at: string;
  observed_error?: string;
  message: string;
}

export interface OpsDeploymentPreflightCheck {
  key: string;
  category?: string;
  label: string;
  status: "pass" | "warning" | "block" | "info";
  message: string;
  detail?: string;
}

export interface OpsDeploymentPreflightGroup {
  category: string;
  label: string;
  total_count: number;
  blocking_count: number;
  warning_count: number;
  pass_count: number;
  info_count: number;
}

export interface OpsDeploymentPreflight {
  project_id: number;
  project: string;
  environment: string;
  generated_at: string;
  ready: boolean;
  status_level: "ready" | "review" | "blocked";
  blocking_count: number;
  warning_count: number;
  pass_count: number;
  info_count: number;
  summary: string;
  next_actions: string[];
  categories: OpsDeploymentPreflightGroup[];
  checks: OpsDeploymentPreflightCheck[];
}

export interface OpsDeploymentPreflightSummary {
  project_id: number;
  project: string;
  environment: string;
  ready: boolean;
  status_level: "ready" | "review" | "blocked";
  blocking_count: number;
  warning_count: number;
  pass_count: number;
  info_count: number;
  summary: string;
  block_reasons: string[];
  warn_reasons: string[];
  next_actions: string[];
  categories: OpsDeploymentPreflightGroup[];
  generated_at: string;
}

export interface OpsDeploymentPreflightOverview {
  environment: string;
  generated_at: string;
  project_count: number;
  ready_count: number;
  review_count: number;
  blocked_count: number;
  warning_count: number;
  categories: OpsDeploymentPreflightGroup[];
  projects: OpsDeploymentPreflightSummary[];
}

export interface OpsDeploymentWorkflowStep {
  id: number;
  workflow_run_id: number;
  sequence: number;
  key: string;
  label: string;
  status: "pending" | "running" | "succeeded" | "failed" | "skipped";
  retryable: boolean;
  external_effect: boolean;
  external_operation_id?: string;
  output_summary?: string;
  error_message?: string;
  started_at?: string;
  completed_at?: string;
}

export interface OpsDeploymentWorkflow {
  id: number;
  kind: string;
  mode: "dry_run" | "production";
  project_id: number;
  project: string;
  environment: string;
  requested_ref: string;
  status: string;
  preflight_status: "ready" | "review" | "blocked" | string;
  created_by_id: number;
  created_by: string;
  approved_by_id?: number;
  approved_by?: string;
  approved_at?: string;
  started_at?: string;
  completed_at?: string;
  previous_ref?: string;
  rollback_ref?: string;
  remote_operation_id?: string;
  health_status?: string;
  last_error?: string;
  preflight?: OpsDeploymentPreflight;
  steps: OpsDeploymentWorkflowStep[];
  created_at: string;
  updated_at: string;
}

export interface OpsOverview {
  environment: string;
  generated_at: string;
  summary: Record<
    string,
    {
      total: number;
      enabled: number;
      attention: number;
      unknown: number;
      healthy: number;
      configured: number;
    }
  >;
  topology: {
    vps: OpsVPS[];
    projects: OpsProject[];
    domains: OpsDomain[];
  };
  attention: Array<{
    kind: string;
    id: number;
    name: string;
    environment: string;
    status: string;
    observed_status?: string;
    health_status?: string;
    message: string;
    target?: string;
    updated_at: string;
  }>;
  recent_audit: OpsOverviewAuditLog[];
}

export interface OpsOverviewAuditLog {
  id: number;
  resource: string;
  action: string;
  status: string;
  username?: string;
  created_at: string;
}

const readPayload = (response: unknown, endpoint: string) =>
  unwrapApiPayload(response, endpoint);

const readObjectPayload = (response: unknown, endpoint: string) =>
  requireApiObject(readPayload(response, endpoint), endpoint);

const readListPayload = (
  response: unknown,
  endpoint: string,
  field: string,
) => {
  const payload = readObjectPayload(response, endpoint);
  requireApiArrayField(payload, field, endpoint);
  return payload;
};

const readEntityPayload = (response: unknown, endpoint: string): any => {
  const payload = readObjectPayload(response, endpoint);
  requireApiNumberField(payload, "id", endpoint);
  return payload;
};

const readDomainSyncResult = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint);
  requireApiNumberField(payload, "domain_id", endpoint);
  requireApiStringField(payload, "observed_status", endpoint);
  requireApiStringField(payload, "message", endpoint);
  return payload;
};

const readVPSSyncResult = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint);
  requireApiNumberField(payload, "vps_id", endpoint);
  requireApiStringField(payload, "observed_status", endpoint);
  requireApiStringField(payload, "message", endpoint);
  return payload;
};

const readProjectSyncResult = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint);
  requireApiNumberField(payload, "project_id", endpoint);
  requireApiStringField(payload, "health_status", endpoint);
  requireApiStringField(payload, "message", endpoint);
  return payload;
};

const readOverviewPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint);
  requireApiStringField(payload, "environment", endpoint);
  requireApiStringField(payload, "generated_at", endpoint);
  const topology = requireApiObjectField(payload, "topology", endpoint);
  requireApiObjectField(payload, "summary", endpoint);
  requireApiArrayField(topology, "vps", endpoint);
  requireApiArrayField(topology, "projects", endpoint);
  requireApiArrayField(topology, "domains", endpoint);
  requireApiArrayField(payload, "attention", endpoint);
  requireApiArrayField(payload, "recent_audit", endpoint);
  return payload;
};

const readDomainDiffPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint);
  requireApiNumberField(payload, "domain_id", endpoint);
  requireApiStringField(payload, "status", endpoint);
  requireApiStringField(payload, "summary", endpoint);
  requireApiArrayField(payload, "items", endpoint);
  return payload;
};

const readDomainPreviewPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint);
  requireApiNumberField(payload, "domain_id", endpoint);
  requireApiArrayField(payload, "warnings", endpoint);
  requireApiObjectField(payload, "dns", endpoint);
  requireApiObjectField(payload, "caddy", endpoint);
  requireApiObjectField(payload, "nginx", endpoint);
  return payload;
};

const readConnectorTestPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint);
  requireApiNumberField(payload, "connector_id", endpoint);
  requireApiBooleanField(payload, "success", endpoint);
  requireApiStringField(payload, "message", endpoint);
  return payload;
};

const readPreflightPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint);
  requireApiNumberField(payload, "project_id", endpoint);
  requireApiBooleanField(payload, "ready", endpoint);
  requireApiStringField(payload, "status_level", endpoint);
  requireApiArrayField(payload, "next_actions", endpoint);
  requireApiArrayField(payload, "categories", endpoint);
  requireApiArrayField(payload, "checks", endpoint);
  return payload;
};

const readPreflightOverviewPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint);
  requireApiStringField(payload, "environment", endpoint);
  requireApiStringField(payload, "generated_at", endpoint);
  requireApiNumberField(payload, "project_count", endpoint);
  requireApiArrayField(payload, "categories", endpoint);
  requireApiArrayField(payload, "projects", endpoint);
  return payload;
};

const readWorkflowPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint);
  requireApiNumberField(payload, "id", endpoint);
  requireApiNumberField(payload, "project_id", endpoint);
  requireApiStringField(payload, "mode", endpoint);
  requireApiStringField(payload, "status", endpoint);
  requireApiArrayField(payload, "steps", endpoint);
  return payload as OpsDeploymentWorkflow;
};

const readWorkflowListPayload = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint);
  requireApiArrayField(payload, "workflows", endpoint);
  return payload.workflows as OpsDeploymentWorkflow[];
};

const readAdminAccountsPayload = (
  response: unknown,
  endpoint: string,
): OpsAdminAccount[] => {
  const payload = readObjectPayload(response, endpoint);
  requireApiArrayField(payload, "accounts", endpoint);
  return payload.accounts as OpsAdminAccount[];
};

const readAdminAccountResult = (response: unknown, endpoint: string) => {
  const payload = readObjectPayload(response, endpoint);
  requireApiNumberField(payload, "user_id", endpoint);
  requireApiStringField(payload, "email", endpoint);
  requireApiStringField(payload, "username", endpoint);
  requireApiStringField(payload, "role", endpoint);
  requireApiStringField(payload, "status", endpoint);
  requireApiBooleanField(payload, "created", endpoint);
  return payload;
};

export default {
  async getOverview(
    environment: OpsEnvironment = "production",
  ): Promise<OpsOverview> {
    const endpoint = "/api/admin/ops/overview";
    return readOverviewPayload(
      await axios.get(endpoint, { params: { environment } }),
      endpoint,
    ) as OpsOverview;
  },
  async listAdminAccounts(search = ""): Promise<OpsAdminAccount[]> {
    const endpoint = "/api/admin/ops/admin-accounts";
    const response = await axios.get(endpoint, {
      params: search.trim() ? { search: search.trim() } : undefined,
    });
    return readAdminAccountsPayload(response, endpoint);
  },
  async ensureAdminAccount(payload: OpsAdminAccountInput) {
    const endpoint = "/api/admin/ops/admin-accounts/ensure";
    return readAdminAccountResult(
      await axios.post(endpoint, payload),
      endpoint,
    );
  },
  async listDomains(environment?: OpsEnvironment) {
    const endpoint = "/api/admin/ops/domains";
    return readListPayload(
      await axios.get(endpoint, {
        params: environment ? { environment } : undefined,
      }),
      endpoint,
      "domains",
    );
  },
  async getDomain(id: number) {
    const endpoint = `/api/admin/ops/domains/${id}`;
    return readEntityPayload(await axios.get(endpoint), endpoint);
  },
  async diffDomain(id: number): Promise<OpsDomainDiff> {
    const endpoint = `/api/admin/ops/domains/${id}/diff`;
    return readDomainDiffPayload(
      await axios.get(endpoint),
      endpoint,
    ) as OpsDomainDiff;
  },
  async previewDomain(id: number): Promise<OpsDomainPreview> {
    const endpoint = `/api/admin/ops/domains/${id}/preview`;
    return readDomainPreviewPayload(
      await axios.get(endpoint),
      endpoint,
    ) as OpsDomainPreview;
  },
  async createDomain(payload: Record<string, unknown>) {
    const endpoint = "/api/admin/ops/domains";
    return readEntityPayload(await axios.post(endpoint, payload), endpoint);
  },
  async updateDomain(id: number, payload: Record<string, unknown>) {
    const endpoint = `/api/admin/ops/domains/${id}`;
    return readEntityPayload(await axios.put(endpoint, payload), endpoint);
  },
  async setDomainEnabled(id: number, enabled: boolean) {
    const endpoint = `/api/admin/ops/domains/${id}/enabled`;
    return readEntityPayload(
      await axios.patch(endpoint, { enabled }),
      endpoint,
    );
  },
  async syncDomain(id: number): Promise<OpsDomainSyncResult> {
    const endpoint = `/api/admin/ops/domains/${id}/sync`;
    return readDomainSyncResult(
      await axios.post(endpoint),
      endpoint,
    ) as OpsDomainSyncResult;
  },
  async listConnectors(environment?: OpsEnvironment) {
    const endpoint = "/api/admin/ops/connectors";
    return readListPayload(
      await axios.get(endpoint, {
        params: environment ? { environment } : undefined,
      }),
      endpoint,
      "connectors",
    );
  },
  async getConnector(id: number) {
    const endpoint = `/api/admin/ops/connectors/${id}`;
    return readEntityPayload(await axios.get(endpoint), endpoint);
  },
  async createConnector(payload: OpsConnectorPayload) {
    const endpoint = "/api/admin/ops/connectors";
    return readEntityPayload(await axios.post(endpoint, payload), endpoint);
  },
  async updateConnector(id: number, payload: OpsConnectorPayload) {
    const endpoint = `/api/admin/ops/connectors/${id}`;
    return readEntityPayload(await axios.put(endpoint, payload), endpoint);
  },
  async setConnectorEnabled(id: number, enabled: boolean) {
    const endpoint = `/api/admin/ops/connectors/${id}/enabled`;
    return readEntityPayload(
      await axios.patch(endpoint, { enabled }),
      endpoint,
    );
  },
  async testConnector(id: number): Promise<OpsConnectorTestResult> {
    const endpoint = `/api/admin/ops/connectors/${id}/test`;
    return readConnectorTestPayload(await axios.post(endpoint), endpoint) as OpsConnectorTestResult;
  },
  async startConnectorOAuth(
    provider: "cloudflare" | "hostinger",
    connectorID?: number,
    returnPath = "/ops/connectors",
    environment?: OpsEnvironment,
  ): Promise<OpsConnectorOAuthStartResult> {
    const endpoint = "/api/admin/ops/connectors/oauth/start";
    const payload: Record<string, unknown> = {
      provider,
      return_path: returnPath,
    };
    if (connectorID) payload.connector_id = connectorID;
    if (environment) payload.environment = environment;
    const result = readObjectPayload(
      await axios.post(endpoint, payload),
      endpoint,
    );
    requireApiStringField(result, "authorization_url", endpoint);
    requireApiStringField(result, "provider", endpoint);
    requireApiNumberField(result, "connector_id", endpoint);
    requireApiStringField(result, "connector_name", endpoint);
    return result as OpsConnectorOAuthStartResult;
  },
  async listVPS(environment?: OpsEnvironment) {
    const endpoint = "/api/admin/ops/vps";
    return readListPayload(
      await axios.get(endpoint, {
        params: environment ? { environment } : undefined,
      }),
      endpoint,
      "vps",
    );
  },
  async getVPS(id: number) {
    const endpoint = `/api/admin/ops/vps/${id}`;
    return readEntityPayload(await axios.get(endpoint), endpoint);
  },
  async createVPS(payload: OpsVPSPayload) {
    const endpoint = "/api/admin/ops/vps";
    return readEntityPayload(await axios.post(endpoint, payload), endpoint);
  },
  async updateVPS(id: number, payload: OpsVPSPayload) {
    const endpoint = `/api/admin/ops/vps/${id}`;
    return readEntityPayload(await axios.put(endpoint, payload), endpoint);
  },
  async setVPSEnabled(id: number, enabled: boolean) {
    const endpoint = `/api/admin/ops/vps/${id}/enabled`;
    return readEntityPayload(
      await axios.patch(endpoint, { enabled }),
      endpoint,
    );
  },
  async syncVPS(id: number): Promise<OpsVPSSyncResult> {
    const endpoint = `/api/admin/ops/vps/${id}/sync`;
    return readVPSSyncResult(
      await axios.post(endpoint),
      endpoint,
    ) as OpsVPSSyncResult;
  },
  async listProjects(environment?: OpsEnvironment) {
    const endpoint = "/api/admin/ops/projects";
    return readListPayload(
      await axios.get(endpoint, {
        params: environment ? { environment } : undefined,
      }),
      endpoint,
      "projects",
    );
  },
  async getProject(id: number) {
    const endpoint = `/api/admin/ops/projects/${id}`;
    return readEntityPayload(await axios.get(endpoint), endpoint);
  },
  async createProject(payload: OpsProjectPayload) {
    const endpoint = "/api/admin/ops/projects";
    return readEntityPayload(await axios.post(endpoint, payload), endpoint);
  },
  async updateProject(id: number, payload: OpsProjectPayload) {
    const endpoint = `/api/admin/ops/projects/${id}`;
    return readEntityPayload(await axios.put(endpoint, payload), endpoint);
  },
  async setProjectEnabled(id: number, enabled: boolean) {
    const endpoint = `/api/admin/ops/projects/${id}/enabled`;
    return readEntityPayload(
      await axios.patch(endpoint, { enabled }),
      endpoint,
    );
  },
  async syncProject(id: number): Promise<OpsProjectSyncResult> {
    const endpoint = `/api/admin/ops/projects/${id}/sync`;
    return readProjectSyncResult(
      await axios.post(endpoint),
      endpoint,
    ) as OpsProjectSyncResult;
  },
  async getProjectPreflight(
    projectID: number,
  ): Promise<OpsDeploymentPreflight> {
    const endpoint = `/api/admin/ops/projects/${projectID}/preflight`;
    return readPreflightPayload(
      await axios.get(endpoint),
      endpoint,
    ) as OpsDeploymentPreflight;
  },
  async getDeploymentPreflightOverview(
    environment?: OpsEnvironment,
  ): Promise<OpsDeploymentPreflightOverview> {
    const endpoint = "/api/admin/ops/deployments/preflight-overview";
    return readPreflightOverviewPayload(
      await axios.get(endpoint, {
        params: environment ? { environment } : undefined,
      }),
      endpoint,
    ) as OpsDeploymentPreflightOverview;
  },
  async listWorkflows(projectID = 0): Promise<OpsDeploymentWorkflow[]> {
    const endpoint = "/api/admin/ops/workflows";
    const response = await axios.get(endpoint, {
      params: projectID > 0 ? { project_id: projectID } : undefined,
    });
    return readWorkflowListPayload(response, endpoint);
  },
  async getWorkflow(id: number): Promise<OpsDeploymentWorkflow> {
    const endpoint = `/api/admin/ops/workflows/${id}`;
    return readWorkflowPayload(await axios.get(endpoint), endpoint);
  },
  async createDryRun(
    projectID: number,
    requestedRef = "",
  ): Promise<OpsDeploymentWorkflow> {
    const endpoint = "/api/admin/ops/workflows";
    return readWorkflowPayload(
      await axios.post(endpoint, {
        project_id: projectID,
        requested_ref: requestedRef,
        mode: "dry_run",
      }),
      endpoint,
    );
  },
  async createProduction(
    projectID: number,
    requestedRef = "master",
  ): Promise<OpsDeploymentWorkflow> {
    const endpoint = "/api/admin/ops/workflows";
    return readWorkflowPayload(
      await axios.post(endpoint, {
        project_id: projectID,
        requested_ref: requestedRef || "master",
        mode: "production",
      }),
      endpoint,
    );
  },
  async validateWorkflow(id: number): Promise<OpsDeploymentWorkflow> {
    const endpoint = `/api/admin/ops/workflows/${id}/validate`;
    return readWorkflowPayload(await axios.post(endpoint), endpoint);
  },
  async approveWorkflow(id: number): Promise<OpsDeploymentWorkflow> {
    const endpoint = `/api/admin/ops/workflows/${id}/approve`;
    return readWorkflowPayload(await axios.post(endpoint), endpoint);
  },
  async executeDryRun(id: number): Promise<OpsDeploymentWorkflow> {
    const endpoint = `/api/admin/ops/workflows/${id}/execute`;
    return readWorkflowPayload(await axios.post(endpoint), endpoint);
  },
  async executeWorkflow(id: number): Promise<OpsDeploymentWorkflow> {
    const endpoint = `/api/admin/ops/workflows/${id}/execute`;
    return readWorkflowPayload(await axios.post(endpoint), endpoint);
  },
  async retryWorkflow(id: number): Promise<OpsDeploymentWorkflow> {
    const endpoint = `/api/admin/ops/workflows/${id}/retry`;
    return readWorkflowPayload(await axios.post(endpoint), endpoint);
  },
  async rollbackWorkflow(id: number): Promise<OpsDeploymentWorkflow> {
    const endpoint = `/api/admin/ops/workflows/${id}/rollback`;
    return readWorkflowPayload(await axios.post(endpoint), endpoint);
  },
  async cancelWorkflow(id: number): Promise<OpsDeploymentWorkflow> {
    const endpoint = `/api/admin/ops/workflows/${id}/cancel`;
    return readWorkflowPayload(await axios.post(endpoint), endpoint);
  },
};
