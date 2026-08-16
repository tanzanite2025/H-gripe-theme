<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>{{ form.id ? '编辑域名绑定' : '新增域名绑定' }}</DialogTitle>
        <DialogDescription>
          这里只维护后台的期望状态，不会直接修改 Cloudflare、Hostinger 或生产网关。
        </DialogDescription>
      </DialogHeader>

      <form class="mt-4 space-y-4" @submit.prevent="emit('save')">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="域名" required description="只填写 hostname，例如 admin.learn.gripe。">
            <Input v-model="form.domain" placeholder="learn.gripe" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="角色" required>
            <select v-model="form.role" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option v-for="option in opsDomainRoleOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField label="环境" required>
            <select v-model="form.environment" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option v-for="option in opsDomainEnvironmentOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField
            label="所属项目"
            :required="requiresProject"
            description="部署 preflight 只采信显式绑定到项目的公网域名。"
          >
            <select
              v-model="form.project_binding_id"
              class="h-10 w-full rounded-md border bg-background px-3 text-sm"
              :disabled="saving || projectsLoading"
            >
              <option :value="null">{{ requiresProject ? '请选择项目' : '不绑定项目' }}</option>
              <option v-for="project in matchingProjects" :key="project.id" :value="project.id">
                {{ domainProjectLabel(project) }}
              </option>
            </select>
            <p v-if="projectsLoading" class="mt-1 text-[10px] text-muted-foreground">正在加载项目...</p>
            <p v-else-if="requiresProject && matchingProjects.length === 0" class="mt-1 text-[10px] text-amber-600">
              当前环境没有可绑定项目，请先在项目中心登记。
            </p>
          </AdminFormField>
          <AdminFormField label="提供商" required>
            <select v-model="form.provider" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option v-for="option in opsDomainProviderOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField v-if="form.provider === 'cloudflare'" label="Cloudflare 只读连接器">
            <select v-model="form.connector_id" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving || connectorsLoading">
              <option :value="null">暂不绑定</option>
              <option v-for="connector in cloudflareConnectors" :key="connector.id" :value="connector.id">
                {{ domainConnectorLabel(connector) }}
              </option>
            </select>
            <p v-if="connectorsLoading" class="mt-1 text-[10px] text-muted-foreground">正在加载连接器...</p>
            <p v-else-if="cloudflareConnectors.length === 0" class="mt-1 text-[10px] text-amber-600">
              请先在连接器中心登记当前环境的 Cloudflare 只读 Token。
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
              <option v-for="option in opsDomainProxyOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField label="TLS 模式">
            <select v-model="form.tls_mode" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option v-for="option in opsDomainTLSOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField v-if="form.role === 'redirect'" label="跳转目标" required>
            <Input v-model="form.redirect_target" placeholder="https://learn.gripe" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="状态">
            <select v-model="form.status" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option v-for="option in opsDomainStatusOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField label="备注" class="md:col-span-2">
            <Textarea v-model="form.notes" class="min-h-24" placeholder="记录绑定来源、切换窗口或维护说明。" :disabled="saving" />
          </AdminFormField>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" :disabled="saving" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="saving || !canEdit">
            <LoaderCircle v-if="saving" class="size-4 animate-spin" />
            <Save v-else class="size-4" />
            {{ saving ? '保存中' : '保存域名' }}
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
import type { OpsConnector, OpsProject } from '@/api/ops'
import {
  domainConnectorLabel,
  domainProjectLabel,
  domainRoleRequiresProject,
  opsDomainEnvironmentOptions,
  opsDomainProviderOptions,
  opsDomainProxyOptions,
  opsDomainRoleOptions,
  opsDomainStatusOptions,
  opsDomainTLSOptions,
  type OpsDomainForm,
} from './opsDomainBindingForm'

const props = withDefaults(defineProps<{
  open?: boolean
  form: OpsDomainForm
  projects?: OpsProject[]
  connectors?: OpsConnector[]
  projectsLoading?: boolean
  connectorsLoading?: boolean
  saving?: boolean
  canEdit?: boolean
}>(), {
  open: false,
  projects: () => [],
  connectors: () => [],
  projectsLoading: false,
  connectorsLoading: false,
  saving: false,
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'save'): void
}>()

const requiresProject = computed(() => domainRoleRequiresProject(props.form.role))
const matchingProjects = computed(() => props.projects.filter((project) => project.environment === props.form.environment))
const cloudflareConnectors = computed(() => props.connectors.filter((connector) => (
  connector.provider === 'cloudflare' &&
  connector.environment === props.form.environment
)))

watch(matchingProjects, (projects) => {
  if (!props.projects.length) return
  if (props.form.project_binding_id && !projects.some((project) => project.id === props.form.project_binding_id)) {
    props.form.project_binding_id = null
  }
}, { immediate: true })

watch([requiresProject, matchingProjects], ([required, projects]) => {
  if (required && props.projects.length && props.form.project_binding_id === null && projects.length === 1) {
    props.form.project_binding_id = projects[0].id
  }
}, { immediate: true })

watch(cloudflareConnectors, (connectors) => {
  if (props.form.provider !== 'cloudflare') {
    props.form.connector_id = null
    return
  }
  if (!props.connectors.length) return
  if (props.form.connector_id && !connectors.some((connector) => connector.id === props.form.connector_id)) {
    props.form.connector_id = null
  }
}, { immediate: true })
</script>
