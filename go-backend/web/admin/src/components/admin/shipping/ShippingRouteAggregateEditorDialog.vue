<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg">
      <form class="space-y-5" @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ mode === 'create' ? '新增聚合线路' : '编辑聚合线路' }}</DialogTitle>
          <DialogDescription>
            聚合线路只登记 4PX、燕文等承运商的已发布线路身份、国家范围和启用状态，不保存向客户收取的运费。
          </DialogDescription>
        </DialogHeader>

        <section class="grid gap-4 md:grid-cols-2">
          <AdminFormField label="聚合线路名称" required :error="errors.name">
            <Input v-model.trim="form.name" placeholder="例如 4PX 美国标准线路" @input="emit('clear-error', 'name')" />
          </AdminFormField>
          <AdminFormField label="排序">
            <Input v-model.number="form.sort_order" type="number" min="0" step="1" />
          </AdminFormField>
          <AdminFormField label="已发布线路" required :error="errors.published_carrier_services" description="选择 4PX 或燕文服务集合中的线路；可选择多条。国家范围随所选服务读取。" class="md:col-span-2">
            <select v-model="selectedPublishedRouteKeys" multiple class="min-h-36 w-full rounded-md border bg-background px-3 py-2 text-sm">
              <option v-for="channel in publishedRouteChoices" :key="channel.key" :value="channel.key">{{ channel.label }}</option>
            </select>
            <p v-if="!publishedRouteChoices.length" class="mt-2 text-xs text-muted-foreground">请先在 4PX 或燕文的服务集合发布线路，并配置对应承运商。</p>
          </AdminFormField>
          <AdminFormField label="配送国家" description="只读取已发布线路的国家范围。" class="md:col-span-2">
            <div class="rounded-md border bg-muted/30 px-3 py-2 font-mono text-xs">{{ form.countries }}</div>
          </AdminFormField>
          <AdminFormField label="说明" class="md:col-span-2">
            <Textarea v-model="form.description" class="min-h-20" placeholder="线路身份、来源或渠道说明" />
          </AdminFormField>
          <div class="flex items-center justify-between gap-3 rounded-lg border px-3 py-2.5 md:col-span-2">
            <div>
              <span class="text-xs font-bold uppercase tracking-wider">启用聚合线路 / ENABLED</span>
              <p class="mt-0.5 text-xs text-muted-foreground">停用后不能被客户线路计费模板用于 Checkout 报价。</p>
            </div>
            <Switch v-model="form.enabled" aria-label="启用聚合线路" />
          </div>
        </section>

        <div class="rounded-lg border border-blue-500/25 bg-blue-500/5 p-3 text-xs leading-5 text-muted-foreground">
          4PX/燕文的内部结算成本属于承运商侧数据；本页不会把它当作客户运费。客户实际支付金额只由“线路计费模板”决定。
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">{{ submitting ? '保存中' : '保存聚合线路' }}</Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { FpxPublishedCollectionReference, YanwenPublishedCollectionReference } from '@/api/shippingServiceCollectionReferenceApi'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import type { ShippingCarrier, ShippingCarrierService, ShippingDialogMode, ShippingErrorMap, ShippingRouteAggregateForm } from '@/modules/shipping/shippingTypes'

const props = withDefaults(defineProps<{
  open?: boolean
  mode?: ShippingDialogMode
  form: ShippingRouteAggregateForm
  errors: ShippingErrorMap
  carrierServices?: ShippingCarrierService[]
  carriers?: ShippingCarrier[]
  fpxChannels?: FpxPublishedCollectionReference[]
  yanwenPublishedChannels?: YanwenPublishedCollectionReference[]
  submitting?: boolean
}>(), {
  open: false,
  mode: 'create',
  carrierServices: () => [],
  carriers: () => [],
  fpxChannels: () => [],
  yanwenPublishedChannels: () => [],
  submitting: false,
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit'): void
  (event: 'clear-error', field: string): void
}>()

const publishedRouteChoices = computed(() => {
  const carrierFor = (provider: string) => props.carriers.find((carrier) => provider === '4PX' ? ['FPX', '4PX'].includes(String(carrier.code).toUpperCase()) : String(carrier.code).toUpperCase() === provider)
  const fpxCarrier = carrierFor('4PX')
  const yanwenCarrier = carrierFor('YANWEN')
  return [
    ...(fpxCarrier ? props.fpxChannels.map((channel) => ({ key: '4PX:' + channel.id, label: '4PX · ' + channel.display_name + ' / ' + channel.service_code, countries: channel.countries, service: { carrier_id: fpxCarrier.id, fpx_channel_id: channel.id, provider_code: '4PX' } })) : []),
    ...(yanwenCarrier ? props.yanwenPublishedChannels.map((channel) => ({ key: 'YANWEN:' + channel.id, label: '燕文 · ' + channel.display_name + ' / ' + channel.product_code, countries: channel.countries, service: { carrier_id: yanwenCarrier.id, yanwen_published_channel_id: channel.id, provider_code: 'YANWEN' } })) : []),
  ]
})
const selectedPublishedRouteKeys = computed({
  get: () => props.form.published_carrier_services.map((service) => service.fpx_channel_id ? '4PX:' + service.fpx_channel_id : 'YANWEN:' + service.yanwen_published_channel_id),
  set: (keys: string[]) => {
    const choices = publishedRouteChoices.value.filter((choice) => keys.includes(choice.key))
    props.form.published_carrier_services = choices.map((choice) => choice.service)
    const countries = new Set<string>()
    for (const choice of choices) {
      try { for (const country of JSON.parse(String(choice.countries || '[]'))) countries.add(String(country)) } catch { /* The server validates collection country data. */ }
    }
    props.form.countries = JSON.stringify(Array.from(countries).sort())
    emit('clear-error', 'published_carrier_services')
  },
})
</script>
