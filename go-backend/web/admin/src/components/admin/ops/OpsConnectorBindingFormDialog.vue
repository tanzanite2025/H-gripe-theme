<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto">
      <DialogHeader>
        <DialogTitle>{{ form.id ? '编辑连接器' : '新增连接器' }}</DialogTitle>
        <DialogDescription>
          凭据只写入后端加密存储。编辑时凭据字段留空会保留原值，页面不会回显旧 Token。
        </DialogDescription>
      </DialogHeader>

      <form class="mt-4 space-y-4" @submit.prevent="emit('save')">
        <div class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="连接名称" required>
            <Input v-model="form.name" placeholder="Cloudflare Production" :disabled="saving" />
          </AdminFormField>
          <AdminFormField label="提供商" required>
            <select v-model="form.provider" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option v-for="option in opsConnectorProviderOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField label="环境">
            <select v-model="form.environment" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option v-for="option in opsConnectorEnvironmentOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField label="认证方式">
            <select v-model="form.auth_type" class="h-10 w-full rounded-md border bg-background px-3 text-sm" :disabled="saving">
              <option v-for="option in opsConnectorAuthTypeOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
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
              <option v-for="option in opsConnectorStatusOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
            </select>
          </AdminFormField>
          <AdminFormField label="备注" class="md:col-span-2">
            <Textarea v-model="form.notes" class="min-h-24" placeholder="记录连接用途、权限边界和轮换说明。" :disabled="saving" />
          </AdminFormField>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" :disabled="saving" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="saving || !canEdit">
            <LoaderCircle v-if="saving" class="size-4 animate-spin" />
            <Save v-else class="size-4" />
            {{ saving ? '保存中' : '保存连接' }}
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
import {
  connectorCredentialFields,
  opsConnectorAuthTypeOptions,
  opsConnectorEnvironmentOptions,
  opsConnectorProviderOptions,
  opsConnectorStatusOptions,
  type OpsConnectorForm,
} from '@/modules/ops/opsConnectorBindingForm'

const props = withDefaults(defineProps<{
  open?: boolean
  form: OpsConnectorForm
  saving?: boolean
  canEdit?: boolean
}>(), {
  open: false,
  saving: false,
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'save'): void
}>()

const credentialFields = computed(() => connectorCredentialFields(props.form.auth_type))

watch(() => props.form.auth_type, (authType) => {
  if (authType === 'environment') {
    props.form.credentials = { token: '', api_key: '', username: '', password: '' }
  }
})
</script>

