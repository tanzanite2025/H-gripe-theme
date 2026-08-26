<template>
  <Card class="h-full min-h-0 overflow-hidden py-0">
    <CardHeader class="shrink-0 border-b bg-muted/30 px-4 py-3">
      <CardTitle class="flex items-center gap-2">
        <UserRound class="size-4 text-primary" />
        客户上下文
      </CardTitle>
      <CardDescription>只读事实源：账号、购物车、心愿单、订单和浏览记录</CardDescription>
    </CardHeader>

    <CardContent class="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
      <div v-if="!selectedConversation" class="flex h-full min-h-0 flex-col items-center justify-center text-center text-muted-foreground">
        <Info class="mb-2 size-7 opacity-55" />
        <p class="text-xs leading-6">选择会话后显示客户上下文。</p>
      </div>

      <div v-else-if="loading" class="flex h-full min-h-0 items-center justify-center text-muted-foreground">
        <LoaderCircle class="size-5 animate-spin" />
      </div>

      <div v-else-if="!customerContext" class="rounded-2xl border border-dashed p-4 text-xs leading-6 text-muted-foreground">
        暂时无法读取客户上下文。消息仍可正常收发。
      </div>

      <template v-else>
        <section class="rounded-2xl border border-primary/20 bg-primary/5 p-3">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="flex items-center gap-2 text-xs font-black uppercase tracking-wider">
              <Clock3 class="size-3.5 text-primary" />
              客户当地时间
            </h3>
            <AdminStatusBadge :tone="customerTimezoneValid ? 'green' : 'amber'">
              {{ customerTimezoneValid ? customerLocalTimePhase : '未采集' }}
            </AdminStatusBadge>
          </div>
          <div class="flex items-end justify-between gap-3">
            <div class="min-w-0">
              <p class="font-mono text-3xl font-black tracking-normal text-foreground">{{ customerLocalTime }}</p>
              <p v-if="customerLocalDate" class="mt-1 text-[11px] text-muted-foreground">{{ customerLocalDate }}</p>
            </div>
            <div class="min-w-0 text-right text-[11px]">
              <p class="truncate font-mono font-bold text-foreground">{{ customerTimezoneValid ? customerContact.timezone : '未采集时区' }}</p>
              <p class="mt-1 truncate text-muted-foreground">{{ customerTimezoneSourceLabel }}</p>
            </div>
          </div>
          <p class="mt-3 rounded-xl bg-background/70 px-3 py-2 text-xs leading-5 text-muted-foreground">
            {{ customerLocalTimeHint }}
          </p>
        </section>

        <section class="rounded-2xl border bg-card p-3">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="flex items-center gap-2 text-xs font-black uppercase tracking-wider">
              <UserCheck class="size-3.5 text-primary" />
              身份
            </h3>
            <AdminStatusBadge :tone="customerAccount ? 'green' : 'amber'">
              {{ customerAccount ? '会员' : '匿名' }}
            </AdminStatusBadge>
          </div>

          <div v-if="customerAccount" class="space-y-2 text-xs">
            <div class="rounded-xl bg-muted/45 p-3">
              <div class="flex flex-wrap items-center gap-2">
                <p class="font-black text-foreground">{{ customerAccount.display_name || customerAccount.username || customerAccount.email }}</p>
                <span
                  v-if="customerAccount.member_tier"
                  class="inline-flex h-5 items-center gap-1 rounded-full border border-amber-500/20 bg-amber-500/10 px-2 text-[10px] font-black text-amber-700"
                  :style="tierStyle(customerAccount.member_tier)"
                  :title="`${customerAccount.member_tier.name} · ${Number(customerAccount.member_tier.total_points || 0)} 积分`"
                >
                  <span v-if="customerAccount.member_tier.icon" class="leading-none">{{ customerAccount.member_tier.icon }}</span>
                  {{ customerAccount.member_tier.name }}
                </span>
              </div>
 <p class="mt-1 break-all text-muted-foreground">{{ customerAccount.email || '未填写邮箱'}}</p>
            </div>
            <dl class="grid grid-cols-2 gap-2 text-[11px]">
              <div class="rounded-xl border p-2">
                <dt class="text-muted-foreground">账号 ID</dt>
                <dd class="mt-1 font-mono font-bold">{{ customerAccount.id }}</dd>
              </div>
              <div class="rounded-xl border p-2">
                <dt class="text-muted-foreground">语言</dt>
 <dd class="mt-1 font-mono font-bold">{{ customerAccount.locale || '-'}}</dd>
              </div>
              <div class="rounded-xl border p-2">
                <dt class="text-muted-foreground">状态</dt>
 <dd class="mt-1 font-mono font-bold">{{ customerAccount.status || '-'}}</dd>
              </div>
              <div class="rounded-xl border p-2">
                <dt class="text-muted-foreground">注册</dt>
                <dd class="mt-1 font-mono font-bold">{{ formatShortDate(customerAccount.created_at) }}</dd>
              </div>
            </dl>
          </div>

          <div v-else class="space-y-3 text-xs leading-6 text-muted-foreground">
            <p class="rounded-xl bg-amber-500/10 p-3 text-amber-700 dark:text-amber-300">
              {{ customerAnonymous?.note || '匿名访客暂未绑定账号。' }}
            </p>
            <div class="rounded-xl border p-3">
              <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">visitor hash</span>
 <span class="mt-1 block font-mono text-foreground">{{ customerAnonymous?.visitor_hash_preview || '未绑定'}}</span>
            </div>
          </div>
        </section>

        <section class="rounded-2xl border bg-card p-3">
          <h3 class="mb-3 flex items-center gap-2 text-xs font-black uppercase tracking-wider">
            <Mail class="size-3.5 text-primary" />
            联系与地区
          </h3>
          <div class="space-y-2 text-xs">
            <div class="flex items-start gap-2 rounded-xl border p-2">
              <Mail class="mt-0.5 size-3.5 text-muted-foreground" />
              <div class="min-w-0">
 <p class="break-all font-bold text-foreground">{{ customerContact.email || '未采集邮箱'}}</p>
 <p class="text-[11px] text-muted-foreground">来源：{{ customerContact.email_source || 'not_captured'}}</p>
              </div>
            </div>
            <div class="flex items-start gap-2 rounded-xl border p-2">
              <Info class="mt-0.5 size-3.5 text-muted-foreground" />
              <div class="min-w-0">
 <p class="font-bold text-foreground">{{ customerContact.locale || '未采集语言'}}</p>
 <p class="text-[11px] text-muted-foreground">来源：{{ customerContact.locale_source || 'not_captured'}}</p>
              </div>
            </div>
            <div class="flex items-start gap-2 rounded-xl border p-2">
              <MapPin class="mt-0.5 size-3.5 text-muted-foreground" />
              <div class="min-w-0">
 <p class="font-bold text-foreground">{{ signalItems[0]?.value || '未采集地区'}}</p>
 <p class="text-[11px] text-muted-foreground">{{ signalItems[0]?.reason || '需要 visitor profile / GeoIP 层'}}</p>
              </div>
            </div>
          </div>
        </section>

        <section class="rounded-2xl border bg-card p-3">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="flex items-center gap-2 text-xs font-black uppercase tracking-wider">
              <ShoppingCart class="size-3.5 text-primary" />
              购物车
            </h3>
            <AdminStatusBadge :tone="customerCart.available ? 'green' : 'amber'">
              {{ customerCart.available ? `${customerCart.item_count || 0} 件` : '未绑定' }}
            </AdminStatusBadge>
          </div>
          <p v-if="!customerCart.available" class="rounded-xl bg-muted/45 p-3 text-xs leading-6 text-muted-foreground">
            {{ customerCart.reason }}
          </p>
          <div v-else class="space-y-2">
            <div class="flex items-center justify-between rounded-xl bg-muted/45 p-3 text-xs">
              <span>合计</span>
              <strong>{{ formatMoney(customerCart.total) }}</strong>
            </div>
            <p v-if="!customerCart.items?.length" class="rounded-xl border border-dashed p-3 text-xs text-muted-foreground">购物车为空</p>
            <article v-for="item in customerCart.items" :key="item.id" class="flex gap-2 rounded-xl border p-2">
              <div class="size-12 shrink-0 overflow-hidden rounded-lg bg-muted">
                <img v-if="item.image" :src="item.image" :alt="item.name" class="size-full object-cover" />
              </div>
              <div class="min-w-0 flex-1 text-xs">
                <p class="truncate font-bold">{{ item.name }}</p>
 <p class="mt-0.5 truncate text-[11px] text-muted-foreground">{{ item.sku || item.variant_name || '无 SKU'}}</p>
                <p class="mt-1 font-mono text-[11px]">x{{ item.quantity }} · {{ formatMoney(item.line_total) }}</p>
              </div>
            </article>
          </div>
        </section>

        <section class="rounded-2xl border bg-card p-3">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="flex items-center gap-2 text-xs font-black uppercase tracking-wider">
              <Heart class="size-3.5 text-primary" />
              心愿单
            </h3>
            <AdminStatusBadge :tone="customerWishlist.available ? 'green' : 'amber'">
              {{ customerWishlist.available ? `${customerWishlist.count || 0} 个` : '不可读' }}
            </AdminStatusBadge>
          </div>
          <p v-if="!customerWishlist.available" class="rounded-xl bg-muted/45 p-3 text-xs leading-6 text-muted-foreground">
            {{ customerWishlist.reason }}
          </p>
          <div v-else class="space-y-2">
            <p v-if="!customerWishlist.items?.length" class="rounded-xl border border-dashed p-3 text-xs text-muted-foreground">暂无心愿单</p>
            <article v-for="item in customerWishlist.items" :key="item.id" class="flex gap-2 rounded-xl border p-2">
              <div class="size-10 shrink-0 overflow-hidden rounded-lg bg-muted">
                <img v-if="item.image" :src="item.image" :alt="item.name" class="size-full object-cover" />
              </div>
              <div class="min-w-0 flex-1 text-xs">
                <p class="truncate font-bold">{{ item.name }}</p>
                <p class="truncate text-[11px] text-muted-foreground">{{ item.sku || `产品 ${item.product_id}` }}</p>
              </div>
            </article>
          </div>
        </section>

        <section class="rounded-2xl border bg-card p-3">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="flex items-center gap-2 text-xs font-black uppercase tracking-wider">
              <PackageCheck class="size-3.5 text-primary" />
              最近订单
            </h3>
            <AdminStatusBadge :tone="customerOrders.available ? 'green' : 'amber'">
              {{ customerOrders.available ? `${customerOrders.total || 0} 单` : '不可读' }}
            </AdminStatusBadge>
          </div>
          <p v-if="!customerOrders.available" class="rounded-xl bg-muted/45 p-3 text-xs leading-6 text-muted-foreground">
            {{ customerOrders.reason }}
          </p>
          <div v-else class="space-y-2">
            <p v-if="!customerOrders.items?.length" class="rounded-xl border border-dashed p-3 text-xs text-muted-foreground">暂无订单</p>
            <article v-for="item in customerOrders.items" :key="item.id" class="rounded-xl border p-2 text-xs">
              <div class="flex items-center justify-between gap-2">
                <strong class="truncate">{{ item.order_number }}</strong>
                <span class="font-mono">{{ formatMoney(item.total_amount) }}</span>
              </div>
              <p class="mt-1 text-[11px] text-muted-foreground">
                {{ item.status }} / {{ item.payment_status }} / {{ item.shipping_status }} · {{ formatShortDate(item.created_at) }}
              </p>
            </article>
          </div>
        </section>

        <section class="rounded-2xl border bg-card p-3">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="flex items-center gap-2 text-xs font-black uppercase tracking-wider">
              <History class="size-3.5 text-primary" />
              浏览历史
            </h3>
            <AdminStatusBadge :tone="customerBrowsing.available ? 'green' : 'amber'">
              {{ customerBrowsing.available ? `${customerBrowsing.count || 0} 条` : '不可读' }}
            </AdminStatusBadge>
          </div>
          <p v-if="!customerBrowsing.available" class="rounded-xl bg-muted/45 p-3 text-xs leading-6 text-muted-foreground">
            {{ customerBrowsing.reason }}
          </p>
          <div v-else class="space-y-2">
            <p v-if="!customerBrowsing.items?.length" class="rounded-xl border border-dashed p-3 text-xs text-muted-foreground">暂无浏览历史</p>
            <article v-for="item in customerBrowsing.items" :key="item.product_id" class="rounded-xl border p-2 text-xs">
              <div class="flex items-center justify-between gap-2">
                <strong>产品 {{ item.product_id }}</strong>
                <span class="font-mono">{{ item.view_count }} 次</span>
              </div>
              <p class="mt-1 text-[11px] text-muted-foreground">最后浏览：{{ formatDate(item.last_viewed_at) }}</p>
            </article>
          </div>
        </section>

        <section class="rounded-2xl border bg-card p-3">
          <h3 class="mb-3 flex items-center gap-2 text-xs font-black uppercase tracking-wider">
            <Info class="size-3.5 text-primary" />
            采集状态
          </h3>
          <div class="space-y-2">
            <div v-for="signal in signalItems" :key="signal.key" class="rounded-xl border p-2 text-xs">
              <div class="flex items-center justify-between gap-2">
                <span class="font-bold">{{ signal.label }}</span>
                <AdminStatusBadge :tone="signalTone(signal.status)">{{ signal.status }}</AdminStatusBadge>
              </div>
              <p class="mt-1 break-words text-[11px] leading-5 text-muted-foreground">
                {{ signal.value || signal.reason || '-' }}
              </p>
            </div>
          </div>
        </section>
      </template>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { Clock3, Heart, History, Info, LoaderCircle, Mail, MapPin, PackageCheck, ShoppingCart, UserCheck, UserRound } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  formatDate,
  formatCustomerLocalDate,
  formatCustomerLocalTime,
  formatMoney,
  formatShortDate,
  customerLocalTimeHint as getCustomerLocalTimeHint,
  customerLocalTimePhase as getCustomerLocalTimePhase,
  customerTimezoneSourceLabel as getCustomerTimezoneSourceLabel,
  isValidCustomerTimezone,
  signalTone,
  tierStyle,
} from '@/lib/customerServicePresentation'
import type {
  CustomerAccount,
  CustomerAnonymous,
  CustomerBrowsing,
  CustomerCart,
  CustomerContact,
  CustomerContext,
  CustomerConversation,
  CustomerOrders,
  CustomerSignal,
  CustomerWishlist,
} from './customerServiceTypes'

const props = withDefaults(defineProps<{
  selectedConversation?: CustomerConversation | null
  customerContext?: CustomerContext | null
  loading?: boolean
}>(), {
  selectedConversation: null,
  customerContext: null,
  loading: false,
})

const customerAccount = computed<CustomerAccount | null>(() => props.customerContext?.customer?.account || null)
const customerAnonymous = computed<CustomerAnonymous | null>(() => props.customerContext?.customer?.anonymous || null)
const customerContact = computed<CustomerContact>(() => props.customerContext?.contact || {})
const customerClockNow = ref(new Date())
const customerCart = computed<CustomerCart>(() => props.customerContext?.cart || { available: false, items: [] })
const customerWishlist = computed<CustomerWishlist>(() => props.customerContext?.wishlist || { available: false, items: [] })
const customerOrders = computed<CustomerOrders>(() => props.customerContext?.orders || { available: false, items: [] })
const customerBrowsing = computed<CustomerBrowsing>(() => props.customerContext?.browsing || { available: false, items: [] })
const customerTimezoneValid = computed(() => isValidCustomerTimezone(customerContact.value.timezone))
const customerLocalTime = computed(() => formatCustomerLocalTime(customerClockNow.value, customerContact.value.timezone))
const customerLocalDate = computed(() => formatCustomerLocalDate(customerClockNow.value, customerContact.value.timezone))
const customerLocalTimePhase = computed(() => getCustomerLocalTimePhase(customerClockNow.value, customerContact.value.timezone))
const customerLocalTimeHint = computed(() => getCustomerLocalTimeHint(customerClockNow.value, customerContact.value.timezone))
const customerTimezoneSourceLabel = computed(() => getCustomerTimezoneSourceLabel(customerContact.value.timezone_source))
interface SignalItem extends CustomerSignal {
  key: string
  label: string
}

const signalItems = computed<SignalItem[]>(() => {
  const signals = props.customerContext?.signals || {}
  return [
    { key: 'region', label: '地区', ...(signals.region || {}) },
    { key: 'cart_session', label: '购物车会话', ...(signals.cart_session || {}) },
    { key: 'email_capture', label: '邮箱采集', ...(signals.email_capture || {}) },
    { key: 'visitor_profile', label: '访客档案', ...(signals.visitor_profile || {}) },
  ]
})

let customerClockTimer: number | null = null

onMounted(() => {
  customerClockTimer = window.setInterval(() => {
    customerClockNow.value = new Date()
  }, 30_000)
})

onBeforeUnmount(() => {
  if (customerClockTimer) {
    window.clearInterval(customerClockTimer)
    customerClockTimer = null
  }
})
</script>
