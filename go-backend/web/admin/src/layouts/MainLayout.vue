<template>
  <TooltipProvider>
    <div class="flex h-screen h-dvh overflow-hidden bg-background">
      <aside
        class="hidden shrink-0 border-r border-dashed border-sidebar-border bg-sidebar transition-[width] duration-200 lg:flex"
        :class="isCollapse ? 'w-[72px]' : 'w-[232px]'"
      >
        <AdminSidebar
          :items="visibleNavigationItems"
          :active-path="route.path"
          :collapsed="isCollapse"
        />
      </aside>

      <Sheet v-model:open="mobileSidebarOpen">
        <SheetContent
          side="left"
          class="gap-0 p-0 border-dashed"
        >
          <SheetTitle class="sr-only">后台导航</SheetTitle>
          <SheetDescription class="sr-only">选择要进入的后台管理模块</SheetDescription>
          <AdminSidebar
            :items="visibleNavigationItems"
            :active-path="route.path"
            @navigate="mobileSidebarOpen = false"
          />
        </SheetContent>
      </Sheet>

      <section class="flex min-w-0 flex-1 flex-col">
        <header class="flex h-14 shrink-0 items-center justify-between border-b border-dashed border-border bg-card px-3 sm:px-4">
          <div class="flex min-w-0 items-center gap-2 sm:gap-3">
            <Button
              variant="ghost"
              size="icon"
              class="lg:hidden rounded-full"
              aria-label="打开导航"
              @click="mobileSidebarOpen = true"
            >
              <Menu class="size-4" />
            </Button>

            <Tooltip>
              <TooltipTrigger as-child>
                <Button
                  variant="ghost"
                  size="icon"
                  class="hidden lg:inline-flex rounded-full"
                  :aria-label="isCollapse ? '展开导航' : '收起导航'"
                  @click="isCollapse = !isCollapse"
                >
                  <PanelLeftOpen v-if="isCollapse" class="size-4" />
                  <PanelLeftClose v-else class="size-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent side="bottom">
                {{ isCollapse ? '展开导航' : '收起导航' }}
              </TooltipContent>
            </Tooltip>

            <div class="min-w-0">
              <span class="hidden text-[9px] font-black uppercase tracking-widest text-muted-foreground/60 sm:block">SYSTEM CONTROL / 运营后台</span>
              <strong class="block truncate text-sm font-black tracking-tighter italic uppercase">{{ routeTitle }}</strong>
            </div>
          </div>

          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <Button variant="ghost" class="h-auto gap-2 px-1.5 py-1 sm:px-2">
                <Avatar class="size-8 rounded-full">
                  <AvatarFallback class="rounded-full bg-primary/10 font-mono text-xs font-black text-primary">
                    {{ userInitials }}
                  </AvatarFallback>
                </Avatar>
                <span class="hidden max-w-36 flex-col items-start leading-tight sm:flex">
                  <strong class="w-full truncate text-xs font-bold">{{ user?.username || '管理员' }}</strong>
                  <small class="mt-0.5 text-[9px] font-black uppercase tracking-widest text-muted-foreground/70">{{ roleLabel }}</small>
                </span>
                <ChevronDown class="size-3.5 text-muted-foreground" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" class="w-60">
              <DropdownMenuLabel class="font-normal">
                <span class="block text-xs font-bold">{{ user?.username || '管理员' }}</span>
                <span class="mt-1 block truncate font-mono text-[10px] text-muted-foreground/70">{{ user?.email }}</span>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem class="text-destructive focus:text-destructive" @select="logoutDialogOpen = true">
                <LogOut class="size-4" />
                退出登录
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </header>

        <main class="min-h-0 flex-1 overflow-auto bg-muted/35 p-3 sm:p-4 lg:p-6">
          <div class="mx-auto w-full max-w-[1600px]">
            <router-view />
          </div>
        </main>
      </section>
    </div>
  </TooltipProvider>

  <AlertDialog v-model:open="logoutDialogOpen">
    <AlertDialogContent>
      <AlertDialogHeader>
        <AlertDialogTitle>退出登录？</AlertDialogTitle>
        <AlertDialogDescription>
          当前管理会话将结束，需要重新登录后才能继续操作。
        </AlertDialogDescription>
      </AlertDialogHeader>
      <AlertDialogFooter>
        <AlertDialogCancel>取消</AlertDialogCancel>
        <AlertDialogAction :disabled="logoutLoading" @click="confirmLogout">
          {{ logoutLoading ? '正在退出' : '退出' }}
        </AlertDialogAction>
      </AlertDialogFooter>
    </AlertDialogContent>
  </AlertDialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  ChevronDown,
  CircleHelp,
  FileText,
  Images,
  LayoutDashboard,
  LogOut,
  Mail,
  Megaphone,
  Menu,
  MessagesSquare,
  Package,
  PanelLeftClose,
  PanelLeftOpen,
  ScrollText,
  Settings,
  ShoppingCart,
  Tags,
  Truck,
  Users
} from '@lucide/vue'
import AdminSidebar from '@/components/admin/AdminSidebar.vue'
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
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger
} from '@/components/ui/dropdown-menu'
import { Sheet, SheetContent, SheetDescription, SheetTitle } from '@/components/ui/sheet'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const isCollapse = ref(false)
const mobileSidebarOpen = ref(false)
const logoutDialogOpen = ref(false)
const logoutLoading = ref(false)
const user = computed(() => authStore.user)

const navigationItems = [
  { path: '/', routeName: 'Dashboard', label: '仪表板', icon: LayoutDashboard },
  { path: '/products', routeName: 'Products', label: '商品管理', icon: Package, permission: 'product:view' },
  { path: '/product-types', routeName: 'ProductTypes', label: '产品模板', icon: Tags, permission: 'product:view' },
  { path: '/orders', routeName: 'Orders', label: '订单管理', icon: ShoppingCart, permission: 'order:view' },
  { path: '/shipping', routeName: 'Shipping', label: '物流管理', icon: Truck, permission: 'shipping:view' },
  { path: '/users', routeName: 'Users', label: '用户管理', icon: Users, permission: 'user:view' },
  { path: '/content', routeName: 'Content', label: '内容管理', icon: FileText, permission: 'content:view' },
  { path: '/faqs', routeName: 'FAQs', label: 'FAQ 管理', icon: CircleHelp, permission: 'faq:view' },
  { path: '/galleries', routeName: 'Galleries', label: '图库管理', icon: Images, permission: 'gallery:view' },
  { path: '/subscriptions', routeName: 'Subscriptions', label: '订阅管理', icon: Mail, permission: 'subscription:view' },
  { path: '/tickets', routeName: 'Tickets', label: '工单管理', icon: MessagesSquare, permission: 'ticket:view' },
  { path: '/marketing', routeName: 'Marketing', label: '营销管理', icon: Megaphone, permission: 'marketing:view' },
  { path: '/settings', routeName: 'Settings', label: '系统设置', icon: Settings, permission: 'settings:view' },
  { path: '/audit-logs', routeName: 'AuditLogs', label: '审计日志', icon: ScrollText, permission: 'logs:view' }
]

const visibleNavigationItems = computed(() =>
  navigationItems.filter((item) => !item.permission || authStore.hasPermission(item.permission))
)
const routeTitle = computed(() => route.meta.title || '仪表板')
const userInitials = computed(() => {
  const identity = user.value?.username || user.value?.email || 'Admin'
  const parts = identity.split(/[\s_-]+/).filter(Boolean)
  const initials = parts.length > 1 ? parts[0][0] + parts[parts.length - 1][0] : parts[0].slice(0, 2)
  return initials.toUpperCase()
})
const roleLabel = computed(() => {
  const labels = {
    admin: '管理员',
    manager: '经理',
    editor: '编辑',
    support: '客服',
    viewer: '查看者'
  }
  return labels[user.value?.role] || '后台用户'
})

const confirmLogout = async () => {
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

watch(
  () => route.fullPath,
  () => {
    mobileSidebarOpen.value = false
  }
)
</script>
