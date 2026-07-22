<template>
  <Card class="gap-0 py-0 shadow-none rounded-[24px] border-dashed border-border/80">
    <div class="uds-glow-bg" />
    <CardContent class="relative z-10 p-3 sm:p-4">
      <slot />
    </CardContent>
  </Card>
</template>

<script setup>
import { Card, CardContent } from '@/components/ui/card'
</script>
