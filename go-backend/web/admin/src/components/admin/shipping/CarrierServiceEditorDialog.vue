<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="full" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-6" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '新增线路服务' : '编辑线路服务' }}</DialogTitle>
          <DialogDescription>
            线路服务连接承运商和运费模板，负责记录国际线路、计费方式、首续重、体积重和时效参数。
          </DialogDescription>
        </DialogHeader>

        <section v-if="selectedCarrierIsFpx" class="rounded-xl border border-orange-500/30 bg-orange-500/5 p-3">
          <AdminFormField label="从 4PX 服务集合选择">
            <Select
              :model-value="form.service_code"
              :disabled="fpxChannels.length === 0"
              @update:model-value="applyFpxChannel"
            >
              <SelectTrigger class="w-full"><SelectValue :placeholder="fpxChannels.length ? '选择已启用的官方服务' : '暂无已发布服务，请先在 4PX 服务集合启用'" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="channel in fpxChannels" :key="channel.id" :value="channel.service_code">
                  {{ channel.display_name }} / {{ channel.service_code }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>
          <p class="mt-2 text-[11px] leading-5 text-muted-foreground">只读取 4PX「服务集合」已启用项；选择后填入线路代码和名称，物流模板、计费和报价规则仍由物流管理维护。</p>
        </section>

        <section class="grid gap-4 lg:grid-cols-4">
          <AdminFormField label="承运商" required :error="errors.carrier_id">
            <Select v-model="form.carrier_id" @update:model-value="emit('clear-error', 'carrier_id')">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择承运商" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="carrier in carriers" :key="carrier.id" :value="String(carrier.id)">
                  {{ carrier.name }} / {{ carrier.code }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="关联运费模板">
            <Select v-model="form.template_id">
              <SelectTrigger class="w-full"><SelectValue placeholder="可选" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="none">暂不绑定模板</SelectItem>
                <SelectItem v-for="template in templates" :key="template.id" :value="String(template.id)">
                  {{ template.name }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="线路代码" required :error="errors.service_code">
            <Input
              v-model.trim="form.service_code"
              class="font-mono uppercase"
              placeholder="DHL-EXP-US"
              @input="emit('clear-error', 'service_code')"
            />
          </AdminFormField>

          <AdminFormField label="排序">
            <Input v-model.number="form.sort_order" type="number" min="0" step="1" />
          </AdminFormField>

          <AdminFormField label="线路名称" required :error="errors.service_name" class="lg:col-span-2">
            <Input v-model.trim="form.service_name" placeholder="例如 DHL Express 美国线" @input="emit('clear-error', 'service_name')" />
          </AdminFormField>

          <AdminFormField label="线路/渠道">
            <Input v-model.trim="form.route_name" placeholder="例如 空运快线 / 邮政小包 / 专线" />
          </AdminFormField>

          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5">
            <div>
              <span class="text-xs font-bold uppercase tracking-wider">启用线路 / ENABLED</span>
              <p class="mt-0.5 text-xs text-muted-foreground">停用后前台报价不会把它作为可用线路展示。</p>
            </div>
            <Switch v-model="form.enabled" aria-label="启用线路服务" />
          </div>

          <AdminFormField
            label="国家/区域"
            class="lg:col-span-2"
            description="JSON 数组或逗号分隔；例如 US, CA。为空代表暂未限制。"
          >
 <Textarea v-model="form.countries" class="min-h-20 font-mono text-xs" placeholder='["US","CA","EU"]'/>
          </AdminFormField>

          <AdminFormField label="计费模式" required :error="errors.billing_mode">
            <Select v-model="form.billing_mode" @update:model-value="emit('clear-error', 'billing_mode')">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择计费模式" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="actual_weight">实重计费</SelectItem>
                <SelectItem value="volumetric_weight">体积重计费</SelectItem>
                <SelectItem value="greater_of_actual_and_volumetric">实重/体积重取大</SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>

          <AdminFormField label="币种">
            <Input v-model.trim="form.currency" class="font-mono uppercase" maxlength="10" placeholder="ISO 4217" />
          </AdminFormField>

          <AdminFormField label="首重 g">
            <Input v-model.number="form.first_weight_grams" type="number" min="0" step="1" />
          </AdminFormField>

          <AdminFormField label="续重单位 g">
            <Input v-model.number="form.additional_weight_grams" type="number" min="0" step="1" />
          </AdminFormField>

          <AdminFormField label="最低计费重 g">
            <Input v-model.number="form.min_charge_weight_grams" type="number" min="0" step="1" />
          </AdminFormField>

          <AdminFormField label="体积重除数">
            <Input v-model.number="form.volumetric_divisor" type="number" min="1" step="1" />
          </AdminFormField>

          <AdminFormField label="燃油附加 %">
            <Input v-model.trim="form.fuel_surcharge_percent_decimal" type="number" min="0" step="0.001" />
          </AdminFormField>

          <AdminFormField label="偏远附加费">
            <Input v-model.number="form.remote_surcharge_minor" type="number" min="0" step="1" />
          </AdminFormField>

          <AdminFormField
            label="偏远邮编段"
            class="lg:col-span-2"
            description="JSON 数组，支持精确邮编、前缀 * 或范围；为空时对该线路全部目的地收取偏远附加费。"
          >
            <Textarea v-model="form.remote_postal_codes" class="min-h-20 font-mono text-xs" placeholder='["10000-10099","967*"]' />
          </AdminFormField>

          <AdminFormField label="最短时效 天">
            <Input v-model.number="form.eta_min_days" type="number" min="0" step="1" />
          </AdminFormField>

          <AdminFormField label="最长时效 天" :error="errors.eta_max_days">
            <Input v-model.number="form.eta_max_days" type="number" min="0" step="1" @input="emit('clear-error', 'eta_max_days')" />
          </AdminFormField>

          <AdminFormField label="说明" class="lg:col-span-4">
            <Textarea v-model="form.description" class="min-h-24" placeholder="内部备注、结算口径、特殊限制或后续 17TRACK 映射说明" />
          </AdminFormField>
        </section>

        <div class="rounded-lg border bg-muted/35 p-3 text-xs text-muted-foreground">
          稳定规则：SKU 提供实际重量，包装规则提供箱规尺寸，线路服务提供计费口径。后续接真实报价时按线路服务计算，不再在 Nuxt 里硬编码运费。
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? '保存中' : '保存线路服务' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, toRefs } from 'vue'
import { LoaderCircle } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import type {
  ShippingCarrier,
  ShippingCarrierServiceForm,
  ShippingDialogMode,
  ShippingErrorMap,
  ShippingTemplate
} from '@/modules/shipping/shippingTypes'

const props = withDefaults(defineProps<{
  open?: boolean
  mode?: ShippingDialogMode
  form: ShippingCarrierServiceForm
  errors: ShippingErrorMap
  carriers?: ShippingCarrier[]
  fpxChannels?: Array<{ id: number; service_code: string; display_name: string; enabled: boolean }>
  templates?: ShippingTemplate[]
  submitting?: boolean
}>(), {
  open: false,
  mode: 'create',
  carriers: () => [],
  fpxChannels: () => [],
  templates: () => [],
  submitting: false
})
const { open, mode, form, errors, carriers, templates, submitting } = toRefs(props)

const selectedCarrierIsFpx = computed(() => {
  const selectedCarrier = props.carriers.find((carrier) => String(carrier.id) === String(props.form.carrier_id))
  return ['4PX', 'FPX'].includes(String(selectedCarrier?.code || '').trim().toUpperCase())
})

const applyFpxChannel = (serviceCode: unknown) => {
  const channel = props.fpxChannels.find((item) => item.service_code === String(serviceCode))
  if (!channel) return
  props.form.service_code = channel.service_code
  props.form.service_name = channel.display_name
}

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit'): void
  (event: 'clear-error', field: string): void
}>()
</script>

