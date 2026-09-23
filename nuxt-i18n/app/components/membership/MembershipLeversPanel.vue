<template>
  <section class="company-section membership-section membership-levers">
    <div class="membership-details">
      <div class="tier-table">
        <h4>{{ t('resourcesMembershipLevers.levels.title', 'Membership Levels') }}</h4>
        <div class="table-wrapper">
          <table>
            <thead>
              <tr>
                <th>{{ t('resourcesMembershipLevers.levels.header.level', 'Level') }}</th>
                <th>{{ t('resourcesMembershipLevers.levels.header.pointsRequired', 'Points Required') }}</th>
                <th>{{ t('resourcesMembershipLevers.levels.header.discountRate', 'Discount Rate') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="tier in tierConfigs" :key="tier.key">
                <td>{{ displayTierName(tier) }}</td>
                <td>{{ tier.min }}{{ tier.max !== null ? `–${tier.max}` : '+' }}</td>
                <td>{{ formatDiscountRate(tier.discount) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="points-rules">
        <h4>{{ t('resourcesMembershipLevers.points.title', 'Points Rules') }}</h4>
        <div class="rule-list">
          <div v-for="rule in pointRuleItems" :key="rule.key" class="rule-item">
            <div class="rule-title">{{ rule.title }}</div>
            <div class="rule-desc">{{ rule.description }}</div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

interface TierConfig {
  key: string
  name: string
  min: number
  max: number | null
  discount: number
}

interface LoyaltyRules {
  points_base_currency: string
  purchase_earn_points_per_currency_unit: number | null
  referral_referrer_points: number | null
  referral_referee_points: number | null
  checkin_base_points: number | null
  checkin_streak_interval_days: number | null
  checkin_streak_bonus_points: number | null
  checkin_max_points: number | null
  points_exchange_rate: number | null
}

const props = defineProps<{
  tierConfigs: TierConfig[]
  loyaltyRules: LoyaltyRules | null
}>()

const { locale, t, te } = useI18n()
const { loadPageMessages } = usePageMessages('resourcesMembershipLevers')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

const normalizeLevelKey = (value: string) =>
  value.toLowerCase().trim().replace(/\s+/g, '-')

const displayTierName = (tier: TierConfig) =>
  t(`resourcesMembershipLevers.levels.rows.${normalizeLevelKey(tier.key)}`, tier.name)

const formatBenefitNumber = (value: number) => {
  const numericValue = Number(value)
  if (!Number.isFinite(numericValue)) return '0'
  return Number.isInteger(numericValue)
    ? String(numericValue)
    : numericValue.toFixed(2).replace(/\.?0+$/, '')
}

const formatDiscountRate = (value: number) => `${formatBenefitNumber(value)}%`

const hasRuleNumber = (value: number | null | undefined): value is number =>
  typeof value === 'number' && Number.isFinite(value)

const formatRuleNumber = (value: number) =>
  Number.isInteger(value) ? String(value) : value.toFixed(2).replace(/\.?0+$/, '')

const translateWithFallback = (
  key: string,
  params: Record<string, string | number>,
  fallback: string,
) => te(key) ? t(key, params) : fallback

const formatPoints = (value: number) => {
  const points = formatRuleNumber(value)
  return value === 1
    ? translateWithFallback(
        'resourcesMembershipLevers.points.pointValueSingular',
        { points },
        `${points} Point`,
      )
    : translateWithFallback(
        'resourcesMembershipLevers.points.pointsValue',
        { points },
        `${points} Points`,
      )
}

const notConfiguredText = computed(() =>
  t('resourcesMembershipLevers.points.notConfigured', 'Not configured'),
)

const pointsBaseCurrency = computed(() =>
  String(props.loyaltyRules?.points_base_currency || 'USD').trim().toUpperCase() || 'USD',
)

const purchaseEarnRuleDescription = computed(() => {
  const pointsPerUnit = props.loyaltyRules?.purchase_earn_points_per_currency_unit
  if (!hasRuleNumber(pointsPerUnit)) return notConfiguredText.value
  if (pointsPerUnit <= 0) {
    return t(
      'resourcesMembershipLevers.points.purchaseEarnDisabled',
      'Purchase earning is not enabled',
    )
  }

  return translateWithFallback(
    'resourcesMembershipLevers.points.purchaseEarnRule',
    {
      points: formatPoints(pointsPerUnit),
      currency: pointsBaseCurrency.value,
    },
    `${formatPoints(pointsPerUnit)} per 1 ${pointsBaseCurrency.value} of product amount after discounts, awarded after order completion`,
  )
})

const redemptionRuleDescription = computed(() => {
  const exchangeRate = props.loyaltyRules?.points_exchange_rate
  if (!hasRuleNumber(exchangeRate)) return notConfiguredText.value
  if (exchangeRate <= 0) {
    return t(
      'resourcesMembershipLevers.points.redemptionDisabled',
      'Points redemption is not enabled',
    )
  }

  return translateWithFallback(
    'resourcesMembershipLevers.points.redemptionDisplayRule',
    {
      points: formatPoints(exchangeRate),
      amount: 1,
      currency: pointsBaseCurrency.value,
    },
    `${formatPoints(exchangeRate)} = ${pointsBaseCurrency.value} 1`,
  )
})

const checkInRuleDescription = computed(() => {
  const base = props.loyaltyRules?.checkin_base_points
  if (!hasRuleNumber(base)) return notConfiguredText.value
  if (base <= 0) {
    return t(
      'resourcesMembershipLevers.points.checkinDisabled',
      'Daily check-in points are not enabled',
    )
  }

  const parts = [
    translateWithFallback(
      'resourcesMembershipLevers.points.checkinBaseRule',
      { base: formatPoints(base) },
      `${formatPoints(base)} per check-in`,
    ),
  ]
  const max = props.loyaltyRules?.checkin_max_points
  const bonus = props.loyaltyRules?.checkin_streak_bonus_points
  const interval = props.loyaltyRules?.checkin_streak_interval_days

  if (hasRuleNumber(max) && max > base) {
    parts.push(translateWithFallback(
      'resourcesMembershipLevers.points.checkinMaxRule',
      { max: formatPoints(max) },
      `up to ${formatPoints(max)}`,
    ))
  }

  if (hasRuleNumber(bonus) && hasRuleNumber(interval) && bonus > 0 && interval > 0) {
    parts.push(translateWithFallback(
      'resourcesMembershipLevers.points.checkinStreakRule',
      { bonus: formatPoints(bonus), interval: formatRuleNumber(interval) },
      `+${formatPoints(bonus)} every ${formatRuleNumber(interval)} consecutive days`,
    ))
  }

  return parts.join('; ')
})

const pointRuleItems = computed(() => [
  {
    key: 'purchase',
    title: t('resourcesMembershipLevers.points.purchaseEarn', 'Order completion'),
    description: purchaseEarnRuleDescription.value,
  },
  {
    key: 'redemption',
    title: t('resourcesMembershipLevers.points.redeem', 'Redemption rate'),
    description: redemptionRuleDescription.value,
  },
  {
    key: 'checkin',
    title: t('resourcesMembershipLevers.points.checkin', 'Daily check-in'),
    description: checkInRuleDescription.value,
  },
])
</script>
