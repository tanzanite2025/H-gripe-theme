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
          <div class="header-mega__grid">
            <article
              v-for="{ card, children } in cardsWithChildren"
              :key="card.id"
              class="header-mega-card"
              :class="[
                `header-mega-card--${card.size}`,
                `header-mega-card--${card.accent}`,
                { 'header-mega-card--has-children': children.length > 0 },
              ]"
            >
              <span class="header-mega-card__glow" aria-hidden="true"></span>

              <NuxtLink
                class="header-mega-card__main"
                :to="localizedTo(card.to)"
                @click="emit('navigate')"
              >
                <span class="header-mega-card__icon" aria-hidden="true">
                  <Icon :name="card.icon" />
                </span>

                <span class="header-mega-card__body">
                  <span v-if="shouldShowCardLabel(card)" class="header-mega-card__label">
                    {{ cardLabel(card) }}
                  </span>
                  <span class="header-mega-card__title">{{ cardTitle(card) }}</span>
                  <span class="header-mega-card__description">{{ card.description }}</span>
                </span>

                <span class="header-mega-card__arrow" aria-hidden="true">
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
                  @click.stop="emit('navigate')"
                >
                  {{ childLabel(child) }}
                </NuxtLink>
              </div>
            </article>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script setup lang="ts">
import { computed, unref } from 'vue'
import { useI18n, useLocalePath } from '#imports'
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

const cardsWithChildren = computed(() => {
  const section = props.section
  if (!section) return []

  return section.cards.map((card) => ({
    card,
    children: getPrimaryMegaNavCardChildren(section, card, localeCodes.value),
  }))
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
</script>

<style scoped>
.header-mega {
  position: absolute;
  left: 50%;
  top: calc(100% + 0.55rem);
  width: min(95vw, 1180px);
  transform: translateX(-50%);
  z-index: 116;
  pointer-events: auto;
}

.header-mega__shell {
  position: relative;
  overflow: hidden;
  border-radius: 30px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background:
    radial-gradient(circle at top left, rgba(64, 255, 170, 0.16), transparent 34%),
    radial-gradient(circle at 80% 20%, rgba(107, 115, 255, 0.18), transparent 30%),
    linear-gradient(135deg, rgba(15, 23, 42, 0.98), rgba(2, 6, 23, 0.98));
  box-shadow:
    0 30px 80px -28px rgba(0, 0, 0, 1),
    inset 0 1px 0 rgba(255, 255, 255, 0.08);
}

.header-mega__shell::before {
  content: '';
  position: absolute;
  inset: 0;
  background-image: radial-gradient(rgba(255, 255, 255, 0.1) 1px, transparent 1px);
  background-size: 18px 18px;
  mask-image: linear-gradient(135deg, rgba(0, 0, 0, 0.8), transparent 70%);
  pointer-events: none;
}

.header-mega__shell::after {
  content: '';
  position: absolute;
  inset: 1px;
  border-radius: 29px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  pointer-events: none;
}

.header-mega__content {
  position: relative;
  z-index: 1;
  max-height: min(690px, calc(100vh - var(--site-header-offset, 92px) - 18px));
  overflow-x: hidden;
  overflow-y: auto;
  padding: 18px;
  scrollbar-width: thin;
  scrollbar-color: rgba(100, 116, 139, 0.7) transparent;
}

.header-mega__content::-webkit-scrollbar {
  width: 8px;
}

.header-mega__content::-webkit-scrollbar-thumb {
  border-radius: 999px;
  background: rgba(100, 116, 139, 0.68);
}

.header-mega__grid {
  display: grid;
  grid-template-columns: repeat(12, minmax(0, 1fr));
  grid-auto-flow: dense;
  gap: 12px;
}

.header-mega-card {
  --mega-accent: #40ffaa;
  --mega-accent-soft: rgba(64, 255, 170, 0.14);
  --mega-accent-shadow: rgba(64, 255, 170, 0.35);
  --mega-card-padding: 16px;
  --mega-card-child-offset: 74px;

  position: relative;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 120px;
  overflow: hidden;
  border-radius: 22px;
  border: 1px solid rgba(255, 255, 255, 0.07);
  background:
    linear-gradient(135deg, rgba(30, 41, 59, 0.88), rgba(15, 23, 42, 0.88)),
    radial-gradient(circle at top left, var(--mega-accent-soft), transparent 62%);
  color: inherit;
  box-shadow: 0 18px 40px -24px rgba(0, 0, 0, 1);
  transition:
    transform 0.22s ease,
    border-color 0.22s ease,
    background 0.22s ease,
    box-shadow 0.22s ease;
}

.header-mega-card::before {
  content: '';
  position: absolute;
  inset: 0 auto 0 0;
  width: 4px;
  background: var(--mega-accent);
  opacity: 0.78;
  transition: width 0.22s ease, opacity 0.22s ease;
}

.header-mega-card:hover {
  transform: translateY(-3px);
  border-color: color-mix(in srgb, var(--mega-accent) 54%, rgba(255, 255, 255, 0.12));
  background:
    linear-gradient(135deg, rgba(36, 48, 70, 0.94), rgba(15, 23, 42, 0.96)),
    radial-gradient(circle at top left, var(--mega-accent-soft), transparent 58%);
  box-shadow:
    0 26px 56px -30px rgba(0, 0, 0, 1),
    0 12px 34px -30px var(--mega-accent-shadow);
}

.header-mega-card:hover::before {
  width: 7px;
  opacity: 1;
}

.header-mega-card__main {
  position: relative;
  z-index: 1;
  display: flex;
  flex: 1 1 auto;
  min-width: 0;
  min-height: inherit;
  gap: 14px;
  padding: var(--mega-card-padding);
  color: inherit;
  text-decoration: none;
}

.header-mega-card--has-children .header-mega-card__main {
  min-height: 0;
  padding-bottom: 10px;
}

.header-mega-card__glow {
  position: absolute;
  right: -40px;
  top: -44px;
  width: 130px;
  height: 130px;
  border-radius: 999px;
  background: var(--mega-accent-soft);
  filter: blur(6px);
  opacity: 0.7;
  pointer-events: none;
}

.header-mega-card__icon {
  position: relative;
  display: inline-flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: var(--mega-accent-soft);
  color: var(--mega-accent);
  transition: all 0.22s ease;
}

.header-mega-card__icon :deep(svg) {
  width: 22px;
  height: 22px;
}

.header-mega-card:hover .header-mega-card__icon {
  background: var(--mega-accent);
  color: #020617;
  transform: scale(1.05) rotate(-4deg);
  box-shadow: 0 16px 32px -18px var(--mega-accent-shadow);
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
  font-size: 10px;
  font-weight: 850;
  letter-spacing: 0.14em;
  line-height: 1.2;
  text-transform: uppercase;
}

.header-mega-card__title {
  margin-top: 6px;
  color: #f8fafc;
  font-size: 15px;
  font-weight: 800;
  line-height: 1.15;
  letter-spacing: -0.02em;
}

.header-mega-card__body > .header-mega-card__title:first-child {
  margin-top: 0;
}

.header-mega-card__description {
  display: -webkit-box;
  margin-top: 8px;
  overflow: hidden;
  color: rgba(203, 213, 225, 0.72);
  font-size: 12px;
  line-height: 1.45;
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
  gap: 6px;
  margin-top: auto;
  padding: 0 var(--mega-card-padding) var(--mega-card-padding) var(--mega-card-child-offset);
}

.header-mega-card__child {
  display: inline-flex;
  min-height: 28px;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, var(--mega-accent) 34%, rgba(255, 255, 255, 0.1));
  background: rgba(15, 23, 42, 0.58);
  padding: 0.38rem 0.62rem;
  color: rgba(241, 245, 249, 0.86);
  font-size: 11px;
  font-weight: 750;
  line-height: 1;
  text-decoration: none;
  transition:
    border-color 0.18s ease,
    background-color 0.18s ease,
    color 0.18s ease,
    transform 0.18s ease;
}

.header-mega-card__child:hover {
  border-color: var(--mega-accent);
  background: color-mix(in srgb, var(--mega-accent) 16%, rgba(15, 23, 42, 0.76));
  color: #ffffff;
  transform: translateY(-1px);
}

.header-mega-card--feature {
  --mega-card-padding: 20px;
  --mega-card-child-offset: 90px;

  grid-column: span 6;
  min-height: 174px;
}

.header-mega-card--feature .header-mega-card__icon {
  width: 56px;
  height: 56px;
  border-radius: 20px;
}

.header-mega-card--feature .header-mega-card__icon :deep(svg) {
  width: 27px;
  height: 27px;
}

.header-mega-card--feature .header-mega-card__title {
  font-size: 20px;
}

.header-mega-card--feature .header-mega-card__description {
  font-size: 13px;
  -webkit-line-clamp: 4;
}

.header-mega-card--wide {
  grid-column: span 6;
  min-height: 142px;
}

.header-mega-card--standard {
  grid-column: span 3;
  min-height: 132px;
}

.header-mega-card--compact {
  --mega-card-child-offset: 68px;

  grid-column: span 3;
  min-height: 104px;
}

.header-mega-card--compact .header-mega-card__main {
  align-items: center;
}

.header-mega-card--compact .header-mega-card__icon {
  width: 38px;
  height: 38px;
  border-radius: 14px;
}

.header-mega-card--compact .header-mega-card__description {
  -webkit-line-clamp: 2;
}

.header-mega-card--mint {
  --mega-accent: #40ffaa;
  --mega-accent-soft: rgba(64, 255, 170, 0.13);
  --mega-accent-shadow: rgba(64, 255, 170, 0.36);
}

.header-mega-card--blue {
  --mega-accent: #38bdf8;
  --mega-accent-soft: rgba(56, 189, 248, 0.14);
  --mega-accent-shadow: rgba(56, 189, 248, 0.34);
}

.header-mega-card--violet {
  --mega-accent: #8b5cf6;
  --mega-accent-soft: rgba(139, 92, 246, 0.16);
  --mega-accent-shadow: rgba(139, 92, 246, 0.34);
}

.header-mega-card--amber {
  --mega-accent: #f59e0b;
  --mega-accent-soft: rgba(245, 158, 11, 0.14);
  --mega-accent-shadow: rgba(245, 158, 11, 0.32);
}

.header-mega-card--rose {
  --mega-accent: #fb7185;
  --mega-accent-soft: rgba(251, 113, 133, 0.14);
  --mega-accent-shadow: rgba(251, 113, 133, 0.32);
}

.header-mega-card--slate {
  --mega-accent: #cbd5e1;
  --mega-accent-soft: rgba(203, 213, 225, 0.12);
  --mega-accent-shadow: rgba(203, 213, 225, 0.26);
}

.header-mega-enter-active,
.header-mega-leave-active {
  transition: opacity 0.18s ease, transform 0.18s ease;
}

.header-mega-enter-from,
.header-mega-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(-8px) scale(0.985);
}

@media (max-width: 1100px) {
  .header-mega__grid {
    grid-template-columns: repeat(8, minmax(0, 1fr));
  }

  .header-mega-card--feature {
    grid-column: span 4;
  }

  .header-mega-card--wide,
  .header-mega-card--standard,
  .header-mega-card--compact {
    grid-column: span 4;
  }
}

@media (max-width: 640px) {
  .header-mega {
    top: calc(100% + 0.35rem);
    width: calc(100vw - 1rem);
  }

  .header-mega__shell {
    border-radius: 22px;
  }

  .header-mega__content {
    max-height: min(72vh, calc(100vh - var(--site-header-offset, 150px) - 12px));
    padding: 10px;
  }

  .header-mega__grid {
    grid-template-columns: 1fr;
    gap: 10px;
  }

  .header-mega-card--feature,
  .header-mega-card--wide,
  .header-mega-card--standard,
  .header-mega-card--compact {
    grid-column: 1 / -1;
  }

  .header-mega-card {
    --mega-card-padding: 12px;
    --mega-card-child-offset: 12px;
    min-height: 0;
    border-radius: 18px;
  }

  .header-mega-card__main {
    gap: 10px;
  }

  .header-mega-card__icon {
    width: 38px;
    height: 38px;
    border-radius: 14px;
  }

  .header-mega-card__icon :deep(svg) {
    width: 20px;
    height: 20px;
  }

  .header-mega-card--feature .header-mega-card__icon {
    width: 42px;
    height: 42px;
    border-radius: 15px;
  }

  .header-mega-card--feature .header-mega-card__icon :deep(svg) {
    width: 22px;
    height: 22px;
  }

  .header-mega-card--feature .header-mega-card__title {
    font-size: 17px;
  }

  .header-mega-card__description {
    -webkit-line-clamp: 2;
  }

  .header-mega-card__children {
    padding-top: 0;
  }

  .header-mega-card__child {
    min-height: 30px;
    font-size: 11px;
  }
}
</style>
