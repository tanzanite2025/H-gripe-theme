<template>
  <div>
    <h2 class="company-page__title">Policies</h2>
    <p class="company-page__intro">
      Review our cookie, privacy, refund & return, and terms of service policies.
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
import { definePageMeta, useHead, useLocalePath } from '#imports'

const localePath = useLocalePath()

const sections = [
  {
    to: '/policies/cookie',
    icon: 'C',
    title: 'Cookie',
    description: 'Learn how we use cookies and similar technologies.',
    cta: 'Cookie policy',
  },
  {
    to: '/policies/privacy',
    icon: 'P',
    title: 'Privacy',
    description: 'How we collect, use, and protect your personal data.',
    cta: 'Privacy policy',
  },
  {
    to: '/policies/refund-return',
    icon: 'R',
    title: 'Refund & Return',
    description: 'Returns, refunds, and exchanges for your orders.',
    cta: 'Refund & return policy',
  },
  {
    to: '/policies/terms',
    icon: 'T',
    title: 'Terms',
    description: 'Terms of service for browsing and purchasing.',
    cta: 'Terms of service',
  },
]

definePageMeta({
  layout: 'products',
})

useHead({
  title: 'Policies',
})
</script>

<style scoped>
.company-page__title {
  margin: 0 0 0.75rem;
  font-size: var(--tz-type-page-title);
  line-height: 1.18;
  font-weight: 600;
  color: #f9fafb;
}

.company-page__intro {
  margin: 0 0 1.25rem;
  font-size: 0.95rem;
  color: var(--tz-text-secondary);
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
  background: rgba(15, 23, 42, 0.9);
  box-shadow:
    0 14px 40px -24px rgba(0, 0, 0, 1),
    0 0 18px rgba(0, 0, 0, 0.9);
  text-decoration: none;
  transition: background-color 0.15s ease, box-shadow 0.15s ease, transform 0.15s ease;
}

.company-card:hover {
  background: rgba(15, 23, 42, 0.98);
  box-shadow:
    0 18px 50px -26px rgba(0, 0, 0, 1),
    0 0 22px rgba(0, 0, 0, 0.95);
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
  color: var(--tz-text-secondary);
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
