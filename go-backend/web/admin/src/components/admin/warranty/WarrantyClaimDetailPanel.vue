<template>
  <Card class="relative overflow-hidden py-0">
    <div v-if="detailLoading" class="absolute inset-0 z-10 flex items-center justify-center bg-card/80 backdrop-blur-[1px]">
      <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载申请详情" />
    </div>

    <CardHeader class="border-b bg-muted/30 px-4 py-3">
      <CardTitle class="flex items-center gap-2 text-sm">
        <FileWarning class="size-4 text-primary" />
        申请详情
      </CardTitle>
      <CardDescription>处理备注、订单行绑定和服务记录分别保存，不互相兜底。</CardDescription>
    </CardHeader>
    <CardContent class="space-y-4 p-4">
      <div v-if="!selectedClaim" class="flex min-h-72 flex-col items-center justify-center text-center text-muted-foreground">
        <FileWarning class="mb-2 size-8 opacity-55" />
        <p class="text-xs leading-6">从左侧列表选择一个保修申请，查看详情并填写处理记录。</p>
      </div>

      <template v-else>
        <section class="rounded-2xl border p-3">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="text-xs font-black uppercase tracking-wider">申请</h3>
            <AdminStatusBadge :tone="claimStatusTone(selectedClaim.status)">
              {{ claimStatusLabel(selectedClaim.status) }}
            </AdminStatusBadge>
          </div>
          <dl class="grid grid-cols-2 gap-2 text-xs">
            <DetailItem label="Claim ID">#{{ selectedClaim.id }}</DetailItem>
            <DetailItem label="Registration">{{ selectedClaim.registration_id || '未绑定' }}</DetailItem>
            <DetailItem label="Order" class="col-span-2">{{ selectedClaim.order_number || '-' }}</DetailItem>
            <DetailItem label="Email" class="col-span-2">{{ selectedClaim.email || selectedClaim.registration?.user?.email || '-' }}</DetailItem>
          </dl>
        </section>

        <section class="rounded-2xl border p-3">
          <div class="mb-2 flex items-center justify-between gap-2">
            <h3 class="text-xs font-black uppercase tracking-wider">订单行绑定</h3>
            <span class="text-[10px] font-mono text-muted-foreground/70">ORDER ITEM</span>
          </div>
          <div v-if="!selectedClaim.order_number" class="rounded-xl bg-muted/35 p-3 text-xs leading-5 text-muted-foreground">
            当前申请没有订单号，不能绑定订单行。这里保持真实为空，不用文本备注兜底。
          </div>
          <div v-else class="space-y-2">
            <Select
              :model-value="orderItemSelection"
              :disabled="!canEdit || orderItemsLoading || orderItemBinding"
              @update:model-value="$emit('update-order-item-selection', $event)"
            >
              <SelectTrigger class="h-9 w-full rounded-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">不绑定订单行</SelectItem>
                <SelectItem v-for="item in orderItems" :key="item.id" :value="String(item.id)">
                  {{ orderItemLabel(item) }}
                </SelectItem>
              </SelectContent>
            </Select>
            <div class="flex items-center justify-between gap-3">
              <span class="truncate text-[10px] text-muted-foreground/70">
                当前：{{ orderItemLabel(selectedClaim.order_item) }}
              </span>
              <Button
                size="sm"
                variant="outline"
                class="rounded-full font-black uppercase tracking-wider"
                :disabled="!canEdit || orderItemsLoading || orderItemBinding"
                @click="$emit('bind-order-item')"
              >
                <LoaderCircle v-if="orderItemBinding" class="size-3.5 animate-spin" />
                保存绑定
              </Button>
            </div>
          </div>
        </section>

        <section class="rounded-2xl border p-3">
          <h3 class="mb-2 text-xs font-black uppercase tracking-wider">问题描述</h3>
          <p class="whitespace-pre-wrap text-xs leading-5 text-foreground">{{ selectedClaim.description || '-' }}</p>
        </section>

        <section class="rounded-2xl border p-3">
          <div class="mb-2 flex items-center justify-between gap-2">
            <h3 class="text-xs font-black uppercase tracking-wider">处理备注</h3>
            <span class="text-[10px] font-mono text-muted-foreground/70">RESOLUTION</span>
          </div>
          <Textarea
            :model-value="resolutionDraft"
            class="min-h-36"
            placeholder="填写处理记录、判定原因、下一步动作。先保存为纯文本事实源，后续富文本/图片单独走规范。"
            :disabled="!canEdit || resolutionSaving"
            @update:model-value="$emit('update-resolution-draft', $event)"
          />
          <div class="mt-3 flex items-center justify-between gap-3">
            <span class="text-[10px] text-muted-foreground/70">最后处理：{{ formatDateTime(selectedClaim.processed_at) }}</span>
            <Button
              size="sm"
              class="rounded-full font-black uppercase tracking-wider"
              :disabled="!canEdit || resolutionSaving"
              @click="$emit('save-resolution')"
            >
              <LoaderCircle v-if="resolutionSaving" class="size-3.5 animate-spin" />
              保存备注
            </Button>
          </div>
        </section>

        <section class="rounded-2xl border p-3">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="text-xs font-black uppercase tracking-wider">服务记录</h3>
            <span class="text-[10px] font-mono text-muted-foreground/70">SERVICE LOG</span>
          </div>

          <div class="space-y-2">
            <div v-if="serviceRecordsLoading" class="flex items-center gap-2 rounded-xl bg-muted/35 p-3 text-xs text-muted-foreground">
              <LoaderCircle class="size-3.5 animate-spin" />
              正在读取服务记录
            </div>
            <div v-else-if="serviceRecords.length === 0" class="rounded-xl bg-muted/35 p-3 text-xs leading-5 text-muted-foreground">
              暂无服务记录。维修、更换、检测、退款等动作应写到这里，不再挤进处理备注。
            </div>
            <template v-else>
              <div v-for="record in serviceRecords" :key="record.id" class="rounded-xl border p-2">
                <div class="mb-1 flex items-center justify-between gap-2">
                  <span class="text-xs font-black">{{ serviceTypeLabel(record.service_type) }}</span>
                  <AdminStatusBadge :tone="serviceStatusTone(record.status)">
                    {{ serviceStatusLabel(record.status) }}
                  </AdminStatusBadge>
                </div>
                <p class="whitespace-pre-wrap text-xs leading-5 text-foreground">{{ record.summary }}</p>
                <p class="mt-1 text-[10px] text-muted-foreground/70">
                  {{ formatMoney(record.cost_amount, record.currency) }} · {{ formatDateTime(record.performed_at || record.created_at) }}
                </p>
              </div>
            </template>
          </div>

          <div class="mt-3 space-y-2 rounded-xl border border-dashed p-2">
            <div class="grid grid-cols-2 gap-2">
              <Select
                :model-value="serviceRecordForm.serviceType"
                :disabled="!canEdit || serviceRecordCreating"
                @update:model-value="updateServiceRecordField('serviceType', $event)"
              >
                <SelectTrigger class="h-9 w-full rounded-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="option in serviceTypeOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <Select
                :model-value="serviceRecordForm.status"
                :disabled="!canEdit || serviceRecordCreating"
                @update:model-value="updateServiceRecordField('status', $event)"
              >
                <SelectTrigger class="h-9 w-full rounded-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="option in serviceStatusOptions" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
            <Textarea
              :model-value="serviceRecordForm.summary"
              class="min-h-24"
              placeholder="新增服务记录，例如：完成检测，更换右侧轴承，等待客户确认。"
              :disabled="!canEdit || serviceRecordCreating"
              @update:model-value="updateServiceRecordField('summary', $event)"
            />
            <div class="grid grid-cols-[minmax(0,1fr)_90px_92px] gap-2">
              <Input
                :model-value="serviceRecordForm.performedAt"
                type="date"
                class="h-9 rounded-full"
                :disabled="!canEdit || serviceRecordCreating"
                @update:model-value="updateServiceRecordField('performedAt', $event)"
              />
              <Input
                :model-value="serviceRecordForm.costAmount"
                type="number"
                min="0"
                step="0.01"
                class="h-9 rounded-full"
                placeholder="费用"
                :disabled="!canEdit || serviceRecordCreating"
                @update:model-value="updateServiceRecordField('costAmount', $event)"
              />
              <Input
                :model-value="serviceRecordForm.currency"
                class="h-9 rounded-full font-mono uppercase"
                maxlength="8"
                placeholder="币种"
                :disabled="!canEdit || serviceRecordCreating"
                @update:model-value="updateServiceRecordField('currency', $event)"
              />
            </div>
            <div class="flex justify-end">
              <Button
                size="sm"
                class="rounded-full font-black uppercase tracking-wider"
                :disabled="!canEdit || serviceRecordCreating || !serviceRecordForm.summary.trim() || !serviceRecordForm.currency.trim()"
                @click="$emit('create-service-record')"
              >
                <LoaderCircle v-if="serviceRecordCreating" class="size-3.5 animate-spin" />
                添加记录
              </Button>
            </div>
          </div>
        </section>
      </template>
    </CardContent>
  </Card>
</template>

<script setup>
import { defineComponent, h } from 'vue'
import { FileWarning, LoaderCircle } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import {
  claimStatusLabel,
  claimStatusTone,
  formatDateTime,
  formatMoney,
  orderItemLabel,
  serviceStatusLabel,
  serviceStatusTone,
  serviceTypeLabel
} from '@/lib/warrantyPresentation'

const DetailItem = defineComponent({
  inheritAttrs: false,
  props: {
    label: { type: String, required: true }
  },
  setup(props, { slots, attrs }) {
    return () => h('div', { class: ['space-y-1 rounded-xl border p-2', attrs.class] }, [
      h('dt', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70' }, props.label),
      h('dd', { class: 'break-words font-bold text-foreground' }, slots.default ? slots.default() : '-')
    ])
  }
})

defineProps({
  detailLoading: { type: Boolean, default: false },
  resolutionSaving: { type: Boolean, default: false },
  orderItemsLoading: { type: Boolean, default: false },
  orderItemBinding: { type: Boolean, default: false },
  serviceRecordsLoading: { type: Boolean, default: false },
  serviceRecordCreating: { type: Boolean, default: false },
  selectedClaim: { type: Object, default: null },
  resolutionDraft: { type: String, default: '' },
  orderItems: { type: Array, default: () => [] },
  orderItemSelection: { type: String, default: 'none' },
  serviceRecords: { type: Array, default: () => [] },
  serviceRecordForm: { type: Object, required: true },
  serviceTypeOptions: { type: Array, required: true },
  serviceStatusOptions: { type: Array, required: true },
  canEdit: { type: Boolean, default: false }
})

const emit = defineEmits([
  'update-order-item-selection',
  'bind-order-item',
  'update-resolution-draft',
  'save-resolution',
  'update-service-record-form',
  'create-service-record'
])

const updateServiceRecordField = (field, value) => {
  emit('update-service-record-form', { [field]: value })
}
</script>
