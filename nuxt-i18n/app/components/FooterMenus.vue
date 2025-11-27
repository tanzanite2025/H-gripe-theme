<template>
  <nav v-if="sections.length" class="footer-menus" aria-label="Footer navigation">
    <div class="footer-menus__grid">
      <section
        v-for="section in sections"
        :key="section.id"
        class="footer-menus__column"
      >
        <h3
          v-if="section.id !== 'resources'"
          class="footer-menus__title"
        >
          <span v-if="section.id === 'support'">
            {{ $t('footer.menus.support', 'Support') }}
          </span>
          <span v-else>
            {{ $t(section.titleKey) }}
          </span>
        </h3>

        <!-- Special brand/contact column for the 'resources' section -->
        <div v-if="section.id === 'resources'" class="footer-menus__brand-text">
          <p class="footer-menus__brand-paragraph">
            {{
              $t(
                'footer.brand.line1',
                'Tanzanite is a brand of Top Sports Co., Limited.'
              )
            }}
          </p>
          <p class="footer-menus__brand-paragraph">
            {{
              $t(
                'footer.brand.line2',
                'Flat 1602, 16/F, Lucky Centre,'
              )
            }}
            <br />
            {{
              $t(
                'footer.brand.line3',
                'No.165-171 Wan Chai Road, Wan Chai'
              )
            }}
            <br />
            {{
              $t(
                'footer.brand.line4',
                'Hong Kong.'
              )
            }}
          </p>
          <p class="footer-menus__brand-paragraph">
            {{ $t('footer.brand.hkLabel', 'Hongkong:') }}
            <br />
            {{ $t('footer.brand.xmLabel', 'Xiamen:') }}
            <br />
            <a
              href="mailto:support@tanzanite.com"
              class="footer-menus__link"
            >
              support@tanzanite.com
            </a>
          </p>
        </div>

        <!-- Support section: link each item to the corresponding /support/... page -->
        <ul v-else-if="section.id === 'support'" class="footer-menus__list">
          <li class="footer-menus__item">
            <NuxtLink
              :to="localePath('/support/faqs')"
              class="footer-menus__link"
            >
              {{ $t('footer.support.allFaqs', "All FAQ'S") }}
            </NuxtLink>
          </li>
          <li class="footer-menus__item">
            <NuxtLink
              :to="localePath('/support/payment')"
              class="footer-menus__link"
            >
              {{ $t('footer.support.payment', 'Payment') }}
            </NuxtLink>
          </li>
          <li class="footer-menus__item">
            <NuxtLink
              :to="localePath('/support/after-sales')"
              class="footer-menus__link"
            >
              {{ $t('footer.support.afterSales', 'After sales') }}
            </NuxtLink>
          </li>
          <li class="footer-menus__item">
            <NuxtLink
              :to="localePath('/support/warranty')"
              class="footer-menus__link"
            >
              {{ $t('footer.support.warranty', 'Warranty') }}
            </NuxtLink>
          </li>
          <li class="footer-menus__item">
            <NuxtLink
              :to="localePath('/support/user-manuals')"
              class="footer-menus__link"
            >
              {{ $t('footer.support.userManuals', 'User Manuals') }}
            </NuxtLink>
          </li>
          <li class="footer-menus__item">
            <NuxtLink
              :to="localePath('/support/product-feedback')"
              class="footer-menus__link"
            >
              {{ $t('footer.support.productFeedback', 'Product Feedback') }}
            </NuxtLink>
          </li>
          <li class="footer-menus__item">
            <NuxtLink
              :to="localePath('/support/spoke-calculator')"
              class="footer-menus__link"
            >
              {{ $t('footer.support.spokecalculator', 'Spokecalculator') }}
            </NuxtLink>
          </li>
        </ul>

        <!-- Default link-list rendering for other columns -->
        <ul v-else class="footer-menus__list">
          <li
            v-for="link in section.links"
            :key="link.labelKey + '::' + link.to"
            class="footer-menus__item"
          >
            <NuxtLink
              v-if="!link.external"
              class="footer-menus__link"
              :to="localePath(link.to)"
            >
              {{ $t(link.labelKey) }}
            </NuxtLink>
            <a
              v-else
              class="footer-menus__link"
              :href="link.to"
              target="_blank"
              rel="noopener noreferrer"
            >
              {{ $t(link.labelKey) }}
            </a>
          </li>
        </ul>
      </section>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useLocalePath } from '#imports'
import type { FooterSection } from '~/utils/footerMenus'
import { footerMenus } from '~/utils/footerMenus'

const props = defineProps<{
  /**
   * Optional override menu structure.
   * If not provided, the default pure-i18n footerMenus config is used.
   */
  menus?: FooterSection[]
}>()

const localePath = useLocalePath()

const sections = computed<FooterSection[]>(() => {
  if (props.menus && props.menus.length) {
    return props.menus
  }
  return footerMenus
})
</script>

<style scoped>
.footer-menus {
  width: 100%;
}

.footer-menus__grid {
  display: grid;
  /* 4 equal-width columns on larger screens */
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 1.5rem;
}

.footer-menus__column {
  text-align: left;
}

.footer-menus__title {
  margin: 0 0 0.75rem;
  font-size: 0.875rem;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: rgba(249, 250, 251, 0.8);
}

.footer-menus__brand-text {
  font-size: 0.875rem;
  color: rgba(249, 250, 251, 0.75);
}

.footer-menus__brand-paragraph {
  margin: 0 0 0.5rem;
}

.footer-menus__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.footer-menus__link {
  font-size: 0.875rem;
  color: rgba(249, 250, 251, 0.75);
  text-decoration: none;
  transition: color 0.15s ease, transform 0.15s ease;
}

.footer-menus__link:hover,
.footer-menus__link:focus-visible {
  color: #e5f2ff;
  transform: translateY(-1px);
}

@media (max-width: 768px) {
  .footer-menus__grid {
    /* On small screens, fall back to 2 columns for readability */
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
