<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="full" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-6" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '新增承运商' : '编辑承运商' }}</DialogTitle>
          <DialogDescription>
            维护物流公司、官方查询链接和可选的承运商接口信息。17TRACK 统一配置会放在后续的追踪设置里。
          </DialogDescription>
        </DialogHeader>

        <div class="grid gap-4 lg:grid-cols-3">
          <AdminFormField label="承运商名称" required :error="errors.name">
            <Input v-model.trim="form.name" placeholder="例如 DHL Express" @input="emit('clear-error', 'name')" />
          </AdminFormField>

          <AdminFormField label="承运商代码" required :error="errors.code">
            <Input
              v-model.trim="form.code"
              class="font-mono uppercase"
              placeholder="例如 DHL"
              @input="emit('clear-error', 'code')"
            />
          </AdminFormField>

          <AdminFormField label="排序">
            <Input v-model.number="form.sort_order" type="number" min="0" step="1" />
          </AdminFormField>

          <AdminFormField label="查询链接模板" class="lg:col-span-2" description="可包含 {tracking_number} 占位符。">
            <Input v-model.trim="form.tracking_url" placeholder="https://example.com/track/{tracking_number}" />
          </AdminFormField>

          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-medium">启用承运商</span>
              <p class="mt-0.5 text-xs text-muted-foreground">停用后前台和后台默认不再作为可选物流公司展示。</p>
            </div>
            <Switch v-model="form.enabled" aria-label="启用承运商" />
          </div>

          <AdminFormField label="联系人">
            <Input v-model.trim="form.contact" placeholder="业务联系人" />
          </AdminFormField>

          <AdminFormField label="电话">
            <Input v-model.trim="form.phone" type="tel" placeholder="+1 000 000 0000" />
          </AdminFormField>

          <AdminFormField label="邮箱">
            <Input v-model.trim="form.email" type="email" placeholder="support@example.com" />
          </AdminFormField>

          <AdminFormField label="承运商 API 地址" class="lg:col-span-3" description="承运商自有接口；17TRACK 全局 API 不放在这里。">
            <Input v-model.trim="form.api_endpoint" placeholder="https://api.carrier.example.com" />
          </AdminFormField>

          <AdminFormField label="API Key" class="lg:col-span-2">
            <Input v-model.trim="form.api_key" class="font-mono" autocomplete="off" />
          </AdminFormField>

          <AdminFormField label="API Secret">
            <Input v-model.trim="form.api_secret" class="font-mono" type="password" autocomplete="new-password" />
          </AdminFormField>

          <AdminFormField
            label="服务区域"
            class="lg:col-span-3"
            description="先按文本/JSON 维护，后续会和配送区域、线路模板打通。"
          >
            <Textarea v-model="form.service_area" class="min-h-24 font-mono text-xs" placeholder='["US", "CA", "EU"]' />
          </AdminFormField>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存承运商' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { LoaderCircle } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  form: { type: Object, required: true },
  errors: { type: Object, required: true },
  submitting: { type: Boolean, default: false },
})

const emit = defineEmits(['update:open', 'submit', 'clear-error'])
</script>
