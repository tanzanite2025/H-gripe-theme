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
  pointsDiscount: number
  stackable: boolean
}

// 礼品卡类型
interface RedeemGiftCardOption {
  id: number
  label: string
  giftcard_value: number
  giftcard_value_cents: number
  points_required: number
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
  referral_referrer_points: number
  referral_referee_points: number
  checkin_base_points: number
  checkin_streak_interval_days: number
  checkin_streak_bonus_points: number
  checkin_max_points: number
  redemption_exchange_rate: number
}

interface LoyaltyProgramConfig {
  id: number
  version: number
  enabled: boolean
  currency: string
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
    pointsDiscount: toFiniteNumber(tier?.points_discount ?? tier?.redeem?.percent_of_total),
    stackable: tier?.redeem?.stack_with_percent ?? tier?.points_discount_stackable ?? true,
  }
}

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
  const tierConfigs = ref<TierConfig[]>([])
  const tierConfigsLoading = ref(false)

  const loadTierConfigs = async () => {
    tierConfigsLoading.value = true
    try {
      // GET /api/v1/marketing/loyalty/levels (公开配置)
      const response = await auth.request<any>('/marketing/loyalty/levels')
      const tiers = Array.isArray(response) ? response : response?.tiers

      if (Array.isArray(tiers)) {
        tierConfigs.value = tiers.map(normalizeTierConfigFromBackend)
      }
    } catch (error) {
      console.error('Failed to load tier configs:', error)
    } finally {
      tierConfigsLoading.value = false
    }
  }

  // ========== 等级折扣 ==========
  const levelDiscounts = computed(() => {
    const lvl = (levelName.value || '').toString().toLowerCase()
    if (!lvl || lvl === '—') return { product: 0, points: 0, stackable: false }

    const config = tierConfigs.value.find(t => t.key === lvl)
    if (config) {
      return {
        product: config.discount,
        points: config.pointsDiscount,
        stackable: config.stackable
      }
    }

    return { product: 0, points: 0, stackable: false }
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
            points_required: Number(option.points_required || 0),
            status: String(option.status || 'active')
          }))
        : []
      const config = rawConfig
        ? { ...rawConfig, redeem_options: normalizedOptions } as LoyaltyProgramConfig
        : null
      loyaltyProgramConfig.value = config
      availableGiftcards.value = normalizedOptions
        ? normalizedOptions.filter((card) => card.status === 'active')
        : []
      loyaltyRules.value = config
        ? {
            version: Number(config.version || 0),
            referral_referrer_points: Number(config.referral_referrer_points || 0),
            referral_referee_points: Number(config.referral_referee_points || 0),
            checkin_base_points: Number(config.checkin_base_points || 0),
            checkin_streak_interval_days: Number(config.checkin_streak_interval_days || 0),
            checkin_streak_bonus_points: Number(config.checkin_streak_bonus_points || 0),
            checkin_max_points: Number(config.checkin_max_points || 0),
            redemption_exchange_rate: Number(config.exchange_rate_points || 0)
          }
        : null
    } catch (error) {
      console.error('Failed to fetch loyalty program config:', error)
      giftcardsError.value = 'Network error'
    } finally {
      giftcardsLoading.value = false
    }
  }

  const fetchAvailableGiftcards = fetchLoyaltyProgramConfig

  const fetchLoyaltyRules = async () => {
    await fetchLoyaltyProgramConfig()
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
