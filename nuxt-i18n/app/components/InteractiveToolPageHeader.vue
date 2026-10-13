<template>
  <header class="interactive-tool-page-header">
    <div class="interactive-tool-page-header__heading-row">
      <h1 :id="headingId" class="interactive-tool-page-header__title">{{ title }}</h1>
    </div>
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
  headingId?: string
}>()
</script>

<style scoped>
.interactive-tool-page-header {
  display: grid;
  gap: 0.375rem;
  min-width: 0;
  width: 100%;
  margin: 0 0 1rem;
  border-bottom: 1px solid var(--tz-border-subtle);
  padding: 0 0 0.625rem;
  color: var(--tz-text-primary);
  text-align: center;
}

.interactive-tool-page-header__heading-row {
  display: flex;
  justify-content: center;
  align-items: baseline;
  gap: 0.75rem;
  min-width: 0;
}

.interactive-tool-page-header__title {
  min-width: 0;
  margin: 0;
  color: var(--tz-text-primary);
  font-size: 1.125rem;
  font-weight: 600;
  line-height: 1.4;
  overflow-wrap: anywhere;
  text-wrap: balance;
}

.interactive-tool-page-header__description {
  max-width: 82ch;
  margin: 0 auto;
  color: var(--tz-text-secondary);
  font-size: var(--tz-type-description);
  line-height: 1.55;
  overflow-wrap: anywhere;
}

.interactive-tool-page-header__details {
  min-width: 0;
}

@media (max-width: 640px) {
  .interactive-tool-page-header {
    margin-bottom: 0.75rem;
  }

  .interactive-tool-page-header__title {
    font-size: 1rem;
  }
}
</style>
