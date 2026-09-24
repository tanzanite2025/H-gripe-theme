/**
 * 会员中心数据管理 composable
 * 统一管理会员等级、积分和优惠券数据
 */
import { ref, computed, onMounted } from 'vue'
import { useAuth } from '~/composables/useAuth'
import { isExpectedOptionalConfigMiss, logUnexpectedApiError } from '~/utils/storefrontApiFailures'

// 等级配置类型
interface TierConfig {
  key: string
  name: string
  min: number
  max: number | null
  discount: number
}

interface LoyaltyRules {
  version: number
  currency: string
  points_base_currency: string
  purchase_earn_points_per_currency_unit: number | null
  purchase_earn_trigger: string
  purchase_earn_amount_basis: string
  referral_referrer_points: number | null
  referral_referee_points: number | null
  checkin_base_points: number | null
  checkin_streak_interval_days: number | null
  checkin_streak_bonus_points: number | null
  checkin_max_points: number | null
}

type LoyaltyRecord = Record<string, unknown>
type LoyaltyTierRecord = LoyaltyRecord & {
  min?: number | string | null
  max?: number | string | null
}
interface MembershipLoadOptions {
  includePublicConfig?: boolean
}

const toFiniteNumber = (value: unknown, fallback = 0) => {
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : fallback
}

const toNullableNumber = (value: unknown) => {
  if (value === null || typeof value === 'undefined' || value === '') return null
  const numberValue = Number(value)
  return Number.isFinite(numberValue) ? numberValue : null
}

const normalizeLoyaltyRules = (raw: any): LoyaltyRules | null => {
  if (!raw || typeof raw !== 'object') return null
  const pointsBaseCurrency = String(raw.points_base_currency || raw.base_currency || 'USD').trim().toUpperCase() || 'USD'

  return {
    version: toFiniteNumber(raw.version),
    currency: pointsBaseCurrency,
    points_base_currency: pointsBaseCurrency,
    purchase_earn_points_per_currency_unit: toNullableNumber(raw.purchase_earn_points_per_currency_unit),
    purchase_earn_trigger: String(raw.purchase_earn_trigger || 'order_completed'),
    purchase_earn_amount_basis: String(raw.purchase_earn_amount_basis || 'order_subtotal_minus_discounts'),
    referral_referrer_points: toNullableNumber(raw.referral_referrer_points),
    referral_referee_points: toNullableNumber(raw.referral_referee_points),
    checkin_base_points: toNullableNumber(raw.checkin_base_points),
    checkin_streak_interval_days: toNullableNumber(raw.checkin_streak_interval_days),
    checkin_streak_bonus_points: toNullableNumber(raw.checkin_streak_bonus_points),
    checkin_max_points: toNullableNumber(raw.checkin_max_points),
  }
}

const emptyTierInfo = (): { current: LoyaltyTierRecord | null; next: LoyaltyTierRecord | null; pct: number } => ({ current: null, next: null, pct: 0 })

const normalizeTierConfigFromBackend = (tier: any): TierConfig => {
  const name = String(tier?.name ?? tier?.label ?? tier?.key ?? '')
  const keySource = String(tier?.key ?? (name || tier?.id || ''))
  const maxPoints = tier?.max_points ?? tier?.max

  return {
    key: keySource.toLowerCase().replace(/\s+/g, '-'),
    name: name || keySource.toUpperCase(),
    min: toFiniteNumber(tier?.min_points ?? tier?.min),
    max: maxPoints === -1 || maxPoints === null || typeof maxPoints === 'undefined'
      ? null
      : toFiniteNumber(maxPoints),
    discount: toFiniteNumber(tier?.discount_rate_decimal),
  }
}

const DEFAULT_TIER_CONFIGS: TierConfig[] = [
  { key: 'ordinary', name: 'Ordinary', min: 0, max: 499, discount: 0 },
  { key: 'bronze', name: 'Bronze', min: 500, max: 1999, discount: 0 },
  { key: 'silver', name: 'Silver', min: 2000, max: 4999, discount: 0 },
  { key: 'gold', name: 'Gold', min: 5000, max: 9999, discount: 0 },
  { key: 'platinum', name: 'Platinum', min: 10000, max: 19999, discount: 0 },
  { key: 'diamond', name: 'Diamond', min: 20000, max: null, discount: 0 },
]

const defaultTierConfigs = () =>
  DEFAULT_TIER_CONFIGS.map(tier => ({ ...tier }))

export function useMembership() {
  const auth = useAuth()

  // ========== 用户数据 ==========
  const userData = computed(() => auth.user.value)
  const loyalty = computed<LoyaltyRecord | null>(() => {
    const value = userData.value?.loyalty
    return value && typeof value === 'object' ? value as LoyaltyRecord : null
  })
  const isLogged = computed(() => !!userData.value)
  const levelName = computed<string>(() => String(loyalty.value?.level || '—'))
  const topTierImage = computed<string>(() => String(loyalty.value?.top_tier_image || ''))
  const points = computed<number>(() => toFiniteNumber(loyalty.value?.points))
  const profileInfo = computed(() => userData.value?.profile || null)
  const tiers = computed<LoyaltyTierRecord[]>(() => {
    const tierList = loyalty.value?.tiers
    return Array.isArray(tierList) ? tierList as LoyaltyTierRecord[] : []
  })

  // ========== 等级进度 ==========
  const tierInfo = computed(() => {
    const pts = points.value
    const tierList = tiers.value
    if (!tierList.length) return emptyTierInfo()

    let current: LoyaltyTierRecord | null = null
    let next: LoyaltyTierRecord | null = null

    for (let i = 0; i < tierList.length; i++) {
      const t = tierList[i]
      if (!t) continue
      const min = toFiniteNumber(t.min)
      const max = toFiniteNumber(t.max, -1)
      const inRange = (max === -1) ? (pts >= min) : (pts >= min && pts <= max)
      if (inRange) {
        current = t
        next = tierList[i + 1] || null
        break
      }
    }

    const firstTier = tierList[0]
    if (!current && firstTier) {
      current = firstTier
      next = tierList[1] || null
    }

    let pct = 100
    if (current) {
      if (next && Number(next.min) > 0) {
        const start = Number(current.min)
        const end = Number(next.min)
        pct = Math.max(0, Math.min(100, Math.floor(((pts - start) / (end - start)) * 100)))
      } else if (Number(current.max) !== -1) {
        const start = Number(current.min)
        const end = Number(current.max)
        pct = Math.max(0, Math.min(100, Math.floor(((pts - start) / Math.max(1, end - start)) * 100)))
      } else {
        pct = 100
      }
    }

    return { current, next, pct }
  })

  // ========== 等级配置 ==========
  const tierConfigs = ref<TierConfig[]>(defaultTierConfigs())
  const tierConfigsLoading = ref(false)

  const loadTierConfigs = async () => {
    tierConfigsLoading.value = true
    try {
      // GET /api/v1/marketing/loyalty/levels (公开配置)
      const response = await auth.request<any>('/marketing/loyalty/levels')
      const tiers = Array.isArray(response) ? response : response?.tiers

      if (Array.isArray(tiers) && tiers.length > 0) {
        tierConfigs.value = tiers.map(normalizeTierConfigFromBackend)
      } else {
        tierConfigs.value = defaultTierConfigs()
      }
    } catch (error) {
      logUnexpectedApiError('Failed to load tier configs:', error, isExpectedOptionalConfigMiss)
      tierConfigs.value = defaultTierConfigs()
    } finally {
      tierConfigsLoading.value = false
    }
  }

  // ========== 等级权益 ==========
  const levelDiscounts = computed(() => {
    const lvl = (levelName.value || '').toString().toLowerCase()
    if (!lvl || lvl === '—') return { discountRate: 0 }

    const config = tierConfigs.value.find(t => t.key === lvl)
    if (config) {
      return {
        discountRate: config.discount,
      }
    }

    return { discountRate: 0 }
  })

  const loyaltyRules = ref<LoyaltyRules | null>(null)
  const fetchLoyaltyRules = async () => {
    try {
      const data = await auth.request<any>('/marketing/loyalty/rules')
      const rawRules = (data?.data || data) as any
      loyaltyRules.value = normalizeLoyaltyRules(rawRules)
    } catch (error) {
      logUnexpectedApiError('Failed to fetch loyalty rules:', error, isExpectedOptionalConfigMiss)
      if (!loyaltyRules.value) {
        loyaltyRules.value = null
      }
    }
  }

  // ========== 登出 ==========
  const doLogout = async () => {
    try {
      await auth.logout()
    } catch (error) {
      console.error('Membership logout failed:', error)
    }
  }

  // ========== 初始化 ==========
  const initMembership = async (options: MembershipLoadOptions = {}) => {
    try {
      await auth.ensureSession()
    } catch (error) {
      console.error('Failed to initialize membership session:', error)
    }

    const tasks: Promise<unknown>[] = []

    const shouldLoadPublicConfig = options.includePublicConfig ?? auth.isAuthenticated.value
    if (shouldLoadPublicConfig) {
      tasks.push(
        loadTierConfigs(),
        fetchLoyaltyRules(),
      )
    }

    await Promise.allSettled(tasks)
  }

  // ========== 刷新数据 ==========
  const refreshData = async () => {
    try {
      await auth.ensureSession()
    } catch (error) {
      console.error('Failed to refresh membership session:', error)
    }

    const tasks: Promise<unknown>[] = []

    if (auth.isAuthenticated.value) {
      tasks.push(
        fetchLoyaltyRules(),
      )
    }

    await Promise.allSettled(tasks)
  }

  return {
    // 用户数据
    userData,
    isLogged,
    levelName,
    topTierImage,
    points,
    profileInfo,
    tiers,
    tierInfo,

    // 等级配置
    tierConfigs,
    tierConfigsLoading,
    loadTierConfigs,
    levelDiscounts,

    loyaltyRules,
    fetchLoyaltyRules,

    // 操作
    doLogout,
    initMembership,
    refreshData,

    // auth 透传
    auth
  }
}
