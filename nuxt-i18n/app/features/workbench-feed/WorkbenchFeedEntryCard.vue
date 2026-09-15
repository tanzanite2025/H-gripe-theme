<template>
  <article class="workbench-feed-entry">
    <header class="workbench-feed-entry__header">
      <div>
        <span class="workbench-feed-entry__number">{{ entry.entry_number }}</span>
        <time class="workbench-feed-entry__date" :datetime="entry.published_at">{{ formatDate(entry.published_at, entry.locale) }}</time>
      </div>
    </header>

    <div v-if="entry.media.length" class="workbench-feed-entry__media" :class="`workbench-feed-entry__media--${Math.min(entry.media.length, 4)}`">
      <button
        v-for="(media, index) in entry.media"
        :key="media.id"
        type="button"
        class="workbench-feed-entry__media-button"
        :aria-label="`${media.caption || 'Workshop photo'} ${index + 1}`"
        @click="$emit('open-media', { entry, index })"
      >
        <StorefrontImage
          :src="media.file_path"
          :alt="media.caption || entry.entry_number"
          :width="media.width"
          :height="media.height"
          preset="content"
          :loading="index === 0 ? 'eager' : 'lazy'"
          class="workbench-feed-entry__image"
        />
        <span v-if="media.caption" class="workbench-feed-entry__caption">{{ media.caption }}</span>
      </button>
    </div>

    <p class="workbench-feed-entry__content">{{ entry.content }}</p>

    <div v-if="entry.tags.length" class="workbench-feed-entry__topics" aria-label="Entry topics">
      <span v-for="tag in entry.tags" :key="tag">{{ tag }}</span>
    </div>

    <footer v-if="entry.tagged_products.length" class="workbench-feed-entry__products">
      <span class="workbench-feed-entry__products-label">Mentioned / used</span>
      <button
        v-for="product in entry.tagged_products"
        :key="product.id"
        type="button"
        class="workbench-feed-entry__product-pill"
        :disabled="!product.product_slug"
        @click="$emit('open-product', product)"
      >
        <span class="workbench-feed-entry__product-dot" aria-hidden="true"></span>
        <span class="workbench-feed-entry__product-title">{{ product.display_title }}</span>
        <span class="workbench-feed-entry__product-price">{{ formatPrice(product.price, product.currency) }}</span>
      </button>
    </footer>
  </article>
</template>

<script setup lang="ts">
import type { WorkbenchFeedEntry, WorkbenchTaggedProduct } from './types'

defineEmits<{
  (event: 'open-media', value: { entry: WorkbenchFeedEntry; index: number }): void
  (event: 'open-product', value: WorkbenchTaggedProduct): void
}>()

defineProps<{
  entry: WorkbenchFeedEntry
}>()

const formatDate = (value: string, locale: string) => {
  const date = new Date(value)
  const dateLocale = String(locale || 'en').replace('_', '-')
  return Number.isNaN(date.getTime()) ? value : new Intl.DateTimeFormat(dateLocale, {
    year: 'numeric', month: 'short', day: 'numeric',
  }).format(date)
}

const formatPrice = (value: number, currency: string) => {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency: /^[A-Z]{3}$/.test(currency) ? currency : 'USD',
      maximumFractionDigits: 2,
    }).format(Number(value || 0))
  } catch {
    return `${currency || 'USD'} ${Number(value || 0).toFixed(2)}`
  }
}
</script>

<style scoped>
.workbench-feed-entry {
  display: grid;
  gap: 1rem;
  padding: 1.35rem 0;
  border-block: 1px solid var(--tz-border-subtle);
}

.workbench-feed-entry + .workbench-feed-entry { border-top: 0; }
.workbench-feed-entry__header { display: flex; justify-content: space-between; }
.workbench-feed-entry__number { color: var(--tz-text-primary); font: 800 0.72rem/1.2 var(--tz-font-ui); letter-spacing: .08em; }
.workbench-feed-entry__date { margin-left: .7rem; color: var(--tz-text-muted); font-size: .75rem; }
.workbench-feed-entry__media { display: grid; gap: .45rem; grid-template-columns: 1fr; }
.workbench-feed-entry__media--2, .workbench-feed-entry__media--3, .workbench-feed-entry__media--4 { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.workbench-feed-entry__media-button { position: relative; min-width: 0; overflow: hidden; padding: 0; border: 1px solid var(--tz-border-subtle); background: var(--tz-surface-subtle); cursor: zoom-in; }
.workbench-feed-entry__media--1 .workbench-feed-entry__media-button { max-height: 38rem; }
.workbench-feed-entry__image { display: block; width: 100%; height: 100%; max-height: 38rem; object-fit: cover; }
.workbench-feed-entry__caption { position: absolute; right: .45rem; bottom: .45rem; max-width: calc(100% - .9rem); padding: .35rem .5rem; color: #fff; background: rgb(0 0 0 / 65%); font-size: .68rem; line-height: 1.35; text-align: left; }
.workbench-feed-entry__content { margin: 0; color: var(--tz-text-primary); font-size: .98rem; line-height: 1.7; white-space: pre-wrap; }
.workbench-feed-entry__topics { display: flex; flex-wrap: wrap; gap: .4rem; color: var(--tz-text-muted); font-size: .72rem; }
.workbench-feed-entry__topics span { padding: .2rem .45rem; border: 1px solid var(--tz-border-subtle); }
.workbench-feed-entry__products { display: flex; flex-wrap: wrap; align-items: center; gap: .45rem; padding-top: .25rem; }
.workbench-feed-entry__products-label { flex: 0 0 100%; color: var(--tz-text-muted); font-size: .67rem; font-weight: 800; letter-spacing: .06em; text-transform: uppercase; }
.workbench-feed-entry__product-pill { display: inline-flex; min-width: 0; max-width: 100%; align-items: center; gap: .4rem; padding: .42rem .65rem; border: 1px solid rgb(5 150 105 / 35%); border-radius: 999px; color: var(--tz-text-primary); background: rgb(5 150 105 / 10%); font-size: .73rem; text-align: left; }
.workbench-feed-entry__product-pill:hover:not(:disabled) { border-color: rgb(5 150 105 / 70%); background: rgb(5 150 105 / 17%); }
.workbench-feed-entry__product-pill:disabled { cursor: default; opacity: .58; }
.workbench-feed-entry__product-dot { width: .45rem; height: .45rem; flex: 0 0 auto; border-radius: 50%; background: #059669; }
.workbench-feed-entry__product-title { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.workbench-feed-entry__product-price { flex: 0 0 auto; color: #047857; font-weight: 800; }

@media (max-width: 520px) {
  .workbench-feed-entry__media--2, .workbench-feed-entry__media--3, .workbench-feed-entry__media--4 { grid-template-columns: 1fr; }
  .workbench-feed-entry__media--2 .workbench-feed-entry__media-button, .workbench-feed-entry__media--3 .workbench-feed-entry__media-button, .workbench-feed-entry__media--4 .workbench-feed-entry__media-button { max-height: 25rem; }
  .workbench-feed-entry__product-pill { width: 100%; }
  .workbench-feed-entry__product-title { flex: 1 1 auto; white-space: normal; }
  .workbench-feed-entry__product-price { margin-left: auto; }
}
</style>
