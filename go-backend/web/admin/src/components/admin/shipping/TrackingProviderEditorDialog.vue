<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="full" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-6" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '新增追踪配置' : '编辑追踪配置' }}</DialogTitle>
          <DialogDescription>
            统一维护 17TRACK / AfterShip 等追踪 Provider 的接口地址、API Key、Webhook 和同步策略；承运商档案不再保存这些凭证。
          </DialogDescription>
        </DialogHeader>

        <section class="grid gap-4 lg:grid-cols-4">
          <AdminFormField label="Provider 名称" required :error="errors.provider_name" class="lg:col-span-2">
            <Input
              v-model.trim="form.provider_name"
              placeholder="例如 17TRACK"
              @input="emit('clear-error', 'provider_name')"
            />
          </AdminFormField>

          <AdminFormField label="Provider 代码" required :error="errors.provider_code">
            <Input
              v-model.trim="form.provider_code"
              class="font-mono uppercase"
              placeholder="17TRACK"
              @input="emit('clear-error', 'provider_code')"
            />
          </AdminFormField>

          <AdminFormField label="环境" required :error="errors.environment">
            <Select v-model="form.environment" @update:model-value="emit('clear-error', 'environment')">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择环境" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="production">Production</SelectItem>
                <SelectItem value="sandbox">Sandbox</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="Base URL" class="lg:col-span-3">
            <Input v-model.trim="form.base_url" class="font-mono" placeholder="例如 https://api.provider.example" />
          </AdminFormField>

          <AdminFormField label="排序">
            <Input v-model.number="form.sort_order" type="number" min="0" step="1" />
          </AdminFormField>

          <AdminFormField label="API Key" class="lg:col-span-2">
            <Input v-model.trim="form.api_key" class="font-mono" type="password" autocomplete="new-password" />
            <p v-if="mode === 'edit' && form.api_key_configured" class="mt-1 text-[10px] text-muted-foreground">
              已配置 API Key；留空保存时会保持原值，输入新值才覆盖。
            </p>
          </AdminFormField>

          <AdminFormField label="Webhook Secret" class="lg:col-span-2">
            <Input v-model.trim="form.webhook_secret" class="font-mono" type="password" autocomplete="new-password" />
            <p v-if="mode === 'edit' && form.webhook_secret_configured" class="mt-1 text-[10px] text-muted-foreground">
              已配置 Webhook Secret；留空保存时会保持原值，输入新值才覆盖。
            </p>
          </AdminFormField>

          <AdminFormField label="Webhook 回调地址" class="lg:col-span-4">
            <Input
              :model-value="webhookUrl || '填写 Provider 代码后自动生成'"
              class="font-mono text-[11px]"
              readonly
            />
          </AdminFormField>

          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-bold uppercase tracking-wider">启用配置 / ENABLED</span>
              <p class="mt-0.5 text-xs text-muted-foreground">停用后不会参与追踪号注册和轨迹同步。</p>
            </div>
            <Switch v-model="form.enabled" aria-label="启用追踪 Provider" />
          </div>

          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-bold uppercase tracking-wider">Webhook</span>
              <p class="mt-0.5 text-xs text-muted-foreground">后续由 Provider 主动推送轨迹事件。</p>
            </div>
            <Switch v-model="form.webhook_enabled" aria-label="启用 Webhook" />
          </div>

          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-bold uppercase tracking-wider">自动注册</span>
              <p class="mt-0.5 text-xs text-muted-foreground">发货后自动把追踪号注册到 Provider。</p>
            </div>
            <Switch v-model="form.auto_register" aria-label="启用自动注册追踪号" />
          </div>

          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-bold uppercase tracking-wider">轮询同步</span>
              <p class="mt-0.5 text-xs text-muted-foreground">Webhook 不可用时按周期拉取轨迹。</p>
            </div>
            <Switch v-model="form.polling_enabled" aria-label="启用轮询同步" />
          </div>

          <AdminFormField label="轮询间隔 分钟">
            <Input v-model.number="form.polling_interval_minutes" type="number" min="1" step="1" />
          </AdminFormField>

          <AdminFormField label="请求超时 秒">
            <Input v-model.number="form.request_timeout_seconds" type="number" min="1" step="1" />
          </AdminFormField>

          <AdminFormField label="说明" class="lg:col-span-4">
            <Textarea
              v-model="form.description"
              class="min-h-24"
              placeholder="内部备注、Provider 套餐、Webhook 回调路径、后续承运商代码映射说明"
            />
          </AdminFormField>
        </section>

        <div class="rounded-lg border bg-muted/35 p-3 text-xs text-muted-foreground">
          稳定规则：承运商只维护物流公司资料；线路服务维护运费和渠道；追踪 Provider 只维护第三方追踪 API 凭证和同步策略。
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存追踪配置' }}
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'

defineProps({
  open: { type: Boolean, default: false },
  mode: { type: String, default: 'create' },
  form: { type: Object, required: true },
  errors: { type: Object, required: true },
  webhookUrl: { type: String, default: '' },
  submitting: { type: Boolean, default: false },
})

const emit = defineEmits(['update:open', 'submit', 'clear-error'])
</script>
