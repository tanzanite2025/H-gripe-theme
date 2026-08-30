<template>
  <Card class="h-full min-h-0 overflow-hidden py-0">
    <template v-if="selectedConversation">
      <CardHeader class="shrink-0 border-b bg-muted/30 px-4 py-3">
        <div class="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
          <div class="min-w-0">
            <div class="flex min-w-0 items-center gap-2">
              <CardTitle class="truncate">{{ selectedConversation.customer_name || '匿名客户' }}</CardTitle>
              <Button
                type="button"
                variant="outline"
                size="icon-sm"
                class="shrink-0"
                aria-label="查看客户上下文"
                @click="emit('open-context')"
              >
                <Info class="size-3.5" />
              </Button>
            </div>
            <CardDescription class="break-all">
              {{ selectedConversation.conversation_id || selectedConversation.ticket_number || selectedConversation.id }}
            </CardDescription>
          </div>

          <div v-if="canEdit" class="flex min-w-0 flex-col gap-1 lg:items-end">
            <div class="flex min-w-0 flex-wrap items-center justify-end gap-2">
              <span class="shrink-0 text-xs text-muted-foreground">当前接待客服：</span>
              <strong class="min-w-0 truncate text-xs font-black text-foreground">
                {{ assigneeName(selectedConversation.assigned_to, assignableAgents, authStore.user) }}
              </strong>
              <Select v-model="transferModel" :disabled="!hasAssignableAgents">
                <SelectTrigger class="h-8 w-40">
                  <SelectValue :placeholder="hasAssignableAgents ? '选择客服' : '暂无可转接客服'" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem v-for="agent in assignableAgents" :key="agent.user_id || agent.id" :value="String(agent.user_id || agent.id)">
                    {{ agentDisplayName(agent) }}
                  </SelectItem>
                </SelectContent>
              </Select>
              <Button
                variant="outline"
                size="sm"
                class="rounded-full"
                :disabled="transferring || !transferModel || !hasAssignableAgents"
                @click="emit('transfer')"
              >
                <ArrowRightLeft v-if="!transferring" class="size-3.5" />
                <LoaderCircle v-else class="size-3.5 animate-spin" />
                转接
              </Button>
            </div>
          </div>
          <div v-else class="text-xs text-muted-foreground lg:text-right">
            当前接待客服：
            <strong class="text-foreground">{{ assigneeName(selectedConversation.assigned_to, assignableAgents, authStore.user) }}</strong>
          </div>
        </div>
      </CardHeader>

      <CardContent class="grid min-h-0 flex-1 grid-rows-[minmax(0,1fr)_auto] gap-0 p-0">
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
 :class="message.is_agent ? 'ml-auto border-blue-200 bg-blue-50/75': 'mr-auto border-border bg-muted/45'"
            >
              <header class="flex flex-wrap items-center justify-between gap-x-4 gap-y-1">
                <div class="flex items-center gap-2">
 <span class="text-xs font-black">{{ message.sender_name || (message.is_agent ? '客服': '客户') }}</span>
                  <AdminStatusBadge :tone="message.is_agent ? 'blue' : 'gray'">
                    {{ message.is_agent ? '客服' : '客户' }}
                  </AdminStatusBadge>
                </div>
                <time class="text-[11px] text-muted-foreground">{{ formatDate(message.created_at) }}</time>
              </header>

              <div
                v-if="isConfigConfirmMessage(message)"
                class="mt-3 rounded-2xl border border-blue-200 bg-blue-50/70 p-3"
              >
                <div class="mb-2 text-[11px] font-black uppercase tracking-widest text-blue-600">
                  配置确认请求
                </div>
                <div class="flex gap-3">
                  <div class="size-16 shrink-0 overflow-hidden rounded-xl bg-muted">
                    <img
                      v-if="configProduct(message).thumbnail"
                      :src="configProduct(message).thumbnail"
                      :alt="configProduct(message).title || 'Product'"
                      class="size-full object-cover"
                    />
                  </div>
                  <div class="min-w-0 flex-1">
                    <a
                      v-if="configProduct(message).url"
                      :href="configProduct(message).url"
                      target="_blank"
                      rel="noreferrer"
                      class="block truncate text-sm font-black text-foreground underline-offset-4 hover:underline"
                    >
                      {{ configProduct(message).title || message.content || message.message }}
                    </a>
                    <p v-else class="truncate text-sm font-black text-foreground">
                      {{ configProduct(message).title || message.content || message.message }}
                    </p>
                    <p v-if="configProduct(message).price" class="mt-1 text-xs font-bold text-emerald-600">
                      {{ configProduct(message).price }}
                    </p>
                    <p v-if="configProduct(message).sku" class="mt-1 text-[11px] text-muted-foreground">
                      SKU：{{ configProduct(message).sku }}
                    </p>
                  </div>
                </div>
                <div class="mt-3 rounded-xl bg-white/70 px-3 py-2 text-xs leading-5 text-muted-foreground">
                  <p v-if="configSelection(message).variant_title" class="font-bold text-foreground">
                    已选：{{ configSelection(message).variant_title }}
                  </p>
                  <div v-if="configOptionRows(message).length" class="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-2">
                    <div
                      v-for="option in configOptionRows(message)"
                      :key="option.key"
                      class="rounded-xl border border-blue-100 bg-blue-50/50 px-2.5 py-2"
                    >
                      <span class="block text-[10px] font-black uppercase tracking-wider text-blue-500">
                        {{ option.label }}
                      </span>
                      <span class="mt-0.5 block font-bold text-foreground">
                        {{ option.value }}<span v-if="option.unit"> {{ option.unit }}</span>
                      </span>
                    </div>
                  </div>
                  <p v-if="configSelection(message).weight_grams" class="mt-2">
                    重量：{{ configSelection(message).weight_grams }}g
                  </p>
                  <p
                    v-if="!configOptionRows(message).length && !configSelection(message).variant_title && !configSelection(message).weight_grams"
                  >
                    客户请求客服确认该产品配置。
                  </p>
                </div>
              </div>

              <div
                v-else-if="isOrderMessage(message)"
                class="mt-3 rounded-2xl border border-emerald-200 bg-emerald-50/70 p-3"
              >
                <div class="mb-2 text-[11px] font-black uppercase tracking-widest text-emerald-600">
                  订单确认请求
                </div>
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0 flex-1">
                    <a
                      v-if="orderPayload(message).url"
                      :href="orderPayload(message).url"
                      target="_blank"
                      rel="noreferrer"
                      class="block truncate text-sm font-black text-foreground underline-offset-4 hover:underline"
                    >
                      {{ orderPayload(message).title || message.content || message.message }}
                    </a>
                    <p v-else class="truncate text-sm font-black text-foreground">
                      {{ orderPayload(message).title || message.content || message.message }}
                    </p>
                    <p class="mt-1 text-xs font-bold text-emerald-700">
                      {{ formatOrderTotal(orderPayload(message)) }}
                    </p>
                  </div>
                  <AdminStatusBadge tone="green">
                    {{ orderPayload(message).status || 'order' }}
                  </AdminStatusBadge>
                </div>
                <div class="mt-2 flex flex-wrap gap-2 text-[11px] text-muted-foreground">
                  <span v-if="orderPayload(message).payment_status">支付：{{ orderPayload(message).payment_status }}</span>
                  <span v-if="orderPayload(message).shipping_status">物流：{{ orderPayload(message).shipping_status }}</span>
                  <span v-if="orderPayload(message).item_count">商品：{{ orderPayload(message).item_count }} 件</span>
                </div>
                <div v-if="orderItems(message).length" class="mt-3 space-y-2">
                  <article
                    v-for="item in orderItems(message).slice(0, 4)"
                    :key="item.id || `${item.product_id}-${item.sku}`"
                    class="rounded-xl border border-emerald-100 bg-white/70 px-3 py-2 text-xs"
                  >
                    <div class="flex items-center justify-between gap-2">
 <p class="truncate font-bold text-foreground">{{ item.title || item.product_name || 'Product'}}</p>
                      <span class="shrink-0 font-mono text-muted-foreground">x{{ item.quantity || 1 }}</span>
                    </div>
                    <p v-if="item.sku" class="mt-1 text-[11px] text-muted-foreground">SKU：{{ item.sku }}</p>
                  </article>
                  <p v-if="orderItems(message).length > 4" class="text-[11px] text-muted-foreground">
                    还有 {{ orderItems(message).length - 4 }} 个商品
                  </p>
                </div>
              </div>

              <div
                v-else-if="isProductMessage(message)"
                class="mt-3 rounded-2xl border border-violet-200 bg-violet-50/70 p-3"
              >
                <div class="mb-2 text-[11px] font-black uppercase tracking-widest text-violet-600">
                  产品链接
                </div>
                <div class="flex gap-3">
                  <div class="size-16 shrink-0 overflow-hidden rounded-xl bg-muted">
                    <img
                      v-if="productPayload(message).thumbnail"
                      :src="productPayload(message).thumbnail"
                      :alt="productPayload(message).title || 'Product'"
                      class="size-full object-cover"
                    />
                    <div v-else class="flex size-full items-center justify-center text-muted-foreground">
                      <Package class="size-5 opacity-50" />
                    </div>
                  </div>
                  <div class="min-w-0 flex-1">
                    <a
                      v-if="productPayload(message).url"
                      :href="productPayload(message).url"
                      target="_blank"
                      rel="noreferrer"
                      class="block truncate text-sm font-black text-foreground underline-offset-4 hover:underline"
                    >
                      {{ productPayload(message).title || message.content || message.message }}
                    </a>
                    <p v-else class="truncate text-sm font-black text-foreground">
                      {{ productPayload(message).title || message.content || message.message }}
                    </p>
                    <p v-if="formatProductPrice(productPayload(message))" class="mt-1 text-xs font-bold text-emerald-600">
                      {{ formatProductPrice(productPayload(message)) }}
                    </p>
                    <p v-if="productPayload(message).sku" class="mt-1 text-[11px] text-muted-foreground">
                      SKU：{{ productPayload(message).sku }}
                    </p>
                    <p v-if="productPayload(message).url" class="mt-1 truncate text-[11px] text-muted-foreground">
                      {{ productPayload(message).url }}
                    </p>
                  </div>
                </div>
              </div>

              <div
                v-else-if="isVideoMessage(message)"
                class="mt-3 rounded-2xl border border-slate-200 bg-slate-50/70 p-3"
              >
                <div class="mb-2 text-[11px] font-black uppercase tracking-widest text-slate-600">
                  视频
                </div>
                <div class="space-y-3">
                  <video
                    v-if="videoPayload(message).url"
                    class="max-h-72 w-full rounded-xl bg-black"
                    controls
                    playsinline
                    preload="metadata"
                    :poster="videoPayload(message).thumbnail || ''"
                  >
                    <source :src="videoPayload(message).url" />
                  </video>
                  <div class="min-w-0">
                    <p class="truncate text-sm font-black text-foreground">
                      {{ videoPayload(message).title || message.content || message.message }}
                    </p>
                    <p v-if="videoPayload(message).url" class="mt-1 truncate text-[11px] text-muted-foreground">
                      {{ videoPayload(message).url }}
                    </p>
                  </div>
                </div>
              </div>

              <p v-else class="mt-2 whitespace-pre-wrap break-words leading-6">{{ message.content || message.message }}</p>
 <div v-if="message.message_type === 'image'&& messageAttachments(message).length" class="mt-2 grid gap-2">
                <img
                  v-for="attachmentUrl in messageAttachments(message)"
                  :key="attachmentUrl"
                  :src="attachmentUrl"
                  alt="客服图片附件"
                  class="max-h-64 max-w-full rounded-xl object-contain"
                />
              </div>
              <a
                v-if="message.message_type === 'link' && message.metadata?.url"
                :href="message.metadata.url"
                target="_blank"
                rel="noreferrer"
                class="mt-2 block rounded-xl border border-emerald-200 bg-emerald-50/60 px-3 py-2 text-xs text-emerald-700 underline-offset-4 hover:underline"
              >
                <span class="block font-black">{{ message.metadata.title || message.content || message.message }}</span>
                <span class="mt-1 block truncate text-emerald-700/70">{{ message.metadata.url }}</span>
              </a>
              <a
                v-for="attachmentUrl in messageAttachments(message)"
                :key="`link-${attachmentUrl}`"
                class="mt-2 inline-flex text-xs font-bold text-primary underline-offset-4 hover:underline"
                :href="attachmentUrl"
                target="_blank"
                rel="noreferrer"
              >
                查看附件
              </a>
            </article>
          </div>
          <div
            v-if="selectedCustomerTyping?.active"
            class="mt-3 inline-flex items-center gap-1.5 rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1.5 text-xs font-bold text-emerald-700"
          >
            <span>{{ selectedCustomerTyping.displayName || '客户' }} 正在输入</span>
            <span class="flex gap-0.5">
              <span class="size-1 animate-pulse rounded-full bg-emerald-500"></span>
              <span class="size-1 animate-pulse rounded-full bg-emerald-500 [animation-delay:120ms]"></span>
              <span class="size-1 animate-pulse rounded-full bg-emerald-500 [animation-delay:240ms]"></span>
            </span>
          </div>
        </div>

        <form v-if="canEdit" class="border-t p-4" @submit.prevent="emit('send-reply')">
          <Textarea
            v-model="replyModel"
            class="min-h-24 resize-none"
            placeholder="输入回复内容，发送后客户侧可在原会话中看到"
            @input="emit('typing-input')"
          />
          <div class="mt-3 flex flex-wrap items-center gap-2">
            <div class="flex min-w-0 flex-1 flex-wrap items-center gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                class="rounded-full"
                :disabled="replying || attachmentUploading"
                @click="imagePickerOpen = true"
                title="从媒体库选择图片"
              >
                <ImagePlus class="size-3.5" />
                媒体库图片
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                class="rounded-full"
                :disabled="replying || attachmentUploading"
                @click="openLocalImagePicker"
                title="拍照或从本地选择图片并直接发送"
              >
                <Camera v-if="!attachmentUploading" class="size-3.5" />
                <LoaderCircle v-else class="size-3.5 animate-spin" />
                拍照
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                class="rounded-full"
                :disabled="replying || attachmentUploading"
                @click="videoPickerOpen = true"
                title="从媒体库选择视频"
              >
                <Video class="size-3.5" />
                媒体库视频
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                class="rounded-full"
                :disabled="replying || attachmentUploading"
                @click="openLocalVideoPicker"
                title="拍视频或从本地选择视频并直接发送"
              >
                <LoaderCircle v-if="attachmentUploading" class="size-3.5 animate-spin" />
                <Video v-else class="size-3.5" />
                拍视频
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                class="rounded-full"
                :disabled="replying || attachmentUploading || !customerOrders.length"
                :title="customerOrders.length ? '从最近订单中选择' : '当前会话暂无可发送订单'"
                @click="orderPickerOpen = true"
              >
                <ShoppingCart class="size-3.5" />
                订单
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                class="rounded-full"
                :disabled="replying"
                @click="productPickerOpen = true"
              >
                <Link2 class="size-3.5" />
                产品链接
              </Button>
            </div>
            <Button type="submit" class="ml-auto rounded-full" :disabled="replying || !replyModel.trim()">
              <LoaderCircle v-if="replying" class="size-4 animate-spin" />
              <Send v-else class="size-4" />
              发送回复
            </Button>
          </div>
        </form>

        <MediaAssetPickerDialog
          v-model:open="imagePickerOpen"
          media-type="image"
          @select="handleImageSelect"
        />
        <MediaAssetPickerDialog
          v-model:open="videoPickerOpen"
          media-type="video"
          @select="handleVideoSelect"
        />
        <GalleryProductPickerDialog
          v-model:open="productPickerOpen"
          @select="handleProductSelect"
        />
        <CustomerServiceOrderPickerDialog
          v-model:open="orderPickerOpen"
          :orders="customerOrders"
          @select="handleOrderSelect"
        />

        <input ref="imageFileInput" class="hidden" type="file" accept="image/*" capture="environment" @change="handleLocalImageUpload" />
        <input ref="videoFileInput" class="hidden" type="file" accept="video/*" capture="environment" @change="handleLocalVideoUpload" />
      </CardContent>
    </template>

    <div v-else class="flex h-full min-h-0 flex-col items-center justify-center p-8 text-center text-muted-foreground">
      <Headset class="mb-3 size-10 opacity-50" />
      <h2 class="text-sm font-black text-foreground">选择一个客户会话</h2>
      <p class="mt-2 max-w-sm text-xs leading-6">
        左侧每个卡片是一条独立 Public Chat 会话。客户之间不会共用消息窗口。
      </p>
    </div>
  </Card>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import {
  ArrowRightLeft,
  Camera,
  Headset,
  Info,
  ImagePlus,
  Link2,
  LoaderCircle,
  MessageCircleOff,
  Package,
  Send,
  ShoppingCart,
  Video,
} from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import GalleryProductPickerDialog from '@/components/admin/gallery/GalleryProductPickerDialog.vue'
import MediaAssetPickerDialog from '@/components/admin/media/MediaAssetPickerDialog.vue'
import CustomerServiceOrderPickerDialog from '@/components/admin/customer-service/CustomerServiceOrderPickerDialog.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import mediaApi from '@/api/media'
import { useAuthStore } from '@/stores/auth'
import { assetTitle } from '@/lib/mediaPresentation'
import { getProductThumbnail } from '@/lib/productMedia'
import { toast } from 'vue-sonner'
import {
  agentDisplayName,
  assigneeName,
  configOptionRows,
  configProduct,
  configSelection,
  formatDate,
  formatOrderTotal,
  formatProductPrice,
  isConfigConfirmMessage,
  isOrderMessage,
  isProductMessage,
  isVideoMessage,
  orderItems,
  orderPayload,
  productPayload,
  videoPayload,
} from '@/lib/customerServicePresentation'
import type { MediaAsset } from '@/api/media'
import type { ProductRecord } from '@/components/admin/product/productEditorTypes'
import type {
  AssignableAgent,
  CustomerContext,
  CustomerConversation,
  CustomerConversationMessage,
  CustomerOrderItem,
  CustomerServiceSendMessagePayload,
  CustomerTypingState,
} from './customerServiceTypes'

interface MediaAssetSelection {
  url: string
  image: MediaAsset
  asset: MediaAsset
}

const props = withDefaults(defineProps<{
  selectedConversation?: CustomerConversation | null
  customerContext?: CustomerContext | null
  messages?: CustomerConversationMessage[]
  messagesLoading?: boolean
  selectedCustomerTyping?: CustomerTypingState | null
  assignableAgents?: AssignableAgent[]
  transferTo?: string
  replyMessage?: string
  transferring?: boolean
  replying?: boolean
  canEdit?: boolean
}>(), {
  selectedConversation: null,
  customerContext: null,
  messages: () => [],
  messagesLoading: false,
  selectedCustomerTyping: null,
  assignableAgents: () => [],
  transferTo: '',
  replyMessage: '',
  transferring: false,
  replying: false,
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'update:transferTo', value: string): void
  (event: 'update:replyMessage', value: string): void
  (event: 'transfer'): void
  (event: 'send-reply'): void
  (event: 'send-message', payload: CustomerServiceSendMessagePayload): void
  (event: 'open-context'): void
  (event: 'typing-input'): void
}>()

const imagePickerOpen = ref(false)
const videoPickerOpen = ref(false)
const productPickerOpen = ref(false)
const orderPickerOpen = ref(false)
const imageFileInput = ref<HTMLInputElement | null>(null)
const videoFileInput = ref<HTMLInputElement | null>(null)
const attachmentUploading = ref(false)
const authStore = useAuthStore()

const customerOrders = computed<CustomerOrderItem[]>(() => (
  Array.isArray(props.customerContext?.orders?.items)
    ? props.customerContext?.orders?.items
    : []
))

const transferModel = computed<string>({
  get: () => props.transferTo,
  set: (value: string) => emit('update:transferTo', value),
})

const hasAssignableAgents = computed(() => props.assignableAgents.length > 0)

const replyModel = computed<string>({
  get: () => props.replyMessage,
  set: (value: string) => emit('update:replyMessage', value),
})

const messageAttachments = (message: CustomerConversationMessage): string[] => {
  const attachments = Array.isArray(message?.attachments)
    ? message.attachments
    : (message?.attachment_url ? [message.attachment_url] : [])
  const unique = new Set<string>()
  attachments.forEach((value) => {
    const attachment = String(value || '').trim()
    if (attachment) unique.add(attachment)
  })
  return Array.from(unique)
}

const toPositiveNumber = (value: unknown): number | null => {
  const parsed = Number(value || 0)
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null
}

const buildProductPath = (slug: string): string => {
  const normalized = String(slug || '').trim()
  return normalized ? `/products/${encodeURIComponent(normalized)}` : ''
}

const buildProductMessageMetadata = (product: ProductRecord) => {
  const thumbnail = getProductThumbnail(product)
  const slug = String(product?.slug || '').trim()
  const priceValue = Number(product?.sale_price ?? product?.price ?? 0)
  const normalizedPriceValue = Number.isFinite(priceValue) ? priceValue : 0
  const currency = String(product?.currency || '').trim().toUpperCase()

  return {
    kind: 'product_reference',
    product_id: toPositiveNumber(product?.id),
    title: String(product?.name || 'Product').trim(),
    slug,
    sku: String(product?.sku || '').trim(),
    url: String(product?.url || buildProductPath(slug)).trim(),
    thumbnail: String(thumbnail?.src || '').trim(),
    price: currency ? `${currency} ${normalizedPriceValue.toFixed(2)}` : (normalizedPriceValue > 0 ? normalizedPriceValue.toFixed(2) : ''),
    price_value: normalizedPriceValue,
  }
}

const buildOrderMessageMetadata = (order: CustomerOrderItem) => {
  const total = Number(order?.total_amount || 0)
  const normalizedTotal = Number.isFinite(total) ? total : 0
  const orderNumber = String(order?.order_number || order?.id || '').trim()
  const items = Array.isArray(order?.items) ? order.items : []
  const normalizedItemCount = Number.isFinite(Number(order?.item_count || items.length || 0))
    ? Number(order?.item_count || items.length || 0)
    : items.length

  return {
    order_number: orderNumber,
    title: orderNumber ? `Order #${orderNumber}` : 'Order',
    status: String(order?.status || '').trim(),
    payment_status: String(order?.payment_status || '').trim(),
    shipping_status: String(order?.shipping_status || '').trim(),
    total: normalizedTotal,
    currency: String(order?.currency || '').trim().toUpperCase(),
    url: String(order?.url || '').trim(),
    thumbnail: String(order?.thumbnail || '').trim(),
    item_count: normalizedItemCount,
    items,
    note: 'Customer asked staff to confirm this order and its purchased configuration.',
  }
}

const buildMediaMessageMetadata = (selection: MediaAssetSelection, mediaType: 'image' | 'video') => ({
  kind: `${mediaType}_reference`,
  title: assetTitle(selection.asset),
  url: selection.url,
  media_type: mediaType,
})

const createMediaSelection = (asset: MediaAsset, url: string): MediaAssetSelection => ({
  url,
  image: asset,
  asset,
})

const sendStructuredMessage = (payload: CustomerServiceSendMessagePayload): void => {
  if (!props.selectedConversation || props.replying) return
  emit('send-message', payload)
}

const sendMediaSelection = (selection: MediaAssetSelection, mediaType: 'image' | 'video'): void => {
  const url = String(selection.url || selection.asset?.url || selection.asset?.access_url || '').trim()
  if (!url) {
    toast.error('附件缺少可用地址，无法发送')
    return
  }

  sendStructuredMessage({
    message: mediaType === 'image' ? '[图片]' : '[视频]',
    messageType: mediaType,
    metadata: buildMediaMessageMetadata(createMediaSelection(selection.asset, url), mediaType),
    attachmentUrl: url,
    attachments: [url],
    toastLabel: mediaType === 'image' ? '图片已发送' : '视频已发送',
  })
}

const handleImageSelect = (selection: MediaAssetSelection): void => {
  imagePickerOpen.value = false
  sendMediaSelection(selection, 'image')
}

const handleVideoSelect = (selection: MediaAssetSelection): void => {
  videoPickerOpen.value = false
  sendMediaSelection(selection, 'video')
}

const handleProductSelect = (product: ProductRecord): void => {
  productPickerOpen.value = false
  const metadata = buildProductMessageMetadata(product)
  sendStructuredMessage({
    message: metadata.title || 'Product',
    messageType: 'product',
    metadata,
    toastLabel: '产品链接已发送',
  })
}

const handleOrderSelect = (order: CustomerOrderItem): void => {
  orderPickerOpen.value = false
  const metadata = buildOrderMessageMetadata(order)
  sendStructuredMessage({
    message: metadata.title || 'Order',
    messageType: 'order',
    metadata,
    toastLabel: '订单已发送',
  })
}

const openLocalImagePicker = (): void => {
  if (props.replying || attachmentUploading.value) return
  imageFileInput.value?.click()
}

const openLocalVideoPicker = (): void => {
  if (props.replying || attachmentUploading.value) return
  videoFileInput.value?.click()
}

const sendUploadedAttachment = async (file: File, mediaType: 'image' | 'video'): Promise<void> => {
  if (!props.selectedConversation || props.replying || attachmentUploading.value) return
  attachmentUploading.value = true
  try {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('media_type', mediaType)
    if (mediaType === 'image') {
      formData.append('image_purpose', 'customer_service_attachment')
    }

    const asset = await mediaApi.uploadAsset(formData)
    const url = String(asset.url || asset.access_url || '').trim()
    if (!url) {
      toast.error('上传成功，但附件地址不可用')
      return
    }

    sendMediaSelection(createMediaSelection(asset, url), mediaType)
  } catch (error) {
    console.error('Failed to upload customer-service attachment:', error)
    toast.error(mediaType === 'image' ? '图片上传失败' : '视频上传失败')
  } finally {
    attachmentUploading.value = false
  }
}

const handleLocalImageUpload = async (event: Event): Promise<void> => {
  const input = event.currentTarget as HTMLInputElement | null
  const file = input?.files?.[0]
  if (input) input.value = ''
  if (!file) return
  await sendUploadedAttachment(file, 'image')
}

const handleLocalVideoUpload = async (event: Event): Promise<void> => {
  const input = event.currentTarget as HTMLInputElement | null
  const file = input?.files?.[0]
  if (input) input.value = ''
  if (!file) return
  await sendUploadedAttachment(file, 'video')
}
</script>
