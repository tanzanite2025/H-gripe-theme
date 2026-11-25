<template>
  <div
    v-show="!isWheelsetSelectionAssistantOpen"
    class="dock-shell fixed inset-x-0 bottom-0 w-full z-[101]"
    :class="{ 'dock-shell--expanded': expanded }"
  >
    <button
      class="dock-shell__toggle"
      type="button"
      :aria-label="expanded ? `${$t('common.close')} ${$t('dockMenu.menu')}` : $t('dockMenu.menu')"
      :aria-expanded="expanded"
      :aria-controls="expanded ? 'storefront-dock-actions' : undefined"
      @click="emit('expand')"
    >
      <svg
        class="dock-shell__expanded-shoulder-contour"
        viewBox="0 0 128 33"
        preserveAspectRatio="none"
        aria-hidden="true"
      >
        <path
          d="M0 32.5C8.84 32.5 16 25.34 16 16.5C16 7.66 23.16 0.5 32 0.5H96C104.84 0.5 112 7.66 112 16.5C112 25.34 119.16 32.5 128 32.5"
          vector-effect="non-scaling-stroke"
        />
      </svg>
      <svg
        class="dock-shell__icon"
        :class="{ 'dock-shell__icon--expanded': expanded }"
        viewBox="0 0 24 24"
        fill="none"
        aria-hidden="true"
      >
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

withDefaults(defineProps<{ expanded?: boolean }>(), { expanded: false })

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
  transition: transform 160ms ease;
}

.dock-shell__expanded-shoulder-contour {
  display: none;
}

.dock-shell__toggle:hover {
  border-color: var(--tz-site-accent, #059669);
  background: var(--tz-site-accent-soft-surface);
}

.dock-shell__toggle:focus-visible {
  outline: 3px solid var(--tz-site-accent, #059669);
  outline-offset: 4px;
}

.dock-shell__icon--expanded {
  transform: translate(-50%, -50%) rotate(180deg);
}

@media (max-width: 767.98px) {
  .dock-shell {
    bottom: calc(0.5rem + var(--tz-safe-area-bottom, 0px));
    padding: 0;
  }

  .dock-shell__toggle {
    width: 6rem;
    height: 2rem;
  }

  .dock-shell--expanded .dock-shell__toggle {
    border: 0;
    background: transparent;
    box-shadow: none;
  }

  .dock-shell--expanded .dock-shell__toggle:hover {
    background: transparent;
  }

  .dock-shell--expanded .dock-shell__expanded-shoulder-contour {
    position: absolute;
    top: 0;
    left: -1rem;
    display: block;
    width: calc(100% + 2rem);
    height: calc(100% + 1px);
    overflow: visible;
    fill: var(--tz-mobile-bottom-chrome-surface);
    stroke: var(--tz-border-subtle);
    stroke-width: 1px;
    pointer-events: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .dock-shell__icon {
    transition: none;
  }
}
</style>
