<template>
  <div class="space-y-4">
    <AdminPageHeader title="业务数据备份与恢复" description="选择整站或业务模块备份到电脑，上传本地包校验，并查看支持的恢复方式。">
      <template #actions>
        <Button as-child variant="outline"><a :href="consolePath" target="_blank" rel="noopener">打开独立恢复入口</a></Button>
        <Button variant="outline" :disabled="connecting" @click="connectIndependentMaintenanceConsole">重新连接</Button>
      </template>
    </AdminPageHeader>
    <section class="space-y-4 rounded-xl border bg-card p-5 sm:p-6" aria-label="选择备份模块">
      <div><h2 class="text-base font-semibold">选择要备份的数据</h2><p class="mt-2 text-sm text-muted-foreground">全部商城数据可以用于整站恢复；也可以分别保存商品、订单、聊天等模块。</p></div>
      <div class="flex flex-wrap gap-4 text-sm">
        <label class="flex items-center gap-2"><input v-model="backupScope" class="accent-admin-selected" type="radio" value="whole" name="admin-backup-scope" :disabled="maintenanceConnected && !maintenanceExportAllowed" @change="sendSelectedBackupScopeToMaintenanceConsole" />全部商城数据</label>
        <label class="flex items-center gap-2"><input v-model="backupScope" class="accent-admin-selected" type="radio" value="modules" name="admin-backup-scope" :disabled="maintenanceConnected && !maintenanceExportAllowed" @change="sendSelectedBackupScopeToMaintenanceConsole" />按模块备份</label>
      </div>
      <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
        <label v-for="option in businessBackupModuleOptions" :key="option.id" class="flex cursor-pointer items-start gap-3 rounded-lg border p-4" :class="{ 'border-admin-selected bg-admin-selected-soft': backupScope === 'whole' || selectedBackupModules.includes(option.id) }">
          <input type="checkbox" class="mt-1 accent-admin-selected" :disabled="maintenanceConnected && !maintenanceExportAllowed" :checked="backupScope === 'whole' || selectedBackupModules.includes(option.id)" @change="toggleBusinessBackupModule(option.id, ($event.target as HTMLInputElement).checked)" />
          <span><strong class="block text-sm">{{ option.label }}</strong><span class="mt-1 block text-xs leading-5 text-muted-foreground">{{ option.description }}</span></span>
        </label>
      </div>
      <p role="status" class="text-sm">{{ backupScope === 'whole' ? '当前范围：全部商城数据与实际文件' : selectedBackupModuleSummary }}</p>
      <div class="flex flex-wrap gap-2"><Button :disabled="!maintenanceConnected || !maintenanceExportAllowed || (backupScope === 'modules' && !selectedBackupModules.length)" @click="openSelectedBackupExportWorkflow">{{ backupScope === 'whole' ? '继续导出全部数据' : '预览所选模块并导出' }}</Button><Button variant="outline" :disabled="!maintenanceConnected || !maintenanceUploadAllowed" @click="openLocalBackupUploadWorkflow">上传本地备份并恢复</Button></div>
      <p v-if="maintenanceOperationBusy" role="status" class="text-sm text-muted-foreground">正在处理备份操作，请等待完成后调整范围。</p>
      <p v-if="!maintenanceConnected" class="text-sm text-muted-foreground">{{ maintenanceConfigured === false ? '本地备份服务尚未配置。模块选择会保留；连接服务后才能生成、下载或上传恢复。' : '连接备份服务后，可以生成备份、下载到电脑或上传本地包。' }}</p>
      <p class="text-xs leading-5 text-muted-foreground">上传后按备份包识别恢复模块，并先验证内容及依赖。整站恢复、各模块支持的恢复范围以验证结果为准。多选模块组合包的正式导入仍待完成。Hostinger 周备份独立管理。</p>
    </section>
    <p class="text-sm text-muted-foreground">整体备份生成期间商城会暂时停止服务。请先安排维护窗口，并单独保管维护凭据和备份解密凭据；原数据库或业务登录不可用时，可直接访问独立恢复入口。</p>
    <p v-if="connectionError" role="alert" class="text-sm text-destructive">{{ connectionError }}</p>
    <p v-if="connecting" role="status" class="rounded-lg border p-4 text-sm text-muted-foreground">正在连接独立备份服务…</p>
    <iframe id="business-backup-maintenance-console" v-if="mountMaintenanceConsole" v-show="!connecting" :key="connectionAttempt" ref="maintenanceConsoleFrame" :src="consolePath" title="独立本地备份与恢复控制台" class="min-h-[1200px] w-full" />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import axios from '@/utils/axios'

const { maintenanceConfigured } = defineProps<{ maintenanceConfigured?: boolean }>()
const businessBackupModuleOptions = [
  { id: 'products', label: '商品与库存', description: '商品详情、多语言、SKU、规格、价格库存及实际图片' },
  { id: 'media', label: '图片与附件', description: '素材原件、派生图片、商品图片和私有附件' },
  { id: 'orders', label: '订单与交易', description: '订单详情、支付退款、发货售后及证据附件' },
  { id: 'customers', label: '客户资料', description: '客户账号、访客、积分及推荐记录' },
  { id: 'customer_service', label: '客户聊天记录', description: '客服会话、全部消息、阅读位置及实际附件' },
  { id: 'configuration', label: '商城业务配置', description: '站点内容、市场税率、交易邮件物流规则及相关图片' },
]
const backupScope = ref<'whole' | 'modules'>('whole')
const selectedBackupModules = ref<string[]>([])
const maintenanceConnected = ref(false)
const maintenanceOperationBusy = ref(false)
const maintenanceExportAllowed = ref(false)
const maintenanceUploadAllowed = ref(false)
const selectedBackupModuleSummary = computed(() => selectedBackupModules.value.length ? '当前范围：' + businessBackupModuleOptions.filter(option => selectedBackupModules.value.includes(option.id)).map(option => option.label).join('、') : '请至少选择一个模块')
function toggleBusinessBackupModule(id: string, checked: boolean) {
  if (backupScope.value === 'whole') selectedBackupModules.value = businessBackupModuleOptions.map(option => option.id)
  backupScope.value = 'modules'
  selectedBackupModules.value = checked ? [...new Set([...selectedBackupModules.value, id])] : selectedBackupModules.value.filter(value => value !== id)
  sendSelectedBackupScopeToMaintenanceConsole()
}
function sendSelectedBackupScopeToMaintenanceConsole() {
  if (!maintenanceConnected.value) return
  maintenanceConsoleFrame.value?.contentWindow?.postMessage({ type: 'local-backup-maintenance-select-scope', mode: backupScope.value, modules: backupScope.value === 'modules' ? [...selectedBackupModules.value].sort() : [] }, window.location.origin)
}
function openSelectedBackupExportWorkflow() {
  if (!maintenanceConnected.value || !maintenanceExportAllowed.value) return
  sendSelectedBackupScopeToMaintenanceConsole()
  maintenanceConsoleFrame.value?.contentWindow?.postMessage({ type: 'local-backup-maintenance-show-workflow', workflow: 'export' }, window.location.origin)
  maintenanceConsoleFrame.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}
function openLocalBackupUploadWorkflow() {
  if (!maintenanceConnected.value || !maintenanceUploadAllowed.value) return
  maintenanceConsoleFrame.value?.contentWindow?.postMessage({ type: 'local-backup-maintenance-show-workflow', workflow: 'upload' }, window.location.origin)
  maintenanceConsoleFrame.value?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

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
  if (errorMessage) {
    maintenanceConnected.value = false
    maintenanceOperationBusy.value = false
    maintenanceExportAllowed.value = false
    maintenanceUploadAllowed.value = false
    mountMaintenanceConsole.value = false
  }
}

async function connectIndependentMaintenanceConsole() {
  if (connecting.value || componentUnmounted) return
  maintenanceConnected.value = false
  maintenanceOperationBusy.value = false
  maintenanceExportAllowed.value = false
  maintenanceUploadAllowed.value = false
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
    maintenanceConnected.value = true
    finishMaintenanceConsoleConnection()
    sendSelectedBackupScopeToMaintenanceConsole()
  } else if (event.data?.type === 'local-backup-maintenance-backup-controls-state') {
    if (typeof event.data.busy !== 'boolean' || typeof event.data.exportAllowed !== 'boolean' || typeof event.data.uploadAllowed !== 'boolean') return
    maintenanceOperationBusy.value = event.data.busy
    maintenanceExportAllowed.value = event.data.exportAllowed
    maintenanceUploadAllowed.value = event.data.uploadAllowed
  } else if (event.data?.type === 'local-backup-maintenance-connection-failed') {
    finishMaintenanceConsoleConnection('独立备份服务未能建立维护会话，请检查服务状态和凭据配置后重新连接。')
  } else if (event.data?.type === 'local-backup-maintenance-disconnected') {
    finishMaintenanceConsoleConnection('维护会话已断开，请点击重新连接。')
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
