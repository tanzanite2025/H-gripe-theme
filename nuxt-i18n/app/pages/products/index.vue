<template>
  <div>
    <h2 class="products-page__title">Product Center</h2>
    <p class="products-page__intro">
      Explore Tanzanite products, tools, charts, and wheelbuilding guides from one place.
      Choose a section below to open the shop, documentation, or blog.
    </p>

    <section class="products-grid">
      <NuxtLink
        v-for="section in sections"
        :key="section.to"
        :to="localePath(section.to)"
        class="products-card"
      >
        <div class="products-card__icon">
          {{ section.icon }}
        </div>
        <div class="products-card__body">
          <h3 class="products-card__title">
            {{ section.title }}
          </h3>
          <p class="products-card__description">
            {{ section.description }}
          </p>
          <span class="products-card__cta">
            {{ section.cta }}
          </span>
        </div>
      </NuxtLink>
    </section>
  </div>
</template>

<script setup lang="ts">
import { useLocalePath } from '#imports'

const localePath = useLocalePath()

const sections = [
  {
    to: '/shop',
    icon: 'S',
    title: 'Shop',
    description: 'Browse Tanzanite products and tools available through the shop.',
    cta: 'Open shop',
  },
  {
    to: '/guides/sizecharts',
    icon: 'C',
    title: 'Tire Size Charts',
    description: 'Reference charts for common tire and rim sizes.',
    cta: 'View size charts',
  },
  {
    to: '/guides/technical',
    icon: 'D',
    title: 'Technical',
    description: 'Technical notes, formulas, and reference docs for wheelbuilding.',
    cta: 'Open technical docs',
  },
  {
    to: '/guides/wheelset-buyers',
    icon: 'G',
    title: 'Wheelset Buyers Guide',
    description: 'High-level guide to choosing the right wheelset.',
    cta: 'Read buyers guide',
  },
  {
    to: '/spoke-calculator',
    icon: 'S',
    title: 'Spoke Calculator',
    description: 'Use the spoke calculator to estimate spoke lengths for your wheel builds.',
    cta: 'Open spoke calculator',
  },
]

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Products',
})
</script>

<style scoped>
.products-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.products-page__intro {
  margin: 0 0 1.25rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.products-grid {
  margin-top: 0.5rem;
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
}

.products-card {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 1rem 1.1rem;
  border-radius: 1rem;
  border: 1px solid rgba(148, 163, 184, 0.3);
  background: rgba(15, 23, 42, 0.75);
  text-decoration: none;
  transition: border-color 0.15s ease, background-color 0.15s ease, transform 0.15s ease;
}

.products-card:hover {
  border-color: rgba(148, 163, 184, 0.6);
  background: rgba(15, 23, 42, 0.95);
  transform: translateY(-1px);
}

.products-card__icon {
  margin-top: 0.1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 9999px;
  background: linear-gradient(135deg, #38bdf8, #22c55e);
  color: #020617;
  font-size: 0.9rem;
  font-weight: 700;
  box-shadow: 0 0 16px rgba(56, 189, 248, 0.35);
}

.products-card__body {
  flex: 1;
}

.products-card__title {
  margin: 0 0 0.25rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: #ffffff;
}

.products-card__description {
  margin: 0;
  font-size: 0.85rem;
  color: rgba(148, 163, 184, 0.9);
}

.products-card__cta {
  margin-top: 0.5rem;
  display: inline-flex;
  align-items: center;
  font-size: 0.78rem;
  font-weight: 500;
  color: #7dd3fc;
}

@media (max-width: 768px) {
  .products-grid {
    grid-template-columns: 1fr;
  }
}
</style>
