<template>
  <Card class="h-full min-h-0 overflow-hidden py-0">
    <CardHeader class="shrink-0 border-b bg-muted/30 px-4 py-3">
      <CardTitle class="flex items-center gap-2">
        <UserRound class="size-4 text-primary" />
        客户上下文
      </CardTitle>
      <CardDescription>只读事实源：账号、订单履约/价格、售后退款、购物车、心愿单和浏览记录</CardDescription>
    </CardHeader>

    <CardContent class="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
      <div v-if="!selectedConversation" class="flex h-full min-h-0 flex-col items-center justify-center text-center text-muted-foreground">
        <Info class="mb-2 size-7 opacity-55" />
        <p class="text-xs leading-6">选择会话后显示客户上下文。</p>
      </div>

      <div v-else-if="loading" class="flex h-full min-h-0 items-center justify-center text-muted-foreground">
        <LoaderCircle class="size-5 animate-spin" />
      </div>

      <div v-else-if="contextError && customerContext" class="rounded-2xl border border-amber-500/30 bg-amber-500/10 p-4 text-xs leading-6 text-amber-800 dark:text-amber-200">
        <p class="font-bold">上下文刷新失败，当前显示的是上次成功读取的数据。</p>
        <p class="mt-1">{{ contextError }}</p>
        <p v-if="contextLastUpdatedAt" class="mt-1 text-[11px] opacity-80">上次更新：{{ formatDate(contextLastUpdatedAt) }}</p>
      </div>

      <div v-else-if="contextError" class="rounded-2xl border border-red-500/30 bg-red-500/10 p-4 text-xs leading-6 text-red-800 dark:text-red-200">
        <p class="font-bold">无法读取客户上下文</p>
        <p class="mt-1">{{ contextError }}</p>
        <p class="mt-1">消息仍可正常收发，请稍后重试。</p>
      </div>

      <div v-else-if="!customerContext" class="rounded-2xl border border-dashed p-4 text-xs leading-6 text-muted-foreground">
        暂无客户上下文数据。若客户未采集时区，时区会明确显示为“未采集时区”。
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
              <p v-if="customerTimezoneValid && customerTimezoneDifference" class="mt-1 truncate text-muted-foreground">{{ customerTimezoneDifference }}</p>
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
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="flex items-center gap-2 text-xs font-black uppercase tracking-wider">
              <MapPin class="size-3.5 text-primary" />
              最近收货地址
            </h3>
            <AdminStatusBadge :tone="factTone(customerShippingAddress.status, customerShippingAddress.available)">
              {{ factStatusLabel(customerShippingAddress.status, customerShippingAddress.available) }}
            </AdminStatusBadge>
          </div>
          <p v-if="!customerShippingAddress.available" class="rounded-xl bg-muted/45 p-3 text-xs leading-6 text-muted-foreground">
            {{ customerShippingAddress.reason || '暂无可用收货地址。' }}
          </p>
          <div v-else class="space-y-2 text-xs">
            <div class="rounded-xl bg-muted/45 p-3">
              <div class="flex items-center justify-between gap-2">
                <strong>{{ customerShippingAddress.recipient_name || '未填写收件人' }}</strong>
                <span class="font-mono text-[11px] text-muted-foreground">{{ customerShippingAddress.source_order_number || '-' }}</span>
              </div>
              <p class="mt-1 leading-5 text-muted-foreground">
                {{ customerShippingAddress.address_line || '-' }}
                <span v-if="customerShippingAddress.city"> · {{ customerShippingAddress.city }}</span>
                <span v-if="customerShippingAddress.state"> · {{ customerShippingAddress.state }}</span>
              </p>
              <p class="mt-1 text-[11px] text-muted-foreground">
                {{ customerShippingAddress.postal_code || '-' }} · {{ customerShippingAddress.country || '-' }}
                <span v-if="customerShippingAddress.phone_present"> · 已留电话</span>
              </p>
            </div>
            <p class="text-[11px] text-muted-foreground">电话和邮箱不会进入客服上下文。</p>
          </div>
        </section>

        <section class="rounded-2xl border bg-card p-3">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="flex items-center gap-2 text-xs font-black uppercase tracking-wider">
              <RotateCcw class="size-3.5 text-primary" />
              售后 / RMA
            </h3>
            <AdminStatusBadge :tone="factTone(customerAfterSales.status, customerAfterSales.available)">
              {{ factStatusLabel(customerAfterSales.status, customerAfterSales.available) }}
            </AdminStatusBadge>
          </div>
          <p v-if="!customerAfterSales.available" class="rounded-xl bg-muted/45 p-3 text-xs leading-6 text-muted-foreground">
            {{ customerAfterSales.reason || '暂无售后事实。' }}
          </p>
          <div v-else class="space-y-2">
            <p v-if="!customerAfterSales.items?.length" class="rounded-xl border border-dashed p-3 text-xs text-muted-foreground">
              {{ customerAfterSales.reason || '暂无售后请求' }}
            </p>
            <article v-for="item in customerAfterSales.items" :key="item.id" class="rounded-xl border p-2 text-xs">
              <div class="flex items-center justify-between gap-2">
                <strong class="truncate">{{ item.type || '售后请求' }}</strong>
                <AdminStatusBadge :tone="statusTone(item.status)">{{ item.status || '-' }}</AdminStatusBadge>
              </div>
              <p class="mt-1 text-[11px] text-muted-foreground">
                {{ item.order_number || `订单 ${item.order_id || '-'}` }} · {{ formatShortDate(item.updated_at || item.created_at) }}
              </p>
              <p v-if="item.reason" class="mt-1 text-xs leading-5 text-muted-foreground">{{ item.reason }}</p>
              <p v-if="item.item_summary" class="mt-1 text-[11px] text-muted-foreground">{{ item.item_summary }}</p>
              <p v-if="item.current_handler_name || item.current_handler_id" class="mt-1 text-[11px] text-muted-foreground">当前处理人：{{ item.current_handler_name || `客服 ${item.current_handler_id}` }}</p>
              <p v-if="item.resolution" class="mt-1 text-[11px] text-muted-foreground">处理结果：{{ item.resolution }}</p>
            </article>
            <p v-if="customerAfterSales.refund_status !== 'available'" class="rounded-xl bg-muted/45 p-2 text-[11px] text-muted-foreground">退款明细：{{ customerAfterSales.refund_reason || '暂无数据' }}</p>
            <article v-for="refund in customerAfterSales.refunds" :key="refund.id" class="rounded-xl border p-2 text-xs">
              <div class="flex items-center justify-between gap-2"><strong>退款 {{ refund.order_number || refund.order_id }}</strong><AdminStatusBadge :tone="statusTone(refund.status)">{{ refund.status || '-' }}</AdminStatusBadge></div>
              <p class="mt-1 text-[11px] text-muted-foreground">{{ formatMoney(refund.amount) }} {{ refund.currency || '' }} · {{ formatShortDate(refund.completed_at || refund.created_at) }}</p>
            </article>
            <p v-if="customerAfterSales.warranty_status !== 'available'" class="rounded-xl bg-muted/45 p-2 text-[11px] text-muted-foreground">保修明细：{{ customerAfterSales.warranty_reason || '暂无数据' }}</p>
            <article v-for="claim in customerAfterSales.warranty_claims" :key="claim.id" class="rounded-xl border p-2 text-xs">
              <div class="flex items-center justify-between gap-2"><strong>保修 {{ claim.order_number || '-' }}</strong><AdminStatusBadge :tone="statusTone(claim.status)">{{ claim.status || '-' }}</AdminStatusBadge></div>
              <p class="mt-1 text-[11px] text-muted-foreground">{{ claim.issue_type || '问题未记录' }}<span v-if="claim.tire_pressure"> · 胎压 {{ claim.tire_pressure }}</span><span v-if="claim.is_tubeless"> · 无内胎</span></p>
            </article>
          </div>
        </section>

        <section class="rounded-2xl border bg-card p-3">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="flex items-center gap-2 text-xs font-black uppercase tracking-wider">
              <ShieldAlert class="size-3.5 text-primary" />
              支付争议
            </h3>
            <AdminStatusBadge :tone="factTone(customerDisputes.status, customerDisputes.available)">
              {{ factStatusLabel(customerDisputes.status, customerDisputes.available) }}
            </AdminStatusBadge>
          </div>
          <p v-if="!customerDisputes.available" class="rounded-xl bg-muted/45 p-3 text-xs leading-6 text-muted-foreground">
            {{ customerDisputes.reason || '暂无支付争议事实。' }}
          </p>
          <div v-else class="space-y-2">
            <p v-if="!customerDisputes.items?.length" class="rounded-xl border border-dashed p-3 text-xs text-muted-foreground">
              {{ customerDisputes.reason || '暂无支付争议' }}
            </p>
            <article v-for="item in customerDisputes.items" :key="`${item.provider}-${item.order_id}-${item.evidence_due_at || item.status}`" class="rounded-xl border p-2 text-xs">
              <div class="flex items-center justify-between gap-2">
                <strong class="uppercase">{{ item.provider || '支付渠道' }}</strong>
                <AdminStatusBadge :tone="statusTone(item.status)">{{ item.status || '-' }}</AdminStatusBadge>
              </div>
              <p class="mt-1 text-[11px] text-muted-foreground">
                {{ item.order_number || `订单 ${item.order_id || '-'}` }} · {{ formatMoney(item.amount) }} {{ item.currency || '' }}
              </p>
              <p v-if="item.reason" class="mt-1 text-xs leading-5 text-muted-foreground">{{ item.reason }}</p>
              <p v-if="item.evidence_due_at" class="mt-1 text-[11px] text-amber-700">证据截止：{{ formatDate(item.evidence_due_at) }}</p>
              <p v-else-if="item.evidence_submitted_at" class="mt-1 text-[11px] text-muted-foreground">证据已提交：{{ formatDate(item.evidence_submitted_at) }}</p>
            </article>
            <p class="text-[11px] text-muted-foreground">仅显示争议摘要，不展示支付凭证或支付提供商 ID。</p>
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
              {{ factStatusLabel(customerCart.status, customerCart.available) }}
            </AdminStatusBadge>
          </div>
          <p v-if="!customerCart.available" class="rounded-xl bg-muted/45 p-3 text-xs leading-6 text-muted-foreground">
            {{ customerCart.reason }}
          </p>
          <div v-else class="space-y-2">
            <div class="flex items-center justify-between rounded-xl bg-muted/45 p-3 text-xs">
              <span>合计</span>
              <strong>{{ formatMoneyWithCurrency(customerCart.total, customerCart.currency) }}</strong>
            </div>
            <p class="text-[11px] text-muted-foreground">价格口径：购物车项目保存的价格快照；当前上下文不代表已锁定库存。</p>
            <p v-if="!customerCart.items?.length" class="rounded-xl border border-dashed p-3 text-xs text-muted-foreground">购物车为空</p>
            <article v-for="item in customerCart.items" :key="item.id" class="flex gap-2 rounded-xl border p-2">
              <div class="size-12 shrink-0 overflow-hidden rounded-lg bg-muted">
                <img v-if="item.image" :src="item.image" :alt="item.name" class="size-full object-cover" />
              </div>
              <div class="min-w-0 flex-1 text-xs">
                <p class="truncate font-bold">{{ item.name }}</p>
 <p class="mt-0.5 truncate text-[11px] text-muted-foreground">{{ item.sku || item.variant_name || '无 SKU'}}</p>
                <p class="mt-1 font-mono text-[11px]">{{ formatMoneyWithCurrency(item.price, item.currency || customerCart.currency) }} × {{ item.quantity }} = {{ formatMoneyWithCurrency(item.line_total, item.currency || customerCart.currency) }}</p>
                <p class="mt-1 text-[11px] text-muted-foreground">库存快照：{{ item.inventory_snapshot === 'not_captured' ? '未采集' : (item.inventory_snapshot || '未知') }}</p>
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
              {{ factStatusLabel(customerWishlist.status, customerWishlist.available) }}
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
              {{ factStatusLabel(customerOrders.status, customerOrders.available) }}
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
                <span class="font-mono">{{ formatMoney(item.total_amount) }} {{ item.currency || '' }}</span>
              </div>
              <p class="mt-1 text-[11px] text-muted-foreground">
                {{ item.status }} / {{ item.payment_status }} / {{ item.shipping_status }} · {{ formatShortDate(item.created_at) }}
              </p>
              <p class="mt-1 text-[11px] text-muted-foreground">
                小计 {{ formatMoney(item.subtotal_amount) }} · 折扣后 {{ formatMoney(item.discounted_subtotal_amount) }} · 运费 {{ formatMoney(item.shipping_fee) }} · 税费 {{ formatMoney(item.tax_amount) }}
              </p>
              <div v-if="item.shipments?.length" class="mt-2 space-y-1">
                <div v-for="shipment in item.shipments" :key="shipment.id" class="flex items-center justify-between gap-2 rounded-lg bg-muted/45 px-2 py-1.5 text-[11px]">
                  <span class="truncate">{{ shipment.carrier || '承运商未记录' }}<span v-if="shipment.carrier_service"> · {{ shipment.carrier_service }}</span></span>
                  <button type="button" class="shrink-0 font-mono underline underline-offset-2" :title="shipment.tracking_number ? '点击复制追踪号' : undefined" @click="copyValue(shipment.tracking_number)">{{ shipment.tracking_number || '无追踪号' }}</button>
                </div>
              </div>
              <div v-if="item.items?.length" class="mt-2 space-y-1">
                <div v-for="line in item.items" :key="line.id" class="flex items-center gap-2 text-[11px]">
                  <img v-if="line.thumbnail" :src="line.thumbnail" :alt="line.name" class="size-7 rounded object-cover" />
                  <span class="min-w-0 flex-1 truncate">{{ line.name || `产品 ${line.product_id}` }} × {{ line.quantity }}</span>
                  <span class="font-mono">{{ formatMoney(line.total_amount) }}</span>
                </div>
              </div>
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
              {{ factStatusLabel(customerBrowsing.status, customerBrowsing.available) }}
            </AdminStatusBadge>
          </div>
          <p v-if="!customerBrowsing.available" class="rounded-xl bg-muted/45 p-3 text-xs leading-6 text-muted-foreground">
            {{ customerBrowsing.reason }}
          </p>
          <div v-else class="space-y-2">
            <p v-if="!customerBrowsing.items?.length" class="rounded-xl border border-dashed p-3 text-xs text-muted-foreground">暂无浏览历史</p>
            <article v-for="item in customerBrowsing.items" :key="item.product_id" class="rounded-xl border p-2 text-xs">
              <div class="flex items-center justify-between gap-2">
                <div class="flex min-w-0 items-center gap-2">
                  <img v-if="item.thumbnail" :src="item.thumbnail" :alt="item.name || `产品 ${item.product_id}`" class="size-8 shrink-0 rounded object-cover" />
                  <strong class="truncate">{{ item.name || `产品 ${item.product_id}` }}</strong>
                </div>
                <span class="shrink-0 font-mono">{{ item.view_count }} 次</span>
              </div>
              <p v-if="item.sku || item.price" class="mt-1 text-[11px] text-muted-foreground">
                <span v-if="item.sku">SKU {{ item.sku }}</span>
                <span v-if="item.sku && item.price"> · </span>
                <span v-if="item.price">{{ item.currency || '' }} {{ item.price }}</span>
              </p>
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
import { Clock3, Heart, History, Info, LoaderCircle, Mail, MapPin, PackageCheck, RotateCcw, ShieldAlert, ShoppingCart, UserCheck, UserRound } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  formatDate,
  formatCustomerLocalDate,
  formatCustomerLocalTime,
  formatMoney,
  formatMoneyWithCurrency,
  formatShortDate,
  customerLocalTimeHint as getCustomerLocalTimeHint,
  customerLocalTimePhase as getCustomerLocalTimePhase,
  customerTimezoneDifference as getCustomerTimezoneDifference,
  customerTimezoneSourceLabel as getCustomerTimezoneSourceLabel,
  isValidCustomerTimezone,
  signalTone,
  statusTone,
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
  CustomerAfterSales,
  CustomerOrders,
  CustomerPaymentDisputes,
  CustomerShippingAddress,
  CustomerSignal,
  CustomerWishlist,
} from '@/modules/customer-service/customerServiceTypes'

const props = withDefaults(defineProps<{
  selectedConversation?: CustomerConversation | null
  customerContext?: CustomerContext | null
  loading?: boolean
  contextError?: string | null
  contextLastUpdatedAt?: Date | null
}>(), {
  selectedConversation: null,
  customerContext: null,
  loading: false,
  contextError: null,
  contextLastUpdatedAt: null,
})

const customerAccount = computed<CustomerAccount | null>(() => props.customerContext?.customer?.account || null)
const customerAnonymous = computed<CustomerAnonymous | null>(() => props.customerContext?.customer?.anonymous || null)
const customerContact = computed<CustomerContact>(() => props.customerContext?.contact || {})
const customerClockNow = ref(new Date())
const customerCart = computed<CustomerCart>(() => props.customerContext?.cart || { available: false, items: [] })
const customerWishlist = computed<CustomerWishlist>(() => props.customerContext?.wishlist || { available: false, items: [] })
const customerOrders = computed<CustomerOrders>(() => props.customerContext?.orders || { available: false, items: [] })
const customerShippingAddress = computed<CustomerShippingAddress>(() => props.customerContext?.shipping_address || { available: false, status: 'unavailable' })
const customerAfterSales = computed<CustomerAfterSales>(() => props.customerContext?.after_sales || { available: false, status: 'unavailable', items: [] })
const customerDisputes = computed<CustomerPaymentDisputes>(() => props.customerContext?.payment_disputes || { available: false, status: 'unavailable', items: [] })
const customerBrowsing = computed<CustomerBrowsing>(() => props.customerContext?.browsing || { available: false, items: [] })
const customerTimezoneValid = computed(() => isValidCustomerTimezone(customerContact.value.timezone))
const customerLocalTime = computed(() => formatCustomerLocalTime(customerClockNow.value, customerContact.value.timezone))
const customerLocalDate = computed(() => formatCustomerLocalDate(customerClockNow.value, customerContact.value.timezone))
const customerLocalTimePhase = computed(() => getCustomerLocalTimePhase(customerClockNow.value, customerContact.value.timezone))
const customerLocalTimeHint = computed(() => getCustomerLocalTimeHint(customerClockNow.value, customerContact.value.timezone))
const customerTimezoneSourceLabel = computed(() => getCustomerTimezoneSourceLabel(customerContact.value.timezone_source))
const customerTimezoneDifference = computed(() => getCustomerTimezoneDifference(customerClockNow.value, customerContact.value.timezone))
const copyValue = async (value?: string) => {
  const normalized = String(value || '').trim()
  if (!normalized || !navigator.clipboard) return
  try {
    await navigator.clipboard.writeText(normalized)
  } catch {
    // Clipboard permission is optional; the tracking number remains visible.
  }
}
const factStatusLabel = (status: string | undefined, available: boolean): string => {
  if (status === 'error') return '读取失败'
  if (status === 'permission_denied') return '无权限'
  if (available || status === 'available') return '可用'
  return '暂无数据'
}
const factTone = (status: string | undefined, available: boolean): 'green' | 'amber' | 'coral' | 'gray' => {
  if (status === 'error') return 'coral'
  if (status === 'permission_denied' || !available) return 'amber'
  return 'green'
}
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

