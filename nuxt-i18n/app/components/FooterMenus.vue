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
            {{ $t(section.titleKey, section.fallback || section.id) }}
          </span>
          <!-- Mobile Toggle Icon -->
          <span class="footer-menus__toggle-icon">
            <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M2.5 4.5L6 8L9.5 4.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </span>
        </h3>

        <!-- Link-list derived from the router's direct child pages -->
        <ul class="footer-menus__list mobile-accordion-content">
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
              {{ link.labelKey ? $t(link.labelKey, link.fallback || link.labelKey) : link.fallback }}
              </NuxtLink>
            <a
              v-else
              class="footer-menus__link"
              :href="link.to"
              target="_blank"
              rel="noopener noreferrer"
            >
              {{ link.labelKey ? $t(link.labelKey, link.fallback || link.labelKey) : link.fallback }}
            </a>
          </li>
        </ul>
      </section>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useLocalePath, useRouter } from '#imports'
import type { FooterSection } from '~/utils/footerMenus'
import { createFooterMenusFromRoutes } from '~/utils/footerMenus'

const props = defineProps<{
  menus?: FooterSection[]
}>()

const localePath = useLocalePath()
const router = useRouter()

const sections = computed<FooterSection[]>(() => {
  if (props.menus) {
    return props.menus
  }
  return createFooterMenusFromRoutes(router.getRoutes())
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
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1.5rem;
}

.footer-menus__column {
  min-width: 0;
  text-align: left;
}

.footer-menus__title {
  margin: 0 0 1.25rem;
  font-size: 0.85rem; 
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--tz-text-primary);
  background: none;
  -webkit-text-fill-color: unset;
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
  color: var(--tz-text-primary);
  transform: translateX(4px);
  text-shadow: none;
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
    border-bottom: 1px solid rgba(20, 32, 43, 0.12); /* Divider */
  }

  .footer-menus__column:last-child {
    border-bottom: none;
    padding-right: 0;
  }

  .footer-menus__title {
    margin: 0;
    padding: 1rem 0; /* Clickable area */
    font-size: 0.85rem;
    background: none;
    -webkit-text-fill-color: unset; /* Reset text fill to allow color change */
    color: var(--tz-text-primary);
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

  @keyframes slideDown {
    from { opacity: 0; transform: translateY(-10px); }
    to { opacity: 1; transform: translateY(0); }
  }
}

@media (min-width: 769px) and (max-width: 1023px) {
  .footer-menus__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 1.25rem;
  }
}

@media (min-width: 1024px) and (max-width: 1279px) {
  .footer-menus__grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 1.25rem;
  }
}

@media (min-width: 1280px) {
  .footer-menus__grid {
    grid-template-columns: repeat(5, minmax(0, 1fr));
  }
}
</style>
