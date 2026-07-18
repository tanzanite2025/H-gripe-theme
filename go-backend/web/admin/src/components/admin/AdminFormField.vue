<template>
  <label class="block space-y-1.5">
    <span class="text-xs font-medium">
      {{ label }}<span v-if="required" class="ml-0.5 text-destructive" aria-hidden="true">*</span>
    </span>
    <slot />
    <span v-if="error" class="block text-xs font-medium text-destructive">{{ error }}</span>
    <span v-else-if="description" class="block text-xs text-muted-foreground">{{ description }}</span>
  </label>
</template>

<script setup>
defineProps({
  label: { type: String, required: true },
  required: { type: Boolean, default: false },
  error: { type: String, default: '' },
  description: { type: String, default: '' }
})
</script>
