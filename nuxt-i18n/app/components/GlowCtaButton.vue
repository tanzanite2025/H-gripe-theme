<template>
  <button
    class="cta-glow-button"
    :type="type || 'button'"
    @click="onClick"
  >
    <span class="cta-glow-button__label">
      <slot>{{ label }}</slot>
    </span>
    <span class="cta-glow-button__icon">
      <svg
        viewBox="0 0 24 24"
        aria-hidden="true"
        class="cta-glow-button__icon-svg"
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
}>()

const emit = defineEmits<{
  (e: 'click', event: MouseEvent): void
}>()

const onClick = (event: MouseEvent) => {
  emit('click', event)
}
</script>

<style scoped>
.cta-glow-button {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.4rem;
  padding: 0.4rem 3.1rem 0.4rem 1.2rem;
  border-radius: 9999px;
  border: none;
  cursor: pointer;
  font-size: 0.95rem;
  font-weight: 600;
  letter-spacing: 0.04em;
  text-transform: none;
  color: #e5e7eb;
  background:
    radial-gradient(circle at top left, rgba(31, 41, 55, 0.96), rgba(15, 23, 42, 0.98));
  box-shadow:
    0 4px 14px -10px rgba(0, 0, 0, 0.9),
    0 0 12px rgba(15, 23, 42, 0.85);
  overflow: hidden;
  height: 2.7rem;
  transition:
    background 0.25s ease,
    box-shadow 0.25s ease,
    transform 0.2s ease;
}

.cta-glow-button__label {
  position: relative;
  white-space: nowrap;
}

.cta-glow-button__icon {
  position: absolute;
  right: 0.3rem;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 2.1rem;
  width: 2.1rem;
  border-radius: 0.9rem;
  background: linear-gradient(135deg, #2dd4bf 0%, #3b82f6 100%);
  box-shadow: 0 4px 12px rgba(45, 212, 191, 0.35);
  z-index: 1;
  transition:
    width 0.3s ease,
    box-shadow 0.3s ease,
    transform 0.2s ease;
}

.cta-glow-button__icon-svg {
  width: 1.1rem;
  height: 1.1rem;
  color: #020617;
  transition: transform 0.3s ease;
}

.cta-glow-button:hover {
  box-shadow:
    0 5px 16px -10px rgba(0, 0, 0, 0.95),
    0 0 14px rgba(15, 23, 42, 0.9);
}

.cta-glow-button:hover .cta-glow-button__icon {
  width: calc(100% - 0.6rem);
  box-shadow:
    0 6px 18px -10px rgba(45, 212, 191, 0.45),
    0 0 18px rgba(15, 23, 42, 0.9);
}

.cta-glow-button:hover .cta-glow-button__icon-svg {
  transform: translateX(0.1rem);
}

.cta-glow-button:active .cta-glow-button__icon {
  transform: scale(0.96);
}
</style>
