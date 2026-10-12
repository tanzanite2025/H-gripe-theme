<template>
  <section
    v-if="loading && !hasSavedAnswers"
    class="storefront-faq-placeholder storefront-faq-placeholder--loading"
    aria-busy="true"
    role="status"
  >
    <p class="storefront-faq-placeholder__message">{{ t('faq.ui.loadingAnswers') }}</p>
    <div class="storefront-faq-placeholder__layout" aria-hidden="true">
      <div class="storefront-faq-placeholder__questions">
        <div v-for="index in 3" :key="index" class="storefront-faq-placeholder__question">
          <span class="storefront-faq-placeholder__line" />
          <span class="storefront-faq-placeholder__line storefront-faq-placeholder__line--short" />
        </div>
      </div>
      <div class="storefront-faq-placeholder__answer">
        <span v-for="index in 4" :key="index" class="storefront-faq-placeholder__line" />
      </div>
    </div>
  </section>
  <section
    v-else-if="failed"
    class="storefront-faq-placeholder storefront-faq-placeholder--failed"
    :class="{ 'storefront-faq-placeholder--with-saved-answers': hasSavedAnswers }"
    role="status"
  >
    <p class="storefront-faq-placeholder__message">
      {{ t(hasSavedAnswers ? 'faq.ui.showingSavedAnswers' : 'faq.ui.loadFailed') }}
    </p>
    <button
      type="button"
      class="storefront-faq-placeholder__retry"
      :disabled="loading"
      :aria-busy="loading"
      @click="emit('retry')"
    >
      {{ t('common.retry') }}
    </button>
  </section>
  <section v-else-if="empty" class="storefront-faq-placeholder" role="status">
    <p class="storefront-faq-placeholder__message">{{ t('faq.ui.emptySection') }}</p>
  </section>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'

withDefaults(defineProps<{
  loading: boolean
  failed: boolean
  hasSavedAnswers?: boolean
  empty?: boolean
}>(), {
  hasSavedAnswers: false,
  empty: false,
})

const emit = defineEmits<{ retry: [] }>()
const { t } = useI18n()
</script>

<style scoped>
.storefront-faq-placeholder {
  box-sizing: border-box;
  width: 100%;
  min-height: 10rem;
  padding: 1.25rem;
  border: 1px solid rgba(20, 32, 43, 0.12);
  border-radius: 1.25rem;
  background: var(--tz-card-surface, #fff);
  color: var(--tz-text-secondary, #475569);
}

.storefront-faq-placeholder__message { margin: 0 0 1rem; line-height: 1.5; }
.storefront-faq-placeholder__layout, .storefront-faq-placeholder__questions { display: grid; gap: 0.75rem; }
.storefront-faq-placeholder__question, .storefront-faq-placeholder__answer {
  display: grid;
  align-content: start;
  gap: 0.7rem;
  padding: 1rem;
  border: 1px solid rgba(20, 32, 43, 0.1);
  border-radius: 1rem;
}
.storefront-faq-placeholder__answer { display: none; }
.storefront-faq-placeholder__line {
  display: block;
  height: 0.65rem;
  border-radius: 999px;
  background: rgba(20, 32, 43, 0.14);
  animation: storefront-faq-placeholder-pulse 1.4s ease-in-out infinite alternate;
}
.storefront-faq-placeholder__line--short { width: 55%; }
.storefront-faq-placeholder__retry {
  padding: 0.6rem 1rem;
  border: 1px solid currentColor;
  border-radius: 0.75rem;
  background: transparent;
  color: var(--tz-text-accent, #047857);
  font: inherit;
  cursor: pointer;
}
.storefront-faq-placeholder__retry:focus-visible { outline: 2px solid currentColor; outline-offset: 3px; }
.storefront-faq-placeholder--with-saved-answers { min-height: 0; margin-bottom: 1rem; }

@media (min-width: 768px) {
  .storefront-faq-placeholder--loading { min-height: 26rem; }
}
@media (min-width: 981px) {
  .storefront-faq-placeholder__layout { grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr); }
  .storefront-faq-placeholder__answer { display: grid; min-height: 20rem; }
}
@keyframes storefront-faq-placeholder-pulse { from { opacity: 0.45; } to { opacity: 1; } }
@media (prefers-reduced-motion: reduce) { .storefront-faq-placeholder__line { animation: none; } }
</style>
