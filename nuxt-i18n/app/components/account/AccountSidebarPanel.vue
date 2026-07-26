<template>
  <section class="account-sidebar" aria-label="Account sidebar">
    <header class="account-sidebar__hero">
      <div class="account-sidebar__hero-main">
        <p class="account-sidebar__kicker">
          {{ t('accountSidebar.kicker', 'Account Hub') }}
        </p>
        <h2>
          {{ isAuthenticated ? displayName : t('accountSidebar.guestTitle', 'Sign in to continue') }}
        </h2>
        <p>
          {{ isAuthenticated
            ? t('accountSidebar.memberSubtitle', 'Points, wishlist, cart and checkout details are grouped here.')
            : t('accountSidebar.guestSubtitle', 'This Dock entry is now reserved for your personal account area.') }}
        </p>
      </div>

      <button v-if="isAuthenticated" type="button" class="account-sidebar__logout" @click="logout">
        <Icon name="lucide:log-out" />
        {{ t('user.logout', 'Logout') }}
      </button>
    </header>

    <AccountLoginPrompt
      v-if="!isAuthenticated"
      @login="openAuthForm('login')"
      @register="openAuthForm('register')"
    />

    <template v-else>
      <div class="account-sidebar__stats" aria-label="Account summary">
        <div>
          <span>{{ t('member.points.unit', 'pts') }}</span>
          <strong>{{ pointsNumber }}</strong>
        </div>
        <div>
          <span>{{ t('accountSidebar.summary.saved', 'Saved') }}</span>
          <strong>{{ wishlistCount }}</strong>
        </div>
        <div>
          <span>{{ t('accountSidebar.summary.cart', 'Cart') }}</span>
          <strong>{{ cartCount }}</strong>
        </div>
      </div>

      <nav class="account-sidebar__tabs" role="tablist" :aria-label="t('accountSidebar.tabs.ariaLabel', 'Account sections')">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          type="button"
          class="account-sidebar__tab"
          :class="{ 'account-sidebar__tab--active': activeTab === tab.id }"
          role="tab"
          :aria-selected="activeTab === tab.id"
          @click="activeTab = tab.id"
        >
          <Icon :name="tab.icon" />
          <span>{{ t(tab.labelKey, tab.label) }}</span>
        </button>
      </nav>

      <div class="account-sidebar__content">
        <AccountPointsTab
          v-show="activeTab === 'points'"
          :level-name="String(levelName || '—')"
          :points="pointsNumber"
          :tier-info="tierInfo"
          :level-discounts="levelDiscounts"
          :coupons="userCoupons"
          :point-cards="userPointCards"
          :loading="membershipLoading"
          @refresh="refreshData"
          @close="closeSidebar"
        />

        <AccountWishlistTab
          v-show="activeTab === 'wishlist'"
          :active="activeTab === 'wishlist'"
          @close="closeSidebar"
        />

        <AccountCartTab
          v-show="activeTab === 'cart'"
          @close="closeSidebar"
        />

        <AccountAddressesTab
          v-show="activeTab === 'addresses'"
          @close="closeSidebar"
        />
      </div>
    </template>

    <LazyAuthModal
      v-model="showAuthModal"
      :default-mode="authMode"
      embedded
      @mode-change="authMode = $event"
      @success="handleAuthSuccess"
    />
  </section>
</template>

<script setup lang="ts">
import { computed, inject, onMounted, ref, watch } from 'vue'
import { useI18n } from '#imports'
import AccountAddressesTab from '~/components/account/AccountAddressesTab.vue'
import AccountCartTab from '~/components/account/AccountCartTab.vue'
import AccountLoginPrompt from '~/components/account/AccountLoginPrompt.vue'
import AccountPointsTab from '~/components/account/AccountPointsTab.vue'
import AccountWishlistTab from '~/components/account/AccountWishlistTab.vue'
import { useAuth } from '~/composables/useAuth'
import { useCart } from '~/composables/useCart'
import { useMembership } from '~/composables/useMembership'
import { useWishlist } from '~/composables/useWishlist'

type AccountTabId = 'points' | 'wishlist' | 'cart' | 'addresses'

const { t } = useI18n()
const auth = useAuth()
const { cartCount } = useCart()
const { items: wishlistItems, loadWishlist } = useWishlist()
const {
  levelName,
  points,
  tierInfo,
  levelDiscounts,
  userCoupons,
  userPointCards,
  tierConfigsLoading,
  assetsLoading,
  initMembership,
  refreshData,
  doLogout,
} = useMembership()

const sidePanel = inject<{ closeLeft?: () => void }>('sidePanel', {})
const showAuthModal = ref(false)
const authMode = ref<'login' | 'register'>('login')
const activeTab = ref<AccountTabId>('points')

const tabs: Array<{ id: AccountTabId; icon: string; labelKey: string; label: string }> = [
  { id: 'points', icon: 'lucide:gem', labelKey: 'accountSidebar.tabs.points', label: 'Points' },
  { id: 'wishlist', icon: 'lucide:heart', labelKey: 'accountSidebar.tabs.wishlist', label: 'Wishlist' },
  { id: 'cart', icon: 'lucide:shopping-cart', labelKey: 'accountSidebar.tabs.cart', label: 'Cart' },
  { id: 'addresses', icon: 'lucide:map-pin', labelKey: 'accountSidebar.tabs.addresses', label: 'Address' },
]

const isAuthenticated = computed(() => auth.isAuthenticated.value)
const pointsNumber = computed(() => Number(points.value || 0))
const wishlistCount = computed(() => wishlistItems.value.length)
const membershipLoading = computed(() => tierConfigsLoading.value || assetsLoading.value || auth.loading.value)

const displayName = computed(() => {
  const user = auth.user.value
  const profile = user?.profile
  return profile?.fullName || user?.display_name || user?.username || user?.email || t('accountSidebar.memberTitle', 'My account')
})

const closeSidebar = () => {
  sidePanel.closeLeft?.()
}

const openAuthForm = (mode: 'login' | 'register') => {
  authMode.value = mode
  showAuthModal.value = true
}

const handleAuthSuccess = async () => {
  showAuthModal.value = false
  await refreshData()
  await loadWishlist()
}

const logout = async () => {
  await doLogout()
  activeTab.value = 'points'
}

const initialiseAccountPanel = async () => {
  await initMembership()
  if (auth.isAuthenticated.value) {
    await loadWishlist()
  }
}

onMounted(() => {
  initialiseAccountPanel()
})

watch(
  () => auth.isAuthenticated.value,
  async (logged) => {
    if (logged) {
      await refreshData()
      await loadWishlist()
    }
  },
)
</script>

<style scoped>
.account-sidebar {
  display: flex;
  width: 100%;
  height: 100%;
  min-height: 0;
  flex-direction: column;
  gap: 0.9rem;
  overflow: hidden;
  padding: 0.35rem 0.15rem 0.75rem;
  color: #ffffff;
}

.account-sidebar__hero {
  display: flex;
  flex: 0 0 auto;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.85rem;
  border-radius: 1.55rem;
  background:
    radial-gradient(circle at top left, rgba(79, 70, 229, 0.2), transparent 42%),
    linear-gradient(135deg, rgba(17, 24, 39, 0.96), rgba(2, 6, 23, 0.92));
  box-shadow: 0 18px 44px -26px rgba(0, 0, 0, 1);
  padding: 1rem;
}

.account-sidebar__hero-main {
  min-width: 0;
}

.account-sidebar__kicker {
  margin: 0 0 0.35rem;
  color: #67e8f9;
  font-size: var(--tz-type-micro-label);
  font-weight: 900;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.account-sidebar__hero h2 {
  margin: 0;
  overflow: hidden;
  color: #ffffff;
  font-size: clamp(1.12rem, 3.5vw, 1.55rem);
  font-weight: 900;
  line-height: 1.08;
  text-overflow: ellipsis;
}

.account-sidebar__hero p:last-child {
  margin: 0.45rem 0 0;
  color: rgba(226, 232, 240, 0.82);
  font-size: 0.8rem;
  line-height: 1.45;
}

.account-sidebar__logout {
  display: inline-flex;
  min-height: 2rem;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.35rem;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  color: rgba(226, 232, 240, 0.88);
  padding: 0 0.7rem;
  font-size: var(--tz-type-micro-label);
  font-weight: 800;
}

.account-sidebar__logout :deep(svg) {
  width: 0.86rem;
  height: 0.86rem;
}

.account-sidebar__stats {
  display: grid;
  flex: 0 0 auto;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.55rem;
}

.account-sidebar__stats div {
  min-height: 4.1rem;
  border-radius: 1.05rem;
  background: rgba(255, 255, 255, 0.055);
  padding: 0.75rem;
}

.account-sidebar__stats span,
.account-sidebar__stats strong {
  display: block;
}

.account-sidebar__stats span {
  color: rgba(203, 213, 225, 0.8);
  font-size: var(--tz-type-micro-label);
  font-weight: 750;
}

.account-sidebar__stats strong {
  margin-top: 0.35rem;
  color: #ffffff;
  font-size: 1.15rem;
  font-weight: 900;
}

.account-sidebar__tabs {
  display: grid;
  flex: 0 0 auto;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.42rem;
  border-radius: 1.25rem;
  background: rgba(255, 255, 255, 0.045);
  padding: 0.35rem;
}

.account-sidebar__tab {
  display: inline-flex;
  min-width: 0;
  min-height: 2.35rem;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  border-radius: 0.95rem;
  color: rgba(226, 232, 240, 0.78);
  font-size: var(--tz-type-micro-label);
  font-weight: 850;
  transition: background 0.18s ease, color 0.18s ease, transform 0.18s ease;
}

.account-sidebar__tab :deep(svg) {
  width: 0.95rem;
  height: 0.95rem;
  flex: 0 0 auto;
}

.account-sidebar__tab span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-sidebar__tab--active {
  background: linear-gradient(135deg, #4efce7, #60a5fa);
  color: #020617;
  transform: translateY(-1px);
}

.account-sidebar__content {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  padding-right: 0.15rem;
}

.account-sidebar__content::-webkit-scrollbar {
  width: 0.32rem;
}

.account-sidebar__content::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgba(148, 163, 184, 0.5);
}

@media (min-width: 768px) {
  .account-sidebar {
    padding-inline: 0.25rem;
  }
}

@media (max-width: 520px) {
  .account-sidebar__hero {
    flex-direction: column;
  }

  .account-sidebar__logout {
    align-self: flex-start;
  }

  .account-sidebar__tabs {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>

