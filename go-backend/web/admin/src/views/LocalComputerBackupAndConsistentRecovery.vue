<template>
  <div class="space-y-4">
    <AdminPageHeader title="业务数据备份与恢复" description="选择整站或业务模块备份到电脑，上传本地包校验，并查看支持的恢复方式。">
      <template #actions>
        <a :href="consolePath" target="_blank" rel="noopener" class="rounded-md border px-3 py-2 text-sm">打开独立恢复入口</a>
        <button type="button" class="rounded-md border px-3 py-2 text-sm" :disabled="connecting" @click="connectIndependentMaintenanceConsole">重新连接</button>
      </template>
    </AdminPageHeader>
    <p class="text-sm text-muted-foreground">整体备份生成期间商城会暂时停止服务。请先安排维护窗口，并单独保管维护凭据和备份解密凭据；原数据库或业务登录不可用时，可直接访问独立恢复入口。</p>
    <p v-if="connectionError" role="alert" class="text-sm text-destructive">{{ connectionError }}</p>
    <p v-if="connecting" role="status" class="rounded-lg border p-4 text-sm text-muted-foreground">正在连接独立备份服务…</p>
    <iframe v-if="mountMaintenanceConsole" v-show="!connecting" :key="connectionAttempt" ref="maintenanceConsoleFrame" :src="consolePath" title="独立本地备份与恢复控制台" class="min-h-[1200px] w-full rounded-lg border bg-background" />
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import axios from '@/utils/axios'

const consolePath = '/local-backup-maintenance/'
const maintenanceConsoleFrame = ref<HTMLIFrameElement | null>(null)
const connecting = ref(false)
const connectionError = ref('')
const mountMaintenanceConsole = ref(false)
const connectionAttempt = ref(0)
let pendingMaintenanceToken = ''
let connectionTimeout: number | undefined
let componentUnmounted = false

function finishMaintenanceConsoleConnection(errorMessage = '') {
  window.clearTimeout(connectionTimeout)
  pendingMaintenanceToken = ''
  connecting.value = false
  connectionError.value = errorMessage
  if (errorMessage) mountMaintenanceConsole.value = false
}

async function connectIndependentMaintenanceConsole() {
  if (connecting.value || componentUnmounted) return
  connecting.value = true
  connectionError.value = ''
  mountMaintenanceConsole.value = false
  pendingMaintenanceToken = ''
  const attempt = ++connectionAttempt.value
  try {
    const response = await axios.post('/api/admin/settings/backup-recovery/maintenance-session', null, { suppressGlobalErrorToast: true })
    if (componentUnmounted || attempt !== connectionAttempt.value) return
    const session = response.data
    if (!session || typeof session.token !== 'string') throw new Error('维护会话响应无效')
    pendingMaintenanceToken = session.token
    mountMaintenanceConsole.value = true
    connectionTimeout = window.setTimeout(() => finishMaintenanceConsoleConnection('独立备份服务未在规定时间内响应，请检查服务是否已启动后重新连接。'), 15000)
  } catch (error: unknown) {
    if (componentUnmounted) return
    const serverError = (error as { response?: { data?: { error?: unknown } } })?.response?.data?.error
    finishMaintenanceConsoleConnection(typeof serverError === 'string' ? serverError : '独立备份服务连接失败，请检查服务配置后重新连接。')
  }
}

function receiveIndependentMaintenanceConsoleConnectionStatus(event: MessageEvent) {
  if (event.origin !== window.location.origin || event.source !== maintenanceConsoleFrame.value?.contentWindow) return
  if (event.data?.type === 'local-backup-maintenance-ready' && pendingMaintenanceToken) {
    maintenanceConsoleFrame.value?.contentWindow?.postMessage({ type: 'local-backup-maintenance-session', token: pendingMaintenanceToken }, window.location.origin)
    pendingMaintenanceToken = ''
  } else if (event.data?.type === 'local-backup-maintenance-connected') {
    finishMaintenanceConsoleConnection()
  } else if (event.data?.type === 'local-backup-maintenance-connection-failed') {
    finishMaintenanceConsoleConnection('独立备份服务未能建立维护会话，请检查服务状态和凭据配置后重新连接。')
  }
}

onMounted(() => {
  window.addEventListener('message', receiveIndependentMaintenanceConsoleConnectionStatus)
  void connectIndependentMaintenanceConsole()
})
onBeforeUnmount(() => {
  componentUnmounted = true
  window.clearTimeout(connectionTimeout)
  pendingMaintenanceToken = ''
  window.removeEventListener('message', receiveIndependentMaintenanceConsoleConnectionStatus)
})
</script>
