<template>
  <Card class="gap-0 py-0 shadow-none">
    <CardContent class="p-3 sm:p-4">
      <slot />
    </CardContent>
  </Card>
</template>

<script setup>
import { Card, CardContent } from '@/components/ui/card'
</script>
