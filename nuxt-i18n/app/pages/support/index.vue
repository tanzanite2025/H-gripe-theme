<template>
  <div>
    <h2 class="support-page__title">Support Center</h2>
    <p class="support-page__intro">
      Find help with orders, payments, warranties, manuals, feedback, and tools. Choose a
      section below to view detailed information or tools for that topic.
    </p>

    <section class="support-grid">
      <NuxtLink
        v-for="section in sections"
        :key="section.to"
        :to="localePath(section.to)"
        class="support-card"
      >
        <div class="support-card__icon">
          {{ section.icon }}
        </div>
        <div class="support-card__body">
          <h3 class="support-card__title">
            {{ section.title }}
          </h3>
          <p class="support-card__description">
            {{ section.description }}
          </p>
          <span class="support-card__cta">
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
    to: '/support/faqs',
    icon: 'Q',
    title: "All FAQ'S",
    description:
      'Browse common questions and quick answers about orders, products, and service.',
    cta: 'Browse FAQs',
  },
  {
    to: '/support/payment',
    icon: 'P',
    title: 'Payment',
    description:
      'Available payment methods, processing times, and basic checkout tips.',
    cta: 'View payment info',
  },
  {
    to: '/support/after-sales',
    icon: 'A',
    title: 'After sales',
    description:
      'Returns, exchanges, repairs, and other after-sales support options.',
    cta: 'View after-sales options',
  },
  {
    to: '/support/warranty',
    icon: 'W',
    title: 'Warranty',
    description:
      'Overview of Tanzanite warranty coverage, terms, and claim process.',
    cta: 'Read warranty details',
  },
  {
    to: '/support/product-feedback',
    icon: 'F',
    title: 'Product Feedback',
    description:
      'Share feedback about products and your riding experience with Tanzanite.',
    cta: 'Give feedback',
  },
]

definePageMeta({
  layout: 'support',
})

useHead({
  title: 'Support',
})
</script>

<style scoped>
.support-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
  text-align: center;
}

.support-page__intro {
  margin: 0 0 1.25rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
  text-align: center;
}

.support-grid {
  margin-top: 0.5rem;
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
}

.support-card {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 1rem 1.1rem;
  border-radius: 1rem;
  border: none;
  background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
  text-decoration: none;
  box-shadow: 0 10px 26px -14px rgba(0, 0, 0, 0.95);
  transition: box-shadow 0.15s ease, background-color 0.15s ease, transform 0.15s ease;
}

.support-card:hover {
  background: radial-gradient(circle at top left, rgba(31, 41, 55, 0.98), rgba(15, 23, 42, 0.99));
  box-shadow: 0 14px 32px -16px rgba(0, 0, 0, 1);
  transform: translateY(-1px);
}

.support-card__icon {
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

.support-card__body {
  flex: 1;
}

.support-card__title {
  margin: 0 0 0.25rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: #ffffff;
}

.support-card__description {
  margin: 0;
  font-size: 0.85rem;
  color: rgba(148, 163, 184, 0.9);
}

.support-card__cta {
  margin-top: 0.5rem;
  display: inline-flex;
  align-items: center;
  font-size: 0.78rem;
  font-weight: 500;
  color: #7dd3fc;
}

@media (max-width: 768px) {
  .support-grid {
    grid-template-columns: 1fr;
  }
}
</style>
