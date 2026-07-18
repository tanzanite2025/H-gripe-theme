<template>
  <router-view />
  <Toaster rich-colors position="top-right" />
</template>

<script setup>
import { onMounted } from 'vue'
import { toast } from 'vue-sonner'
import { Toaster } from '@/components/ui/sonner'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

onMounted(async () => {
  const result = await authStore.initAuth()

  if (result.authenticated && result.permissionsUpdated) {
    toast.info('权限已更新，请注意菜单变化', { id: 'permissions-updated' })
  }

  if (result.warning) {
    console.warn('[App] Auth warning:', result.warning)
  }
})
</script>
