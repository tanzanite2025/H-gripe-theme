<template>
  <div
    class="storefront-data-notice"
    :class="[
      `storefront-data-notice--${tone}`,
      { 'storefront-data-notice--compact': compact },
    ]"
    :role="role"
    :aria-live="role === 'status' ? 'polite' : undefined"
  >
    <span class="storefront-data-notice__icon" aria-hidden="true">
      <Icon :name="iconName" class="h-4 w-4" />
    </span>

    <div class="storefront-data-notice__copy">
      <p class="storefront-data-notice__title">{{ title }}</p>
      <p v-if="description" class="storefront-data-notice__description">{{ description }}</p>
    </div>

    <div v-if="$slots.actions" class="storefront-data-notice__actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type StorefrontDataNoticeTone = 'fallback' | 'error' | 'empty'

const props = withDefaults(defineProps<{
  tone?: StorefrontDataNoticeTone
  title: string
  description?: string
  compact?: boolean
  role?: 'status' | 'alert'
}>(), {
  tone: 'fallback',
  description: '',
  compact: false,
  role: 'status',
})

const iconName = computed(() => {
  if (props.tone === 'error') return 'lucide:cloud-off'
  if (props.tone === 'empty') return 'lucide:circle-dashed'
  return 'lucide:info'
})
</script>

<style scoped>
.storefront-data-notice {
  display: flex;
  min-width: 0;
  align-items: flex-start;
  gap: 0.7rem;
  padding: 0.75rem 0.85rem;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 0.75rem;
  background: var(--tz-card-surface);
  color: var(--tz-text-secondary);
}

.storefront-data-notice--compact {
  padding: 0.55rem 0.7rem;
}

.storefront-data-notice--fallback {
  border-color: var(--tz-site-accent);
  background: var(--tz-site-accent-soft-surface);
}

.storefront-data-notice--error {
  border-color: var(--tz-status-danger-text);
  background: var(--tz-status-danger-bg);
}

.storefront-data-notice--empty {
  border-style: dashed;
  border-color: rgba(148, 163, 184, 0.26);
}

.storefront-data-notice__icon {
  display: grid;
  width: 1.65rem;
  height: 1.65rem;
  flex: 0 0 auto;
  place-items: center;
  border: 1px solid var(--tz-border-subtle);
  border-radius: 999px;
  color: var(--tz-text-secondary);
  background: var(--tz-surface-muted);
}

.storefront-data-notice--fallback .storefront-data-notice__icon {
  border-color: rgba(5, 150, 105, 0.32);
  color: var(--tz-site-accent);
}

.storefront-data-notice--error .storefront-data-notice__icon {
  border-color: var(--tz-status-danger-text);
  color: var(--tz-status-danger-text);
}

.storefront-data-notice__copy {
  min-width: 0;
  flex: 1 1 auto;
}

.storefront-data-notice__title,
.storefront-data-notice__description {
  margin: 0;
}

.storefront-data-notice__title {
  color: var(--tz-text-primary);
  font-size: 0.78rem;
  font-weight: 750;
  line-height: 1.3;
}

.storefront-data-notice__description {
  margin-top: 0.2rem;
  color: var(--tz-text-muted);
  font-size: 0.72rem;
  line-height: 1.45;
}

.storefront-data-notice__actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.4rem;
}

.storefront-data-notice :deep(.storefront-data-notice-action) {
  display: inline-flex;
  min-height: 1.85rem;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  padding: 0.35rem 0.7rem;
  border: 1px solid var(--tz-border-strong);
  border-radius: 999px;
  background: var(--tz-surface-muted);
  color: var(--tz-text-primary);
  font-size: 0.72rem;
  font-weight: 750;
  line-height: 1.2;
  text-decoration: none;
  white-space: nowrap;
  cursor: pointer;
  transition:
    background-color 160ms ease,
    border-color 160ms ease,
    color 160ms ease;
}

.storefront-data-notice :deep(.storefront-data-notice-action:hover:not(:disabled)) {
  border-color: var(--tz-site-accent-hover);
  background: var(--tz-site-accent-soft-surface);
  color: var(--tz-site-accent);
}

.storefront-data-notice :deep(.storefront-data-notice-action:disabled) {
  cursor: wait;
  opacity: 0.58;
}

.storefront-data-notice :deep(.storefront-data-notice-action svg) {
  width: 0.85rem;
  height: 0.85rem;
}

@media (max-width: 480px) {
  .storefront-data-notice {
    flex-wrap: wrap;
  }

  .storefront-data-notice__actions {
    width: 100%;
    padding-left: 2.35rem;
  }
}
</style>
