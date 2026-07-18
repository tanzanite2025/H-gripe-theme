<template>
  <div class="flex h-full min-h-0 flex-col bg-sidebar text-sidebar-foreground">
    <RouterLink
      :to="{ name: 'Dashboard' }"
      class="flex h-14 shrink-0 items-center gap-2.5 border-b border-sidebar-border px-4 text-sidebar-foreground no-underline"
      :class="collapsed ? 'justify-center px-0' : ''"
      aria-label="返回仪表板"
      @click="emit('navigate')"
    >
      <span class="flex size-8 shrink-0 items-center justify-center rounded-lg bg-sidebar-primary text-sm font-bold text-sidebar-primary-foreground">
        T
      </span>
      <span v-if="!collapsed" class="flex min-w-0 flex-col leading-none">
        <strong class="truncate text-sm font-semibold">Tanzanite</strong>
        <small class="mt-1 text-[9px] font-semibold text-muted-foreground">ADMIN</small>
      </span>
    </RouterLink>

    <span v-if="!collapsed" class="px-4 pb-2 pt-5 text-[10px] font-semibold text-muted-foreground">
      工作台
    </span>

    <nav class="min-h-0 flex-1 overflow-y-auto px-2 py-2" aria-label="后台导航">
      <TooltipProvider :delay-duration="0">
        <Tooltip v-for="item in items" :key="item.path">
          <TooltipTrigger as-child>
            <RouterLink
              :to="{ name: item.routeName }"
              class="mb-1 flex h-9 items-center gap-2.5 rounded-lg px-2.5 text-sm font-medium text-muted-foreground no-underline transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
              :class="[
                collapsed ? 'justify-center px-0' : '',
                isActive(item.path) ? 'bg-sidebar-accent text-sidebar-accent-foreground' : ''
              ]"
              :aria-current="isActive(item.path) ? 'page' : undefined"
              @click="emit('navigate')"
            >
              <component :is="item.icon" class="size-4 shrink-0" aria-hidden="true" />
              <span v-if="!collapsed" class="truncate">{{ item.label }}</span>
            </RouterLink>
          </TooltipTrigger>
          <TooltipContent v-if="collapsed" side="right">
            {{ item.label }}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </nav>
  </div>
</template>

<script setup>
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger
} from '@/components/ui/tooltip'

const props = defineProps({
  collapsed: {
    type: Boolean,
    default: false
  },
  activePath: {
    type: String,
    required: true
  },
  items: {
    type: Array,
    required: true
  }
})

const emit = defineEmits(['navigate'])

const isActive = (path) => {
  if (path === '/') return props.activePath === '/'
  return props.activePath === path || props.activePath.startsWith(path + '/')
}
</script>
