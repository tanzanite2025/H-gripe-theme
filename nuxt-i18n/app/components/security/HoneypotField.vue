<template>
  <div
    class="honeypot-field"
    aria-hidden="true"
    inert
  >
    <label :for="fieldId">{{ label }}</label>
    <input
      :id="fieldId"
      v-model="model"
      type="text"
      :name="name"
      tabindex="-1"
      autocomplete="off"
      :aria-label="label"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    name: string
    label?: string
  }>(),
  {
    label: 'Additional website',
  },
)

const model = defineModel<string>({ default: '' })
const fieldId = computed(() => `decoy-${props.name.replace(/[^a-z0-9_-]/gi, '-')}`)
</script>

<style scoped>
.honeypot-field {
  position: absolute;
  top: -9999px;
  left: -9999px;
  width: 1px;
  height: 1px;
  overflow: hidden;
  opacity: 0;
  pointer-events: none;
}
</style>
