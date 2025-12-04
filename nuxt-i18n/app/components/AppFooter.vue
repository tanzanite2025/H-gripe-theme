<template>
  <footer class="app-footer">
    <div class="footer-content">
      <div class="footer-subscription">
        <SubscriptionOptIn
          label="Subscribe for new products & blog updates"
        />
      </div>

      <div class="footer-widgets">
        <slot name="widgets" />
      </div>

      <div class="footer-menus-wrapper">
        <FooterMenus />
      </div>

      <div class="footer-bottom">
        <div class="footer-bottom__social">
          <SocialIcons :items="footerSocialItems" />
        </div>

        <div class="footer-bottom__info">
          <p class="footer-info__text">
            &copy;
            <span>{{ currentYear }}</span>
            <span class="footer-site">
              Top Sports Co., Limited. All Rights Reserved. Tanzanite® is a registered
              trademark.
            </span>
          </p>
        </div>

        <div class="footer-bottom__buttons">
          <div class="footer-info__buttons">
            <NuxtLink
              to="/privacy"
              class="footer-info__button"
              target="_blank"
              rel="noopener noreferrer"
            >
              Privacy Policy
            </NuxtLink>
            <NuxtLink
              to="/terms"
              class="footer-info__button"
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
  /* 增大底部 padding，让文字区域整体上移，
     预留空间给底部浮动 Dock 覆盖 */
  padding: 1.5rem 1.5rem 4.75rem;
  background: rgba(0, 0, 0, 0.85);
  color: #f9fafb;
}

.footer-content {
  max-width: 960px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  text-align: center;
}

.footer-subscription {
  width: 100%;
  max-width: 480px;
  margin-bottom: 2rem;
  padding: 1.25rem 1.5rem;
  border-radius: 0.75rem;
  background: radial-gradient(circle at top, #020617 0, #020617 40%, #020617 100%);
  box-shadow: 0 16px 40px rgba(15, 23, 42, 0.6);
}

.footer-widgets {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  justify-content: center;
}

.footer-menus-wrapper {
  width: 100%;
  margin-top: 0;
}

.footer-bottom {
  width: 100%;
  margin-top: 0.5rem;
  margin-bottom: 0.75rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
}

.footer-bottom__social {
  display: flex;
  justify-content: center;
}

.footer-bottom__info {
  font-size: 0.875rem;
  color: rgba(249, 250, 251, 0.7);
}

.footer-bottom__buttons {
  display: flex;
  justify-content: center;
  padding-bottom: 8px;
}

.footer-info__text {
  margin: 0;
}

.footer-info__buttons {
  margin-top: 0.5rem;
  display: inline-flex;
  align-items: center;
  gap: 0.75rem;
}

.footer-info__button {
  padding: 0.25rem 0.85rem;
  border-radius: 999px;
  border: 1px solid rgba(249, 250, 251, 0.7);
  background: transparent;
  color: rgba(249, 250, 251, 0.85);
  font-size: 0.75rem;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease, border-color 0.15s ease;
}

.footer-info__button:hover,
.footer-info__button:focus-visible {
  background-color: rgba(249, 250, 251, 0.12);
  border-color: rgba(249, 250, 251, 0.95);
  color: #ffffff;
}

.footer-site {
  margin: 0 0.5rem;
  font-weight: 600;
}

@media (min-width: 768px) {
  .footer-bottom {
    flex-direction: row;
    align-items: center;
    justify-content: space-between;
  }

  .footer-bottom__social,
  .footer-bottom__info,
  .footer-bottom__buttons {
    flex: 1;
  }

  /* Reorder for desktop: left = copyright, center = buttons, right = social */
  .footer-bottom__info {
    order: 1;
    text-align: left;
  }

  .footer-bottom__buttons {
    order: 2;
    justify-content: center;
    padding-bottom: 0;
  }

  .footer-bottom__social {
    order: 3;
    justify-content: flex-end;
  }
}

@media (max-width: 768px) {
  .app-footer {
    /* 移动端 Dock 通常更高，底部多留一些空间，但不需要额外一整行的高度 */
    padding: 1.5rem 1.25rem 4.75rem;
  }
}
</style>
