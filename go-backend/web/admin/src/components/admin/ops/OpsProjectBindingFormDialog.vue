<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>{{ form.id ? '编辑项目绑定' : '新增项目绑定' }}</DialogTitle>
        <DialogDescription>
          项目状态来自后台维护的声明式记录。实际健康状态只能通过同步更新，不在表单里手工覆盖。
        </DialogDescription>
      </DialogHeader>

      <form class="mt-4 space-y-4" @submit.prevent="emit('save')">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="项目名称" required>
            <Input v-model="form.name" placeholder="commerce-platform" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="环境" required>
            <select v-model="form.environment" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option v-for="option in opsProjectEnvironmentOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField label="绑定 VPS" required>
            <select v-model.number="form.vps_binding_id" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option :value="0">请选择 VPS</option>
              <option v-for="vps in compatibleVPS" :key="vps.id" :value="vps.id">
                {{ projectVPSLabel(vps) }}{{ vps.enabled ? '' : ' · 已停用' }}
              </option>
            </select>
          </AdminFormField>
          <AdminFormField label="Hostinger 连接器">
            <select v-model="form.connector_id" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option :value="null">沿用 VPS 连接器</option>
              <option v-for="connector in compatibleConnectors" :key="connector.id" :value="connector.id">
                {{ connector.name }} · {{ connector.environment }}
              </option>
            </select>
          </AdminFormField>
          <AdminFormField label="Hostinger 项目 ID">
            <Input v-model="form.provider_resource_id" placeholder="可在 Hostinger 确认后补录" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="Compose 来源">
            <Input v-model="form.compose_source" placeholder="compose.prod.yml" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="Compose 项目名">
            <Input v-model="form.compose_project_name" placeholder="commerce-platform" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="共享网关网络">
            <Input v-model="form.gateway_network" placeholder="shared-edge" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="网关别名">
            <Input v-model="form.gateway_alias" placeholder="theme-web" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="服务列表" description="用逗号分隔，例如 db, redis, api。">
            <Input v-model="form.services" placeholder="db, redis, api, storefront, admin, web" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="网络边界" description="用逗号分隔。">
            <Input v-model="form.networks" placeholder="db, cache, app, shared-edge" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="卷边界" description="用逗号分隔。">
            <Input v-model="form.volumes" placeholder="postgres, redis, uploads" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="当前镜像标签">
            <Input v-model="form.current_image_tag" placeholder="master 或 sha-..." :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="当前 Commit SHA">
            <Input v-model="form.current_commit_sha" placeholder="可选：完整 SHA" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="期望状态">
            <select v-model="form.status" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option v-for="option in opsProjectStatusOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField label="最后部署时间">
            <Input v-model="form.last_deployment_at" type="datetime-local" :disabled="saving" />
          </AdminFormField>
          <div class="md:col-span-2 rounded-lg border border-border/80 p-3">
            <div class="flex flex-wrap items-center justify-between gap-3">
              <p class="text-xs font-black text-foreground">Quick Buy 限流</p>
              <label class="inline-flex items-center gap-2 text-xs font-bold">
                <input v-model="form.quickBuyRateLimit.enabled" type="checkbox" class="size-4" :disabled="saving" />
                启用
              </label>
            </div>
            <div class="mt-3 grid gap-3 md:grid-cols-4">
              <AdminFormField label="IP 每分钟">
                <Input v-model.number="form.quickBuyRateLimit.ipRequestsPerMinute" type="number" min="1" step="1" :disabled="saving || !form.quickBuyRateLimit.enabled" />
              </AdminFormField>
              <AdminFormField label="IP Burst">
                <Input v-model.number="form.quickBuyRateLimit.ipBurst" type="number" min="1" step="1" :disabled="saving || !form.quickBuyRateLimit.enabled" />
              </AdminFormField>
              <AdminFormField label="Session 每分钟">
                <Input v-model.number="form.quickBuyRateLimit.sessionRequestsPerMinute" type="number" min="1" step="1" :disabled="saving || !form.quickBuyRateLimit.enabled" />
              </AdminFormField>
              <AdminFormField label="Session Burst">
                <Input v-model.number="form.quickBuyRateLimit.sessionBurst" type="number" min="1" step="1" :disabled="saving || !form.quickBuyRateLimit.enabled" />
              </AdminFormField>
              <AdminFormField label="Edge IP 每分钟">
                <Input v-model.number="form.quickBuyRateLimit.edgeIPRequestsPerMinute" type="number" min="1" step="1" :disabled="saving || !form.quickBuyRateLimit.enabled || !form.quickBuyRateLimit.caddyRateLimitEnabled" />
              </AdminFormField>
              <AdminFormField label="Edge IP Burst">
                <Input v-model.number="form.quickBuyRateLimit.edgeIPBurst" type="number" min="1" step="1" :disabled="saving || !form.quickBuyRateLimit.enabled || !form.quickBuyRateLimit.caddyRateLimitEnabled" />
              </AdminFormField>
              <div class="flex items-center md:col-span-2">
                <label class="inline-flex items-center gap-2 text-xs font-bold">
                  <input v-model="form.quickBuyRateLimit.caddyRateLimitEnabled" type="checkbox" class="size-4" :disabled="saving || !form.quickBuyRateLimit.enabled" />
                  Caddy 边缘桶
                </label>
              </div>
            </div>
          </div>
          <AdminFormField label="备份策略" class="md:col-span-2">
            <Textarea v-model="form.backup_policy" class="min-h-20" placeholder="每日 PostgreSQL、uploads 备份，异地保存和恢复演练。" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="恢复说明" class="md:col-span-2">
            <Textarea v-model="form.restore_notes" class="min-h-20" placeholder="记录最近一次恢复演练、备份位置和待处理事项。" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="备注" class="md:col-span-2">
            <Textarea v-model="form.notes" class="min-h-24" placeholder="记录项目边界、发布来源和维护说明。" :disabled="saving" />
          </AdminFormField>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" :disabled="saving" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="saving || !canEdit">
            <LoaderCircle v-if="saving" class="size-4 animate-spin" />
            <Save v-else class="size-4" />
            {{ saving ? '保存中' : '保存项目' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { LoaderCircle, Save } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
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
import { Textarea } from '@/components/ui/textarea'
import type { OpsConnector, OpsVPS } from '@/api/ops'
import {
  opsProjectEnvironmentOptions,
  opsProjectStatusOptions,
  projectVPSLabel,
  type OpsProjectForm,
} from './opsProjectBindingForm'

const props = withDefaults(defineProps<{
  open?: boolean
  form: OpsProjectForm
  vps?: OpsVPS[]
  connectors?: OpsConnector[]
  saving?: boolean
  canEdit?: boolean
}>(), {
  open: false,
  vps: () => [],
  connectors: () => [],
  saving: false,
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'save'): void
}>()

const compatibleVPS = computed(() => props.vps.filter((vps) => vps.environment === props.form.environment))
const compatibleConnectors = computed(() => props.connectors.filter((connector) => (
  connector.provider === 'hostinger' &&
  connector.environment === props.form.environment
)))

watch(compatibleVPS, (vps) => {
  if (!props.vps.length) return
  if (props.form.vps_binding_id && !vps.some((item) => item.id === props.form.vps_binding_id)) {
    props.form.vps_binding_id = 0
  }
}, { immediate: true })

watch(compatibleConnectors, (connectors) => {
  if (!props.connectors.length) return
  if (props.form.connector_id && !connectors.some((connector) => connector.id === props.form.connector_id)) {
    props.form.connector_id = null
  }
}, { immediate: true })
</script>
