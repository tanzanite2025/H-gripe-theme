<template>
  <div
    v-show="!isWheelsetSelectionAssistantOpen"
    class="dock-shell fixed inset-x-0 bottom-0 w-full z-[101]"
  >
    <button
      class="dock-shell__toggle"
      type="button"
      :aria-label="$t('dockMenu.menu')"
      aria-expanded="false"
      @click="emit('expand')"
    >
      <svg class="dock-shell__icon" viewBox="0 0 24 24" fill="none" aria-hidden="true">
        <path
          d="m6 14 6-6 6 6M6 20l6-6 6 6"
          transform="translate(0 -2)"
          stroke="currentColor"
          stroke-width="2.25"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from '#imports'
import { useWheelsetSelectionAssistantModalState } from '~/composables/useWheelsetSelectionAssistantModalState'

const { isOpen: isWheelsetSelectionAssistantOpen } = useWheelsetSelectionAssistantModalState()
const { t: $t } = useI18n()

const emit = defineEmits<{
  expand: []
}>()
</script>

<style scoped>
.dock-shell {
  display: flex;
  justify-content: center;
  padding: 0 1rem max(0.75rem, calc(0.75rem + var(--tz-safe-area-bottom, 0px)));
  pointer-events: none;
}

.dock-shell__toggle {
  position: relative;
  display: flex;
  box-sizing: border-box;
  width: 10rem;
  height: 2.5rem;
  align-items: center;
  justify-content: center;
  padding: 0;
  border: 1px solid var(--tz-border-strong);
  border-radius: 999px;
  background: var(--tz-input-surface);
  color: var(--tz-site-accent, #059669);
  cursor: pointer;
  line-height: 0;
  pointer-events: auto;
  box-shadow:
    0 8px 20px rgba(20, 32, 43, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
}

.dock-shell__icon {
  display: block;
  position: absolute;
  top: 50%;
  left: 50%;
  flex: 0 0 auto;
  width: 1.75rem;
  height: 1.75rem;
  margin: 0;
  transform: translate(-50%, -50%);
}

.dock-shell__toggle:hover {
  border-color: var(--tz-site-accent, #059669);
  background: var(--tz-site-accent-soft-surface);
}
</style>
