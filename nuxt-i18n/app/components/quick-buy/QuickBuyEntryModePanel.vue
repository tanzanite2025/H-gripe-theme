<script setup lang="ts">
import { computed } from 'vue'

const emit = defineEmits<{
  directSelect: []
  wheelsetSelectionAssistant: []
}>()

const { t } = useI18n()

const entryModes = computed(() => [
  {
    key: 'direct-select',
    enabled: true,
    icon: 'lucide:sliders-horizontal',
    title: t('quickBuy.entry.directSelect.title', 'I know my specs'),
    description: t(
      'quickBuy.entry.directSelect.description',
      'Choose configured items now.',
    ),
    action: t('quickBuy.entry.directSelect.action', 'Start'),
  },
  {
    key: 'ready-match',
    enabled: false,
    icon: 'lucide:circle-help',
    title: t('quickBuy.entry.readyMatch.title', 'Match ready wheelsets'),
    description: t(
      'quickBuy.entry.readyMatch.description',
      'Coming next.',
    ),
    action: '?',
  },
  {
    key: 'wheelset-selection-assistant',
    enabled: true,
    icon: 'lucide:circle-help',
    title: t('quickBuy.entry.wheelsetSelectionAssistant.title', 'Find my bike specs'),
    description: t(
      'quickBuy.entry.wheelsetSelectionAssistant.description',
      'Answer guided fit questions.',
    ),
    action: t('quickBuy.entry.wheelsetSelectionAssistant.action', 'Start'),
  },
])

const handleEntryModeClick = (key: string, enabled: boolean) => {
  if (!enabled) return
  if (key === 'direct-select') {
    emit('directSelect')
  }
  if (key === 'wheelset-selection-assistant') {
    emit('wheelsetSelectionAssistant')
  }
}
</script>

<template>
  <div class="quickbuy-entry-mode-panel">
    <div class="quickbuy-entry-mode-panel__header">
      <span class="quickbuy-entry-mode-panel__eyebrow">
        {{ t('quickBuy.entry.eyebrow', 'QUICKBUY') }}
      </span>
      <h2 class="quickbuy-entry-mode-panel__title">
        {{ t('quickBuy.entry.title', 'Choose how to start') }}
      </h2>
    </div>

    <div class="quickbuy-entry-mode-panel__grid">
      <button
        v-for="mode in entryModes"
        :key="mode.key"
        class="quickbuy-entry-mode-card"
        :class="{ 'quickbuy-entry-mode-card--disabled': !mode.enabled }"
        type="button"
        :disabled="!mode.enabled"
        :aria-disabled="!mode.enabled"
        @click="handleEntryModeClick(mode.key, mode.enabled)"
      >
        <span class="quickbuy-entry-mode-card__icon" aria-hidden="true">
          <Icon :name="mode.icon" class="h-5 w-5" />
        </span>
        <span class="quickbuy-entry-mode-card__content">
          <span class="quickbuy-entry-mode-card__title">
            {{ mode.title }}
          </span>
          <span class="quickbuy-entry-mode-card__description">
            {{ mode.description }}
          </span>
        </span>
        <span class="quickbuy-entry-mode-card__action">
          {{ mode.action }}
        </span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.quickbuy-entry-mode-panel {
  display: flex;
  width: 100%;
  min-height: 100%;
  box-sizing: border-box;
  flex-direction: column;
  justify-content: center;
  gap: 0.75rem;
  padding: 0.75rem;
}

.quickbuy-entry-mode-panel__header {
  display: grid;
  gap: 0.35rem;
  justify-items: center;
  text-align: center;
}

.quickbuy-entry-mode-panel__eyebrow {
  color: rgba(181, 255, 109, 0.86);
  font-size: 0.68rem;
  font-weight: 800;
  letter-spacing: 0.14em;
  text-transform: uppercase;
}

.quickbuy-entry-mode-panel__title {
  margin: 0;
  color: #ffffff;
  font-size: 1.08rem;
  font-weight: 850;
  letter-spacing: 0;
  line-height: 1.16;
}

.quickbuy-entry-mode-panel__grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.5rem;
  width: 100%;
}

.quickbuy-entry-mode-card {
  display: flex;
  min-width: 0;
  min-height: 4.75rem;
  align-items: center;
  gap: 0.65rem;
  box-sizing: border-box;
  padding: 0.75rem;
  border: 0;
  border-radius: 0.75rem;
  color: white;
  background:
    linear-gradient(180deg, var(--quickbuy-panel-surface-raised, #17171b), #101014);
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.03),
    inset 0 0 0 1px rgba(0, 0, 0, 0.56),
    0 12px 28px rgba(0, 0, 0, 0.26);
  text-align: left;
  transition: background 160ms ease, opacity 160ms ease, transform 160ms ease;
}

.quickbuy-entry-mode-card:hover:not(:disabled) {
  background:
    linear-gradient(180deg, #202026, var(--quickbuy-panel-surface-raised, #17171b));
  transform: translateY(-1px);
}

.quickbuy-entry-mode-card--disabled {
  cursor: not-allowed;
  opacity: 0.42;
}

.quickbuy-entry-mode-card__icon {
  display: grid;
  width: 2rem;
  height: 2rem;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 999px;
  color: #050505;
  background: #ffffff;
  box-shadow: 0 8px 18px rgba(0, 0, 0, 0.28);
}

.quickbuy-entry-mode-card--disabled .quickbuy-entry-mode-card__icon {
  color: rgba(255, 255, 255, 0.74);
  background: var(--quickbuy-control-surface-raised, #151519);
  box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.68);
}

.quickbuy-entry-mode-card__content {
  display: grid;
  min-width: 0;
  gap: 0.45rem;
}

.quickbuy-entry-mode-card__title {
  color: #ffffff;
  font-size: 0.88rem;
  font-weight: 800;
  letter-spacing: 0;
  line-height: 1.25;
}

.quickbuy-entry-mode-card__description {
  color: rgba(255, 255, 255, 0.68);
  font-size: 0.75rem;
  line-height: 1.35;
}

.quickbuy-entry-mode-card__action {
  display: inline-flex;
  min-height: 2rem;
  flex: 0 0 auto;
  max-width: 100%;
  align-items: center;
  justify-content: center;
  overflow-wrap: anywhere;
  padding: 0.42rem 0.72rem;
  border-radius: 999px;
  color: #050505;
  background: #ffffff;
  font-size: 0.75rem;
  font-weight: 800;
  line-height: 1.2;
  text-align: center;
}

.quickbuy-entry-mode-card--disabled .quickbuy-entry-mode-card__action {
  width: 2rem;
  min-width: 2rem;
  padding: 0;
  color: rgba(255, 255, 255, 0.78);
  background: var(--quickbuy-control-surface-raised, #151519);
}

@media (max-width: 767px) {
  .quickbuy-entry-mode-panel {
    padding: 0.65rem;
  }

  .quickbuy-entry-mode-card {
    min-height: 4.5rem;
  }
}
</style>
