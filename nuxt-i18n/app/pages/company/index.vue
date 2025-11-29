<template>
  <div>
    <h2 class="company-page__title">Company</h2>
    <p class="company-page__intro">
      Learn more about Tanzanite, our story, membership & points program, and the
      picture warehouse. Choose a section below to explore.
    </p>

    <section class="company-grid">
      <NuxtLink
        v-for="section in sections"
        :key="section.to"
        :to="localePath(section.to)"
        class="company-card"
      >
        <div class="company-card__icon">
          {{ section.icon }}
        </div>
        <div class="company-card__body">
          <h3 class="company-card__title">
            {{ section.title }}
          </h3>
          <p class="company-card__description">
            {{ section.description }}
          </p>
          <span class="company-card__cta">
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
    to: '/company/ourstory',
    icon: 'O',
    title: 'Our Story',
    description: 'Background of Tanzanite and the vision behind our products.',
    cta: 'Read our story',
  },
  {
    to: '/company/membershipandpoints',
    icon: 'M',
    title: 'Membership and Points',
    description: 'Overview of the membership program, benefits, and earning points.',
    cta: 'View membership info',
  },
  {
    to: '/company/picture-warehouse',
    icon: 'P',
    title: 'Picture warehouse',
    description: 'Central library for Tanzanite product and lifestyle photos.',
    cta: 'Open picture warehouse',
  },
]

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Company',
})
</script>

<style scoped>
.company-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
}

.company-page__intro {
  margin: 0 0 1.25rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
}

.company-grid {
  margin-top: 0.5rem;
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
}

.company-card {
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

.company-card:hover {
  border-color: rgba(148, 163, 184, 0.6);
  background: rgba(15, 23, 42, 0.95);
  transform: translateY(-1px);
}

.company-card__icon {
  margin-top: 0.1rem;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.5rem;
  height: 2.5rem;
  border-radius: 9999px;
  background: linear-gradient(135deg, #a855f7, #6366f1);
  color: #020617;
  font-size: 0.9rem;
  font-weight: 700;
  box-shadow: 0 0 16px rgba(129, 140, 248, 0.45);
}

.company-card__body {
  flex: 1;
}

.company-card__title {
  margin: 0 0 0.25rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: #ffffff;
}

.company-card__description {
  margin: 0;
  font-size: 0.85rem;
  color: rgba(148, 163, 184, 0.9);
}

.company-card__cta {
  margin-top: 0.5rem;
  display: inline-flex;
  align-items: center;
  font-size: 0.78rem;
  font-weight: 500;
  color: #a5b4fc;
}

@media (max-width: 768px) {
  .company-grid {
    grid-template-columns: 1fr;
  }
}
</style>
