<template>
  <button
    class="chat-start-button"
    :type="type || 'button'"
    :disabled="disabled"
    @click="onClick"
  >
    <span class="chat-start-button__label">
      <slot>{{ label }}</slot>
    </span>
    <span class="chat-start-button__overlay">
      <svg
        class="chat-start-button__icon"
        viewBox="0 0 24 24"
        aria-hidden="true"
      >
        <path d="M0 0h24v24H0z" fill="none" />
        <path
          d="M16.172 11 10.808 5.636l1.414-1.414L20 12l-7.778 7.778-1.414-1.414L16.172 13H4v-2z"
          fill="currentColor"
        />
      </svg>
    </span>
  </button>
</template>

<script setup lang="ts">
const props = defineProps<{
  label?: string
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'click', event: MouseEvent): void
}>()

const onClick = (event: MouseEvent) => {
  if (props.disabled) return
  emit('click', event)
}
</script>

<style scoped>
.chat-start-button {
  position: relative;
  width: 100%;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 0.625rem 3.1rem 0.625rem 1.2rem;
  border-radius: 9999px;
  border: 1px solid rgba(181, 255, 109, 0.28);
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 600;
  letter-spacing: 0.02em;
  transition:
    background 0.25s ease,
    box-shadow 0.25s ease,
    transform 0.2s ease,
    opacity 0.2s ease;
  overflow: hidden;
}

.chat-start-button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.chat-start-button__label {
  position: relative;
  white-space: nowrap;
}

.chat-start-button__overlay {
  position: absolute;
  right: 0.3rem;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 2.1rem;
  width: 2.1rem;
  border-radius: 0.9rem;
}

.chat-start-button__icon {
  width: 1.1rem;
  height: 1.1rem;
  transition: transform 0.3s ease;
}

/* Desktop / hover-capable devices: dark glass base + overlay slide animation */
@media (hover: hover) and (pointer: fine) {
  .chat-start-button {
    color: #e5e7eb;
    background:
      linear-gradient(135deg, rgba(24, 24, 24, 0.98), rgba(5, 5, 5, 0.98));
    box-shadow:
      0 10px 30px -18px rgba(0, 0, 0, 1),
      inset 0 0 0 1px rgba(255, 255, 255, 0.06),
      0 0 18px rgba(181, 255, 109, 0.12);
  }

  .chat-start-button__overlay {
    background: #b5ff6d;
    box-shadow: 0 4px 12px rgba(181, 255, 109, 0.22);
    z-index: 1;
    transition:
      width 0.3s ease,
      box-shadow 0.3s ease,
      transform 0.2s ease;
  }

  .chat-start-button__icon {
    color: #020617;
  }

  .chat-start-button:hover {
    box-shadow:
      0 5px 16px -10px rgba(0, 0, 0, 0.95),
      inset 0 0 0 1px rgba(181, 255, 109, 0.18),
      0 0 14px rgba(181, 255, 109, 0.16);
  }

  .chat-start-button:hover .chat-start-button__overlay {
    width: calc(100% - 0.6rem);
    box-shadow:
      0 6px 18px -10px rgba(181, 255, 109, 0.48),
      0 0 18px rgba(181, 255, 109, 0.18);
  }

  .chat-start-button:hover .chat-start-button__icon {
    transform: translateX(0.1rem);
  }

  .chat-start-button:active .chat-start-button__overlay {
    transform: scale(0.96);
  }
}

/* Mobile / touch devices: static gradient CTA without overlay animation */
@media (hover: none) and (pointer: coarse) {
  .chat-start-button {
    padding: 0.625rem 1rem;
    color: #e5e7eb;
    background: linear-gradient(135deg, rgba(24, 24, 24, 0.98), rgba(5, 5, 5, 0.98));
    box-shadow:
      0 10px 30px -18px rgba(0, 0, 0, 1),
      inset 0 0 0 1px rgba(255, 255, 255, 0.06),
      0 0 18px rgba(181, 255, 109, 0.12);
  }

  .chat-start-button__label {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.45rem;
  }

  .chat-start-button__overlay {
    position: static;
    width: auto;
    height: auto;
    background: transparent;
    box-shadow: none;
  }

  .chat-start-button__icon {
    color: #b5ff6d;
  }

  .chat-start-button:active {
    transform: scale(0.98);
    box-shadow: 0 6px 18px rgba(181, 255, 109, 0.24);
  }
}
</style>
