<template>
  <header class="interactive-tool-page-header">
    <p v-if="eyebrow" class="interactive-tool-page-header__eyebrow">{{ eyebrow }}</p>
    <h1 :id="headingId" class="interactive-tool-page-header__title">{{ title }}</h1>
    <p v-if="description" class="interactive-tool-page-header__description">{{ description }}</p>
    <div v-if="$slots.default" class="interactive-tool-page-header__details">
      <slot />
    </div>
  </header>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  eyebrow?: string
  headingId?: string
}>()
</script>

<style scoped>
.interactive-tool-page-header {
  position: relative;
  display: grid;
  gap: 0.75rem;
  min-width: 0;
  width: 100%;
  margin: 0 0 1.5rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 1rem;
  background: linear-gradient(120deg, var(--tz-surface-subtle), var(--tz-card-surface) 65%);
  padding: clamp(1.25rem, 2vw, 1.75rem);
  color: var(--tz-text-primary);
  text-align: start;
}

.interactive-tool-page-header::before {
  content: '';
  width: 2.5rem;
  height: 3px;
  border-radius: 999px;
  background: var(--tz-site-accent);
}

.interactive-tool-page-header__eyebrow {
  margin: 0;
  color: var(--tz-text-accent);
  font-size: var(--tz-type-micro-label);
  font-weight: 700;
  letter-spacing: 0.08em;
  line-height: 1.5;
}

.interactive-tool-page-header__title {
  min-width: 0;
  margin: 0;
  color: var(--tz-text-primary);
  font-size: var(--tz-type-page-title);
  font-weight: 700;
  letter-spacing: -0.015em;
  line-height: 1.3;
  overflow-wrap: anywhere;
  text-wrap: balance;
}

.interactive-tool-page-header__description {
  max-width: 82ch;
  margin: 0;
  color: var(--tz-text-secondary);
  font-size: 0.9375rem;
  line-height: 1.65;
  overflow-wrap: anywhere;
}

.interactive-tool-page-header__details {
  min-width: 0;
}

@media (max-width: 640px) {
  .interactive-tool-page-header {
    margin-bottom: 1.25rem;
  }
}
</style>
