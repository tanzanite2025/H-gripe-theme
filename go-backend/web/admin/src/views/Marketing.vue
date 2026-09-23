<template>
  <div class="space-y-4">
    <AdminPageHeader title="营销管理" description="管理优惠券、积分规则和会员等级" />

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
      :format-currency="formatCurrency"
      :format-date="formatDate"
      :loyalty-loading="loyaltyLoading"
      :loyalty-transactions="loyaltyTransactions"
      :loyalty-filters="loyaltyFilters"
      :loyalty-pagination="loyaltyPagination"
      :loyalty-form="loyaltyForm"
      :loyalty-errors="loyaltyErrors"
      :loyalty-submitting="loyaltySubmitting"
      :loyalty-type-name="loyaltyTypeName"
      :loyalty-settings="loyaltySettings"
      :points-base-currency="pointsBaseCurrency"
      :loyalty-program-version="loyaltyProgramVersion"
      :loyalty-program-loading="loyaltyProgramLoading"
      :loyalty-program-saving="loyaltyProgramSaving"
      :levels-loading="levelsLoading"
      :levels="levels"
      :levels-using-fallback="levelsUsingFallback"
      :promotion-risk-loading="promotionRiskLoading"
      :promotion-risk-analysis="promotionRiskAnalysis"
      :promotion-risk-error="promotionRiskError"
      :format-rate="formatRate"
      @coupon-filter-change="applyCouponFilter"
      @create-coupon="showCreateCouponDialog"
      @edit-coupon="showEditCouponDialog"
      @delete-coupon="requestDeleteCoupon"
      @update-coupon-page="updateCouponPage"
      @update-coupon-page-size="updateCouponPageSize"
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
      @refresh-promotion-risk-analysis="fetchPromotionRiskAnalysis(true)"
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
import { BadgePercent, Coins, Crown, ShieldAlert } from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import MarketingEditorDialogs from '@/components/admin/marketing/MarketingEditorDialogs.vue'
import MarketingTabsPanel from '@/components/admin/marketing/MarketingTabsPanel.vue'
import { useRouteTab } from '@/composables/useRouteTab'
import type { CouponErrors, CouponForm } from '@/components/admin/marketing/CouponEditorDialog.vue'
import type { MemberLevelErrors, MemberLevelForm } from '@/components/admin/marketing/MemberLevelEditorDialog.vue'
import type {
  CouponRecord,
  LoyaltyAdjustmentForm,
  LoyaltyErrors,
  LoyaltyTransaction,
  MemberLevel,
  PromotionRiskAnalysis,
} from '@/modules/marketing/marketingTypes'
import {
  couponStatus,
  couponValue,
  formatCurrency,
  formatDate,
  formatMoney,
  formatRate,
  loyaltyTypeName,
  toDateTimeLocal,
  toISO
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
}

const authStore = useAuthStore()
const activeTab = useRouteTab({
  defaultValue: 'coupons',
  values: ['coupons', 'loyalty', 'levels', 'risk'],
  routes: {
    coupons: 'MarketingCoupons',
    loyalty: ['MarketingLoyaltyTransactions', 'MarketingLoyaltyRules'],
    levels: 'MarketingLevels',
    risk: 'MarketingPromotionRisk',
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
  id: null, code: '', type: 'fixed', currency: 'USD', value: '0', description: '', min_amount: '0', max_discount: '0',
  usage_limit: 0, usage_limit_per_user: 0, start_date: '', end_date: '', applicable_products: '',
  excluded_products: '', applicable_categories: '', enabled: true
})

const couponsLoaded = ref(false)
const levelsLoaded = ref(false)

const loyaltyLoading = ref(false)
const loyaltyTransactions = ref<LoyaltyTransaction[]>([])
const loyaltyFilters = reactive({ user_id: '' })
const loyaltyPagination = reactive({ page: 1, pageSize: 20, total: 0 })
const loyaltySubmitting = ref(false)
const loyaltyErrors = reactive<LoyaltyErrors>({})
const loyaltyForm = reactive<LoyaltyAdjustmentForm>({ user_id: '', points: 0, description: '' })
const loyaltySettings = reactive({
  points_redemption_enabled: true,
  points_redemption_currency: 'USD',
  points_exchange_rate: 100,
  tz_loyalty_purchase_earn_points_per_currency_unit: 1,
  tz_loyalty_referral_referrer_points: 100,
  tz_loyalty_referral_referee_points: 50,
  tz_loyalty_checkin_base_points: 10,
  tz_loyalty_checkin_streak_interval_days: 7,
  tz_loyalty_checkin_streak_bonus_points: 5,
  tz_loyalty_checkin_max_points: 50
})
const loyaltyProgramVersion = ref(0)
const pointsBaseCurrency = ref('USD')
const loyaltyProgramLoading = ref(false)
const loyaltyProgramSaving = ref(false)
const loyaltyProgramLoaded = ref(false)
const levelsLoading = ref(false)
const levels = ref<MemberLevel[]>([])
const levelsUsingFallback = ref(false)
const levelDialogVisible = ref(false)
const levelDialogMode = ref('create')
const levelSubmitting = ref(false)
const levelErrors = reactive<MemberLevelErrors>({})
const levelForm = reactive<MemberLevelEditorForm>({
  id: null, name: '', min_points: 0, max_points: 0, discount_rate_decimal: '0',
  sort_order: 0, benefits: '', icon: '', color: '#059669'
})
const promotionRiskLoading = ref(false)
const promotionRiskAnalysis = ref<PromotionRiskAnalysis | null>(null)
const promotionRiskError = ref<string | null>(null)
const promotionRiskLoaded = ref(false)

const DEFAULT_MEMBER_LEVELS = [
  { name: 'Ordinary', min_points: 0, max_points: 499, discount_rate_decimal: '0', benefits: '[]', color: '#f8fafc', sort_order: 0 },
  { name: 'Bronze', min_points: 500, max_points: 1999, discount_rate_decimal: '0', benefits: '[]', color: '#b87333', sort_order: 10 },
  { name: 'Silver', min_points: 2000, max_points: 4999, discount_rate_decimal: '0', benefits: '[]', color: '#c0c0c0', sort_order: 20 },
  { name: 'Gold', min_points: 5000, max_points: 9999, discount_rate_decimal: '0', benefits: '[]', color: '#d4af37', sort_order: 30 },
  { name: 'Platinum', min_points: 10000, max_points: 19999, discount_rate_decimal: '0', benefits: '[]', color: '#e5e4e2', sort_order: 40 },
  { name: 'Diamond', min_points: 20000, max_points: 999999999, discount_rate_decimal: '0', benefits: '[]', color: '#b9f2ff', sort_order: 50 }
]

const defaultMemberLevels = (): MemberLevel[] =>
  DEFAULT_MEMBER_LEVELS.map((level) => ({ ...level, id: null, is_fallback: true }))

const confirmation = reactive<ConfirmationState>({ open: false, type: '', target: null, title: '', description: '' })

const statCount = (value: unknown, unit: string) => `${Number(value || 0).toLocaleString('zh-CN')} ${unit}`
const couponRuleCount = computed(() => stats.value.coupons?.total ?? couponPagination.total ?? coupons.value.length)
const loyaltyProgramVersionLabel = computed(() => {
  if (!loyaltyProgramLoaded.value) return loyaltyProgramLoading.value ? '加载中' : '未加载'
  return loyaltyProgramVersion.value > 0 ? `v${loyaltyProgramVersion.value}` : '未发布'
})
const memberLevelRuleCount = computed(() => levels.value.length || DEFAULT_MEMBER_LEVELS.length)
const promotionRiskSeverity = computed(() => promotionRiskAnalysis.value?.summary?.severity || 'info')
const promotionRiskSummaryLabel = computed(() => {
  if (!promotionRiskLoaded.value) return promotionRiskLoading.value ? '加载中' : '未加载'
  const summary = promotionRiskAnalysis.value?.summary
  if (!summary) return '未加载'
  if (Number(summary.zero_total_risk_count || 0) > 0) return `${summary.zero_total_risk_count} 高危`
  if (Number(summary.gateway_minimum_risk_count || 0) > 0) return `${summary.gateway_minimum_risk_count} 预警`
  return '正常'
})

const promotionRiskTone = (severity?: string) => {
  if (severity === 'critical') return 'coral'
  if (severity === 'warning') return 'amber'
  return 'green'
}

const statItems = computed(() => [
  { key: 'coupon-rules', label: '优惠券规则', value: statCount(couponRuleCount.value, '条'), icon: BadgePercent, tone: 'gray' },
  { key: 'loyalty-program', label: '积分规则版本', value: loyaltyProgramVersionLabel.value, icon: Coins, tone: 'amber' },
  { key: 'member-level-rules', label: '会员等级规则', value: statCount(memberLevelRuleCount.value, '级'), icon: Crown, tone: 'green' },
  { key: 'promotion-risk', label: '优惠叠加风险', value: promotionRiskSummaryLabel.value, icon: ShieldAlert, tone: promotionRiskTone(promotionRiskSeverity.value) }
])

const apiData = (response: any) => response.data?.data ?? response.data ?? {}
const hasPermission = (permission: string) => authStore.hasPermission(permission)
const clearErrors = (errors: Record<string, unknown>) => Object.keys(errors).forEach((key) => delete errors[key])
const clearCouponError = (field: keyof CouponForm) => { delete couponErrors[field] }
const clearLoyaltyError = (field: keyof LoyaltyAdjustmentForm) => { delete loyaltyErrors[field] }
const clearLevelError = (field: keyof MemberLevelForm) => { delete levelErrors[field] }

const normalizeCurrencyCode = (currency: unknown) => String(currency || '').trim().toUpperCase()
const minorToMajor = (minor: unknown, currency: unknown) => Number(minor || 0) / (['JPY', 'KRW', 'CLP'].includes(normalizeCurrencyCode(currency)) ? 1 : 100)
const majorToMinor = (major: unknown, currency: unknown) => Math.round(Number(major || 0) * (['JPY', 'KRW', 'CLP'].includes(normalizeCurrencyCode(currency)) ? 1 : 100))

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/marketing/stats')
    stats.value = apiData(response) || {}
  } catch (error) {
    console.error('Failed to fetch marketing stats:', error)
  }
}

const fetchPromotionRiskAnalysis = async (force = false) => {
  if (!force && promotionRiskLoaded.value) return
  if (!force && promotionRiskLoading.value) return

  promotionRiskLoading.value = true
  promotionRiskError.value = null
  try {
    const response = await axios.get('/api/admin/marketing/risk-analysis')
    const data = apiData(response)
    const analysis = data?.analysis
    if (!analysis || typeof analysis !== 'object') {
      throw new Error('优惠风险分析接口返回了无效数据')
    }
    promotionRiskAnalysis.value = analysis
    promotionRiskLoaded.value = true
  } catch (error) {
    console.error('Failed to fetch promotion risk analysis:', error)
    promotionRiskAnalysis.value = null
    promotionRiskLoaded.value = false
    promotionRiskError.value = error instanceof Error && error.message
      ? error.message
      : '优惠风险分析接口返回异常，请重试'
    toast.error('优惠风险分析加载失败')
  } finally {
    promotionRiskLoading.value = false
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
    id: null, code: '', type: 'fixed', currency: 'USD', value: '0', description: '', min_amount: '0', max_discount: '0',
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
      id: data.id, code: data.code || '', type: data.type || 'fixed', currency: data.currency || 'USD',
      value: String(data.type === 'percentage' ? Number(data.value_rate_decimal || 0) : minorToMajor(data.value_minor, data.currency)),
      description: data.description || '', min_amount: String(minorToMajor(data.min_amount_minor, data.currency)), max_discount: String(minorToMajor(data.max_discount_minor, data.currency)),
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
    code: couponForm.code.trim().toUpperCase(), type: couponForm.type, currency: couponForm.currency.trim().toUpperCase(),
    value_minor: couponForm.type === 'fixed' ? majorToMinor(couponForm.value, couponForm.currency) : 0,
    value_rate_decimal: couponForm.type === 'percentage' ? String(couponForm.value) : '',
    description: couponForm.description, min_amount_minor: majorToMinor(couponForm.min_amount, couponForm.currency), max_discount_minor: majorToMinor(couponForm.max_discount, couponForm.currency),
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
    await Promise.all([fetchCoupons(), fetchStats(), fetchPromotionRiskAnalysis(true)])
  } catch (error) {
    console.error('Failed to save coupon:', error)
  } finally {
    couponSubmitting.value = false
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
  Object.assign(loyaltySettings, {
    points_redemption_enabled: Boolean(config.enabled),
    points_redemption_currency: normalizeCurrencyCode(config.currency || 'USD') || 'USD',
    points_exchange_rate: Number(config.exchange_rate_points || 100),
    tz_loyalty_purchase_earn_points_per_currency_unit: Number(config.purchase_earn_points_per_currency_unit ?? 1),
    tz_loyalty_referral_referrer_points: Number(config.referral_referrer_points || 0),
    tz_loyalty_referral_referee_points: Number(config.referral_referee_points || 0),
    tz_loyalty_checkin_base_points: Number(config.checkin_base_points || 0),
    tz_loyalty_checkin_streak_interval_days: Number(config.checkin_streak_interval_days || 1),
    tz_loyalty_checkin_streak_bonus_points: Number(config.checkin_streak_bonus_points || 0),
    tz_loyalty_checkin_max_points: Number(config.checkin_max_points || 0)
  })
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

const refreshLoyaltyProgramConfig = () => fetchLoyaltyProgramConfig(true)

const saveLoyaltyProgramConfig = async () => {
  const pointsCurrency = normalizeCurrencyCode(loyaltySettings.points_redemption_currency)
  if (!/^[A-Z]{3}$/.test(pointsCurrency)) {
    toast.error('请输入有效的积分抵扣币种')
    return
  }
  if (Number(loyaltySettings.points_exchange_rate) <= 0) {
    toast.error('积分抵扣比例必须大于 0')
    return
  }

  loyaltyProgramSaving.value = true
  try {
    const response = await axios.put('/api/admin/marketing/loyalty/program-config', {
      enabled: Boolean(loyaltySettings.points_redemption_enabled),
      currency: pointsCurrency,
      exchange_rate_points: Number(loyaltySettings.points_exchange_rate),
      purchase_earn_points_per_currency_unit: Number(loyaltySettings.tz_loyalty_purchase_earn_points_per_currency_unit),
      referral_referrer_points: Number(loyaltySettings.tz_loyalty_referral_referrer_points),
      referral_referee_points: Number(loyaltySettings.tz_loyalty_referral_referee_points),
      checkin_base_points: Number(loyaltySettings.tz_loyalty_checkin_base_points),
      checkin_streak_interval_days: Number(loyaltySettings.tz_loyalty_checkin_streak_interval_days),
      checkin_streak_bonus_points: Number(loyaltySettings.tz_loyalty_checkin_streak_bonus_points),
      checkin_max_points: Number(loyaltySettings.tz_loyalty_checkin_max_points)
    })
    applyLoyaltyProgramConfig(apiData(response).config)
    loyaltyProgramLoaded.value = true
    await fetchPromotionRiskAnalysis(true)
    toast.success('积分规则已生成新版本')
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
    id: null, name: '', min_points: 0, max_points: 0, discount_rate_decimal: '0',
    sort_order: 0, benefits: '', icon: '', color: '#059669'
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
      discount_rate_decimal: String(data.discount_rate_decimal ?? '0'),
      sort_order: Number(data.sort_order || 0), benefits: data.benefits || '', icon: data.icon || '', color: data.color || '#059669'
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
    discount_rate_decimal: String(levelForm.discount_rate_decimal || '0').trim(),
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
    await Promise.all([fetchLevels(), fetchPromotionRiskAnalysis(true)])
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
      await Promise.all([fetchCoupons(), fetchStats(), fetchPromotionRiskAnalysis(true)])
    } else if (type === 'level') {
      await axios.delete(`/api/admin/marketing/levels/${target.id}`)
      toast.success('会员等级已删除')
      await Promise.all([fetchLevels(), fetchPromotionRiskAnalysis(true)])
    }
  } catch (error) {
    console.error('Failed to delete marketing item:', error)
  }
}

const ensureActiveTabLoaded = () => {
  if (activeTab.value === 'coupons' && !couponsLoaded.value) return fetchCoupons()
  if (activeTab.value === 'loyalty') {
    return loyaltyProgramLoaded.value ? Promise.resolve() : fetchLoyaltyProgramConfig()
  }
  if (activeTab.value === 'levels' && !levelsLoaded.value) return fetchLevels()
  if (activeTab.value === 'risk' && !promotionRiskLoaded.value) return fetchPromotionRiskAnalysis()
  return Promise.resolve()
}

watch(activeTab, ensureActiveTabLoaded)

onMounted(() => Promise.all([
  fetchStats(),
  fetchLoyaltyProgramConfig(),
  fetchLevels(),
  fetchPromotionRiskAnalysis(),
  ensureActiveTabLoaded()
]))
</script>

