<template>
  <footer class="app-footer">
    <div class="footer-content">
      
      <!-- Desktop: Side-by-Side Layout Wrapper -->
      <div class="footer-main-row">
        
        <!-- Left: Subscription -->
        <div class="footer-subscription">
          <SubscriptionOptIn
            label="Subscribe for new products & blog updates"
          />
          <!-- Social Icons moved here -->
          <div class="footer-subscription__social">
            <SocialIcons :items="footerSocialItems" />
          </div>
          
          <!-- Payment Icons -->
          <div class="footer-subscription__payment">
            <span v-for="icon in paymentIcons" :key="icon.src" class="payment-icon-tile">
              <img
                :src="icon.src"
                :alt="icon.alt"
                :class="['payment-icon-tile__img', icon.className]"
                loading="lazy"
                decoding="async"
              />
            </span>
          </div>
        </div>

        <!-- Right: Menus -->
        <div class="footer-menus-wrapper">
          <FooterMenus />
        </div>

      </div>

      <div v-if="$slots.widgets" class="footer-widgets">
        <slot name="widgets" />
      </div>

      <div class="footer-bottom">
        <div class="footer-bottom__info">
          <p class="footer-info__text">
            &copy;
            <span>{{ currentYear }}</span>
            <span class="footer-site">
              Top Sports Co., Limited. All Rights Reserved. Tanzanite® is a registered
              trademark.
            </span>
          </p>
          
          <!-- Moved buttons here for better layout grouping -->
          <div class="footer-info__buttons">
            <NuxtLink
              to="/policies/privacy"
              class="footer-info__link"
              target="_blank"
              rel="noopener noreferrer"
            >
              Privacy Policy
            </NuxtLink>
            <span class="footer-info__sep">|</span>
            <NuxtLink
              to="/policies/cookie"
              class="footer-info__link"
              target="_blank"
              rel="noopener noreferrer"
            >
              Cookie Policy
            </NuxtLink>
            <span class="footer-info__sep">|</span>
            <NuxtLink
              to="/policies/refund-return"
              class="footer-info__link"
              target="_blank"
              rel="noopener noreferrer"
            >
              Refund &amp; Return
            </NuxtLink>
            <span class="footer-info__sep">|</span>
            <NuxtLink
              to="/policies/terms"
              class="footer-info__link"
              target="_blank"
              rel="noopener noreferrer"
            >
              Terms of Service
            </NuxtLink>
          </div>
        </div>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRuntimeConfig } from '#imports'
import SocialIcons from '~/components/SocialIcons.vue'
import FooterMenus from '~/components/FooterMenus.vue'
import SubscriptionOptIn from '~/components/SubscriptionOptIn.vue'

const config = useRuntimeConfig()

const currentYear = computed(() => new Date().getFullYear())

const siteTitle = computed(() => {
  const title = (config.public as { siteTitle?: string }).siteTitle
  return title && title.trim().length ? title : 'Tanzanite'
})

interface FooterSocialItem {
  url: string
  label: string
  network?: string
  size?: number
}

const footerSocialItems: FooterSocialItem[] = [
  {
    network: 'twitter',
    url: 'https://twitter.com',
    label: 'Twitter',
    size: 24,
  },
  {
    network: 'instagram',
    url: 'https://instagram.com',
    label: 'Instagram',
    size: 24,
  },
  {
    network: 'github',
    url: 'https://github.com',
    label: 'GitHub',
    size: 24,
  },
]

interface PaymentIcon {
  src: string
  alt: string
  className?: string
}

const paymentIcons: PaymentIcon[] = [
  { src: '/icons/payment/paypal.svg', alt: 'PayPal' },
  { src: '/icons/payment/visa.svg', alt: 'Visa' },
  { src: '/icons/payment/mastercard.svg', alt: 'Mastercard' },
  { src: '/icons/payment/amex.svg', alt: 'American Express' },
  { src: '/icons/payment/discover.svg', alt: 'Discover' },
  { src: '/icons/payment/jcb.svg', alt: 'JCB' },
  { src: '/icons/payment/diners.svg', alt: 'Diners Club' },
  { src: '/icons/payment/alipay.svg?v=6', alt: 'Alipay', className: 'payment-icon--alipay' },
  { src: '/icons/payment/unionpay.svg', alt: 'UnionPay' },
  { src: '/icons/payment/wechatpay.svg', alt: 'WeChat Pay' },
  { src: '/icons/payment/applepay.svg?v=7', alt: 'Apple Pay' },
  { src: '/icons/payment/googlepay.svg', alt: 'Google Pay' },
  { src: '/icons/payment/stripe.svg', alt: 'Stripe' },
  { src: '/icons/payment/default.svg', alt: 'Card Payment' },
]
</script>

<style scoped>
.app-footer {
  /* 增大底部 padding，预留空间给底部浮动 Dock (Desktop: 8rem, Mobile: 6rem) */
  padding: 1rem 1.5rem 8rem;
  background: linear-gradient(140deg, #0c0f17 0%, #141925 45%, #1b2230 100%);
  color: #f5f6fa;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  /* Mobile: centered text */
  text-align: center;
}

.footer-main-row {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.5rem;
  margin-bottom: 0.5rem;
}

.footer-subscription {
  width: 100%;
  max-width: 480px;
  padding: 0 1rem;
}

.footer-menus-wrapper {
  width: 100%;
}

.footer-subscription__social {
  margin-top: 1.5rem;
  display: flex;
  justify-content: center; /* Mobile center */
}

.footer-subscription__payment {
  margin-top: 1rem;
  display: flex;
  justify-content: center; /* Mobile center */
  gap: 0.5rem;
  flex-wrap: wrap;
}

.payment-icon-tile {
  width: 44px;
  height: 26px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  overflow: hidden;
  background: transparent;
  border: 0;
}

.payment-icon-tile__img {
  width: 100%;
  height: 100%;
  display: block;
  object-fit: contain;
}

.payment-icon-tile__img.payment-icon--alipay {
  transform: scale(1.28);
  transform-origin: center;
}

.footer-bottom__info {
  font-size: 0.875rem;
  color: rgba(245, 246, 250, 0.78);
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

.footer-info__buttons {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.8rem;
}

.footer-info__link {
  color: rgba(245, 246, 250, 0.65);
  text-decoration: none;
  transition: color 0.2s;
}

.footer-info__link:hover {
  color: #ffffff;
  text-decoration: underline;
}

.footer-info__sep {
  color: rgba(245, 246, 250, 0.35);
}

.footer-site {
  margin: 0 0.25rem;
  font-weight: 600;
}

/* 平板断点 (768px - 1024px): 保持移动端布局 */
@media (min-width: 768px) and (max-width: 1023px) {
  .footer-content {
    text-align: center;
  }

  .footer-main-row {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2rem;
  }

  .footer-subscription {
    max-width: 480px;
    text-align: center;
  }
  
  .footer-subscription__social {
    justify-content: center;
  }

  .footer-subscription__payment {
    justify-content: center;
  }

  .footer-bottom__info {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }
}

/* 桌面端 (1024px+): 并排布局 */
@media (min-width: 1024px) {
  .footer-content {
    text-align: left;
    align-items: stretch;
  }

  .footer-main-row {
    display: grid;
    grid-template-columns: 360px 1fr;
    align-items: start;
    gap: 4rem;
    margin-bottom: 0;
  }

  .footer-subscription {
    max-width: 100%;
    padding: 0;
    text-align: left;
  }
  
  .footer-subscription__social {
    justify-content: flex-start;
  }

  .footer-subscription__payment {
    justify-content: flex-start;
  }
  
  .footer-subscription :deep(label) {
    text-align: left;
  }
  
  .footer-menus-wrapper {
    margin-top: 0.5rem;
  }

  .footer-bottom {
    display: flex;
    margin-top: 0;
    padding-top: 0.5rem;
    border-top: 1px solid rgba(255, 255, 255, 0.08);
  }

  .footer-bottom__info {
    flex-direction: row;
    flex-wrap: wrap;
    align-items: center;
    text-align: left;
    gap: 1rem;
  }
}

@media (max-width: 768px) {
  .app-footer {
    /* 移动端 Dock 通常更高，底部多留一些空间 */
    padding: 1rem 1.25rem 6rem;
  }
  
  .footer-main-row {
    gap: 1.5rem;
    margin-bottom: 0.5rem;
  }
  
  .footer-content {
    gap: 0.5rem;
  }
}
</style>
