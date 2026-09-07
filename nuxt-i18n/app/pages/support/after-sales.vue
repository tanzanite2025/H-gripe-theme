<template>
  <div class="after-sales-page">
    <header class="after-sales-page__header">
      <p class="after-sales-page__eyebrow">
        <Icon name="lucide:shield-check" aria-hidden="true" />
        {{ t('supportAfterSales.eyebrow') }}
      </p>
      <h1 class="after-sales-page__title">{{ t('supportAfterSales.title') }}</h1>
      <p class="after-sales-page__intro">{{ t('supportAfterSales.intro') }}</p>
      <p class="after-sales-page__updated">{{ t('supportAfterSales.updated') }}</p>
    </header>

    <section class="after-sales-summary" :aria-label="t('supportAfterSales.summaryLabel')">
      <article class="after-sales-summary__item">
        <Icon name="lucide:rotate-ccw" aria-hidden="true" />
        <div>
          <h2>{{ t('supportAfterSales.summary.returns.title') }}</h2>
          <p>{{ t('supportAfterSales.summary.returns.description') }}</p>
        </div>
      </article>
      <article class="after-sales-summary__item">
        <Icon name="lucide:badge-check" aria-hidden="true" />
        <div>
          <h2>{{ t('supportAfterSales.summary.warranty.title') }}</h2>
          <p>{{ t('supportAfterSales.summary.warranty.description') }}</p>
        </div>
      </article>
      <article class="after-sales-summary__item">
        <Icon name="lucide:life-buoy" aria-hidden="true" />
        <div>
          <h2>{{ t('supportAfterSales.summary.crashReplacement.title') }}</h2>
          <p>{{ t('supportAfterSales.summary.crashReplacement.description') }}</p>
        </div>
      </article>
      <article class="after-sales-summary__item">
        <Icon name="lucide:package-check" aria-hidden="true" />
        <div>
          <h2>{{ t('supportAfterSales.summary.delivery.title') }}</h2>
          <p>{{ t('supportAfterSales.summary.delivery.description') }}</p>
        </div>
      </article>
    </section>

    <div class="after-sales-policy">
      <section id="returns" class="after-sales-policy__section">
        <h2>{{ t('supportAfterSales.sections.returns.title') }}</h2>

        <h3>{{ t('supportAfterSales.sections.returns.stockedTitle') }}</h3>
        <ul>
          <li>{{ t('supportAfterSales.sections.returns.stocked.0') }}</li>
          <li>{{ t('supportAfterSales.sections.returns.stocked.1') }}</li>
          <li>{{ t('supportAfterSales.sections.returns.stocked.2') }}</li>
          <li>{{ t('supportAfterSales.sections.returns.stocked.3') }}</li>
        </ul>

        <h3>{{ t('supportAfterSales.sections.returns.customTitle') }}</h3>
        <ul>
          <li>{{ t('supportAfterSales.sections.returns.custom.0') }}</li>
          <li>{{ t('supportAfterSales.sections.returns.custom.1') }}</li>
          <li>{{ t('supportAfterSales.sections.returns.custom.2') }}</li>
        </ul>

        <p>{{ t('supportAfterSales.sections.returns.refundIntro') }}<NuxtLink :to="localePath('/policies/refund-cancellation')">{{ t('supportAfterSales.sections.returns.refundLink') }}</NuxtLink>{{ t('supportAfterSales.sections.returns.refundSuffix') }}</p>
      </section>

      <section id="warranty" class="after-sales-policy__section">
        <h2>{{ t('supportAfterSales.sections.warranty.title') }}</h2>
        <ul>
          <li>{{ t('supportAfterSales.sections.warranty.bullets.0') }}</li>
          <li>{{ t('supportAfterSales.sections.warranty.bullets.1') }}</li>
          <li>{{ t('supportAfterSales.sections.warranty.bullets.2') }}</li>
          <li>{{ t('supportAfterSales.sections.warranty.bullets.3') }}</li>
          <li>{{ t('supportAfterSales.sections.warranty.bullets.4') }}</li>
          <li>{{ t('supportAfterSales.sections.warranty.bullets.5') }}</li>
          <li>{{ t('supportAfterSales.sections.warranty.bullets.6') }}</li>
        </ul>
        <p>{{ t('supportAfterSales.sections.warranty.shipping') }}</p>
      </section>

      <section id="exclusions" class="after-sales-policy__section">
        <h2>{{ t('supportAfterSales.sections.exclusions.title') }}</h2>
        <ul>
          <li>{{ t('supportAfterSales.sections.exclusions.bullets.0') }}</li>
          <li>{{ t('supportAfterSales.sections.exclusions.bullets.1') }}</li>
          <li>{{ t('supportAfterSales.sections.exclusions.bullets.2') }}</li>
          <li>{{ t('supportAfterSales.sections.exclusions.bullets.3') }}</li>
          <li>{{ t('supportAfterSales.sections.exclusions.bullets.4') }}</li>
        </ul>
      </section>

      <section id="accidental-damage" class="after-sales-policy__section">
        <h2>{{ t('supportAfterSales.sections.accidental.title') }}</h2>
        <ul>
          <li>{{ t('supportAfterSales.sections.accidental.bullets.0') }}</li>
          <li>{{ t('supportAfterSales.sections.accidental.bullets.1') }}</li>
          <li>{{ t('supportAfterSales.sections.accidental.bullets.2') }}</li>
          <li>{{ t('supportAfterSales.sections.accidental.bullets.3') }}</li>
        </ul>
      </section>

      <section id="delivery" class="after-sales-policy__section">
        <h2>{{ t('supportAfterSales.sections.delivery.title') }}</h2>
        <ul>
          <li>{{ t('supportAfterSales.sections.delivery.bullets.0') }}</li>
          <li>{{ t('supportAfterSales.sections.delivery.bullets.1') }}</li>
          <li>{{ t('supportAfterSales.sections.delivery.bullets.2') }}</li>
          <li>{{ t('supportAfterSales.sections.delivery.bullets.3') }}</li>
        </ul>
      </section>

      <section id="claims" class="after-sales-policy__section">
        <h2>{{ t('supportAfterSales.sections.claims.title') }}</h2>
        <ol>
          <li>{{ t('supportAfterSales.sections.claims.prepare') }}</li>
          <li>{{ t('supportAfterSales.sections.claims.warrantyPrefix') }}<NuxtLink :to="localePath('/support/warranty')">{{ t('supportAfterSales.sections.claims.warrantyLink') }}</NuxtLink>{{ t('supportAfterSales.sections.claims.warrantySuffix') }}</li>
          <li>{{ t('supportAfterSales.sections.claims.authorization') }}</li>
          <li>{{ t('supportAfterSales.sections.claims.review') }}</li>
        </ol>
        <p v-if="supportEmail">{{ t('supportAfterSales.sections.claims.contactPrefix') }}<NuxtLink :to="localePath('/company/contact')">{{ t('supportAfterSales.sections.claims.contactLink') }}</NuxtLink>{{ t('supportAfterSales.sections.claims.contactEmailJoiner') }}<a :href="`mailto:${supportEmail}`">{{ supportEmail }}</a>{{ t('supportAfterSales.sections.claims.contactSuffix') }}</p>
        <p v-else>{{ t('supportAfterSales.sections.claims.contactPrefix') }}<NuxtLink :to="localePath('/company/contact')">{{ t('supportAfterSales.sections.claims.contactLink') }}</NuxtLink>{{ t('supportAfterSales.sections.claims.contactSuffix') }}</p>
      </section>
    </div>

    <footer class="after-sales-page__footer">
      <NuxtLink :to="localePath('/support/warranty')" class="after-sales-page__footer-link">
        <Icon name="lucide:arrow-right" aria-hidden="true" />
        {{ t('supportAfterSales.footer.warrantyLink') }}
      </NuxtLink>
      <NuxtLink :to="localePath('/policies/refund-cancellation')" class="after-sales-page__footer-link">
        <Icon name="lucide:arrow-right" aria-hidden="true" />
        {{ t('supportAfterSales.footer.refundLink') }}
      </NuxtLink>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useHead, useI18n, useLocalePath } from '#imports'
import { useSiteSettings } from '~/composables/usePublicSettings'
import { usePageMessages } from '~/composables/usePageMessages'

definePageMeta({
  layout: 'support',
  footerLabelKey: 'support.nav.afterSales',
  footerLabelFallback: 'After sales',
})

const { locale, t } = useI18n()
const localePath = useLocalePath()
const { loadPageMessages } = usePageMessages('supportAfterSales')
const { siteSettings } = useSiteSettings()
const supportEmail = computed(() => siteSettings.value.contactEmail?.trim() || '')

await loadPageMessages(locale.value)

watch(locale, (nextLocale) => {
  void loadPageMessages(nextLocale)
})

useHead(() => ({
  title: t('supportAfterSales.metaTitle'),
}))
</script>

<style scoped>
.after-sales-page {
  width: 100%;
  color: var(--tz-text-secondary);
}

.after-sales-page__header {
  max-width: 54rem;
  margin-bottom: 2rem;
}

.after-sales-page__eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  margin: 0 0 0.7rem;
  color: var(--tz-site-accent);
  font-size: 0.72rem;
  font-weight: 800;
  letter-spacing: 0.14em;
  line-height: 1.3;
  text-transform: uppercase;
}

.after-sales-page__eyebrow :deep(svg) {
  width: 1rem;
  height: 1rem;
}

.after-sales-page__title {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: clamp(1.75rem, 3vw, 2.6rem);
  font-weight: 800;
  line-height: 1.12;
}

.after-sales-page__intro {
  max-width: 48rem;
  margin: 1rem 0 0;
  font-size: 1rem;
  line-height: 1.75;
}

.after-sales-page__updated {
  margin: 0.85rem 0 0;
  color: var(--tz-text-muted);
  font-size: 0.78rem;
  line-height: 1.6;
}

.after-sales-summary {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 0.75rem;
  margin-bottom: 2.5rem;
}

.after-sales-summary__item {
  display: flex;
  min-width: 0;
  gap: 0.75rem;
  padding: 1rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 8px;
  background: var(--tz-surface-card);
}

.after-sales-summary__item > :deep(svg) {
  width: 1.25rem;
  height: 1.25rem;
  flex: 0 0 auto;
  margin-top: 0.1rem;
  color: var(--tz-site-accent);
}

.after-sales-summary__item h2 {
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 0.95rem;
  font-weight: 800;
  line-height: 1.3;
}

.after-sales-summary__item p {
  margin: 0.35rem 0 0;
  font-size: 0.8rem;
  line-height: 1.55;
}

.after-sales-policy {
  max-width: 62rem;
}

.after-sales-policy__section {
  padding: 1.75rem 0;
  border-top: 1px solid var(--tz-border-subtle);
}

.after-sales-policy__section:first-child {
  padding-top: 0;
  border-top: 0;
}

.after-sales-policy__section h2 {
  margin: 0 0 1rem;
  color: var(--tz-text-primary);
  font-size: 1.2rem;
  font-weight: 800;
  line-height: 1.3;
}

.after-sales-policy__section h3 {
  margin: 1.25rem 0 0.55rem;
  color: var(--tz-text-primary);
  font-size: 0.98rem;
  font-weight: 750;
  line-height: 1.4;
}

.after-sales-policy__section ul,
.after-sales-policy__section ol {
  display: grid;
  gap: 0.55rem;
  margin: 0;
  padding-inline-start: 1.25rem;
  font-size: 0.92rem;
  line-height: 1.7;
}

.after-sales-policy__section p {
  margin: 1rem 0 0;
  font-size: 0.92rem;
  line-height: 1.75;
}

.after-sales-policy__section a,
.after-sales-page__footer-link {
  color: var(--tz-site-accent);
  text-decoration: underline;
  text-underline-offset: 0.18em;
}

.after-sales-page__footer {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem 1.5rem;
  margin-top: 0.5rem;
  padding-top: 1.5rem;
  border-top: 1px solid var(--tz-border-subtle);
}

.after-sales-page__footer-link {
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  font-size: 0.9rem;
  font-weight: 700;
}

.after-sales-page__footer-link :deep(svg) {
  width: 1rem;
  height: 1rem;
  text-decoration: none;
}

@media (max-width: 960px) {
  .after-sales-summary {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 640px) {
  .after-sales-summary {
    grid-template-columns: minmax(0, 1fr);
  }

  .after-sales-page__title {
    font-size: 1.8rem;
  }
}
</style>
