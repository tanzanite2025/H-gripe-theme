<template>
  <div ref="rootRef" class="spoke-calculator-select">
    <button
      :id="id"
      ref="buttonRef"
      type="button"
      class="spoke-calculator-select__button"
      :class="{ 'spoke-calculator-select__button--placeholder': !selectedOption }"
      :disabled="disabled"
      :aria-label="ariaLabel"
      :aria-expanded="open"
      :aria-controls="menuId"
      aria-haspopup="listbox"
      @click="toggleMenu"
      @keydown="onButtonKeydown"
    >
      <span class="spoke-calculator-select__label">{{ selectedOption?.label || placeholder }}</span>
      <span class="spoke-calculator-select__chevron" aria-hidden="true">▾</span>
    </button>

    <div
      v-if="open"
      :id="menuId"
      class="spoke-calculator-select__menu"
      :class="{ 'spoke-calculator-select__menu--viewport-bound': keepMenuWithinViewport }"
      :style="menuStyle"
      role="listbox"
      :aria-labelledby="id"
    >
      <button
        v-for="(option, index) in options"
        :key="`${option.value}`"
        type="button"
        class="spoke-calculator-select__option"
        :class="{
          'spoke-calculator-select__option--active': index === highlightedIndex,
          'spoke-calculator-select__option--selected': option.value === modelValue,
        }"
        role="option"
        :aria-selected="option.value === modelValue"
        @mouseenter="highlightedIndex = index"
        @click="selectOption(option.value)"
      >
        {{ option.label }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

type SelectValue = string | number | null

interface SelectOption {
  label: string
  value: SelectValue
}

const props = withDefaults(defineProps<{
  id: string
  modelValue: SelectValue
  options: SelectOption[]
  placeholder?: string
  disabled?: boolean
  ariaLabel?: string
  keepMenuWithinViewport?: boolean
}>(), {
  placeholder: 'Select',
  disabled: false,
  ariaLabel: undefined,
  keepMenuWithinViewport: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: SelectValue]
}>()

const open = ref(false)
const rootRef = ref<HTMLElement | null>(null)
const buttonRef = ref<HTMLButtonElement | null>(null)
const highlightedIndex = ref(0)
const menuId = computed(() => `${props.id}-menu`)
const selectedOption = computed(() => props.options.find(option => option.value === props.modelValue) || null)
const menuStyle = ref<Record<string, string>>({})

const updateMenuPosition = () => {
  if (!props.keepMenuWithinViewport || !open.value || !buttonRef.value) return

  const buttonRect = buttonRef.value.getBoundingClientRect()
  const viewportMargin = 8
  const menuGap = 5
  const preferredMenuHeight = 240
  const spaceBelow = Math.max(0, window.innerHeight - buttonRect.bottom - viewportMargin - menuGap)
  const spaceAbove = Math.max(0, buttonRect.top - viewportMargin - menuGap)
  const openBelow = spaceBelow > spaceAbove + 32
  const availableHeight = Math.max(0, openBelow ? spaceBelow : spaceAbove)
  const menuTop = openBelow ? buttonRect.bottom + menuGap : buttonRect.top - menuGap

  menuStyle.value = {
    top: `${menuTop}px`,
    left: `${buttonRect.left}px`,
    width: `${buttonRect.width}px`,
    maxHeight: `${Math.min(preferredMenuHeight, availableHeight)}px`,
    transform: openBelow ? 'none' : 'translateY(-100%)',
  }
}

const closeMenu = () => {
  open.value = false
}

const openMenu = () => {
  if (props.disabled || props.options.length === 0) return
  const selectedIndex = props.options.findIndex(option => option.value === props.modelValue)
  highlightedIndex.value = selectedIndex >= 0 ? selectedIndex : 0
  open.value = true
  if (props.keepMenuWithinViewport) void nextTick(updateMenuPosition)
}

const toggleMenu = () => {
  if (open.value) {
    closeMenu()
  } else {
    openMenu()
  }
}

const selectOption = (value: SelectValue) => {
  emit('update:modelValue', value)
  closeMenu()
}

const moveHighlight = (direction: number) => {
  if (props.options.length === 0) return
  const lastIndex = props.options.length - 1
  const nextIndex = highlightedIndex.value + direction
  highlightedIndex.value = nextIndex < 0 ? lastIndex : nextIndex > lastIndex ? 0 : nextIndex
}

const onButtonKeydown = (event: KeyboardEvent) => {
  if (event.key === 'ArrowDown') {
    event.preventDefault()
    if (!open.value) openMenu()
    else moveHighlight(1)
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    if (!open.value) openMenu()
    else moveHighlight(-1)
  } else if (event.key === 'Enter' || event.key === ' ') {
    event.preventDefault()
    if (!open.value) {
      openMenu()
    } else {
      selectOption(props.options[highlightedIndex.value]?.value ?? null)
    }
  } else if (event.key === 'Escape') {
    event.preventDefault()
    closeMenu()
  }
}

const onDocumentPointerDown = (event: PointerEvent) => {
  if (!rootRef.value?.contains(event.target as Node)) {
    closeMenu()
  }
}

const onViewportChange = () => {
  updateMenuPosition()
}

watch(
  () => props.disabled,
  (disabled) => {
    if (disabled) closeMenu()
  }
)

onMounted(() => {
  document.addEventListener('pointerdown', onDocumentPointerDown)
  window.addEventListener('resize', onViewportChange)
  window.addEventListener('scroll', onViewportChange, true)
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onDocumentPointerDown)
  window.removeEventListener('resize', onViewportChange)
  window.removeEventListener('scroll', onViewportChange, true)
})
</script>

<style scoped>
.spoke-calculator-select {
  position: relative;
  width: 100%;
}

.spoke-calculator-select__button {
  display: flex;
  width: 100%;
  min-height: 3.125rem;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  border: 1px solid var(--spoke-border, var(--tz-border-subtle));
  border-radius: 9999px;
  background: var(--spoke-control-surface, var(--tz-input-surface));
  color: var(--tz-text-primary);
  padding: 0.75rem 0.875rem;
  font-size: 0.75rem;
  line-height: 1rem;
  text-align: left;
  transition:
    background-color 0.18s ease,
    border-color 0.18s ease,
    box-shadow 0.18s ease;
}

.spoke-calculator-select__button:hover,
.spoke-calculator-select__button[aria-expanded='true'] {
  border-color: var(--spoke-border-strong, var(--tz-border-strong));
  background: var(--spoke-result-surface, var(--tz-surface-subtle));
}

.spoke-calculator-select__button:focus-visible {
  outline: none;
  border-color: var(--spoke-border-strong, var(--tz-border-strong));
  box-shadow: 0 0 0 1px var(--spoke-focus-ring, var(--tz-form-control-focus-ring));
}

.spoke-calculator-select__button:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}

.spoke-calculator-select__button--placeholder {
  color: var(--tz-text-muted);
}

.spoke-calculator-select__label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.spoke-calculator-select__chevron {
  color: var(--tz-text-muted);
  font-size: 0.75rem;
  line-height: 1;
}

.spoke-calculator-select__menu {
  position: absolute;
  z-index: 30;
  top: calc(100% + 0.35rem);
  left: 0;
  width: 100%;
  max-height: 15rem;
  overflow-y: auto;
  border: 1px solid var(--spoke-border-strong, var(--tz-border-strong));
  border-radius: 0.5rem;
  background: var(--spoke-control-surface, var(--tz-input-surface));
  padding: 0.35rem;
  box-shadow: 0 18px 42px rgba(20, 32, 43, 0.14);
}

.spoke-calculator-select__menu--viewport-bound {
  position: fixed;
  top: auto;
  left: auto;
  width: auto;
  max-height: min(15rem, calc(100dvh - 1rem));
}

.spoke-calculator-select__option {
  display: block;
  width: 100%;
  border: 0;
  border-radius: 9999px;
  background: transparent;
  color: var(--tz-text-secondary);
  padding: 0.65rem 0.75rem;
  font-size: 0.75rem;
  line-height: 1rem;
  text-align: left;
  transition:
    background-color 0.16s ease,
    color 0.16s ease;
}

.spoke-calculator-select__option:hover,
.spoke-calculator-select__option--active {
  background: var(--tz-surface-muted);
  color: var(--tz-text-primary);
}

.spoke-calculator-select__option--selected {
  background: var(--tz-surface-subtle);
  color: var(--tz-text-primary);
}
</style>
