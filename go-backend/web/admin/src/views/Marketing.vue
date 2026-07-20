<template>
  <div class="space-y-4">
    <AdminPageHeader title="营销管理" description="管理优惠券、礼品卡和会员等级" />

    <AdminStatsGrid :items="statItems" />

    <Tabs v-model="activeTab" class="gap-4">
      <TabsList variant="line" class="h-10 w-full justify-start overflow-x-auto rounded-none border-b bg-transparent p-0">
        <TabsTrigger value="coupons" class="h-9 flex-none px-4">
          <BadgePercent class="size-4" />
          优惠券
        </TabsTrigger>
        <TabsTrigger value="giftcards" class="h-9 flex-none px-4">
          <Gift class="size-4" />
          礼品卡
        </TabsTrigger>
        <TabsTrigger value="levels" class="h-9 flex-none px-4">
          <Crown class="size-4" />
          会员等级
        </TabsTrigger>
      </TabsList>

      <TabsContent value="coupons" class="space-y-3">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <label class="w-48 space-y-1.5">
            <span class="text-xs font-medium text-muted-foreground">状态</span>
            <Select v-model="couponFilters.status" @update:model-value="applyCouponFilter">
              <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="active">生效中</SelectItem>
                <SelectItem value="expired">已过期</SelectItem>
                <SelectItem value="disabled">已停用</SelectItem>
              </SelectContent>
            </Select>
          </label>
          <Button v-if="hasPermission('marketing:create')" size="sm" @click="showCreateCouponDialog">
            <Plus class="size-3.5" />
            创建优惠券
          </Button>
        </div>

        <AdminTablePanel :loading="couponsLoading">
          <Table class="min-w-[1080px]">
            <TableHeader>
              <TableRow>
                <TableHead class="w-36">优惠码</TableHead>
                <TableHead class="w-24">类型</TableHead>
                <TableHead class="w-28 text-right">折扣值</TableHead>
                <TableHead>描述</TableHead>
                <TableHead class="w-32 text-right">最低消费</TableHead>
                <TableHead class="w-28">使用情况</TableHead>
                <TableHead class="w-60">有效期</TableHead>
                <TableHead class="w-24">状态</TableHead>
                <TableHead class="w-16 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="coupons.length === 0" :colspan="9">
                <div class="flex flex-col items-center text-muted-foreground">
                  <BadgePercent class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无优惠券</span>
                </div>
              </TableEmpty>
              <TableRow v-for="coupon in coupons" :key="coupon.id">
                <TableCell class="font-mono text-xs font-semibold">{{ coupon.code }}</TableCell>
                <TableCell>{{ coupon.type === 'fixed' ? '固定金额' : '百分比' }}</TableCell>
                <TableCell class="text-right font-medium tabular-nums">{{ couponValue(coupon) }}</TableCell>
                <TableCell class="max-w-64 truncate text-muted-foreground">{{ coupon.description || '-' }}</TableCell>
                <TableCell class="text-right tabular-nums">¥{{ formatMoney(coupon.min_amount) }}</TableCell>
                <TableCell class="tabular-nums">{{ coupon.used_count || 0 }} / {{ coupon.usage_limit || '不限' }}</TableCell>
                <TableCell class="text-xs text-muted-foreground">
                  {{ formatDate(coupon.start_date) }}<br />{{ formatDate(coupon.end_date) }}
                </TableCell>
                <TableCell>
                  <AdminStatusBadge :tone="couponStatus(coupon).tone">{{ couponStatus(coupon).label }}</AdminStatusBadge>
                </TableCell>
                <TableCell class="text-right">
                  <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                      <Button variant="ghost" size="icon" :aria-label="`管理优惠券 ${coupon.code}`">
                        <MoreHorizontal class="size-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" class="w-36">
                      <DropdownMenuItem v-if="hasPermission('marketing:edit')" @select="showEditCouponDialog(coupon)">
                        <Pencil class="size-4" />
                        编辑
                      </DropdownMenuItem>
                      <DropdownMenuSeparator v-if="hasPermission('marketing:delete')" />
                      <DropdownMenuItem
                        v-if="hasPermission('marketing:delete')"
                        class="text-destructive focus:text-destructive"
                        @select="requestDeleteCoupon(coupon)"
                      >
                        <Trash2 class="size-4" />
                        删除
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
          <template #footer>
            <AdminPagination
              :page="couponPagination.page"
              :page-size="couponPagination.pageSize"
              :total="couponPagination.total"
              @update:page="updateCouponPage"
              @update:page-size="updateCouponPageSize"
            />
          </template>
        </AdminTablePanel>
      </TabsContent>

      <TabsContent value="giftcards" class="space-y-3">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <label class="w-48 space-y-1.5">
            <span class="text-xs font-medium text-muted-foreground">状态</span>
            <Select v-model="giftCardFilters.status" @update:model-value="applyGiftCardFilter">
              <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="active">活跃</SelectItem>
                <SelectItem value="used">已使用</SelectItem>
                <SelectItem value="expired">已过期</SelectItem>
                <SelectItem value="cancelled">已取消</SelectItem>
              </SelectContent>
            </Select>
          </label>
          <Button v-if="hasPermission('marketing:create')" size="sm" @click="showCreateGiftCardDialog">
            <Plus class="size-3.5" />
            创建礼品卡
          </Button>
        </div>

        <AdminTablePanel :loading="giftCardsLoading">
          <Table class="min-w-[980px]">
            <TableHeader>
              <TableRow>
                <TableHead class="w-44">卡号</TableHead>
                <TableHead class="w-32 text-right">初始金额</TableHead>
                <TableHead class="w-32 text-right">余额</TableHead>
                <TableHead>收件人</TableHead>
                <TableHead class="w-24">状态</TableHead>
                <TableHead class="w-44">到期时间</TableHead>
                <TableHead class="w-44">创建时间</TableHead>
                <TableHead class="w-16 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="giftCards.length === 0" :colspan="8">
                <div class="flex flex-col items-center text-muted-foreground">
                  <Gift class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无礼品卡</span>
                </div>
              </TableEmpty>
              <TableRow v-for="giftCard in giftCards" :key="giftCard.id">
                <TableCell class="font-mono text-xs font-semibold">{{ giftCard.code }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ formatCurrency(giftCard.initial_value, giftCard.currency) }}</TableCell>
                <TableCell class="text-right font-semibold tabular-nums">{{ formatCurrency(giftCard.balance, giftCard.currency) }}</TableCell>
                <TableCell>
                  <span class="block font-medium">{{ giftCard.recipient_name || '-' }}</span>
                  <span class="block text-xs text-muted-foreground">{{ giftCard.recipient_email || '-' }}</span>
                </TableCell>
                <TableCell>
                  <AdminStatusBadge :tone="giftCardStatusTone(giftCard.status)">{{ giftCardStatusName(giftCard.status) }}</AdminStatusBadge>
                </TableCell>
                <TableCell class="text-xs text-muted-foreground">{{ formatDate(giftCard.expires_at) }}</TableCell>
                <TableCell class="text-xs text-muted-foreground">{{ formatDate(giftCard.created_at) }}</TableCell>
                <TableCell class="text-right">
                  <Button variant="ghost" size="icon" :aria-label="`查看礼品卡 ${giftCard.code}`" @click="viewGiftCard(giftCard)">
                    <Eye class="size-4" />
                  </Button>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
          <template #footer>
            <AdminPagination
              :page="giftCardPagination.page"
              :page-size="giftCardPagination.pageSize"
              :total="giftCardPagination.total"
              @update:page="updateGiftCardPage"
              @update:page-size="updateGiftCardPageSize"
            />
          </template>
        </AdminTablePanel>
      </TabsContent>

      <TabsContent value="levels" class="space-y-3">
        <div class="flex justify-end">
          <Button v-if="hasPermission('marketing:create')" size="sm" @click="showCreateLevelDialog">
            <Plus class="size-3.5" />
            创建等级
          </Button>
        </div>

        <AdminTablePanel :loading="levelsLoading">
          <Table class="min-w-[900px]">
            <TableHeader>
              <TableRow>
                <TableHead>等级名称</TableHead>
                <TableHead class="w-52">积分范围</TableHead>
                <TableHead class="w-28 text-right">折扣率</TableHead>
                <TableHead class="w-28 text-right">积分倍数</TableHead>
                <TableHead>权益说明</TableHead>
                <TableHead class="w-20 text-right">排序</TableHead>
                <TableHead class="w-16 text-right">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="levels.length === 0" :colspan="7">
                <div class="flex flex-col items-center text-muted-foreground">
                  <Crown class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无会员等级</span>
                </div>
              </TableEmpty>
              <TableRow v-for="level in levels" :key="level.id">
                <TableCell>
                  <div class="flex items-center gap-2">
                    <span class="size-3 rounded-full border" :style="{ backgroundColor: level.color || '#94a3b8' }" />
                    <span class="font-medium">{{ level.name }}</span>
                  </div>
                </TableCell>
                <TableCell class="tabular-nums">{{ level.min_points }} - {{ level.max_points }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ formatRate(level.discount_rate) }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ Number(level.points_multiplier || 1).toFixed(2) }}x</TableCell>
                <TableCell class="max-w-72 truncate text-muted-foreground">{{ level.benefits || '-' }}</TableCell>
                <TableCell class="text-right tabular-nums">{{ level.sort_order || 0 }}</TableCell>
                <TableCell class="text-right">
                  <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                      <Button variant="ghost" size="icon" :aria-label="`管理会员等级 ${level.name}`">
                        <MoreHorizontal class="size-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" class="w-36">
                      <DropdownMenuItem v-if="hasPermission('marketing:edit')" @select="showEditLevelDialog(level)">
                        <Pencil class="size-4" />
                        编辑
                      </DropdownMenuItem>
                      <DropdownMenuSeparator v-if="hasPermission('marketing:delete')" />
                      <DropdownMenuItem
                        v-if="hasPermission('marketing:delete')"
                        class="text-destructive focus:text-destructive"
                        @select="requestDeleteLevel(level)"
                      >
                        <Trash2 class="size-4" />
                        删除
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </AdminTablePanel>
      </TabsContent>
    </Tabs>

    <CouponEditorDialog
      v-model:open="couponDialogVisible"
      :mode="couponDialogMode"
      :form="couponForm"
      :errors="couponErrors"
      :submitting="couponSubmitting"
      @submit="submitCouponForm"
      @clear-error="clearCouponError"
    />
    <GiftCardEditorDialog
      v-model:open="giftCardDialogVisible"
      :form="giftCardForm"
      :errors="giftCardErrors"
      :submitting="giftCardSubmitting"
      @submit="submitGiftCardForm"
      @clear-error="clearGiftCardError"
    />
    <MemberLevelEditorDialog
      v-model:open="levelDialogVisible"
      :mode="levelDialogMode"
      :form="levelForm"
      :errors="levelErrors"
      :submitting="levelSubmitting"
      @submit="submitLevelForm"
      @clear-error="clearLevelError"
    />

    <Dialog v-model:open="giftCardDetailVisible">
      <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
        <DialogHeader>
          <DialogTitle>礼品卡详情</DialogTitle>
          <DialogDescription v-if="currentGiftCard" class="font-mono">{{ currentGiftCard.code }}</DialogDescription>
        </DialogHeader>

        <div v-if="giftCardDetailLoading" class="flex h-52 items-center justify-center">
          <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载礼品卡详情" />
        </div>
        <div v-else-if="currentGiftCard" class="space-y-6">
          <dl class="grid overflow-hidden rounded-lg border sm:grid-cols-3">
            <DetailItem label="状态">
              <AdminStatusBadge :tone="giftCardStatusTone(currentGiftCard.status)">{{ giftCardStatusName(currentGiftCard.status) }}</AdminStatusBadge>
            </DetailItem>
            <DetailItem label="初始金额">{{ formatCurrency(currentGiftCard.initial_value, currentGiftCard.currency) }}</DetailItem>
            <DetailItem label="当前余额"><strong>{{ formatCurrency(currentGiftCard.balance, currentGiftCard.currency) }}</strong></DetailItem>
            <DetailItem label="收件人">{{ currentGiftCard.recipient_name || '-' }}</DetailItem>
            <DetailItem label="收件邮箱">{{ currentGiftCard.recipient_email || '-' }}</DetailItem>
            <DetailItem label="发送人">{{ currentGiftCard.sender_name || '-' }}</DetailItem>
            <DetailItem label="到期时间">{{ formatDate(currentGiftCard.expires_at) }}</DetailItem>
            <DetailItem label="创建时间">{{ formatDate(currentGiftCard.created_at) }}</DetailItem>
            <DetailItem label="更新时间">{{ formatDate(currentGiftCard.updated_at) }}</DetailItem>
          </dl>

          <div v-if="currentGiftCard.message" class="space-y-1.5">
            <h3 class="text-sm font-semibold">祝福语</h3>
            <p class="whitespace-pre-wrap rounded-lg border bg-muted/30 p-3 text-sm leading-6">{{ currentGiftCard.message }}</p>
          </div>

          <div v-if="hasPermission('marketing:edit')" class="flex flex-col gap-2 border-t pt-5 sm:flex-row sm:items-end">
            <label class="w-full space-y-1.5 sm:w-52">
              <span class="text-xs font-medium">更新状态</span>
              <Select v-model="giftCardStatusUpdate">
                <SelectTrigger class="w-full"><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="active">活跃</SelectItem>
                  <SelectItem value="used">已使用</SelectItem>
                  <SelectItem value="expired">已过期</SelectItem>
                  <SelectItem value="cancelled">已取消</SelectItem>
                </SelectContent>
              </Select>
            </label>
            <Button :disabled="giftCardStatusSubmitting || giftCardStatusUpdate === currentGiftCard.status" @click="updateGiftCardStatus">
              <LoaderCircle v-if="giftCardStatusSubmitting" class="size-4 animate-spin" />
              更新状态
            </Button>
          </div>

          <section class="space-y-3">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-semibold">交易记录</h3>
              <span class="text-xs text-muted-foreground">{{ giftCardTransactions.length }} 条</span>
            </div>
            <div class="overflow-x-auto rounded-lg border">
              <Table class="min-w-[620px]">
                <TableHeader>
                  <TableRow>
                    <TableHead class="w-24">类型</TableHead>
                    <TableHead class="w-32 text-right">金额</TableHead>
                    <TableHead class="w-32 text-right">交易后余额</TableHead>
                    <TableHead>备注</TableHead>
                    <TableHead class="w-44">时间</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableEmpty v-if="giftCardTransactions.length === 0" :colspan="5">暂无交易记录</TableEmpty>
                  <TableRow v-for="transaction in giftCardTransactions" :key="transaction.id">
                    <TableCell>{{ transactionTypeName(transaction.type) }}</TableCell>
                    <TableCell class="text-right tabular-nums">{{ formatCurrency(transaction.amount, currentGiftCard.currency) }}</TableCell>
                    <TableCell class="text-right tabular-nums">{{ formatCurrency(transaction.balance, currentGiftCard.currency) }}</TableCell>
                    <TableCell class="text-muted-foreground">{{ transaction.note || '-' }}</TableCell>
                    <TableCell class="text-xs text-muted-foreground">{{ formatDate(transaction.created_at) }}</TableCell>
                  </TableRow>
                </TableBody>
              </Table>
            </div>
          </section>
        </div>
      </DialogContent>
    </Dialog>

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      :title="confirmation.title"
      :description="confirmation.description"
      confirm-label="删除"
      destructive
      @confirm="executeDelete"
    />
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { toast } from 'vue-sonner'
import { BadgePercent, Crown, Eye, Gift, LoaderCircle, MoreHorizontal, Pencil, Plus, TicketCheck, Trash2, UsersRound } from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import CouponEditorDialog from '@/components/admin/marketing/CouponEditorDialog.vue'
import GiftCardEditorDialog from '@/components/admin/marketing/GiftCardEditorDialog.vue'
import MemberLevelEditorDialog from '@/components/admin/marketing/MemberLevelEditorDialog.vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const DetailItem = defineComponent({
  props: { label: { type: String, required: true } },
  setup(props, { slots }) {
    return () => h('div', { class: 'border-b p-3 last:border-b-0 sm:border-b sm:border-r sm:nth-[3n]:border-r-0' }, [
      h('dt', { class: 'text-xs font-medium text-muted-foreground' }, props.label),
      h('dd', { class: 'mt-1 break-words text-sm' }, slots.default?.())
    ])
  }
})

const authStore = useAuthStore()
const activeTab = ref('coupons')
const stats = ref({})

const couponsLoading = ref(false)
const coupons = ref([])
const couponFilters = reactive({ status: 'all' })
const couponPagination = reactive({ page: 1, pageSize: 20, total: 0 })
const couponDialogVisible = ref(false)
const couponDialogMode = ref('create')
const couponSubmitting = ref(false)
const couponErrors = reactive({})
const couponForm = reactive({
  id: null, code: '', type: 'fixed', value: 0, description: '', min_amount: 0, max_discount: 0,
  usage_limit: 0, usage_limit_per_user: 0, start_date: '', end_date: '', applicable_products: '',
  excluded_products: '', applicable_categories: '', enabled: true
})

const giftCardsLoading = ref(false)
const giftCards = ref([])
const giftCardFilters = reactive({ status: 'all' })
const giftCardPagination = reactive({ page: 1, pageSize: 20, total: 0 })
const giftCardDialogVisible = ref(false)
const giftCardSubmitting = ref(false)
const giftCardErrors = reactive({})
const giftCardForm = reactive({
  code: '', initial_value: 0, currency: 'USD', recipient_email: '', recipient_name: '', sender_name: '',
  message: '', cover_image: '', expires_at: ''
})
const giftCardDetailVisible = ref(false)
const giftCardDetailLoading = ref(false)
const currentGiftCard = ref(null)
const giftCardTransactions = ref([])
const giftCardStatusUpdate = ref('active')
const giftCardStatusSubmitting = ref(false)

const levelsLoading = ref(false)
const levels = ref([])
const levelDialogVisible = ref(false)
const levelDialogMode = ref('create')
const levelSubmitting = ref(false)
const levelErrors = reactive({})
const levelForm = reactive({
  id: null, name: '', min_points: 0, max_points: 0, discount_rate: 0, points_multiplier: 1,
  sort_order: 0, benefits: '', icon: '', color: '#2563eb'
})

const confirmation = reactive({ open: false, type: '', target: null, title: '', description: '' })

const statItems = computed(() => [
  { key: 'coupon-total', label: '优惠券总数', value: stats.value.coupons?.total || 0, icon: BadgePercent, tone: 'gray' },
  { key: 'coupon-active', label: '活跃优惠券', value: stats.value.coupons?.active || 0, icon: TicketCheck, tone: 'green' },
  { key: 'coupon-used', label: '已使用次数', value: stats.value.coupons?.used || 0, icon: Gift, tone: 'blue' },
  { key: 'members', label: '会员总数', value: stats.value.loyalty?.total_members || 0, icon: UsersRound, tone: 'amber' }
])

const apiData = (response) => response.data?.data ?? response.data ?? {}
const hasPermission = (permission) => authStore.hasPermission(permission)
const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
const formatMoney = (amount) => Number(amount || 0).toFixed(2)
const formatCurrency = (amount, currency = 'USD') => {
  try {
    return new Intl.NumberFormat('zh-CN', { style: 'currency', currency: currency || 'USD' }).format(Number(amount || 0))
  } catch {
    return `${currency || ''} ${formatMoney(amount)}`.trim()
  }
}
const formatRate = (rate) => `${Number(rate || 0).toFixed(2)}%`
const couponValue = (coupon) => coupon.type === 'percentage' ? `${formatMoney(coupon.value)}%` : `¥${formatMoney(coupon.value)}`
const couponStatus = (coupon) => {
  const now = Date.now()
  if (!coupon.enabled) return { label: '已停用', tone: 'gray' }
  if (coupon.end_date && now > new Date(coupon.end_date).getTime()) return { label: '已过期', tone: 'amber' }
  if (coupon.start_date && now < new Date(coupon.start_date).getTime()) return { label: '未开始', tone: 'blue' }
  return { label: '生效中', tone: 'green' }
}
const giftCardStatusName = (status) => ({ active: '活跃', used: '已使用', expired: '已过期', cancelled: '已取消' })[status] || status || '-'
const giftCardStatusTone = (status) => ({ active: 'green', used: 'blue', expired: 'amber', cancelled: 'coral' })[status] || 'gray'
const transactionTypeName = (type) => ({ issue: '发行', use: '消费', refund: '退款' })[type] || type || '-'
const toDateTimeLocal = (value) => {
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
  return local.toISOString().slice(0, 16)
}
const toISO = (value) => value ? new Date(value).toISOString() : null
const clearErrors = (errors) => Object.keys(errors).forEach((key) => delete errors[key])
const clearCouponError = (field) => { delete couponErrors[field] }
const clearGiftCardError = (field) => { delete giftCardErrors[field] }
const clearLevelError = (field) => { delete levelErrors[field] }

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/marketing/stats')
    stats.value = apiData(response) || {}
  } catch (error) {
    console.error('Failed to fetch marketing stats:', error)
  }
}

const fetchCoupons = async () => {
  couponsLoading.value = true
  try {
    const response = await axios.get('/api/admin/marketing/coupons', {
      params: { page: couponPagination.page, page_size: couponPagination.pageSize, status: couponFilters.status }
    })
    const data = apiData(response)
    coupons.value = Array.isArray(data) ? data : data.coupons || []
    couponPagination.total = response.data.pagination?.total ?? coupons.value.length
  } catch (error) {
    console.error('Failed to fetch coupons:', error)
  } finally {
    couponsLoading.value = false
  }
}
const applyCouponFilter = () => { couponPagination.page = 1; fetchCoupons() }
const updateCouponPage = (page) => { couponPagination.page = page; fetchCoupons() }
const updateCouponPageSize = (pageSize) => { couponPagination.pageSize = pageSize; couponPagination.page = 1; fetchCoupons() }
const resetCouponForm = () => {
  Object.assign(couponForm, {
    id: null, code: '', type: 'fixed', value: 0, description: '', min_amount: 0, max_discount: 0,
    usage_limit: 0, usage_limit_per_user: 0, start_date: '', end_date: '', applicable_products: '',
    excluded_products: '', applicable_categories: '', enabled: true
  })
  clearErrors(couponErrors)
}
const showCreateCouponDialog = () => { couponDialogMode.value = 'create'; resetCouponForm(); couponDialogVisible.value = true }
const showEditCouponDialog = async (coupon) => {
  couponDialogMode.value = 'edit'
  try {
    const response = await axios.get(`/api/admin/marketing/coupons/${coupon.id}`)
    const data = apiData(response).coupon || coupon
    Object.assign(couponForm, {
      id: data.id, code: data.code || '', type: data.type || 'fixed', value: Number(data.value || 0),
      description: data.description || '', min_amount: Number(data.min_amount || 0), max_discount: Number(data.max_discount || 0),
      usage_limit: Number(data.usage_limit || 0), usage_limit_per_user: Number(data.usage_limit_per_user || 0),
      start_date: toDateTimeLocal(data.start_date), end_date: toDateTimeLocal(data.end_date),
      applicable_products: data.applicable_products || '', excluded_products: data.excluded_products || '',
      applicable_categories: data.applicable_categories || '', enabled: data.enabled !== false
    })
    clearErrors(couponErrors)
    couponDialogVisible.value = true
  } catch (error) {
    console.error('Failed to fetch coupon detail:', error)
  }
}
const validateCoupon = () => {
  clearErrors(couponErrors)
  if (!couponForm.code.trim()) couponErrors.code = '请输入优惠码'
  if (Number(couponForm.value) <= 0) couponErrors.value = '折扣值必须大于 0'
  else if (couponForm.type === 'percentage' && Number(couponForm.value) > 100) couponErrors.value = '百分比不能大于 100'
  if (!couponForm.start_date) couponErrors.start_date = '请选择开始时间'
  if (!couponForm.end_date) couponErrors.end_date = '请选择结束时间'
  else if (couponForm.start_date && new Date(couponForm.end_date) <= new Date(couponForm.start_date)) couponErrors.end_date = '结束时间必须晚于开始时间'
  if (Object.keys(couponErrors).length) { toast.error('请检查优惠券表单'); return false }
  return true
}
const submitCouponForm = async () => {
  if (!validateCoupon()) return
  couponSubmitting.value = true
  const payload = {
    code: couponForm.code.trim().toUpperCase(), type: couponForm.type, value: Number(couponForm.value),
    description: couponForm.description, min_amount: Number(couponForm.min_amount || 0), max_discount: Number(couponForm.max_discount || 0),
    usage_limit: Number(couponForm.usage_limit || 0), usage_limit_per_user: Number(couponForm.usage_limit_per_user || 0),
    start_date: toISO(couponForm.start_date), end_date: toISO(couponForm.end_date), applicable_products: couponForm.applicable_products,
    excluded_products: couponForm.excluded_products, applicable_categories: couponForm.applicable_categories, enabled: couponForm.enabled
  }
  try {
    if (couponDialogMode.value === 'create') {
      await axios.post('/api/admin/marketing/coupons', payload)
      toast.success('优惠券创建成功')
    } else {
      await axios.put(`/api/admin/marketing/coupons/${couponForm.id}`, payload)
      toast.success('优惠券更新成功')
    }
    couponDialogVisible.value = false
    await Promise.all([fetchCoupons(), fetchStats()])
  } catch (error) {
    console.error('Failed to save coupon:', error)
  } finally {
    couponSubmitting.value = false
  }
}

const fetchGiftCards = async () => {
  giftCardsLoading.value = true
  try {
    const response = await axios.get('/api/admin/marketing/gift-cards', {
      params: { page: giftCardPagination.page, page_size: giftCardPagination.pageSize, status: giftCardFilters.status }
    })
    const data = apiData(response)
    giftCards.value = data.gift_cards || []
    giftCardPagination.total = response.data.pagination?.total ?? giftCards.value.length
  } catch (error) {
    console.error('Failed to fetch gift cards:', error)
  } finally {
    giftCardsLoading.value = false
  }
}
const applyGiftCardFilter = () => { giftCardPagination.page = 1; fetchGiftCards() }
const updateGiftCardPage = (page) => { giftCardPagination.page = page; fetchGiftCards() }
const updateGiftCardPageSize = (pageSize) => { giftCardPagination.pageSize = pageSize; giftCardPagination.page = 1; fetchGiftCards() }
const resetGiftCardForm = () => {
  Object.assign(giftCardForm, {
    code: '', initial_value: 0, currency: 'USD', recipient_email: '', recipient_name: '', sender_name: '',
    message: '', cover_image: '', expires_at: ''
  })
  clearErrors(giftCardErrors)
}
const showCreateGiftCardDialog = () => { resetGiftCardForm(); giftCardDialogVisible.value = true }
const validateGiftCard = () => {
  clearErrors(giftCardErrors)
  if (!giftCardForm.code.trim()) giftCardErrors.code = '请输入卡号'
  if (Number(giftCardForm.initial_value) <= 0) giftCardErrors.initial_value = '金额必须大于 0'
  if (Object.keys(giftCardErrors).length) { toast.error('请检查礼品卡表单'); return false }
  return true
}
const submitGiftCardForm = async () => {
  if (!validateGiftCard()) return
  giftCardSubmitting.value = true
  try {
    await axios.post('/api/admin/marketing/gift-cards', {
      code: giftCardForm.code.trim().toUpperCase(), initial_value: Number(giftCardForm.initial_value),
      currency: giftCardForm.currency.trim().toUpperCase() || 'USD', recipient_email: giftCardForm.recipient_email.trim(),
      recipient_name: giftCardForm.recipient_name.trim(), sender_name: giftCardForm.sender_name.trim(),
      message: giftCardForm.message, cover_image: giftCardForm.cover_image.trim(), expires_at: toISO(giftCardForm.expires_at)
    })
    toast.success('礼品卡创建成功')
    giftCardDialogVisible.value = false
    await fetchGiftCards()
  } catch (error) {
    console.error('Failed to create gift card:', error)
  } finally {
    giftCardSubmitting.value = false
  }
}
const viewGiftCard = async (giftCard) => {
  currentGiftCard.value = giftCard
  giftCardTransactions.value = []
  giftCardStatusUpdate.value = giftCard.status
  giftCardDetailVisible.value = true
  giftCardDetailLoading.value = true
  try {
    const response = await axios.get(`/api/admin/marketing/gift-cards/${giftCard.id}`)
    const data = apiData(response)
    currentGiftCard.value = data.gift_card || giftCard
    giftCardTransactions.value = data.transactions || []
    giftCardStatusUpdate.value = currentGiftCard.value.status
  } catch (error) {
    console.error('Failed to fetch gift card detail:', error)
  } finally {
    giftCardDetailLoading.value = false
  }
}
const updateGiftCardStatus = async () => {
  giftCardStatusSubmitting.value = true
  try {
    const response = await axios.patch(`/api/admin/marketing/gift-cards/${currentGiftCard.value.id}/status`, { status: giftCardStatusUpdate.value })
    currentGiftCard.value = apiData(response).gift_card || { ...currentGiftCard.value, status: giftCardStatusUpdate.value }
    toast.success('礼品卡状态已更新')
    await fetchGiftCards()
  } catch (error) {
    console.error('Failed to update gift card status:', error)
  } finally {
    giftCardStatusSubmitting.value = false
  }
}

const fetchLevels = async () => {
  levelsLoading.value = true
  try {
    const response = await axios.get('/api/admin/marketing/levels')
    levels.value = apiData(response).levels || []
  } catch (error) {
    console.error('Failed to fetch member levels:', error)
  } finally {
    levelsLoading.value = false
  }
}
const resetLevelForm = () => {
  Object.assign(levelForm, {
    id: null, name: '', min_points: 0, max_points: 0, discount_rate: 0, points_multiplier: 1,
    sort_order: 0, benefits: '', icon: '', color: '#2563eb'
  })
  clearErrors(levelErrors)
}
const showCreateLevelDialog = () => { levelDialogMode.value = 'create'; resetLevelForm(); levelDialogVisible.value = true }
const showEditLevelDialog = async (level) => {
  levelDialogMode.value = 'edit'
  try {
    const response = await axios.get(`/api/admin/marketing/levels/${level.id}`)
    const data = apiData(response).level || level
    Object.assign(levelForm, {
      id: data.id, name: data.name || '', min_points: Number(data.min_points || 0), max_points: Number(data.max_points || 0),
      discount_rate: Number(data.discount_rate || 0), points_multiplier: Number(data.points_multiplier || 1),
      sort_order: Number(data.sort_order || 0), benefits: data.benefits || '', icon: data.icon || '', color: data.color || '#2563eb'
    })
    clearErrors(levelErrors)
    levelDialogVisible.value = true
  } catch (error) {
    console.error('Failed to fetch member level detail:', error)
  }
}
const validateLevel = () => {
  clearErrors(levelErrors)
  if (!levelForm.name.trim()) levelErrors.name = '请输入等级名称'
  if (Number(levelForm.min_points) < 0) levelErrors.min_points = '最小积分不能为负数'
  if (Number(levelForm.max_points) < Number(levelForm.min_points)) levelErrors.max_points = '最大积分不能小于最小积分'
  if (Object.keys(levelErrors).length) { toast.error('请检查会员等级表单'); return false }
  return true
}
const submitLevelForm = async () => {
  if (!validateLevel()) return
  levelSubmitting.value = true
  const payload = {
    name: levelForm.name.trim(), min_points: Number(levelForm.min_points), max_points: Number(levelForm.max_points),
    discount_rate: Number(levelForm.discount_rate || 0), points_multiplier: Number(levelForm.points_multiplier || 1),
    sort_order: Number(levelForm.sort_order || 0), benefits: levelForm.benefits, icon: levelForm.icon, color: levelForm.color
  }
  try {
    if (levelDialogMode.value === 'create') {
      await axios.post('/api/admin/marketing/levels', payload)
      toast.success('会员等级创建成功')
    } else {
      await axios.put(`/api/admin/marketing/levels/${levelForm.id}`, payload)
      toast.success('会员等级更新成功')
    }
    levelDialogVisible.value = false
    await fetchLevels()
  } catch (error) {
    console.error('Failed to save member level:', error)
  } finally {
    levelSubmitting.value = false
  }
}

const requestDeleteCoupon = (coupon) => Object.assign(confirmation, {
  open: true, type: 'coupon', target: coupon, title: '删除优惠券？',
  description: `优惠券 ${coupon.code} 将被永久删除，此操作不可恢复。`
})
const requestDeleteLevel = (level) => Object.assign(confirmation, {
  open: true, type: 'level', target: level, title: '删除会员等级？',
  description: `会员等级“${level.name}”将被永久删除，此操作不可恢复。`
})
const executeDelete = async () => {
  const { type, target } = confirmation
  confirmation.open = false
  try {
    if (type === 'coupon') {
      await axios.delete(`/api/admin/marketing/coupons/${target.id}`)
      toast.success('优惠券已删除')
      await Promise.all([fetchCoupons(), fetchStats()])
    } else if (type === 'level') {
      await axios.delete(`/api/admin/marketing/levels/${target.id}`)
      toast.success('会员等级已删除')
      await fetchLevels()
    }
  } catch (error) {
    console.error('Failed to delete marketing item:', error)
  }
}

onMounted(() => Promise.all([fetchStats(), fetchCoupons(), fetchGiftCards(), fetchLevels()]))
</script>
