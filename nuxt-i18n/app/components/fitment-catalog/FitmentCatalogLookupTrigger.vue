<template>
  <button
    type="button"
    class="fitment-catalog-lookup-trigger"
    :aria-label="label || t('fitmentCatalog.trigger.label')"
    :title="title || label || t('fitmentCatalog.trigger.label')"
    @click="emit('open')"
  >
    <Icon name="lucide:circle-help" class="fitment-catalog-lookup-trigger__icon" aria-hidden="true" />
  </button>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'

withDefaults(defineProps<{
  label?: string
  title?: string
}>(), {
  label: '',
  title: '',
})

const emit = defineEmits<{
  open: []
}>()

const { t } = useI18n()
</script>

<style scoped>
.fitment-catalog-lookup-trigger {
  display: inline-grid;
  width: 2rem;
  height: 2rem;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 999px;
  color: var(--tz-action-primary);
  background: var(--tz-surface-subtle);
  cursor: pointer;
  transition: border-color 160ms ease, color 160ms ease, background-color 160ms ease;
}

.fitment-catalog-lookup-trigger:hover {
  border-color: var(--tz-action-primary);
  color: var(--tz-action-primary-hover);
  background: var(--tz-surface-muted);
}

.fitment-catalog-lookup-trigger:focus-visible {
  outline: 2px solid var(--tz-action-primary);
  outline-offset: 3px;
}

.fitment-catalog-lookup-trigger__icon {
  width: 1.05rem;
  height: 1.05rem;
}
</style>
