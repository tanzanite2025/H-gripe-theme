<template>
  <label class="block space-y-1">
    <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">
      {{ label }}<span v-if="required" class="ml-0.5 text-destructive" aria-hidden="true">*</span>
    </span>
    <slot />
    <span v-if="error" class="block text-[10px] font-bold text-destructive">{{ error }}</span>
    <span v-else-if="description" class="block text-[9px] font-black uppercase tracking-widest text-muted-foreground/60 mt-0.5">{{ description }}</span>
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
