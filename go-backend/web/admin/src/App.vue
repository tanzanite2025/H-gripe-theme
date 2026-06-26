<template>
  <router-view />
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const authStore = useAuthStore()

onMounted(async () => {
  // 页面加载时初始化认证状态并验证权限
  const result = await authStore.initAuth()

  if (result.authenticated) {
    if (result.permissionsUpdated) {
      ElMessage.info('权限已更新，请注意菜单变化')
    }

    if (result.warning) {
      console.warn('[App] Auth warning:', result.warning)
    }
  }
})
</script>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

#app {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial,
    'Noto Sans', sans-serif, 'Apple Color Emoji', 'Segoe UI Emoji', 'Segoe UI Symbol',
    'Noto Color Emoji';
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}
</style>
