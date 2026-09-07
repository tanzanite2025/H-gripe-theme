<template>
  <div>
    <div class="warranty-card">
      <Icon name="lucide:shield-check" class="warranty-card__icon" aria-hidden="true" />
      <div class="warranty-card__content">
        <h3 class="warranty-card__title">{{ t('supportWarrantyCheck.title', 'Warranty Check') }}</h3>
        <p class="warranty-card__desc">
          {{ t('supportWarrantyCheck.cardDesc', 'Enter your order number to check shipped items, warranty time, and service history.') }}
        </p>
      </div>
      <NuxtLink :to="localePath('/support/warranty-check')" class="warranty-card__btn">
        {{ t('supportWarrantyCheck.checkNow', 'Check Now') }}
        <span class="arrow">→</span>
      </NuxtLink>
    </div>

    <section class="membership-section">
      <div class="membership-grid" :class="{ 'membership-grid--modal': isModal }">
        <div class="membership-col">
          <div class="member-header">
            <div class="member-avatar">
              <BadgeAvatar :logged="isLogged" :level="String(levelName)" :topTierImageUrl="String(topTierImage)" />
            </div>
            <div v-if="isLogged && profileInfo?.fullName" class="member-name">
              {{ profileInfo.fullName }}
            </div>
            <div v-if="isLogged" class="member-level">
              <span class="level-badge">{{ displayLevelName }}</span>
              <span class="level-points">{{ points }} {{ t('member.points.unit', 'pts') }}</span>
            </div>
            <div class="member-actions">
              <template v-if="!isLogged">
                <button class="btn-primary" type="button" @click="emit('open-auth', 'register')">
                  {{ t('user.register', 'Register') }}
                </button>
                <button class="btn-secondary" type="button" @click="emit('open-auth', 'login')">
                  {{ t('user.login', 'Login') }}
                </button>
              </template>
              <template v-else>
                <button class="btn-danger" type="button" @click="emit('logout')">
                  {{ t('user.logout', 'Logout') }}
                </button>
              </template>
            </div>
          </div>
        </div>

        <div class="membership-col">
          <div class="member-card member-benefits-card">
            <h4 class="card-title">{{ t('resourcesMembershipMyInfo.benefits.title', 'My Benefits') }}</h4>
            <div class="member-stats">
              <div class="stat-item">
                <Icon name="lucide:tag" class="stat-icon" aria-hidden="true" />
                <div class="stat-content">
                  <span class="stat-copy">
                    <span class="stat-label">{{ t('resourcesMembershipMyInfo.benefits.level', 'Level') }}</span>
                    <span class="stat-desc">
                      {{ t('resourcesMembershipMyInfo.benefits.levelDesc', 'Your membership tier reflects accumulated activity and unlocks the benefit rules connected to that tier.') }}
                    </span>
                  </span>
                  <span class="stat-value" :class="{ highlight: !isLogged }">
                    {{ isLogged ? displayLevelName : '?' }}
                  </span>
                </div>
              </div>
              <div class="stat-item">
                <Icon name="lucide:badge-percent" class="stat-icon" aria-hidden="true" />
                <div class="stat-content">
                  <span class="stat-copy">
                    <span class="stat-label">{{ t('resourcesMembershipMyInfo.benefits.discountRate', 'Discount Rate') }}</span>
                    <span class="stat-desc">
                      {{ t('resourcesMembershipMyInfo.benefits.discountRateDesc', 'The member-level price discount configured in the backend.') }}
                    </span>
                  </span>
                  <span class="stat-value" :class="{ highlight: !isLogged }">
                    {{ isLogged ? formatDiscountRate(levelDiscounts.discountRate) : '?' }}
                  </span>
                </div>
              </div>
            </div>

            <div class="member-assets">
              <div class="asset-item">
                <Icon name="lucide:ticket-percent" class="asset-icon" aria-hidden="true" />
                <div class="asset-content">
                  <span class="asset-label">{{ t('resourcesMembershipMyInfo.benefits.coupons', 'Coupons') }}</span>
                  <span class="asset-value">{{ isLogged ? `× ${userCoupons}` : '?' }}</span>
                </div>
              </div>
              <div class="asset-item">
                <Icon name="lucide:credit-card" class="asset-icon" aria-hidden="true" />
                <div class="asset-content">
                  <span class="asset-label">{{ t('resourcesMembershipMyInfo.benefits.giftCards', 'Gift Cards') }}</span>
                  <span class="asset-value">{{ isLogged ? `× ${userPointCards}` : '?' }}</span>
                </div>
              </div>
            </div>

            <div v-if="isLogged" class="tier-progress">
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: `${tierInfo.pct}%` }"></div>
              </div>
              <div class="progress-labels">
                <span>{{ tierInfo.current ? tierInfo.current.min : 0 }}</span>
                <span class="progress-pct">{{ tierInfo.pct }}%</span>
                <span>{{ tierInfo.next ? tierInfo.next.min : 'MAX' }}</span>
              </div>
            </div>
          </div>
        </div>

        <div v-if="isLogged && profileInfo" class="membership-col membership-col--full">
          <div class="profile-info">
            <h4>{{ t('profile.title', 'Profile') }}</h4>
            <div v-if="profileInfo.fullName" class="profile-item">
              <span class="profile-label">{{ t('profile.fullName', 'Full Name') }}</span>
              <span class="profile-value">{{ profileInfo.fullName }}</span>
            </div>
            <div v-if="profileInfo.company" class="profile-item">
              <span class="profile-label">{{ t('profile.company', 'Company') }}</span>
              <span class="profile-value">{{ profileInfo.company }}</span>
            </div>
            <div v-if="profileInfo.country" class="profile-item">
              <span class="profile-label">{{ t('profile.country', 'Country/Region') }}</span>
              <span class="profile-value">{{ profileInfo.country }}</span>
            </div>
            <div v-if="profileInfo.phone" class="profile-item">
              <span class="profile-label">{{ t('profile.phone', 'Phone') }}</span>
              <span class="profile-value">{{ profileInfo.phone }}</span>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import BadgeAvatar from '~/components/BadgeAvatar.vue'
import { usePageMessages } from '~/composables/usePageMessages'

interface ProfileInfo {
  fullName?: string
  company?: string
  country?: string
  phone?: string
}

interface TierRecord {
  min?: number | string | null
}

interface TierInfo {
  current: TierRecord | null
  next: TierRecord | null
  pct: number
}

interface LevelDiscounts {
  discountRate: number
}

const props = withDefaults(defineProps<{
  isModal?: boolean
  isLogged: boolean
  levelName: string
  topTierImage: string
  points: number
  profileInfo?: ProfileInfo | null
  tierInfo: TierInfo
  levelDiscounts: LevelDiscounts
  userCoupons: number
  userPointCards: number
}>(), {
  isModal: false,
  profileInfo: null,
})

const emit = defineEmits<{
  (event: 'open-auth', mode: 'login' | 'register'): void
  (event: 'logout'): void
}>()

const { locale, t } = useI18n()
const localePath = useLocalePath()
const membershipMessages = usePageMessages('resourcesMembershipMyInfo')
const warrantyMessages = usePageMessages('supportWarrantyCheck')

await Promise.all([
  membershipMessages.loadPageMessages(locale.value),
  warrantyMessages.loadPageMessages(locale.value),
])

watch(locale, (nextLocale) => {
  void Promise.all([
    membershipMessages.loadPageMessages(nextLocale),
    warrantyMessages.loadPageMessages(nextLocale),
  ])
})

const normalizeLevelKey = (value: string) =>
  value.toLowerCase().trim().replace(/\s+/g, '-')

const displayLevelName = computed(() => {
  const rawLevel = String(props.levelName || '')
  if (!rawLevel || rawLevel === '—') return rawLevel
  return t(`resourcesMembershipMyInfo.levels.rows.${normalizeLevelKey(rawLevel)}`, rawLevel)
})

const formatBenefitNumber = (value: number) => {
  const numericValue = Number(value)
  if (!Number.isFinite(numericValue)) return '0'
  return Number.isInteger(numericValue)
    ? String(numericValue)
    : numericValue.toFixed(2).replace(/\.?0+$/, '')
}

const formatDiscountRate = (value: number) => `${formatBenefitNumber(value)}%`
</script>
