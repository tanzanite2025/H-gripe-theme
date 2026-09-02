<template>
  <div
    v-show="!isWheelsetSelectionAssistantOpen"
    class="dock-shell fixed inset-x-0 bottom-0 w-full z-[101]"
  >
    <div class="dock-shell__surface mx-auto w-full md:max-w-[500px] rounded-none px-1 py-2.5 md:px-4 md:py-3">
      <button
        class="dock-shell__icon-button"
        type="button"
        aria-label="Open sidebar"
        @click="emit('intent', 'sidebar')"
      >
        <svg class="dock-shell__icon" viewBox="0 0 24 24" fill="none">
          <path d="M18 20a6 6 0 0 0-12 0" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
          <circle cx="12" cy="10" r="4" stroke="currentColor" stroke-width="2" />
          <path d="m17 8 2 2 4-4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>

      <button
        class="dock-shell__icon-button"
        type="button"
        aria-label="Open chat"
        @click="emit('intent', 'chat')"
      >
        <svg class="dock-shell__icon" viewBox="0 0 48 48" fill="none">
          <g transform="translate(24 24) scale(1.2) translate(-24 -24)">
            <path
              d="M31 12H15.5C12.46 12 10 14.46 10 17.5V30C10 33.04 12.46 35.5 15.5 35.5H20V42L28 35.5H37C40.04 35.5 42.5 33.04 42.5 30V21"
              stroke="currentColor"
              stroke-width="3.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
            <path
              d="M29.5 18.5L40 8M33 8H40V15"
              stroke="currentColor"
              stroke-width="3.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </g>
        </svg>
      </button>

      <button
        class="dock-shell__icon-button"
        type="button"
        aria-label="Open quick buy"
        @click="emit('intent', 'quick-buy')"
      >
        <svg class="dock-shell__icon" viewBox="0 0 24 24" fill="none">
          <path d="m13 2-9 11h7l-1 9 9-11h-7l1-9Z" fill="currentColor" />
        </svg>
      </button>

      <button
        class="dock-shell__cart"
        type="button"
        aria-label="Open cart"
        @click="emit('intent', 'cart')"
      >
        <span class="dock-shell__currency">$</span>
        <span class="dock-shell__total">0.00</span>
        <span class="dock-shell__count">0</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useWheelsetSelectionAssistantModalState } from '~/composables/useWheelsetSelectionAssistantModalState'

type GradientDockMenuShellIntent = 'sidebar' | 'chat' | 'quick-buy' | 'cart'

const { isOpen: isWheelsetSelectionAssistantOpen } = useWheelsetSelectionAssistantModalState()

const emit = defineEmits<{
  intent: [value: GradientDockMenuShellIntent]
}>()
</script>

<style scoped>
.dock-shell {
  min-height: var(--tz-bottom-dock-height, 4.5rem);
  background: var(--tz-input-surface);
  border-top: 1px solid var(--tz-border-subtle);
  box-shadow:
    0 -12px 32px rgba(20, 32, 43, 0.12),
    inset 0 1px 0 rgba(255, 255, 255, 0.9);
  color: var(--tz-text-secondary);
}

.dock-shell__surface {
  display: grid;
  max-width: min(100%, 500px);
  grid-template-columns: repeat(3, minmax(3rem, 1fr)) minmax(9.25rem, 1.85fr);
  column-gap: 0.25rem;
  align-items: center;
}

.dock-shell__icon-button {
  display: flex;
  min-width: 0;
  height: 2.75rem;
  align-items: center;
  justify-content: center;
  border: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  padding: 0;
}

.dock-shell__icon {
  display: block;
  width: 2rem;
  height: 2rem;
}

.dock-shell__cart {
  display: flex;
  width: 100%;
  min-width: 0;
  height: 2.5rem;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  padding-inline: 0.6rem;
  border: 1px solid rgba(255, 255, 255, 0.82);
  border-radius: 999px;
  background: #ffffff;
  color: var(--tz-text-primary);
  font-weight: 900;
  line-height: 1;
  white-space: nowrap;
  box-shadow:
    0 10px 24px rgba(0, 0, 0, 0.18),
    inset 0 1px 0 rgba(255, 255, 255, 0.82);
  cursor: pointer;
}

.dock-shell__currency {
  display: inline-grid;
  width: 1.25rem;
  height: 1.25rem;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 999px;
  color: var(--tz-text-primary);
}

.dock-shell__total {
  min-width: 0;
  max-width: 5.75rem;
  overflow: hidden;
  font-size: 1.05rem;
  letter-spacing: 0;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dock-shell__count {
  display: inline-flex;
  min-width: 1.875rem;
  height: 1.875rem;
  flex: 0 0 auto;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  background: var(--tz-site-accent, #059669);
  color: var(--tz-text-primary);
  font-size: 1rem;
  font-weight: 900;
  padding: 0 0.35rem;
}

@media (max-width: 767px) {
  .dock-shell {
    background: var(--tz-mobile-bottom-chrome-surface);
  }

  .dock-shell__surface {
    grid-template-columns: repeat(3, minmax(2.75rem, 1fr)) minmax(9.15rem, 1.85fr);
    padding-bottom: max(0.625rem, calc(0.625rem + var(--tz-safe-area-bottom, 0px)));
  }

  .dock-shell__icon-button {
    height: 2.75rem;
  }

  .dock-shell__icon {
    width: 1.75rem;
    height: 1.75rem;
  }

  .dock-shell__cart {
    padding-inline: 0.4rem;
  }
}
</style>
