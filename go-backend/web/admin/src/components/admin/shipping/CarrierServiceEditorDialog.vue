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

        <section v-if="selectedCarrierIsYanwen" class="rounded-xl border border-orange-500/30 bg-orange-500/5 p-3">
          <AdminFormField label="从燕文服务集合选择">
            <Select
              :model-value="form.yanwen_published_channel_id ? String(form.yanwen_published_channel_id) : ''"
              :disabled="yanwenPublishedChannels.length === 0"
              @update:model-value="applyYanwenChannel"
            >
              <SelectTrigger class="w-full"><SelectValue :placeholder="yanwenPublishedChannels.length ? '选择已启用的官方服务' : '暂无已发布服务，请先在燕文服务集合启用'" /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="channel in yanwenPublishedChannels" :key="channel.id" :value="String(channel.id)">
                  {{ channel.display_name }} / {{ channel.product_code }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>
          <p class="mt-2 text-[11px] leading-5 text-muted-foreground">只读取燕文「服务集合」已启用项；选择后填入产品代码、名称和配送地区。</p>
        </section>

        <section class="grid gap-4 lg:grid-cols-4">
          <AdminFormField label="承运商" required :error="errors.carrier_id">
            <Select v-model="form.carrier_id" @update:model-value="handleCarrierSelectionChange">
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
                <SelectItem v-for="template in carrierBindableTemplates" :key="template.id" :value="String(template.id)">
                  {{ template.name }}
                </SelectItem>
              </SelectContent>
            </Select>
            <p class="mt-1 text-[11px] leading-4 text-muted-foreground">系统免邮模板只按国家范围控制下单，不绑定具体承运线路。</p>
          </AdminFormField>

          <AdminFormField label="线路代码" required :error="errors.service_code">
            <Input
              v-model.trim="form.service_code"
              class="font-mono uppercase"
              placeholder="DHL-EXP-US"
              :readonly="selectedCarrierUsesPublishedCollection"
              :class="{ 'bg-muted/30 text-muted-foreground': selectedCarrierUsesPublishedCollection }"
              @input="emit('clear-error', 'service_code')"
            />
            <p v-if="selectedCarrierUsesPublishedCollection" class="mt-1 text-[11px] leading-4 text-muted-foreground">线路代码由已选服务集合记录提供；需要更换代码请切换精选服务。</p>
          </AdminFormField>

          <AdminFormField label="排序">
            <Input v-model.number="form.sort_order" type="number" min="0" step="1" />
          </AdminFormField>

          <AdminFormField label="线路名称" required :error="errors.service_name" class="lg:col-span-2">
            <Input
              v-model.trim="form.service_name"
              placeholder="例如 DHL Express 美国线"
              :readonly="selectedCarrierUsesPublishedCollection"
              :class="{ 'bg-muted/30 text-muted-foreground': selectedCarrierUsesPublishedCollection }"
              @input="emit('clear-error', 'service_name')"
            />
            <p v-if="selectedCarrierUsesPublishedCollection" class="mt-1 text-[11px] leading-4 text-muted-foreground">线路名称由服务集合提供；业务备注请填写在线路/渠道或说明。</p>
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
            v-if="selectedCarrierUsesPublishedCollection"
            label="配送地区"
            class="lg:col-span-2"
            description="地区由 4PX/燕文服务集合提供，不能在物流管理中编辑。"
          >
            <div class="flex min-h-20 items-start rounded-md border bg-muted/30 px-3 py-2 font-mono text-xs text-muted-foreground">
              {{ serviceCountriesLabel(form.countries) }}
            </div>
          </AdminFormField>

          <AdminFormField
            v-else
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
import type {
  FpxPublishedCollectionReference,
  YanwenPublishedCollectionReference,
} from '@/api/shippingServiceCollectionReferenceApi'

const props = withDefaults(defineProps<{
  open?: boolean
  mode?: ShippingDialogMode
  form: ShippingCarrierServiceForm
  errors: ShippingErrorMap
  carriers?: ShippingCarrier[]
  fpxChannels?: FpxPublishedCollectionReference[]
  yanwenPublishedChannels?: YanwenPublishedCollectionReference[]
  templates?: ShippingTemplate[]
  submitting?: boolean
}>(), {
  open: false,
  mode: 'create',
  carriers: () => [],
  fpxChannels: () => [],
  yanwenPublishedChannels: () => [],
  templates: () => [],
  submitting: false
})
const { open, mode, form, errors, carriers, templates, submitting } = toRefs(props)

const selectedCarrierIsFpx = computed(() => {
  const selectedCarrier = props.carriers.find((carrier) => String(carrier.id) === String(props.form.carrier_id))
  return ['4PX', 'FPX'].includes(String(selectedCarrier?.code || '').trim().toUpperCase())
})
const selectedCarrierIsYanwen = computed(() => {
  const selectedCarrier = props.carriers.find((carrier) => String(carrier.id) === String(props.form.carrier_id))
  return String(selectedCarrier?.code || '').trim().toUpperCase() === 'YANWEN'
})
const selectedCarrierUsesPublishedCollection = computed(() => {
  const selectedCarrier = props.carriers.find((carrier) => String(carrier.id) === String(props.form.carrier_id))
  const carrierCode = String(selectedCarrier?.code || '').trim().toUpperCase()
  return ['4PX', 'FPX', 'YANWEN'].includes(carrierCode) || String(props.form.service_code || '').trim().toUpperCase().startsWith('YANWEN:')
})
const carrierBindableTemplates = computed(() => props.templates.filter((template) => (
  template.template_kind !== 'system_free_shipping' && template.type !== 'free_shipping'
)))

const serviceCountriesLabel = (value: unknown) => {
  const raw = String(value || '').trim()
  if (!raw) return '未限制（服务集合未提供地区）'
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed) && parsed.length) return parsed.join(', ')
  } catch {
    // Keep compatibility with older comma-separated records.
  }
  return raw.replace(/[\[\]"']/g, '').replace(/[,，;|]+/g, ', ')
}

const handleCarrierSelectionChange = (carrierID: unknown) => {
  const selectedCarrier = props.carriers.find((carrier) => String(carrier.id) === String(carrierID))
  const carrierCode = String(selectedCarrier?.code || '').trim().toUpperCase()
  props.form.fpx_channel_id = null
  props.form.yanwen_published_channel_id = null
  props.form.provider_code = ['4PX', 'FPX'].includes(carrierCode) ? '4PX' : carrierCode
  emit('clear-error', 'carrier_id')
}

const applyFpxChannel = (serviceCode: unknown) => {
  const channel = props.fpxChannels.find((item) => item.service_code === String(serviceCode))
  if (!channel) return
  props.form.service_code = channel.service_code
  props.form.service_name = channel.display_name
  props.form.provider_code = '4PX'
  props.form.fpx_channel_id = channel.id
  props.form.yanwen_published_channel_id = null
  props.form.countries = channel.countries || '[]'
}

const applyYanwenChannel = (channelID: unknown) => {
  const channel = props.yanwenPublishedChannels.find((item) => String(item.id) === String(channelID))
  if (!channel) return
  const productCode = String(channel.product_code || '').trim().replace(/^YANWEN:/i, '')
  props.form.service_code = `YANWEN:${productCode}`
  props.form.service_name = channel.display_name
  props.form.provider_code = 'YANWEN'
  props.form.fpx_channel_id = null
  props.form.yanwen_published_channel_id = channel.id
  props.form.countries = channel.countries || '[]'
}

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit'): void
  (event: 'clear-error', field: string): void
}>()
</script>
