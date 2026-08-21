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
            :class="[
              `header-mega__grid--${section.id}`,
              { 'header-mega__grid--separated-columns': shouldUseSeparatedColumns },
            ]"
          >
            <div
              v-for="column in menuColumns"
              :key="column.id"
              class="header-mega__column"
              :class="`header-mega__column--${column.kind}`"
            >
              <article
                v-for="{ card, children, displaySize } in column.cards"
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
                  <span class="header-mega-card__icon" aria-hidden="true">
                    <Icon :name="card.icon" />
                  </span>

                  <span class="header-mega-card__body">
                    <span v-if="shouldShowCardLabel(card)" class="header-mega-card__label">
                      {{ cardLabel(card) }}
                    </span>
                    <span class="header-mega-card__title">{{ cardTitle(card) }}</span>
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
import type {
  PrimaryMegaNavCard,
  PrimaryMegaNavCardSize,
  PrimaryMegaNavSection,
} from '~/utils/primaryMegaNav'
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

type MegaMenuCardItem = {
  card: PrimaryMegaNavCard
  children: PageSubNavigationChild[]
  displaySize: PrimaryMegaNavCardSize
}

type MegaMenuColumn = {
  id: string
  kind: 'flat' | 'standalone' | 'nested'
  cards: MegaMenuCardItem[]
}

const separatedColumnSectionIds = new Set<PrimaryMegaNavSection['id']>([
  'support',
  'company',
  'guides',
])

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

const displaySizeForCard = (
  card: PrimaryMegaNavCard,
  children: PageSubNavigationChild[],
): PrimaryMegaNavCardSize => {
  if (children.length > 0 && (card.size === 'compact' || card.size === 'standard')) {
    return 'wide'
  }

  return card.size
}

const cardsWithChildren = computed<MegaMenuCardItem[]>(() => {
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

const shouldUseSeparatedColumns = computed(() => {
  const sectionId = props.section?.id
  return sectionId ? separatedColumnSectionIds.has(sectionId) : false
})

const menuColumns = computed<MegaMenuColumn[]>(() => {
  const cards = cardsWithChildren.value

  if (!shouldUseSeparatedColumns.value) {
    return [{ id: 'flat', kind: 'flat', cards }]
  }

  const standaloneCards = cards.filter(({ card, children }) => {
    return !children.length && !shouldShowProductCategoryNavigationCardsInsideCard(card)
  })
  const nestedColumns: [MegaMenuCardItem[], MegaMenuCardItem[]] = [[], []]

  cards
    .filter(({ card, children }) => {
      return children.length > 0 || shouldShowProductCategoryNavigationCardsInsideCard(card)
    })
    .forEach((item, index) => {
      if (index % 2 === 0) {
        nestedColumns[0].push(item)
        return
      }

      nestedColumns[1].push(item)
    })

  const columns: MegaMenuColumn[] = []

  if (standaloneCards.length > 0) {
    columns.push({ id: 'standalone', kind: 'standalone', cards: standaloneCards })
  }

  if (nestedColumns[0].length > 0) {
    columns.push({ id: 'nested-a', kind: 'nested', cards: nestedColumns[0] })
  }

  if (nestedColumns[1].length > 0) {
    columns.push({ id: 'nested-b', kind: 'nested', cards: nestedColumns[1] })
  }

  return columns
})

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
  --header-mega-max-height: min(560px, calc(var(--tz-mobile-safe-viewport-height, 100vh) - var(--site-header-overlay-offset, 92px) - 18px));

  height: auto;
  max-height: var(--header-mega-max-height);
  overflow: hidden;
  border-radius: 0;
  border: 1px solid rgba(255, 255, 255, 0.18);
  border-right: 0;
  border-left: 0;
  border-top-color: rgba(255, 255, 255, 0.26);
  border-bottom-color: rgba(255, 255, 255, 0.24);
  background: linear-gradient(180deg, #19191d 0%, #111116 100%);
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
  height: auto;
  max-height: var(--header-mega-max-height);
  overflow-x: hidden;
  overflow-y: auto;
  padding: 18px clamp(16px, 3.4vw, 48px) 50px;
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

.header-mega__column {
  min-width: 0;
}

.header-mega__column--flat {
  display: contents;
}

.header-mega__grid--products {
  display: grid;
  height: auto;
  min-height: 0;
  grid-template-columns: repeat(3, minmax(0, 1fr)) minmax(0, 0.92fr);
  grid-template-rows: repeat(4, minmax(0, auto));
  align-items: start;
  align-content: start;
  column-count: initial;
  column-gap: normal;
  gap: 14px;
}

.header-mega__grid--products .header-mega-card {
  height: auto;
  min-height: 0;
  margin: 0;
  break-inside: avoid;
  overflow: visible;
}

.header-mega__grid--products .header-mega-card--card-shop {
  grid-column: 1 / 4;
  grid-row: 1 / -1;
  height: auto;
  overflow: hidden;
}

.header-mega__grid--products .header-mega-card--card-membership-and-points,
.header-mega__grid--products .header-mega-card--card-picture-warehouse,
.header-mega__grid--products .header-mega-card--card-spoke-calculator,
.header-mega__grid--products .header-mega-card--card-blog {
  grid-column: 4;
}

.header-mega__grid--support,
.header-mega__grid--company,
.header-mega__grid--guides {
  display: grid;
  height: auto;
  min-height: 0;
  align-items: start;
  align-content: start;
  column-count: initial;
  column-gap: normal;
  gap: 14px;
}

.header-mega__grid--separated-columns {
  grid-template-columns: minmax(220px, 0.64fr) repeat(2, minmax(0, 1fr));
  grid-template-rows: none;
  gap: 12px clamp(20px, 3vw, 40px);
}

.header-mega__grid--separated-columns .header-mega__column {
  display: flex;
  min-width: 0;
  flex-direction: column;
  align-items: stretch;
}

.header-mega__grid--separated-columns .header-mega__column--standalone {
  gap: 22px;
}

.header-mega__grid--separated-columns .header-mega__column--nested {
  gap: 28px;
}

.header-mega__grid--separated-columns .header-mega-card {
  height: auto;
  min-height: 0;
  margin: 0;
  break-inside: avoid;
  overflow: visible;
}

.header-mega__grid--separated-columns .header-mega-card--has-children .header-mega-card__children {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, var(--mega-child-width)));
  justify-content: start;
  align-items: start;
  column-gap: 12px;
  row-gap: 8px;
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
  background: var(--tz-card-surface, #111116);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.06),
    0 12px 28px -24px rgba(0, 0, 0, 1);
}

.header-mega__product-category-navigation :deep(.product-category-navigation-cards__media) {
  border-bottom-color: rgba(181, 255, 109, 0.08);
  background: rgba(255, 255, 255, 0.04);
}

.header-mega__product-category-navigation :deep(.product-category-navigation-cards__footer) {
  background: var(--tz-card-surface, #111116);
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
  --mega-card-title-size: 16px;
  --mega-main-link-width: 260px;
  --mega-child-width: 184px;

  position: relative;
  display: flex;
  flex-direction: column;
  align-self: start;
  break-inside: avoid;
  min-width: 0;
  width: 100%;
  min-height: 0;
  margin: 0 0 14px;
  overflow: visible;
  vertical-align: top;
  border: 0;
  border-radius: 0;
  background: transparent;
  color: inherit;
  box-shadow: none;
}

.header-mega-card::before {
  display: none;
}

.header-mega-card__main {
  position: relative;
  z-index: 1;
  display: flex;
  flex: 1 1 auto;
  width: 100%;
  min-width: 0;
  min-height: inherit;
  align-items: center;
  gap: 12px;
  border-radius: 10px;
  padding: var(--mega-card-padding);
  color: inherit;
  text-decoration: none;
  transition:
    background-color 0.18s ease,
    color 0.18s ease;
}

.header-mega-card__main:hover,
.header-mega-card__main:focus-visible {
  background: rgba(255, 255, 255, 0.03);
}

.header-mega-card__main:focus-visible {
  outline: 1px solid rgba(181, 255, 109, 0.58);
  outline-offset: 2px;
}

.header-mega-card--has-children .header-mega-card__main {
  flex: 0 0 auto;
  min-height: 0;
  padding-bottom: 10px;
}

.header-mega-card__glow {
  display: none;
}

.header-mega-card__icon {
  display: inline-flex;
  width: 26px;
  height: 26px;
  flex: 0 0 26px;
  align-items: center;
  justify-content: center;
  color: var(--mega-accent);
}

.header-mega-card__icon :deep(svg) {
  width: 20px;
  height: 20px;
  stroke-width: 2.2;
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
  font-size: var(--mega-card-title-size);
  font-weight: 800;
  line-height: 1.15;
  letter-spacing: 0;
}

.header-mega-card__body > .header-mega-card__title:first-child {
  margin-top: 0;
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

.header-mega-card__main:hover .header-mega-card__arrow,
.header-mega-card__main:focus-visible .header-mega-card__arrow {
  color: #ffffff;
  transform: translate(3px, -3px);
}

.header-mega-card__children {
  position: relative;
  z-index: 2;
  display: grid;
  width: fit-content;
  grid-template-columns: repeat(2, minmax(0, var(--mega-child-width)));
  justify-content: start;
  align-content: flex-start;
  gap: 8px 16px;
  margin-top: 0;
  padding: 0 var(--mega-card-padding) var(--mega-card-padding);
}

.header-mega-card__child {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  width: 100%;
  max-width: 100%;
  min-width: 0;
  min-height: 34px;
  align-items: center;
  justify-content: stretch;
  gap: 8px;
  border-radius: 8px;
  border: 0;
  background: transparent;
  padding: 0.45rem 0.52rem;
  color: rgba(241, 245, 249, 0.86);
  font-size: 14px;
  font-weight: 800;
  line-height: 1.2;
  text-decoration: none;
  transform-origin: left center;
  transition:
    background-color 0.18s ease,
    color 0.18s ease,
    transform 0.18s ease;
}

.header-mega-card__child:hover {
  background: rgba(255, 255, 255, 0.03);
  color: #ffffff;
  transform: translateX(2px);
}

.header-mega-card__child:focus-visible {
  outline: none;
}

.header-mega-card__child-label {
  position: relative;
  display: block;
  min-width: 0;
  max-width: 100%;
  color: rgba(248, 250, 252, 0.92);
  font-size: 14px;
  overflow: visible;
  overflow-wrap: anywhere;
  text-overflow: clip;
  text-align: left;
  white-space: normal;
  transition: color 0.18s ease;
}

.header-mega-card__child-label::after {
  position: absolute;
  bottom: -0.28rem;
  left: 0;
  width: 0;
  height: 2px;
  border-radius: 999px;
  background: var(--mega-accent);
  box-shadow: 0 0 8px rgba(181, 255, 109, 0.58);
  content: '';
  opacity: 0;
  transition:
    width 0.28s cubic-bezier(0.4, 0, 0.2, 1),
    opacity 0.18s ease;
}

.header-mega-card__child:hover .header-mega-card__child-label,
.header-mega-card__child:focus-visible .header-mega-card__child-label {
  color: #ffffff;
}

.header-mega-card__child:hover .header-mega-card__child-label::after,
.header-mega-card__child:focus-visible .header-mega-card__child-label::after {
  width: 100%;
  opacity: 1;
}

.header-mega-card__child-description {
  display: none;
}

.header-mega-card__child:hover .header-mega-card__child-description {
  color: rgba(248, 250, 252, 0.72);
}

.header-mega-card__child-arrow {
  display: inline-flex;
  flex: 0 0 14px;
  align-items: center;
  justify-content: center;
  justify-self: end;
  margin-left: 0;
  color: rgba(226, 232, 240, 0.58);
  transition:
    color 0.18s ease,
    transform 0.18s ease;
}

.header-mega-card__child-arrow :deep(svg) {
  width: 14px;
  height: 14px;
}

.header-mega-card__child:hover .header-mega-card__child-arrow,
.header-mega-card__child:focus-visible .header-mega-card__child-arrow {
  color: #ffffff;
}

.header-mega-card--feature {
  --mega-card-padding: 20px;
  --mega-card-child-offset: 90px;
}

.header-mega-card--compact {
  --mega-card-child-offset: 68px;
}

.header-mega-card--compact .header-mega-card__main {
  align-items: center;
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
    grid-template-columns: repeat(3, minmax(0, 1fr)) minmax(0, 0.92fr);
    grid-template-rows: repeat(4, minmax(0, auto));
    align-items: start;
    align-content: start;
    gap: 14px;
  }

  .header-mega__grid--products .header-mega-card--card-shop {
    grid-column: 1 / 4;
    grid-row: 1 / -1;
  }

  .header-mega__grid--products .header-mega-card--card-membership-and-points,
  .header-mega__grid--products .header-mega-card--card-picture-warehouse,
  .header-mega__grid--products .header-mega-card--card-spoke-calculator,
  .header-mega__grid--products .header-mega-card--card-blog {
    grid-column: 4;
  }

  .header-mega__grid--products .header-mega-card {
    height: auto;
    overflow: visible;
  }

  .header-mega__grid--separated-columns {
    grid-template-columns: minmax(200px, 0.68fr) repeat(2, minmax(0, 1fr));
    gap: 12px clamp(16px, 2.8vw, 36px);
  }

  .header-mega__grid--separated-columns .header-mega__column--standalone {
    gap: 20px;
  }

  .header-mega__grid--separated-columns .header-mega__column--nested {
    gap: 24px;
  }

  .header-mega__grid--separated-columns .header-mega-card--has-children .header-mega-card__children {
    grid-template-columns: repeat(2, minmax(0, var(--mega-child-width)));
    column-gap: 10px;
    row-gap: 6px;
  }

  .header-mega__grid--separated-columns .header-mega-card__child {
    width: 100%;
    justify-content: stretch;
  }

  .header-mega__grid--separated-columns .header-mega-card__child-label {
    text-align: left;
    white-space: normal;
    overflow-wrap: anywhere;
  }

  .header-mega__grid--support,
  .header-mega__grid--company,
  .header-mega__grid--guides {
    height: auto;
    min-height: 0;
    grid-template-columns: minmax(200px, 0.68fr) repeat(2, minmax(0, 1fr));
    grid-template-rows: none;
    align-items: start;
    align-content: start;
    gap: 12px clamp(16px, 2.8vw, 36px);
  }
}

@media (max-width: 767px) {
  .header-mega {
    position: fixed;
    inset: calc(var(--header-mega-mobile-top, var(--site-header-overlay-offset, 7rem)) + 2px) 2px 2px;
    width: auto;
    transform: none;
    z-index: 10010;
  }

  .header-mega__shell {
    height: auto;
    max-height: var(--header-mega-max-height);
    border: 1px solid rgba(172, 172, 172, 0.32);
    border-radius: 14px;
    box-sizing: border-box;
    box-shadow:
      0 24px 64px -28px #000000,
      inset 0 1px 0 rgba(255, 255, 255, 0.1);
  }

  .header-mega__content {
    height: auto;
    max-height: var(--header-mega-max-height);
    padding: 10px 10px calc(48px + env(safe-area-inset-bottom));
  }

  .header-mega__collapse {
    bottom: calc(10px + env(safe-area-inset-bottom));
  }

  .header-mega__grid {
    column-count: 1;
    column-gap: 0;
  }

  .header-mega__grid--products {
    display: block;
    column-count: 1;
    column-gap: 0;
  }

  .header-mega__grid--separated-columns,
  .header-mega__grid--support,
  .header-mega__grid--company,
  .header-mega__grid--guides {
    display: block;
    height: auto;
    column-count: 1;
    column-gap: 0;
  }

  .header-mega__grid--separated-columns .header-mega__column {
    display: block;
  }

  .header-mega__grid--separated-columns .header-mega__column--standalone,
  .header-mega__grid--separated-columns .header-mega__column--nested {
    gap: 0;
  }

  .header-mega__grid--products .header-mega-card--card-shop {
    column-span: all;
  }

  .header-mega__grid--products .header-mega-card,
  .header-mega__grid--support .header-mega-card,
  .header-mega__grid--company .header-mega-card,
  .header-mega__grid--guides .header-mega-card {
    margin: 0 0 5px;
  }

  .header-mega__product-category-navigation {
    padding: 0 var(--mega-card-padding) var(--mega-card-padding);
    overflow: visible;
  }

  .header-mega-card {
    --mega-card-padding: 8px;
    --mega-card-child-offset: 12px;
    --mega-card-title-size: 13px;
    min-height: 0;
    width: 100%;
    margin: 0 0 5px;
    break-inside: avoid;
    border-radius: 10px;
  }

  .header-mega-card__main {
    gap: 7px;
    align-items: center;
    padding: 8px;
  }

  .header-mega-card--has-children .header-mega-card__main,
  .header-mega-card--card-shop .header-mega-card__main {
    padding-bottom: 4px;
  }

  .header-mega-card__title,
  .header-mega-card--feature .header-mega-card__title {
    font-size: var(--mega-card-title-size);
  }

  .header-mega-card__children {
    width: 100%;
    margin-top: 0;
    padding-top: 0;
    padding-right: 8px;
    padding-bottom: 5px;
    padding-left: 8px;
    gap: 3px 6px;
  }

  .header-mega-card__child {
    min-height: 25px;
    justify-content: stretch;
    width: 100%;
    gap: 6px;
    padding: 0.2rem 0.3rem;
    background: transparent;
    font-size: 12px;
    line-height: 1.1;
  }

  .header-mega-card__child-label {
    min-width: 0;
    max-width: 100%;
    font-size: 12px;
    line-height: 1.1;
    white-space: normal;
    overflow-wrap: anywhere;
  }

  .header-mega-card__title {
    white-space: normal;
    overflow-wrap: anywhere;
    line-height: 1.2;
  }

  .header-mega-card__child-arrow {
    display: inline-flex;
    flex: 0 0 12px;
    justify-self: end;
    margin-left: 0;
    color: rgba(226, 232, 240, 0.72);
  }

  .header-mega-card__child-arrow :deep(svg) {
    width: 12px;
    height: 12px;
  }

  .header-mega__grid--separated-columns .header-mega-card--has-children .header-mega-card__children {
    display: grid;
    width: 100%;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    column-gap: 6px;
    row-gap: 2px;
  }
}
</style>
