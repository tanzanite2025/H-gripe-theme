<template>
  <div class="wheelset-selection-outcome-panel">
    <div class="wheelset-selection-outcome-panel__copy">
      <span class="wheelset-selection-outcome-panel__eyebrow">
        {{ t('wheelsetSelectionAssistant.outcome.eyebrow') }}
      </span>
      <h3>{{ prompt }}</h3>
      <p v-if="helper">{{ helper }}</p>
    </div>

    <div class="wheelset-selection-outcome-panel__actions">
      <button
        v-if="canGoBack"
        type="button"
        class="wheelset-selection-outcome-panel__back"
        @click="emit('back')"
      >
        <Icon name="lucide:arrow-left" class="h-4 w-4" />
        <span>{{ t('wheelsetSelectionAssistant.outcome.back') }}</span>
      </button>
      <button
        type="button"
        class="wheelset-selection-outcome-panel__contact"
        @click="emit('contactSupport')"
      >
        <Icon name="lucide:message-circle" class="h-4 w-4" />
        <span>{{ t('wheelsetSelectionAssistant.outcome.contact') }}</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'

defineProps<{
  prompt: string
  helper?: string
  canGoBack: boolean
}>()

const { t } = useI18n()

const emit = defineEmits<{
  back: []
  contactSupport: []
}>()
</script>

<style scoped>
.wheelset-selection-outcome-panel {
  display: flex;
  min-height: 100%;
  flex-direction: column;
  justify-content: space-between;
  gap: 1.5rem;
  padding: 1rem;
}

.wheelset-selection-outcome-panel__copy {
  display: grid;
  gap: 0.55rem;
  align-content: start;
}

.wheelset-selection-outcome-panel__eyebrow {
  color: var(--tz-site-accent, #059669);
  font-size: 0.66rem;
  font-weight: 850;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.wheelset-selection-outcome-panel__copy h3 {
  color: var(--tz-text-primary, #0f172a);
  font-size: 1.05rem;
  font-weight: 800;
  line-height: 1.4;
}

.wheelset-selection-outcome-panel__copy p {
  color: var(--tz-text-muted, #94a3b8);
  font-size: 0.85rem;
  line-height: 1.55;
}

.wheelset-selection-outcome-panel__actions {
  display: grid;
  gap: 0.65rem;
}

.wheelset-selection-outcome-panel__back,
.wheelset-selection-outcome-panel__contact {
  display: inline-flex;
  min-height: 2.5rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border-radius: 0.7rem;
  padding: 0 0.9rem;
  font-size: 0.82rem;
  font-weight: 800;
}

.wheelset-selection-outcome-panel__back {
  color: var(--tz-text-secondary);
  background: var(--quickbuy-control-surface-raised, var(--tz-surface-subtle));
}

.wheelset-selection-outcome-panel__contact {
  color: #ffffff;
  background: var(--tz-site-accent, #059669);
}
</style>
