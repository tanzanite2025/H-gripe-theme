<template>
  <transition name="header-mega">
    <div
      v-if="section"
      :id="panelId"
      class="header-mega"
      role="dialog"
      :aria-label="`${sectionLabel} menu`"
      @click.stop
    >
      <div class="header-mega__shell">
        <div class="header-mega__content">
          <div
            class="header-mega__grid"
            :class="`header-mega__grid--${section.id}`"
          >
            <article
              v-for="{ card, children, displaySize } in cardsWithChildren"
              :key="card.id"
              class="header-mega-card"
              :class="[
                `header-mega-card--${displaySize}`,
                `header-mega-card--${card.accent}`,
                `header-mega-card--card-${card.id}`,
                { 'header-mega-card--has-children': children.length > 0 },
              ]"
            >
              <span class="header-mega-card__glow" aria-hidden="true"></span>

              <NuxtLink
                class="header-mega-card__main"
                :to="localizedTo(card.to)"
                @click="scheduleNavigateClose"
              >
                <span class="header-mega-card__body">
                  <span v-if="shouldShowCardLabel(card)" class="header-mega-card__label">
                    {{ cardLabel(card) }}
                  </span>
                  <span class="header-mega-card__title">{{ cardTitle(card) }}</span>
                  <span
                    v-if="!shouldShowProductCategoryNavigationCardsInsideCard(card)"
                    class="header-mega-card__description"
                  >
                    {{ card.description }}
                  </span>
                </span>

                <span
                  v-if="!children.length && !shouldShowProductCategoryNavigationCardsInsideCard(card)"
                  class="header-mega-card__arrow"
                  aria-hidden="true"
                >
                  <Icon name="lucide:arrow-up-right" />
                </span>
              </NuxtLink>

              <div
                v-if="children.length"
                class="header-mega-card__children"
                :aria-label="`${cardTitle(card)} sections`"
              >
                <NuxtLink
                  v-for="child in children"
                  :key="child.id"
                  class="header-mega-card__child"
                  :to="localizedTo(child.to)"
                  :aria-label="childAccessibleLabel(child)"
                  @click.stop="scheduleNavigateClose"
                >
                  <span class="header-mega-card__child-label">{{ childLabel(child) }}</span>
                  <span class="header-mega-card__child-arrow" aria-hidden="true">
                    <Icon name="lucide:arrow-up-right" />
                  </span>
                </NuxtLink>
              </div>

              <ProductCategoryNavigationCards
                v-if="shouldShowProductCategoryNavigationCardsInsideCard(card)"
                class="header-mega__product-category-navigation"
                density="compact"
                :product-category-display-limit="4"
                @navigate="scheduleNavigateClose"
              />
            </article>
          </div>
        </div>

        <button
          type="button"
          class="header-mega__collapse"
          aria-label="Close menu"
          @click="closePanel"
        >
          <Icon name="lucide:chevron-up" />
        </button>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { computed, unref } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import ProductCategoryNavigationCards from '~/components/shop/ProductCategoryNavigationCards.vue'
import type { PrimaryMegaNavCard, PrimaryMegaNavSection } from '~/utils/primaryMegaNav'
import {
  getPrimaryMegaNavCardChildren,
  type PageSubNavigationChild,
} from '~/utils/pageSubNavigation'

const props = defineProps<{
  section: PrimaryMegaNavSection | null
  panelId: string
}>()

const emit = defineEmits<{
  navigate: []
}>()

const { t, locales } = useI18n() as any
const localePath = useLocalePath()

const localeCodes = computed(() => {
  return (unref(locales) || [])
    .map((item: any) => (typeof item === 'string' ? item : item?.code))
    .filter(Boolean)
})

const sectionLabel = computed(() => {
  const section = props.section
  if (!section) return ''
  return t(section.labelKey, section.labelFallback) as string
})

const localizedTo = (to: string) => {
  if (/^https?:\/\//i.test(to)) return to

  const hashIndex = to.indexOf('#')
  const withoutHash = hashIndex >= 0 ? to.slice(0, hashIndex) : to
  const hash = hashIndex >= 0 ? to.slice(hashIndex) : ''

  const queryIndex = withoutHash.indexOf('?')
  const path = queryIndex >= 0 ? withoutHash.slice(0, queryIndex) : withoutHash
  const query = queryIndex >= 0 ? withoutHash.slice(queryIndex) : ''

  return `${localePath(path || '/')}${query}${hash}`
}

const scheduleNavigateClose = () => {
  if (typeof window === 'undefined') return
  window.setTimeout(() => {
    emit('navigate')
  }, 0)
}

const closePanel = () => {
  emit('navigate')
}

const displaySizeForCard = (card: PrimaryMegaNavCard, children: PageSubNavigationChild[]) => {
  if (children.length > 0 && (card.size === 'compact' || card.size === 'standard')) {
    return 'wide'
  }

  return card.size
}

const cardsWithChildren = computed(() => {
  const section = props.section
  if (!section) return []

  return section.cards.map((card) => {
    const children = getPrimaryMegaNavCardChildren(section, card, localeCodes.value)

    return {
      card,
      children,
      displaySize: displaySizeForCard(card, children),
    }
  })
})

const shouldShowProductCategoryNavigationCardsInsideCard = (card: PrimaryMegaNavCard) => {
  return props.section?.id === 'products' && card.id === 'shop'
}

const cardLabel = (card: PrimaryMegaNavCard) => {
  return t(card.labelKey, card.labelFallback) as string
}

const cardTitle = (card: PrimaryMegaNavCard) => {
  return card.title || cardLabel(card)
}

const normalizeLabel = (value: string) => {
  return value.trim().replace(/\s+/g, ' ').toLowerCase()
}

const shouldShowCardLabel = (card: PrimaryMegaNavCard) => {
  return normalizeLabel(cardLabel(card)) !== normalizeLabel(cardTitle(card))
}

const childLabel = (child: PageSubNavigationChild) => {
  if (child.labelKey) return t(child.labelKey, child.fallback || child.label || child.id) as string
  return child.label || child.fallback || child.id
}

const childDescription = (child: PageSubNavigationChild) => {
  if (child.descriptionKey) return t(child.descriptionKey, child.description || '') as string
  return child.description || ''
}

const childAccessibleLabel = (child: PageSubNavigationChild) => {
  const label = childLabel(child)
  const description = childDescription(child)
  return description ? `${label}: ${description}` : label
}
</script>

<style scoped>
.header-mega {
  position: absolute;
  left: 50%;
  top: calc(100% - 1px);
  width: 100vw;
  max-width: none;
  transform: translateX(-50%);
  transform-origin: top center;
  z-index: 116;
  pointer-events: auto;
  will-change: opacity, transform, clip-path;
  clip-path: inset(0 0 0 0);
}

.header-mega__shell {
  position: relative;
  height: min(560px, calc(var(--tz-mobile-safe-viewport-height, 100vh) - var(--site-header-offset, 92px) - 18px));
  overflow: hidden;
  border-radius: 0;
  border: 1px solid rgba(255, 255, 255, 0.26);
  border-right: 0;
  border-left: 0;
  border-top-color: rgba(255, 255, 255, 0.38);
  border-bottom-color: rgba(255, 255, 255, 0.34);
  background: #000000;
  box-shadow:
    0 30px 80px -28px rgba(0, 0, 0, 1),
    inset 0 1px 0 rgba(255, 255, 255, 0.12),
    inset 0 -1px 0 rgba(255, 255, 255, 0.12);
}

.header-mega__shell::before {
  display: none;
}

.header-mega__shell::after {
  display: none;
}

.header-mega__content {
  position: relative;
  z-index: 1;
  box-sizing: border-box;
  height: 100%;
  max-height: none;
  overflow-x: hidden;
  overflow-y: auto;
  padding: 18px clamp(18px, 4vw, 56px) 54px;
  scrollbar-width: thin;
  scrollbar-color: rgba(100, 116, 139, 0.7) transparent;
}

.header-mega__collapse {
  position: absolute;
  right: 50%;
  bottom: 10px;
  z-index: 4;
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  transform: translateX(50%);
  border: 1px solid rgba(255, 255, 255, 0.32);
  border-radius: 999px;
  background: #ffffff;
  color: #050505;
  box-shadow: 0 14px 34px -24px rgba(0, 0, 0, 1);
  transition:
    border-color 0.18s ease,
    transform 0.18s ease;
}

.header-mega__collapse:hover,
.header-mega__collapse:focus-visible {
  border-color: rgba(255, 255, 255, 0.72);
  transform: translateX(50%) translateY(-1px);
}

.header-mega__collapse:focus-visible {
  outline: 2px solid rgba(255, 255, 255, 0.42);
  outline-offset: 2px;
}

.header-mega__collapse :deep(svg) {
  width: 1rem;
  height: 1rem;
  stroke-width: 2.4;
}

.header-mega__content::-webkit-scrollbar {
  width: 8px;
}

.header-mega__content::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgba(100, 116, 139, 0.68);
}

.header-mega__grid {
  column-count: 3;
  column-gap: 16px;
}

.header-mega__grid--products {
  display: grid;
  height: 100%;
  min-height: 0;
  grid-template-columns: minmax(0, 2fr) minmax(0, 1fr) minmax(0, 1fr);
  grid-template-rows: repeat(2, minmax(0, 1fr));
  align-items: stretch;
  column-count: initial;
  column-gap: normal;
  gap: 14px;
}

.header-mega__grid--products .header-mega-card {
  height: 100%;
  min-height: 0;
  margin: 0;
  break-inside: auto;
  overflow: hidden;
}

.header-mega__grid--products .header-mega-card--card-shop {
  grid-column: 1;
  grid-row: 1 / -1;
}

.header-mega__grid--products .header-mega-card--card-membership-and-points {
  grid-column: 2;
  grid-row: 1;
}

.header-mega__grid--products .header-mega-card--card-picture-warehouse {
  grid-column: 2;
  grid-row: 2;
}

.header-mega__grid--products .header-mega-card--card-spoke-calculator {
  grid-column: 3;
  grid-row: 1 / -1;
}

.header-mega__product-category-navigation {
  width: 100%;
  min-width: 0;
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
  padding: 0 var(--mega-card-padding) var(--mega-card-padding);
}

.header-mega__product-category-navigation :deep(.product-category-navigation-cards__all-link) {
  border-color: rgba(181, 255, 109, 0.62);
  background: rgba(181, 255, 109, 0.14);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.header-mega__product-category-navigation :deep(.product-category-navigation-cards__item) {
  border-color: rgba(181, 255, 109, 0);
  background: rgba(255, 255, 255, 0.055);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.06),
    0 12px 28px -24px rgba(0, 0, 0, 1);
}

.header-mega__product-category-navigation :deep(.product-category-navigation-cards__media) {
  border-bottom-color: rgba(181, 255, 109, 0.08);
  background: rgba(255, 255, 255, 0.04);
}

.header-mega__product-category-navigation :deep(.product-category-navigation-cards__footer) {
  background: rgba(11, 13, 17, 0.94);
}

.header-mega-card--card-shop .header-mega-card__main {
  flex: 0 0 auto;
  min-height: 0;
  padding-bottom: 10px;
}

.header-mega-card {
  --mega-accent: #B5FF6D;
  --mega-accent-soft: rgba(181, 255, 109, 0.14);
  --mega-accent-shadow: rgba(181, 255, 109, 0.35);
  --mega-card-padding: 16px;
  --mega-card-child-offset: 74px;

  position: relative;
  display: flex;
  flex-direction: column;
  align-self: start;
  break-inside: avoid;
  min-width: 0;
  width: 100%;
  min-height: 132px;
  margin: 0 0 14px;
  overflow: hidden;
  vertical-align: top;
  border: 1px solid rgba(181, 255, 109, 0);
  border-radius: 16px;
  background: linear-gradient(180deg, rgba(18, 21, 25, 0.98), rgba(10, 12, 16, 0.98));
  color: inherit;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.06),
    0 18px 44px -34px rgba(0, 0, 0, 1);
  transition:
    border-color 0.18s ease,
    background 0.18s ease,
    box-shadow 0.18s ease,
    color 0.22s ease;
}

.header-mega-card::before {
  display: none;
}

.header-mega-card:hover {
  border-color: rgba(181, 255, 109, 0.2);
  background: linear-gradient(180deg, rgba(22, 25, 30, 0.98), rgba(12, 14, 18, 0.98));
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.08),
    0 20px 48px -32px rgba(0, 0, 0, 1);
}

.header-mega-card__main {
  position: relative;
  z-index: 1;
  display: flex;
  flex: 1 1 auto;
  min-width: 0;
  min-height: inherit;
  gap: 12px;
  padding: var(--mega-card-padding);
  color: inherit;
  text-decoration: none;
}

.header-mega-card--has-children .header-mega-card__main {
  flex: 0 0 auto;
  min-height: 0;
  padding-bottom: 10px;
}

.header-mega-card--has-children .header-mega-card__description {
  -webkit-line-clamp: 2;
}

.header-mega-card__glow {
  display: none;
}

.header-mega-card__body {
  position: relative;
  display: flex;
  min-width: 0;
  flex: 1 1 auto;
  flex-direction: column;
}

.header-mega-card__label {
  color: var(--mega-accent);
  font-size: var(--tz-type-micro-label);
  font-weight: 850;
  letter-spacing: 0.14em;
  line-height: 1.2;
  text-transform: uppercase;
}

.header-mega-card__title {
  margin-top: 6px;
  color: #f8fafc;
  font-size: 20px;
  font-weight: 800;
  line-height: 1.15;
  letter-spacing: 0;
}

.header-mega-card__body > .header-mega-card__title:first-child {
  margin-top: 0;
}

.header-mega-card__description {
  display: -webkit-box;
  margin-top: 8px;
  overflow: hidden;
  color: rgba(203, 213, 225, 0.72);
  font-size: 17px;
  line-height: 1.48;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
}

.header-mega-card__arrow {
  position: relative;
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  margin-left: auto;
  border-radius: 999px;
  color: rgba(226, 232, 240, 0.68);
  transition: all 0.22s ease;
}

.header-mega-card__arrow :deep(svg) {
  width: 16px;
  height: 16px;
}

.header-mega-card:hover .header-mega-card__arrow {
  color: #ffffff;
  transform: translate(3px, -3px);
}

.header-mega-card__children {
  position: relative;
  z-index: 2;
  display: flex;
  flex-wrap: wrap;
  align-content: flex-start;
  gap: 8px;
  margin-top: 0;
  padding: 0 var(--mega-card-padding) var(--mega-card-padding);
}

.header-mega-card__child {
  display: inline-flex;
  max-width: 100%;
  min-height: 38px;
  align-items: center;
  justify-content: center;
  gap: 7px;
  border-radius: 999px;
  border: 1px solid rgba(181, 255, 109, 0.18);
  background: rgba(255, 255, 255, 0.055);
  padding: 0.56rem 0.92rem 0.56rem 1rem;
  color: rgba(241, 245, 249, 0.86);
  font-size: 14px;
  font-weight: 800;
  line-height: 1.1;
  text-decoration: none;
  transition:
    border-color 0.18s ease,
    background-color 0.18s ease,
    color 0.18s ease,
    transform 0.18s ease;
}

.header-mega-card__child:hover {
  border-color: var(--mega-accent);
  background: var(--mega-accent-soft);
  color: #ffffff;
  transform: translateY(-1px);
}

.header-mega-card__child:focus-visible {
  outline: 2px solid rgba(181, 255, 109, 0.54);
  outline-offset: 2px;
}

.header-mega-card__child-label {
  min-width: 0;
  max-width: 100%;
  color: rgba(248, 250, 252, 0.92);
  font-size: 14px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.header-mega-card__child-description {
  display: none;
}

.header-mega-card__child:hover .header-mega-card__child-description {
  color: rgba(248, 250, 252, 0.72);
}

.header-mega-card__child-arrow {
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  color: rgba(226, 232, 240, 0.58);
  transition:
    color 0.18s ease,
    transform 0.18s ease;
}

.header-mega-card__child-arrow :deep(svg) {
  width: 14px;
  height: 14px;
}

.header-mega-card__child:hover .header-mega-card__child-arrow {
  color: #ffffff;
  transform: translate(2px, -2px);
}

.header-mega-card--feature {
  --mega-card-padding: 20px;
  --mega-card-child-offset: 90px;
}

.header-mega-card--feature .header-mega-card__description {
  font-size: 17px;
  -webkit-line-clamp: 4;
}

.header-mega-card--compact {
  --mega-card-child-offset: 68px;
}

.header-mega-card--compact .header-mega-card__main {
  align-items: center;
}

.header-mega-card--compact .header-mega-card__description {
  -webkit-line-clamp: 2;
}

.header-mega-card--mint {
  --mega-accent: #B5FF6D;
  --mega-accent-soft: rgba(181, 255, 109, 0.13);
  --mega-accent-shadow: rgba(181, 255, 109, 0.36);
}

.header-mega-card--blue {
  --mega-accent: #B5FF6D;
  --mega-accent-soft: rgba(181, 255, 109, 0.13);
  --mega-accent-shadow: rgba(181, 255, 109, 0.34);
}

.header-mega-card--violet {
  --mega-accent: #B5FF6D;
  --mega-accent-soft: rgba(181, 255, 109, 0.13);
  --mega-accent-shadow: rgba(181, 255, 109, 0.34);
}

.header-mega-card--amber {
  --mega-accent: #B5FF6D;
  --mega-accent-soft: rgba(181, 255, 109, 0.13);
  --mega-accent-shadow: rgba(181, 255, 109, 0.34);
}

.header-mega-card--rose {
  --mega-accent: #B5FF6D;
  --mega-accent-soft: rgba(181, 255, 109, 0.13);
  --mega-accent-shadow: rgba(181, 255, 109, 0.34);
}

.header-mega-card--slate {
  --mega-accent: #B5FF6D;
  --mega-accent-soft: rgba(181, 255, 109, 0.13);
  --mega-accent-shadow: rgba(181, 255, 109, 0.34);
}

.header-mega-enter-active,
.header-mega-leave-active {
  transition:
    opacity 0.24s ease,
    transform 0.36s cubic-bezier(0.2, 0.8, 0.2, 1),
    clip-path 0.36s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.header-mega-enter-from,
.header-mega-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(-16px);
  clip-path: inset(0 0 100% 0);
}

.header-mega-enter-to,
.header-mega-leave-from {
  opacity: 1;
  transform: translateX(-50%) translateY(0);
  clip-path: inset(0 0 0 0);
}

@media (max-width: 1100px) {
  .header-mega__grid {
    column-count: 2;
  }

  .header-mega__grid--products {
    height: auto;
    min-height: 0;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    grid-template-rows: auto;
    align-items: start;
    gap: 14px;
  }

  .header-mega__grid--products .header-mega-card--card-shop {
    grid-column: 1 / -1;
    grid-row: auto;
  }

  .header-mega__grid--products .header-mega-card--card-membership-and-points,
  .header-mega__grid--products .header-mega-card--card-picture-warehouse,
  .header-mega__grid--products .header-mega-card--card-spoke-calculator {
    grid-column: auto;
    grid-row: auto;
  }

  .header-mega__grid--products .header-mega-card {
    height: auto;
    overflow: visible;
  }
}

@media (max-width: 767px) {
  .header-mega {
    position: fixed;
    inset: var(--header-mega-mobile-top, var(--site-header-offset, 7rem)) 0 0 0;
    width: 100vw;
    transform: none;
    z-index: 1300;
  }

  .header-mega__shell {
    height: 100%;
    border-radius: 0;
    box-sizing: border-box;
  }

  .header-mega__content {
    height: 100%;
    max-height: none;
    padding: 10px 10px calc(48px + env(safe-area-inset-bottom));
  }

  .header-mega__collapse {
    bottom: calc(10px + env(safe-area-inset-bottom));
  }

  .header-mega__grid {
    display: grid;
    grid-template-columns: 1fr;
    column-count: initial;
    column-gap: normal;
    gap: 10px;
  }

  .header-mega-card--feature,
  .header-mega-card--wide,
  .header-mega-card--standard,
  .header-mega-card--compact {
    grid-column: 1 / -1;
  }

  .header-mega__product-category-navigation {
    padding: 0 var(--mega-card-padding) var(--mega-card-padding);
    overflow: visible;
  }

  .header-mega-card {
    --mega-card-padding: 12px;
    --mega-card-child-offset: 12px;
    min-height: 0;
    width: auto;
    margin: 0;
    break-inside: auto;
    border-radius: 16px;
  }

  .header-mega-card__main {
    gap: 10px;
  }

  .header-mega-card__title,
  .header-mega-card--feature .header-mega-card__title {
    font-size: 17px;
  }

  /* Mobile keeps the category title and actions; descriptions are desktop-only. */
  .header-mega-card__description {
    display: none;
  }

  .header-mega-card__children {
    margin-top: auto;
    padding-top: 0;
    gap: 7px;
  }

  .header-mega-card__child {
    min-height: 36px;
    justify-content: center;
    padding: 0.55rem 0.82rem;
    border-color: rgba(181, 255, 109, 0.16);
    background: rgba(255, 255, 255, 0.055);
    font-size: 13px;
    line-height: 1.1;
  }

  .header-mega-card__child-label {
    font-size: 13px;
  }

  .header-mega-card__child-arrow {
    display: none;
  }
}
</style>
