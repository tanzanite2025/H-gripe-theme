<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="运维中心 / 域名中心"
      description="统一维护主域名、别名、后台域、跳转域和验证域的绑定关系。"
    >
      <template #actions>
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
        <Button variant="outline" :disabled="loading" @click="loadDomains">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
          刷新
        </Button>
        <Button v-if="canEdit" @click="openCreate">
          <Plus class="size-4" />
          新增域名
        </Button>
      </template>
    </AdminPageHeader>

    <section class="grid gap-3 md:grid-cols-4">
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">已登记域名</p>
        <p class="mt-2 text-2xl font-black">{{ domains.length }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">当前启用</p>
        <p class="mt-2 text-2xl font-black text-emerald-600">{{ enabledCount }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">生产域名</p>
        <p class="mt-2 text-2xl font-black text-primary">{{ productionCount }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">需处理 / 未同步</p>
        <p class="mt-2 text-2xl font-black text-amber-600">{{ attentionCount }}</p>
      </div>
    </section>

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
              <p class="mt-1 text-[10px] text-muted-foreground">{{ domain.zone || '未登记 Zone' }}</p>
            </TableCell>
            <TableCell class="max-w-56">
              <p class="truncate font-mono text-[10px]" :title="domain.target">期望：{{ domain.target || '-' }}</p>
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

    <Dialog v-model:open="dialogOpen">
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>{{ form.id ? '编辑域名绑定' : '新增域名绑定' }}</DialogTitle>
          <DialogDescription>
            这里只维护后台的期望状态，不会直接修改 Cloudflare、Hostinger 或生产网关。
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-4" @submit.prevent="saveDomain">
          <div class="grid gap-4 md:grid-cols-2">
            <AdminFormField label="域名" required description="只填写 hostname，例如 admin.learn.gripe。">
              <Input v-model="form.domain" placeholder="learn.gripe" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="角色" required>
              <select v-model="form.role" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in roleOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="环境" required>
              <select v-model="form.environment" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in environmentOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField
              label="所属项目"
              :required="deploymentRoleRequiresProject"
              description="部署 preflight 只采信显式绑定到项目的公网域名。"
            >
              <select
                v-model="form.project_binding_id"
                class="h-10 w-full rounded-md border bg-background px-3 text-sm"
                :disabled="saving || projectsLoading"
              >
                <option :value="null">{{ deploymentRoleRequiresProject ? '请选择项目' : '不绑定项目' }}</option>
                <option v-for="project in matchingProjects" :key="project.id" :value="project.id">
                  {{ project.name }}{{ project.enabled ? '' : '（已停用）' }}
                </option>
              </select>
              <p v-if="projectsLoading" class="mt-1 text-[10px] text-muted-foreground">正在加载项目...</p>
              <p v-else-if="deploymentRoleRequiresProject && matchingProjects.length === 0" class="mt-1 text-[10px] text-amber-600">
                当前环境没有可绑定项目，请先在项目中心登记。
              </p>
            </AdminFormField>
            <AdminFormField label="提供商" required>
              <select v-model="form.provider" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in providerOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField v-if="form.provider === 'cloudflare'" label="Cloudflare 只读连接器">
              <select v-model="form.connector_id" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving || connectorsLoading">
                <option :value="null">暂不绑定</option>
                <option v-for="connector in cloudflareConnectors" :key="connector.id" :value="connector.id">
                  {{ connector.name }}{{ connector.enabled ? '' : '（已停用）' }}
                </option>
              </select>
              <p v-if="connectorsLoading" class="mt-1 text-[10px] text-muted-foreground">正在加载连接器...</p>
              <p v-else-if="cloudflareConnectors.length === 0" class="mt-1 text-[10px] text-amber-600">
                请先在连接器中心登记 Cloudflare 只读 Token。
              </p>
            </AdminFormField>
            <AdminFormField label="Cloudflare Zone / DNS Zone">
              <Input v-model="form.zone" placeholder="learn.gripe" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="目标">
              <Input v-model="form.target" placeholder="theme-web:8080 或 VPS IP" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="代理模式">
              <select v-model="form.proxy_mode" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in proxyOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="TLS 模式">
              <select v-model="form.tls_mode" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in tlsOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField v-if="form.role === 'redirect'" label="跳转目标" required>
              <Input v-model="form.redirect_target" placeholder="https://learn.gripe" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="状态">
              <select v-model="form.status" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in statusOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="备注" class="md:col-span-2">
              <Textarea v-model="form.notes" class="min-h-24" placeholder="记录绑定来源、切换窗口或维护说明。" :disabled="saving" />
            </AdminFormField>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="dialogOpen = false">取消</Button>
            <Button type="submit" :disabled="saving || !canEdit">
              <LoaderCircle v-if="saving" class="size-4 animate-spin" />
              <Save v-else class="size-4" />
              {{ saving ? '保存中' : '保存域名' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>

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
                <div><span class="text-muted-foreground">Zone：</span>{{ preview.dns.zone || '-' }}</div>
                <div><span class="text-muted-foreground">Type：</span>{{ preview.dns.record_type }}</div>
                <div><span class="text-muted-foreground">Name：</span>{{ preview.dns.name }}</div>
                <div class="sm:col-span-2"><span class="text-muted-foreground">Content：</span>{{ preview.dns.content || '-' }}</div>
                <div><span class="text-muted-foreground">Proxy：</span>{{ proxyLabel(preview.dns.proxy_mode) }}</div>
                <div><span class="text-muted-foreground">TLS：</span>{{ tlsLabel(preview.dns.tls_mode) }}</div>
                <div v-if="preview.dns.redirect" class="sm:col-span-2">
                  <span class="text-muted-foreground">Redirect：</span>{{ preview.dns.redirect_target || '-' }}
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
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { Copy, FileCode2, GitCompareArrows, LoaderCircle, Pencil, Plus, Power, RefreshCw, Save } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Textarea } from '@/components/ui/textarea'
import opsApi, { type OpsDomainDiff, type OpsDomainPreview, type OpsDomainSyncResult } from '@/api/ops'
import { useAuthStore } from '@/stores/auth'

interface OpsDomain {
  id: number
  domain: string
  connector_id?: number | null
  project_binding_id?: number | null
  role: string
  environment: string
  provider: string
  zone: string
  target: string
  proxy_mode: string
  tls_mode: string
  redirect_target: string
  status: string
  observed_status: string
  observed_target: string
  observed_proxy_mode: string
  observed_tls_mode: string
  observed_source: string
  last_observed_at?: string
  observed_error?: string
  enabled: boolean
  notes: string
}

interface OpsDomainForm {
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

interface OpsConnectorOption {
  id: number
  name: string
  provider: string
  enabled: boolean
}

interface OpsProjectOption {
  id: number
  name: string
  environment: string
  enabled: boolean
}

const authStore = useAuthStore()
const canEdit = computed(() => authStore.hasPermission('ops:domain:edit'))
const canSync = computed(() => authStore.hasPermission('ops:domain:sync'))
const domains = ref<OpsDomain[]>([])
const connectors = ref<OpsConnectorOption[]>([])
const projects = ref<OpsProjectOption[]>([])
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

const PreviewCodeBlock = defineComponent({
  props: {
    filename: { type: String, required: true },
    content: { type: String, required: true },
  },
  emits: ['copy'],
  setup(props, { emit }) {
    return () => h('div', { class: 'overflow-hidden rounded-lg border border-dashed border-border/70' }, [
      h('div', { class: 'flex items-center justify-between gap-3 border-b border-dashed border-border/70 bg-muted/40 px-3 py-2' }, [
        h('span', { class: 'truncate font-mono text-[10px] font-bold text-muted-foreground' }, props.filename),
        h(Button, {
          size: 'sm',
          variant: 'ghost',
          onClick: () => emit('copy'),
        }, () => [h(Copy, { class: 'size-4' }), '复制']),
      ]),
      h('pre', { class: 'max-h-96 overflow-auto whitespace-pre-wrap break-words p-3 text-xs leading-6' }, props.content),
    ])
  },
})

const roleOptions = [
  { value: 'canonical', label: '主域名' },
  { value: 'alias', label: '别名' },
  { value: 'admin', label: '后台域' },
  { value: 'redirect', label: '跳转域' },
  { value: 'verification', label: '验证域' },
  { value: 'internal', label: '内部域' },
]
const environmentOptions = [
  { value: 'production', label: '生产' },
  { value: 'staging', label: '预发布' },
  { value: 'test', label: '测试' },
  { value: 'local', label: '本地' },
]
const providerOptions = [
  { value: 'cloudflare', label: 'Cloudflare' },
  { value: 'hostinger', label: 'Hostinger' },
  { value: 'other', label: '其他' },
]
const proxyOptions = [
  { value: 'proxied', label: '已代理' },
  { value: 'dns_only', label: 'DNS only' },
  { value: 'unknown', label: '未知' },
]
const tlsOptions = [
  { value: 'full_strict', label: 'Full (strict)' },
  { value: 'full', label: 'Full' },
  { value: 'flexible', label: 'Flexible' },
  { value: 'off', label: '关闭' },
  { value: 'unknown', label: '未知' },
]
const statusOptions = [
  { value: 'active', label: '正常' },
  { value: 'pending', label: '待确认' },
  { value: 'disabled', label: '已停用' },
  { value: 'drifted', label: '配置漂移' },
  { value: 'error', label: '错误' },
]

const emptyForm = (): OpsDomainForm => ({
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

const form = reactive<OpsDomainForm>(emptyForm())

const cloudflareConnectors = computed(() => connectors.value.filter((connector) => connector.provider === 'cloudflare'))
const cloudflareDomains = computed(() => domains.value.filter((domain) => domain.provider === 'cloudflare'))
const matchingProjects = computed(() => projects.value.filter((project) => project.environment === form.environment))
const deploymentRoleRequiresProject = computed(() => ['canonical', 'alias', 'admin', 'redirect'].includes(form.role))
const enabledCount = computed(() => domains.value.filter((domain) => domain.enabled).length)
const productionCount = computed(() => domains.value.filter((domain) => domain.environment === 'production').length)
const attentionCount = computed(() => domains.value.filter((domain) => (
  ['pending', 'drifted', 'error'].includes(domain.status)
  || ['unknown', 'drifted', 'error'].includes(domain.observed_status)
)).length)

const assignForm = (domain?: Partial<OpsDomain>): void => {
  Object.assign(form, emptyForm(), domain || {})
}

const loadDomains = async (): Promise<void> => {
  loading.value = true
  try {
    const data = await opsApi.listDomains()
    domains.value = Array.isArray(data?.domains) ? data.domains : []
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '域名列表加载失败')
  } finally {
    loading.value = false
  }
}

const loadConnectors = async (): Promise<void> => {
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

const openCreate = (): void => {
  assignForm()
  dialogOpen.value = true
}

const openEdit = (domain: OpsDomain): void => {
  assignForm(domain)
  dialogOpen.value = true
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
  if (deploymentRoleRequiresProject.value && !form.project_binding_id) {
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
    await loadDomains()
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
const roleLabel = (value: string): string => optionLabel(roleOptions, value)
const environmentLabel = (value: string): string => optionLabel(environmentOptions, value)
const projectLabel = (projectID?: number | null): string => (
  projects.value.find((project) => project.id === projectID)?.name || '未绑定'
)
const providerLabel = (value: string): string => optionLabel(providerOptions, value)
const proxyLabel = (value: string): string => optionLabel(proxyOptions, value)
const tlsLabel = (value: string): string => optionLabel(tlsOptions, value)
const statusLabel = (value: string): string => optionLabel(statusOptions, value)
const observedStatusLabel = (value: string): string => optionLabel([
  { value: 'unknown', label: '未同步' },
  { value: 'matched', label: '已匹配' },
  { value: 'drifted', label: '漂移' },
  { value: 'error', label: '检查错误' },
], value)
const diffStatusLabel = (value: string): string => optionLabel([
  { value: 'unknown', label: '未确认' },
  { value: 'matched', label: '已匹配' },
  { value: 'drifted', label: '有差异' },
  { value: 'error', label: '检查错误' },
], value)
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

onMounted(() => {
  void Promise.all([loadDomains(), loadConnectors(), loadProjects()])
})
</script>
