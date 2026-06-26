/**
 * 会员等级配置 - 与后端保持一致
 */
import type { MemberTier } from '../types/cart-calculation-types'

export const MEMBER_TIERS: Record<string, MemberTier> = {
  ordinary: { name: 'Ordinary', min: 0, max: 499, discount: 0 },
  bronze: { name: 'Bronze', min: 500, max: 1999, discount: 5 },
  silver: { name: 'Silver', min: 2000, max: 4999, discount: 10 },
  gold: { name: 'Gold', min: 5000, max: 9999, discount: 15 },
  platinum: { name: 'Platinum', min: 10000, max: null, discount: 20 },
}

/**
 * 根据积分获取会员等级
 */
export const getTierByPoints = (points: number): MemberTier => {
  const defaultTier: MemberTier = { name: 'Ordinary', min: 0, max: 499, discount: 0 }

  for (const [key, tier] of Object.entries(MEMBER_TIERS)) {
    if (tier.max === null) {
      if (points >= tier.min) return tier
    } else {
      if (points >= tier.min && points <= tier.max) return tier
    }
  }

  return MEMBER_TIERS.ordinary ?? defaultTier
}
