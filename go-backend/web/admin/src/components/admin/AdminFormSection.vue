<template>
  <section class="space-y-4 border-t border-dashed pt-6 first:border-t-0 first:pt-0">
    <div class="space-y-1">
      <h3 class="text-sm font-black uppercase tracking-tighter text-foreground">{{ title }}</h3>
      <p v-if="description" class="max-w-2xl text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">
        {{ description }}
      </p>
    </div>
    <div class="min-w-0">
      <slot />
    </div>
  </section>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  title: string
  description?: string
}>(), {
  description: ''
})
</script>
