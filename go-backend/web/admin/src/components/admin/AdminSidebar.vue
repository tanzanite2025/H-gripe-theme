<template>
  <div class="flex h-full min-h-0 flex-col bg-sidebar text-sidebar-foreground">
    <span v-if="!collapsed" class="mx-auto w-full max-w-48 shrink-0 px-2 pb-2 pt-5 text-[10px] font-semibold text-muted-foreground">
      工作台
    </span>

    <nav
      class="min-h-0 flex-1 overflow-y-auto px-0 pb-3"
      :class="collapsed ? 'pt-5' : 'pt-2'"
      aria-label="后台导航"
    >
      <TooltipProvider :delay-duration="0">
        <div
          class="mx-auto flex flex-col"
          :class="collapsed ? 'w-full items-center px-2' : 'w-full max-w-48 px-2'"
        >
          <Tooltip v-for="item in items" :key="item.path">
            <TooltipTrigger as-child>
              <RouterLink
                :to="{ name: item.routeName }"
                class="mb-1 flex h-9 items-center gap-2.5 rounded-lg px-2.5 text-sm font-medium text-muted-foreground no-underline transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                :class="[
                  collapsed ? 'w-9 justify-center px-0' : 'w-full',
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
        </div>
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
