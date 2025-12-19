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
              href="mailto:support@tanzanite.site"
              class="footer-menus__link"
            >
              support@tanzanite.site
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
              :to="localePath('/support/product-feedback')"
              class="footer-menus__link"
            >
              {{ $t('footer.support.productFeedback', 'Product Feedback') }}
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
  margin: 0 0 1rem;
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: 0.1em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.4);
}

.footer-menus__brand-text {
  font-size: 0.8rem;
  line-height: 1.6;
  color: rgba(255, 255, 255, 0.5);
}

.footer-menus__brand-paragraph {
  margin: 0 0 0.75rem;
}

.footer-menus__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.footer-menus__link {
  font-size: 0.875rem;
  color: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  display: inline-block;
  transition: color 0.2s ease, transform 0.2s ease;
}

.footer-menus__link:hover,
.footer-menus__link:focus-visible {
  color: #ffffff;
  transform: translateX(4px);
}

@media (max-width: 768px) {
  .footer-menus__grid {
    display: flex;
    overflow-x: auto;
    gap: 2rem;
    padding-bottom: 1rem;
    
    /* Scroll snap for better UX */
    scroll-snap-type: x mandatory;
    
    /* Extend scroll area to screen edges (counteracting parent padding) */
    margin-left: -1.25rem;
    margin-right: -1.25rem;
    padding-left: 1.25rem;
    padding-right: 1.25rem;
    
    /* Hide scrollbar for cleaner look */
    scrollbar-width: none; /* Firefox */
    -ms-overflow-style: none; /* IE/Edge */
    -webkit-overflow-scrolling: touch;
  }
  
  .footer-menus__grid::-webkit-scrollbar {
    display: none; /* Chrome/Safari */
  }

  .footer-menus__column {
    /* Ensure minimum width to prevent squishing on small screens */
    min-width: 160px;
    flex-shrink: 0;
    scroll-snap-align: start;
  }
  
  /* Add some spacer at the end so the last item isn't flush with edge */
  .footer-menus__column:last-child {
    padding-right: 1.25rem;
    box-sizing: content-box;
  }
}
</style>
