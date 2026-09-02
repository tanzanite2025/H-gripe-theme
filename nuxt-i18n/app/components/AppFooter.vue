<template>
  <footer class="app-footer app-footer-background-surface">
    <div class="footer-content">
      
      <div class="footer-feature-row">
        <FooterSiteOverview :links="siteOverviewLinks" />

        <div class="footer-subscription">
          <SubscriptionOptIn
            label="Subscribe for new products & blog updates"
          />
          <div class="footer-subscription__social">
            <SocialIcons />
          </div>
        </div>

      </div>

      <div class="footer-menus-wrapper">
        <FooterMenus :menus="routeMenuSections" />
      </div>

      <div v-if="$slots.widgets" class="footer-widgets">
        <slot name="widgets" />
      </div>

      <div class="footer-bottom">
        <div class="footer-bottom__info">
          <div class="footer-info__text-wrapper">
            <p class="footer-info__text">
              &copy; {{ currentYear }}. All pages use HTTPS with SSL encryption. Payments are securely processed; no card data stored.
            </p>
            <svg class="footer-secure-icon" width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" title="SSL Secure">
              <path d="M12 2L3 5V11C3 16.55 6.84 21.74 12 23C17.16 21.74 21 16.55 21 11V5L12 2Z" fill="#059669" stroke="#059669" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M12 2L3 5V11C3 16.55 6.84 21.74 12 23" fill="#059669" stroke="#059669" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <text x="50%" y="54%" dominant-baseline="middle" text-anchor="middle" fill="white" font-size="6" font-weight="bold" font-family="MapleUILatin">SSL</text>
              <rect x="10.5" y="14" width="3" height="2" rx="0.5" fill="#F59E0B" />
              <path d="M10.5 14V13C10.5 12.4477 10.9477 12 11.5 12C12.0523 12 12.5 12.4477 12.5 13V14" stroke="#F59E0B" stroke-width="0.5"/>
            </svg>
          </div>

          <div class="footer-bottom__utilities">
            <div class="footer-bottom__payment" aria-label="Accepted payment methods">
              <span v-for="icon in paymentIcons" :key="icon.src" class="payment-icon-tile">
                <img
                  :src="icon.src"
                  :alt="icon.alt"
                  :width="icon.width"
                  :height="icon.height"
                  :class="['payment-icon-tile__img', icon.className]"
                  loading="lazy"
                  decoding="async"
                />
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from '#imports'
import SocialIcons from '~/components/SocialIcons.vue'
import FooterSiteOverview from '~/components/FooterSiteOverview.vue'
import FooterMenus from '~/components/FooterMenus.vue'
import SubscriptionOptIn from '~/components/SubscriptionOptIn.vue'
import type { FooterSection } from '~/utils/footerMenus'
import { createFooterMenusFromRoutes } from '~/utils/footerMenus'

const currentYear = computed(() => new Date().getFullYear())
const router = useRouter()
const footerSections = computed<FooterSection[]>(() => (
  createFooterMenusFromRoutes(router.getRoutes())
))
const siteOverviewLinks = computed(() => (
  footerSections.value.find(section => section.id === 'siteOverview')?.links || []
))
const routeMenuSections = computed<FooterSection[]>(() => (
  footerSections.value.filter(section => section.id !== 'siteOverview')
))

interface PaymentIcon {
  src: string
  alt: string
  width: number
  height: number
  className?: string
}

const paymentIcons: PaymentIcon[] = [
  { src: '/icons/payment/paypal.svg', alt: 'PayPal', width: 200, height: 120 },
  { src: '/icons/payment/visa.svg', alt: 'Visa', width: 200, height: 120 },
  { src: '/icons/payment/mastercard.svg', alt: 'Mastercard', width: 200, height: 120 },
  { src: '/icons/payment/amex.svg', alt: 'American Express', width: 200, height: 120 },
  { src: '/icons/payment/discover.svg', alt: 'Discover', width: 200, height: 120 },
  { src: '/icons/payment/jcb.svg', alt: 'JCB', width: 200, height: 120 },
  { src: '/icons/payment/diners.svg', alt: 'Diners Club', width: 200, height: 120 },
  { src: '/icons/payment/alipay.svg?v=6', alt: 'Alipay', width: 200, height: 120, className: 'payment-icon--alipay' },
  { src: '/icons/payment/unionpay.svg', alt: 'UnionPay', width: 200, height: 120 },
  { src: '/icons/payment/wechatpay.svg', alt: 'WeChat Pay', width: 200, height: 120 },
  { src: '/icons/payment/applepay.svg?v=7', alt: 'Apple Pay', width: 200, height: 120 },
  { src: '/icons/payment/googlepay.svg', alt: 'Google Pay', width: 200, height: 120 },
  { src: '/icons/payment/default.svg', alt: 'Card Payment', width: 750, height: 471 },
]
</script>

<style scoped>
.app-footer {
  /* Keep the final footer row clear of the fixed bottom dock. */
  padding: 1rem 1.5rem var(--tz-bottom-dock-content-clearance, 6.5rem);
  color: var(--tz-text-primary);
  content-visibility: auto;
  contain-intrinsic-size: 32rem;
}

 .app-footer-background-surface {
  background: var(--tz-card-surface);
 }

.footer-content {
  width: 100%;
  max-width: none;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  /* Mobile: centered text */
  text-align: center;
}

.footer-feature-row {
  width: 100%;
  box-sizing: border-box;
  padding: 1.5rem clamp(0.5rem, 1.8vw, 2rem);
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
  align-items: start;
  gap: clamp(1.5rem, 4vw, 5rem);
  background: var(--tz-surface-subtle);
  border-block: 1px solid var(--tz-border-subtle);
}

.footer-subscription {
  width: 100%;
  max-width: 100%;
  padding: 0;
  background: transparent;
  border-radius: 0;
  box-shadow: none;
}

.footer-menus-wrapper {
  width: 100%;
}

.footer-subscription :deep(form) {
  width: 100%;
}

.footer-subscription :deep(.subscription-opt-in__control) {
  flex-direction: column !important;
  gap: 0.5rem;
}

.footer-subscription :deep(.subscription-opt-in__input) {
  width: 100%;
  flex: 0 0 auto;
  min-width: 0;
}

.footer-subscription :deep(.subscription-opt-in__button) {
  width: 100%;
  flex: 0 0 auto;
}

.footer-subscription__social {
  display: flex;
  justify-content: flex-start;
  margin-top: 0.85rem;
}

.payment-icon-tile {
  width: 44px;
  height: 28px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  /* Revert to simple style */
  background: transparent;
  border: none;
  border-radius: 4px; /* Optional, for image corners if needed */
  transition: transform 0.2s;
}

.payment-icon-tile:hover {
  transform: translateY(-2px);
  background: transparent;
}

.payment-icon-tile__img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: contain;
  padding: 0; /* Remove internal padding */
}

.payment-icon-tile__img.payment-icon--alipay {
  transform: scale(1.2);
  transform-origin: center;
}

.footer-bottom__info {
  font-size: 0.875rem;
  color: var(--tz-text-secondary);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  width: 100%;
}

.footer-info__text {
  margin: 0;
  line-height: 1.5;
}

.footer-info__text-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.footer-bottom__utilities {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: center;
  gap: 1rem;
  flex-wrap: wrap;
}

.footer-bottom__payment {
  display: flex;
  justify-content: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.footer-secure-icon {
  margin-top: -2px; /* Visual alignment */
}

.footer-site {
  margin: 0 0.25rem;
  font-weight: 600;
}

/* Tablet */
@media (min-width: 768px) and (max-width: 1023px) {
  .footer-content {
    text-align: left;
  }

  .footer-feature-row {
    gap: 2rem;
  }

  .footer-subscription {
    max-width: none;
    text-align: left;
  }
  
  .footer-bottom__info {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }
}

@media (min-width: 1024px) {
  .footer-content {
    text-align: left;
    align-items: stretch;
  }

  .footer-subscription {
    max-width: 100%;
    text-align: left;
  }

  .footer-subscription :deep(label) {
    text-align: left;
  }

  .footer-bottom {
    display: flex;
    margin-top: 0;
    padding-top: 0.5rem;
    border-top: 1px solid rgba(20, 32, 43, 0.1);
  }

  .footer-bottom__info {
    flex-direction: column;
    align-items: center;
    text-align: center;
    justify-content: center;
    gap: 0.75rem;
  }
}

@media (max-width: 768px) {
  .app-footer {
    padding: 1rem 1.25rem var(--tz-bottom-dock-content-clearance, 5rem);
  }

  .app-footer-background-surface {
    background: var(--tz-card-surface);
  }
  
  .footer-feature-row {
    display: flex;
    flex-direction: column;
    align-items: stretch; /* Ensure children take full width */
    gap: 2rem; /* Increase gap */
    padding-bottom: 2rem;
  }
  
  .footer-content {
    gap: 0.5rem;
  }
}
</style>
