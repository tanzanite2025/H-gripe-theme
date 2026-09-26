<script setup lang="ts">
import { computed, ref } from 'vue'
import { ClipboardCheck, Download, Plus } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const canManage = computed(() => authStore.hasPermission('logistics:fpx:manage'))

const rmaSearch = ref('')
const rmaStatus = ref('')
const rmaDialogOpen = ref(false)
const rmaForm = ref({ orderNo: '', customer: '', carrier: '', reason: '', notes: '' })
const rmaRecords = ref<Array<{
  id: string
  orderNo: string
  customer: string
  carrier: string
  inspection: string
  disposition: string
  status: string
}>>([])
const filteredRmaRecords = computed(() => {
  const query = rmaSearch.value.trim().toLowerCase()
  return rmaRecords.value.filter((record) => {
    const values = [record.id, record.orderNo, record.customer, record.carrier, record.inspection]
    const matchesSearch = !query || values.some((value) => value.toLowerCase().includes(query))
    const matchesStatus = !rmaStatus.value || record.status === rmaStatus.value
    return matchesSearch && matchesStatus
  })
})
const canCreateRma = computed(() => canManage.value && rmaForm.value.orderNo.trim().length > 0 && rmaForm.value.carrier.trim().length > 0)
const saveRmaDraft = () => {
  if (!canCreateRma.value) return
  const form = rmaForm.value
  rmaRecords.value.unshift({
    id: 'RMA-DRAFT-' + Date.now(),
    orderNo: form.orderNo.trim(),
    customer: form.customer.trim() || '待同步',
    carrier: form.carrier.trim(),
    inspection: '待质检',
    disposition: '待判定',
    status: '当前会话草稿',
  })
  rmaForm.value = { orderNo: '', customer: '', carrier: '', reason: '', notes: '' }
  rmaDialogOpen.value = false
}
</script>

<template>
      <section class="space-y-5 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-rose-600">Reverse Logistics & RMA Inspection</p>
            <h2 class="mt-1 text-lg font-black tracking-tight">逆向退件台账</h2>
            <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">记录内部退件与承运商委托信息。4PX 公开直发文档未确认独立 RMA 创建接口，真实逆向运输需单独确认承运商能力。</p>
          </div>
          <Button size="sm" :disabled="!canManage" @click="rmaDialogOpen = true"><Plus class="size-3.5" />登记退件台账</Button>
        </div>
        <div class="grid gap-3 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4 md:grid-cols-[minmax(0,1fr)_220px_auto]">
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">搜索退件</span><Input v-model="rmaSearch" placeholder="退货单号 / 原订单号 / 买家 / 承运商" /></label>
          <label class="space-y-1.5"><span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground">状态</span><select v-model="rmaStatus" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"><option value="">全部状态</option><option value="当前会话草稿">当前会话草稿</option><option value="待承运商接收">待承运商接收</option><option value="运输中">运输中</option><option value="已签收">已签收</option></select></label>
          <div class="flex items-end"><Button variant="ghost" size="sm" class="w-full" @click="rmaSearch = ''; rmaStatus = ''">重置筛选</Button></div>
        </div>
      </section>

      <section class="overflow-hidden rounded-[28px] border border-dashed border-border/80 bg-card shadow-sm">
        <div class="flex items-center justify-between gap-3 border-b border-dashed border-border/80 px-5 py-4 sm:px-6">
          <div><h2 class="text-base font-black">退件运输台账</h2><p class="mt-1 text-xs text-muted-foreground">{{ filteredRmaRecords.length }} 条记录 · 仅记录内部流程与承运商状态</p></div>
          <span class="rounded-full border border-amber-500/30 bg-amber-500/10 px-3 py-1.5 text-[11px] font-bold text-amber-700 dark:text-amber-300">承运商能力待确认</span>
        </div>
        <div v-if="filteredRmaRecords.length === 0" class="flex min-h-64 flex-col items-center justify-center gap-3 px-6 py-12 text-center">
          <span class="flex size-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground"><ClipboardCheck class="size-5" /></span>
          <div><p class="text-sm font-black">暂无退件运输记录</p><p class="mt-1 max-w-lg text-xs leading-5 text-muted-foreground">4PX 全球直发公开文档未提供独立 RMA 创建/查询接口；下单接口仅提供境内/境外异常退件策略和地址字段。这里先承载内部退件台账，接入承运商退件能力后再提交真实运输委托。</p></div>
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-[980px] text-left text-sm">
            <thead class="bg-muted/30 text-[10px] font-black uppercase tracking-wider text-muted-foreground"><tr><th class="px-5 py-3">退货单号 / 原订单号</th><th class="px-4 py-3">买家 / 承运商</th><th class="px-4 py-3">运输状态</th><th class="px-4 py-3">签收凭证</th><th class="px-4 py-3">处置决策</th><th class="px-4 py-3">状态</th><th class="px-5 py-3 text-right">操作</th></tr></thead>
            <tbody class="divide-y divide-dashed divide-border/70"><tr v-for="record in filteredRmaRecords" :key="record.id"><td class="px-5 py-4"><p class="font-mono text-xs font-bold">{{ record.id }}</p><p class="mt-1 text-[11px] text-muted-foreground">{{ record.orderNo }}</p></td><td class="px-4 py-4"><p class="font-semibold">{{ record.customer }}</p><p class="mt-1 text-[11px] text-muted-foreground">{{ record.carrier }}</p></td><td class="px-4 py-4">{{ record.inspection }}</td><td class="px-4 py-4 text-xs text-muted-foreground">待同步凭证</td><td class="px-4 py-4">{{ record.disposition }}</td><td class="px-4 py-4"><span class="rounded-full bg-amber-500/10 px-2.5 py-1 text-[10px] font-bold text-amber-700 dark:text-amber-300">{{ record.status }}</span></td><td class="px-5 py-4 text-right"><Button variant="ghost" size="sm" disabled><Download class="size-3.5" />证据包</Button></td></tr></tbody>
          </table>
        </div>
      </section>

      <Dialog v-model:open="rmaDialogOpen">
        <DialogContent size="lg">
          <DialogHeader><DialogTitle>登记逆向退件</DialogTitle><DialogDescription>当前仅保存为会话草稿，离开页面后不会持久化；4PX 公开直发接口没有独立 RMA 委托方法，真实退件需要配置承运商或内部退件流程。</DialogDescription></DialogHeader>
          <div class="grid gap-4 py-2 sm:grid-cols-2">
            <label class="space-y-1.5"><span class="text-xs font-bold">原订单号</span><Input v-model="rmaForm.orderNo" placeholder="例如 TZ-20260905-8120" /></label>
            <label class="space-y-1.5"><span class="text-xs font-bold">退件买家</span><Input v-model="rmaForm.customer" placeholder="买家姓名，可选" /></label>
            <label class="space-y-1.5 sm:col-span-2"><span class="text-xs font-bold">退件承运商</span><select v-model="rmaForm.carrier" class="h-9 w-full rounded-md border border-border bg-background px-3 text-sm"><option value="">请选择承运商</option><option value="4PX（商务授权待确认）">4PX（商务授权待确认）</option><option value="FedEx">FedEx</option><option value="UPS">UPS</option></select></label>
            <label class="space-y-1.5 sm:col-span-2"><span class="text-xs font-bold">退货原因</span><Input v-model="rmaForm.reason" placeholder="规格不兼容 / 质保问题 / 外包装破损" /></label>
            <label class="space-y-1.5 sm:col-span-2"><span class="text-xs font-bold">处置备注</span><textarea v-model="rmaForm.notes" rows="4" class="w-full rounded-md border border-border bg-background px-3 py-2 text-sm" placeholder="记录客户诉求、检查重点和后续索赔要求" /></label>
          </div>
          <DialogFooter><Button variant="outline" @click="rmaDialogOpen = false">取消</Button><Button :disabled="!canCreateRma" @click="saveRmaDraft">保存会话草稿</Button></DialogFooter>
        </DialogContent>
      </Dialog>
</template>
