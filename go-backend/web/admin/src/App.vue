<template>
  <div
    v-if="isBusy"
    class="fixed inset-x-0 top-0 z-[100] h-0.5 overflow-hidden bg-primary/10"
    aria-live="polite"
    aria-label="正在加载"
  >
    <div class="h-full w-1/3 animate-[admin-loading-bar_1.1s_ease-in-out_infinite] bg-gradient-to-r from-primary via-sky-400 to-primary" />
  </div>
  <router-view />
  <Toaster rich-colors position="top-right" />
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { toast } from 'vue-sonner'
import { Toaster } from '@/components/ui/sonner'
import { useAuthStore } from '@/stores/auth'
import router from '@/router'

const authStore = useAuthStore()
const routeLoading = ref(false)
const apiLoading = ref(false)
const isBusy = computed(() => routeLoading.value || apiLoading.value)

const removeBeforeResolve = router.beforeResolve(() => {
  routeLoading.value = true
})
const removeAfterEach = router.afterEach(() => {
  window.requestAnimationFrame(() => {
    routeLoading.value = false
  })
})
const removeOnError = router.onError(() => {
  routeLoading.value = false
})

const onApiLoading = (event) => {
  apiLoading.value = Boolean(event.detail?.loading)
}

onMounted(async () => {
  window.addEventListener('admin-api-loading', onApiLoading)
  const result = await authStore.initAuth()

  if (result.authenticated && result.permissionsUpdated) {
    toast.info('权限已更新，请注意菜单变化', { id: 'permissions-updated' })
  }

  if (result.warning) {
    console.warn('[App] Auth warning:', result.warning)
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('admin-api-loading', onApiLoading)
  removeBeforeResolve()
  removeAfterEach()
  removeOnError()
})
</script>

<style>
@keyframes admin-loading-bar {
  0% {
    transform: translateX(-120%);
  }
  100% {
    transform: translateX(320%);
  }
}
</style>
