<script setup lang="ts">
import { computed, ref } from 'vue'
import { Activity, AlertTriangle, Calculator, CheckCircle2, ClipboardCheck, Download, Globe2, PackageCheck, Plus, Power, Printer, RadioTower, RefreshCw, Search, Send, Settings2, Trash2, Truck } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const waybillSearch = ref('')
const waybillChannel = ref('')
const waybillWarehouse = ref('')
const waybillStatus = ref('')
const waybillRefreshing = ref(false)
const waybillSelectedIds = ref<string[]>([])
const waybillActionMessage = ref('')
const canShip = computed(() => authStore.hasPermission('logistics:yanwen:ship'))
const canCancel = computed(() => authStore.hasPermission('logistics:yanwen:cancel'))
const waybillRecords = ref<Array<{
  id: string
  orderNo: string
  createdAt: string
  waybill: string
  lastMile: string
  channel: string
  warehouse: string
  destination: string
  consignee: string
  description: string
  weight: string
  status: string
}>>([])
const filteredWaybills = computed(() => {
  const query = waybillSearch.value.trim().toLowerCase()
  return waybillRecords.value.filter((record) => {
    const matchesSearch = !query || [record.orderNo, record.waybill, record.lastMile, record.consignee, record.description]
      .some((value) => value.toLowerCase().includes(query))
    const matchesChannel = !waybillChannel.value || record.channel === waybillChannel.value
    const matchesWarehouse = !waybillWarehouse.value || record.warehouse === waybillWarehouse.value
    const matchesStatus = !waybillStatus.value || record.status === waybillStatus.value
    return matchesSearch && matchesChannel && matchesWarehouse && matchesStatus
  })
})
const hasSelectedWaybills = computed(() => waybillSelectedIds.value.length > 0)
const refreshWaybills = async () => {
  waybillRefreshing.value = true
  waybillActionMessage.value = ''
  await Promise.resolve()
  waybillRefreshing.value = false
}
const resetWaybillFilters = () => {
  waybillSearch.value = ''
  waybillChannel.value = ''
  waybillWarehouse.value = ''
  waybillStatus.value = ''
}
const toggleWaybill = (id: string) => {
  waybillSelectedIds.value = waybillSelectedIds.value.includes(id)
    ? waybillSelectedIds.value.filter((selectedId) => selectedId !== id)
    : [...waybillSelectedIds.value, id]
}
const submitWaybills = () => {
  if (!canShip.value || !hasSelectedWaybills.value) return
  waybillActionMessage.value = 'express.order.create 尚未接入，未提交真实燕文运单。'
}
const printWaybillLabels = () => {
  if (!canShip.value || !hasSelectedWaybills.value) return
  waybillActionMessage.value = 'express.order.label.get 尚未接入，未生成真实面单。'
}
const cancelWaybill = (id: string) => {
  if (!canCancel.value) return
  waybillActionMessage.value = `express.order.cancel 尚未接入，未取消运单 ${id}。`
}

const canManage = computed(() => authStore.hasPermission('logistics:yanwen:manage'))
</script>

<template>
      <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-emerald-600">Waybill Fulfillment &amp; Labels</p>
            <h2 class="mt-1 text-lg font-black tracking-tight">专线运单与面单中心</h2>
            <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">处理商城待发货订单、燕文运单、10×10 热敏面单和交运前拦截。真实操作会调用 express.order.create、express.order.label.get 与 express.order.cancel。</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <Button variant="outline" size="sm" :disabled="waybillRefreshing" @click="refreshWaybills"><RefreshCw :class="['size-3.5', { 'animate-spin': waybillRefreshing }]" />刷新</Button>
            <Button size="sm" :disabled="!canShip || !hasSelectedWaybills" :title="canShip ? undefined : '需要 logistics:yanwen:ship 权限'" @click="submitWaybills"><Send class="size-3.5" />批量推单</Button>
            <Button variant="outline" size="sm" :disabled="!canShip || !hasSelectedWaybills" :title="canShip ? undefined : '需要 logistics:yanwen:ship 权限'" @click="printWaybillLabels"><Printer class="size-3.5" />批量打面单</Button>
          </div>
        </div>

        <div class="grid gap-3 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4 md:grid-cols-2 xl:grid-cols-[minmax(0,1.4fr)_repeat(3,minmax(0,1fr))_auto]">
          <label class="space-y-1.5"><span class="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-wider text-muted-foreground/80"><Search class="size-3" />搜索运单台账</span><Input v-model="waybillSearch" placeholder="商城单号 / 燕文单号 / 尾程单号 / 买家" /></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">渠道</span><select v-model="waybillChannel" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm"><option value="">全部渠道</option><option value="燕文专线挂号">燕文专线挂号</option><option value="燕文特快专线">燕文特快专线</option><option value="燕文经济小包">燕文经济小包</option></select></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">交货仓</span><Input v-model="waybillWarehouse" placeholder="仓库代码 / 名称（待主数据同步）" /></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">状态</span><select v-model="waybillStatus" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm"><option value="">全部状态</option><option value="待推单">待推单</option><option value="已建单">已建单</option><option value="已打面单">已打面单</option><option value="揽收">揽收</option><option value="运输中">运输中</option></select></label>
          <div class="flex items-end"><Button variant="ghost" size="sm" class="w-full" @click="resetWaybillFilters">重置筛选</Button></div>
        </div>

        <p v-if="waybillActionMessage" class="rounded-2xl border border-amber-500/30 bg-amber-500/10 px-4 py-3 text-xs leading-5 text-amber-700 dark:text-amber-300">{{ waybillActionMessage }}</p>

        <section class="overflow-hidden rounded-[24px] border border-dashed border-border/80">
          <div class="flex items-center justify-between gap-3 border-b border-dashed border-border/80 px-5 py-4">
            <div><h3 class="text-sm font-black">专线运单全透明台账</h3><p class="mt-1 text-[11px] text-muted-foreground">{{ filteredWaybills.length }} 条记录 · 面单和轨迹只展示真实接口返回</p></div>
            <span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold text-muted-foreground">{{ waybillSelectedIds.length }} 已选择</span>
          </div>
          <div v-if="filteredWaybills.length === 0" class="flex min-h-64 flex-col items-center justify-center gap-3 px-6 py-12 text-center">
            <span class="flex size-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground"><PackageCheck class="size-5" /></span>
            <div><p class="text-sm font-black">暂无燕文运单记录</p><p class="mt-1 max-w-lg text-xs leading-5 text-muted-foreground">接入 express.order.create、express.order.get 和 express.order.getlist 后，将展示订单、燕文单号、尾程单号、申报品名、重量与交运状态。</p></div>
          </div>
          <div v-else class="overflow-x-auto">
            <table class="w-full min-w-[1220px] text-left text-sm">
              <thead class="bg-muted/30 text-[10px] font-black uppercase tracking-wider text-muted-foreground"><tr><th class="w-12 px-5 py-3"><span class="sr-only">选择</span></th><th class="px-4 py-3">订单编号 / 时间</th><th class="px-4 py-3">燕文单号 / 尾程单号</th><th class="px-4 py-3">渠道 / 交货仓</th><th class="px-4 py-3">目的国 / 买家</th><th class="px-4 py-3">申报品名 / 重量</th><th class="px-4 py-3">状态</th><th class="px-5 py-3 text-right">操作</th></tr></thead>
              <tbody class="divide-y divide-dashed divide-border/70"><tr v-for="record in filteredWaybills" :key="record.id"><td class="px-5 py-4"><input type="checkbox" :checked="waybillSelectedIds.includes(record.id)" class="size-4 rounded border-border" @change="toggleWaybill(record.id)" /></td><td class="px-4 py-4"><p class="font-mono text-xs font-bold">{{ record.orderNo }}</p><p class="mt-1 text-[11px] text-muted-foreground">{{ record.createdAt }}</p></td><td class="px-4 py-4"><p class="font-mono text-xs font-bold">{{ record.waybill || '尚未推单' }}</p><p class="mt-1 font-mono text-[11px] text-muted-foreground">{{ record.lastMile || '—' }}</p></td><td class="px-4 py-4"><p class="font-semibold">{{ record.channel }}</p><p class="mt-1 text-[11px] text-muted-foreground">{{ record.warehouse }}</p></td><td class="px-4 py-4"><p class="font-semibold">{{ record.destination }}</p><p class="mt-1 text-[11px] text-muted-foreground">{{ record.consignee }}</p></td><td class="px-4 py-4"><p>{{ record.description }}</p><p class="mt-1 text-[11px] text-muted-foreground">{{ record.weight }}</p></td><td class="px-4 py-4"><span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold text-muted-foreground">{{ record.status }}</span></td><td class="px-5 py-4 text-right"><div class="flex justify-end gap-1"><Button variant="ghost" size="sm" :disabled="!canShip" :title="canShip ? undefined : '需要 logistics:yanwen:ship 权限'"><Download class="size-3.5" />面单</Button><Button variant="ghost" size="sm" :disabled="!canCancel" :title="canCancel ? undefined : '需要 logistics:yanwen:cancel 权限'" @click="cancelWaybill(record.id)">取消</Button></div></td></tr></tbody>
            </table>
          </div>
        </section>
      </section>
</template>
