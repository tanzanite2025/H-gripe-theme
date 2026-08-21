<template>
  <section class="space-y-4">
    <div class="flex flex-col gap-3 border border-dashed border-border/80 bg-muted/20 p-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex items-start gap-3">
        <GitBranch class="mt-0.5 size-5 shrink-0 text-foreground" />
        <div>
          <p class="text-sm font-black">GitHub / GHCR 对接</p>
          <p class="mt-1 text-xs text-muted-foreground">
            查看仓库读取、镜像读取和发布引用所依赖的连接器状态。
          </p>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <Button
          size="sm"
          :disabled="githubLoading || githubOAuthLoading"
          @click="emit('connect')"
        >
          <LoaderCircle v-if="githubOAuthLoading" class="size-4 animate-spin" />
          <GitBranch v-else class="size-4" />
          连接 GitHub
        </Button>
        <Button variant="outline" size="sm" :disabled="githubLoading" @click="emit('refresh')">
          <RefreshCw :class="['size-4', githubLoading ? 'animate-spin' : '']" />
          刷新连接
        </Button>
        <Button variant="ghost" size="sm" @click="emit('configure')">
          连接器配置
          <ExternalLink class="size-4" />
        </Button>
      </div>
    </div>

    <div v-if="githubLoading" class="border border-dashed p-8 text-center text-xs text-muted-foreground">
      正在读取 GitHub / GHCR 连接状态
    </div>
    <div v-else class="grid gap-3 lg:grid-cols-2">
      <Card v-for="provider in githubProviderCards" :key="provider.value" size="sm">
        <CardHeader class="border-b border-dashed border-border/70">
          <div class="flex items-start justify-between gap-3">
            <div class="flex items-center gap-3">
              <span class="flex size-9 items-center justify-center border bg-muted/40">
                <component :is="provider.icon" class="size-4" />
              </span>
              <div>
                <CardTitle>{{ provider.label }}</CardTitle>
                <CardDescription>{{ provider.description }}</CardDescription>
              </div>
            </div>
            <AdminStatusBadge :tone="providerTone(provider.status)">
              {{ connectorStatusLabel(provider.status) }}
            </AdminStatusBadge>
          </div>
        </CardHeader>
        <CardContent class="space-y-3 pt-3">
          <div class="grid grid-cols-3 gap-2 text-center">
            <div class="border border-dashed p-2">
              <p class="text-[10px] text-muted-foreground">连接</p>
              <p class="mt-1 text-lg font-black">{{ provider.connectors.length }}</p>
            </div>
            <div class="border border-dashed p-2">
              <p class="text-[10px] text-muted-foreground">凭据</p>
              <p class="mt-1 text-lg font-black">{{ provider.configuredCount }}</p>
            </div>
            <div class="border border-dashed p-2">
              <p class="text-[10px] text-muted-foreground">测试通过</p>
              <p class="mt-1 text-lg font-black text-emerald-600">{{ provider.testedCount }}</p>
            </div>
          </div>

          <div v-if="!provider.connectors.length" class="border border-dashed p-5 text-center text-xs text-muted-foreground">
            尚未登记 {{ provider.label }} 连接器
          </div>
          <div v-else class="space-y-2">
            <div
              v-for="connector in provider.connectors"
              :key="connector.id"
              class="border border-dashed p-3"
            >
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <p class="truncate text-xs font-black">{{ connector.name }}</p>
                  <p class="mt-1 truncate text-[10px] text-muted-foreground">
                    {{ environmentLabel(connector.environment) }} · {{ connector.endpoint || defaultConnectorEndpoint(connector.provider) }}
                  </p>
                </div>
                <AdminStatusBadge :tone="connectorTone(connector)">
                  {{ connectorStatusLabel(connector.status) }}
                </AdminStatusBadge>
              </div>
              <div class="mt-3 grid gap-2 text-[10px] text-muted-foreground sm:grid-cols-2">
                <p>
                  凭据：
                  <span :class="connector.credential_configured ? 'text-emerald-700' : 'text-amber-700'">
                    {{ connector.credential_configured ? '已配置' : '未配置' }}
                  </span>
                </p>
                <p>
                  测试：
                  <span :class="connector.last_test_status === 'success' ? 'text-emerald-700' : 'text-amber-700'">
                    {{ testStatusLabel(connector.last_test_status) }}
                  </span>
                </p>
                <p class="sm:col-span-2">
                  作用域：
                  <span class="font-mono">{{ connector.scopes || '未登记' }}</span>
                </p>
              </div>
              <div class="mt-3 flex flex-wrap items-center justify-between gap-2 border-t border-dashed pt-2">
                <AdminStatusBadge :tone="scopeTone(connector, provider.value)">
                  {{ scopeLabel(connector, provider.value) }}
                </AdminStatusBadge>
                <Button
                  size="icon"
                  variant="ghost"
                  :title="`测试 ${connector.name}`"
                  :disabled="testingConnectorId === connector.id"
                  @click="emit('test', connector)"
                >
                  <LoaderCircle v-if="testingConnectorId === connector.id" class="size-4 animate-spin" />
                  <PlugZap v-else class="size-4" />
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>

    <Card size="sm">
      <CardHeader class="border-b border-dashed border-border/70">
        <CardTitle>项目发布引用</CardTitle>
        <CardDescription>当前环境项目台账中的 Compose 来源、镜像标签和 Commit。</CardDescription>
      </CardHeader>
      <CardContent class="pt-3">
        <div v-if="projects.length === 0" class="py-6 text-center text-xs text-muted-foreground">
          当前环境没有可展示的项目发布引用
        </div>
        <div v-else class="space-y-2">
          <div
            v-for="project in projects"
            :key="project.id"
            class="grid gap-2 border-b border-dashed border-border/60 py-3 last:border-b-0 sm:grid-cols-[1fr_1.4fr_1fr_1fr]"
          >
            <div class="min-w-0">
              <p class="truncate text-xs font-black">{{ project.name }}</p>
              <p class="mt-1 text-[10px] text-muted-foreground">{{ environmentLabel(project.environment) }}</p>
            </div>
            <p class="truncate font-mono text-[10px] text-muted-foreground" :title="project.compose_source">
              {{ project.compose_source || '未登记仓库 / Compose 来源' }}
            </p>
            <p class="truncate font-mono text-[10px] text-muted-foreground">
              镜像 {{ project.current_image_tag || '-' }}
            </p>
            <p class="truncate font-mono text-[10px] text-muted-foreground">
              Commit {{ project.current_commit_sha || '-' }}
            </p>
          </div>
        </div>
      </CardContent>
    </Card>
  </section>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  ExternalLink,
  GitBranch,
  LoaderCircle,
  Package,
  PlugZap,
  RefreshCw,
} from '@lucide/vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { environmentLabel } from '@/lib/deploymentPreflightPresentation'
import type { OpsConnector, OpsProject } from '@/api/ops'

const props = defineProps<{
  connectors: readonly OpsConnector[]
  projects: readonly OpsProject[]
  githubLoading: boolean
  githubOAuthLoading: boolean
  testingConnectorId: number
}>()

const emit = defineEmits<{
  connect: []
  refresh: []
  configure: []
  test: [connector: OpsConnector]
}>()

const githubProviderCards = computed(() => [
  {
    value: 'github',
    label: 'GitHub',
    description: '仓库、Commit 和发布引用',
    icon: GitBranch,
    connectors: props.connectors.filter((connector) => connector.provider === 'github'),
  },
  {
    value: 'ghcr',
    label: 'GHCR',
    description: '容器镜像和 packages:read',
    icon: Package,
    connectors: props.connectors.filter((connector) => connector.provider === 'ghcr'),
  },
].map((provider) => ({
  ...provider,
  configuredCount: provider.connectors.filter((connector) => connector.credential_configured).length,
  testedCount: provider.connectors.filter((connector) => connector.last_test_status === 'success').length,
  status: connectorGroupStatus(provider.connectors),
})))

const connectorGroupStatus = (connectors: readonly OpsConnector[]): string => {
  if (!connectors.length) return 'not_connected'
  if (connectors.some((connector) => connector.enabled && connector.status === 'active')) {
    return 'active'
  }
  if (connectors.some((connector) => connector.status === 'error')) return 'attention'
  return 'pending'
}

const connectorStatusLabel = (value: string): string => ({
  active: '已连接',
  pending: '待验证',
  error: '需要处理',
  disabled: '已停用',
  not_connected: '未连接',
}[value] || value || '未登记')

const providerTone = (value: string): AdminStatusTone => {
  if (value === 'active') return 'green'
  if (value === 'attention' || value === 'error') return 'coral'
  if (value === 'pending') return 'amber'
  return 'gray'
}

const connectorTone = (connector: OpsConnector): AdminStatusTone => {
  if (!connector.enabled || connector.status === 'disabled') return 'gray'
  if (connector.status === 'error' || connector.last_test_status === 'failed') return 'coral'
  if (connector.status === 'active' && connector.credential_configured) return 'green'
  return 'amber'
}

const testStatusLabel = (value: string): string =>
  value === 'success' ? '已通过' : value === 'failed' ? '失败' : '未测试'

const defaultConnectorEndpoint = (provider: string): string =>
  provider === 'github' || provider === 'ghcr'
    ? 'https://api.github.com/user'
    : '未设置测试接口'

const hasPackageReadScope = (value: string): boolean => {
  const scopes = new Set(value.toLowerCase().split(/[,\s]+/).filter(Boolean))
  return scopes.has('packages:read') || scopes.has('read:packages')
}

const hasRepoReadScope = (value: string): boolean => {
  const scopes = new Set(value.toLowerCase().split(/[,\s]+/).filter(Boolean))
  return scopes.has('repo') || scopes.has('repo:read') || scopes.has('contents:read')
}

const scopeLabel = (connector: OpsConnector, provider: string): string => {
  if (!connector.scopes.trim()) return '未登记作用域'
  if (provider === 'ghcr') return hasPackageReadScope(connector.scopes) ? '具备镜像读取作用域' : '缺少 packages:read'
  return hasRepoReadScope(connector.scopes) ? '具备仓库读取作用域' : '缺少 repo:read'
}

const scopeTone = (connector: OpsConnector, provider: string): AdminStatusTone => {
  if (!connector.scopes.trim()) return 'amber'
  return provider === 'ghcr'
    ? hasPackageReadScope(connector.scopes) ? 'green' : 'amber'
    : hasRepoReadScope(connector.scopes) ? 'green' : 'amber'
}
</script>
