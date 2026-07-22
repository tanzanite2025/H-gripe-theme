<script setup lang="ts">
import type { SelectTriggerProps } from 'reka-ui'

import type { HTMLAttributes } from 'vue'
import { ChevronDownIcon } from '@lucide/vue'
import { reactiveOmit } from '@vueuse/core'
import { SelectIcon, SelectTrigger, useForwardProps } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<SelectTriggerProps & { class?: HTMLAttributes['class'], size?: 'sm' | 'default' }>(),
  { size: 'default' },
)

const delegatedProps = reactiveOmit(props, 'class', 'size')
const forwardedProps = useForwardProps(delegatedProps)
</script>

<template>
  <SelectTrigger
    data-slot="select-trigger"
    :data-size="size"
    v-bind="forwardedProps"
    :class="cn(
      'bg-muted/50 border-none focus-visible:ring-ring/50 gap-1.5 rounded-xl py-1.5 pr-2.5 pl-3 text-xs font-bold transition-all select-none focus-visible:ring-2 data-[size=default]:h-9 data-[size=sm]:h-8 *:data-[slot=select-value]:gap-1.5 [&_svg:not([class*=size-])]:size-3.5 flex w-full items-center justify-between whitespace-nowrap outline-none disabled:cursor-not-allowed disabled:opacity-50 [&_[data-slot=select-value][data-placeholder]]:text-[11px] [&_[data-slot=select-value][data-placeholder]]:font-normal [&_[data-slot=select-value][data-placeholder]]:text-muted-foreground/50',
      props.class,
    )"
  >
    <slot />
    <SelectIcon as-child>
      <ChevronDownIcon class="text-muted-foreground size-4 pointer-events-none" />
    </SelectIcon>
  </SelectTrigger>
</template>
