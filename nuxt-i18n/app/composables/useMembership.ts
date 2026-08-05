/**
 * 会员中心数据管理 composable
 * 统一管理会员等级、积分、优惠券、礼品卡等数据
 */
import { ref, computed, onMounted } from 'vue'
import { useAuth } from '~/composables/useAuth'

// 等级配置类型
interface TierConfig {
  key: string
  name: string
  min: number
  max: number | null
  discount: number
}

// 礼品卡类型
interface RedeemGiftCardOption {
  id: number
  label: string
  giftcard_value: number
  giftcard_value_cents: number
  currency: string
  points_required: number
  remaining_quantity: number
  status: string
  cover_image?: string
}

interface UserGiftCard {
  id: number
  code: string
  initial_value: number
  balance: number
  initial_value_cents: number
  balance_cents: number
  currency: string
  status: string
  expires_at?: string | null
  created_at: string
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
  redemption_exchange_rate: number | null
}

interface LoyaltyProgramConfig {
  id: number
  version: number
  enabled: boolean
  currency: string
  points_base_currency?: string
  purchase_earn_points_per_currency_unit?: number
  exchange_rate_points: number
  min_redeem_points: number
  max_value_per_day_cents: number
  card_expiry_days: number
  referral_referrer_points: number
  referral_referee_points: number
  checkin_base_points: number
  checkin_streak_interval_days: number
  checkin_streak_bonus_points: number
  checkin_max_points: number
  redeem_options: RedeemGiftCardOption[]
}

type LoyaltyRecord = Record<string, unknown>
type LoyaltyTierRecord = LoyaltyRecord & {
  min?: number | string | null
  max?: number | string | null
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
    redemption_exchange_rate: toNullableNumber(raw.redemption_exchange_rate ?? raw.exchange_rate_points)
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
    discount: toFiniteNumber(tier?.discount_rate ?? tier?.discount),
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
      console.error('Failed to load tier configs:', error)
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

  // ========== 用户资产（优惠券、积分卡） ==========
  const userCoupons = ref(0)
  const userPointCards = ref(0)
  const assetsLoading = ref(false)

  const fetchUserAssets = async () => {
    if (!isLogged.value) {
      userCoupons.value = 0
      userPointCards.value = 0
      return
    }

    assetsLoading.value = true
    try {
      const data = await auth.request<any>('/marketing/loyalty/assets')
      if (data) {
        userCoupons.value = data.coupons || 0
        userPointCards.value = data.point_cards || 0
      }
    } catch (error) {
      console.error('获取用户资产失败:', error)
    } finally {
      assetsLoading.value = false
    }
  }

  // ========== 礼品卡 ==========
  const availableGiftcards = ref<RedeemGiftCardOption[]>([])
  const loyaltyProgramConfig = ref<LoyaltyProgramConfig | null>(null)
  const loyaltyRules = ref<LoyaltyRules | null>(null)
  const userGiftCards = ref<UserGiftCard[]>([])
  const giftcardsLoading = ref(false)
  const giftcardsError = ref('')
  const redeemingCardId = ref<number | null>(null)
  const redeemMessage = ref('')
  const redeemSuccess = ref(false)

  const fetchLoyaltyProgramConfig = async () => {
    giftcardsLoading.value = true
    giftcardsError.value = ''

    try {
      const data = await auth.request<any>('/marketing/loyalty/config')
      const rawConfig = (data?.data || data) as any
      const normalizedOptions: RedeemGiftCardOption[] = Array.isArray(rawConfig?.redeem_options)
        ? rawConfig.redeem_options.map((option: any) => ({
            id: Number(option.id),
            label: String(option.label || ''),
            giftcard_value: Number(option.value ?? option.giftcard_value ?? 0),
            giftcard_value_cents: Number(option.value_cents ?? option.giftcard_value_cents ?? 0),
            currency: String(option.currency || rawConfig?.currency || ''),
            points_required: Number(option.points_required || 0),
            remaining_quantity: Number(option.remaining_quantity ?? 0),
            status: String(option.status || 'active')
          }))
        : []
      const config = rawConfig
        ? { ...rawConfig, redeem_options: normalizedOptions } as LoyaltyProgramConfig
        : null
      loyaltyProgramConfig.value = config
      availableGiftcards.value = normalizedOptions
        ? normalizedOptions.filter((card) => card.status === 'active' && card.remaining_quantity > 0)
        : []
      if (!loyaltyRules.value) {
        loyaltyRules.value = config
          ? normalizeLoyaltyRules(config)
          : null
      }
    } catch (error) {
      console.error('Failed to fetch loyalty program config:', error)
      giftcardsError.value = 'Network error'
    } finally {
      giftcardsLoading.value = false
    }
  }

  const fetchAvailableGiftcards = fetchLoyaltyProgramConfig

  const fetchLoyaltyRules = async () => {
    try {
      const data = await auth.request<any>('/marketing/loyalty/rules')
      const rawRules = (data?.data || data) as any
      loyaltyRules.value = normalizeLoyaltyRules(rawRules)
    } catch (error) {
      console.error('Failed to fetch loyalty rules:', error)
      if (!loyaltyRules.value) {
        loyaltyRules.value = null
      }
    }
  }

  const fetchUserGiftCards = async () => {
    if (!isLogged.value) {
      userGiftCards.value = []
      return
    }
    try {
      const data = await auth.request<any>('/marketing/loyalty/gift-cards')
      userGiftCards.value = Array.isArray(data?.gift_cards) ? data.gift_cards : []
    } catch (error) {
      console.error('Failed to fetch user gift cards:', error)
      userGiftCards.value = []
    }
  }

  const handleRedeemGiftcard = async (card: RedeemGiftCardOption) => {
    if (redeemingCardId.value) return

    if (!isLogged.value) {
      redeemSuccess.value = false
      redeemMessage.value = 'Please login to redeem gift cards'
      setTimeout(() => { redeemMessage.value = '' }, 3000)
      return
    }

    redeemingCardId.value = card.id
    redeemMessage.value = ''
    redeemSuccess.value = false

    try {
      const idempotencyKey = typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
        ? crypto.randomUUID()
        : `${Date.now()}-${Math.random().toString(36).slice(2)}`
      const data = await auth.request<any>('/marketing/loyalty/redeem', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Idempotency-Key': idempotencyKey
        },
        body: JSON.stringify({
          option_id: Number(card.id),
          idempotency_key: idempotencyKey
        })
      })

      if (data && (data.success || data.card_code)) {
        redeemSuccess.value = true
        redeemMessage.value = `Redeemed successfully! Card code: ${data.card_code}`

        await auth.ensureSession()
        await fetchLoyaltyProgramConfig()
        await fetchUserAssets()
        await fetchUserGiftCards()

        setTimeout(() => { redeemMessage.value = '' }, 3000)
      } else {
        redeemSuccess.value = false
        redeemMessage.value = data.message || 'Redemption failed'
      }
    } catch (error) {
      console.error('Failed to redeem gift card:', error)
      redeemSuccess.value = false
      redeemMessage.value = 'Network error, please try again later'
    } finally {
      redeemingCardId.value = null
    }
  }

  // ========== 登出 ==========
  const doLogout = async () => {
    try {
      await auth.logout()
    } catch { }
  }

  // ========== 初始化 ==========
  const initMembership = async () => {
    try {
      await auth.ensureSession()
    } catch (error) {
      console.error('Failed to initialize membership session:', error)
    }

    await Promise.allSettled([
      loadTierConfigs(),
      fetchUserAssets(),
      fetchLoyaltyProgramConfig(),
      fetchLoyaltyRules(),
      fetchUserGiftCards()
    ])
  }

  // ========== 刷新数据 ==========
  const refreshData = async () => {
    try {
      await auth.ensureSession()
    } catch (error) {
      console.error('Failed to refresh membership session:', error)
    }

    await Promise.allSettled([
      fetchUserAssets(),
      fetchLoyaltyProgramConfig(),
      fetchLoyaltyRules(),
      fetchUserGiftCards()
    ])
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

    // 用户资产
    userCoupons,
    userPointCards,
    assetsLoading,
    fetchUserAssets,

    // 礼品卡
    availableGiftcards,
    loyaltyProgramConfig,
    loyaltyRules,
    userGiftCards,
    giftcardsLoading,
    giftcardsError,
    redeemingCardId,
    redeemMessage,
    redeemSuccess,
    fetchAvailableGiftcards,
    fetchLoyaltyProgramConfig,
    fetchLoyaltyRules,
    fetchUserGiftCards,
    handleRedeemGiftcard,

    // 操作
    doLogout,
    initMembership,
    refreshData,

    // auth 透传
    auth
  }
}
