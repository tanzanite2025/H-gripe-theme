<template>
  <div class="space-y-4">
    <AdminPageHeader title="营销管理" description="管理优惠券、礼品卡、积分规则和会员等级" />

    <AdminStatsGrid :items="statItems" />

    <MarketingTabsPanel
      :active-tab="activeTab"
      :active-sub-tab="activeSubTab"
      :can-create="hasPermission('marketing:create')"
      :can-edit="hasPermission('marketing:edit')"
      :can-delete="hasPermission('marketing:delete')"
      :coupons-loading="couponsLoading"
      :coupons="coupons"
      :coupon-filters="couponFilters"
      :coupon-pagination="couponPagination"
      :coupon-value="couponValue"
      :coupon-status="couponStatus"
      :format-money="formatMoney"
      :format-date="formatDate"
      :gift-cards-loading="giftCardsLoading"
      :gift-cards="giftCards"
      :gift-card-filters="giftCardFilters"
      :gift-card-pagination="giftCardPagination"
      :format-currency="formatCurrency"
      :gift-card-status-name="giftCardStatusName"
      :gift-card-status-tone="giftCardStatusTone"
      :loyalty-loading="loyaltyLoading"
      :loyalty-transactions="loyaltyTransactions"
      :loyalty-filters="loyaltyFilters"
      :loyalty-pagination="loyaltyPagination"
      :loyalty-form="loyaltyForm"
      :loyalty-errors="loyaltyErrors"
      :loyalty-submitting="loyaltySubmitting"
      :loyalty-type-name="loyaltyTypeName"
      :loyalty-settings="loyaltySettings"
      :redeem-settings="redeemSettings"
      :points-base-currency="pointsBaseCurrency"
      :loyalty-program-version="loyaltyProgramVersion"
      :loyalty-program-loading="loyaltyProgramLoading"
      :loyalty-program-saving="loyaltyProgramSaving"
      :redeem-currency-options="redeemCurrencyOptions"
      :redeem-currencies-loading="redeemCurrenciesLoading"
      :levels-loading="levelsLoading"
      :levels="levels"
      :levels-using-fallback="levelsUsingFallback"
      :format-rate="formatRate"
      @coupon-filter-change="applyCouponFilter"
      @create-coupon="showCreateCouponDialog"
      @edit-coupon="showEditCouponDialog"
      @delete-coupon="requestDeleteCoupon"
      @update-coupon-page="updateCouponPage"
      @update-coupon-page-size="updateCouponPageSize"
      @gift-card-filter-change="applyGiftCardFilter"
      @view-gift-card="viewGiftCard"
      @update-gift-card-page="updateGiftCardPage"
      @update-gift-card-page-size="updateGiftCardPageSize"
      @loyalty-filter-change="applyLoyaltyFilter"
      @update-loyalty-page="updateLoyaltyPage"
      @update-loyalty-page-size="updateLoyaltyPageSize"
      @submit-loyalty-adjustment="submitLoyaltyAdjustment"
      @clear-loyalty-error="clearLoyaltyError"
      @refresh-loyalty-program-config="refreshLoyaltyProgramConfig"
      @save-loyalty-program-config="saveLoyaltyProgramConfig"
      @create-level="showCreateLevelDialog"
      @edit-level="showEditLevelDialog"
      @delete-level="requestDeleteLevel"
    />

    <MarketingEditorDialogs
      v-model:coupon-open="couponDialogVisible"
      v-model:level-open="levelDialogVisible"
      :coupon-mode="couponDialogMode"
      :coupon-form="couponForm"
      :coupon-errors="couponErrors"
      :coupon-submitting="couponSubmitting"
      :level-mode="levelDialogMode"
      :level-form="levelForm"
      :level-errors="levelErrors"
      :level-submitting="levelSubmitting"
      @submit-coupon="submitCouponForm"
      @submit-level="submitLevelForm"
      @clear-coupon-error="clearCouponError"
      @clear-level-error="clearLevelError"
    />

    <GiftCardDetailDialog
      v-model:open="giftCardDetailVisible"
      v-model:status-update="giftCardStatusUpdate"
      :current-gift-card="currentGiftCard"
      :loading="giftCardDetailLoading"
      :transactions="giftCardTransactions"
      :status-submitting="giftCardStatusSubmitting"
      :can-edit="hasPermission('marketing:edit')"
      :format-currency="formatCurrency"
      :format-date="formatDate"
      :gift-card-status-name="giftCardStatusName"
      :gift-card-status-tone="giftCardStatusTone"
      :gift-card-status-options="giftCardStatusOptions"
      :transaction-type-name="transactionTypeName"
      @update-status="updateGiftCardStatus"
    />

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

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import { BadgePercent, Coins, Crown, Gift } from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import GiftCardDetailDialog from '@/components/admin/marketing/GiftCardDetailDialog.vue'
import MarketingEditorDialogs from '@/components/admin/marketing/MarketingEditorDialogs.vue'
import MarketingTabsPanel from '@/components/admin/marketing/MarketingTabsPanel.vue'
import { useRouteTab } from '@/composables/useRouteTab'
import type { CouponErrors, CouponForm } from '@/components/admin/marketing/CouponEditorDialog.vue'
import type { MemberLevelErrors, MemberLevelForm } from '@/components/admin/marketing/MemberLevelEditorDialog.vue'
import type { GiftCardTransaction } from '@/components/admin/marketing/GiftCardDetailPanel.vue'
import type {
  CouponRecord,
  GiftCardRecord,
  GiftCardRedeemOption,
  GiftCardRedeemSettings,
  LoyaltyAdjustmentForm,
  LoyaltyErrors,
  LoyaltyTransaction,
  MemberLevel,
} from '@/components/admin/marketing/marketingTypes'
import {
  couponStatus,
  couponValue,
  formatCurrency,
  formatDate,
  formatMoney,
  formatRate,
  giftCardStatusName,
  giftCardStatusOptions,
  giftCardStatusTone,
  loyaltyTypeName,
  toDateTimeLocal,
  toISO,
  transactionTypeName
} from '@/lib/marketingPresentation'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

interface MarketingStats {
  coupons?: {
    total?: number
  }
}

interface CouponEditorForm extends CouponForm {
  id: string | number | null
}

interface MemberLevelEditorForm extends MemberLevelForm {
  id: string | number | null
  sort_order: number | string
  icon: string
  color: string
}

interface GiftCardDetail extends GiftCardRecord {
  id: string | number
}

interface ConfirmationState {
  open: boolean
  type: 'coupon' | 'level' | ''
  target: CouponRecord | MemberLevel | null
  title: string
  description: string
}

interface LoyaltyProgramConfig {
  version?: number | string
  points_base_currency?: string
  available_currencies?: Array<{ code?: string }>
  purchase_earn_points_per_currency_unit?: number | string
  referral_referrer_points?: number | string
  referral_referee_points?: number | string
  checkin_base_points?: number | string
  checkin_streak_interval_days?: number | string
  checkin_streak_bonus_points?: number | string
  checkin_max_points?: number | string
  enabled?: boolean
  currency?: string
  exchange_rate_points?: number | string
  min_redeem_points?: number | string
  max_value_per_day?: number | string
  card_expiry_days?: number | string
  redeem_options?: Array<GiftCardRedeemOption & { id?: string | number }>
}

const authStore = useAuthStore()
const activeTab = useRouteTab({
  defaultValue: 'coupons',
  values: ['coupons', 'giftcards', 'loyalty', 'levels'],
  routes: {
    coupons: 'MarketingCoupons',
    giftcards: 'MarketingGiftCards',
    loyalty: ['MarketingLoyaltyTransactions', 'MarketingLoyaltyRules'],
    levels: 'MarketingLevels',
  },
})
const activeSubTab = useRouteTab({
  defaultValue: 'transactions',
  values: ['transactions', 'rules'],
  routes: {
    transactions: 'MarketingLoyaltyTransactions',
    rules: 'MarketingLoyaltyRules',
  },
  enabled: () => activeTab.value === 'loyalty',
})
const stats = ref<MarketingStats>({})

const couponsLoading = ref(false)
const coupons = ref<CouponRecord[]>([])
const couponFilters = reactive({ status: 'all' })
const couponPagination = reactive({ page: 1, pageSize: 20, total: 0 })
const couponDialogVisible = ref(false)
const couponDialogMode = ref('create')
const couponSubmitting = ref(false)
const couponErrors = reactive<CouponErrors>({})
const couponForm = reactive<CouponEditorForm>({
  id: null, code: '', type: 'fixed', value: 0, description: '', min_amount: 0, max_discount: 0,
  usage_limit: 0, usage_limit_per_user: 0, start_date: '', end_date: '', applicable_products: '',
  excluded_products: '', applicable_categories: '', enabled: true
})

const giftCardsLoading = ref(false)
const giftCards = ref<GiftCardRecord[]>([])
const giftCardFilters = reactive({ status: 'all' })
const giftCardPagination = reactive({ page: 1, pageSize: 20, total: 0 })
const giftCardDetailVisible = ref(false)
const giftCardDetailLoading = ref(false)
const currentGiftCard = ref<GiftCardDetail | null>(null)
const giftCardTransactions = ref<GiftCardTransaction[]>([])
const giftCardStatusUpdate = ref('active')
const giftCardStatusSubmitting = ref(false)
const couponsLoaded = ref(false)
const giftCardsLoaded = ref(false)
const levelsLoaded = ref(false)

const loyaltyLoading = ref(false)
const loyaltyTransactions = ref<LoyaltyTransaction[]>([])
const loyaltyFilters = reactive({ user_id: '' })
const loyaltyPagination = reactive({ page: 1, pageSize: 20, total: 0 })
const loyaltySubmitting = ref(false)
const loyaltyErrors = reactive<LoyaltyErrors>({})
const loyaltyForm = reactive<LoyaltyAdjustmentForm>({ user_id: '', points: 0, description: '' })
const loyaltySettings = reactive({
  tz_loyalty_purchase_earn_points_per_currency_unit: 1,
  tz_loyalty_referral_referrer_points: 100,
  tz_loyalty_referral_referee_points: 50,
  tz_loyalty_checkin_base_points: 10,
  tz_loyalty_checkin_streak_interval_days: 7,
  tz_loyalty_checkin_streak_bonus_points: 5,
  tz_loyalty_checkin_max_points: 50
})
const redeemSettings = reactive({
  tz_redeem_enabled: true,
  tz_redeem_currency: '',
  tz_redeem_exchange_rate: 100,
  tz_redeem_min_points: 1000,
  tz_redeem_max_value_per_day: 500,
  tz_redeem_card_expiry_days: 365,
  options: []
})
const loyaltyProgramVersion = ref(0)
const pointsBaseCurrency = ref('USD')
const loyaltyProgramLoading = ref(false)
const loyaltyProgramSaving = ref(false)
const loyaltyProgramLoaded = ref(false)
const redeemCurrencyOptions = ref<string[]>([])
const redeemCurrenciesLoading = ref(false)
const redeemCurrenciesLoaded = ref(false)

const levelsLoading = ref(false)
const levels = ref<MemberLevel[]>([])
const levelsUsingFallback = ref(false)
const levelDialogVisible = ref(false)
const levelDialogMode = ref('create')
const levelSubmitting = ref(false)
const levelErrors = reactive<MemberLevelErrors>({})
const levelForm = reactive<MemberLevelEditorForm>({
  id: null, name: '', min_points: 0, max_points: 0, discount_rate: 0,
  sort_order: 0, benefits: '', icon: '', color: '#B5FF6D'
})

const DEFAULT_MEMBER_LEVELS = [
  { name: 'Ordinary', min_points: 0, max_points: 499, discount_rate: 0, benefits: '[]', color: '#f8fafc', sort_order: 0 },
  { name: 'Bronze', min_points: 500, max_points: 1999, discount_rate: 0, benefits: '[]', color: '#b87333', sort_order: 10 },
  { name: 'Silver', min_points: 2000, max_points: 4999, discount_rate: 0, benefits: '[]', color: '#c0c0c0', sort_order: 20 },
  { name: 'Gold', min_points: 5000, max_points: 9999, discount_rate: 0, benefits: '[]', color: '#d4af37', sort_order: 30 },
  { name: 'Platinum', min_points: 10000, max_points: 19999, discount_rate: 0, benefits: '[]', color: '#e5e4e2', sort_order: 40 },
  { name: 'Diamond', min_points: 20000, max_points: 999999999, discount_rate: 0, benefits: '[]', color: '#b9f2ff', sort_order: 50 }
]

const defaultMemberLevels = (): MemberLevel[] =>
  DEFAULT_MEMBER_LEVELS.map((level) => ({ ...level, id: null, is_fallback: true }))

const confirmation = reactive<ConfirmationState>({ open: false, type: '', target: null, title: '', description: '' })

const statCount = (value: unknown, unit: string) => `${Number(value || 0).toLocaleString('zh-CN')} ${unit}`
const couponRuleCount = computed(() => stats.value.coupons?.total ?? couponPagination.total ?? coupons.value.length)
const giftCardRedeemStatus = computed(() => {
  if (!loyaltyProgramLoaded.value) return loyaltyProgramLoading.value ? '加载中' : '未加载'
  return redeemSettings.tz_redeem_enabled ? '已启用' : '已停用'
})
const loyaltyProgramVersionLabel = computed(() => {
  if (!loyaltyProgramLoaded.value) return loyaltyProgramLoading.value ? '加载中' : '未加载'
  return loyaltyProgramVersion.value > 0 ? `v${loyaltyProgramVersion.value}` : '未发布'
})
const memberLevelRuleCount = computed(() => levels.value.length || DEFAULT_MEMBER_LEVELS.length)

const statItems = computed(() => [
  { key: 'coupon-rules', label: '优惠券规则', value: statCount(couponRuleCount.value, '条'), icon: BadgePercent, tone: 'gray' },
  { key: 'gift-card-redeem', label: '礼品卡兑换', value: giftCardRedeemStatus.value, icon: Gift, tone: loyaltyProgramLoaded.value && redeemSettings.tz_redeem_enabled ? 'green' : 'gray' },
  { key: 'loyalty-program', label: '积分规则版本', value: loyaltyProgramVersionLabel.value, icon: Coins, tone: 'amber' },
  { key: 'member-level-rules', label: '会员等级规则', value: statCount(memberLevelRuleCount.value, '级'), icon: Crown, tone: 'green' }
])

const apiData = (response: any) => response.data?.data ?? response.data ?? {}
const hasPermission = (permission: string) => authStore.hasPermission(permission)
const clearErrors = (errors: Record<string, unknown>) => Object.keys(errors).forEach((key) => delete errors[key])
const clearCouponError = (field: keyof CouponForm) => { delete couponErrors[field] }
const clearLoyaltyError = (field: keyof LoyaltyAdjustmentForm) => { delete loyaltyErrors[field] }
const clearLevelError = (field: keyof MemberLevelForm) => { delete levelErrors[field] }

const normalizeCurrencyCode = (currency: unknown) => String(currency || '').trim().toUpperCase()

const applyRedeemCurrencySelection = () => {
  if (redeemCurrencyOptions.value.length === 0) return

  const selected = normalizeCurrencyCode(redeemSettings.tz_redeem_currency)
  redeemSettings.tz_redeem_currency = redeemCurrencyOptions.value.includes(selected)
    ? selected
    : redeemCurrencyOptions.value[0]
}

const fetchRedeemCurrencies = async (force = false) => {
  if (!force && redeemCurrenciesLoaded.value) return

  redeemCurrenciesLoading.value = true
  try {
    const response = await axios.get('/api/admin/settings/currency-policy')
    const policy = response.data?.policy || {}
    redeemCurrencyOptions.value = Array.isArray(policy.available_currencies)
      ? policy.available_currencies.map((currency) => normalizeCurrencyCode(currency.code)).filter((currency) => /^[A-Z]{3}$/.test(currency))
      : []
    redeemCurrenciesLoaded.value = true
    applyRedeemCurrencySelection()
  } catch (error) {
    console.error('Failed to fetch redeem currencies:', error)
    redeemCurrencyOptions.value = []
  } finally {
    redeemCurrenciesLoading.value = false
  }
}

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
    couponsLoaded.value = true
  } catch (error) {
    console.error('Failed to fetch coupons:', error)
  } finally {
    couponsLoading.value = false
  }
}
const applyCouponFilter = () => { couponPagination.page = 1; fetchCoupons() }
const updateCouponPage = (page: number) => { couponPagination.page = page; fetchCoupons() }
const updateCouponPageSize = (pageSize: number) => { couponPagination.pageSize = pageSize; couponPagination.page = 1; fetchCoupons() }
const resetCouponForm = () => {
  Object.assign(couponForm, {
    id: null, code: '', type: 'fixed', value: 0, description: '', min_amount: 0, max_discount: 0,
    usage_limit: 0, usage_limit_per_user: 0, start_date: '', end_date: '', applicable_products: '',
    excluded_products: '', applicable_categories: '', enabled: true
  })
  clearErrors(couponErrors)
}
const showCreateCouponDialog = () => { couponDialogMode.value = 'create'; resetCouponForm(); couponDialogVisible.value = true }
const showEditCouponDialog = async (coupon: CouponRecord) => {
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
    giftCardsLoaded.value = true
  } catch (error) {
    console.error('Failed to fetch gift cards:', error)
  } finally {
    giftCardsLoading.value = false
  }
}
const applyGiftCardFilter = () => { giftCardPagination.page = 1; fetchGiftCards() }
const updateGiftCardPage = (page: number) => { giftCardPagination.page = page; fetchGiftCards() }
const updateGiftCardPageSize = (pageSize: number) => { giftCardPagination.pageSize = pageSize; giftCardPagination.page = 1; fetchGiftCards() }
const viewGiftCard = async (giftCard: GiftCardRecord) => {
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
    giftCardStatusUpdate.value = currentGiftCard.value?.status || 'active'
  } catch (error) {
    console.error('Failed to fetch gift card detail:', error)
  } finally {
    giftCardDetailLoading.value = false
  }
}
const updateGiftCardStatus = async () => {
  if (!currentGiftCard.value) return
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

const fetchLoyaltyTransactions = async () => {
  if (!loyaltyFilters.user_id) {
    loyaltyTransactions.value = []
    loyaltyPagination.total = 0
    return
  }
  loyaltyLoading.value = true
  try {
    const response = await axios.get('/api/admin/marketing/loyalty/transactions', {
      params: { user_id: loyaltyFilters.user_id, page: loyaltyPagination.page, page_size: loyaltyPagination.pageSize }
    })
    const data = apiData(response)
    loyaltyTransactions.value = data.transactions || []
    loyaltyPagination.total = response.data.pagination?.total ?? data.total ?? loyaltyTransactions.value.length
  } catch (error) {
    console.error('Failed to fetch loyalty transactions:', error)
  } finally {
    loyaltyLoading.value = false
  }
}
const applyLoyaltyFilter = () => { loyaltyPagination.page = 1; fetchLoyaltyTransactions() }
const updateLoyaltyPage = (page: number) => { loyaltyPagination.page = page; fetchLoyaltyTransactions() }
const updateLoyaltyPageSize = (pageSize: number) => { loyaltyPagination.pageSize = pageSize; loyaltyPagination.page = 1; fetchLoyaltyTransactions() }
const validateLoyaltyAdjustment = () => {
  clearErrors(loyaltyErrors)
  if (!Number(loyaltyForm.user_id)) loyaltyErrors.user_id = '请输入用户 ID'
  if (!Number(loyaltyForm.points)) loyaltyErrors.points = '积分不能为 0'
  if (!loyaltyForm.description.trim()) loyaltyErrors.description = '请输入调整原因'
  if (Object.keys(loyaltyErrors).length) { toast.error('请检查积分调整表单'); return false }
  return true
}
const submitLoyaltyAdjustment = async () => {
  if (!validateLoyaltyAdjustment()) return
  loyaltySubmitting.value = true
  try {
    await axios.post('/api/admin/marketing/loyalty/transactions', {
      user_id: Number(loyaltyForm.user_id),
      points: Number(loyaltyForm.points),
      description: loyaltyForm.description.trim()
    })
    toast.success('积分调整已写入流水')
    loyaltyFilters.user_id = String(loyaltyForm.user_id)
    loyaltyForm.points = 0
    loyaltyForm.description = ''
    await Promise.all([fetchLoyaltyTransactions(), fetchStats()])
  } catch (error) {
    console.error('Failed to adjust loyalty points:', error)
  } finally {
    loyaltySubmitting.value = false
  }
}

const applyLoyaltyProgramConfig = (config?: LoyaltyProgramConfig) => {
  if (!config) return
  loyaltyProgramVersion.value = Number(config.version || 0)
  pointsBaseCurrency.value = normalizeCurrencyCode(config.points_base_currency || 'USD') || 'USD'
  const catalogCurrencies = Array.isArray(config.available_currencies)
    ? config.available_currencies.map((currency) => normalizeCurrencyCode(currency.code)).filter((currency) => /^[A-Z]{3}$/.test(currency))
    : []
  if (catalogCurrencies.length > 0) {
    redeemCurrencyOptions.value = catalogCurrencies
    redeemCurrenciesLoaded.value = true
  }
  Object.assign(loyaltySettings, {
    tz_loyalty_purchase_earn_points_per_currency_unit: Number(config.purchase_earn_points_per_currency_unit ?? 1),
    tz_loyalty_referral_referrer_points: Number(config.referral_referrer_points || 0),
    tz_loyalty_referral_referee_points: Number(config.referral_referee_points || 0),
    tz_loyalty_checkin_base_points: Number(config.checkin_base_points || 0),
    tz_loyalty_checkin_streak_interval_days: Number(config.checkin_streak_interval_days || 1),
    tz_loyalty_checkin_streak_bonus_points: Number(config.checkin_streak_bonus_points || 0),
    tz_loyalty_checkin_max_points: Number(config.checkin_max_points || 0)
  })
  Object.assign(redeemSettings, {
    tz_redeem_enabled: Boolean(config.enabled),
    tz_redeem_currency: normalizeCurrencyCode(config.currency),
    tz_redeem_exchange_rate: Number(config.exchange_rate_points || 0),
    tz_redeem_min_points: Number(config.min_redeem_points || 0),
    tz_redeem_max_value_per_day: Number(config.max_value_per_day ?? 0),
    tz_redeem_card_expiry_days: Number(config.card_expiry_days || 0),
    options: Array.isArray(config.redeem_options)
      ? config.redeem_options.map((option, index) => ({
          key: String(option.id || `option-${index}`),
          value: Number(option.value ?? 0),
          currency: normalizeCurrencyCode(option.currency || config.currency),
          stock_quantity: Number(option.stock_quantity ?? 0),
          redeemed_quantity: Number(option.redeemed_quantity ?? 0),
          remaining_quantity: Number(option.remaining_quantity ?? 0),
        }))
      : []
  })
  applyRedeemCurrencySelection()
}

const fetchLoyaltyProgramConfig = async (force = false) => {
  if (!force && loyaltyProgramLoaded.value) return
  loyaltyProgramLoading.value = true
  try {
    const response = await axios.get('/api/admin/marketing/loyalty/program-config')
    applyLoyaltyProgramConfig(apiData(response).config)
    loyaltyProgramLoaded.value = true
  } catch (error) {
    console.error('Failed to fetch loyalty program config:', error)
  } finally {
    loyaltyProgramLoading.value = false
  }
}

const refreshLoyaltyProgramConfig = () => Promise.all([
  fetchLoyaltyProgramConfig(true),
  fetchRedeemCurrencies(true)
])

const saveLoyaltyProgramConfig = async () => {
  const redeemCurrency = normalizeCurrencyCode(redeemSettings.tz_redeem_currency)
  const redeemOptions = (redeemSettings.options || [])
    .map((option) => ({
      value_cents: Math.round(Number(option.value) * 100),
      currency: normalizeCurrencyCode(option.currency || redeemCurrency),
      stock_quantity: Math.max(0, Math.floor(Number(option.stock_quantity || 0))),
    }))
    .filter((option) => option.value_cents > 0)

  if (redeemSettings.tz_redeem_enabled && redeemOptions.length === 0) {
    toast.error('启用积分兑换时，至少需要一个有效的兑换面值')
    return
  }
  if (!redeemCurrency) {
    toast.error('请在礼品卡页面设置默认币种')
    return
  }
  if (redeemOptions.some((option) => !redeemCurrencyOptions.value.includes(option.currency))) {
    toast.error('兑换面值中存在无效礼品卡币种')
    applyRedeemCurrencySelection()
    return
  }

  loyaltyProgramSaving.value = true
  try {
    const response = await axios.put('/api/admin/marketing/loyalty/program-config', {
      enabled: Boolean(redeemSettings.tz_redeem_enabled),
      currency: redeemCurrency,
      exchange_rate_points: Number(redeemSettings.tz_redeem_exchange_rate),
      min_redeem_points: Number(redeemSettings.tz_redeem_min_points),
      max_value_per_day_cents: Math.round(Number(redeemSettings.tz_redeem_max_value_per_day) * 100),
      card_expiry_days: Number(redeemSettings.tz_redeem_card_expiry_days),
      purchase_earn_points_per_currency_unit: Number(loyaltySettings.tz_loyalty_purchase_earn_points_per_currency_unit),
      referral_referrer_points: Number(loyaltySettings.tz_loyalty_referral_referrer_points),
      referral_referee_points: Number(loyaltySettings.tz_loyalty_referral_referee_points),
      checkin_base_points: Number(loyaltySettings.tz_loyalty_checkin_base_points),
      checkin_streak_interval_days: Number(loyaltySettings.tz_loyalty_checkin_streak_interval_days),
      checkin_streak_bonus_points: Number(loyaltySettings.tz_loyalty_checkin_streak_bonus_points),
      checkin_max_points: Number(loyaltySettings.tz_loyalty_checkin_max_points),
      redeem_values_cents: redeemOptions.map((option) => option.value_cents),
      redeem_options: redeemOptions
    })
    applyLoyaltyProgramConfig(apiData(response).config)
    loyaltyProgramLoaded.value = true
    toast.success('积分与兑换规则已生成新版本')
  } catch (error) {
    console.error('Failed to save loyalty program config:', error)
  } finally {
    loyaltyProgramSaving.value = false
  }
}

const fetchLevels = async () => {
  levelsLoading.value = true
  try {
    const response = await axios.get('/api/admin/marketing/levels')
    const fetchedLevels = apiData(response).levels
    const normalizedLevels = Array.isArray(fetchedLevels) ? fetchedLevels : []
    levels.value = normalizedLevels.length > 0 ? normalizedLevels : defaultMemberLevels()
    levelsUsingFallback.value = normalizedLevels.length === 0
    levelsLoaded.value = true
  } catch (error) {
    console.error('Failed to fetch member levels:', error)
    levels.value = defaultMemberLevels()
    levelsUsingFallback.value = true
    levelsLoaded.value = true
  } finally {
    levelsLoading.value = false
  }
}
const resetLevelForm = () => {
  Object.assign(levelForm, {
    id: null, name: '', min_points: 0, max_points: 0, discount_rate: 0,
    sort_order: 0, benefits: '', icon: '', color: '#B5FF6D'
  })
  clearErrors(levelErrors)
}
const showCreateLevelDialog = () => { levelDialogMode.value = 'create'; resetLevelForm(); levelDialogVisible.value = true }
const showEditLevelDialog = async (level: MemberLevel) => {
  if (!level?.id) {
    toast.error('会员等级接口还未返回真实数据，请确认后端迁移已执行并重启服务')
    return
  }
  levelDialogMode.value = 'edit'
  try {
    const response = await axios.get(`/api/admin/marketing/levels/${level.id}`)
    const data = apiData(response).level || level
    Object.assign(levelForm, {
      id: data.id, name: data.name || '', min_points: Number(data.min_points || 0), max_points: Number(data.max_points || 0),
      discount_rate: Number(data.discount_rate || 0),
      sort_order: Number(data.sort_order || 0), benefits: data.benefits || '', icon: data.icon || '', color: data.color || '#B5FF6D'
    })
    clearErrors(levelErrors)
    levelDialogVisible.value = true
  } catch (error) {
    console.error('Failed to fetch member level detail:', error)
  }
}
const validateLevel = () => {
  clearErrors(levelErrors)
  if (levelDialogMode.value === 'create' && !levelForm.name.trim()) levelErrors.name = '请输入等级名称'
  if (Number(levelForm.min_points) < 0) levelErrors.min_points = '最小积分不能为负数'
  if (Number(levelForm.max_points) < Number(levelForm.min_points)) levelErrors.max_points = '最大积分不能小于最小积分'
  if (Object.keys(levelErrors).length) { toast.error('请检查会员等级表单'); return false }
  return true
}
const submitLevelForm = async () => {
  if (!validateLevel()) return
  levelSubmitting.value = true
  const rulePayload = {
    min_points: Number(levelForm.min_points),
    max_points: Number(levelForm.max_points),
    discount_rate: Number(levelForm.discount_rate || 0),
    benefits: levelForm.benefits
  }
  const payload = levelDialogMode.value === 'create'
    ? {
        ...rulePayload,
        name: levelForm.name.trim(),
        sort_order: Number(levelForm.sort_order || 0),
        icon: levelForm.icon,
        color: levelForm.color
      }
    : rulePayload
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

const requestDeleteCoupon = (coupon: CouponRecord) => Object.assign(confirmation, {
  open: true, type: 'coupon', target: coupon, title: '删除优惠券？',
  description: `优惠券 ${coupon.code} 将被永久删除，此操作不可恢复。`
})
const requestDeleteLevel = (level: MemberLevel) => Object.assign(confirmation, {
  open: true, type: 'level', target: level, title: '删除会员等级？',
  description: `会员等级“${level.name}”将被永久删除，此操作不可恢复。`
})
const executeDelete = async () => {
  const { type, target } = confirmation
  if (!target) return
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

const ensureActiveTabLoaded = () => {
  if (activeTab.value === 'coupons' && !couponsLoaded.value) return fetchCoupons()
  if (activeTab.value === 'giftcards') {
    return Promise.all([
      giftCardsLoaded.value ? Promise.resolve() : fetchGiftCards(),
      loyaltyProgramLoaded.value ? Promise.resolve() : fetchLoyaltyProgramConfig(),
      redeemCurrenciesLoaded.value ? Promise.resolve() : fetchRedeemCurrencies()
    ])
  }
  if (activeTab.value === 'loyalty') {
    return Promise.all([
      loyaltyProgramLoaded.value ? Promise.resolve() : fetchLoyaltyProgramConfig(),
      redeemCurrenciesLoaded.value ? Promise.resolve() : fetchRedeemCurrencies()
    ])
  }
  if (activeTab.value === 'levels' && !levelsLoaded.value) return fetchLevels()
  return Promise.resolve()
}

watch(activeTab, ensureActiveTabLoaded)

onMounted(() => Promise.all([
  fetchStats(),
  fetchLoyaltyProgramConfig(),
  fetchLevels(),
  ensureActiveTabLoaded()
]))
</script>
