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
  points_required: number
  status: string
  cover_image?: string
}

const emptyTierInfo = () => ({ current: null, next: null, pct: 0 })

export function useMembership() {
  const auth = useAuth()

  // ========== 用户数据 ==========
  const userData = computed(() => auth.user.value)
  const isLogged = computed(() => !!userData.value)
  const levelName = computed(() => userData.value?.loyalty?.level || '—')
  const topTierImage = computed(() => userData.value?.loyalty?.top_tier_image || '')
  const points = computed(() => userData.value?.loyalty?.points ?? 0)
  const profileInfo = computed(() => userData.value?.profile || null)
  const tiers = computed(() => {
    const tierList = userData.value?.loyalty?.tiers
    return Array.isArray(tierList) ? tierList : []
  })

  // ========== 等级进度 ==========
  const tierInfo = computed(() => {
    const pts = points.value
    const tierList = tiers.value as any[]
    if (!tierList.length) return emptyTierInfo()

    let current: any = null
    let next: any = null

    for (let i = 0; i < tierList.length; i++) {
      const t = tierList[i]
      const min = Number(t.min)
      const max = Number(t.max)
      const inRange = (max === -1) ? (pts >= min) : (pts >= min && pts <= max)
      if (inRange) {
        current = t
        next = tierList[i + 1] || null
        break
      }
    }

    if (!current && tierList.length) {
      current = tierList[0]
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
        tierConfigs.value = tiers.map((tier: any) => ({
          key: tier.key,
          name: tier.name ?? tier.label ?? String(tier.key || '').toUpperCase(),
          min: Number(tier.min ?? 0),
          max: typeof tier.max === 'number' ? (tier.max === -1 ? null : tier.max) : null,
          discount: Number(tier.discount ?? 0),
          // 这里将积分折扣近似映射为可用积分抵扣的最大订单百分比
          pointsDiscount: Number(tier.redeem?.percent_of_total ?? 0),
          // 是否允许与百分比折扣叠加
          stackable: tier.redeem?.stack_with_percent ?? true,
        }))
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
  const giftcardsLoading = ref(false)
  const giftcardsError = ref('')
  const redeemingCardId = ref<number | null>(null)
  const redeemMessage = ref('')
  const redeemSuccess = ref(false)

  const fetchAvailableGiftcards = async () => {
    giftcardsLoading.value = true
    giftcardsError.value = ''

    try {
      const data = await auth.request<any>('/marketing/loyalty/redeem-options')
      const allCards = Array.isArray(data?.items) ? data.items : (Array.isArray(data) ? data : [])
      availableGiftcards.value = allCards.filter((card: any) => card.status === 'active')
    } catch (error) {
      console.error('Failed to fetch gift cards:', error)
      giftcardsError.value = 'Network error'
    } finally {
      giftcardsLoading.value = false
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
      const data = await auth.request<any>('/marketing/loyalty/redeem', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          giftcard_value: Number(card.giftcard_value)
        })
      })

      if (data && (data.success || data.card_code)) {
        redeemSuccess.value = true
        redeemMessage.value = `Redeemed successfully! Card code: ${data.card_code}`

        await auth.ensureSession()
        await fetchAvailableGiftcards()
        await fetchUserAssets()

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

  // ========== 邀请链接 ==========
  const inviteLoading = ref(false)
  const inviteMsg = ref('')

  const handleCopyInviteLink = async () => {
    try {
      inviteLoading.value = true
      inviteMsg.value = ''
      const data = await auth.request<any>('/marketing/loyalty/referral', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' }
      })
      if (!data || data.error) throw new Error((data && data.message) || 'Failed to generate referral link')
      const url = String(data && data.url)
      if (typeof navigator !== 'undefined' && navigator.share) {
        try { await navigator.share({ url }) } catch { }
      }
      if (typeof navigator !== 'undefined') {
        await navigator.clipboard.writeText(url)
      }
      inviteMsg.value = 'Invitation link copied'
    } catch (e) {
      inviteMsg.value = String(e instanceof Error ? e.message : 'Failed to generate referral link')
    } finally {
      inviteLoading.value = false
      setTimeout(() => { inviteMsg.value = '' }, 15000)
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
      fetchAvailableGiftcards()
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
      fetchAvailableGiftcards()
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
    giftcardsLoading,
    giftcardsError,
    redeemingCardId,
    redeemMessage,
    redeemSuccess,
    fetchAvailableGiftcards,
    handleRedeemGiftcard,

    // 邀请
    inviteLoading,
    inviteMsg,
    handleCopyInviteLink,

    // 操作
    doLogout,
    initMembership,
    refreshData,

    // auth 透传
    auth
  }
}
