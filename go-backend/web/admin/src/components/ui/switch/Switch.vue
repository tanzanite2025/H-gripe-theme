<script setup lang="ts">
import type { SwitchRootEmits, SwitchRootProps } from 'reka-ui'
import type { HTMLAttributes } from 'vue'
import { reactiveOmit } from '@vueuse/core'
import {
  SwitchRoot,
  SwitchThumb,
  useForwardPropsEmits,
} from 'reka-ui'
import { cn } from '@/lib/utils'

const props = withDefaults(defineProps<SwitchRootProps & {
  class?: HTMLAttributes['class']
  size?: 'sm' | 'default'
}>(), {
  size: 'default',
})

const emits = defineEmits<SwitchRootEmits>()
const delegatedProps = reactiveOmit(props, 'class', 'size')
const forwarded = useForwardPropsEmits(delegatedProps, emits)
</script>

<template>
  <SwitchRoot
    v-slot="slotProps"
    data-slot="switch"
    :data-size="size"
    v-bind="forwarded"
    :class="cn(
      'admin-switch peer group/switch relative inline-flex shrink-0 items-center rounded-full border outline-none transition-colors after:absolute after:-inset-x-3 after:-inset-y-2 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-600 data-[size=default]:h-[26px] data-[size=default]:w-[46px] data-[size=sm]:h-[22px] data-[size=sm]:w-[36px] data-[state=unchecked]:border-slate-400 data-[state=unchecked]:bg-slate-300 data-[state=checked]:border-emerald-700 data-[state=checked]:bg-emerald-500 data-[disabled]:cursor-not-allowed data-[disabled]:opacity-60 dark:data-[state=unchecked]:border-slate-600 dark:data-[state=unchecked]:bg-slate-700 dark:data-[state=checked]:border-emerald-400 dark:data-[state=checked]:bg-emerald-600',
      props.class,
    )"
  >
    <SwitchThumb
      data-slot="switch-thumb"
      class="admin-switch-thumb pointer-events-none block rounded-full bg-white shadow-[0_1px_3px_rgb(15_23_42_/_0.35)] ring-0 transition-transform dark:bg-slate-100"
    >
      <slot name="thumb" v-bind="slotProps" />
    </SwitchThumb>
  </SwitchRoot>
</template>

<style>
.admin-switch[data-state='unchecked'] {
  border-color: #94a3b8;
  background: #cbd5e1;
}

.admin-switch[data-state='checked'] {
  border-color: #047857;
  background: #10b981;
}

.admin-switch[data-disabled] {
  cursor: not-allowed;
  opacity: 0.6;
}

.admin-switch-thumb {
  margin-left: 3px;
  display: block;
  border-radius: 999px;
  background: #ffffff;
  box-shadow: 0 1px 3px rgb(15 23 42 / 0.35);
  pointer-events: none;
  transition: transform 160ms ease;
}

.admin-switch[data-size='default'] .admin-switch-thumb {
  width: 20px;
  height: 20px;
}

.admin-switch[data-size='sm'] .admin-switch-thumb {
  width: 16px;
  height: 16px;
}

.admin-switch[data-size='default'][data-state='checked'] .admin-switch-thumb {
  transform: translateX(20px);
}

.admin-switch[data-size='sm'][data-state='checked'] .admin-switch-thumb {
  transform: translateX(14px);
}

.dark .admin-switch[data-state='unchecked'] {
  border-color: #475569;
  background: #334155;
}

.dark .admin-switch[data-state='checked'] {
  border-color: #34d399;
  background: #059669;
}
</style>
