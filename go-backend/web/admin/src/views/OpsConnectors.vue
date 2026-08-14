<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="运维中心 / 连接器中心"
      description="登记 Cloudflare、Hostinger 和代码发布平台连接；这里只做连接测试，不直接执行发布或 DNS 写入。"
    >
      <template #actions>
        <Button variant="outline" :disabled="loading" @click="loadConnectors">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
          刷新
        </Button>
        <Button v-if="canEdit" @click="openCreate">
          <Plus class="size-4" />
          新增连接
        </Button>
      </template>
    </AdminPageHeader>

    <section class="grid gap-3 md:grid-cols-4">
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">已登记连接</p>
        <p class="mt-2 text-2xl font-black">{{ connectors.length }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">当前启用</p>
        <p class="mt-2 text-2xl font-black text-emerald-600">{{ enabledCount }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">凭据已配置</p>
        <p class="mt-2 text-2xl font-black text-primary">{{ credentialCount }}</p>
      </div>
      <div class="rounded-2xl border border-dashed border-border/80 bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">需处理状态</p>
        <p class="mt-2 text-2xl font-black text-amber-600">{{ attentionCount }}</p>
      </div>
    </section>

    <AdminTablePanel :loading="loading" :batch-visible="false">
      <Table class="min-w-[1120px]">
        <TableHeader>
          <TableRow>
            <TableHead>连接名称</TableHead>
            <TableHead>提供商 / 环境</TableHead>
            <TableHead>认证方式</TableHead>
            <TableHead>凭据状态</TableHead>
            <TableHead>范围</TableHead>
            <TableHead>测试状态</TableHead>
            <TableHead>状态</TableHead>
            <TableHead class="w-48 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="connectors.length === 0" :colspan="8">
            <div class="py-8 text-center text-xs text-muted-foreground">还没有登记连接器</div>
          </TableEmpty>
          <TableRow v-for="connector in connectors" :key="connector.id">
            <TableCell>
              <div class="min-w-0">
                <p class="truncate font-bold">{{ connector.name }}</p>
                <p class="mt-1 truncate text-[10px] text-muted-foreground" :title="connector.endpoint">
                  {{ connector.endpoint || defaultEndpointLabel(connector.provider) }}
                </p>
              </div>
            </TableCell>
            <TableCell>
              <p class="text-xs font-bold">{{ providerLabel(connector.provider) }}</p>
              <p class="mt-1 text-[10px] text-muted-foreground">{{ environmentLabel(connector.environment) }}</p>
            </TableCell>
            <TableCell class="text-xs">{{ authTypeLabel(connector.auth_type) }}</TableCell>
            <TableCell>
              <AdminStatusBadge :tone="connector.credential_configured ? 'green' : 'amber'">
                {{ connector.credential_configured ? '已配置' : '未配置' }}
              </AdminStatusBadge>
              <p v-if="connector.credential_fields?.length" class="mt-1 text-[10px] text-muted-foreground">
                {{ connector.credential_fields.join(', ') }}
              </p>
              <p v-else-if="connector.credential_ref" class="mt-1 truncate text-[10px] text-muted-foreground">
                {{ connector.credential_ref }}
              </p>
            </TableCell>
            <TableCell class="max-w-44 truncate text-xs" :title="connector.scopes">
              {{ connector.scopes || '-' }}
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="testTone(connector.last_test_status)">
                {{ testLabel(connector.last_test_status) }}
              </AdminStatusBadge>
              <p v-if="connector.last_tested_at" class="mt-1 text-[10px] text-muted-foreground">
                {{ formatDate(connector.last_tested_at) }}
              </p>
              <p v-if="connector.last_error" class="mt-1 max-w-36 truncate text-[10px] text-rose-600" :title="connector.last_error">
                {{ connector.last_error }}
              </p>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="statusTone(connector.status)">
                {{ statusLabel(connector.status) }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right">
              <div class="flex justify-end gap-1">
                <Button
                  v-if="canEdit"
                  size="icon"
                  variant="ghost"
                  :title="`测试连接 ${connector.name}`"
                  :disabled="testingID === connector.id"
                  @click="testConnector(connector)"
                >
                  <LoaderCircle v-if="testingID === connector.id" class="size-4 animate-spin" />
                  <PlugZap v-else class="size-4" />
                </Button>
                <Button
                  v-if="canEdit"
                  size="icon"
                  variant="ghost"
                  :title="connector.enabled ? '停用连接' : '启用连接'"
                  @click="toggleConnector(connector)"
                >
                  <Power class="size-4" />
                </Button>
                <Button v-if="canEdit" size="icon" variant="ghost" title="编辑连接" @click="openEdit(connector)">
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
          <DialogTitle>{{ form.id ? '编辑连接器' : '新增连接器' }}</DialogTitle>
          <DialogDescription>
            凭据只写入后端加密存储。编辑时凭据字段留空会保留原值，页面不会回显旧 Token。
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-4" @submit.prevent="saveConnector">
          <div class="grid gap-4 md:grid-cols-2">
            <AdminFormField label="连接名称" required>
              <Input v-model="form.name" placeholder="Cloudflare Production" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="提供商" required>
              <select v-model="form.provider" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in providerOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="环境">
              <select v-model="form.environment" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in environmentOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="认证方式">
              <select v-model="form.auth_type" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in authTypeOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="测试接口" description="Cloudflare 留空会使用 Token 验证接口；Hostinger 建议填写官方只读接口。">
              <Input v-model="form.endpoint" type="url" placeholder="https://api.example.com/health" :disabled="saving" />
            </AdminFormField>
            <AdminFormField v-if="form.auth_type === 'environment'" label="环境变量名" required>
              <Input v-model="form.credential_ref" placeholder="HOSTINGER_API_TOKEN" :disabled="saving" />
            </AdminFormField>
            <AdminFormField v-else label="凭据引用">
              <Input v-model="form.credential_ref" placeholder="可选：外部密钥系统引用" :disabled="saving" />
            </AdminFormField>
            <AdminFormField label="权限范围" description="用逗号或空格记录允许的范围，不会自动扩大外部权限。">
              <Input v-model="form.scopes" placeholder="zones:read, dns:read" :disabled="saving" />
            </AdminFormField>
            <template v-if="form.auth_type !== 'none' && form.auth_type !== 'manual' && form.auth_type !== 'environment'">
              <AdminFormField
                v-for="field in credentialFields"
                :key="field.key"
                :label="field.label"
                :description="form.id ? '留空保留已保存凭据；填写后将覆盖这一字段。' : '提交后只显示字段名，不显示原值。'"
              >
                <Input
                  v-model="form.credentials[field.key]"
                  :type="field.secret ? 'password' : 'text'"
                  autocomplete="new-password"
                  :placeholder="field.placeholder"
                  :disabled="saving"
                />
              </AdminFormField>
            </template>
            <AdminFormField label="状态">
              <select v-model="form.status" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
                <option v-for="option in statusOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="备注" class="md:col-span-2">
              <Textarea v-model="form.notes" class="min-h-24" placeholder="记录连接用途、权限边界和轮换说明。" :disabled="saving" />
            </AdminFormField>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" :disabled="saving" @click="dialogOpen = false">取消</Button>
            <Button type="submit" :disabled="saving || !canEdit">
              <LoaderCircle v-if="saving" class="size-4 animate-spin" />
              <Save v-else class="size-4" />
              {{ saving ? '保存中' : '保存连接' }}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { LoaderCircle, Pencil, PlugZap, Plus, Power, RefreshCw, Save } from '@lucide/vue'
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
import { Textarea } from '@/components/ui/textarea'
import opsApi, { type OpsConnectorPayload } from '@/api/ops'
import { useAuthStore } from '@/stores/auth'

interface OpsConnector {
  id: number
  name: string
  provider: string
  environment: string
  endpoint: string
  auth_type: string
  credential_ref: string
  credential_configured: boolean
  credential_fields: string[]
  scopes: string
  status: string
  enabled: boolean
  last_test_status: string
  last_tested_at?: string
  last_error: string
  notes: string
}

interface CredentialField {
  key: string
  label: string
  placeholder: string
  secret: boolean
}

interface OpsConnectorForm {
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

const authStore = useAuthStore()
const canEdit = computed(() => authStore.hasPermission('ops:connector:edit'))
const connectors = ref<OpsConnector[]>([])
const loading = ref(false)
const saving = ref(false)
const testingID = ref(0)
const dialogOpen = ref(false)

const providerOptions = [
  { value: 'cloudflare', label: 'Cloudflare' },
  { value: 'hostinger', label: 'Hostinger' },
  { value: 'github', label: 'GitHub' },
  { value: 'ghcr', label: 'GitHub / GHCR' },
  { value: 'other', label: '其他' },
]
const environmentOptions = [
  { value: 'production', label: '生产' },
  { value: 'staging', label: '预发布' },
  { value: 'test', label: '测试' },
  { value: 'local', label: '本地' },
]
const authTypeOptions = [
  { value: 'api_token', label: 'API Token' },
  { value: 'api_key', label: 'API Key' },
  { value: 'bearer', label: 'Bearer Token' },
  { value: 'basic', label: 'Basic Auth' },
  { value: 'environment', label: '后端环境变量' },
  { value: 'manual', label: '手动登记' },
  { value: 'none', label: '无需认证' },
]
const statusOptions = [
  { value: 'active', label: '正常' },
  { value: 'pending', label: '待测试' },
  { value: 'disabled', label: '已停用' },
  { value: 'error', label: '测试失败' },
]

const emptyForm = (): OpsConnectorForm => ({
  id: 0,
  name: '',
  provider: 'cloudflare',
  environment: 'production',
  endpoint: '',
  auth_type: 'api_token',
  credential_ref: '',
  credentials: { token: '', api_key: '', username: '', password: '' },
  scopes: '',
  status: 'pending',
  enabled: true,
  notes: '',
})
const form = reactive<OpsConnectorForm>(emptyForm())

const enabledCount = computed(() => connectors.value.filter((connector) => connector.enabled).length)
const credentialCount = computed(() => connectors.value.filter((connector) => connector.credential_configured).length)
const attentionCount = computed(() => connectors.value.filter((connector) => ['pending', 'error'].includes(connector.status)).length)

const credentialFields = computed<CredentialField[]>(() => {
  if (form.auth_type === 'basic') {
    return [
      { key: 'username', label: '用户名', placeholder: '只读账号用户名', secret: false },
      { key: 'password', label: '密码', placeholder: '只读账号密码', secret: true },
    ]
  }
  if (form.auth_type === 'api_key') {
    return [{ key: 'api_key', label: 'API Key', placeholder: '粘贴 API Key', secret: true }]
  }
  return [{ key: 'token', label: 'Token', placeholder: '粘贴 Token，不会回显', secret: true }]
})

const assignForm = (connector?: Partial<OpsConnector>): void => {
  Object.assign(form, emptyForm(), connector || {})
  form.credentials = { token: '', api_key: '', username: '', password: '' }
}

watch(() => form.auth_type, (authType) => {
  if (authType === 'environment') form.credentials = { token: '', api_key: '', username: '', password: '' }
})

const loadConnectors = async (): Promise<void> => {
  loading.value = true
  try {
    const data = await opsApi.listConnectors()
    connectors.value = Array.isArray(data?.connectors) ? data.connectors : []
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '连接器列表加载失败')
  } finally {
    loading.value = false
  }
}

const openCreate = (): void => {
  assignForm()
  dialogOpen.value = true
}

const openEdit = (connector: OpsConnector): void => {
  assignForm(connector)
  dialogOpen.value = true
}

const saveConnector = async (): Promise<void> => {
  if (!form.name.trim()) {
    toast.error('请输入连接名称')
    return
  }
  if (form.auth_type === 'environment' && !form.credential_ref.trim()) {
    toast.error('请输入后端环境变量名')
    return
  }
  saving.value = true
  try {
    const credentials = Object.fromEntries(
      Object.entries(form.credentials).filter(([, value]) => String(value || '').trim())
    )
    const payload: OpsConnectorPayload = {
      name: form.name.trim(),
      provider: form.provider,
      environment: form.environment,
      endpoint: form.endpoint.trim(),
      auth_type: form.auth_type,
      credential_ref: form.credential_ref.trim(),
      credentials,
      scopes: form.scopes.trim(),
      status: form.status,
      enabled: form.enabled,
      notes: form.notes.trim(),
    }
    if (form.id) {
      await opsApi.updateConnector(form.id, payload)
      toast.success('连接器已保存')
    } else {
      await opsApi.createConnector(payload)
      toast.success('连接器已创建')
    }
    dialogOpen.value = false
    await loadConnectors()
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '连接器保存失败')
  } finally {
    saving.value = false
  }
}

const toggleConnector = async (connector: OpsConnector): Promise<void> => {
  try {
    const updated = await opsApi.setConnectorEnabled(connector.id, !connector.enabled)
    const index = connectors.value.findIndex((item) => item.id === connector.id)
    if (index >= 0) connectors.value[index] = updated
    toast.success(updated.enabled ? '连接器已启用' : '连接器已停用')
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '连接器状态更新失败')
  }
}

const testConnector = async (connector: OpsConnector): Promise<void> => {
  testingID.value = connector.id
  try {
    const result = await opsApi.testConnector(connector.id)
    toast[result.success ? 'success' : 'error'](result.message || (result.success ? '连接测试成功' : '连接测试失败'))
    await loadConnectors()
  } catch (error: any) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '连接测试失败')
  } finally {
    testingID.value = 0
  }
}

const optionLabel = (options: Array<{ value: string; label: string }>, value: string): string => (
  options.find((option) => option.value === value)?.label || value || '-'
)
const providerLabel = (value: string): string => optionLabel(providerOptions, value)
const environmentLabel = (value: string): string => optionLabel(environmentOptions, value)
const authTypeLabel = (value: string): string => optionLabel(authTypeOptions, value)
const statusLabel = (value: string): string => optionLabel(statusOptions, value)
const testLabel = (value: string): string => value === 'success' ? '成功' : value === 'failed' ? '失败' : '未测试'
const defaultEndpointLabel = (provider: string): string => provider === 'cloudflare' ? 'Cloudflare Token 验证接口' : '未设置测试接口'
const formatDate = (value: string): string => new Date(value).toLocaleString()
const statusTone = (value: string): AdminStatusTone => {
  if (value === 'active') return 'green'
  if (value === 'error') return 'coral'
  if (value === 'pending') return 'amber'
  return 'gray'
}
const testTone = (value: string): AdminStatusTone => value === 'success' ? 'green' : value === 'failed' ? 'coral' : 'gray'

onMounted(loadConnectors)
</script>
