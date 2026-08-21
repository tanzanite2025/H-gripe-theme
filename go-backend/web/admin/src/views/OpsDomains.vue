<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="运维中心 / 域名中心"
      description="统一维护主域名、别名、后台域、跳转域和验证域的绑定关系。"
    >
      <template #actions>
        <select
          v-model="environmentFilter"
          class="h-9 rounded-md border bg-background px-3 text-sm"
          aria-label="筛选域名环境"
          :disabled="loading || syncingAll"
          :options="opsDomainEnvironmentOptions"
          @update:model-value="changeEnvironment"
        />
        <Button
          v-if="canSync"
          variant="outline"
          :disabled="loading || syncingAll || cloudflareDomains.length === 0"
          @click="syncAllCloudflare"
        >
          <LoaderCircle v-if="syncingAll" class="size-4 animate-spin" />
          <RefreshCw v-else class="size-4" />
          同步 Cloudflare
        </Button>
        <Button variant="outline" :disabled="loading" @click="refreshDomainsPage">
 <RefreshCw :class="['size-4', loading ? 'animate-spin': '']" />
          刷新
        </Button>
        <Button v-if="canEdit" @click="openCreate">
          <Plus class="size-4" />
          新增域名
        </Button>
      </template>
    </AdminPageHeader>

    <OpsSummaryCards :items="summaryCards" />

    <AdminTablePanel :loading="loading" :batch-visible="false">
      <Table class="min-w-[1440px]">
        <TableHeader>
          <TableRow>
            <TableHead>域名</TableHead>
            <TableHead>角色</TableHead>
            <TableHead>环境</TableHead>
            <TableHead>所属项目</TableHead>
            <TableHead>提供商 / Zone</TableHead>
            <TableHead>目标（期望 / 实际）</TableHead>
            <TableHead>代理 / TLS（期望 / 实际）</TableHead>
            <TableHead>状态（期望 / 实际）</TableHead>
            <TableHead class="w-44 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="domains.length === 0" :colspan="9">
            <div class="py-8 text-center text-xs text-muted-foreground">还没有登记域名</div>
          </TableEmpty>
          <TableRow v-for="domain in domains" :key="domain.id">
            <TableCell>
              <div class="min-w-0">
                <p class="truncate font-bold">{{ domain.domain }}</p>
                <p v-if="domain.redirect_target" class="mt-1 truncate text-[10px] text-muted-foreground">
                  跳转至 {{ domain.redirect_target }}
                </p>
              </div>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="roleTone(domain.role)">{{ roleLabel(domain.role) }}</AdminStatusBadge>
            </TableCell>
            <TableCell class="text-xs">{{ environmentLabel(domain.environment) }}</TableCell>
            <TableCell>
              <p class="max-w-40 truncate text-xs font-bold">{{ projectLabel(domain.project_binding_id) }}</p>
            </TableCell>
            <TableCell>
              <p class="text-xs font-bold">{{ providerLabel(domain.provider) }}</p>
 <p class="mt-1 text-[10px] text-muted-foreground">{{ domain.zone || '未登记 Zone'}}</p>
            </TableCell>
            <TableCell class="max-w-56">
 <p class="truncate font-mono text-[10px]" :title="domain.target">期望：{{ domain.target || '-'}}</p>
              <p class="mt-1 truncate font-mono text-[10px] text-muted-foreground" :title="domain.observed_target">
                实际：{{ domain.observed_target || '未同步' }}
              </p>
            </TableCell>
            <TableCell>
              <p class="text-[10px]">期望：{{ proxyLabel(domain.proxy_mode) }} / {{ tlsLabel(domain.tls_mode) }}</p>
              <p class="mt-1 text-[10px] text-muted-foreground">
                实际：{{ proxyLabel(domain.observed_proxy_mode) }} / {{ tlsLabel(domain.observed_tls_mode) }}
              </p>
            </TableCell>
            <TableCell>
              <div class="flex flex-col items-start gap-1">
                <AdminStatusBadge :tone="statusTone(domain.status)">
                  期望：{{ statusLabel(domain.status) }}
                </AdminStatusBadge>
                <AdminStatusBadge :tone="observedStatusTone(domain.observed_status)">
                  实际：{{ observedStatusLabel(domain.observed_status) }}
                </AdminStatusBadge>
                <span v-if="domain.observed_source || domain.last_observed_at" class="text-[9px] text-muted-foreground">
                  {{ domain.observed_source || '未设置来源' }} · {{ formatDate(domain.last_observed_at) }}
                </span>
              </div>
            </TableCell>
            <TableCell class="text-right">
              <div class="flex justify-end gap-1">
                <Button
                  v-if="canSync && domain.provider === 'cloudflare'"
                  size="icon"
                  variant="ghost"
                  :title="`同步 ${domain.domain}`"
                  :disabled="syncingID === domain.id || syncingAll"
                  @click="syncDomain(domain)"
                >
                  <LoaderCircle v-if="syncingID === domain.id" class="size-4 animate-spin" />
                  <RefreshCw v-else class="size-4" />
                </Button>
                <Button
                  size="icon"
                  variant="ghost"
                  :title="`查看 ${domain.domain} 的期望 / 实际差异`"
                  :disabled="diffLoadingID === domain.id"
                  @click="openDiff(domain)"
                >
                  <LoaderCircle v-if="diffLoadingID === domain.id" class="size-4 animate-spin" />
                  <GitCompareArrows v-else class="size-4" />
                </Button>
                <Button
                  size="icon"
                  variant="ghost"
                  :title="`预览 ${domain.domain} 的 DNS / 网关配置`"
                  :disabled="previewLoadingID === domain.id"
                  @click="openPreview(domain)"
                >
                  <LoaderCircle v-if="previewLoadingID === domain.id" class="size-4 animate-spin" />
                  <FileCode2 v-else class="size-4" />
                </Button>
                <Button
                  v-if="canEdit"
                  size="icon"
                  variant="ghost"
                  :title="domain.enabled ? '停用域名' : '启用域名'"
                  @click="toggleDomain(domain)"
                >
                  <Power class="size-4" />
                </Button>
                <Button v-if="canEdit" size="icon" variant="ghost" title="编辑域名" @click="openEdit(domain)">
                  <Pencil class="size-4" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>

    <OpsDomainBindingFormDialog
      :open="dialogOpen"
      :form="form"
      :projects="projects"
      :connectors="connectors"
      :projects-loading="projectsLoading"
      :connectors-loading="connectorsLoading"
      :saving="saving"
      :can-edit="canEdit"
      @update:open="dialogOpen = $event"
      @save="saveDomain"
    />

    <Dialog v-model:open="diffOpen">
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>{{ diff?.domain || '域名状态差异' }}</DialogTitle>
          <DialogDescription>
            只读差异明细，来自后台期望状态和最近一次同步得到的实际状态；不会写入外部平台。
          </DialogDescription>
        </DialogHeader>

        <div v-if="diff" class="space-y-4">
          <div class="rounded-lg border border-dashed border-border/70 p-3">
            <div class="flex flex-wrap items-center gap-2">
              <AdminStatusBadge :tone="diffTone(diff.status)">{{ diffStatusLabel(diff.status) }}</AdminStatusBadge>
              <span class="text-xs text-muted-foreground">{{ diff.summary }}</span>
            </div>
            <div class="mt-2 grid gap-2 text-[10px] text-muted-foreground sm:grid-cols-2">
              <p>来源：{{ diff.observed_source || '未同步' }}</p>
              <p>时间：{{ formatDate(diff.last_observed_at) }}</p>
              <p v-if="diff.observed_error" class="sm:col-span-2 text-rose-600">错误：{{ diff.observed_error }}</p>
            </div>
          </div>

          <div class="overflow-hidden rounded-lg border border-dashed border-border/70">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>项目</TableHead>
                  <TableHead>期望</TableHead>
                  <TableHead>实际</TableHead>
                  <TableHead>状态</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="item in diff.items" :key="item.key">
                  <TableCell class="font-bold">{{ item.label }}</TableCell>
                  <TableCell class="font-mono text-[10px]">{{ item.desired }}</TableCell>
                  <TableCell class="font-mono text-[10px]">{{ item.observed }}</TableCell>
                  <TableCell>
                    <div class="flex flex-col items-start gap-1">
                      <AdminStatusBadge :tone="diffTone(item.status)">
                        {{ diffStatusLabel(item.status) }}
                      </AdminStatusBadge>
                      <span v-if="item.message" class="text-[10px] text-muted-foreground">{{ item.message }}</span>
                    </div>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </div>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="diffOpen = false">关闭</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <Dialog v-model:open="previewOpen">
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>{{ preview?.domain || '域名配置预览' }}</DialogTitle>
          <DialogDescription>
            只读配置草稿，用于部署前检查；不会写入 Cloudflare、Caddy、Nginx 或生产网关。
          </DialogDescription>
        </DialogHeader>

        <div v-if="preview" class="space-y-4">
          <div v-if="preview.warnings.length" class="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3">
            <p class="text-xs font-black text-amber-800">需要人工确认</p>
            <ul class="mt-2 space-y-1 text-xs text-amber-800/90">
              <li v-for="warning in preview.warnings" :key="warning">{{ warning }}</li>
            </ul>
          </div>

          <Tabs v-model="previewTab" class="w-full">
            <TabsList class="h-9 w-full justify-start overflow-x-auto rounded-xl bg-muted/50 p-1">
              <TabsTrigger value="dns" class="min-w-24">DNS</TabsTrigger>
              <TabsTrigger value="caddy" class="min-w-24">Caddy</TabsTrigger>
              <TabsTrigger value="nginx" class="min-w-24">Nginx</TabsTrigger>
            </TabsList>

            <TabsContent value="dns" class="mt-4">
              <div class="grid gap-2 rounded-lg border border-dashed border-border/70 p-3 text-xs sm:grid-cols-2">
                <div><span class="text-muted-foreground">Provider：</span>{{ providerLabel(preview.dns.provider) }}</div>
 <div><span class="text-muted-foreground">Zone：</span>{{ preview.dns.zone || '-'}}</div>
                <div><span class="text-muted-foreground">Type：</span>{{ preview.dns.record_type }}</div>
                <div><span class="text-muted-foreground">Name：</span>{{ preview.dns.name }}</div>
 <div class="sm:col-span-2"><span class="text-muted-foreground">Content：</span>{{ preview.dns.content || '-'}}</div>
                <div><span class="text-muted-foreground">Proxy：</span>{{ proxyLabel(preview.dns.proxy_mode) }}</div>
                <div><span class="text-muted-foreground">TLS：</span>{{ tlsLabel(preview.dns.tls_mode) }}</div>
                <div v-if="preview.dns.redirect" class="sm:col-span-2">
 <span class="text-muted-foreground">Redirect：</span>{{ preview.dns.redirect_target || '-'}}
                </div>
              </div>
            </TabsContent>

            <TabsContent value="caddy" class="mt-4">
              <PreviewCodeBlock
                :filename="preview.caddy.filename"
                :content="preview.caddy.content"
                @copy="copyPreview(preview.caddy.content)"
              />
            </TabsContent>

            <TabsContent value="nginx" class="mt-4">
              <PreviewCodeBlock
                :filename="preview.nginx.filename"
                :content="preview.nginx.content"
                @copy="copyPreview(preview.nginx.content)"
              />
            </TabsContent>
          </Tabs>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="previewOpen = false">关闭</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { toast } from 'vue-sonner'
import { Copy, FileCode2, GitCompareArrows, LoaderCircle, Pencil, Plus, Power, RefreshCw } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import OpsEnvironmentSelect from '@/components/admin/ops/OpsEnvironmentSelect.vue'
import OpsSummaryCards, { type OpsSummaryCardItem } from '@/components/admin/ops/OpsSummaryCards.vue'
import OpsDomainBindingFormDialog from '@/components/admin/ops/OpsDomainBindingFormDialog.vue'
import {
  assignOpsDomainForm,
  domainRoleRequiresProject,
  emptyOpsDomainForm,
  opsDomainDiffStatusOptions,
  opsDomainEnvironmentOptions,
  opsDomainObservedStatusOptions,
  opsDomainProviderOptions,
  opsDomainProxyOptions,
  opsDomainRoleOptions,
  opsDomainStatusOptions,
  opsDomainTLSOptions,
  type OpsDomainForm,
} from '@/components/admin/ops/opsDomainBindingForm'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import opsApi, {
  type OpsConnector,
  type OpsDomain,
  type OpsDomainDiff,
  type OpsDomainPreview,
  type OpsDomainSyncResult,
  type OpsEnvironment,
  type OpsProject,
} from '@/api/ops'
import { readOpsEnvironmentQuery, withOpsEnvironmentQuery } from '@/lib/opsEnvironment'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const canEdit = computed(() => authStore.hasPermission('ops:domain:edit'))
const canSync = computed(() => authStore.hasPermission('ops:domain:sync'))
const domains = ref<OpsDomain[]>([])
const connectors = ref<OpsConnector[]>([])
const projects = ref<OpsProject[]>([])
const loading = ref(false)
const connectorsLoading = ref(false)
const projectsLoading = ref(false)
const saving = ref(false)
const syncingID = ref(0)
const syncingAll = ref(false)
const dialogOpen = ref(false)
const diffOpen = ref(false)
const diffLoadingID = ref(0)
const diff = ref<OpsDomainDiff | null>(null)
const previewOpen = ref(false)
const previewLoadingID = ref(0)
const previewTab = ref('dns')
const preview = ref<OpsDomainPreview | null>(null)
const environmentFilter = ref<OpsEnvironment | ''>(readOpsEnvironmentQuery(route.query.environment))

const PreviewCodeBlock = defineComponent({
  props: {
    filename: { type: String, required: true },
    content: { type: String, required: true },
  },
  emits: ['copy'],
  setup(props, { emit }) {
 return () => h('div', { class: 'overflow-hidden rounded-lg border border-dashed border-border/70'}, [
 h('div', { class: 'flex items-center justify-between gap-3 border-b border-dashed border-border/70 bg-muted/40 px-3 py-2'}, [
 h('span', { class: 'truncate font-mono text-[10px] font-bold text-muted-foreground'}, props.filename),
        h(Button, {
          size: 'sm',
          variant: 'ghost',
          onClick: () => emit('copy'),
 }, () => [h(Copy, { class: 'size-4'}), '复制']),
      ]),
 h('pre', { class: 'max-h-96 overflow-auto whitespace-pre-wrap break-words p-3 text-xs leading-6'}, props.content),
    ])
  },
})

const form = reactive<OpsDomainForm>(emptyOpsDomainForm())

const cloudflareDomains = computed(() => domains.value.filter((domain) => domain.provider === 'cloudflare'))
const enabledCount = computed(() => domains.value.filter((domain) => domain.enabled).length)
const productionCount = computed(() => domains.value.filter((domain) => domain.environment === 'production').length)
const attentionCount = computed(() => domains.value.filter((domain) => (
  ['pending', 'drifted', 'error'].includes(domain.status)
  || ['unknown', 'drifted', 'error'].includes(domain.observed_status)
)).length)
const summaryCards = computed<readonly OpsSummaryCardItem[]>(() => [
  { key: 'domains', label: '已登记域名', value: domains.value.length, detail: `当前启用 ${enabledCount.value}` },
  { key: 'enabled', label: '当前启用', value: enabledCount.value, detail: `共 ${domains.value.length} 个域名`, tone: 'green' },
  { key: 'production', label: '生产域名', value: productionCount.value, detail: '当前列表中的生产环境', tone: 'primary' },
  { key: 'attention', label: '需处理 / 未同步', value: attentionCount.value, detail: '期望状态或实际状态异常', tone: 'amber' },
])

const assignForm = (domain?: Partial<OpsDomain>): void => {
  assignOpsDomainForm(form, domain)
  if (!domain && environmentFilter.value) {
    form.environment = environmentFilter.value
  }
}

const loadDomains = async (): Promise<void> => {
  loading.value = true
  try {
    const data = await opsApi.listDomains(environmentFilter.value || undefined)
    domains.value = Array.isArray(data?.domains) ? data.domains : []
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '域名列表加载失败')
  } finally {
    loading.value = false
  }
}

const loadConnectors = async (): Promise<void> => {
  if (!canEdit.value) {
    connectors.value = []
    return
  }
  connectorsLoading.value = true
  try {
    const data = await opsApi.listConnectors()
    connectors.value = Array.isArray(data?.connectors) ? data.connectors : []
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'Cloudflare 连接器加载失败')
  } finally {
    connectorsLoading.value = false
  }
}

const loadProjects = async (): Promise<void> => {
  projectsLoading.value = true
  try {
    const data = await opsApi.listProjects()
    projects.value = Array.isArray(data?.projects) ? data.projects : []
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '项目列表加载失败')
  } finally {
    projectsLoading.value = false
  }
}

const changeEnvironment = (): void => {
  void router.replace({ query: withOpsEnvironmentQuery(route.query, environmentFilter.value) })
  void loadDomains()
}

const refreshDomainsPage = async (): Promise<void> => {
  await Promise.all([
    loadDomains(),
    loadProjects(),
    loadConnectors(),
  ])
}

const openCreate = (): void => {
  assignForm()
  dialogOpen.value = true
  if (!projects.value.length) void loadProjects()
  if (!connectors.value.length) void loadConnectors()
}

const openEdit = (domain: OpsDomain): void => {
  assignForm(domain)
  dialogOpen.value = true
  if (!projects.value.length) void loadProjects()
  if (!connectors.value.length) void loadConnectors()
}

const openDiff = async (domain: OpsDomain): Promise<void> => {
  diffLoadingID.value = domain.id
  try {
    diff.value = await opsApi.diffDomain(domain.id)
    diffOpen.value = true
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || `${domain.domain} 差异加载失败`)
  } finally {
    diffLoadingID.value = 0
  }
}

const openPreview = async (domain: OpsDomain): Promise<void> => {
  previewLoadingID.value = domain.id
  previewTab.value = 'dns'
  try {
    preview.value = await opsApi.previewDomain(domain.id)
    previewOpen.value = true
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || `${domain.domain} 预览生成失败`)
  } finally {
    previewLoadingID.value = 0
  }
}

const copyPreview = async (content: string): Promise<void> => {
  try {
    await navigator.clipboard.writeText(content)
    toast.success('配置草稿已复制')
  } catch {
    toast.error('复制失败，请手动复制')
  }
}

const saveDomain = async (): Promise<void> => {
  if (!form.domain.trim()) {
    toast.error('请输入域名')
    return
  }
  if (domainRoleRequiresProject(form.role) && !form.project_binding_id) {
    toast.error('请选择所属项目')
    return
  }
  saving.value = true
  try {
    const payload = {
      domain: form.domain.trim(),
      connector_id: form.provider === 'cloudflare' ? form.connector_id : null,
      project_binding_id: form.project_binding_id,
      role: form.role,
      environment: form.environment,
      provider: form.provider,
      zone: form.zone.trim(),
      target: form.target.trim(),
      proxy_mode: form.proxy_mode,
      tls_mode: form.tls_mode,
      redirect_target: form.redirect_target.trim(),
      status: form.status,
      enabled: form.enabled,
      notes: form.notes.trim(),
    }
    if (form.id) {
      await opsApi.updateDomain(form.id, payload)
      toast.success('域名绑定已保存')
    } else {
      await opsApi.createDomain(payload)
      toast.success('域名绑定已创建')
    }
    dialogOpen.value = false
    await loadDomains()
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '域名保存失败')
  } finally {
    saving.value = false
  }
}

const applyDomainSyncResult = (domain: OpsDomain, result: OpsDomainSyncResult): void => {
  const index = domains.value.findIndex((item) => item.id === domain.id)
  if (index < 0) return
  domains.value[index] = {
    ...domain,
    observed_status: result.observed_status,
    observed_target: result.observed_target,
    observed_proxy_mode: result.observed_proxy_mode,
    observed_tls_mode: result.observed_tls_mode,
    observed_source: result.observed_source,
    last_observed_at: result.last_observed_at,
    observed_error: result.observed_error || '',
  }
}

const syncMessage = (result: OpsDomainSyncResult): string => {
  if (result.observed_status === 'matched') return `${result.domain} 已匹配 Cloudflare 实际状态`
  if (result.observed_status === 'drifted') return `${result.domain} 检测到配置差异`
  return result.message || `${result.domain} 同步失败`
}

const syncDomain = async (domain: OpsDomain): Promise<void> => {
  syncingID.value = domain.id
  try {
    const result = await opsApi.syncDomain(domain.id)
    const tone = result.observed_status === 'matched'
      ? 'success'
      : result.observed_status === 'drifted' ? 'warning' : 'error'
    toast[tone](syncMessage(result))
    applyDomainSyncResult(domain, result)
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || `${domain.domain} 同步失败`)
  } finally {
    syncingID.value = 0
  }
}

const syncAllCloudflare = async (): Promise<void> => {
  syncingAll.value = true
  let matched = 0
  let drifted = 0
  let failed = 0
  try {
    for (const domain of cloudflareDomains.value) {
      try {
        const result = await opsApi.syncDomain(domain.id)
        applyDomainSyncResult(domain, result)
        if (result.observed_status === 'matched') matched += 1
        else if (result.observed_status === 'drifted') drifted += 1
        else failed += 1
      } catch {
        failed += 1
      }
    }
    await loadDomains()
    if (failed > 0) {
      toast.error(`Cloudflare 同步完成：${matched} 个匹配，${drifted} 个漂移，${failed} 个失败`)
    } else if (drifted > 0) {
      toast.warning(`Cloudflare 同步完成：${matched} 个匹配，${drifted} 个漂移`)
    } else {
      toast.success(`Cloudflare 同步完成：${matched} 个域名已匹配`)
    }
  } finally {
    syncingAll.value = false
  }
}

const toggleDomain = async (domain: OpsDomain): Promise<void> => {
  try {
    const updated = await opsApi.setDomainEnabled(domain.id, !domain.enabled)
    const index = domains.value.findIndex((item) => item.id === domain.id)
    if (index >= 0) domains.value[index] = updated
    toast.success(updated.enabled ? '域名已启用' : '域名已停用')
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '域名状态更新失败')
  }
}

const optionLabel = (options: Array<{ value: string; label: string }>, value: string): string => (
  options.find((option) => option.value === value)?.label || value || '-'
)
const roleLabel = (value: string): string => optionLabel(opsDomainRoleOptions, value)
const environmentLabel = (value: string): string => optionLabel(opsDomainEnvironmentOptions, value)
const projectLabel = (projectID?: number | null): string => (
  projects.value.find((project) => project.id === projectID)?.name || '未绑定'
)
const providerLabel = (value: string): string => optionLabel(opsDomainProviderOptions, value)
const proxyLabel = (value: string): string => optionLabel(opsDomainProxyOptions, value)
const tlsLabel = (value: string): string => optionLabel(opsDomainTLSOptions, value)
const statusLabel = (value: string): string => optionLabel(opsDomainStatusOptions, value)
const observedStatusLabel = (value: string): string => optionLabel(opsDomainObservedStatusOptions, value)
const diffStatusLabel = (value: string): string => optionLabel(opsDomainDiffStatusOptions, value)
const formatDate = (value?: string): string => value ? new Date(value).toLocaleString('zh-CN') : '未同步'

const roleTone = (value: string): AdminStatusTone => value === 'canonical' ? 'blue' : value === 'admin' ? 'coral' : 'gray'
const statusTone = (value: string): AdminStatusTone => {
  if (value === 'active') return 'green'
  if (value === 'pending' || value === 'drifted') return 'amber'
  if (value === 'error') return 'coral'
  return 'gray'
}
const observedStatusTone = (value: string): AdminStatusTone => {
  if (value === 'matched') return 'green'
  if (value === 'drifted' || value === 'unknown') return 'amber'
  if (value === 'error') return 'coral'
  return 'gray'
}
const diffTone = (value: string): AdminStatusTone => {
  if (value === 'matched') return 'green'
  if (value === 'drifted' || value === 'unknown') return 'amber'
  if (value === 'error') return 'coral'
  return 'gray'
}

watch(() => route.query.environment, (value) => {
  const nextEnvironment = readOpsEnvironmentQuery(value)
  if (nextEnvironment === environmentFilter.value) return
  environmentFilter.value = nextEnvironment
  void loadDomains()
})

onMounted(refreshDomainsPage)
</script>
