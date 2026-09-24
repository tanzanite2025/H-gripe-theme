<template>
  <div class="website-name-page">
    <section class="website-name-content" aria-labelledby="website-name-title">
      <div class="website-name-content__heading">
        <p v-if="copy.status" class="website-name-status">{{ copy.status }}</p>
        <p class="website-name-eyebrow">{{ copy.eyebrow }}</p>
        <h1 id="website-name-title">{{ copy.title }}</h1>
      </div>

      <div v-if="copy.intro || copy.body || copy.note" class="website-name-content__copy">
        <p v-if="copy.intro" class="website-name-content__intro">{{ copy.intro }}</p>
        <SafeRichText v-if="copy.body" class="website-name-content__body" :html="copy.body" />
        <span v-if="copy.note" class="website-name-content__note">{{ copy.note }}</span>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { definePageMeta, useHead, useI18n } from '#imports'
import { useWebsiteNameSettings } from '~/composables/useWebsiteNameSettings'

const { locale } = useI18n()
const { websiteNameSettings: copy } = useWebsiteNameSettings(locale)

definePageMeta({
  layout: 'products',
  footerLabelKey: 'footer.links.whyThisName',
  footerLabelFallback: 'Why This Name',
})

useHead(() => ({
  title: copy.value.title,
}))
</script>

<style scoped>
.website-name-page {
  width: min(100%, 76rem);
  margin-inline: auto;
  color: var(--tz-text-primary);
}

.website-name-content {
  display: grid;
  grid-template-columns: minmax(11rem, 0.4fr) minmax(0, 1fr);
  gap: 3rem;
  padding: clamp(1.5rem, 4vw, 3rem) 0 1.5rem;
}

.website-name-eyebrow {
  margin: 0;
  color: var(--tz-text-accent);
  font-size: 0.66rem;
  font-weight: 850;
  letter-spacing: 0.18em;
  line-height: 1.4;
  text-transform: uppercase;
}

.website-name-status {
  margin: 0 0 0.75rem;
  color: var(--tz-text-muted);
  font-size: 0.72rem;
  font-weight: 750;
  letter-spacing: 0.08em;
  line-height: 1.4;
  text-transform: uppercase;
}

.website-name-content h1 {
  margin: 0.5rem 0 0;
  color: var(--tz-text-primary);
  font-size: var(--tz-type-website-title);
  font-weight: 800;
  letter-spacing: 0;
  line-height: 1.08;
}

.website-name-content__copy {
  border-top: 1px solid var(--tz-border-subtle);
  border-bottom: 1px solid var(--tz-border-subtle);
  padding: 1.25rem 0 1.5rem;
}

.website-name-content__copy p {
  max-width: 48rem;
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 1.08rem;
  line-height: 1.8;
  white-space: pre-line;
}

.website-name-content__body {
  max-width: 48rem;
  margin-top: 1.5rem;
  color: var(--tz-text-secondary);
  font-size: 1.08rem;
  line-height: 1.8;
}

.website-name-content__body :deep(p),
.website-name-content__body :deep(ul),
.website-name-content__body :deep(ol),
.website-name-content__body :deep(blockquote) {
  margin: 0 0 1rem;
}

.website-name-content__body :deep(h2),
.website-name-content__body :deep(h3) {
  margin: 1.5rem 0 0.65rem;
  color: var(--tz-text-primary);
  font-weight: 800;
  line-height: 1.25;
}

.website-name-content__body :deep(h2) {
  font-size: 1.45rem;
}

.website-name-content__body :deep(h3) {
  font-size: 1.2rem;
}

.website-name-content__body :deep(ul),
.website-name-content__body :deep(ol) {
  padding-left: 1.35rem;
  list-style-position: outside;
}

.website-name-content__body :deep(ul) {
  list-style-type: disc;
}

.website-name-content__body :deep(ol) {
  list-style-type: decimal;
}

.website-name-content__body :deep(blockquote) {
  margin-left: 0;
  border-left: 3px solid var(--tz-text-accent);
  padding-left: 1rem;
  color: var(--tz-text-muted);
}

.website-name-content__body :deep(a) {
  color: var(--tz-text-accent);
  text-decoration: underline;
}

.website-name-content__note {
  display: inline-flex;
  margin-top: 2rem;
  color: var(--tz-text-muted);
  font-size: 0.75rem;
  font-weight: 750;
  line-height: 1.3;
}

@media (max-width: 767px) {
  .website-name-content {
    display: block;
    padding-top: 1.5rem;
  }

  .website-name-content__heading {
    margin-bottom: 1.5rem;
  }

  .website-name-content h1 {
    font-size: var(--tz-type-website-title-mobile);
  }
}
</style>
