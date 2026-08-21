<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="服务中心 / 服务总览"
      description="统一查看服务连接、服务器、项目、域名和网络关系；具体配置与发布操作在运维中心完成。"
    >
      <template #actions>
        <select
          v-model="environmentFilter"
          class="h-9 border bg-background px-3 text-sm"
          aria-label="筛选服务绑定环境"
          :disabled="loading"
          @change="changeEnvironment"
        >
          <option value="">全部环境</option>
          <option v-for="option in opsEnvironmentOptions" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
        <Button size="icon" variant="outline" title="刷新服务绑定状态" :disabled="loading" @click="loadOverview">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
        </Button>
      </template>
    </AdminPageHeader>

    <section class="flex flex-col gap-3 border border-dashed border-border/80 bg-muted/20 p-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex items-start gap-3">
        <Link2 class="mt-0.5 size-5 shrink-0 text-primary" />
        <div>
          <p class="text-sm font-black">当前总览：{{ environmentLabel(serviceOverview?.environment) }}</p>
          <p class="mt-1 text-xs text-muted-foreground">
            这里是服务和线上资源的唯一总览入口；运维中心只保留 VPS、项目、域名维护，以及检查、同步和发布动作。
          </p>
        </div>
      </div>
      <AdminStatusBadge :tone="loading ? 'gray' : 'blue'">
        {{ generatedLabel }}
      </AdminStatusBadge>
    </section>

    <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-5">
      <div class="border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">已登记 VPS</p>
        <p class="mt-2 text-2xl font-black">{{ vpsList.length }}</p>
        <p class="mt-1 text-[10px] text-muted-foreground">健康 {{ healthyVPSCount }} · 待检查 {{ vpsAttentionCount }}</p>
      </div>
      <div class="border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">已绑定项目</p>
        <p class="mt-2 text-2xl font-black">{{ projectList.length }}</p>
        <p class="mt-1 text-[10px] text-muted-foreground">健康 {{ healthyProjectCount }} · 待检查 {{ projectAttentionCount }}</p>
      </div>
      <div class="border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">已登记域名</p>
        <p class="mt-2 text-2xl font-black">{{ domainList.length }}</p>
        <p class="mt-1 text-[10px] text-muted-foreground">已匹配 {{ matchedDomainCount }} · 待检查 {{ domainAttentionCount }}</p>
      </div>
      <div class="border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">服务连接</p>
        <p class="mt-2 text-2xl font-black">{{ totalConnectionCount }}</p>
        <p class="mt-1 text-[10px] text-muted-foreground">可用 {{ activeConnectionCount }} · 待处理 {{ connectorAttentionCount }}</p>
      </div>
      <div class="border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">网络记录</p>
        <p class="mt-2 text-2xl font-black">{{ networkCounts?.total || 0 }}</p>
        <p class="mt-1 text-[10px] text-muted-foreground">待验证 {{ networkCounts?.attention || 0 }}</p>
      </div>
    </section>

    <section class="space-y-3">
      <div class="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h2 class="text-base font-black">第三方服务</h2>
          <p class="mt-1 text-xs text-muted-foreground">这里展示实际连接状态，不把“登记资源”误认为“已经授权”。</p>
        </div>
      </div>

      <div class="grid gap-3 md:grid-cols-2">
        <button
          v-for="provider in providers"
          :key="provider.id"
          type="button"
          class="group border bg-card p-4 text-left transition-colors"
          :class="provider.route ? 'hover:border-primary/60 hover:bg-muted/20' : 'cursor-default opacity-80'"
          :disabled="!provider.route"
          @click="openProvider(provider)"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="flex min-w-0 items-center gap-3">
              <div class="flex size-10 shrink-0 items-center justify-center border bg-muted/40">
                <component :is="providerIcon(provider.id)" class="size-5" />
              </div>
              <div class="min-w-0">
                <p class="truncate text-sm font-black">{{ provider.label }}</p>
                <p class="mt-1 text-[10px] text-muted-foreground">
                  {{ providerRouteLabel(provider) }}
                </p>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <AdminStatusBadge :tone="providerTone(provider.status)">
                {{ statusLabel(provider.status) }}
              </AdminStatusBadge>
              <ChevronRight v-if="provider.route" class="size-4 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
            </div>
          </div>

          <div class="mt-5 grid grid-cols-3 gap-3 border-t border-dashed pt-3">
            <div>
              <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">连接</p>
              <p class="mt-1 text-lg font-black">{{ provider.connection_count }}</p>
            </div>
            <div>
              <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">可用</p>
              <p class="mt-1 text-lg font-black text-emerald-600">{{ provider.active_connection_count }}</p>
            </div>
            <div>
              <p class="text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">关联资源</p>
              <p class="mt-1 text-lg font-black">{{ providerResourceCount(provider) }}</p>
            </div>
          </div>
        </button>
      </div>
    </section>

    <section class="border bg-card">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-dashed px-4 py-3">
        <div>
          <h2 class="text-base font-black">当前绑定资产</h2>
          <p class="mt-1 text-xs text-muted-foreground">按 VPS → 项目 → 域名查看实际关系。</p>
        </div>
        <div class="flex items-center gap-1">
          <Button variant="ghost" size="sm" @click="navigateTo('/ops/vps')">
            VPS 维护
            <ChevronRight class="size-4" />
          </Button>
          <Button variant="ghost" size="sm" @click="navigateTo('/ops/deployments', { tab: 'overview' })">
            发布中心
            <ChevronRight class="size-4" />
          </Button>
        </div>
      </div>

      <div class="grid gap-4 p-4 xl:grid-cols-[1.15fr_1fr_1fr]">
        <section>
          <div class="mb-3 flex items-center justify-between gap-2">
            <div class="flex items-center gap-2">
              <Server class="size-4 text-blue-700" />
              <h3 class="text-sm font-black">VPS</h3>
            </div>
            <Button variant="ghost" size="icon" title="查看 VPS 中心" @click="navigateTo('/ops/vps')">
              <ChevronRight class="size-4" />
            </Button>
          </div>
          <div v-if="vpsList.length === 0" class="border border-dashed p-5 text-center text-xs text-muted-foreground">
            当前环境没有登记 VPS
          </div>
          <div v-else class="space-y-2">
            <button
              v-for="vps in vpsList"
              :key="vps.id"
              type="button"
              class="flex w-full items-start justify-between gap-3 border border-dashed p-3 text-left hover:border-primary/50 hover:bg-muted/20"
              @click="navigateTo('/ops/vps')"
            >
              <span class="min-w-0">
                <span class="block truncate text-xs font-black">{{ vps.name }}</span>
                <span class="mt-1 block truncate font-mono text-[10px] text-muted-foreground">
                  {{ vps.hostname || vps.ipv4 || '未登记地址' }}
                </span>
                <span class="mt-1 block text-[10px] text-muted-foreground">
                  {{ providerLabel(vps.provider) }} · {{ projectCountForVPS(vps.id) }} 个项目
                </span>
              </span>
              <AdminStatusBadge class="shrink-0" :tone="observedTone(vps.observed_status)">
                {{ observedLabel(vps.observed_status) }}
              </AdminStatusBadge>
            </button>
          </div>
        </section>

        <section>
          <div class="mb-3 flex items-center justify-between gap-2">
            <div class="flex items-center gap-2">
              <Workflow class="size-4 text-emerald-700" />
              <h3 class="text-sm font-black">项目</h3>
            </div>
            <Button variant="ghost" size="icon" title="查看项目中心" @click="navigateTo('/ops/projects')">
              <ChevronRight class="size-4" />
            </Button>
          </div>
          <div v-if="projectList.length === 0" class="border border-dashed p-5 text-center text-xs text-muted-foreground">
            当前环境没有绑定项目
          </div>
          <div v-else class="space-y-2">
            <button
              v-for="project in projectList"
              :key="project.id"
              type="button"
              class="flex w-full items-start justify-between gap-3 border border-dashed p-3 text-left hover:border-primary/50 hover:bg-muted/20"
              @click="navigateTo('/ops/projects')"
            >
              <span class="min-w-0">
                <span class="block truncate text-xs font-black">{{ project.name }}</span>
                <span class="mt-1 block truncate text-[10px] text-muted-foreground">
                  VPS：{{ project.vps_name || '未绑定' }}
                </span>
                <span class="mt-1 block truncate font-mono text-[10px] text-muted-foreground">
                  {{ project.compose_project_name || '未登记 Compose 项目' }}
                </span>
              </span>
              <AdminStatusBadge class="shrink-0" :tone="healthTone(project.health_status)">
                {{ healthLabel(project.health_status) }}
              </AdminStatusBadge>
            </button>
          </div>
        </section>

        <section>
          <div class="mb-3 flex items-center justify-between gap-2">
            <div class="flex items-center gap-2">
              <Globe2 class="size-4 text-amber-700" />
              <h3 class="text-sm font-black">域名 / DNS</h3>
            </div>
            <Button variant="ghost" size="icon" title="查看域名中心" @click="navigateTo('/ops/domains')">
              <ChevronRight class="size-4" />
            </Button>
          </div>
          <div v-if="domainList.length === 0" class="border border-dashed p-5 text-center text-xs text-muted-foreground">
            当前环境没有登记域名
          </div>
          <div v-else class="space-y-2">
            <button
              v-for="domain in domainList"
              :key="domain.id"
              type="button"
              class="flex w-full items-start justify-between gap-3 border border-dashed p-3 text-left hover:border-primary/50 hover:bg-muted/20"
              @click="navigateTo('/ops/domains')"
            >
              <span class="min-w-0">
                <span class="block truncate text-xs font-black">{{ domain.domain }}</span>
                <span class="mt-1 block truncate font-mono text-[10px] text-muted-foreground">
                  {{ domain.zone || '未登记 Zone' }} · {{ domain.target || '未登记目标' }}
                </span>
                <span class="mt-1 block text-[10px] text-muted-foreground">
                  {{ providerLabel(domain.provider) }} · {{ proxyLabel(domain.proxy_mode) }} · {{ tlsLabel(domain.tls_mode) }}
                </span>
              </span>
              <AdminStatusBadge class="shrink-0" :tone="domainTone(domain)">
                {{ domainStatusLabel(domain) }}
              </AdminStatusBadge>
            </button>
          </div>
        </section>
      </div>
    </section>

    <section class="flex flex-col gap-3 border border-dashed border-primary/30 bg-primary/5 p-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex items-start gap-3">
        <Cloud class="mt-0.5 size-5 shrink-0 text-primary" />
        <div>
          <p class="text-sm font-black">Cloudflare 能力详情</p>
          <p class="mt-1 text-xs text-muted-foreground">
            {{ cloudflareSummary }}
          </p>
        </div>
      </div>
      <Button variant="outline" @click="navigateTo('/services/cloudflare', { tab: cloudflareDomains.length ? 'zones' : 'connections' })">
        查看 Cloudflare
        <ChevronRight class="size-4" />
      </Button>
    </section>

    <section class="border bg-card">
      <div class="flex flex-wrap items-center justify-between gap-3 border-b border-dashed px-4 py-3">
        <div class="flex items-start gap-3">
          <ShieldAlert class="mt-0.5 size-5 shrink-0 text-amber-700" />
          <div>
            <h2 class="text-base font-black">网络与访问记录</h2>
            <p class="mt-1 text-xs text-muted-foreground">端口、防火墙、DNS 和边缘代理按实际归属汇总。</p>
          </div>
        </div>
        <AdminStatusBadge :tone="networkCounts?.attention ? 'amber' : 'gray'">
          {{ networkCounts?.attention ? `${networkCounts.attention} 项待核验` : '暂无待核验项' }}
        </AdminStatusBadge>
      </div>

      <div class="grid gap-px border-b bg-border sm:grid-cols-4">
        <div class="bg-card p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">声明规则</p>
          <p class="mt-2 text-2xl font-black">{{ networkCounts?.explicit_rule_count || 0 }}</p>
        </div>
        <div class="bg-card p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">DNS / 边缘派生</p>
          <p class="mt-2 text-2xl font-black">{{ networkCounts?.inferred_item_count || 0 }}</p>
        </div>
        <div class="bg-card p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">关联 VPS</p>
          <p class="mt-2 text-2xl font-black">{{ networkCounts?.vps_count || 0 }}</p>
        </div>
        <div class="bg-card p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">待验证</p>
          <p class="mt-2 text-2xl font-black">{{ networkCounts?.unknown || 0 }}</p>
        </div>
      </div>

      <div v-if="networkItems.length === 0" class="p-6 text-center text-xs text-muted-foreground">
        当前环境没有端口、防火墙或 DNS / 边缘记录
      </div>
      <div v-else class="grid gap-3 p-4 lg:grid-cols-2">
        <button
          v-for="item in networkItems"
          :key="item.key"
          type="button"
          class="group flex min-w-0 flex-col gap-3 border border-dashed p-3 text-left transition-colors hover:border-primary/50 hover:bg-muted/20"
          @click="openNetworkItem(item)"
        >
          <div class="flex min-w-0 items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="truncate text-xs font-black">{{ item.name }}</p>
              <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground">
                {{ networkTargetLabel(item) }}
              </p>
            </div>
            <div class="flex shrink-0 items-center gap-2">
              <AdminStatusBadge :tone="networkStateTone(item)">
                {{ networkStateLabel(item) }}
              </AdminStatusBadge>
              <ChevronRight class="size-4 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
            </div>
          </div>

          <div class="flex flex-wrap gap-1.5 text-[10px]">
            <span class="border bg-muted/30 px-2 py-1 font-black">{{ networkKindLabel(item.kind) }}</span>
            <span class="border bg-muted/30 px-2 py-1">{{ networkScopeLabel(item.scope) }}</span>
            <span class="border bg-muted/30 px-2 py-1">归属：{{ networkOwnerLabel(item) }}</span>
            <span class="border bg-muted/30 px-2 py-1">管理：{{ networkManagerLabel(item.managed_by) }}</span>
          </div>

          <div class="grid gap-1 text-[10px] text-muted-foreground">
            <p v-if="item.vps_name" class="truncate">VPS：{{ item.vps_name }}</p>
            <p v-if="item.project_name" class="truncate">项目：{{ item.project_name }}</p>
            <p v-if="item.domain_name" class="truncate">域名：{{ item.domain_name }}</p>
            <p v-if="item.connector_name" class="truncate">服务连接：{{ item.connector_name }}</p>
          </div>
        </button>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { ChevronRight, Cloud, GitBranch, Globe2, Link2, Package, RefreshCw, Server, ShieldAlert, Workflow } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import servicesApi, {
  type ServiceCenterOverview,
  type ServiceCenterProvider,
} from '@/api/services'
import type {
  OpsDomain,
  OpsEnvironment,
  OpsNetworkSummary,
  OpsNetworkSummaryItem,
  OpsProject,
  OpsVPS,
} from '@/api/ops'
import {
  opsEnvironmentOptions,
  readOpsEnvironmentQuery,
  withOpsEnvironmentQuery,
} from '@/lib/opsEnvironment'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const serviceOverview = ref<ServiceCenterOverview | null>(null)
const environmentFilter = ref<OpsEnvironment | ''>(readOpsEnvironmentQuery(route.query.environment))

const providers = computed<ServiceCenterProvider[]>(() => serviceOverview.value?.providers || [])
const vpsList = computed<OpsVPS[]>(() => serviceOverview.value?.assets?.vps || [])
const projectList = computed<OpsProject[]>(() => serviceOverview.value?.assets?.projects || [])
const domainList = computed<OpsDomain[]>(() => serviceOverview.value?.assets?.domains || [])
const networkSummary = computed<OpsNetworkSummary | null>(() => serviceOverview.value?.network || null)
const networkCounts = computed(() => networkSummary.value?.summary)
const networkItems = computed<OpsNetworkSummaryItem[]>(() => networkSummary.value?.items || [])
const healthyVPSCount = computed(() => vpsList.value.filter((vps) => vps.observed_status === 'healthy').length)
const vpsAttentionCount = computed(() => vpsList.value.filter((vps) => (
  ['pending', 'drifted', 'error'].includes(vps.status)
  || ['unknown', 'degraded', 'offline'].includes(vps.observed_status)
)).length)
const healthyProjectCount = computed(() => projectList.value.filter((project) => project.health_status === 'healthy').length)
const projectAttentionCount = computed(() => projectList.value.filter((project) => (
  ['pending', 'drifted', 'error'].includes(project.status)
  || ['unknown', 'degraded', 'offline'].includes(project.health_status)
)).length)
const matchedDomainCount = computed(() => domainList.value.filter((domain) => domain.observed_status === 'matched').length)
const domainAttentionCount = computed(() => domainList.value.filter((domain) => (
  ['pending', 'drifted', 'error'].includes(domain.status)
  || ['unknown', 'drifted', 'error'].includes(domain.observed_status)
)).length)
const totalConnectionCount = computed(() => providers.value.reduce((total, provider) => total + provider.connection_count, 0))
const activeConnectionCount = computed(() => providers.value.reduce((total, provider) => total + provider.active_connection_count, 0))
const connectorAttentionCount = computed(() => providers.value
  .filter((provider) => ['attention', 'pending'].includes(provider.status))
  .reduce((total, provider) => total + provider.connection_count, 0))
const cloudflareDomains = computed(() => domainList.value.filter((domain) => domain.provider === 'cloudflare'))
const cloudflareProvider = computed(() => providers.value.find((provider) => provider.id === 'cloudflare'))
const generatedLabel = computed(() => serviceOverview.value?.generated_at ? formatDate(serviceOverview.value.generated_at) : '尚未生成')
const cloudflareSummary = computed(() => {
  const provider = cloudflareProvider.value
  if (!provider || provider.status === 'not_connected') {
    return '当前没有可用的 Cloudflare 连接。需要授权或检查连接状态时，进入 Cloudflare 详情页处理。'
  }
  if (cloudflareDomains.value.length === 0) {
    return `Cloudflare 已有 ${provider.active_connection_count} 个可用连接，但当前没有关联 Zone 或域名。`
  }
  return `当前关联 ${cloudflareDomains.value.length} 个域名；DNS、代理/TLS 和 Cache Rules 的实际能力在 Cloudflare 详情页按 Zone 查看。`
})

const loadOverview = async (): Promise<void> => {
  loading.value = true
  try {
    serviceOverview.value = await servicesApi.getOverview(environmentFilter.value || undefined)
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '服务中心加载失败')
  } finally {
    loading.value = false
  }
}

const changeEnvironment = (): void => {
  void router.replace({ query: withOpsEnvironmentQuery(route.query, environmentFilter.value) })
  void loadOverview()
}

const navigateTo = (path: string, extraQuery: Record<string, string> = {}): void => {
  const query: Record<string, string> = { ...extraQuery }
  if (environmentFilter.value) query.environment = environmentFilter.value
  void router.push({ path, query })
}

const openProvider = (provider: ServiceCenterProvider): void => {
  if (provider.id === 'github' || provider.id === 'ghcr') {
    navigateTo('/ops/deployments', { tab: 'github' })
    return
  }
  if (provider.route) navigateTo(provider.route)
}

const environmentLabel = (value?: string): string => {
  if (!value) return '全部环境'
  return opsEnvironmentOptions.find((option) => option.value === value)?.label || value
}

const providerIcon = (provider: string) => ({
  cloudflare: Cloud,
  hostinger: Server,
  github: GitBranch,
  ghcr: Package,
}[provider] || Cloud)

const providerRouteLabel = (provider: ServiceCenterProvider): string => ({
  cloudflare: '进入 Cloudflare 能力详情',
  hostinger: '进入 VPS 中心查看资源',
  github: '进入发布中心查看 GitHub / GHCR',
  ghcr: '进入发布中心查看 GitHub / GHCR',
}[provider.id] || (provider.route ? '查看详情' : '暂未提供详情页'))

const statusLabel = (status: ServiceCenterProvider['status']): string => ({
  active: '已连接',
  attention: '需要处理',
  pending: '待验证',
  not_connected: '未连接',
  not_configured: '未接入',
}[status] || status)

const providerTone = (status: ServiceCenterProvider['status']): AdminStatusTone => (
  status === 'active' ? 'green' : status === 'attention' ? 'coral' : status === 'pending' ? 'amber' : 'gray'
)

const providerResourceCount = (provider: ServiceCenterProvider): number => {
  if (provider.id === 'hostinger') {
    return vpsList.value.filter((vps) => vps.provider === 'hostinger').length
  }
  return provider.resource_count
}

const openNetworkItem = (item: OpsNetworkSummaryItem): void => {
  if (item.managed_by === 'cloudflare') {
    navigateTo('/services/cloudflare', { tab: 'zones' })
    return
  }
  if (item.vps_binding_id || item.owner_kind === 'vps') {
    navigateTo('/ops/vps')
    return
  }
  if (item.project_binding_id || item.owner_kind === 'project') {
    navigateTo('/ops/projects')
    return
  }
  if (item.domain_binding_id || item.owner_kind === 'domain') {
    navigateTo('/ops/domains')
    return
  }
  navigateTo('/ops/vps')
}

const providerLabel = (value: string): string => ({
  hostinger: 'Hostinger',
  cloudflare: 'Cloudflare',
  github: 'GitHub',
  ghcr: 'GHCR',
  other: '其他 Provider',
}[value] || value || '未指定 Provider')

const networkKindLabel = (value: string): string => ({
  rule: '声明规则',
  domain_dns: 'DNS / 边缘派生',
}[value] || value || '网络记录')

const networkScopeLabel = (value: string): string => ({
  os_firewall: '主机防火墙',
  gateway: '共享网关',
  edge: '边缘代理',
  dns: 'DNS',
  origin: '源站',
}[value] || value || '未分类')

const networkManagerLabel = (value: string): string => ({
  cloudflare: 'Cloudflare',
  hostinger: 'Hostinger',
  os_firewall: '主机防火墙',
  manual: '手动登记',
  other: '其他',
}[value] || value || '未指定')

const networkOwnerLabel = (item: OpsNetworkSummaryItem): string => (
  item.owner_name
  || item.vps_name
  || item.project_name
  || item.domain_name
  || item.connector_name
  || ({
    vps: 'VPS',
    project: '项目',
    domain: '域名',
    connector: '服务连接',
    manual: '手动登记',
  }[item.owner_kind] || '未指定')
)

const networkTargetLabel = (item: OpsNetworkSummaryItem): string => {
  const protocol = item.protocol ? item.protocol.toUpperCase() : ''
  if (item.ports) return [protocol, item.ports].filter(Boolean).join(' ')
  if (item.target) return item.target
  return networkScopeLabel(item.scope)
}

const networkStateLabel = (item: OpsNetworkSummaryItem): string => {
  if (item.observed_state === 'error' || item.effective_state === 'error') return '检查异常'
  if (item.observed_state === 'drifted' || item.effective_state === 'drifted') return '发现漂移'
  if (item.observed_state === 'unknown' || item.effective_state === 'unknown') return '待验证'
  if (item.observed_state === 'closed' || item.effective_state === 'closed') return '已关闭'
  if (item.observed_state === 'restricted' || item.effective_state === 'restricted') return '受限'
  return item.kind === 'domain_dns' ? '已登记' : '已确认'
}

const networkStateTone = (item: OpsNetworkSummaryItem): AdminStatusTone => {
  if (item.observed_state === 'error' || item.effective_state === 'error') return 'coral'
  if (item.observed_state === 'drifted' || item.effective_state === 'drifted') return 'amber'
  if (item.observed_state === 'unknown' || item.effective_state === 'unknown') return 'amber'
  if (item.observed_state === 'closed' || item.effective_state === 'closed') return 'gray'
  return 'green'
}

const projectCountForVPS = (vpsID: number): number => projectList.value.filter((project) => project.vps_binding_id === vpsID).length

const observedLabel = (value: string): string => ({
  healthy: '健康',
  degraded: '降级',
  unknown: '未同步',
  offline: '离线',
}[value] || value || '未同步')

const healthLabel = (value: string): string => observedLabel(value)

const observedTone = (value: string): AdminStatusTone => (
  value === 'healthy' ? 'green' : value === 'degraded' || value === 'unknown' ? 'amber' : value === 'offline' ? 'coral' : 'gray'
)

const healthTone = (value: string): AdminStatusTone => observedTone(value)

const proxyLabel = (value: string): string => ({
  proxied: '已代理',
  dns_only: '仅 DNS',
  unknown: '代理待确认',
}[value] || value || '代理待确认')

const tlsLabel = (value: string): string => ({
  full_strict: 'TLS Full Strict',
  full: 'TLS Full',
  flexible: 'TLS Flexible',
  off: 'TLS 关闭',
  unknown: 'TLS 待确认',
}[value] || value || 'TLS 待确认')

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

const formatDate = (value?: string): string => value ? new Date(value).toLocaleString('zh-CN') : '-'

watch(() => route.query.environment, (value) => {
  const nextEnvironment = readOpsEnvironmentQuery(value)
  if (nextEnvironment === environmentFilter.value) return
  environmentFilter.value = nextEnvironment
  void loadOverview()
})

onMounted(() => {
  void loadOverview()
})
</script>
