<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="full" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-6" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '新增承运商映射' : '编辑承运商映射' }}</DialogTitle>
          <DialogDescription>
            把本地承运商或线路服务映射到 17TRACK / AfterShip 等 Provider 的 carrier code；后续注册追踪号时统一读取这里。
          </DialogDescription>
        </DialogHeader>

        <section class="grid gap-4 lg:grid-cols-4">
          <AdminFormField label="追踪 Provider" required :error="errors.provider_id">
            <Select v-model="form.provider_id" @update:model-value="emit('clear-error', 'provider_id')">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择 Provider" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="provider in providers" :key="provider.id" :value="String(provider.id)">
                  {{ provider.provider_name }} / {{ provider.provider_code }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="映射层级" required :error="errors.scope">
            <Select v-model="form.scope" @update:model-value="emit('clear-error', 'scope')">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择层级" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="carrier">承运商</SelectItem>
                <SelectItem value="carrier_service">线路服务</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField v-if="form.scope === 'carrier'" label="本地承运商" required :error="errors.carrier_id" class="lg:col-span-2">
            <Select v-model="form.carrier_id" @update:model-value="emit('clear-error', 'carrier_id')">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择承运商" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="carrier in carriers" :key="carrier.id" :value="String(carrier.id)">
                  {{ carrier.name }} / {{ carrier.code }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField v-else label="本地线路服务" required :error="errors.carrier_service_id" class="lg:col-span-2">
            <Select v-model="form.carrier_service_id" @update:model-value="emit('clear-error', 'carrier_service_id')">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择线路服务" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="service in carrierServices" :key="service.id" :value="String(service.id)">
                  {{ service.service_name }} / {{ service.service_code }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="Provider Carrier Code" required :error="errors.provider_carrier_code" class="lg:col-span-2">
            <Input
              v-model.trim="form.provider_carrier_code"
              class="font-mono"
              placeholder="例如 190271 / dhl"
              @input="emit('clear-error', 'provider_carrier_code')"
            />
          </AdminFormField>

          <AdminFormField label="Provider Carrier Name">
            <Input v-model.trim="form.provider_carrier_name" placeholder="例如 DHL Express" />
          </AdminFormField>

          <AdminFormField label="优先级">
            <Input v-model.number="form.priority" type="number" min="0" step="1" />
          </AdminFormField>

          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-bold uppercase tracking-wider">启用映射 / ENABLED</span>
              <p class="mt-0.5 text-xs text-muted-foreground">停用后不会用于追踪号注册或轨迹查询。</p>
            </div>
            <Switch v-model="form.enabled" aria-label="启用承运商映射" />
          </div>

          <AdminFormField label="说明" class="lg:col-span-3">
            <Textarea
              v-model="form.description"
              class="min-h-24"
              placeholder="Provider 代码来源、适用线路、特殊限制或内部备注"
            />
          </AdminFormField>
        </section>

        <div class="rounded-lg border bg-muted/35 p-3 text-xs text-muted-foreground">
          稳定规则：线路服务映射优先于承运商映射；没有线路级映射时，后续追踪同步可以回退到承运商级映射。
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存映射' }}
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
  providers: { type: Array, default: () => [] },
  carriers: { type: Array, default: () => [] },
  carrierServices: { type: Array, default: () => [] },
  submitting: { type: Boolean, default: false },
})

const emit = defineEmits(['update:open', 'submit', 'clear-error'])
</script>
