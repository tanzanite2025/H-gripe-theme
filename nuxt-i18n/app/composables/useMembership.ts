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
interface GiftCard {
  id: number
  card_code: string
  balance: string
  points_spent: number
  status: string
  cover_image?: string
}

export function useMembership() {
  const auth = useAuth()

  // ========== 用户数据 ==========
  const userData = computed(() => auth.user.value)
  const isLogged = computed(() => !!userData.value)
  const levelName = computed(() => userData.value?.loyalty?.level || '—')
  const topTierImage = computed(() => userData.value?.loyalty?.top_tier_image || '')
  const points = computed(() => userData.value?.loyalty?.points ?? 0)
  const profileInfo = computed(() => userData.value?.profile || null)
  const tiers = computed(() => (userData.value?.loyalty?.tiers || []) as any[])

  // ========== 等级进度 ==========
  const tierInfo = computed(() => {
    const pts = points.value
    const tierList = tiers.value as any[]
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
      const response = await $fetch<{ tiers?: Record<string, any> }>('/wp-json/tanzanite/v1/loyalty/settings')
      if (response?.tiers) {
        tierConfigs.value = Object.entries(response.tiers).map(([key, config]) => ({
          key,
          name: config.name,
          min: config.min,
          max: config.max,
          discount: config.discount,
          pointsDiscount: config.points_discount || 0,
          stackable: config.stackable !== false
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
      const base = typeof window !== 'undefined' ? window.location.origin : ''
      const res = await fetch(`${base}/wp-json/mytheme/v1/user/assets`, {
        method: 'GET',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' }
      })
      
      if (res.ok) {
        const data = await res.json()
        if (data.success) {
          userCoupons.value = data.data?.coupons || 0
          userPointCards.value = data.data?.point_cards || 0
        }
      }
    } catch (error) {
      console.error('获取用户资产失败:', error)
    } finally {
      assetsLoading.value = false
    }
  }

  // ========== 礼品卡 ==========
  const availableGiftcards = ref<GiftCard[]>([])
  const giftcardsLoading = ref(false)
  const giftcardsError = ref('')
  const redeemingCardId = ref<number | null>(null)
  const redeemMessage = ref('')
  const redeemSuccess = ref(false)

  const fetchAvailableGiftcards = async () => {
    giftcardsLoading.value = true
    giftcardsError.value = ''
    
    try {
      const base = typeof window !== 'undefined' ? window.location.origin : ''
      const res = await fetch(`${base}/wp-json/tanzanite/v1/giftcards`, {
        method: 'GET',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' }
      })
      
      if (res.ok) {
        const data = await res.json()
        const allCards = data.items || data || []
        availableGiftcards.value = allCards.filter((card: any) => card.status === 'active')
      } else {
        giftcardsError.value = 'Failed to load gift cards'
      }
    } catch (error) {
      console.error('Failed to fetch gift cards:', error)
      giftcardsError.value = 'Network error'
    } finally {
      giftcardsLoading.value = false
    }
  }

  const handleRedeemGiftcard = async (card: GiftCard) => {
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
      const base = typeof window !== 'undefined' ? window.location.origin : ''
      const res = await fetch(`${base}/wp-json/tanzanite/v1/redeem/exchange`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          points_to_spend: card.points_spent,
          giftcard_value: parseFloat(card.balance)
        })
      })
      
      const data = await res.json()
      
      if (res.ok && data.success) {
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
      const base = typeof window !== 'undefined' ? window.location.origin : ''
      const res = await fetch(`${base}/wp-json/tanzanite/v1/loyalty/referral/generate`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' }
      })
      const data = await res.json()
      if (!res.ok) throw new Error((data && data.message) || 'Failed to generate referral link')
      const url = String(data && data.url)
      if (typeof navigator !== 'undefined' && navigator.share) {
        try { await navigator.share({ url }) } catch {}
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
    } catch {}
  }

  // ========== 初始化 ==========
  const initMembership = async () => {
    await auth.ensureSession()
    await Promise.all([
      loadTierConfigs(),
      fetchUserAssets(),
      fetchAvailableGiftcards()
    ])
  }

  // ========== 刷新数据 ==========
  const refreshData = async () => {
    await auth.ensureSession()
    await fetchUserAssets()
    await fetchAvailableGiftcards()
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
