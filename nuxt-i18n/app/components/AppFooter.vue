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
            <img src="/icons/payment/paypal.svg?v=4" alt="PayPal" loading="lazy" />
            <img src="/icons/payment/visa.svg?v=4" alt="Visa" loading="lazy" />
            <img src="/icons/payment/mastercard.svg?v=4" alt="Mastercard" loading="lazy" />
            <img src="/icons/payment/amex.svg?v=4" alt="American Express" loading="lazy" />
            <img src="/icons/payment/discover.svg?v=4" alt="Discover" loading="lazy" />
            <img src="/icons/payment/jcb.svg?v=4" alt="JCB" loading="lazy" />
            <img src="/icons/payment/diners.svg?v=4" alt="Diners Club" loading="lazy" />
            <img src="/icons/payment/alipay.svg?v=4" alt="Alipay" loading="lazy" />
            <img src="/icons/payment/stripe.svg?v=4" alt="Stripe" loading="lazy" />
          </div>
        </div>

        <!-- Right: Menus -->
        <div class="footer-menus-wrapper">
          <FooterMenus />
        </div>

      </div>

      <div class="footer-widgets">
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
              to="/privacy"
              class="footer-info__link"
              target="_blank"
              rel="noopener noreferrer"
            >
              Privacy Policy
            </NuxtLink>
            <span class="footer-info__sep">|</span>
            <NuxtLink
              to="/terms"
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
</script>

<style scoped>
.app-footer {
  /* 增大底部 padding，预留空间给底部浮动 Dock (Desktop: 8rem, Mobile: 6rem) */
  padding: 1.5rem 1.5rem 8rem;
  background: rgba(0, 0, 0, 0.85);
  color: #f9fafb;
}

.footer-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1rem;
  /* Mobile: centered text */
  text-align: center;
}

.footer-main-row {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3rem;
  margin-bottom: 2rem;
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

.footer-subscription__payment img {
  height: 32px;
  width: auto;
  display: block;
}

.footer-bottom__info {
  font-size: 0.875rem;
  color: rgba(249, 250, 251, 0.7);
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
  color: rgba(249, 250, 251, 0.6);
  text-decoration: none;
  transition: color 0.2s;
}

.footer-info__link:hover {
  color: #ffffff;
  text-decoration: underline;
}

.footer-info__sep {
  color: rgba(255, 255, 255, 0.2);
}

.footer-site {
  margin: 0 0.25rem;
  font-weight: 600;
}

@media (min-width: 768px) {
  .footer-content {
    text-align: left;
    align-items: stretch;
  }

  .footer-main-row {
    display: grid;
    grid-template-columns: 360px 1fr;
    align-items: start;
    gap: 4rem;
    margin-bottom: 0; /* Removed bottom margin entirely */
  }

  .footer-subscription {
    max-width: 100%; /* Fill the grid column */
    padding: 0;
    text-align: left;
  }
  
  .footer-subscription__social {
    justify-content: flex-start; /* Desktop left align */
  }

  .footer-subscription__payment {
    justify-content: flex-start; /* Desktop left align */
  }
  
  /* Ensure label in subscription aligns left on desktop */
  .footer-subscription :deep(label) {
    text-align: left;
  }
  
  .footer-menus-wrapper {
    margin-top: 0.5rem; /* Visual alignment with subscription input */
  }

  .footer-bottom {
    /* Simplified bottom bar since social moved out */
    display: flex;
    margin-top: 0;
    padding-top: 1rem;
    border-top: 1px solid rgba(255, 255, 255, 0.05);
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
    padding: 1.5rem 1.25rem 6rem;
  }
}
</style>
