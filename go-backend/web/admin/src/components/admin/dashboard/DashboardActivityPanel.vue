<template>
  <Card class="gap-0 py-0 shadow-none rounded-[24px] border-dashed border-border/80">
    <div class="uds-glow-bg" />
    <Tabs :model-value="activeActivity" class="relative z-10" @update:model-value="updateActiveActivity">
      <CardHeader class="flex flex-col gap-3 border-b border-dashed border-border/70 py-3.5 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <CardTitle class="text-sm font-black tracking-tighter italic uppercase">最近活动</CardTitle>
          <CardDescription class="mt-0.5 text-[9px] font-black uppercase tracking-widest opacity-60">最新业务动态记录</CardDescription>
        </div>
        <TabsList variant="line" class="rounded-full bg-muted/40 p-1">
          <TabsTrigger value="orders" class="rounded-full text-xs font-bold px-3">订单</TabsTrigger>
          <TabsTrigger value="users" class="rounded-full text-xs font-bold px-3">用户</TabsTrigger>
          <TabsTrigger value="tickets" class="rounded-full text-xs font-bold px-3">工单</TabsTrigger>
        </TabsList>
      </CardHeader>

      <CardContent class="pb-4 pt-2">
        <TabsContent value="orders" class="mt-0">
          <div class="flex min-h-9 items-center justify-between">
            <strong class="text-xs font-black uppercase tracking-wider text-foreground/80">最近订单</strong>
            <Button variant="link" size="sm" class="px-0 font-bold text-xs" @click="emit('navigate', '/orders')">
              查看全部
              <ArrowRight class="size-3.5 ml-1" />
            </Button>
          </div>
          <EmptyActivity v-if="recentOrders.length === 0" label="暂无订单" />
          <div v-else class="grid grid-cols-1 gap-x-6 md:grid-cols-2">
            <div v-for="order in recentOrders" :key="order.id" class="flex min-w-0 items-center justify-between gap-4 border-b border-dashed border-border/60 py-2.5">
              <div class="min-w-0">
                <strong class="block truncate text-xs font-mono font-bold">#{{ order.order_number }}</strong>
                <span class="mt-0.5 block truncate text-[11px] font-mono text-muted-foreground">¥{{ formatNumber(order.total_amount) }}</span>
              </div>
              <AdminStatusBadge :tone="orderStatusTone(order.status)">
                {{ getOrderStatusName(order.status) }}
              </AdminStatusBadge>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="users" class="mt-0">
          <div class="flex min-h-9 items-center justify-between">
            <strong class="text-xs font-black uppercase tracking-wider text-foreground/80">最近用户</strong>
            <Button variant="link" size="sm" class="px-0 font-bold text-xs" @click="emit('navigate', '/access/admin-users')">
              查看全部
              <ArrowRight class="size-3.5 ml-1" />
            </Button>
          </div>
          <EmptyActivity v-if="recentUsers.length === 0" label="暂无用户" />
          <div v-else class="grid grid-cols-1 gap-x-6 md:grid-cols-2">
            <div v-for="recentUser in recentUsers" :key="recentUser.id" class="flex min-w-0 items-center justify-between gap-4 border-b border-dashed border-border/60 py-2.5">
              <div class="min-w-0">
                <strong class="block truncate text-xs font-bold">{{ recentUser.username }}</strong>
                <span class="mt-0.5 block truncate text-[11px] font-mono text-muted-foreground">{{ recentUser.email }}</span>
              </div>
              <AdminStatusBadge :tone="roleTone(recentUser.role)">
                {{ getRoleName(recentUser.role) }}
              </AdminStatusBadge>
            </div>
          </div>
        </TabsContent>

        <TabsContent value="tickets" class="mt-0">
          <div class="flex min-h-9 items-center justify-between">
            <strong class="text-xs font-black uppercase tracking-wider text-foreground/80">最近工单</strong>
            <Button variant="link" size="sm" class="px-0 font-bold text-xs" @click="emit('navigate', '/tickets')">
              查看全部
              <ArrowRight class="size-3.5 ml-1" />
            </Button>
          </div>
          <EmptyActivity v-if="recentTickets.length === 0" label="暂无工单" />
          <div v-else class="grid grid-cols-1 gap-x-6 md:grid-cols-2">
            <div v-for="ticket in recentTickets" :key="ticket.id" class="flex min-w-0 items-center justify-between gap-4 border-b border-dashed border-border/60 py-2.5">
              <div class="min-w-0">
                <strong class="block truncate text-xs font-bold">{{ ticket.subject }}</strong>
                <span class="mt-0.5 block truncate text-[11px] font-mono text-muted-foreground">{{ ticket.category }}</span>
              </div>
              <AdminStatusBadge :tone="ticketStatusTone(ticket.status)">
                {{ getTicketStatusName(ticket.status) }}
              </AdminStatusBadge>
            </div>
          </div>
        </TabsContent>
      </CardContent>
    </Tabs>
  </Card>
</template>

<script setup lang="ts">
import { defineComponent, h } from 'vue'
import { ArrowRight, Inbox } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type {
  DashboardActivity,
  DashboardLabelResolver,
  DashboardNumberFormatter,
  DashboardRecentOrder,
  DashboardRecentTicket,
  DashboardRecentUser,
  DashboardToneResolver
} from './dashboardTypes'

const EmptyActivity = defineComponent({
  props: {
    label: { type: String, required: true },
  },
  setup(props) {
    return () => h('div', { class: 'flex min-h-36 flex-col items-center justify-center text-muted-foreground' }, [
      h(Inbox, { class: 'mb-2 size-6 opacity-55' }),
      h('span', { class: 'text-xs' }, props.label),
    ])
  },
})

withDefaults(defineProps<{
  activeActivity?: DashboardActivity
  recentOrders?: DashboardRecentOrder[]
  recentUsers?: DashboardRecentUser[]
  recentTickets?: DashboardRecentTicket[]
  formatNumber: DashboardNumberFormatter
  getOrderStatusName: DashboardLabelResolver
  orderStatusTone: DashboardToneResolver
  getRoleName: DashboardLabelResolver
  roleTone: DashboardToneResolver
  getTicketStatusName: DashboardLabelResolver
  ticketStatusTone: DashboardToneResolver
}>(), {
  activeActivity: 'orders',
  recentOrders: () => [],
  recentUsers: () => [],
  recentTickets: () => []
})

const emit = defineEmits<{
  (event: 'update:active-activity', value: DashboardActivity): void
  (event: 'navigate', path: string): void
}>()

const activityValues: DashboardActivity[] = ['orders', 'users', 'tickets']
const isDashboardActivity = (value: unknown): value is DashboardActivity => (
  typeof value === 'string' && activityValues.includes(value as DashboardActivity)
)

const updateActiveActivity = (value: unknown): void => {
  if (isDashboardActivity(value)) emit('update:active-activity', value)
}
</script>
