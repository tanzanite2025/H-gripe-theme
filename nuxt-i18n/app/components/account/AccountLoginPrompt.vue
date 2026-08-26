<template>
  <div class="account-login-prompt">
    <div class="account-login-prompt__card">
      <div class="account-login-prompt__icon">
        <Icon name="lucide:user-round-check" />
      </div>
      <div class="account-login-prompt__copy">
        <p class="account-login-prompt__eyebrow">
          {{ t('accountSidebar.loginPrompt.eyebrow', 'Private account area') }}
        </p>
        <h3>{{ t('accountSidebar.loginPrompt.title', 'Sign in before managing personal data') }}</h3>
        <p>
          {{ t('accountSidebar.loginPrompt.description', 'Points, wishlist, cart checkout and shipping address tools live here after login.') }}
        </p>
      </div>
    </div>

    <div class="account-login-prompt__actions">
      <button type="button" class="account-login-prompt__primary" @click="$emit('login')">
        {{ t('authModal.actions.signIn', 'Sign in') }}
      </button>
      <button type="button" class="account-login-prompt__secondary" @click="$emit('register')">
        {{ t('authModal.actions.signUp', 'Create account') }}
      </button>
    </div>

    <div class="account-login-prompt__preview" aria-label="Account features">
      <div v-for="feature in features" :key="feature.id" class="account-login-prompt__feature">
        <Icon :name="feature.icon" />
        <div>
          <strong>{{ t(feature.titleKey, feature.title) }}</strong>
          <span>{{ t(feature.descKey, feature.desc) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'

defineEmits<{
  (event: 'login'): void
  (event: 'register'): void
}>()

const { t } = useI18n()

const features = [
  {
    id: 'points',
    icon: 'lucide:gem',
    titleKey: 'accountSidebar.features.points.title',
    title: 'Points',
    descKey: 'accountSidebar.features.points.desc',
    desc: 'Membership level and points balance',
  },
  {
    id: 'wishlist',
    icon: 'lucide:heart',
    titleKey: 'accountSidebar.features.wishlist.title',
    title: 'Wishlist',
    descKey: 'accountSidebar.features.wishlist.desc',
    desc: 'Saved products and quick add-to-cart',
  },
  {
    id: 'cart',
    icon: 'lucide:shopping-cart',
    titleKey: 'accountSidebar.features.cart.title',
    title: 'Cart',
    descKey: 'accountSidebar.features.cart.desc',
    desc: 'Review products and checkout faster',
  },
  {
    id: 'addresses',
    icon: 'lucide:map-pin',
    titleKey: 'accountSidebar.features.addresses.title',
    title: 'Addresses',
    descKey: 'accountSidebar.features.addresses.desc',
    desc: 'Prepare shipping details for checkout',
  },
]
</script>

<style scoped>
.account-login-prompt {
  display: flex;
  min-height: 0;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 1rem;
}

.account-login-prompt__card {
  display: flex;
  gap: 0.9rem;
  border-radius: 1.5rem;
  background:
    radial-gradient(circle at top left, rgba(5, 150, 105, 0.06), transparent 48%),
    var(--tz-card-surface);
  box-shadow: 0 18px 42px -24px rgba(15, 23, 42, 0.12);
  padding: 1rem;
}

.account-login-prompt__icon {
  display: inline-flex;
  width: 3rem;
  height: 3rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 1rem;
  border: 1px solid var(--tz-border-strong);
  background: #ffffff;
  color: var(--tz-text-primary);
}

.account-login-prompt__icon :deep(svg) {
  width: 1.45rem;
  height: 1.45rem;
}

.account-login-prompt__copy {
  min-width: 0;
}

.account-login-prompt__eyebrow {
  margin: 0 0 0.3rem;
  color: #059669;
  font-size: var(--tz-type-micro-label);
  font-weight: 800;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.account-login-prompt__copy h3 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: clamp(1rem, 3.5vw, 1.35rem);
  font-weight: 800;
  line-height: 1.18;
}

.account-login-prompt__copy p:last-child {
  margin: 0.45rem 0 0;
  color: rgba(232, 232, 232, 0.82);
  font-size: 0.84rem;
  line-height: 1.55;
}

.account-login-prompt__actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.7rem;
}

.account-login-prompt__primary,
.account-login-prompt__secondary {
  min-height: 2.75rem;
  border-radius: 999px;
  font-size: 0.86rem;
  font-weight: 800;
  transition: transform 0.18s ease, filter 0.18s ease, background 0.18s ease;
}

.account-login-prompt__primary {
  border: 1px solid var(--tz-border-strong);
  background: #ffffff;
  color: var(--tz-text-primary);
}

.account-login-prompt__secondary {
  border: 1px solid var(--tz-border-subtle);
  background: var(--tz-surface-subtle);
  color: var(--tz-text-primary);
}

.account-login-prompt__primary:hover,
.account-login-prompt__secondary:hover {
  transform: translateY(-1px);
  filter: brightness(1.05);
}

.account-login-prompt__preview {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0.7rem;
}

.account-login-prompt__feature {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 0.75rem;
  align-items: center;
  border-radius: 1.1rem;
  background: var(--tz-surface-subtle);
  padding: 0.8rem;
}

.account-login-prompt__feature :deep(svg) {
  width: 1.1rem;
  height: 1.1rem;
  color: #059669;
}

.account-login-prompt__feature strong,
.account-login-prompt__feature span {
  display: block;
}

.account-login-prompt__feature strong {
  color: var(--tz-text-primary);
  font-size: 0.86rem;
}

.account-login-prompt__feature span {
  margin-top: 0.1rem;
  color: var(--tz-text-secondary);
  font-size: 0.76rem;
  line-height: 1.4;
}
</style>

