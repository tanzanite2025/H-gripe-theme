<template>
  <nav v-if="sections.length" class="footer-menus" aria-label="Footer navigation">
    <div v-if="compactSections.length" class="footer-menus__compact-row">
      <button
        v-if="hasCompactMenuOverflow"
        type="button"
        class="footer-menus__scroll-button footer-menus__scroll-button--previous"
        :disabled="!canScrollCompactLeft"
        aria-label="Scroll footer navigation left"
        @click="scrollCompactMenus(-1)"
      >
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
          <path d="M10 3L5 8L10 13" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </button>

      <div
        ref="compactMenuViewport"
        class="footer-menus__compact-viewport"
        @scroll="updateCompactScrollState"
      >
        <div class="footer-menus__grid">
          <section
            v-for="section in compactSections"
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

            <!-- Route groups, nested pages, and runtime SHOP links. -->
            <ul class="footer-menus__list mobile-accordion-content">
              <FooterMenuNavigationItem
                v-for="(item, itemIndex) in section.links"
                :key="getFooterMenuNavigationItemKey(item, itemIndex)"
                :item="item"
              />
            </ul>
          </section>
        </div>
      </div>

      <button
        v-if="hasCompactMenuOverflow"
        type="button"
        class="footer-menus__scroll-button footer-menus__scroll-button--next"
        :disabled="!canScrollCompactRight"
        aria-label="Scroll footer navigation right"
        @click="scrollCompactMenus(1)"
      >
        <svg width="16" height="16" viewBox="0 0 16 16" fill="none" aria-hidden="true">
          <path d="M6 3L11 8L6 13" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
      </button>
    </div>

    <section
      v-if="guidesSection"
      class="footer-menus__guides-row"
      :class="{ 'is-open': isOpen(guidesSection.id) }"
    >
      <h3 class="footer-menus__title" @click="toggleSection(guidesSection.id)">
        <span class="footer-menus__title-text">
          {{ $t(guidesSection.titleKey, guidesSection.fallback || guidesSection.id) }}
        </span>
        <!-- Mobile Toggle Icon -->
        <span class="footer-menus__toggle-icon">
          <svg width="12" height="12" viewBox="0 0 12 12" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M2.5 4.5L6 8L9.5 4.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </span>
      </h3>

      <!-- Keep the complete documentation hierarchy together in its own row. -->
      <ul class="footer-menus__list footer-menus__list--guides mobile-accordion-content">
        <FooterMenuNavigationItem
          v-for="(item, itemIndex) in guidesSection.links"
          :key="getFooterMenuNavigationItemKey(item, itemIndex)"
          :item="item"
        />
      </ul>
    </section>
  </nav>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from '#imports'
import FooterMenuNavigationItem from '~/components/FooterMenuNavigationItem.vue'
import type { FooterNavigationItem, FooterSection } from '~/utils/footerMenus'
import { createFooterMenusFromRoutes } from '~/utils/footerMenus'

const props = defineProps<{
  menus?: FooterSection[]
}>()

const router = useRouter()

const sections = computed<FooterSection[]>(() => {
  if (props.menus) {
    return props.menus
  }
  return createFooterMenusFromRoutes(router.getRoutes())
})

const compactSections = computed<FooterSection[]>(() => (
  sections.value.filter(section => section.id !== 'guides')
))

const guidesSection = computed<FooterSection | undefined>(() => (
  sections.value.find(section => section.id === 'guides')
))

const openSections = ref<Record<string, boolean>>({})

const compactMenuViewport = ref<HTMLElement | null>(null)
const hasCompactMenuOverflow = ref(false)
const canScrollCompactLeft = ref(false)
const canScrollCompactRight = ref(false)

const updateCompactScrollState = () => {
  const viewport = compactMenuViewport.value
  if (!viewport) {
    hasCompactMenuOverflow.value = false
    canScrollCompactLeft.value = false
    canScrollCompactRight.value = false
    return
  }

  const maximumScrollLeft = Math.max(viewport.scrollWidth - viewport.clientWidth, 0)
  hasCompactMenuOverflow.value = maximumScrollLeft > 1
  canScrollCompactLeft.value = viewport.scrollLeft > 1
  canScrollCompactRight.value = viewport.scrollLeft < maximumScrollLeft - 1
}

const scrollCompactMenus = (direction: -1 | 1) => {
  const viewport = compactMenuViewport.value
  if (!viewport) return

  viewport.scrollBy({
    left: direction * Math.max(viewport.clientWidth * 0.8, 240),
    behavior: 'smooth',
  })
}

let compactMenuResizeObserver: ResizeObserver | undefined

onMounted(async () => {
  await nextTick()

  const viewport = compactMenuViewport.value
  if (viewport) {
    if (typeof ResizeObserver !== 'undefined') {
      compactMenuResizeObserver = new ResizeObserver(updateCompactScrollState)
      compactMenuResizeObserver.observe(viewport)
    }
  }

  window.addEventListener('resize', updateCompactScrollState)
  updateCompactScrollState()
})

watch(sections, async () => {
  await nextTick()
  updateCompactScrollState()
}, { deep: true, flush: 'post' })

onBeforeUnmount(() => {
  compactMenuResizeObserver?.disconnect()
  window.removeEventListener('resize', updateCompactScrollState)
})

const toggleSection = (id: string) => {
  openSections.value[id] = !openSections.value[id]
}

const isOpen = (id: string) => {
  return !!openSections.value[id]
}

const getFooterMenuNavigationItemKey = (
  item: FooterNavigationItem,
  index: number,
) => `${item.to || item.labelKey || item.fallback || 'group'}-${index}`
</script>

<style scoped>
.footer-menus {
  width: 100%;
}

.footer-menus__compact-row {
  display: flex;
  align-items: stretch;
  min-width: 0;
  gap: 0.5rem;
}

.footer-menus__compact-viewport {
  min-width: 0;
  flex: 1 1 auto;
  overflow-x: auto;
  scrollbar-width: none;
  scroll-behavior: smooth;
}

.footer-menus__compact-viewport::-webkit-scrollbar {
  display: none;
}

.footer-menus__grid {
  display: flex;
  flex-wrap: nowrap;
  gap: 1.5rem;
  width: max-content;
  min-width: 100%;
}

.footer-menus__column {
  min-width: 0;
  flex: 1 0 clamp(11rem, 16vw, 17rem);
  text-align: left;
}

.footer-menus__guides-row {
  width: 100%;
  margin-top: 1.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--tz-border-subtle);
  text-align: left;
}

.footer-menus__scroll-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  align-self: center;
  flex: 0 0 2.25rem;
  width: 2.25rem;
  height: 2.25rem;
  padding: 0;
  color: var(--tz-text-secondary);
  background: var(--tz-surface-subtle);
  border: 1px solid var(--tz-border-subtle);
  border-radius: 999px;
  cursor: pointer;
  transition: color 0.2s ease, background-color 0.2s ease, opacity 0.2s ease;
}

.footer-menus__scroll-button:hover:not(:disabled),
.footer-menus__scroll-button:focus-visible:not(:disabled) {
  color: var(--tz-text-primary);
  background: var(--tz-card-surface);
}

.footer-menus__scroll-button:disabled {
  cursor: default;
  opacity: 0.35;
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

.footer-menus__column.is-open .footer-menus__toggle-icon,
.footer-menus__guides-row.is-open .footer-menus__toggle-icon {
  transform: rotate(180deg);
}

.footer-menus__list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

/* Each top-level guide group gets one column; its child links stay vertical. */
.footer-menus__list--guides {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  align-items: start;
  column-gap: 2rem;
  row-gap: 1.5rem;
}

/* Mobile Accordion Styles */
@media (max-width: 768px) {
  .footer-menus__compact-row {
    display: block;
  }

  .footer-menus__compact-viewport {
    overflow-x: visible;
  }

  .footer-menus__grid {
    display: flex;
    flex-direction: column; /* Vertical stack */
    width: 100%;
    min-width: 0;
    overflow-x: visible; /* No scroll */
    gap: 0; /* Gap handled by padding inside columns or items */
    margin: 0;
    padding: 0;
    scroll-snap-type: none;
    scrollbar-width: auto;
  }
  
  .footer-menus__column {
    min-width: auto;
    flex: 0 0 auto;
    flex-shrink: 1;
    border-bottom: 1px solid rgba(20, 32, 43, 0.12); /* Divider */
  }

  .footer-menus__column:last-child {
    border-bottom: none;
    padding-right: 0;
  }

  .footer-menus__guides-row {
    margin-top: 0;
    padding-top: 0;
    border-top: none;
    border-bottom: 1px solid rgba(20, 32, 43, 0.12);
  }

  .footer-menus__scroll-button {
    display: none;
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

  .footer-menus__column.is-open .mobile-accordion-content,
  .footer-menus__guides-row.is-open .mobile-accordion-content {
    display: block;
  }

  .footer-menus__column.is-open .footer-menus__list.mobile-accordion-content,
  .footer-menus__guides-row.is-open .footer-menus__list.mobile-accordion-content {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  @keyframes slideDown {
    from { opacity: 0; transform: translateY(-10px); }
    to { opacity: 1; transform: translateY(0); }
  }
}

@media (min-width: 769px) and (max-width: 1023px) {
  .footer-menus__grid {
    gap: 1.25rem;
  }

  .footer-menus__list--guides {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (min-width: 1024px) and (max-width: 1279px) {
  .footer-menus__list--guides {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }
}
</style>
