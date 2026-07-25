<template>
  <Card class="min-w-0 gap-0 py-0 shadow-none rounded-[24px] border-dashed border-border/80">
    <div class="uds-glow-bg" />
    <CardHeader class="relative z-10 flex flex-row items-center justify-between border-b border-dashed border-border/70 py-3.5">
      <div>
        <CardTitle class="text-sm font-black tracking-tighter italic uppercase">销售趋势</CardTitle>
        <CardDescription class="mt-0.5 text-[9px] font-black uppercase tracking-widest opacity-60">最近 30 天表现</CardDescription>
      </div>
      <Tooltip>
        <TooltipTrigger as-child>
          <Button
            variant="outline"
            size="icon"
            class="rounded-full border-dashed size-8"
            aria-label="刷新销售趋势"
            :disabled="chartLoading"
            @click="emit('refresh')"
          >
            <RefreshCw class="size-3.5" :class="chartLoading ? 'animate-spin' : ''" />
          </Button>
        </TooltipTrigger>
        <TooltipContent class="font-bold text-xs">刷新销售趋势</TooltipContent>
      </Tooltip>
    </CardHeader>
    <CardContent class="relative z-10 flex h-80 items-center justify-center p-4">
      <div v-if="chartLoading" class="w-full space-y-4">
        <Skeleton class="h-4 w-36 rounded-full" />
        <Skeleton class="h-56 w-full rounded-2xl" />
      </div>
      <v-chart v-else-if="chartOption" class="h-full w-full" :option="chartOption" autoresize />
      <div v-else class="flex flex-col items-center text-center text-muted-foreground">
        <ChartNoAxesCombined class="mb-3 size-8 opacity-55 text-primary" />
        <p class="text-xs font-black uppercase tracking-wider text-foreground/75">暂无销售数据</p>
      </div>
    </CardContent>
  </Card>
</template>

<script setup>
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import { ChartNoAxesCombined, RefreshCw } from '@lucide/vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'

use([CanvasRenderer, LineChart, GridComponent, TooltipComponent, LegendComponent])

defineProps({
  chartLoading: { type: Boolean, default: false },
  chartOption: { type: Object, default: null },
})

const emit = defineEmits(['refresh'])
</script>
