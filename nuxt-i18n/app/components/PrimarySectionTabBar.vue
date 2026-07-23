<template>
  <div
    v-if="currentSection && currentSection.cards.length"
    class="primary-section-tabs-shell"
  >
    <nav
      class="primary-section-tabs"
      :class="{
        'primary-section-tabs--open': mobileOpen,
        'primary-section-tabs--scrolling': isScrolling,
      }"
      :aria-label="`${sectionLabel} navigation`"
    >
      <button
        type="button"
        class="primary-section-tabs__mobile-trigger"
        :aria-expanded="mobileOpen"
        :aria-label="mobileOpen ? 'Collapse section navigation' : 'Expand section navigation'"
        @click="mobileOpen = !mobileOpen"
      >
        <span class="primary-section-tabs__hamburger" aria-hidden="true">
          <span />
          <span />
          <span />
        </span>
        <span class="primary-section-tabs__mobile-text">
          {{ sectionLabel }} / {{ activeLabel }}
        </span>
      </button>

      <div class="primary-section-tabs__list" role="tablist">
        <NuxtLink
          v-for="card in currentSection.cards"
          :key="card.id"
          :to="localizedTo(card.to)"
          class="primary-section-tabs__item"
          :class="{ 'primary-section-tabs__item--active': isCardActive(card) }"
          role="tab"
          :aria-selected="isCardActive(card)"
          :aria-current="isCardActive(card) ? 'page' : undefined"
          @click="mobileOpen = false"
        >
          <Icon :name="card.icon" class="primary-section-tabs__icon" aria-hidden="true" />
          <span>{{ cardTitle(card) }}</span>
        </NuxtLink>
      </div>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, unref, watch } from 'vue'
import { useI18n, useLocalePath, useRoute } from '#imports'
import { useScrollIdleVisibility } from '~/composables/useScrollIdleVisibility'
import {
  findPrimaryMegaNavSectionByPath,
  primaryMegaNavPathMatches,
  primaryMegaNavSections,
  type PrimaryMegaNavCard,
} from '~/utils/primaryMegaNav'

const route = useRoute()
const localePath = useLocalePath()
const { locales, t } = useI18n()
const mobileOpen = ref(false)
const { isScrolling } = useScrollIdleVisibility()

const localeCodes = computed(() =>
  (unref(locales) || [])
    .map((item: any) => (typeof item === 'string' ? item : item?.code))
    .filter(Boolean)
)

const currentSection = computed(() =>
  findPrimaryMegaNavSectionByPath(route.path || '/', primaryMegaNavSections, localeCodes.value)
)

const splitTarget = (to: string) => {
  const hashIndex = to.indexOf('#')
  const beforeHash = hashIndex >= 0 ? to.slice(0, hashIndex) : to
  const hash = hashIndex >= 0 ? to.slice(hashIndex) : ''
  const queryIndex = beforeHash.indexOf('?')

  return {
    path: queryIndex >= 0 ? beforeHash.slice(0, queryIndex) || '/' : beforeHash || '/',
    query: queryIndex >= 0 ? beforeHash.slice(queryIndex) : '',
    hash,
  }
}

const localizedTo = (to: string) => {
  const target = splitTarget(to)
  return `${localePath(target.path)}${target.query}${target.hash}`
}

const cardTitle = (card: PrimaryMegaNavCard) => {
  return card.title || (t(card.labelKey, card.labelFallback) as string)
}

const hasExactHashActive = computed(() => {
  const section = currentSection.value
  if (!section) return false

  return section.cards.some((card) => {
    const target = splitTarget(card.to)
    return Boolean(
      target.hash &&
      primaryMegaNavPathMatches(route.path || '/', target.path, localeCodes.value) &&
      route.hash === target.hash
    )
  })
})

const isCardActive = (card: PrimaryMegaNavCard) => {
  const target = splitTarget(card.to)
  const pathActive = primaryMegaNavPathMatches(route.path || '/', target.path, localeCodes.value)
  if (!pathActive) return false
  if (target.hash) return route.hash === target.hash
  if (hasExactHashActive.value) return false
  return true
}

const activeLabel = computed(() => {
  const active = currentSection.value?.cards.find((card) => isCardActive(card))
  return active ? cardTitle(active) : sectionLabel.value
})

const sectionLabel = computed(() => {
  const section = currentSection.value
  return section ? (t(section.labelKey, section.labelFallback) as string) : ''
})

watch(
  () => route.fullPath,
  () => {
    mobileOpen.value = false
  }
)

watch(isScrolling, (scrolling) => {
  if (scrolling) mobileOpen.value = false
})
</script>

<style scoped>
.primary-section-tabs-shell {
  --primary-section-tab-bar-height: 58px;
  --primary-section-tab-bar-gap: clamp(0.8rem, 0.95vw, 1.15rem);
  height: calc(
    var(--primary-section-tab-bar-height) +
    var(--primary-section-tab-bar-gap)
  );
}

.primary-section-tabs {
  --primary-section-tab-bar-height: 58px;
  --primary-section-tab-bar-gap: clamp(0.8rem, 0.95vw, 1.15rem);
  box-sizing: border-box;
  position: fixed;
  inset-inline-start: 0;
  top: calc(var(--site-header-offset, 120px) + var(--primary-section-tab-bar-gap));
  z-index: 88;
  width: 100vw;
  width: 100dvw;
  max-width: 100vw;
  max-width: 100dvw;
  min-height: var(--primary-section-tab-bar-height);
  margin: 0;
  padding: 0.55rem clamp(1rem, 3vw, 3.5rem);
  border: 1px solid rgba(45, 212, 191, 0.14);
  border-right: 0;
  border-left: 0;
  border-radius: 0;
  background:
    linear-gradient(135deg, rgba(15, 23, 42, 0.93), rgba(2, 6, 23, 0.8)),
    radial-gradient(circle at 50% 0%, rgba(45, 212, 191, 0.13), transparent 46%);
  backdrop-filter: blur(18px);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.05),
    0 16px 34px rgba(0, 0, 0, 0.34);
  transition:
    transform 0.22s ease,
    opacity 0.18s ease,
    visibility 0s linear 0s;
}

.primary-section-tabs--scrolling {
  transform: translateY(calc(-100% - var(--primary-section-tab-bar-gap)));
  opacity: 0;
  visibility: hidden;
  pointer-events: none;
  transition:
    transform 0.22s ease,
    opacity 0.18s ease,
    visibility 0s linear 0.22s;
}

.primary-section-tabs__mobile-trigger {
  display: none;
}

.primary-section-tabs__list {
  display: flex;
  width: 100%;
  align-items: center;
  justify-content: center;
  gap: clamp(0.35rem, 0.55vw, 0.7rem);
}

.primary-section-tabs__item {
  min-width: 0;
  flex: 0 1 auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.45rem;
  white-space: nowrap;
  border: 0;
  border-radius: 9999px;
  padding: 0.7rem clamp(0.72rem, 1.02vw, 1.18rem);
  background: rgba(15, 23, 42, 0.82);
  color: rgba(226, 232, 240, 0.88);
  font-size: clamp(0.7rem, 0.72vw, 0.86rem);
  font-weight: 800;
  line-height: 1;
  text-decoration: none;
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.04),
    0 10px 22px rgba(0, 0, 0, 0.24);
  transition:
    background-color 0.18s ease,
    color 0.18s ease,
    transform 0.18s ease,
    box-shadow 0.18s ease;
}

.primary-section-tabs__item:hover {
  color: #ffffff;
  background: rgba(30, 41, 59, 0.94);
  transform: translateY(-1px);
}

.primary-section-tabs__item--active {
  color: #0f172a;
  background: #ffffff;
  box-shadow:
    0 14px 28px rgba(255, 255, 255, 0.12),
    0 10px 24px rgba(0, 0, 0, 0.28);
}

.primary-section-tabs__icon {
  width: 0.92rem;
  height: 0.92rem;
  flex: 0 0 auto;
}

@media (min-width: 901px) {
  .primary-section-tabs__list {
    overflow: visible;
  }
}

@media (max-width: 900px) {
  .primary-section-tabs-shell {
    --primary-section-tab-bar-height: 68px;
    --primary-section-tab-bar-gap: 0.7rem;
  }

  .primary-section-tabs {
    --primary-section-tab-bar-height: 68px;
    --primary-section-tab-bar-gap: 0.7rem;
    padding: 0.55rem clamp(0.9rem, 4vw, 1.25rem);
  }

  .primary-section-tabs__mobile-trigger {
    display: flex;
    width: 100%;
    align-items: center;
    gap: 0.75rem;
    border: 0;
    border-radius: 0.9rem;
    padding: 0.82rem 0.95rem;
    background: rgba(15, 23, 42, 0.88);
    color: #f8fafc;
    font-size: 0.92rem;
    font-weight: 800;
    cursor: pointer;
    text-align: left;
  }

  .primary-section-tabs__hamburger {
    display: inline-flex;
    width: 1.05rem;
    flex: 0 0 auto;
    flex-direction: column;
    gap: 0.22rem;
  }

  .primary-section-tabs__hamburger span {
    display: block;
    height: 2px;
    border-radius: 9999px;
    background: currentColor;
    transition:
      opacity 0.18s ease,
      transform 0.18s ease;
  }

  .primary-section-tabs--open .primary-section-tabs__hamburger span:nth-child(1) {
    transform: translateY(0.34rem) rotate(45deg);
  }

  .primary-section-tabs--open .primary-section-tabs__hamburger span:nth-child(2) {
    opacity: 0;
  }

  .primary-section-tabs--open .primary-section-tabs__hamburger span:nth-child(3) {
    transform: translateY(-0.34rem) rotate(-45deg);
  }

  .primary-section-tabs__mobile-text {
    min-width: 0;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .primary-section-tabs__list {
    display: none;
    flex-direction: column;
    align-items: stretch;
    justify-content: flex-start;
    gap: 0.42rem;
    padding-top: 0.48rem;
  }

  .primary-section-tabs--open .primary-section-tabs__list {
    display: flex;
  }

  .primary-section-tabs__item {
    width: 100%;
    justify-content: flex-start;
    padding: 0.82rem 0.95rem;
    font-size: 0.88rem;
  }
}
</style>
