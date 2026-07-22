<template>
  <div class="space-y-3">
    <div class="flex flex-wrap items-end justify-between gap-3">
      <div>
        <h2 class="text-sm font-black tracking-tighter italic uppercase">追踪任务</h2>
        <p class="mt-1 text-xs text-muted-foreground">这里展示订单发货后生成的运单同步状态；后续定时轮询和 17TRACK webhook 都以这张任务表为落点。</p>
      </div>
      <div class="flex flex-wrap gap-2">
        <Button variant="outline" size="sm" :disabled="loading.trackingShipments" @click="fetchTrackingShipments">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading.trackingShipments }]" />
          刷新任务
        </Button>
        <Button v-if="canEdit" size="sm" :disabled="syncingDueTrackingShipments" @click="syncDueTrackingShipments">
          <RefreshCw :class="['size-3.5', { 'animate-spin': syncingDueTrackingShipments }]" />
          同步到期任务
        </Button>
      </div>
    </div>

    <section class="grid gap-3 md:grid-cols-3 xl:grid-cols-6">
      <div v-for="item in trackingShipmentStatusCards" :key="item.key" class="rounded-lg border bg-card p-3">
        <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">{{ item.label }}</p>
        <p class="mt-1 text-2xl font-black tabular-nums">{{ item.value }}</p>
      </div>
    </section>

    <section class="rounded-lg border bg-card p-3 shadow-xs">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div class="flex flex-wrap items-center gap-2">
            <h3 class="text-sm font-black tracking-tighter italic uppercase">自动轮询器</h3>
            <AdminStatusBadge :tone="trackingPollingState.enabled ? (trackingPollingState.running ? 'blue' : 'green') : 'gray'">
              {{ trackingPollingState.enabled ? (trackingPollingState.running ? '运行中' : '已启用') : '已禁用' }}
            </AdminStatusBadge>
          </div>
          <p class="mt-1 text-xs text-muted-foreground">
            后端 scheduler 的当前状态。这里不取数据库历史，只显示当前服务进程内最近一次运行结果，用来判断自动同步链路是否活着。
          </p>
        </div>
        <Button variant="outline" size="sm" :disabled="loading.trackingPolling" @click="fetchTrackingPollingState">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading.trackingPolling }]" />
          刷新轮询状态
        </Button>
      </div>

      <div class="mt-3 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <div class="rounded-md border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">POLICY / 策略</p>
          <p class="mt-1 text-xs font-bold">间隔 {{ trackingPollingIntervalLabel }}</p>
          <p class="mt-0.5 text-[10px] text-muted-foreground">每批 {{ trackingPollingState.batch_limit || 0 }} 条</p>
        </div>
        <div class="rounded-md border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">LAST RUN / 最近运行</p>
          <p class="mt-1 font-mono text-xs">{{ formatDate(trackingPollingState.last_started_at) }}</p>
          <p class="mt-0.5 text-[10px] text-muted-foreground">结束 {{ formatDate(trackingPollingState.last_finished_at) }}</p>
        </div>
        <div class="rounded-md border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">RESULT / 结果</p>
          <p class="mt-1 text-xs font-bold">
            命中 {{ trackingPollingState.last_matched || 0 }} / 成功 {{ trackingPollingState.last_synced || 0 }} / 失败 {{ trackingPollingState.last_failed || 0 }}
          </p>
          <p class="mt-0.5 text-[10px] text-muted-foreground">耗时 {{ trackingPollingDurationLabel }}</p>
        </div>
        <div class="rounded-md border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">ERROR / 最近错误</p>
          <p class="mt-1 line-clamp-2 text-xs" :class="trackingPollingState.last_error ? 'text-destructive' : 'text-muted-foreground'">
            {{ trackingPollingState.last_error || trackingPollingLastItemError || '无错误' }}
          </p>
        </div>
      </div>
    </section>

    <section class="rounded-lg border bg-card p-3 shadow-xs">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <div class="flex flex-wrap items-center gap-2">
            <h3 class="text-sm font-black tracking-tighter italic uppercase">Webhook 推送</h3>
            <AdminStatusBadge :tone="trackingWebhookState.last_accepted ? 'green' : (trackingWebhookState.last_error ? 'coral' : 'gray')">
              {{ trackingWebhookState.last_accepted ? '最近成功' : (trackingWebhookState.last_error ? '最近失败' : '暂无回调') }}
            </AdminStatusBadge>
          </div>
          <p class="mt-1 text-xs text-muted-foreground">
            展示当前服务进程最近一次 17TRACK 或内部 webhook 处理结果。用于排查 URL、验签密钥、Provider 开关和运单匹配问题。
          </p>
        </div>
        <Button variant="outline" size="sm" :disabled="loading.trackingWebhook" @click="fetchTrackingWebhookState">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading.trackingWebhook }]" />
          刷新 Webhook 状态
        </Button>
      </div>

      <div class="mt-3 grid gap-3 md:grid-cols-2 xl:grid-cols-4">
        <div class="rounded-md border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">RECEIVED / 最近接收</p>
          <p class="mt-1 font-mono text-xs">{{ formatDate(trackingWebhookState.last_received_at) }}</p>
          <p class="mt-0.5 text-[10px] text-muted-foreground">HTTP {{ trackingWebhookState.last_http_status || '-' }} / {{ trackingWebhookDurationLabel }}</p>
        </div>
        <div class="rounded-md border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SIGNATURE / 验签</p>
          <p class="mt-1 text-xs font-bold">{{ trackingWebhookSignatureLabel }}</p>
          <p class="mt-0.5 text-[10px] text-muted-foreground">Provider {{ trackingWebhookState.last_provider_code || '-' }}</p>
        </div>
        <div class="rounded-md border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SHIPMENT / 命中任务</p>
          <p class="mt-1 font-mono text-xs">{{ trackingWebhookState.last_tracking_number || '-' }}</p>
          <p class="mt-0.5 text-[10px] text-muted-foreground">
            订单 #{{ trackingWebhookState.last_order_id || '-' }} / {{ trackingWebhookState.last_carrier_code || '-' }} / {{ trackingWebhookState.last_event_count || 0 }} 事件
          </p>
        </div>
        <div class="rounded-md border bg-muted/25 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">ERROR / 最近错误</p>
          <p class="mt-1 line-clamp-2 text-xs" :class="trackingWebhookState.last_error ? 'text-destructive' : 'text-muted-foreground'">
            {{ trackingWebhookState.last_error || '无错误' }}
          </p>
        </div>
      </div>
    </section>

    <section class="rounded-lg border bg-card p-3 shadow-xs">
      <div class="grid gap-3 lg:grid-cols-6">
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SYNC / 同步状态</span>
          <Select v-model="filters.sync_status">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部同步状态</SelectItem>
              <SelectItem value="pending">待同步</SelectItem>
              <SelectItem value="syncing">同步中</SelectItem>
              <SelectItem value="synced">已同步</SelectItem>
              <SelectItem value="failed">同步失败</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">REGISTER / 登记状态</span>
          <Select v-model="filters.registration_status">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部登记状态</SelectItem>
              <SelectItem value="pending">待登记</SelectItem>
              <SelectItem value="registered">已登记</SelectItem>
              <SelectItem value="failed">登记失败</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">PROVIDER / 服务商</span>
          <Select v-model="filters.provider_id">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部 Provider</SelectItem>
              <SelectItem v-for="provider in trackingProvidersList" :key="provider.id" :value="String(provider.id)">
                {{ provider.provider_name }} / {{ provider.provider_code }}
              </SelectItem>
            </SelectContent>
          </Select>
        </label>

        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">CARRIER / 承运商</span>
          <Select v-model="filters.carrier_id" @update:model-value="handleCarrierFilterChange">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部承运商</SelectItem>
              <SelectItem v-for="carrier in carriersList" :key="carrier.id" :value="String(carrier.id)">
                {{ carrier.name }} / {{ carrier.code }}
              </SelectItem>
            </SelectContent>
          </Select>
        </label>

        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SERVICE / 线路</span>
          <Select v-model="filters.carrier_service_id">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部线路服务</SelectItem>
              <SelectItem v-for="service in filteredCarrierServices" :key="service.id" :value="String(service.id)">
                {{ service.service_name }} / {{ service.service_code }}
              </SelectItem>
            </SelectContent>
          </Select>
        </label>

        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SCOPE / 轮询范围</span>
          <Select v-model="filters.due_only">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部任务</SelectItem>
              <SelectItem value="true">仅到期任务</SelectItem>
            </SelectContent>
          </Select>
        </label>
      </div>

      <div class="mt-3 grid gap-3 lg:grid-cols-[1fr_12rem_12rem_auto]">
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">KEYWORD / 订单、单号、承运商代码、错误</span>
          <Input
            v-model="filters.keyword"
            class="h-9"
            placeholder="例如：#1001 / 1Z999 / 190271 / api key"
            @keyup.enter="fetchTrackingShipments"
          />
        </label>

        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">ENABLED / 启用</span>
          <Select v-model="filters.enabled">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部</SelectItem>
              <SelectItem value="true">只看启用</SelectItem>
              <SelectItem value="false">只看停用</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">LIMIT / 数量</span>
          <Select v-model="filters.limit">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="50">最近 50 条</SelectItem>
              <SelectItem value="100">最近 100 条</SelectItem>
              <SelectItem value="200">最近 200 条</SelectItem>
              <SelectItem value="500">最近 500 条</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <div class="flex items-end gap-2">
          <Button variant="outline" size="sm" class="h-9" @click="resetFilters">
            重置
          </Button>
          <Button size="sm" class="h-9" :disabled="loading.trackingShipments" @click="fetchTrackingShipments">
            应用筛选
          </Button>
        </div>
      </div>
    </section>

    <AdminTablePanel :loading="loading.trackingShipments">
      <Table class="min-w-[1360px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-24">订单</TableHead>
            <TableHead>运单</TableHead>
            <TableHead>Provider</TableHead>
            <TableHead>本地承运商 / 线路</TableHead>
            <TableHead class="w-28">同步状态</TableHead>
            <TableHead class="w-24 text-right">事件</TableHead>
            <TableHead class="w-40">最后同步</TableHead>
            <TableHead class="w-40">下次同步</TableHead>
            <TableHead>错误</TableHead>
            <TableHead class="w-52 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="trackingShipments.length === 0" :colspan="10">
            <div class="flex flex-col items-center text-muted-foreground">
              <Radar class="mb-2 size-7 opacity-55" />
              <span class="text-xs">暂无追踪任务；订单填写物流单号后会自动生成。</span>
            </div>
          </TableEmpty>
          <TableRow v-for="shipment in trackingShipments" :key="shipment.id">
            <TableCell class="font-mono text-xs font-bold">#{{ shipment.order_id }}</TableCell>
            <TableCell>
              <span class="block font-mono text-xs font-bold">{{ shipment.tracking_number || '-' }}</span>
              <span class="block font-mono text-[10px] text-muted-foreground/70">{{ shipment.provider_carrier_code || '-' }}</span>
            </TableCell>
            <TableCell>
              <span class="block font-bold text-xs">{{ trackingShipmentProviderName(shipment) }}</span>
              <AdminStatusBadge :tone="trackingShipmentRegistrationTone(shipment.registration_status)" class="mt-1">
                {{ trackingShipmentRegistrationLabel(shipment.registration_status) }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell>
              <span class="block text-xs font-bold">{{ trackingShipmentCarrierLabel(shipment) }}</span>
              <span class="block text-[10px] text-muted-foreground/70">{{ trackingShipmentServiceLabel(shipment) }}</span>
            </TableCell>
            <TableCell>
              <AdminStatusBadge :tone="trackingShipmentStatusTone(shipment.sync_status)">
                {{ trackingShipmentStatusLabel(shipment.sync_status) }}
              </AdminStatusBadge>
            </TableCell>
            <TableCell class="text-right font-mono text-xs tabular-nums">{{ shipment.event_count || 0 }}</TableCell>
            <TableCell class="font-mono text-[10px] text-muted-foreground">{{ formatDate(shipment.last_synced_at) }}</TableCell>
            <TableCell class="font-mono text-[10px] text-muted-foreground">{{ formatDate(shipment.next_sync_at) }}</TableCell>
            <TableCell class="max-w-80 truncate text-xs text-destructive">{{ shipment.last_error || '-' }}</TableCell>
            <TableCell class="text-right">
              <div class="inline-flex items-center gap-1">
                <Button variant="outline" size="sm" :disabled="loading.trackingEvents" @click="openTrackingEvents(shipment)">
                  事件
                </Button>
                <Button
                  v-if="canEdit"
                  variant="outline"
                  size="sm"
                  :disabled="isRegisteringTrackingShipment(shipment)"
                  @click="registerTrackingShipment(shipment)"
                >
                  <RefreshCw :class="['size-3.5', { 'animate-spin': isRegisteringTrackingShipment(shipment) }]" />
                  登记
                </Button>
                <Button
                  v-if="canEdit"
                  variant="outline"
                  size="sm"
                  :disabled="isSyncingTrackingShipment(shipment)"
                  @click="syncTrackingShipment(shipment)"
                >
                  <RefreshCw :class="['size-3.5', { 'animate-spin': isSyncingTrackingShipment(shipment) }]" />
                  同步
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>

    <Dialog v-model:open="eventDialogOpen">
      <DialogContent size="xl" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
        <DialogHeader>
          <DialogTitle>追踪事件时间线</DialogTitle>
          <DialogDescription>
            订单 #{{ selectedTrackingShipment?.order_id || '-' }} · 运单 {{ selectedTrackingShipment?.tracking_number || '-' }}
          </DialogDescription>
        </DialogHeader>

        <section v-if="selectedTrackingShipment" class="grid gap-3 md:grid-cols-4">
          <div class="rounded-lg border bg-muted/25 p-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">PROVIDER</p>
            <p class="mt-1 text-xs font-bold">{{ trackingShipmentProviderName(selectedTrackingShipment) }}</p>
            <p class="mt-0.5 font-mono text-[10px] text-muted-foreground">{{ selectedTrackingShipment.provider_carrier_code || '-' }}</p>
          </div>
          <div class="rounded-lg border bg-muted/25 p-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">CARRIER</p>
            <p class="mt-1 text-xs font-bold">{{ trackingShipmentCarrierLabel(selectedTrackingShipment) }}</p>
            <p class="mt-0.5 text-[10px] text-muted-foreground">{{ trackingShipmentServiceLabel(selectedTrackingShipment) }}</p>
          </div>
          <div class="rounded-lg border bg-muted/25 p-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SYNC</p>
            <AdminStatusBadge :tone="trackingShipmentStatusTone(selectedTrackingShipment.sync_status)" class="mt-1">
              {{ trackingShipmentStatusLabel(selectedTrackingShipment.sync_status) }}
            </AdminStatusBadge>
            <p class="mt-2 font-mono text-[10px] text-muted-foreground">{{ formatDate(selectedTrackingShipment.last_synced_at) }}</p>
          </div>
          <div class="rounded-lg border bg-muted/25 p-3">
            <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">EVENTS</p>
            <p class="mt-1 text-2xl font-black tabular-nums">{{ trackingEvents.length }}</p>
            <p class="mt-0.5 text-[10px] text-muted-foreground">任务记录 {{ selectedTrackingShipment.event_count || 0 }} 条</p>
          </div>
        </section>

        <div v-if="loading.trackingEvents" class="flex items-center justify-center gap-2 rounded-lg border border-dashed p-8 text-sm text-muted-foreground">
          <RefreshCw class="size-4 animate-spin" />
          正在读取轨迹事件...
        </div>

        <div v-else-if="trackingEvents.length === 0" class="flex flex-col items-center rounded-lg border border-dashed p-8 text-muted-foreground">
          <Radar class="mb-2 size-8 opacity-55" />
          <p class="text-sm font-bold">暂无轨迹事件</p>
          <p class="mt-1 text-xs">同步或 Webhook 成功后会写入 tracking_events，并在这里展示。</p>
        </div>

        <ol v-else class="relative ml-2 space-y-4 border-l pl-5">
          <li v-for="event in trackingEvents" :key="event.id || `${event.tracking_number}-${event.event_time}-${event.description}`" class="relative">
            <span class="absolute -left-[1.75rem] top-1 flex size-3 rounded-full border-2 border-background bg-primary shadow" />
            <div class="rounded-lg border bg-card p-4 shadow-xs">
              <div class="flex flex-wrap items-start justify-between gap-3">
                <div>
                  <div class="flex flex-wrap items-center gap-2">
                    <AdminStatusBadge :tone="trackingEventStatusTone(event.status)">
                      {{ event.status || 'UNKNOWN' }}
                    </AdminStatusBadge>
                    <span class="font-mono text-xs text-muted-foreground">{{ event.provider_carrier_code || selectedTrackingShipment?.provider_carrier_code || '-' }}</span>
                  </div>
                  <p class="mt-2 text-sm font-bold">{{ event.description || '无事件描述' }}</p>
                  <p class="mt-1 text-xs text-muted-foreground">{{ event.location || '无地点信息' }}</p>
                </div>
                <div class="text-right">
                  <p class="font-mono text-xs font-bold">{{ formatDate(event.event_time) }}</p>
                  <p class="mt-1 font-mono text-[10px] text-muted-foreground">created {{ formatDate(event.created_at) }}</p>
                </div>
              </div>
            </div>
          </li>
        </ol>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { Radar, RefreshCw } from '@lucide/vue'
import shippingApi from '@/api/shipping'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'

const props = defineProps({
  trackingProviders: {
    type: Array,
    default: () => [],
  },
  carriers: {
    type: Array,
    default: () => [],
  },
  carrierServices: {
    type: Array,
    default: () => [],
  },
  canEdit: {
    type: Boolean,
    default: false,
  },
})

const emit = defineEmits(['count-change'])

const trackingShipments = ref([])
const trackingPollingState = ref({})
const trackingWebhookState = ref({})
const trackingEvents = ref([])
const selectedTrackingShipment = ref(null)
const eventDialogOpen = ref(false)
const syncingDueTrackingShipments = ref(false)
const registeringTrackingShipmentIds = ref(new Set())
const syncingTrackingShipmentIds = ref(new Set())
const filters = reactive(defaultFilters())
const loading = reactive({
  trackingShipments: false,
  trackingPolling: false,
  trackingWebhook: false,
  trackingEvents: false,
})

const trackingProvidersList = computed(() => Array.isArray(props.trackingProviders) ? props.trackingProviders : [])
const carriersList = computed(() => Array.isArray(props.carriers) ? props.carriers : [])
const carrierServicesList = computed(() => Array.isArray(props.carrierServices) ? props.carrierServices : [])

function defaultFilters() {
  return {
    sync_status: 'all',
    registration_status: 'all',
    provider_id: 'all',
    carrier_id: 'all',
    carrier_service_id: 'all',
    enabled: 'all',
    due_only: 'all',
    keyword: '',
    limit: '200',
  }
}

function resetReactive(target, defaults) {
  Object.keys(target).forEach((key) => delete target[key])
  Object.assign(target, defaults)
}

function isTrackingShipmentDue(shipment) {
  const status = shipment?.sync_status || 'pending'
  if (status === 'pending') return true

  const nextSyncAt = shipment?.next_sync_at ? Date.parse(shipment.next_sync_at) : Number.NaN
  if (!Number.isFinite(nextSyncAt)) {
    return status === 'failed'
  }

  return nextSyncAt <= Date.now()
}

const trackingShipmentStatusCards = computed(() => {
  const counts = trackingShipments.value.reduce((acc, shipment) => {
    const status = shipment.sync_status || 'missing'
    acc[status] = (acc[status] || 0) + 1
    if (shipment.registration_status === 'failed') {
      acc.registrationFailed = (acc.registrationFailed || 0) + 1
    }
    if (isTrackingShipmentDue(shipment)) {
      acc.due = (acc.due || 0) + 1
    }
    return acc
  }, {})

  return [
    { key: 'total', label: '当前结果', value: trackingShipments.value.length },
    { key: 'pending', label: '待同步', value: counts.pending || 0 },
    { key: 'synced', label: '已同步', value: counts.synced || 0 },
    { key: 'failed', label: '同步失败', value: counts.failed || 0 },
    { key: 'registrationFailed', label: '登记失败', value: counts.registrationFailed || 0 },
    { key: 'due', label: '到期轮询', value: counts.due || 0 },
  ]
})

const filteredCarrierServices = computed(() => {
  const carrierID = Number(filters.carrier_id || 0)
  if (!carrierID) return carrierServicesList.value
  return carrierServicesList.value.filter((service) => Number(service.carrier_id || 0) === carrierID)
})

const trackingPollingIntervalLabel = computed(() => {
  const seconds = Number(trackingPollingState.value.interval_seconds || 0)
  if (seconds <= 0) return trackingPollingState.value.interval || '未配置'
  if (seconds % 3600 === 0) return `${seconds / 3600} 小时`
  if (seconds % 60 === 0) return `${seconds / 60} 分钟`
  return `${seconds} 秒`
})

const trackingPollingDurationLabel = computed(() => {
  const ms = Number(trackingPollingState.value.last_duration_ms || 0)
  if (ms <= 0) return '-'
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)} 秒`
  return `${ms} ms`
})

const trackingPollingLastItemError = computed(() => {
  const errors = Array.isArray(trackingPollingState.value.last_errors) ? trackingPollingState.value.last_errors : []
  const first = errors.find((item) => item?.error)
  if (!first) return ''
  const order = first.order_id ? `订单 #${first.order_id}` : first.tracking_number || '任务'
  return `${order}: ${first.error}`
})

const trackingWebhookDurationLabel = computed(() => {
  const ms = Number(trackingWebhookState.value.last_duration_ms || 0)
  if (ms <= 0) return '-'
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)} 秒`
  return `${ms} ms`
})

const trackingWebhookSignatureLabel = computed(() => {
  if (!trackingWebhookState.value.last_signature_checked) return '未验签'
  return trackingWebhookState.value.last_signature_valid ? '验签通过' : '验签失败'
})

const trackingShipmentFilterParam = (value) => {
  const normalized = String(value ?? '').trim()
  return normalized && normalized !== 'all' ? normalized : undefined
}

const buildTrackingShipmentQuery = () => {
  const params = {
    sync_status: trackingShipmentFilterParam(filters.sync_status),
    registration_status: trackingShipmentFilterParam(filters.registration_status),
    provider_id: trackingShipmentFilterParam(filters.provider_id),
    carrier_id: trackingShipmentFilterParam(filters.carrier_id),
    carrier_service_id: trackingShipmentFilterParam(filters.carrier_service_id),
    enabled: trackingShipmentFilterParam(filters.enabled),
    due_only: filters.due_only === 'true' ? 'true' : undefined,
    keyword: filters.keyword?.trim() || undefined,
    limit: Number(filters.limit || 200),
  }

  return Object.fromEntries(Object.entries(params).filter(([, value]) => value !== undefined && value !== ''))
}

const fetchTrackingShipments = async () => {
  loading.trackingShipments = true
  try {
    trackingShipments.value = await shippingApi.listTrackingShipments(buildTrackingShipmentQuery())
    emit('count-change', trackingShipments.value.length)
  } catch (error) {
    console.error('Failed to fetch tracking shipments:', error)
  } finally {
    loading.trackingShipments = false
  }
}

const fetchTrackingPollingState = async () => {
  loading.trackingPolling = true
  try {
    trackingPollingState.value = await shippingApi.getTrackingPollingState()
  } catch (error) {
    console.error('Failed to fetch tracking polling state:', error)
  } finally {
    loading.trackingPolling = false
  }
}

const fetchTrackingWebhookState = async () => {
  loading.trackingWebhook = true
  try {
    trackingWebhookState.value = await shippingApi.getTrackingWebhookState()
  } catch (error) {
    console.error('Failed to fetch tracking webhook state:', error)
  } finally {
    loading.trackingWebhook = false
  }
}

const fetchTrackingEvents = async (shipment) => {
  const orderId = Number(shipment?.order_id || 0)
  if (!orderId) {
    trackingEvents.value = []
    return
  }

  loading.trackingEvents = true
  try {
    trackingEvents.value = await shippingApi.listTrackingEvents(orderId)
  } catch (error) {
    console.error('Failed to fetch tracking events:', error)
    trackingEvents.value = []
    toast.error(error.response?.data?.error || '读取轨迹事件失败')
  } finally {
    loading.trackingEvents = false
  }
}

const openTrackingEvents = async (shipment) => {
  selectedTrackingShipment.value = shipment
  trackingEvents.value = []
  eventDialogOpen.value = true
  await fetchTrackingEvents(shipment)
}

const refresh = async () => {
  await Promise.all([fetchTrackingShipments(), fetchTrackingPollingState(), fetchTrackingWebhookState()])
}

const resetFilters = async () => {
  resetReactive(filters, defaultFilters())
  await fetchTrackingShipments()
}

const handleCarrierFilterChange = () => {
  if (filters.carrier_service_id === 'all') return

  const selectedService = carrierServicesList.value.find((service) => String(service.id) === String(filters.carrier_service_id))
  if (!selectedService) {
    filters.carrier_service_id = 'all'
    return
  }

  const carrierID = Number(filters.carrier_id || 0)
  if (carrierID > 0 && Number(selectedService.carrier_id || 0) !== carrierID) {
    filters.carrier_service_id = 'all'
  }
}

const syncDueTrackingShipments = async () => {
  syncingDueTrackingShipments.value = true
  try {
    const result = await shippingApi.syncDueTrackingShipments({ limit: 20 })
    toast.success(`追踪任务同步完成：成功 ${result.synced || 0}，失败 ${result.failed || 0}`)
    await Promise.all([fetchTrackingShipments(), fetchTrackingPollingState()])
  } catch (error) {
    console.error('Failed to sync due tracking shipments:', error)
    toast.error(error.response?.data?.error || '追踪任务同步失败')
  } finally {
    syncingDueTrackingShipments.value = false
  }
}

const isRegisteringTrackingShipment = (shipment) => registeringTrackingShipmentIds.value.has(Number(shipment.order_id))
const isSyncingTrackingShipment = (shipment) => syncingTrackingShipmentIds.value.has(Number(shipment.order_id))

const registerTrackingShipment = async (shipment) => {
  const orderId = Number(shipment?.order_id || 0)
  if (!orderId || isRegisteringTrackingShipment(shipment)) return

  registeringTrackingShipmentIds.value = new Set(registeringTrackingShipmentIds.value).add(orderId)
  try {
    await shippingApi.registerTrackingShipment(orderId)
    toast.success(`订单 #${orderId} 运单已登记到 Provider`)
    await fetchTrackingShipments()
  } catch (error) {
    console.error('Failed to register tracking shipment:', error)
    toast.error(error.response?.data?.error || '追踪任务登记失败')
    await fetchTrackingShipments()
  } finally {
    const next = new Set(registeringTrackingShipmentIds.value)
    next.delete(orderId)
    registeringTrackingShipmentIds.value = next
  }
}

const syncTrackingShipment = async (shipment) => {
  const orderId = Number(shipment?.order_id || 0)
  if (!orderId || isSyncingTrackingShipment(shipment)) return

  syncingTrackingShipmentIds.value = new Set(syncingTrackingShipmentIds.value).add(orderId)
  try {
    const result = await shippingApi.syncTrackingShipment(orderId)
    const eventCount = result.tracking?.events?.length ?? result.tracking?.shipment?.event_count ?? 0
    toast.success(`订单 #${orderId} 轨迹已同步：${eventCount} 条事件`)
    await fetchTrackingShipments()
  } catch (error) {
    console.error('Failed to sync tracking shipment:', error)
    toast.error(error.response?.data?.error || '单条追踪任务同步失败')
    await fetchTrackingShipments()
  } finally {
    const next = new Set(syncingTrackingShipmentIds.value)
    next.delete(orderId)
    syncingTrackingShipmentIds.value = next
  }
}

const trackingShipmentStatusLabel = (status) => {
  const labels = {
    pending: '待同步',
    syncing: '同步中',
    synced: '已同步',
    failed: '同步失败',
  }
  return labels[status] || status || '未建立'
}

const trackingShipmentStatusTone = (status) => {
  const tones = {
    pending: 'gray',
    syncing: 'blue',
    synced: 'green',
    failed: 'coral',
  }
  return tones[status] || 'gray'
}

const trackingShipmentRegistrationLabel = (status) => {
  const labels = {
    pending: '待登记',
    registered: '已登记',
    failed: '登记失败',
  }
  return labels[status] || status || '未登记'
}

const trackingShipmentRegistrationTone = (status) => {
  const tones = {
    pending: 'gray',
    registered: 'green',
    failed: 'coral',
  }
  return tones[status] || 'gray'
}

const trackingEventStatusTone = (status) => {
  const normalized = String(status || '').toLowerCase()
  if (normalized.includes('delivered') || normalized.includes('签收') || normalized.includes('妥投')) return 'green'
  if (normalized.includes('exception') || normalized.includes('fail') || normalized.includes('returned') || normalized.includes('异常') || normalized.includes('退回')) return 'coral'
  if (normalized.includes('transit') || normalized.includes('pickup') || normalized.includes('arriv') || normalized.includes('运输') || normalized.includes('揽收')) return 'blue'
  return 'gray'
}

const trackingShipmentProviderName = (shipment) => {
  if (shipment.provider?.provider_name) return shipment.provider.provider_name
  const provider = trackingProvidersList.value.find((item) => Number(item.id) === Number(shipment.tracking_provider_id))
  return provider?.provider_name || `Provider #${shipment.tracking_provider_id || '-'}`
}

const trackingShipmentCarrierLabel = (shipment) => {
  if (shipment.carrier?.name) return shipment.carrier.name
  const carrier = carriersList.value.find((item) => Number(item.id) === Number(shipment.carrier_id))
  return carrier?.name || '未绑定承运商'
}

const trackingShipmentServiceLabel = (shipment) => {
  if (shipment.carrier_service?.service_name) return shipment.carrier_service.service_name
  const service = carrierServicesList.value.find((item) => Number(item.id) === Number(shipment.carrier_service_id))
  return service?.service_name || '未绑定线路'
}

const formatDate = (value) => value ? new Date(value).toLocaleString('zh-CN') : '-'

defineExpose({ refresh, refreshShipments: fetchTrackingShipments })

onMounted(() => refresh())
</script>
