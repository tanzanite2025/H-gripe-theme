<template>
  <div class="space-y-4">
    <AdminPageHeader title="Google Merchant" description="独立维护 Google 分发资料；不会修改站内商品录入内容。">
      <template #actions>
        <Button variant="outline" :disabled="loading" @click="refresh">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
          刷新
        </Button>
        <Button v-if="canEdit" @click="openCreate">
          <Plus class="size-4" />
          配置同步商品
        </Button>
      </template>
    </AdminPageHeader>

    <section class="rounded-2xl border bg-muted/20 p-4">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Channel Connection</p>
          <h2 class="mt-1 text-sm font-black">Google 账号与 API 连接</h2>
          <p class="mt-1 max-w-3xl text-xs leading-5 text-muted-foreground">
            账号连接和 Merchant 配置只属于这个渠道页面，不会写入站内商品。
          </p>
        </div>
        <AdminStatusBadge :tone="connectionTone">{{ connectionLabel }}</AdminStatusBadge>
      </div>

      <div class="mt-4 grid gap-3 md:grid-cols-5">
        <div class="rounded-xl border bg-background/70 p-3">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Google 账号</p>
          <p class="mt-2 truncate text-sm font-bold">{{ connection.google_account_email || '尚未绑定' }}</p>
        </div>
        <AdminFormField label="Merchant Account ID">
          <Input v-model="connectionForm.merchant_account_id" placeholder="例如 123456789" />
        </AdminFormField>
        <AdminFormField label="Data Source ID">
          <Input v-model="connectionForm.data_source_id" placeholder="例如 987654321" />
        </AdminFormField>
        <AdminFormField label="Storefront Base URL">
          <Input v-model="connectionForm.storefront_base_url" placeholder="https://tanzanite.site" />
        </AdminFormField>
        <div class="flex items-end justify-end gap-2">
          <Button
            v-if="canEdit"
            variant="outline"
            :disabled="connectionSaving"
            @click="saveConnection"
          >
            <Save class="size-4" />
            {{ connectionSaving ? '保存中' : '保存配置' }}
          </Button>
          <Button
            v-if="canEdit"
            :disabled="oauthStarting || !connection.oauth_configured || !connection.token_encryption_configured"
            @click="connectGoogle"
          >
            <Link2 class="size-4" />
            {{ connection.connected ? '重新连接' : '连接 Google' }}
          </Button>
          <Button
            v-if="canEdit && connection.connected"
            size="icon"
            variant="ghost"
            class="text-destructive"
            title="断开 Google 连接"
            :disabled="disconnecting"
            @click="disconnectGoogle"
          >
            <Unlink2 class="size-4" />
          </Button>
        </div>
      </div>

      <p v-if="connection.last_error" class="mt-3 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs text-destructive">
        {{ connection.last_error }}
      </p>
      <p v-if="!connection.oauth_configured || !connection.token_encryption_configured" class="mt-3 text-xs text-amber-700 dark:text-amber-200">
        服务端尚未完成 Google Merchant OAuth 配置，连接按钮暂不可用。
      </p>
      <p v-else-if="connection.connected" class="mt-3 text-xs text-muted-foreground">
        Merchant Account: {{ connection.merchant_account_id || '未配置' }} · Data Source: {{ connection.data_source_id || '未配置' }} · Storefront: {{ connection.storefront_base_url || '未配置' }}
      </p>
    </section>

    <section class="border-y bg-muted/10 py-4">
      <div class="flex flex-wrap items-center justify-between gap-3 px-1">
        <div>
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/60">Remote Catalog</p>
          <h2 class="mt-1 text-sm font-black">Google Merchant 已处理商品</h2>
        </div>
        <Button
          variant="outline"
          :disabled="remoteLoading || !connection.connected || !connection.merchant_account_id"
          @click="refreshRemoteProducts"
        >
          <RefreshCw :class="['size-4', remoteLoading ? 'animate-spin' : '']" />
          刷新 Google 商品
        </Button>
      </div>
    </section>

    <AdminTablePanel v-if="connection.connected && connection.merchant_account_id" :loading="remoteLoading" :batch-visible="false">
      <Table class="min-w-[980px]">
        <TableHeader>
          <TableRow>
            <TableHead>Google 商品</TableHead>
            <TableHead>Offer ID</TableHead>
            <TableHead>价格</TableHead>
            <TableHead>库存</TableHead>
            <TableHead>处理状态</TableHead>
            <TableHead>目标市场</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="remoteProducts.length === 0" :colspan="6">
            <div class="py-7 text-center text-xs text-muted-foreground">Google Merchant 中暂未返回已处理商品</div>
          </TableEmpty>
          <TableRow v-for="remoteProduct in remoteProducts" :key="remoteProduct.name">
            <TableCell>
              <div class="flex min-w-0 items-center gap-3">
                <img
                  v-if="remoteProduct.product_attributes?.image_link"
                  :src="remoteProduct.product_attributes.image_link"
                  :alt="remoteProduct.product_attributes?.title || remoteProduct.offer_id"
                  class="size-10 shrink-0 rounded border object-cover"
                />
                <div class="min-w-0">
                  <a
                    v-if="remoteProduct.product_attributes?.link"
                    :href="remoteProduct.product_attributes.link"
                    target="_blank"
                    rel="noreferrer"
                    class="block truncate text-sm font-bold hover:text-primary"
                  >
                    {{ remoteProduct.product_attributes?.title || remoteProduct.offer_id }}
                  </a>
                  <p v-else class="truncate text-sm font-bold">{{ remoteProduct.product_attributes?.title || remoteProduct.offer_id }}</p>
                  <p class="mt-0.5 truncate text-xs text-muted-foreground">{{ remoteProduct.name }}</p>
                </div>
              </div>
            </TableCell>
            <TableCell class="font-mono text-xs">{{ remoteProduct.offer_id || '-' }}</TableCell>
            <TableCell class="font-mono text-xs">
              <span>{{ formatRemotePrice(remoteProduct.product_attributes?.sale_price || remoteProduct.product_attributes?.price) }}</span>
              <span v-if="remoteProduct.product_attributes?.sale_price" class="ml-1 text-muted-foreground line-through">
                {{ formatRemotePrice(remoteProduct.product_attributes?.price) }}
              </span>
            </TableCell>
            <TableCell><AdminStatusBadge :tone="remoteAvailabilityTone(remoteProduct.product_attributes?.availability)">{{ remoteAvailabilityLabel(remoteProduct.product_attributes?.availability) }}</AdminStatusBadge></TableCell>
            <TableCell><AdminStatusBadge :tone="remoteStatusTone(remoteProduct)">{{ remoteStatusLabel(remoteProduct) }}</AdminStatusBadge></TableCell>
            <TableCell class="font-mono text-xs">{{ remoteProduct.content_language || '-' }} / {{ remoteProduct.feed_label || '-' }}</TableCell>
          </TableRow>
        </TableBody>
      </Table>
      <div v-if="remoteNextPageToken" class="flex justify-end border-t px-4 py-3">
        <Button size="sm" variant="outline" :disabled="remoteLoading" @click="loadMoreRemoteProducts">
          <ChevronRight class="size-4" />
          加载更多
        </Button>
      </div>
    </AdminTablePanel>

    <section v-else class="border-y py-5 text-center text-xs text-muted-foreground">
      连接 Google 账号并填写 Merchant Account ID 后，可在这里查看 Google 已处理商品。
    </section>

    <AdminTablePanel :loading="loading" :batch-visible="false">
      <Table class="min-w-[1080px]">
        <TableHeader>
          <TableRow>
            <TableHead>站内商品</TableHead>
            <TableHead>SKU</TableHead>
            <TableHead>目标市场</TableHead>
            <TableHead>Offer ID</TableHead>
            <TableHead>状态</TableHead>
            <TableHead>最近校验</TableHead>
            <TableHead class="w-64 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableEmpty v-if="offers.length === 0" :colspan="7">
            <div class="py-8 text-center text-xs text-muted-foreground">还没有选择要同步的 SKU</div>
          </TableEmpty>
          <TableRow v-for="offer in offers" :key="offer.id">
            <TableCell class="font-bold">{{ offer.product?.name || `商品 #${offer.product_id}` }}</TableCell>
            <TableCell class="font-mono text-xs">{{ offer.variant?.sku || `SKU #${offer.variant_id}` }}</TableCell>
            <TableCell class="font-mono text-xs">{{ offer.target_country || '-' }} / {{ offer.currency_code || '-' }}</TableCell>
            <TableCell class="font-mono text-xs">{{ offer.offer_id }}</TableCell>
            <TableCell><AdminStatusBadge :tone="statusTone(offer.sync_status)">{{ statusLabel(offer.sync_status) }}</AdminStatusBadge></TableCell>
            <TableCell class="text-xs text-muted-foreground">
              <div>{{ formatDate(offer.last_validated_at) }}</div>
              <div v-if="offer.last_sync_at" class="mt-1">同步 {{ formatDate(offer.last_sync_at) }}</div>
              <div v-if="offer.last_error" class="mt-1 max-w-56 truncate text-destructive" :title="offer.last_error">{{ offer.last_error }}</div>
            </TableCell>
            <TableCell class="text-right">
              <div class="flex justify-end gap-1">
                <Button v-if="canEdit" size="sm" variant="outline" @click="validate(offer)">校验</Button>
                <Button
                  v-if="canSubmitToGoogle"
                  size="sm"
                  variant="outline"
                  :disabled="!canSync(offer) || syncingOfferId === offer.id"
                  :title="syncButtonTitle(offer)"
                  @click="syncOffer(offer)"
                >
                  <RefreshCw :class="['size-4', syncingOfferId === offer.id ? 'animate-spin' : '']" />
                  {{ syncingOfferId === offer.id ? '同步中' : offer.sync_status === 'sync_failed' ? '重试' : '同步' }}
                </Button>
                <Button
                  v-if="canSubmitToGoogle && hasRemoteSubmission(offer)"
                  size="sm"
                  variant="outline"
                  class="text-destructive"
                  :disabled="!canRemoveRemote(offer) || removingRemoteOfferId === offer.id"
                  :title="removeRemoteButtonTitle(offer)"
                  @click="removeRemote(offer)"
                >
                  <Unlink2 :class="['size-4', removingRemoteOfferId === offer.id ? 'animate-pulse' : '']" />
                  {{ removingRemoteOfferId === offer.id ? '撤回中' : '撤回' }}
                </Button>
                <Button v-if="canEdit" size="icon" variant="ghost" title="编辑同步资料" @click="openEdit(offer)">
                  <Pencil class="size-4" />
                </Button>
                <Button
                  v-if="canEdit"
                  size="icon"
                  variant="ghost"
                  class="text-destructive"
                  :disabled="hasRemoteSubmission(offer)"
                  :title="hasRemoteSubmission(offer) ? '请先从 Google 撤回，再删除本地配置' : '移除同步配置'"
                  @click="remove(offer)"
                >
                  <Trash2 class="size-4" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </AdminTablePanel>

    <Dialog :open="dialogOpen" @update:open="dialogOpen = $event">
      <DialogContent size="lg">
        <DialogHeader>
          <DialogTitle>{{ form.id ? '编辑 Google 同步资料' : '选择 SKU 并配置同步资料' }}</DialogTitle>
          <DialogDescription>这些字段只保存到 Google 渠道记录，不会写入站内商品。</DialogDescription>
        </DialogHeader>
        <form class="space-y-4" @submit.prevent="save">
          <div class="grid gap-4 md:grid-cols-2">
            <AdminFormField label="站内商品" required>
              <select v-model.number="form.product_id" class="h-10 w-full rounded-md border bg-background px-3 text-sm" @change="syncVariant">
                <option :value="0">请选择商品</option>
                <option v-for="product in products" :key="product.id" :value="product.id">{{ product.name }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="SKU" required>
              <select v-model.number="form.variant_id" class="h-10 w-full rounded-md border bg-background px-3 text-sm">
                <option :value="0">请选择 SKU</option>
                <option v-for="variant in selectedProduct?.variants || []" :key="variant.id" :value="variant.id">{{ variant.sku }}</option>
              </select>
            </AdminFormField>
            <AdminFormField label="Offer ID" required><Input v-model="form.offer_id" placeholder="留空自动按 SKU 生成" /></AdminFormField>
            <AdminFormField label="品牌" required><Input v-model="form.brand" placeholder="Google 渠道品牌" /></AdminFormField>
            <AdminFormField label="Google 商品分类" required><Input v-model="form.google_product_category" placeholder="分类名称或分类 ID" /></AdminFormField>
            <AdminFormField label="商品成色" required>
              <select v-model="form.condition" class="h-10 w-full rounded-md border bg-background px-3 text-sm">
                <option value="new">全新</option><option value="used">二手</option><option value="refurbished">翻新</option>
              </select>
            </AdminFormField>
            <AdminFormField label="目标国家" required><Input v-model="form.target_country" maxlength="2" placeholder="US" /></AdminFormField>
            <AdminFormField label="内容语言" required><Input v-model="form.content_language" maxlength="2" placeholder="en" /></AdminFormField>
            <AdminFormField label="结算币种" required><Input v-model="form.currency_code" maxlength="3" placeholder="USD" /></AdminFormField>
            <AdminFormField label="Feed label"><Input v-model="form.feed_label" maxlength="20" placeholder="留空使用目标国家，如 US" /></AdminFormField>
            <AdminFormField label="GTIN"><Input v-model="form.gtin" placeholder="可选" /></AdminFormField>
            <AdminFormField label="MPN"><Input v-model="form.mpn" placeholder="可选" /></AdminFormField>
            <AdminFormField label="商品标识状态" required>
              <select v-model="identifierValue" class="h-10 w-full rounded-md border bg-background px-3 text-sm">
                <option value="">请选择</option><option value="true">有商品标识</option><option value="false">确实没有</option>
              </select>
            </AdminFormField>
            <AdminFormField label="市场价格覆盖"><Input v-model.number="form.price_override" type="number" min="0" step="0.01" placeholder="留空使用站内价格" /></AdminFormField>
            <AdminFormField label="市场促销价覆盖"><Input v-model.number="form.sale_price_override" type="number" min="0" step="0.01" placeholder="可选" /></AdminFormField>
          </div>
          <AdminFormField label="同步标题"><Input v-model="form.title" placeholder="留空可后续按站内商品标题映射" /></AdminFormField>
          <AdminFormField label="同步描述"><Textarea v-model="form.description" class="min-h-24" placeholder="填写 Google 渠道描述" /></AdminFormField>
          <DialogFooter>
            <Button type="button" variant="outline" @click="dialogOpen = false">取消</Button>
            <Button type="submit" :disabled="saving">{{ saving ? '保存中' : '保存同步资料' }}</Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { ChevronRight, Link2, Pencil, Plus, RefreshCw, Save, Trash2, Unlink2 } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import googleMerchantApi from '@/api/googleMerchant'
import productApi from '@/api/products'
import { useAuthStore } from '@/stores/auth'
import { useRoute, useRouter } from 'vue-router'

type GoogleMerchantID = number

interface GoogleMerchantVariant {
  id: GoogleMerchantID
  sku?: string
}

interface GoogleMerchantProduct {
  id: GoogleMerchantID
  name?: string
  variants?: GoogleMerchantVariant[]
}

interface GoogleMerchantOffer {
  id: GoogleMerchantID | null
  product_id: GoogleMerchantID
  variant_id: GoogleMerchantID
  offer_id: string
  title: string
  description: string
  brand: string
  condition: string
  google_product_category: string
  gtin: string
  mpn: string
  identifier_exists: boolean | null
  target_country: string
  content_language: string
  currency_code: string
  feed_label: string
  price_override: number | null
  sale_price_override: number | null
  publication_status: string
  sync_status?: string
  last_validated_at?: string | number | Date | null
  last_sync_at?: string | number | Date | null
  last_error?: string
  product?: { name?: string }
  variant?: { sku?: string }
}

interface GoogleMerchantConnection {
  configured: boolean
  oauth_configured: boolean
  token_encryption_configured: boolean
  connected: boolean
  status: string
  google_account_email: string
  merchant_account_id: string
  data_source_id: string
  storefront_base_url: string
  last_connected_at: string | number | Date | null
  last_error: string
}

interface GoogleMerchantConnectionForm {
  merchant_account_id: string
  data_source_id: string
  storefront_base_url: string
}

interface RemoteProductPrice {
  amount_micros?: number | string
  currency_code?: string
}

interface RemoteProductStatus {
  item_level_issues?: unknown[]
  destination_statuses?: Array<{ approved_countries?: unknown[] }>
}

interface RemoteProduct {
  name: string
  offer_id?: string
  content_language?: string
  feed_label?: string
  archived?: boolean
  product_status?: RemoteProductStatus
  product_attributes?: {
    image_link?: string
    title?: string
    link?: string
    price?: RemoteProductPrice
    sale_price?: RemoteProductPrice
    availability?: string
  }
}

type GoogleMerchantOfferForm = GoogleMerchantOffer

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()
const canEdit = computed(() => authStore.hasPermission('merchant:edit'))
const canSubmitToGoogle = computed(() => authStore.hasPermission('merchant:sync'))
const loading = ref(false)
const saving = ref(false)
const connectionSaving = ref(false)
const oauthStarting = ref(false)
const disconnecting = ref(false)
const remoteLoading = ref(false)
const products = ref<GoogleMerchantProduct[]>([])
const offers = ref<GoogleMerchantOffer[]>([])
const remoteProducts = ref<RemoteProduct[]>([])
const remoteNextPageToken = ref('')
const dialogOpen = ref(false)
const syncingOfferId = ref<GoogleMerchantID | null>(null)
const removingRemoteOfferId = ref<GoogleMerchantID | null>(null)
const form = reactive<GoogleMerchantOfferForm>({
  id: null,
  product_id: 0,
  variant_id: 0,
  offer_id: '',
  title: '',
  description: '',
  brand: '',
  condition: 'new',
  google_product_category: '',
  gtin: '',
  mpn: '',
  identifier_exists: null,
  target_country: '',
  content_language: '',
  currency_code: '',
  feed_label: '',
  price_override: null,
  sale_price_override: null,
  publication_status: 'draft'
})
const connection = reactive<GoogleMerchantConnection>({
  configured: false,
  oauth_configured: false,
  token_encryption_configured: false,
  connected: false,
  status: 'disconnected',
  google_account_email: '',
  merchant_account_id: '',
  data_source_id: '',
  storefront_base_url: '',
  last_connected_at: null,
  last_error: ''
})
const connectionForm = reactive<GoogleMerchantConnectionForm>({
  merchant_account_id: '',
  data_source_id: '',
  storefront_base_url: ''
})
const identifierValue = computed({
  get: () => form.identifier_exists == null ? '' : String(form.identifier_exists),
  set: (value: string) => { form.identifier_exists = value === '' ? null : value === 'true' }
})
const selectedProduct = computed(() => products.value.find((item) => item.id === Number(form.product_id)) || null)

const emptyForm = (): GoogleMerchantOfferForm => ({
  id: null, product_id: 0, variant_id: 0, offer_id: '', title: '', description: '', brand: '',
  condition: 'new', google_product_category: '', gtin: '', mpn: '', identifier_exists: null,
  target_country: '', content_language: '', currency_code: '', feed_label: '',
  price_override: null, sale_price_override: null, publication_status: 'draft'
})
const reset = (values: Partial<GoogleMerchantOfferForm> = {}) => Object.assign(form, emptyForm(), values)
const applyConnection = (values: Partial<GoogleMerchantConnection> = {}) => {
  Object.assign(connection, {
    ...connection,
    ...values
  })
  connectionForm.merchant_account_id = values.merchant_account_id || ''
  connectionForm.data_source_id = values.data_source_id || ''
  connectionForm.storefront_base_url = values.storefront_base_url || ''
}
const syncVariant = () => {
  if (!selectedProduct.value?.variants?.some((variant) => variant.id === Number(form.variant_id))) {
    form.variant_id = selectedProduct.value?.variants?.[0]?.id || 0
  }
}
const refresh = async () => {
  loading.value = true
  try {
    const [connectionPayload, productPayload, offerPayload] = await Promise.all([
      googleMerchantApi.getConnection(),
      productApi.list({ page: 1, page_size: 100, status: 'active' }),
      googleMerchantApi.listOffers()
    ])
    applyConnection(connectionPayload.connection || {})
    products.value = productPayload.products || []
    offers.value = offerPayload.offers || []
    if (connection.connected && connection.merchant_account_id) {
      await refreshRemoteProducts()
    } else {
      remoteProducts.value = []
      remoteNextPageToken.value = ''
    }
  } catch (error) {
    toast.error(error?.response?.data?.message || 'Google 同步资料读取失败')
  } finally {
    loading.value = false
  }
}
const refreshRemoteProducts = async ({ append = false, pageToken = '' } = {}) => {
  if (!connection.connected || !connection.merchant_account_id) {
    remoteProducts.value = []
    remoteNextPageToken.value = ''
    return
  }
  remoteLoading.value = true
  try {
    const result = await googleMerchantApi.listRemoteProducts({ page_size: 50, page_token: pageToken })
    const records = result.products || []
    remoteProducts.value = append ? [...remoteProducts.value, ...records] : records
    remoteNextPageToken.value = result.next_page_token || ''
  } catch (error) {
    if (!append) {
      remoteProducts.value = []
      remoteNextPageToken.value = ''
    }
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'Google 远程商品读取失败')
  } finally {
    remoteLoading.value = false
  }
}
const loadMoreRemoteProducts = () => refreshRemoteProducts({ append: true, pageToken: remoteNextPageToken.value })
const saveConnection = async () => {
  connectionSaving.value = true
  try {
    const result = await googleMerchantApi.updateConnection({ ...connectionForm })
    applyConnection(result.connection || {})
    toast.success('Google Merchant 配置已保存')
  } catch (error) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'Google Merchant 配置保存失败')
  } finally {
    connectionSaving.value = false
  }
}
const connectGoogle = async () => {
  oauthStarting.value = true
  try {
    const result = await googleMerchantApi.startOAuth()
    if (!result.authorization_url) throw new Error('授权地址为空')
    window.location.assign(result.authorization_url)
  } catch (error) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || error.message || 'Google 连接启动失败')
    oauthStarting.value = false
  }
}
const disconnectGoogle = async () => {
  if (!window.confirm('确定断开 Google Merchant 连接吗？')) return
  disconnecting.value = true
  try {
    await googleMerchantApi.disconnect()
    await refresh()
    toast.success('Google Merchant 已断开')
  } catch (error) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'Google Merchant 断开失败')
  } finally {
    disconnecting.value = false
  }
}
const handleOAuthResult = async () => {
  const status = String(route.query.google_merchant_status || '')
  if (!status) return
  const message = String(route.query.google_merchant_message || '')
  if (status === 'connected') {
    toast.success('Google Merchant 已连接')
  } else {
    toast.error(message || 'Google Merchant 连接未完成')
  }
  const query = { ...route.query }
  delete query.google_merchant_status
  delete query.google_merchant_message
  await router.replace({ query })
}
const openCreate = () => {
  reset()
  const productID = Number(route.query.product_id || 0)
  if (productID) {
    form.product_id = productID
    syncVariant()
  }
  dialogOpen.value = true
}
const openEdit = (offer: GoogleMerchantOffer) => {
  reset({ ...offer, id: offer.id, product_id: offer.product_id, variant_id: offer.variant_id })
  dialogOpen.value = true
}
const save = async () => {
  saving.value = true
  try {
    const { id, product, variant, sync_status, last_validated_at, last_sync_at, last_error, ...payload } = form
    const result = form.id ? await googleMerchantApi.updateOffer(form.id, payload) : await googleMerchantApi.createOffer(payload)
    toast.success('Google 同步资料已保存')
    dialogOpen.value = false
    await refresh()
    return result
  } catch (error) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '同步资料保存失败')
  } finally {
    saving.value = false
  }
}
const validate = async (offer: GoogleMerchantOffer) => {
  try {
    await googleMerchantApi.validateOffer(offer.id)
    toast.success('校验通过，已标记为待同步')
    await refresh()
  } catch (error) {
    toast.error(error?.response?.data?.message || error?.response?.data?.error || '校验未通过')
  }
}
const canSync = (offer: GoogleMerchantOffer) => {
  return Boolean(
      connection.connected &&
      connection.merchant_account_id &&
      connection.data_source_id &&
      connection.storefront_base_url &&
      canSubmitToGoogle.value &&
      offer.publication_status === 'ready' &&
      offer.sync_status !== 'syncing'
  )
}
const hasRemoteSubmission = (offer: GoogleMerchantOffer | null | undefined) => Boolean(offer?.last_sync_at && offer.sync_status !== 'removed')
const canRemoveRemote = (offer: GoogleMerchantOffer) => {
  return Boolean(
      canSubmitToGoogle.value &&
      hasRemoteSubmission(offer) &&
      connection.connected &&
      connection.merchant_account_id &&
      connection.data_source_id
  )
}
const syncButtonTitle = (offer: GoogleMerchantOffer) => {
  if (!canSubmitToGoogle.value) return '缺少 Google Merchant 同步权限'
  if (!connection.connected) return '请先连接 Google Merchant'
  if (!connection.merchant_account_id || !connection.data_source_id) return '请先配置 Merchant Account ID 和 Data Source ID'
  if (!connection.storefront_base_url) return '请先配置 Storefront Base URL'
  if (offer.publication_status !== 'ready') return '请先完成校验'
  return offer.sync_status === 'sync_failed' ? '重试此 SKU' : '同步此 SKU'
}
const removeRemoteButtonTitle = (offer: GoogleMerchantOffer) => {
  if (!canSubmitToGoogle.value) return '缺少 Google Merchant 同步权限'
  if (!hasRemoteSubmission(offer)) return '此 SKU 尚未提交到 Google'
  if (!connection.connected) return '请先连接 Google Merchant'
  if (!connection.merchant_account_id || !connection.data_source_id) return '请先配置 Merchant Account ID 和 Data Source ID'
  return '从 Google Merchant 撤回此 SKU，本地同步配置会保留'
}
const syncOffer = async (offer: GoogleMerchantOffer) => {
  syncingOfferId.value = offer.id
  try {
    await googleMerchantApi.syncOffer(offer.id)
    toast.success('单个 SKU 已提交到 Google Merchant')
    await refresh()
  } catch (error) {
    await refresh()
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'SKU 同步失败，可稍后重试')
  } finally {
    syncingOfferId.value = null
  }
}
const removeRemote = async (offer: GoogleMerchantOffer) => {
  if (!window.confirm(`确定从 Google Merchant 撤回 ${offer.offer_id}？本地同步配置会保留。`)) return
  removingRemoteOfferId.value = offer.id
  try {
    await googleMerchantApi.removeRemoteOffer(offer.id)
    toast.success('SKU 已从 Google Merchant 撤回，本地配置已保留')
    await refresh()
  } catch (error) {
    await refresh()
    toast.error(error?.response?.data?.message || error?.response?.data?.error || 'Google 撤回失败，可稍后重试')
  } finally {
    removingRemoteOfferId.value = null
  }
}
const remove = async (offer: GoogleMerchantOffer) => {
  if (hasRemoteSubmission(offer)) {
    toast.error('请先从 Google 撤回，再删除本地同步配置')
    return
  }
  if (!window.confirm(`移除 ${offer.offer_id} 的 Google 同步配置？`)) return
  await googleMerchantApi.deleteOffer(offer.id)
  await refresh()
}
const statusLabel = (status?: string) => ({
  ready: '待同步',
  not_synced: '未校验',
  validation_failed: '校验失败',
  syncing: '同步中',
  synced: '已提交',
  sync_failed: '同步失败',
  removed: '已撤回'
})[status || ''] || status || '-'
const statusTone = (status?: string) => {
  if (status === 'ready' || status === 'synced') return 'green'
  if (status === 'validation_failed' || status === 'sync_failed') return 'coral'
  if (status === 'syncing') return 'amber'
  return 'gray'
}
const connectionLabel = computed(() => {
  if (!connection.oauth_configured || !connection.token_encryption_configured) return '服务端待配置'
  if (connection.connected) return 'Google 已连接'
  if (connection.status === 'error') return '连接异常'
  return '未连接'
})
const connectionTone = computed(() => {
  if (!connection.oauth_configured || !connection.token_encryption_configured) return 'amber'
  if (connection.connected) return 'green'
  if (connection.status === 'error') return 'coral'
  return 'gray'
})
const normalizedAvailability = (availability?: string) => String(availability || '').toLowerCase()
const remoteAvailabilityLabel = (availability?: string) => ({
  in_stock: '有货',
  out_of_stock: '缺货',
  preorder: '预售',
  backorder: '补货中'
})[normalizedAvailability(availability)] || availability || '-'
const remoteAvailabilityTone = (availability?: string) => normalizedAvailability(availability) === 'in_stock' ? 'green' : 'gray'
const remoteStatusLabel = (product: RemoteProduct) => {
  if (product?.product_status?.item_level_issues?.length) return `问题 ${product.product_status.item_level_issues.length}`
  if (product?.archived) return '已归档'
  if (product?.product_status?.destination_statuses?.some((item) => item.approved_countries?.length)) return '已批准'
  return '处理中'
}
const remoteStatusTone = (product: RemoteProduct) => {
  if (product?.product_status?.item_level_issues?.length) return 'coral'
  if (product?.archived) return 'gray'
  if (product?.product_status?.destination_statuses?.some((item) => item.approved_countries?.length)) return 'green'
  return 'amber'
}
const formatRemotePrice = (price?: RemoteProductPrice) => {
  const amount = Number(price?.amount_micros)
  if (!Number.isFinite(amount) || !price?.currency_code) return '-'
  return `${(amount / 1_000_000).toFixed(2)} ${price.currency_code}`
}
const formatDate = (value?: string | number | Date | null) => value ? new Date(value).toLocaleString('zh-CN') : '-'

onMounted(async () => {
  await handleOAuthResult()
  await refresh()
})
</script>
