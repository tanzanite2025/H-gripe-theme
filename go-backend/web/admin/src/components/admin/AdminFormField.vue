<template>
  <div class="block space-y-1">
    <label v-if="forId" :for="forId" class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
      {{ label }}<span v-if="required" class="ml-0.5 text-destructive" aria-hidden="true">*</span>
    </label>
    <span v-else class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
      {{ label }}<span v-if="required" class="ml-0.5 text-destructive" aria-hidden="true">*</span>
    </span>
    <slot />
    <span v-if="error" class="block text-[10px] font-bold text-destructive">{{ error }}</span>
    <span v-else-if="description" class="block text-[9px] font-black uppercase tracking-widest text-muted-foreground/60 mt-0.5">{{ description }}</span>
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  label: string
  forId?: string
  required?: boolean
  error?: string
  description?: string
}>(), {
  forId: '',
  required: false,
  error: '',
  description: ''
})
</script>
