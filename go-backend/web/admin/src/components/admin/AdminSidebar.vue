<template>
  <div class="relative flex h-full min-h-0 w-full flex-col bg-sidebar text-sidebar-foreground">
    <span v-if="!collapsed" class="pointer-events-none absolute inset-x-0 top-0 z-10 mx-auto w-full max-w-48 px-2 pb-2 pt-4 text-center text-[9px] font-black uppercase tracking-widest text-muted-foreground/60">
      WORKSPACE / 工作台
    </span>

    <nav
      class="flex min-h-0 flex-1 items-center justify-center overflow-y-auto px-0 py-3"
      aria-label="后台导航"
    >
      <TooltipProvider :delay-duration="0">
        <div
          class="admin-sidebar__menu"
          :class="collapsed ? 'admin-sidebar__menu--collapsed' : 'admin-sidebar__menu--expanded'"
        >
          <Tooltip v-for="item in items" :key="item.path">
            <TooltipTrigger as-child>
              <RouterLink
                :to="{ name: item.routeName }"
                class="admin-sidebar__link"
                :class="[
                  collapsed ? 'admin-sidebar__link--collapsed' : 'admin-sidebar__link--expanded',
                  isActive(item.path) ? 'admin-sidebar__link--active' : ''
                ]"
                :aria-current="isActive(item.path) ? 'page' : undefined"
                @click="emit('navigate')"
              >
                <span class="admin-sidebar__glow" aria-hidden="true" />
                <component :is="item.icon" class="admin-sidebar__icon" aria-hidden="true" />
                <span v-if="!collapsed" class="truncate">{{ item.label }}</span>
              </RouterLink>
            </TooltipTrigger>
            <TooltipContent v-if="collapsed" side="right" class="font-bold text-xs">
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

<style scoped>
.admin-sidebar__menu {
  display: flex;
  min-height: 100%;
  margin-inline: auto;
  padding-block: 1rem;
  flex-direction: column;
  justify-content: center;
}

.admin-sidebar__menu--expanded {
  width: calc(100% - 1rem);
  max-width: 12rem;
  padding-inline: 0.5rem;
}

.admin-sidebar__menu--collapsed {
  width: calc(100% - 0.5rem);
  max-width: 3rem;
  align-items: center;
  padding-inline: 0.5rem;
}

.admin-sidebar__link {
  position: relative;
  isolation: isolate;
  display: flex;
  height: 2.25rem;
  margin-bottom: 0.25rem;
  align-items: center;
  gap: 0.625rem;
  overflow: hidden;
  border: 1px solid transparent;
  border-radius: 999px;
  color: var(--muted-foreground);
  font-size: 0.75rem;
  font-weight: 900;
  letter-spacing: 0.04em;
  line-height: 1rem;
  text-decoration: none;
  text-transform: uppercase;
  transition:
    color 160ms ease,
    background-color 160ms ease,
    border-color 160ms ease,
    box-shadow 160ms ease,
    transform 160ms ease;
}

.admin-sidebar__link--expanded {
  width: 100%;
  justify-content: center;
  padding-inline: 0.75rem;
}

.admin-sidebar__link--collapsed {
  width: 2.25rem;
  justify-content: center;
  padding-inline: 0;
}

.admin-sidebar__glow {
  position: absolute;
  inset: 1px;
  z-index: -1;
  border-radius: inherit;
  background:
    linear-gradient(135deg, color-mix(in oklch, var(--primary) 16%, transparent), transparent 48%),
    linear-gradient(
      180deg,
      color-mix(in oklch, var(--sidebar-accent) 92%, transparent),
      color-mix(in oklch, var(--sidebar-accent) 60%, transparent)
    );
  opacity: 0;
  transition: opacity 160ms ease;
}

.admin-sidebar__icon {
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
  transition: color 160ms ease, transform 160ms ease;
}

.admin-sidebar__link:hover,
.admin-sidebar__link:focus-visible {
  transform: translateY(-1px);
  border-color: color-mix(in oklch, var(--sidebar-border) 85%, transparent);
  color: var(--sidebar-accent-foreground);
  box-shadow:
    0 8px 18px color-mix(in oklch, var(--sidebar-ring) 10%, transparent),
    inset 0 1px 0 hsl(0 0% 100% / 0.12);
}

.admin-sidebar__link:hover .admin-sidebar__glow,
.admin-sidebar__link:focus-visible .admin-sidebar__glow,
.admin-sidebar__link--active .admin-sidebar__glow {
  opacity: 1;
}

.admin-sidebar__link:hover .admin-sidebar__icon,
.admin-sidebar__link:focus-visible .admin-sidebar__icon {
  color: var(--sidebar-ring);
  transform: scale(1.05);
}

.admin-sidebar__link:focus-visible {
  outline: 2px solid color-mix(in oklch, var(--sidebar-ring) 34%, transparent);
  outline-offset: 2px;
}

.admin-sidebar__link:active {
  transform: translateY(0) scale(0.98);
}

.admin-sidebar__link--active {
  border-color: color-mix(in oklch, var(--sidebar-ring) 28%, transparent);
  color: var(--sidebar-accent-foreground);
  box-shadow:
    0 8px 20px color-mix(in oklch, var(--sidebar-ring) 12%, transparent),
    inset 0 1px 0 hsl(0 0% 100% / 0.14);
}

.admin-sidebar__link--active .admin-sidebar__icon {
  color: var(--sidebar-ring);
}
</style>
