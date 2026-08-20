<template>
 <div class="space-y-4">
    <AdminPageHeader title="服务中心 / Cloudflare" description="查看 Cloudflare 授权、Zone/DNS、代理/TLS 与 Cache Rules 的实际状态">
      <template #actions>
        <select
          v-model="environmentFilter"
 class="h-9 border bg-background px-3 text-sm"
          aria-label="筛选 Cloudflare 环境"
          :disabled="loading || Boolean(oauthPending)"
          @change="changeEnvironment"
        >
          <option value="">全部环境</option>
          <option v-for="option in opsConnectorEnvironmentOptions" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
        <Button size="icon" variant="outline" title="刷新 Cloudflare 状态" :disabled="loading" @click="loadCloudflare">
 <RefreshCw :class="['size-4', loading ? 'animate-spin': '']" />
        </Button>
        <Button
          v-if="canManage"
          :disabled="Boolean(oauthPending)"
          :title="connections.length ? '重新授权 Cloudflare' : '连接 Cloudflare'"
          @click="startOAuth()"
        >
 <LoaderCircle v-if="oauthPending" class="size-4 animate-spin" />
 <KeyRound v-else class="size-4" />
          {{ connections.length ? '重新授权' : '连接 Cloudflare' }}
        </Button>
      </template>
    </AdminPageHeader>

    <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
 <div v-for="item in summaryItems" :key="item.label" class="border bg-card p-3">
 <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">{{ item.label }}</p>
 <p class="mt-2 text-2xl font-black" :class="item.tone">{{ item.value }}</p>
      </div>
    </section>

    <section class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
      <button
        v-for="item in capabilityItems"
        :key="item.key"
        type="button"
        class="border border-dashed border-border/80 bg-card p-3 text-left transition-colors hover:border-primary/50 hover:bg-muted/20"
        @click="activeTab = item.tab"
      >
        <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">{{ item.label }}</p>
        <p class="mt-2 text-sm font-black">{{ item.value }}</p>
        <p class="mt-1 text-[10px] text-muted-foreground">{{ item.detail }}</p>
      </button>
    </section>

    <Tabs v-model="activeTab">
      <TabsList variant="line" class="max-w-full overflow-x-auto">
        <TabsTrigger value="connections"><Link2 class="size-3.5" />授权与权限</TabsTrigger>
        <TabsTrigger value="zones"><Globe2 class="size-3.5" />Zone / DNS</TabsTrigger>
        <TabsTrigger value="cache"><Gauge class="size-3.5" />Cache Rules</TabsTrigger>
        <TabsTrigger value="status"><Activity class="size-3.5" />状态</TabsTrigger>
      </TabsList>

 <TabsContent value="connections" class="mt-3">
 <div v-if="connections.length === 0" class="border border-dashed p-8 text-center text-sm text-muted-foreground">
          尚未连接 Cloudflare
        </div>
 <section v-else class="grid gap-3 xl:grid-cols-2">
 <article v-for="connection in connections" :key="connection.id" class="border bg-card p-4">
 <div class="flex items-start justify-between gap-4">
 <div class="min-w-0">
 <p class="truncate font-black">{{ connection.name }}</p>
 <p class="mt-1 text-xs text-muted-foreground">{{ environmentLabel(connection.environment) }}</p>
              </div>
              <AdminStatusBadge :tone="connectionTone(connection.status)">
                {{ connectionStatusLabel(connection.status) }}
              </AdminStatusBadge>
            </div>

 <dl class="mt-5 grid gap-3 text-xs sm:grid-cols-2">
              <div>
 <dt class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">授权</dt>
 <dd class="mt-1 font-bold">{{ connection.credential_configured ? '已配置': '待配置'}}</dd>
              </div>
              <div>
 <dt class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">权限范围</dt>
 <dd class="mt-1 truncate font-bold" :title="connection.scopes">{{ connection.scopes || '-'}}</dd>
              </div>
              <div>
 <dt class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">最近测试</dt>
 <dd class="mt-1 font-bold">{{ formatDate(connection.last_tested_at) }}</dd>
              </div>
              <div>
 <dt class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">测试状态</dt>
 <dd class="mt-1 font-bold">{{ testLabel(connection.last_test_status) }}</dd>
              </div>
            </dl>

 <p v-if="connection.last_error" class="mt-4 border-l-2 border-rose-500 pl-3 text-xs text-rose-700">
              {{ connection.last_error }}
            </p>

 <div v-if="canManage" class="mt-5 flex justify-end gap-2 border-t border-dashed pt-3">
              <Button
                size="icon"
                variant="outline"
                :title="`测试连接 ${connection.name}`"
                :disabled="testingID === connection.id || Boolean(oauthPending)"
                @click="testConnection(connection.id)"
              >
 <LoaderCircle v-if="testingID === connection.id" class="size-4 animate-spin" />
 <PlugZap v-else class="size-4" />
              </Button>
              <Button
                size="icon"
                variant="outline"
                :title="`重新授权 ${connection.name}`"
                :disabled="Boolean(oauthPending)"
                @click="startOAuth(connection.id)"
              >
 <KeyRound class="size-4" />
              </Button>
            </div>
          </article>
        </section>
      </TabsContent>

 <TabsContent value="zones" class="mt-3">
 <div v-if="zones.length === 0" class="border border-dashed p-8 text-center text-sm text-muted-foreground">
          尚未绑定 Cloudflare Zone
        </div>
 <section v-else class="space-y-3">
 <article v-for="zone in zones" :key="zoneKey(zone)" class="border bg-card">
 <div class="flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3">
 <div class="min-w-0">
 <p class="truncate font-black">{{ zone.name }}</p>
 <p class="mt-1 text-xs text-muted-foreground">
                  {{ environmentLabel(zone.environment) }}<span v-if="zone.connector_name"> · {{ zone.connector_name }}</span>
                </p>
              </div>
              <AdminStatusBadge tone="blue">{{ zone.domain_count }} 个域名</AdminStatusBadge>
            </div>
 <div class="divide-y">
 <div v-for="domain in zone.domains" :key="domain.id" class="grid gap-3 px-4 py-3 text-xs sm:grid-cols-[minmax(0,1fr)_auto_auto_auto] sm:items-center">
 <div class="min-w-0">
 <p class="truncate font-bold">{{ domain.domain }}</p>
 <p class="mt-1 truncate text-[10px] text-muted-foreground">{{ domain.target || '-'}}</p>
                </div>
                <span>{{ roleLabel(domain.role) }}</span>
                <span>{{ proxyLabel(domain.proxy_mode) }}</span>
                <AdminStatusBadge :tone="domainTone(domain)">
                  {{ domainStatusLabel(domain) }}
                </AdminStatusBadge>
              </div>
            </div>
          </article>
        </section>
      </TabsContent>

 <TabsContent value="cache" class="mt-3">
 <div class="flex flex-wrap items-center justify-between gap-2 border bg-card p-3">
          <select
            v-model="cacheTargetKey"
 class="h-9 min-w-64 border bg-background px-3 text-sm"
            aria-label="选择 Cloudflare Cache Rules Zone"
            :disabled="cacheTargets.length === 0 || cacheLoading"
          >
            <option v-for="target in cacheTargets" :key="target.key" :value="target.key">
              {{ target.zone }} · {{ environmentLabel(target.environment) }}
            </option>
          </select>
          <Button
            size="icon"
            variant="outline"
            title="刷新 Cache Rules"
            :disabled="cacheLoading || !selectedCacheTarget"
            @click="loadCacheRules"
          >
 <RefreshCw :class="['size-4', cacheLoading ? 'animate-spin': '']" />
          </Button>
        </div>

 <div v-if="cacheTargets.length === 0" class="mt-3 border border-dashed p-8 text-center text-sm text-muted-foreground">
          尚未绑定可读取缓存规则的 Cloudflare Zone
        </div>
 <div v-else-if="cacheLoading" class="mt-3 border border-dashed p-8 text-center text-sm text-muted-foreground">
          正在读取 Cache Rules
        </div>
 <div v-else-if="cacheRules" class="mt-3 space-y-3">
 <section class="grid gap-3 md:grid-cols-3">
 <div class="border bg-card p-4">
 <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">来源策略</p>
 <p class="mt-3 text-sm font-black">{{ cacheOriginPolicyLabel(cacheRules.origin_cache_control_status) }}</p>
 <AdminStatusBadge class="mt-3" :tone="cacheOriginPolicyTone(cacheRules.origin_cache_control_status)">
                {{ cacheRules.origin_cache_control_status }}
              </AdminStatusBadge>
            </div>
 <div class="border bg-card p-4">
 <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">Ruleset</p>
 <p class="mt-3 truncate text-sm font-black" :title="cacheRules.ruleset_name || ''">
                {{ cacheRules.ruleset_name || '未配置' }}
              </p>
 <AdminStatusBadge class="mt-3" :tone="cacheRules.ruleset_configured ? 'green': 'gray'">
                {{ cacheRules.ruleset_configured ? '已读取' : '无规则' }}
              </AdminStatusBadge>
            </div>
 <div class="border bg-card p-4">
 <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">规则数</p>
 <p class="mt-3 text-sm font-black">{{ cacheRules.rules.length }}</p>
 <p class="mt-3 text-[10px] text-muted-foreground">{{ cacheRules.zone }}</p>
            </div>
          </section>

 <div v-if="cacheRules.rules.length === 0" class="border border-dashed p-8 text-center text-sm text-muted-foreground">
            该 Zone 没有 Cache Rules
          </div>
 <section v-else class="divide-y border bg-card">
 <article v-for="rule in cacheRules.rules" :key="rule.id" class="grid gap-4 px-4 py-4 lg:grid-cols-[minmax(0,1fr)_minmax(10rem,0.7fr)_auto] lg:items-center">
 <div class="min-w-0">
 <div class="flex flex-wrap items-center gap-2">
 <p class="truncate text-sm font-black">{{ rule.description || rule.id }}</p>
                  <AdminStatusBadge :tone="cacheRuleOriginTone(rule.origin_cache_control_status)">
                    {{ cacheOriginPolicyLabel(rule.origin_cache_control_status) }}
                  </AdminStatusBadge>
                </div>
 <p class="mt-2 truncate font-mono text-[10px] text-muted-foreground" :title="rule.expression">{{ rule.expression }}</p>
              </div>
 <dl class="grid grid-cols-2 gap-3 text-xs">
                <div>
 <dt class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">动作</dt>
 <dd class="mt-1 font-bold">{{ rule.action || '-'}}</dd>
                </div>
                <div>
 <dt class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">Edge TTL</dt>
 <dd class="mt-1 font-bold">{{ rule.edge_ttl_mode || '-'}}</dd>
                </div>
              </dl>
 <div class="flex items-center justify-between gap-3 lg:justify-end">
 <span class="text-xs font-bold">{{ rule.enabled ? '已启用': '已停用'}}</span>
                <Switch
                  size="sm"
                  :model-value="rule.enabled"
                  :disabled="!canManage || cacheMutationRuleID === rule.id"
                  :aria-label="`${rule.enabled ? '停用' : '启用'}缓存规则 ${rule.description || rule.id}`"
                  @update:model-value="requestRuleEnabledChange(rule, Boolean($event))"
                />
              </div>
            </article>
          </section>
        </div>
      </TabsContent>

 <TabsContent value="status" class="mt-3">
 <div v-if="connections.length === 0" class="border border-dashed p-8 text-center text-sm text-muted-foreground">
          暂无连接状态记录
        </div>
 <section v-else class="divide-y border bg-card">
 <div v-for="connection in connections" :key="`status-${connection.id}`" class="grid gap-2 px-4 py-3 text-xs sm:grid-cols-[minmax(0,1fr)_auto_auto] sm:items-center">
 <div class="min-w-0">
 <p class="truncate font-bold">{{ connection.name }}</p>
 <p class="mt-1 truncate text-[10px] text-muted-foreground">{{ connection.last_error || '无最近错误'}}</p>
            </div>
            <span>{{ formatDate(connection.last_tested_at) }}</span>
            <AdminStatusBadge :tone="testTone(connection.last_test_status)">
              {{ testLabel(connection.last_test_status) }}
            </AdminStatusBadge>
          </div>
        </section>
      </TabsContent>
    </Tabs>

    <AdminConfirmDialog
      :open="Boolean(pendingRuleChange)"
      title="变更 Cloudflare Cache Rule"
      :description="pendingRuleChangeDescription"
      confirm-label="应用变更"
      :destructive="Boolean(pendingRuleChange?.rule.enabled && !pendingRuleChange?.enabled)"
      @update:open="handleRuleChangeDialog"
      @confirm="applyPendingRuleChange"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { Activity, Gauge, Globe2, KeyRound, Link2, LoaderCircle, PlugZap, RefreshCw } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { opsConnectorEnvironmentOptions } from '@/components/admin/ops/opsConnectorBindingForm'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import servicesApi, {
  type CloudflareCacheRule,
  type CloudflareCacheRules,
  type ServiceCenterCloudflare,
  type ServiceCenterCloudflareZone,
} from '@/api/services'
import type { OpsConnector, OpsDomain, OpsEnvironment } from '@/api/ops'
import { readOpsEnvironmentQuery, withOpsEnvironmentQuery } from '@/lib/opsEnvironment'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const canManage = computed(() => authStore.hasPermission('services:manage'))
const loading = ref(false)
const oauthPending = ref(false)
const testingID = ref(0)
const cloudflareTabs = ['connections', 'zones', 'cache', 'status']
const resolveCloudflareTab = (value: unknown): string => {
  const tab = Array.isArray(value) ? value[0] : value
  return typeof tab === 'string' && cloudflareTabs.includes(tab) ? tab : 'connections'
}
const activeTab = ref(resolveCloudflareTab(route.query.tab))
const environmentFilter = ref<OpsEnvironment | ''>(readOpsEnvironmentQuery(route.query.environment))
const cloudflare = ref<ServiceCenterCloudflare | null>(null)
const cacheTargetKey = ref('')
const cacheRules = ref<CloudflareCacheRules | null>(null)
const cacheLoading = ref(false)
const cacheMutationRuleID = ref('')
const pendingRuleChange = ref<{ rule: CloudflareCacheRule; enabled: boolean } | null>(null)

const connections = computed<OpsConnector[]>(() => cloudflare.value?.connections || [])
const zones = computed<ServiceCenterCloudflareZone[]>(() => cloudflare.value?.zones || [])
const cacheTargets = computed(() => zones.value
  .filter((zone): zone is ServiceCenterCloudflareZone & { connector_id: number } => Boolean(zone.connector_id))
  .map((zone) => ({
    key: `${zone.environment}:${zone.connector_id}:${zone.name}`,
    connectorID: zone.connector_id,
    zone: zone.name,
    environment: zone.environment,
  })))
const selectedCacheTarget = computed(() => cacheTargets.value.find((target) => target.key === cacheTargetKey.value) || null)
const summaryItems = computed(() => [
  { label: '连接', value: cloudflare.value?.connection_count || 0, tone: '' },
  { label: '已启用', value: cloudflare.value?.active_connection_count || 0, tone: 'text-emerald-600' },
  { label: '授权', value: cloudflare.value?.credential_configured_count || 0, tone: 'text-primary' },
  { label: 'Zone', value: cloudflare.value?.zone_count || 0, tone: '' },
  { label: '待处理', value: cloudflare.value?.attention_count || 0, tone: 'text-amber-600' },
])
const capabilityItems = computed(() => {
  const configuredConnections = connections.value.filter((connection) => connection.credential_configured).length
  const proxiedDomains = (cloudflare.value?.domains || []).filter((domain) => domain.proxy_mode === 'proxied').length
  return [
    {
      key: 'authorization',
      label: 'API 授权',
      value: configuredConnections ? `${configuredConnections} 个已授权` : '未完成授权',
      detail: connections.value.length ? '查看权限范围与最近测试' : '需要先连接 Cloudflare',
      tab: 'connections',
    },
    {
      key: 'dns',
      label: 'Zone / DNS',
      value: `${zones.value.length} 个 Zone`,
      detail: `${cloudflare.value?.domains.length || 0} 个域名绑定`,
      tab: 'zones',
    },
    {
      key: 'proxy',
      label: '代理 / TLS',
      value: proxiedDomains ? `${proxiedDomains} 个已代理` : '暂无已代理域名',
      detail: (cloudflare.value?.domains.length || 0) ? '进入 Zone / DNS 查看每个域名' : '尚未关联域名',
      tab: 'zones',
    },
    {
      key: 'cache',
      label: 'Cache Rules',
      value: cacheTargets.value.length ? '可按 Zone 查看' : '暂无可用 Zone',
      detail: cacheTargets.value.length ? '进入后读取并管理规则' : '需要先绑定 Cloudflare Zone',
      tab: 'cache',
    },
  ]
})
const pendingRuleChangeDescription = computed(() => {
  if (!pendingRuleChange.value || !selectedCacheTarget.value) return ''
  const { rule, enabled } = pendingRuleChange.value
  return `${selectedCacheTarget.value.zone} / ${rule.description || rule.id}：${rule.enabled ? '已启用' : '已停用'} -> ${enabled ? '已启用' : '已停用'}`
})

const loadCloudflare = async (): Promise<void> => {
  loading.value = true
  try {
    cloudflare.value = await servicesApi.getCloudflare(environmentFilter.value || undefined)
    syncCacheTarget()
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'Cloudflare 服务加载失败')
  } finally {
    loading.value = false
  }
}

const loadCacheRules = async (): Promise<void> => {
  const target = selectedCacheTarget.value
  if (!target) {
    cacheRules.value = null
    return
  }
  cacheLoading.value = true
  try {
    cacheRules.value = await servicesApi.getCloudflareCacheRules(target.connectorID, target.zone)
  } catch (error: any) {
    cacheRules.value = null
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'Cache Rules 加载失败')
  } finally {
    cacheLoading.value = false
  }
}

const syncCacheTarget = (): void => {
  if (cacheTargets.value.some((target) => target.key === cacheTargetKey.value)) return
  cacheTargetKey.value = cacheTargets.value[0]?.key || ''
  cacheRules.value = null
}

const requestRuleEnabledChange = (rule: CloudflareCacheRule, enabled: boolean): void => {
  if (!canManage.value || enabled === rule.enabled) return
  pendingRuleChange.value = { rule, enabled }
}

const handleRuleChangeDialog = (open: boolean): void => {
  if (!open) pendingRuleChange.value = null
}

const applyPendingRuleChange = async (): Promise<void> => {
  const target = selectedCacheTarget.value
  const change = pendingRuleChange.value
  if (!target || !change) return
  cacheMutationRuleID.value = change.rule.id
  try {
    cacheRules.value = await servicesApi.setCloudflareCacheRuleEnabled(
      target.connectorID,
      target.zone,
      change.rule.id,
      change.enabled,
    )
    toast.success(change.enabled ? 'Cache Rule 已启用' : 'Cache Rule 已停用')
    pendingRuleChange.value = null
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'Cache Rule 变更失败')
  } finally {
    cacheMutationRuleID.value = ''
  }
}

const changeEnvironment = (): void => {
  void router.replace({ query: withOpsEnvironmentQuery(route.query, environmentFilter.value) })
  void loadCloudflare()
}

const oauthReturnPath = (): string => {
  const query = new URLSearchParams()
  if (environmentFilter.value) query.set('environment', environmentFilter.value)
  return `/services/cloudflare${query.size ? `?${query.toString()}` : ''}`
}

const startOAuth = async (connectorID?: number): Promise<void> => {
  oauthPending.value = true
  try {
    const result = await servicesApi.startCloudflareOAuth(
      connectorID,
      oauthReturnPath(),
      environmentFilter.value || 'production',
    )
    window.location.assign(result.authorization_url)
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'Cloudflare 授权启动失败')
    oauthPending.value = false
  }
}

const testConnection = async (id: number): Promise<void> => {
  testingID.value = id
  try {
    const result = await servicesApi.testCloudflareConnection(id)
    toast.success(result.message || 'Cloudflare 连接测试成功')
    await loadCloudflare()
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'Cloudflare 连接测试失败')
  } finally {
    testingID.value = 0
  }
}

const handleOAuthReturn = (): void => {
  const query = new URLSearchParams(window.location.search)
  const status = query.get('ops_oauth_status')
  if (!status) return
  const connected = status === 'connected' || status === 'connected_with_warnings'
  const message = query.get('ops_oauth_message') || (connected ? 'Cloudflare 已连接' : 'Cloudflare 授权失败')
  toast[connected ? 'success' : 'error'](message)
  window.history.replaceState({}, document.title, `${window.location.pathname}${window.location.hash}`)
}

const environmentLabel = (value: string): string => (
  opsConnectorEnvironmentOptions.find((option) => option.value === value)?.label || value || '-'
)
const formatDate = (value?: string): string => value ? new Date(value).toLocaleString('zh-CN') : '未测试'
const testLabel = (value: string): string => value === 'success' ? '成功' : value === 'failed' ? '失败' : '未测试'
const connectionStatusLabel = (value: string): string => ({
  active: '已连接',
  pending: '待验证',
  error: '异常',
  disabled: '已停用',
}[value] || value)
const connectionTone = (value: string): AdminStatusTone => (
  value === 'active' ? 'green' : value === 'error' ? 'coral' : value === 'pending' ? 'amber' : 'gray'
)
const testTone = (value: string): AdminStatusTone => value === 'success' ? 'green' : value === 'failed' ? 'coral' : 'gray'
const roleLabel = (value: string): string => ({
  canonical: '主域名',
  alias: '别名',
  admin: '后台',
  redirect: '重定向',
  verification: '验证',
  internal: '内部',
}[value] || value)
const proxyLabel = (value: string): string => ({
  proxied: '已代理',
  dns_only: '仅 DNS',
  unknown: '待确认',
}[value] || value)
const domainStatusLabel = (domain: OpsDomain): string => (
  domain.observed_status === 'drifted' || domain.status === 'drifted'
    ? '有漂移'
    : domain.observed_status === 'error' || domain.status === 'error'
      ? '异常'
      : domain.observed_status === 'matched'
        ? '已匹配'
        : '待检查'
)
const domainTone = (domain: OpsDomain): AdminStatusTone => (
  domain.observed_status === 'drifted' || domain.status === 'drifted'
    ? 'amber'
    : domain.observed_status === 'error' || domain.status === 'error'
      ? 'coral'
      : domain.observed_status === 'matched'
        ? 'green'
        : 'gray'
)
const zoneKey = (zone: ServiceCenterCloudflareZone): string => `${zone.environment}-${zone.connector_id || 0}-${zone.name}`
const cacheOriginPolicyLabel = (value: string): string => ({
  overridden: '覆盖源站',
  respected: '尊重源站',
  no_rules: '无 Cache Rules',
  not_applicable: '不适用',
}[value] || value)
const cacheOriginPolicyTone = (value: string): AdminStatusTone => (
  value === 'overridden' ? 'amber' : value === 'respected' ? 'green' : 'gray'
)
const cacheRuleOriginTone = (value: string): AdminStatusTone => (
  value === 'overridden' ? 'amber' : value === 'respected' ? 'green' : 'gray'
)

watch(() => route.query.environment, (value) => {
  const nextEnvironment = readOpsEnvironmentQuery(value)
  if (nextEnvironment === environmentFilter.value) return
  environmentFilter.value = nextEnvironment
  void loadCloudflare()
})
watch(() => route.query.tab, (value) => {
  const nextTab = resolveCloudflareTab(value)
  if (nextTab !== activeTab.value) activeTab.value = nextTab
})
watch(cacheTargetKey, () => {
  cacheRules.value = null
  if (activeTab.value === 'cache') void loadCacheRules()
})
watch(activeTab, (value) => {
  if (value === 'cache') void loadCacheRules()
})

onMounted(() => {
  handleOAuthReturn()
  void loadCloudflare()
})
</script>
