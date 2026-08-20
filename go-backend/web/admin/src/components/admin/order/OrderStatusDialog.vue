<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg">
      <form @submit.prevent="emit('submit')">
        <DialogHeader>
          <DialogTitle>{{ fulfillmentMode ? '确认发货' : '状态管理' }}</DialogTitle>
          <DialogDescription>
            {{ fulfillmentMode ? `为订单 ${statusForm.order_number} 填写发货信息。` : `更新订单 ${statusForm.order_number} 的履约状态。` }}
          </DialogDescription>
        </DialogHeader>

        <div class="space-y-4 py-5">
          <label v-if="!fulfillmentMode" class="block space-y-1">
            <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">STATUS / 订单状态</span>
            <Select v-model="statusForm.status">
              <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="option in editableOrderStatusOptions" :key="option.value" :value="option.value">{{ option.label }}</SelectItem>
              </SelectContent>
            </Select>
          </label>

          <label v-if="!fulfillmentMode" class="block space-y-1">
            <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SHIPPING / 物流状态</span>
            <Select v-model="statusForm.shipping_status">
              <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem v-for="option in editableShippingStatusOptions" :key="option.value" :value="option.value">{{ option.label }}</SelectItem>
              </SelectContent>
            </Select>
          </label>

          <label class="block space-y-1">
            <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">TRACKING / 物流单号</span>
            <Input v-model="statusForm.tracking_number" />
          </label>

          <label class="block space-y-1">
            <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">PROVIDER / 追踪服务商</span>
            <Select v-model="statusForm.tracking_provider_id">
              <SelectTrigger class="w-full"><SelectValue placeholder="请选择追踪 Provider" /></SelectTrigger>
              <SelectContent>
                <SelectItem value="none">请选择追踪 Provider</SelectItem>
                <SelectItem v-for="provider in trackingProviders" :key="provider.id" :value="String(provider.id)">
                  {{ provider.provider_name }} / {{ provider.provider_code }}
                </SelectItem>
              </SelectContent>
            </Select>
          </label>

          <div class="grid gap-3 sm:grid-cols-2">
            <label class="block space-y-1">
              <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">CARRIER / 本地承运商</span>
              <Select v-model="statusForm.carrier_id" @update:model-value="(value) => emit('carrier-change', String(value))">
                <SelectTrigger class="w-full"><SelectValue placeholder="可选：选择承运商" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">不指定承运商</SelectItem>
                  <SelectItem v-for="carrier in carriers" :key="carrier.id" :value="String(carrier.id)">
                    {{ carrier.name }} / {{ carrier.code }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </label>

            <label class="block space-y-1">
              <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SERVICE / 线路服务</span>
              <Select v-model="statusForm.carrier_service_id" @update:model-value="(value) => emit('carrier-service-change', String(value))">
                <SelectTrigger class="w-full"><SelectValue placeholder="可选：选择线路服务" /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">不指定线路服务</SelectItem>
                  <SelectItem v-for="service in filteredStatusCarrierServices" :key="service.id" :value="String(service.id)">
                    {{ service.service_name }} / {{ service.service_code }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </label>
          </div>

          <div class="rounded-lg border bg-muted/35 p-3 text-xs text-muted-foreground">
            保存时系统会按“线路服务映射优先、承运商映射其次”解析 Provider Carrier Code。
            当前可预览：<span class="font-mono font-bold text-foreground">{{ resolvedProviderCarrierCodeLabel }}</span>
          </div>
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" @click="emit('update:open', false)">取消</Button>
          <Button type="submit" :disabled="submitting">
            <LoaderCircle v-if="submitting" class="size-4 animate-spin" />
            {{ submitting ? (fulfillmentMode ? '正在发货' : '正在保存') : (fulfillmentMode ? '确认发货' : '保存') }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { LoaderCircle } from '@lucide/vue'
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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import type {
  OrderStatusForm,
  OrderStatusOption,
  ShippingCarrier,
  ShippingCarrierService,
  TrackingProvider
} from './orderTypes'

withDefaults(defineProps<{
  open?: boolean
  statusForm: OrderStatusForm
  editableOrderStatusOptions?: OrderStatusOption[]
  editableShippingStatusOptions?: OrderStatusOption[]
  trackingProviders?: TrackingProvider[]
  carriers?: ShippingCarrier[]
  filteredStatusCarrierServices?: ShippingCarrierService[]
  resolvedProviderCarrierCodeLabel?: string
  fulfillmentMode?: boolean
  submitting?: boolean
}>(), {
  open: false,
  editableOrderStatusOptions: () => [],
  editableShippingStatusOptions: () => [],
  trackingProviders: () => [],
  carriers: () => [],
  filteredStatusCarrierServices: () => [],
  resolvedProviderCarrierCodeLabel: '未匹配映射',
  fulfillmentMode: false,
  submitting: false
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
  (event: 'submit'): void
  (event: 'carrier-change', value: string): void
  (event: 'carrier-service-change', value: string): void
}>()
</script>
