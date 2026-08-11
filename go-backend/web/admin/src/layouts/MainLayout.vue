<template>
  <TooltipProvider>
    <div class="flex h-screen h-dvh overflow-hidden bg-background">
      <aside
        class="hidden shrink-0 border-r border-dashed border-slate-200 bg-white shadow-sm transition-[width] duration-300 ease-[cubic-bezier(0.4,0,0.2,1)] lg:flex"
        :class="isCollapse ? 'w-[76px]' : 'w-[250px]'"
      >
        <AdminSidebar
          :items="visibleNavigationItems"
          :active-path="route.path"
          :collapsed="isCollapse"
          :brand-initial="brandInitial"
          :brand-name="brandName"
          :panel-label="panelLabel"
          :role-label="roleLabel"
          :user="user"
          :user-initials="userInitials"
          @request-logout="requestLogout"
        />
      </aside>

      <Sheet v-model:open="mobileSidebarOpen">
        <SheetContent
          side="left"
          class="gap-0 p-0 border-dashed"
        >
          <SheetTitle class="sr-only">{{ t('layout.navigation') }}</SheetTitle>
          <SheetDescription class="sr-only">{{ t('layout.selectModule') }}</SheetDescription>
          <AdminSidebar
            :items="visibleNavigationItems"
            :active-path="route.path"
            :brand-initial="brandInitial"
            :brand-name="brandName"
            :panel-label="panelLabel"
            :role-label="roleLabel"
            :user="user"
            :user-initials="userInitials"
            @request-logout="requestLogout"
            @navigate="mobileSidebarOpen = false"
          />
        </SheetContent>
      </Sheet>

      <section class="flex min-w-0 flex-1 flex-col">
        <header class="flex h-16 shrink-0 items-center border-b border-dashed border-slate-200 bg-white/80 px-3 backdrop-blur sm:px-4">
          <div class="flex min-w-0 items-center gap-2 sm:gap-3">
            <Button
              variant="ghost"
              size="icon"
              class="lg:hidden rounded-full"
              :aria-label="t('layout.openNavigation')"
              @click="mobileSidebarOpen = true"
            >
              <Menu class="size-4" />
            </Button>

            <Tooltip>
              <TooltipTrigger as-child>
                <Button
                  variant="ghost"
                  size="icon"
                  class="hidden rounded-full lg:inline-flex"
                  :aria-label="isCollapse ? t('layout.expandNavigation') : t('layout.collapseNavigation')"
                  @click="isCollapse = !isCollapse"
                >
                  <PanelLeftOpen v-if="isCollapse" class="size-4" />
                  <PanelLeftClose v-else class="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent side="bottom">
                {{ isCollapse ? t('layout.expandNavigation') : t('layout.collapseNavigation') }}
              </TooltipContent>
            </Tooltip>

            <div class="min-w-0">
              <span v-if="panelLabel" class="hidden text-[9px] font-black uppercase tracking-widest text-muted-foreground/60 sm:block">{{ panelLabel }}</span>
              <strong class="block truncate text-sm font-black tracking-tighter italic uppercase">{{ routeTitle }}</strong>
            </div>
          </div>
          <AdminLanguageSwitcher class="ml-auto" />
        </header>

        <main class="min-h-0 flex-1 overflow-auto bg-muted/35 p-3 sm:p-4 lg:p-6">
          <div class="mx-auto h-full min-h-0 w-full max-w-none sm:w-[95%]">
            <router-view />
          </div>
        </main>
      </section>
    </div>
  </TooltipProvider>

  <AlertDialog v-model:open="logoutDialogOpen">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>{{ t('layout.logoutTitle') }}</AlertDialogTitle>
        <AlertDialogDescription>
          {{ t('layout.logoutDescription') }}
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>{{ t('common.cancel') }}</AlertDialogCancel>
        <AlertDialogAction :disabled="logoutLoading" @click="confirmLogout">
          {{ logoutLoading ? t('common.loggingOut') : t('common.logout') }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch, watchEffect } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Menu,
  PanelLeftClose,
  PanelLeftOpen,
} from '@lucide/vue'
import AdminSidebar from '@/components/admin/AdminSidebar.vue'
import AdminLanguageSwitcher from '@/components/admin/AdminLanguageSwitcher.vue'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle
} from '@/components/ui/alert-dialog'
import { Button } from '@/components/ui/button'
import { Sheet, SheetContent, SheetDescription, SheetTitle } from '@/components/ui/sheet'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { useAdminBranding } from '@/composables/useAdminBranding'
import { useAuthStore } from '@/stores/auth'
import { adminNavigationItems, findActiveNavigationEntry, filterNavigationItems } from '@/lib/adminNavigation'
import { translateAdminNavigation, useAdminI18n } from '@/i18n'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const { t } = useAdminI18n()

const isCollapse = ref(false)
const mobileSidebarOpen = ref(false)
const logoutDialogOpen = ref(false)
const logoutLoading = ref(false)
const user = computed(() => authStore.user)
const { brandInitial, brandName, panelLabel, loadAdminBranding, setAdminDocumentTitle } = useAdminBranding()

const translatedNavigationItems = computed(() => translateAdminNavigation(adminNavigationItems, t))
const visibleNavigationItems = computed(() => filterNavigationItems(translatedNavigationItems.value, (permission) => authStore.hasPermission(permission)))
const activeNavigationEntry = computed(() => findActiveNavigationEntry(visibleNavigationItems.value, route.path))
const routeTitle = computed(() => {
  const parentLabel = activeNavigationEntry.value?.parent?.label
  const itemLabel = activeNavigationEntry.value?.item?.label

  if (parentLabel && itemLabel && parentLabel !== itemLabel) return `${parentLabel} / ${itemLabel}`
  return itemLabel || parentLabel || route.meta.title || t('layout.dashboardFallback')
})
const userInitials = computed(() => {
  const identity = user.value?.username || user.value?.email || 'Admin'
  const parts = identity.split(/[\s_-]+/).filter(Boolean)
  const initials = parts.length > 1 ? parts[0][0] + parts[parts.length - 1][0] : parts[0].slice(0, 2)
  return initials.toUpperCase()
})
const roleLabel = computed(() => {
  const labels: Record<string, string> = {
    admin: t('roles.admin'),
    manager: t('roles.manager'),
    editor: t('roles.editor'),
    support: t('roles.support'),
    viewer: t('roles.viewer')
  }
  return labels[user.value?.role || ''] || t('roles.backofficeUser')
})

const requestLogout = (): void => {
  mobileSidebarOpen.value = false
  logoutDialogOpen.value = true
}

const confirmLogout = async (): Promise<void> => {
  if (logoutLoading.value) return

  logoutLoading.value = true
  try {
    await authStore.logout()
    await router.push('/login')
  } finally {
    logoutLoading.value = false
    logoutDialogOpen.value = false
  }
}

onMounted(() => {
  loadAdminBranding()
})

watchEffect(() => {
  setAdminDocumentTitle(routeTitle.value)
})

watch(
  () => route.fullPath,
  () => {
    mobileSidebarOpen.value = false
  }
)
</script>
