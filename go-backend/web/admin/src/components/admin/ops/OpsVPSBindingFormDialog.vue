<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>{{ form.id ? '编辑 VPS 绑定' : '新增 VPS 绑定' }}</DialogTitle>
        <DialogDescription>
          这里记录 Hostinger 或其他提供商的资源基线。观察状态只有同步后才代表实际状态。
        </DialogDescription>
      </DialogHeader>

      <form class="mt-4 space-y-4" @submit.prevent="emit('save')">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="资源名称" required>
            <Input v-model="form.name" placeholder="Hostinger Production VPS" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="提供商" required>
            <select v-model="form.provider" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option v-for="option in opsVPSProviderOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField label="环境" required>
            <select v-model="form.environment" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option v-for="option in opsVPSEnvironmentOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField label="连接器">
            <select v-model="form.connector_id" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option :value="null">未绑定连接器</option>
              <option v-for="connector in compatibleConnectors" :key="connector.id" :value="connector.id">
                {{ connector.name }} · {{ connector.environment }}
              </option>
            </select>
          </AdminFormField>
          <AdminFormField label="Hostinger VPS ID / 资源 ID">
            <Input v-model="form.provider_resource_id" placeholder="1834903" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="主机名">
            <Input v-model="form.hostname" placeholder="srv1834903.hstgr.cloud" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="IPv4">
            <Input v-model="form.ipv4" placeholder="2.25.85.201" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="区域">
            <Input v-model="form.region" placeholder="Hostinger region" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="操作系统">
            <Input v-model="form.operating_system" placeholder="Ubuntu 24.04 LTS" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="期望状态">
            <select v-model="form.status" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option v-for="option in opsVPSStatusOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField label="备注" class="md:col-span-2">
            <Textarea v-model="form.notes" class="min-h-24" placeholder="记录区域、维护窗口、资源边界或确认来源。" :disabled="saving" />
          </AdminFormField>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" :disabled="saving" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="saving || !canEdit">
            <LoaderCircle v-if="saving" class="size-4 animate-spin" />
            <Save v-else class="size-4" />
            {{ saving ? '保存中' : '保存 VPS' }}
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
import type { OpsConnector } from '@/api/ops'
import {
  opsVPSEnvironmentOptions,
  opsVPSProviderOptions,
  opsVPSStatusOptions,
  type OpsVPSForm,
} from '@/modules/ops/opsVPSBindingForm'

const props = withDefaults(defineProps<{
  open?: boolean
  form: OpsVPSForm
  connectors?: OpsConnector[]
  saving?: boolean
  canEdit?: boolean
}>(), {
  open: false,
  connectors: () => [],
  saving: false,
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'save'): void
}>()

const compatibleConnectors = computed(() => props.connectors.filter((connector) => (
  connector.provider === props.form.provider &&
  connector.environment === props.form.environment
)))

watch(
  compatibleConnectors,
  (connectors) => {
    if (!props.connectors.length) {
      return
    }
    if (props.form.connector_id && !connectors.some((connector) => connector.id === props.form.connector_id)) {
      props.form.connector_id = null
    }
  },
  { immediate: true },
)
</script>

