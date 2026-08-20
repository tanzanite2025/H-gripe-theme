<script setup lang="ts">
import { computed } from 'vue'
import { clearError, useHead } from '#imports'

const props = defineProps<{
  error: {
    statusCode?: number
    statusMessage?: string
    message?: string
  }
}>()

const statusCode = computed(() => Number(props.error?.statusCode || 500))
const title = computed(() => (statusCode.value === 404 ? 'Page not found' : 'Storefront unavailable'))
const description = computed(() => (
  statusCode.value === 404
    ? 'The page may have moved or the address may be incomplete.'
    : 'The storefront hit a temporary error while rendering this page.'
))

useHead({
  link: [
    { rel: 'stylesheet', href: '/fonts/maple-ui.css' },
  ],
})

const returnHome = (): Promise<void> => clearError({ redirect: '/' })
</script>

<template>
  <main class="storefront-error">
    <section class="storefront-error__content" aria-labelledby="storefront-error-title">
      <p class="storefront-error__code">{{ statusCode }}</p>
      <h1 id="storefront-error-title" class="storefront-error__title">{{ title }}</h1>
      <p class="storefront-error__description">{{ description }}</p>
      <button class="storefront-error__button" type="button" @click="returnHome">
        Back to storefront
      </button>
    </section>
  </main>
</template>

<style scoped>
.storefront-error {
  display: grid;
  min-height: 100vh;
  place-items: center;
  background: #050507;
  color: #f8fafc;
  font-family: 'MapleUILatin', 'MapleUICJK';
  padding: 1.5rem;
}

.storefront-error__content {
  width: min(100%, 32rem);
  text-align: center;
}

.storefront-error__code {
  color: #b5ff6d;
  font-size: clamp(4rem, 12vw, 7rem);
  font-weight: 900;
  line-height: 0.95;
}

.storefront-error__title {
  color: #ffffff;
  font-size: clamp(1.4rem, 3vw, 2.1rem);
  font-weight: 900;
  line-height: 1.1;
  margin-top: 1rem;
}

.storefront-error__description {
  color: #cbd5e1;
  font-size: 0.95rem;
  line-height: 1.6;
  margin: 0.9rem auto 0;
  max-width: 27rem;
}

.storefront-error__button {
  align-items: center;
  background: #b5ff6d;
  border: 0;
  color: #050507;
  cursor: pointer;
  display: inline-flex;
  font: inherit;
  font-size: 0.9rem;
  font-weight: 900;
  justify-content: center;
  margin-top: 1.5rem;
  min-height: 2.75rem;
  padding: 0 1.25rem;
}

.storefront-error__button:focus-visible {
  outline: 2px solid #ffffff;
  outline-offset: 3px;
}
</style>
