<script setup lang="ts">
import { computed, ref } from 'vue'
import { AlertTriangle, Download, Package, Printer, RefreshCw, Search, Send } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const directSearch = ref('')
const directChannel = ref('')
const directStatus = ref('')
const directRefreshing = ref(false)
const directSelectedIds = ref<string[]>([])

const directOrders = ref<Array<{
  id: string
  orderNo: string
  waybill: string
  product: string
  channel: string
  destination: string
  consignee: string
  status: string
}>>([])

const filteredDirectOrders = computed(() => {
  const query = directSearch.value.trim().toLowerCase()
  return directOrders.value.filter((order) => {
    const matchesSearch = !query || [order.orderNo, order.waybill, order.product, order.consignee].some((value) => value.toLowerCase().includes(query))
    const matchesChannel = !directChannel.value || order.channel === directChannel.value
    const matchesStatus = !directStatus.value || order.status === directStatus.value
    return matchesSearch && matchesChannel && matchesStatus
  })
})

const refreshDirectOrders = async () => {
  directRefreshing.value = true
  await Promise.resolve()
  directRefreshing.value = false
}

const resetDirectFilters = () => {
  directSearch.value = ''
  directChannel.value = ''
  directStatus.value = ''
}

const toggleDirectOrder = (id: string) => {
  directSelectedIds.value = directSelectedIds.value.includes(id)
    ? directSelectedIds.value.filter((selectedId) => selectedId !== id)
    : [...directSelectedIds.value, id]
}

const canShip = computed(() => authStore.hasPermission('logistics:fpx:ship'))
const hasSelectedDirectOrders = computed(() => directSelectedIds.value.length > 0)
</script>

<template>
      <section class="space-y-4 rounded-[28px] border border-dashed border-border/80 bg-card p-5 shadow-sm sm:p-6">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <p class="text-[10px] font-black uppercase tracking-[0.18em] text-orange-600">Bulky Direct Shipping & Labels</p>
            <h2 class="mt-1 text-lg font-black tracking-tight">大件专线直发与面单</h2>
            <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">处理商城待发货的轮组、车架大件，后续接入 `ds.xms.order.create`、面单和交运状态同步。</p>
          </div>
          <div class="flex flex-wrap gap-2">
            <Button variant="outline" size="sm" :disabled="directRefreshing" @click="refreshDirectOrders">
              <RefreshCw :class="['size-3.5', { 'animate-spin': directRefreshing }]" />
              刷新
            </Button>
            <Button size="sm" :disabled="!canShip || !hasSelectedDirectOrders" :title="canShip ? undefined : '需要 logistics:fpx:ship 权限'">
              <Send class="size-3.5" />
              批量直发推单
            </Button>
            <Button variant="outline" size="sm" :disabled="!canShip || !hasSelectedDirectOrders" :title="canShip ? undefined : '需要 logistics:fpx:ship 权限'">
              <Download class="size-3.5" />
              获取面单 PDF
            </Button>
          </div>
        </div>

        <div class="grid gap-3 rounded-[24px] border border-dashed border-border/80 bg-muted/20 p-4 md:grid-cols-2 xl:grid-cols-[minmax(0,1.4fr)_repeat(2,minmax(0,1fr))_auto]">
          <label class="space-y-1.5">
            <span class="flex items-center gap-1.5 text-[10px] font-black uppercase tracking-wider text-muted-foreground/80"><Search class="size-3" /> 搜索台账</span>
            <Input v-model="directSearch" placeholder="商城单号 / 4PX 单号 / 客户 / 货品" />
          </label>
          <label class="space-y-1.5">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">渠道</span>
            <select v-model="directChannel" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
              <option value="">全部渠道</option>
              <option value="4PX全球特快">4PX 全球特快</option>
              <option value="大包专线">大包专线</option>
              <option value="优先派送">优先派送</option>
            </select>
          </label>
          <label class="space-y-1.5">
            <span class="text-[10px] font-black uppercase tracking-wider text-muted-foreground/80">状态</span>
            <select v-model="directStatus" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
              <option value="">全部状态</option>
              <option value="待推单">待推单</option>
              <option value="已建单">已建单</option>
              <option value="已打面单">已打面单</option>
              <option value="交运在途">交运在途</option>
            </select>
          </label>
          <div class="flex items-end">
            <Button variant="ghost" size="sm" class="w-full" @click="resetDirectFilters">重置筛选</Button>
          </div>
        </div>
      </section>

      <section class="overflow-hidden rounded-[28px] border border-dashed border-border/80 bg-card shadow-sm">
        <div class="flex flex-col gap-2 border-b border-dashed border-border/80 px-5 py-4 sm:flex-row sm:items-center sm:justify-between sm:px-6">
          <div>
            <h2 class="text-base font-black">大件直发台账</h2>
            <p class="mt-1 text-xs text-muted-foreground">{{ filteredDirectOrders.length }} 条记录 · 当前数据源待接入</p>
          </div>
          <span class="inline-flex w-fit items-center gap-1.5 rounded-full border border-amber-500/30 bg-amber-500/10 px-3 py-1.5 text-[11px] font-bold text-amber-700 dark:text-amber-300">
            <AlertTriangle class="size-3.5" />
            暂无同步数据
          </span>
        </div>

        <div v-if="filteredDirectOrders.length === 0" class="flex min-h-64 flex-col items-center justify-center gap-3 px-6 py-12 text-center">
          <span class="flex size-12 items-center justify-center rounded-2xl bg-muted text-muted-foreground"><Package class="size-5" /></span>
          <div>
            <p class="text-sm font-black">暂时没有可展示的 4PX 直发订单</p>
            <p class="mt-1 max-w-md text-xs leading-5 text-muted-foreground">完成 4PX 接口配置并接入商城待发货订单后，这里将显示尺寸、计费重、运单号、面单和交运状态。</p>
          </div>
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-[980px] text-left text-sm">
            <thead class="bg-muted/30 text-[10px] font-black uppercase tracking-wider text-muted-foreground">
              <tr>
                <th class="w-12 px-5 py-3"><span class="sr-only">选择</span></th>
                <th class="px-4 py-3">订单编号</th>
                <th class="px-4 py-3">4PX 单号 / 跟踪号</th>
                <th class="px-4 py-3">货品属性</th>
                <th class="px-4 py-3">目的国 / 收件人</th>
                <th class="px-4 py-3">状态</th>
                <th class="px-5 py-3 text-right">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-dashed divide-border/70">
              <tr v-for="order in filteredDirectOrders" :key="order.id" class="align-top">
                <td class="px-5 py-4"><input type="checkbox" :checked="directSelectedIds.includes(order.id)" class="size-4 accent-primary" @change="toggleDirectOrder(order.id)" /></td>
                <td class="px-4 py-4"><p class="font-mono text-xs font-bold">{{ order.orderNo }}</p><p class="mt-1 text-[11px] text-muted-foreground">待接入下单时间</p></td>
                <td class="px-4 py-4"><p class="font-mono text-xs font-bold">{{ order.waybill || '待推单' }}</p></td>
                <td class="px-4 py-4"><p class="font-semibold">{{ order.product }}</p><p class="mt-1 text-[11px] text-muted-foreground">尺寸与计费重待同步</p></td>
                <td class="px-4 py-4"><p class="font-semibold">{{ order.destination }}</p><p class="mt-1 text-[11px] text-muted-foreground">{{ order.consignee }}</p></td>
                <td class="px-4 py-4"><span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-bold">{{ order.status }}</span></td>
                <td class="px-5 py-4 text-right"><Button variant="ghost" size="sm" :disabled="!canShip"><Printer class="size-3.5" />面单</Button></td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
</template>
