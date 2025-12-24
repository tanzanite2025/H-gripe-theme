<template>
  <div>
    <h2 class="company-page__title">{{ t('company.index.title') }}</h2>
    <p class="company-page__intro">
      {{ t('company.index.intro') }}
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
import { computed } from 'vue'
import { definePageMeta, useHead, useI18n, useLocalePath } from '#imports'

const localePath = useLocalePath()
const { t } = useI18n()

const sections = computed(() => [
  {
    to: '/company/about',
    icon: 'A',
    title: t('company.nav.about'),
    description: t('company.index.cards.about.description'),
    cta: t('company.index.cards.about.cta'),
  },
  {
    to: '/company/ourstory',
    icon: 'O',
    title: t('company.nav.ourStory'),
    description: t('company.index.cards.ourStory.description'),
    cta: t('company.index.cards.ourStory.cta'),
  },
  {
    to: '/company/global-partners',
    icon: 'G',
    title: t('company.nav.globalPartners'),
    description: t('company.index.cards.globalPartners.description'),
    cta: t('company.index.cards.globalPartners.cta'),
  },
  {
    to: '/company/oem-odm',
    icon: 'O',
    title: t('company.nav.oemOdm'),
    description: t('company.index.cards.oemOdm.description'),
    cta: t('company.index.cards.oemOdm.cta'),
  },
  {
    to: '/company/certificates',
    icon: 'C',
    title: t('company.nav.certificates'),
    description: t('company.index.cards.certificates.description'),
    cta: t('company.index.cards.certificates.cta'),
  },
  {
    to: '/company/contact',
    icon: 'C',
    title: t('company.nav.contact'),
    description: t('company.index.cards.contact.description'),
    cta: t('company.index.cards.contact.cta'),
  },
])

definePageMeta({
  layout: 'products',
})

useHead({
  title: t('company.index.title'),
})
</script>

<style scoped>
.company-page__title {
  margin: 0 0 0.75rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: #f9fafb;
  text-align: center;
}

.company-page__intro {
  margin: 0 0 1.25rem;
  font-size: 0.95rem;
  color: rgba(148, 163, 184, 0.9);
  text-align: center;
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
