<template>
  <Card class="min-w-0 gap-0 py-0 shadow-none rounded-[24px] border-dashed border-border/80">
    <div class="uds-glow-bg" />
    <CardHeader class="relative z-10 border-b border-dashed border-border/70 py-3.5">
      <CardTitle class="text-sm font-black tracking-tighter uppercase">快速操作</CardTitle>
      <CardDescription class="text-[9px] font-black uppercase tracking-widest opacity-60">常用管理入口</CardDescription>
    </CardHeader>
    <CardContent class="relative z-10 grid grid-cols-1 gap-1.5 p-3 sm:grid-cols-2 xl:grid-cols-1">
      <Button
        v-for="action in actions"
        :key="action.path"
        variant="ghost"
        class="h-9 w-full justify-start gap-2.5 px-3 rounded-full hover:bg-primary/10 transition-all font-bold text-xs"
        @click="emit('navigate', action.path)"
      >
        <span class="flex size-6 shrink-0 items-center justify-center rounded-full" :class="metricToneClass(action.tone)">
          <component :is="action.icon" class="size-3" />
        </span>
        <span class="truncate tracking-tight">{{ action.label }}</span>
        <ArrowRight class="ml-auto size-3.5 text-muted-foreground/60" />
      </Button>
    </CardContent>
  </Card>
</template>

<script setup lang="ts">
import { ArrowRight } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import type { DashboardMetricToneClass, DashboardQuickAction } from './dashboardTypes'

withDefaults(defineProps<{
  actions?: DashboardQuickAction[]
  metricToneClass: DashboardMetricToneClass
}>(), {
  actions: () => []
})

const emit = defineEmits<{
  (event: 'navigate', path: string): void
}>()
</script>
