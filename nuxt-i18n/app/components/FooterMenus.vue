<template>
  <nav v-if="sections.length" class="footer-menus" aria-label="Footer navigation">
    <div class="footer-menus__grid">
      <section
        v-for="section in sections"
        :key="section.id"
        class="footer-menus__column"
        :class="{ 'is-open': isOpen(section.id) }"
      >
        <h3 class="footer-menus__title" @click="toggleSection(section.id)">
          <span class="footer-menus__title-text">
            <span v-if="section.id === 'support'">
              {{ $t('footer.menus.support', 'Support') }}
            </span>
            <span v-else-if="section.id === 'policies'">
              {{ $t('footer.menus.policies', 'Policies') }}
            </span>
            <span v-else>
              {{ $t(section.titleKey) }}
            </span>
          </span>
          <!-- Mobile Toggle Icon -->
          <span class="footer-menus__toggle-icon">
            <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M2.5 4.5L6 8L9.5 4.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </span>
        </h3>

        <!-- Special brand/contact column for the 'resources' section -->
        <div v-if="section.id === 'resources'" class="footer-menus__brand-text mobile-accordion-content">
          <template v-if="canRenderSiteSettings">
            <p v-if="siteName" class="footer-menus__brand-name">
              {{ siteName }}
            </p>
            <p v-if="siteDescription" class="footer-menus__brand-paragraph">
              {{ siteDescription }}
            </p>
            <p v-if="contactEmail" class="footer-menus__brand-paragraph">
              <a
                :href="`mailto:${contactEmail}`"
                class="footer-menus__link"
              >
                {{ contactEmail }}
              </a>
            </p>
            <p v-if="contactPhone" class="footer-menus__brand-paragraph">
              <a
                :href="`tel:${contactPhone}`"
                class="footer-menus__link"
              >
                {{ contactPhone }}
              </a>
            </p>
          </template>
        </div>

        <!-- Support section link list -->
        <ul v-else-if="section.id === 'support'" class="footer-menus__list mobile-accordion-content">
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

        <!-- Default link-list -->
        <ul v-else class="footer-menus__list mobile-accordion-content">
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
import { computed, onMounted, ref } from 'vue'
import { useLocalePath } from '#imports'
import type { FooterSection } from '~/utils/footerMenus'
import { footerMenus } from '~/utils/footerMenus'
import { useSiteSettings } from '~/composables/usePublicSettings'

const props = defineProps<{
  menus?: FooterSection[]
}>()

const localePath = useLocalePath()
const { siteSettings } = useSiteSettings()

const siteName = computed(() => siteSettings.value.siteTitle?.trim() || '')
const siteDescription = computed(() => siteSettings.value.siteDescription?.trim() || '')
const contactEmail = computed(() => siteSettings.value.contactEmail?.trim() || '')
const contactPhone = computed(() => siteSettings.value.contactPhone?.trim() || '')
const canRenderSiteSettings = ref(false)

onMounted(() => {
  canRenderSiteSettings.value = true
})

const sections = computed<FooterSection[]>(() => {
  if (props.menus && props.menus.length) {
    return props.menus
  }
  return footerMenus
})

const openSections = ref<Record<string, boolean>>({})

const toggleSection = (id: string) => {
  openSections.value[id] = !openSections.value[id]
}

const isOpen = (id: string) => {
  return !!openSections.value[id]
}
</script>

<style scoped>
.footer-menus {
  width: 100%;
}

.footer-menus__grid {
  display: grid;
  /* Keep policies aligned as a dedicated fifth column on larger screens. */
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 1.5rem;
}

.footer-menus__column {
  text-align: left;
}

.footer-menus__title {
  margin: 0 0 1.25rem;
  font-size: 0.85rem; 
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  /* Modern Gradient Text */
  background: linear-gradient(90deg, #e2e8f0 0%, #94a3b8 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  display: flex;
  align-items: center;
  justify-content: space-between;
  cursor: pointer; /* Pointer for interactive feel */
  user-select: none;
}

.footer-menus__title-text {
  /* Ensure gradient text logic works on the span if needed, 
     but parent has it. If parent is flex, gradient on flex item might break in some browsers if not careful.
     Putting gradient on text node only. */
}

.footer-menus__toggle-icon {
  display: none; /* Hidden on Desktop */
  color: var(--tz-text-muted);
  transition: transform 0.3s ease;
}

.footer-menus__column.is-open .footer-menus__toggle-icon {
  transform: rotate(180deg);
}

.footer-menus__brand-text {
  font-size: 0.85rem;
  line-height: 1.7;
  color: var(--tz-text-secondary);
}

.footer-menus__brand-name {
  margin: 0 0 0.85rem;
  font-size: 0.92rem;
  font-weight: 700;
  line-height: 1.45;
  color: #f8fafc;
}

.footer-menus__brand-paragraph {
  margin: 0 0 1rem;
}

.footer-menus__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

.footer-menus__link {
  font-size: 0.9rem;
  font-weight: 500;
  color: var(--tz-text-secondary);
  text-decoration: none;
  display: inline-block;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
}

.footer-menus__link:hover,
.footer-menus__link:focus-visible {
  color: #ffffff;
  transform: translateX(4px);
  text-shadow: 0 0 10px rgba(255, 255, 255, 0.5); 
}

/* Mobile Accordion Styles */
@media (max-width: 768px) {
  .footer-menus__grid {
    display: flex;
    flex-direction: column; /* Vertical stack */
    overflow-x: visible; /* No scroll */
    gap: 0; /* Gap handled by padding inside columns or items */
    margin: 0;
    padding: 0;
    scroll-snap-type: none;
    scrollbar-width: auto;
  }
  
  .footer-menus__column {
    min-width: auto;
    flex-shrink: 1;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05); /* Divider */
  }

  .footer-menus__column:last-child {
    border-bottom: none;
    padding-right: 0;
  }

  .footer-menus__title {
    margin: 0;
    padding: 1rem 0; /* Clickable area */
    font-size: 0.85rem;
    background: none; /* Remove gradient on mobile? Or keep? usually solid white is better for reading or keep theme. */
    -webkit-text-fill-color: unset; /* Reset text fill to allow color change */
    color: #f1f5f9;
    display: flex;
    justify-content: space-between;
  }

  .footer-menus__toggle-icon {
    display: block; /* Show Icon */
  }

  /* Hide content by default, show when open */
  .mobile-accordion-content {
    display: none;
    padding-bottom: 1rem;
    padding-left: 0.5rem; /* Indent slightly */
    animation: slideDown 0.3s ease-out;
  }

  .footer-menus__column.is-open .mobile-accordion-content {
    display: block;
  }

  .footer-menus__column.is-open .footer-menus__list.mobile-accordion-content {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .footer-menus__item {
    margin: 0;
  }

  .footer-menus__link {
    display: inline-flex;
    align-items: center;
    min-height: 2.25rem;
    padding-block: 0.2rem;
    line-height: 1.45;
  }

  .footer-menus__brand-paragraph .footer-menus__link {
    min-height: 2.25rem;
  }
  
  @keyframes slideDown {
    from { opacity: 0; transform: translateY(-10px); }
    to { opacity: 1; transform: translateY(0); }
  }
}
</style>
