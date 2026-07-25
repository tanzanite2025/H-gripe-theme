<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="客服对话"
      description="处理网页 Public Chat 会话；客户侧只创建和读取自己的对话，客服侧在这里回复"
    >
      <template #actions>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" :disabled="loading" @click="refreshInbox">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <AdminFilterPanel>
      <form class="grid gap-3 lg:grid-cols-[minmax(220px,1.2fr)_140px_140px_180px_130px_auto_auto]" @submit.prevent="applyFilters">
        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
            <Input v-model="filters.search" class="h-9 pl-9" placeholder="客户、邮箱、会话 ID、消息内容" />
          </div>
        </label>

        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">STATUS / 状态</span>
          <Select v-model="filters.status">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部</SelectItem>
              <SelectItem value="pending">待处理</SelectItem>
              <SelectItem value="active">进行中</SelectItem>
              <SelectItem value="closed">已关闭</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">IDENTITY / 身份</span>
          <Select v-model="filters.identity">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部客户</SelectItem>
              <SelectItem value="account">会员</SelectItem>
              <SelectItem value="anonymous">匿名访客</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">ASSIGNEE / 负责人</span>
          <Select v-model="filters.assignedTo">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部客服</SelectItem>
              <SelectItem v-for="agent in assignableAgents" :key="agent.user_id || agent.id" :value="String(agent.user_id || agent.id)">
                {{ agent.name || agent.email || `用户 ${agent.user_id || agent.id}` }}
              </SelectItem>
            </SelectContent>
          </Select>
        </label>

        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">UNREAD / 未读</span>
          <Select v-model="filters.unread">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部</SelectItem>
              <SelectItem value="unread">只看未读</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <label class="space-y-1 block">
          <span class="block text-[10px] font-black uppercase tracking-widest text-transparent select-none">QUERY</span>
          <Button type="submit" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" :disabled="loading">
            <Search class="size-3.5" />
            查询
          </Button>
        </label>

        <label class="space-y-1 block">
          <span class="block text-[10px] font-black uppercase tracking-widest text-transparent select-none">ACTION</span>
          <Button type="button" variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="resetFilters">
            <RotateCcw class="size-3.5" />
            重置
          </Button>
        </label>
      </form>
    </AdminFilterPanel>

    <section class="grid min-h-[calc(100dvh-280px)] gap-4 xl:grid-cols-[320px_minmax(0,1fr)_320px] 2xl:grid-cols-[360px_minmax(0,1fr)_360px]">
      <Card class="min-h-[420px] overflow-hidden py-0">
        <CardHeader class="border-b bg-muted/30 px-4 py-3">
          <CardTitle>会话列表</CardTitle>
          <CardDescription>按客户会话隔离，每个卡片对应一个 conversation</CardDescription>
        </CardHeader>

        <div class="min-h-0 flex-1 overflow-y-auto p-3">
          <div v-if="loading" class="flex h-52 items-center justify-center text-muted-foreground">
            <LoaderCircle class="size-5 animate-spin" />
          </div>
          <div v-else-if="filteredConversations.length === 0" class="flex h-52 flex-col items-center justify-center text-muted-foreground">
            <MessageCircleOff class="mb-2 size-7 opacity-55" />
            <span class="text-xs">暂无客服对话</span>
          </div>
          <div v-else class="space-y-2">
            <button
              v-for="conversation in filteredConversations"
              :key="conversation.id"
              type="button"
              class="group w-full rounded-2xl border border-dashed p-3 text-left transition-all hover:border-primary/45 hover:bg-muted/45"
              :class="selectedConversation?.id === conversation.id ? 'border-primary/60 bg-primary/5 shadow-[0_8px_28px_rgba(15,23,42,0.08)]' : 'border-border/80 bg-card'"
              @click="selectConversation(conversation)"
            >
              <div class="flex items-start gap-3">
                <div class="flex size-10 shrink-0 items-center justify-center rounded-full bg-primary/10 font-black text-primary">
                  {{ initials(conversation.customer_name) }}
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center justify-between gap-2">
                    <strong class="truncate text-xs font-black">{{ conversation.customer_name || '匿名客户' }}</strong>
                    <AdminStatusBadge :tone="statusTone(conversation.display_status)">
                      {{ statusLabel(conversation.display_status) }}
                    </AdminStatusBadge>
                  </div>
                  <p class="mt-1 line-clamp-2 text-xs leading-5 text-muted-foreground">
                    {{ conversation.last_message || '暂无消息' }}
                  </p>
                  <div class="mt-2 flex flex-wrap items-center gap-2 text-[10px] font-bold uppercase tracking-wider text-muted-foreground/70">
                    <span>#{{ conversation.ticket_number || conversation.id }}</span>
                    <span>·</span>
                    <span>{{ assigneeName(conversation.assigned_to) }}</span>
                    <span v-if="conversation.unread_count > 0" class="rounded-full bg-rose-500/10 px-2 py-0.5 text-rose-600">
                      {{ conversation.unread_count }} 未读
                    </span>
                  </div>
                </div>
              </div>
            </button>
          </div>
        </div>

        <CardFooter class="justify-between gap-3 text-xs text-muted-foreground">
          <span>共 {{ pagination.total }} 条</span>
          <div class="flex items-center gap-2">
            <Button variant="outline" size="sm" :disabled="pagination.page <= 1 || loading" @click="changePage(pagination.page - 1)">上一页</Button>
            <span class="font-mono text-[10px]">{{ pagination.page }} / {{ totalPages }}</span>
            <Button variant="outline" size="sm" :disabled="pagination.page >= totalPages || loading" @click="changePage(pagination.page + 1)">下一页</Button>
          </div>
        </CardFooter>
      </Card>

      <Card class="min-h-[520px] overflow-hidden py-0">
        <template v-if="selectedConversation">
          <CardHeader class="border-b bg-muted/30 px-4 py-3">
            <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
              <div class="min-w-0">
                <CardTitle class="truncate">{{ selectedConversation.customer_name || '匿名客户' }}</CardTitle>
                <CardDescription class="break-all">
                  {{ selectedConversation.conversation_id || selectedConversation.ticket_number || selectedConversation.id }}
                </CardDescription>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <AdminStatusBadge :tone="statusTone(selectedConversation.display_status)">
                  {{ statusLabel(selectedConversation.display_status) }}
                </AdminStatusBadge>
                <AdminStatusBadge v-if="selectedConversation.visitor_anonymous" tone="amber">匿名</AdminStatusBadge>
                <AdminStatusBadge v-else tone="green">会员</AdminStatusBadge>
              </div>
            </div>
          </CardHeader>

          <CardContent class="grid min-h-0 flex-1 grid-rows-[auto_minmax(0,1fr)_auto] gap-0 p-0">
            <div class="flex flex-wrap items-center justify-between gap-3 border-b px-4 py-3">
              <div class="text-xs text-muted-foreground">
                当前负责人：
                <strong class="text-foreground">{{ assigneeName(selectedConversation.assigned_to) }}</strong>
              </div>

              <div v-if="hasPermission('ticket:edit')" class="flex flex-wrap items-center gap-2">
                <Select v-model="transferTo">
                  <SelectTrigger class="h-9 w-44"><SelectValue placeholder="选择客服" /></SelectTrigger>
                  <SelectContent>
                    <SelectItem v-for="agent in assignableAgents" :key="agent.user_id || agent.id" :value="String(agent.user_id || agent.id)">
                      {{ agent.name || agent.email || `用户 ${agent.user_id || agent.id}` }}
                    </SelectItem>
                  </SelectContent>
                </Select>
                <Button variant="outline" size="sm" class="rounded-full" :disabled="transferring || !transferTo" @click="transferConversation">
                  <ArrowRightLeft v-if="!transferring" class="size-3.5" />
                  <LoaderCircle v-else class="size-3.5 animate-spin" />
                  转接
                </Button>
              </div>
            </div>

            <div class="relative min-h-0 overflow-y-auto px-4 py-4">
              <div v-if="messagesLoading" class="absolute inset-0 z-10 flex items-center justify-center bg-card/75">
                <LoaderCircle class="size-5 animate-spin text-primary" />
              </div>
              <div v-else-if="messages.length === 0" class="flex h-72 flex-col items-center justify-center text-muted-foreground">
                <MessageCircleOff class="mb-2 size-7 opacity-55" />
                <span class="text-xs">暂无消息记录</span>
              </div>
              <div v-else class="space-y-3">
                <article
                  v-for="message in messages"
                  :key="message.id"
                  class="max-w-[86%] rounded-2xl border px-4 py-3 text-sm"
                  :class="message.is_agent ? 'ml-auto border-blue-200 bg-blue-50/75' : 'mr-auto border-border bg-muted/45'"
                >
                  <header class="flex flex-wrap items-center justify-between gap-x-4 gap-y-1">
                    <div class="flex items-center gap-2">
                      <span class="text-xs font-black">{{ message.sender_name || (message.is_agent ? '客服' : '客户') }}</span>
                      <AdminStatusBadge :tone="message.is_agent ? 'blue' : 'gray'">
                        {{ message.is_agent ? '客服' : '客户' }}
                      </AdminStatusBadge>
                    </div>
                    <time class="text-[11px] text-muted-foreground">{{ formatDate(message.created_at) }}</time>
                  </header>
                  <p class="mt-2 whitespace-pre-wrap break-words leading-6">{{ message.content || message.message }}</p>
                  <a
                    v-if="message.attachment_url"
                    class="mt-2 inline-flex text-xs font-bold text-primary underline-offset-4 hover:underline"
                    :href="message.attachment_url"
                    target="_blank"
                    rel="noreferrer"
                  >
                    查看附件
                  </a>
                </article>
              </div>
            </div>

            <form v-if="hasPermission('ticket:edit')" class="border-t p-4" @submit.prevent="sendReply">
              <Textarea v-model="replyMessage" class="min-h-24 resize-none" placeholder="输入回复内容，发送后客户侧可在原会话中看到" />
              <div class="mt-3 flex justify-end">
                <Button type="submit" class="rounded-full" :disabled="replying || !replyMessage.trim()">
                  <LoaderCircle v-if="replying" class="size-4 animate-spin" />
                  <Send v-else class="size-4" />
                  发送回复
                </Button>
              </div>
            </form>
          </CardContent>
        </template>

        <div v-else class="flex h-full min-h-[520px] flex-col items-center justify-center p-8 text-center text-muted-foreground">
          <Headset class="mb-3 size-10 opacity-50" />
          <h2 class="text-sm font-black text-foreground">选择一个客户会话</h2>
          <p class="mt-2 max-w-sm text-xs leading-6">
            左侧每个卡片是一条独立 Public Chat 会话。客户之间不会共用消息窗口。
          </p>
        </div>
      </Card>

      <Card class="min-h-[520px] overflow-hidden py-0">
        <CardHeader class="border-b bg-muted/30 px-4 py-3">
          <CardTitle class="flex items-center gap-2">
            <UserRound class="size-4 text-primary" />
            客户上下文
          </CardTitle>
          <CardDescription>只读事实源：账号、购物车、心愿单、订单和浏览记录</CardDescription>
        </CardHeader>

        <CardContent class="min-h-0 space-y-4 overflow-y-auto p-4">
          <div v-if="!selectedConversation" class="flex min-h-[420px] flex-col items-center justify-center text-center text-muted-foreground">
            <Info class="mb-2 size-7 opacity-55" />
            <p class="text-xs leading-6">选择会话后显示客户上下文。</p>
          </div>

          <div v-else-if="contextLoading" class="flex min-h-[420px] items-center justify-center text-muted-foreground">
            <LoaderCircle class="size-5 animate-spin" />
          </div>

          <div v-else-if="!customerContext" class="rounded-2xl border border-dashed p-4 text-xs leading-6 text-muted-foreground">
            暂时无法读取客户上下文。消息仍可正常收发。
          </div>

          <template v-else>
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
                  <p class="font-black text-foreground">{{ customerAccount.display_name || customerAccount.username || customerAccount.email }}</p>
                  <p class="mt-1 break-all text-muted-foreground">{{ customerAccount.email || '未填写邮箱' }}</p>
                </div>
                <dl class="grid grid-cols-2 gap-2 text-[11px]">
                  <div class="rounded-xl border p-2">
                    <dt class="text-muted-foreground">账号 ID</dt>
                    <dd class="mt-1 font-mono font-bold">{{ customerAccount.id }}</dd>
                  </div>
                  <div class="rounded-xl border p-2">
                    <dt class="text-muted-foreground">语言</dt>
                    <dd class="mt-1 font-mono font-bold">{{ customerAccount.locale || '-' }}</dd>
                  </div>
                  <div class="rounded-xl border p-2">
                    <dt class="text-muted-foreground">状态</dt>
                    <dd class="mt-1 font-mono font-bold">{{ customerAccount.status || '-' }}</dd>
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
                  <span class="mt-1 block font-mono text-foreground">{{ customerAnonymous?.visitor_hash_preview || '未绑定' }}</span>
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
                    <p class="break-all font-bold text-foreground">{{ customerContact.email || '未采集邮箱' }}</p>
                    <p class="text-[11px] text-muted-foreground">来源：{{ customerContact.email_source || 'not_captured' }}</p>
                  </div>
                </div>
                <div class="flex items-start gap-2 rounded-xl border p-2">
                  <Info class="mt-0.5 size-3.5 text-muted-foreground" />
                  <div class="min-w-0">
                    <p class="font-bold text-foreground">{{ customerContact.locale || '未采集语言' }}</p>
                    <p class="text-[11px] text-muted-foreground">来源：{{ customerContact.locale_source || 'not_captured' }}</p>
                  </div>
                </div>
                <div class="flex items-start gap-2 rounded-xl border p-2">
                  <MapPin class="mt-0.5 size-3.5 text-muted-foreground" />
                  <div class="min-w-0">
                    <p class="font-bold text-foreground">{{ signalItems[0]?.value || '未采集地区' }}</p>
                    <p class="text-[11px] text-muted-foreground">{{ signalItems[0]?.reason || '需要 visitor profile / GeoIP 层' }}</p>
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
                    <p class="mt-0.5 truncate text-[11px] text-muted-foreground">{{ item.sku || item.variant_name || '无 SKU' }}</p>
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
    </section>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  ArrowRightLeft,
  Clock3,
  Heart,
  Headset,
  History,
  Info,
  LoaderCircle,
  Mail,
  MapPin,
  MessageCircleOff,
  MessagesSquare,
  PackageCheck,
  RefreshCw,
  RotateCcw,
  Search,
  Send,
  ShoppingCart,
  UserCheck,
  UserRound
} from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()
const loading = ref(false)
const messagesLoading = ref(false)
const contextLoading = ref(false)
const replying = ref(false)
const transferring = ref(false)
const conversations = ref([])
const messages = ref([])
const customerContext = ref(null)
const selectedConversation = ref(null)
const replyMessage = ref('')
const transferTo = ref('')
const assignableAgents = ref([])
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const filters = reactive({ search: '', status: 'all', identity: 'all', assignedTo: 'all', unread: 'all' })
const realtimeSource = ref(null)
let realtimeReconnectTimer = null
let realtimeRefreshTimer = null

const apiData = (response) => response.data?.data ?? response.data ?? {}
const hasPermission = (permission) => authStore.hasPermission(permission)

const totalPages = computed(() => Math.max(1, Math.ceil((pagination.total || 0) / pagination.pageSize)))

const filteredConversations = computed(() => conversations.value)

const statItems = computed(() => {
  const total = conversations.value.length
  const unread = conversations.value.reduce((sum, item) => sum + Number(item.unread_count || 0), 0)
  const active = conversations.value.filter((item) => statusDisplayValue(item.display_status || item.status) === 'active').length
  const closed = conversations.value.filter((item) => statusDisplayValue(item.display_status || item.status) === 'closed').length

  return [
    { key: 'total', label: '当前页会话', value: total, icon: MessagesSquare, tone: 'gray' },
    { key: 'unread', label: '未读消息', value: unread, icon: Clock3, tone: unread > 0 ? 'coral' : 'gray' },
    { key: 'active', label: '进行中', value: active, icon: Headset, tone: 'blue' },
    { key: 'closed', label: '已关闭', value: closed, icon: UserCheck, tone: 'green' }
  ]
})

const customerAccount = computed(() => customerContext.value?.customer?.account || null)
const customerAnonymous = computed(() => customerContext.value?.customer?.anonymous || null)
const customerContact = computed(() => customerContext.value?.contact || {})
const customerCart = computed(() => customerContext.value?.cart || { available: false, items: [] })
const customerWishlist = computed(() => customerContext.value?.wishlist || { available: false, items: [] })
const customerOrders = computed(() => customerContext.value?.orders || { available: false, items: [] })
const customerBrowsing = computed(() => customerContext.value?.browsing || { available: false, items: [] })
const signalItems = computed(() => {
  const signals = customerContext.value?.signals || {}
  return [
    { key: 'region', label: '地区', ...(signals.region || {}) },
    { key: 'cart_session', label: '购物车会话', ...(signals.cart_session || {}) },
    { key: 'email_capture', label: '邮箱采集', ...(signals.email_capture || {}) },
    { key: 'visitor_profile', label: '访客档案', ...(signals.visitor_profile || {}) }
  ]
})

const fetchConversations = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/customer-service/conversations', {
      params: {
        page: pagination.page,
        page_size: pagination.pageSize,
        search: filters.search.trim() || undefined,
        status: filters.status !== 'all' ? filters.status : undefined,
        identity: filters.identity !== 'all' ? filters.identity : undefined,
        assigned_to: filters.assignedTo !== 'all' ? filters.assignedTo : undefined,
        unread: filters.unread === 'unread' ? 'true' : undefined
      }
    })
    const data = apiData(response)
    conversations.value = data.conversations || []
    pagination.total = data.pagination?.total ?? conversations.value.length

    if (selectedConversation.value) {
      const refreshed = conversations.value.find((item) => Number(item.id) === Number(selectedConversation.value.id))
      if (refreshed) {
        selectedConversation.value = refreshed
      } else {
        selectedConversation.value = null
        messages.value = []
        customerContext.value = null
      }
    }
  } catch (error) {
    console.error('Failed to fetch customer-service conversations:', error)
  } finally {
    loading.value = false
  }
}

const fetchContext = async (conversationID) => {
  if (!conversationID) {
    customerContext.value = null
    return
  }

  contextLoading.value = true
  try {
    const response = await axios.get(`/api/admin/customer-service/conversations/${conversationID}/context`)
    const data = apiData(response)
    customerContext.value = data.context || null
  } catch (error) {
    console.error('Failed to fetch customer-service context:', error)
    customerContext.value = null
  } finally {
    contextLoading.value = false
  }
}

const fetchAgents = async () => {
  try {
    const response = await axios.get('/api/admin/customer-service/agents')
    const data = apiData(response)
    assignableAgents.value = data.agents || []
  } catch (error) {
    console.error('Failed to fetch public chat agents:', error)
    assignableAgents.value = []
  }
}

const refreshInbox = async () => {
  await Promise.all([fetchConversations(), fetchAgents()])
  if (selectedConversation.value) {
    await Promise.all([
      fetchMessages(selectedConversation.value.id),
      fetchContext(selectedConversation.value.id)
    ])
  }
}

const buildAdminCustomerServiceEventURL = () => {
  const query = new URLSearchParams({ scope: 'inbox' })
  const baseURL = String(axios.defaults?.baseURL || '').replace(/\/$/, '')
  const path = `/api/admin/customer-service/events?${query.toString()}`
  return baseURL ? `${baseURL}${path}` : path
}

const connectCustomerServiceRealtime = () => {
  if (typeof window === 'undefined' || !('EventSource' in window)) return

  closeCustomerServiceRealtimeSource()
  const source = new EventSource(buildAdminCustomerServiceEventURL(), { withCredentials: true })
  realtimeSource.value = source

  ;[
    'conversation.message.created',
    'conversation.messages.read',
    'conversation.assigned',
    'conversation.status.changed',
    'conversation.context.updated'
  ].forEach((eventType) => {
    source.addEventListener(eventType, handleCustomerServiceRealtimeEvent)
  })

  source.onerror = () => {
    if (realtimeSource.value === source) {
      realtimeSource.value = null
    }
    source.close()
    if (!realtimeReconnectTimer) {
      realtimeReconnectTimer = window.setTimeout(() => {
        realtimeReconnectTimer = null
        connectCustomerServiceRealtime()
      }, 5000)
    }
  }
}

const closeCustomerServiceRealtimeSource = () => {
  if (realtimeSource.value) {
    realtimeSource.value.close()
    realtimeSource.value = null
  }
}

const closeCustomerServiceRealtime = () => {
  closeCustomerServiceRealtimeSource()
  if (realtimeReconnectTimer) {
    window.clearTimeout(realtimeReconnectTimer)
    realtimeReconnectTimer = null
  }
  if (realtimeRefreshTimer) {
    window.clearTimeout(realtimeRefreshTimer)
    realtimeRefreshTimer = null
  }
}

const handleCustomerServiceRealtimeEvent = (event) => {
  try {
    scheduleCustomerServiceRealtimeRefresh(JSON.parse(event.data || '{}'))
  } catch (error) {
    console.warn('Invalid customer-service realtime event:', error)
  }
}

const scheduleCustomerServiceRealtimeRefresh = (event) => {
  if (realtimeRefreshTimer) {
    window.clearTimeout(realtimeRefreshTimer)
  }

  realtimeRefreshTimer = window.setTimeout(async () => {
    realtimeRefreshTimer = null
    await fetchConversations()

    if (!selectedConversation.value || Number(event.ticket_id) !== Number(selectedConversation.value.id)) {
      return
    }

    if (event.type === 'conversation.message.created') {
      await fetchMessages(selectedConversation.value.id)
    }
    if (['conversation.context.updated', 'conversation.assigned'].includes(event.type)) {
      await fetchContext(selectedConversation.value.id)
    }
  }, 350)
}

const fetchMessages = async (conversationID) => {
  if (!conversationID) return
  messagesLoading.value = true
  try {
    const response = await axios.get(`/api/admin/customer-service/conversations/${conversationID}/messages`)
    const data = apiData(response)
    messages.value = data.messages || []
    await axios.post(`/api/admin/customer-service/conversations/${conversationID}/messages/mark-read`)
  } catch (error) {
    console.error('Failed to fetch customer-service messages:', error)
    messages.value = []
  } finally {
    messagesLoading.value = false
  }
}

const selectConversation = async (conversation) => {
  selectedConversation.value = conversation
  replyMessage.value = ''
  transferTo.value = conversation.assigned_to ? String(conversation.assigned_to) : ''
  await Promise.all([
    fetchMessages(conversation.id),
    fetchContext(conversation.id)
  ])
  await fetchConversations()
}

const sendReply = async () => {
  if (!selectedConversation.value || !replyMessage.value.trim()) return
  const message = replyMessage.value.trim()
  replying.value = true
  try {
    await axios.post(`/api/admin/customer-service/conversations/${selectedConversation.value.id}/messages`, { message })
    replyMessage.value = ''
    toast.success('回复已发送')
    await Promise.all([
      fetchMessages(selectedConversation.value.id),
      fetchConversations(),
      fetchContext(selectedConversation.value.id)
    ])
  } catch (error) {
    console.error('Failed to send customer-service reply:', error)
  } finally {
    replying.value = false
  }
}

const transferConversation = async () => {
  if (!selectedConversation.value || !transferTo.value) return
  transferring.value = true
  try {
    await axios.patch(`/api/admin/customer-service/conversations/${selectedConversation.value.id}/transfer`, {
      assigned_to: Number(transferTo.value)
    })
    toast.success('会话已转接')
    await Promise.all([
      fetchConversations(),
      fetchContext(selectedConversation.value.id)
    ])
  } catch (error) {
    console.error('Failed to transfer customer-service conversation:', error)
  } finally {
    transferring.value = false
  }
}

const changePage = async (page) => {
  pagination.page = Math.max(1, Math.min(page, totalPages.value))
  await fetchConversations()
}

const applyFilters = async () => {
  pagination.page = 1
  await fetchConversations()
}

const resetFilters = async () => {
  filters.search = ''
  filters.status = 'all'
  filters.identity = 'all'
  filters.assignedTo = 'all'
  filters.unread = 'all'
  pagination.page = 1
  await fetchConversations()
}

const initials = (name) => String(name || '?').trim().slice(0, 2).toUpperCase()
const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
const formatShortDate = (dateString) => dateString ? new Date(dateString).toLocaleDateString('zh-CN') : '-'
const formatMoney = (value) => `$${Number(value || 0).toFixed(2)}`

const statusDisplayValue = (status) => {
  if (['resolved', 'closed'].includes(status)) return 'closed'
  if (['in_progress', 'active'].includes(status)) return 'active'
  if (['open', 'pending'].includes(status)) return 'pending'
  return status || 'pending'
}

const statusLabel = (status) => ({
  active: '进行中',
  pending: '待处理',
  closed: '已关闭',
  open: '待处理',
  in_progress: '处理中',
  resolved: '已解决'
})[status] || status || '-'

const statusTone = (status) => ({
  active: 'blue',
  pending: 'amber',
  closed: 'gray',
  open: 'amber',
  in_progress: 'blue',
  resolved: 'green'
})[status] || 'gray'

const signalTone = (status) => {
  if (['captured', 'verified', 'bound', 'created', 'linked'].includes(status)) return 'green'
  if (['missing', 'not_captured', 'not_linked', 'not_created', 'missing_user'].includes(status)) return 'amber'
  return 'gray'
}

const assigneeName = (assignedTo) => {
  if (!assignedTo) return '未分配'
  const agent = assignableAgents.value.find((item) => Number(item.user_id || item.id) === Number(assignedTo))
  return agent?.name || agent?.email || `用户 ${assignedTo}`
}

onMounted(async () => {
  await refreshInbox()
  connectCustomerServiceRealtime()
})

onBeforeUnmount(closeCustomerServiceRealtime)
</script>
