<template>
  <aside class="wheelset-selection-support-cta">
    <div class="wheelset-selection-support-cta__copy">
      <strong>{{ title }}</strong>
      <p>{{ description }}</p>
    </div>

    <button
      type="button"
      class="wheelset-selection-support-cta__button"
      @click="emit('contactSupport')"
    >
      <Icon name="lucide:message-circle" class="h-4 w-4" />
      <span>{{ actionLabel }}</span>
    </button>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from '#imports'

const props = withDefaults(defineProps<{
  title?: string
  description?: string
  actionLabel?: string
}>(), {
  title: '',
  description: '',
  actionLabel: '',
})

const { t } = useI18n()
const title = computed(() => props.title || t(
  'wheelsetSelectionAssistant.support.title',
  'Contact support directly',
))
const description = computed(() => props.description || t(
  'wheelsetSelectionAssistant.support.description',
  'Contact support with your bike model or requirements.',
))
const actionLabel = computed(() => props.actionLabel || t(
  'wheelsetSelectionAssistant.support.action',
  'Contact support',
))

const emit = defineEmits<{
  contactSupport: []
}>()
</script>

<style scoped>
.wheelset-selection-support-cta {
  display: flex;
  width: 100%;
  box-sizing: border-box;
  flex: 0 0 auto;
  align-items: center;
  gap: 0.75rem;
  margin-top: 0;
  border-radius: 0.75rem;
  background: var(--quickbuy-panel-surface-soft, #16171d);
  padding: 0.875rem;
}

.wheelset-selection-support-cta__copy {
  display: grid;
  min-width: 0;
  flex: 1 1 auto;
  gap: 0.25rem;
}

.wheelset-selection-support-cta__copy strong {
  color: var(--tz-text-primary, #f8fafc);
  font-size: 0.875rem;
  font-weight: 800;
}

.wheelset-selection-support-cta__copy p {
  color: var(--tz-text-muted, #94a3b8);
  font-size: 0.78rem;
  line-height: 1.45;
}

.wheelset-selection-support-cta__button {
  display: inline-flex;
  flex: 0 0 auto;
  min-height: 2.5rem;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  border-radius: 0.75rem;
  background: #ffffff;
  padding: 0 0.875rem;
  color: #101014;
  font-size: 0.875rem;
  font-weight: 800;
  white-space: nowrap;
  transition: background 160ms ease;
}

.wheelset-selection-support-cta__button:hover {
  background: #f1f5f9;
}

@media (max-width: 767px) {
  .wheelset-selection-support-cta__copy p {
    display: none;
  }
}

</style>
