<template>
  <section class="guide-category-child-route-navigation-cards" :aria-labelledby="headingId">
    <header class="guide-category-child-route-navigation-cards__header">
      <p class="guide-category-child-route-navigation-cards__eyebrow">{{ eyebrow }}</p>
      <h1 :id="headingId" class="guide-category-child-route-navigation-cards__title">
        {{ heading }}
      </h1>
      <p v-if="description" class="guide-category-child-route-navigation-cards__description">
        {{ description }}
      </p>
    </header>

    <ul class="guide-category-child-route-navigation-cards__list">
      <li
        v-for="card in cards"
        :key="card.id"
        class="guide-category-child-route-navigation-cards__list-item"
      >
        <NuxtLink
          class="guide-category-child-route-navigation-cards__card"
          :to="localizedRoute(card.to)"
        >
          <span class="guide-category-child-route-navigation-cards__card-topline">
            <span class="guide-category-child-route-navigation-cards__index" aria-hidden="true">
              {{ String(cards.indexOf(card) + 1).padStart(2, '0') }}
            </span>
            <Icon name="lucide:arrow-up-right" aria-hidden="true" />
          </span>
          <span class="guide-category-child-route-navigation-cards__card-title">
            {{ card.label }}
          </span>
          <span v-if="card.description" class="guide-category-child-route-navigation-cards__card-description">
            {{ card.description }}
          </span>
          <span class="guide-category-child-route-navigation-cards__card-action">
            {{ openLabel }}
            <Icon name="lucide:arrow-right" aria-hidden="true" />
          </span>
        </NuxtLink>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
import { useId } from 'vue'
import { useLocalePath } from '#imports'

export interface GuideCategoryChildRouteNavigationCard {
  id: string
  label: string
  description?: string
  to: string
}

withDefaults(defineProps<{
  eyebrow?: string
  heading: string
  description?: string
  cards: readonly GuideCategoryChildRouteNavigationCard[]
  openLabel?: string
}>(), {
  eyebrow: 'Guides',
  description: '',
  openLabel: 'Open guide',
})

const headingId = `guide-category-child-route-navigation-cards-${useId()}`
const localePath = useLocalePath()

const localizedRoute = (route: string) => localePath(route)
</script>

<style scoped>
.guide-category-child-route-navigation-cards {
  width: 100%;
  margin: 0 auto;
  color: var(--tz-text-primary);
}

.guide-category-child-route-navigation-cards__header {
  max-width: 760px;
  margin-bottom: 2rem;
}

.guide-category-child-route-navigation-cards__eyebrow {
  margin: 0 0 0.55rem;
  color: var(--tz-text-accent, #047857);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.12em;
  line-height: 1.2;
  text-transform: uppercase;
}

.guide-category-child-route-navigation-cards__title {
  margin: 0;
  font-size: clamp(1.8rem, 4vw, 2.8rem);
  font-weight: 750;
  letter-spacing: 0;
  line-height: 1.08;
}

.guide-category-child-route-navigation-cards__description {
  margin: 0.85rem 0 0;
  color: var(--tz-text-secondary);
  font-size: 1rem;
  line-height: 1.65;
}

.guide-category-child-route-navigation-cards__list {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(18rem, 100%), 1fr));
  gap: 1rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.guide-category-child-route-navigation-cards__list-item {
  min-width: 0;
}

.guide-category-child-route-navigation-cards__card {
  display: flex;
  min-height: 13rem;
  height: 100%;
  flex-direction: column;
  gap: 0.8rem;
  border: 1px solid var(--tz-border-subtle, #d8e1e8);
  border-radius: 8px;
  background: var(--tz-card-surface, #fff);
  box-shadow: 0 10px 24px rgba(20, 32, 43, 0.07);
  color: inherit;
  padding: 1.2rem;
  text-decoration: none;
  transition: border-color 0.18s ease, box-shadow 0.18s ease, transform 0.18s ease;
}

.guide-category-child-route-navigation-cards__card:hover,
.guide-category-child-route-navigation-cards__card:focus-visible {
  border-color: var(--tz-text-accent, #059669);
  box-shadow: 0 14px 28px rgba(20, 32, 43, 0.12);
  transform: translateY(-2px);
}

.guide-category-child-route-navigation-cards__card:focus-visible {
  outline: 2px solid var(--tz-text-accent, #059669);
  outline-offset: 3px;
}

.guide-category-child-route-navigation-cards__card-topline,
.guide-category-child-route-navigation-cards__card-action {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.guide-category-child-route-navigation-cards__card-topline {
  color: var(--tz-text-accent, #059669);
}

.guide-category-child-route-navigation-cards__card-topline :deep(svg) {
  width: 1.1rem;
  height: 1.1rem;
}

.guide-category-child-route-navigation-cards__index {
  color: var(--tz-text-muted);
  font-size: 0.75rem;
  font-weight: 800;
  letter-spacing: 0.1em;
}

.guide-category-child-route-navigation-cards__card-title {
  max-width: 26rem;
  font-size: 1.2rem;
  font-weight: 750;
  line-height: 1.25;
}

.guide-category-child-route-navigation-cards__card-description {
  flex: 1 1 auto;
  color: var(--tz-text-secondary);
  font-size: 0.9rem;
  line-height: 1.55;
}

.guide-category-child-route-navigation-cards__card-action {
  justify-content: flex-start;
  margin-top: auto;
  color: var(--tz-text-accent, #047857);
  font-size: 0.78rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}

.guide-category-child-route-navigation-cards__card-action :deep(svg) {
  width: 0.95rem;
  height: 0.95rem;
}

@media (max-width: 640px) {
  .guide-category-child-route-navigation-cards__header {
    margin-bottom: 1.35rem;
  }

  .guide-category-child-route-navigation-cards__card {
    min-height: 11rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .guide-category-child-route-navigation-cards__card {
    transition: none;
  }
}
</style>
