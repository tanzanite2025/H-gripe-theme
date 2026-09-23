<template>
  <div class="flex h-full min-h-0 flex-col gap-3 overflow-auto">
    <AdminPageHeader
      title="客服会话保留运维"
      description="仅管理员可执行；所有操作都需要显式会话 ID 和运维原因"
    >
      <template #actions>
        <Button variant="outline" size="sm" :disabled="loading" @click="preview">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
          刷新资格
        </Button>
      </template>
    </AdminPageHeader>

    <Card v-if="!isAdmin" class="border-destructive/30 bg-destructive/5 shadow-none">
      <CardContent class="flex items-start gap-3 p-4 text-sm font-bold text-destructive">
        <ShieldAlert class="mt-0.5 size-4 shrink-0" />
        当前账号不是管理员，后端不会接受 retention 运维命令。
      </CardContent>
    </Card>

    <Card v-else class="shadow-none">
      <CardHeader class="pb-3">
        <CardTitle class="text-sm font-black uppercase tracking-tight">运维目标</CardTitle>
        <CardDescription class="text-xs">先预览资格，再确认软删除或恢复期结束后的 purge。</CardDescription>
      </CardHeader>
      <CardContent class="grid gap-4 p-4 pt-0 lg:grid-cols-[1fr_1fr_auto] lg:items-end">
        <label class="grid gap-1.5 text-[10px] font-black uppercase tracking-wider text-muted-foreground">
          会话 ID 列表
          <Textarea
            v-model="idsInput"
            rows="2"
            placeholder="例如：101, 102, 103"
            :disabled="loading || busy"
            class="rounded-xl text-xs normal-case tracking-normal"
          />
          <span class="font-medium normal-case tracking-normal text-muted-foreground/70">最多 100 个，支持逗号、空格或换行分隔。</span>
        </label>
        <label class="grid gap-1.5 text-[10px] font-black uppercase tracking-wider text-muted-foreground">
          运维原因
          <Textarea
            v-model="reason"
            rows="2"
            maxlength="500"
            placeholder="例如：已完成留存策略审批"
            :disabled="loading || busy"
            class="rounded-xl text-xs normal-case tracking-normal"
          />
          <span class="font-medium normal-case tracking-normal text-muted-foreground/70">原因会写入 retention 审计记录。</span>
        </label>
        <div class="flex flex-wrap gap-2 lg:justify-end">
          <Button size="sm" :disabled="loading || busy" @click="preview">
            <ScanSearch class="size-3.5" />
            预览资格
          </Button>
          <Button
            size="sm"
            variant="outline"
            :disabled="loading || busy || !canSoftDelete"
            @click="requestAction('soft-delete')"
          >
            <ArchiveX class="size-3.5" />
            软删除
          </Button>
          <Button
            size="sm"
            variant="destructive"
            :disabled="loading || busy || !canPurge"
            @click="requestAction('purge')"
          >
            <Trash2 class="size-3.5" />
            执行 purge
          </Button>
        </div>
      </CardContent>
      <div v-if="validationError" class="border-t border-destructive/15 px-4 py-3 text-xs font-bold text-destructive">
        {{ validationError }}
      </div>
    </Card>

    <Card v-if="isAdmin" class="shadow-none">
      <CardHeader class="pb-3">
        <CardTitle class="text-sm font-black uppercase tracking-tight">自动清理策略</CardTitle>
        <CardDescription class="text-xs">修改后无需重启服务；关闭时保留数据，不会自动 purge。</CardDescription>
      </CardHeader>
      <CardContent class="grid gap-4 p-4 pt-0 md:grid-cols-5 md:items-end">
        <label class="flex items-center gap-2 text-xs font-bold"><input v-model="runtimeConfig.enabled" type="checkbox" />启用自动 purge</label>
        <label class="grid gap-1 text-xs text-muted-foreground">运行间隔（秒）<input v-model.number="runtimeConfig.interval_seconds" class="rounded-md border bg-background px-2 py-1 text-sm text-foreground" type="number" min="60" max="604800" /></label>
        <label class="grid gap-1 text-xs text-muted-foreground">最小保留（天）<input v-model.number="runtimeConfig.minimum_retention_days" class="rounded-md border bg-background px-2 py-1 text-sm text-foreground" type="number" min="1" max="36500" /></label>
        <label class="grid gap-1 text-xs text-muted-foreground">恢复窗口（天）<input v-model.number="runtimeConfig.recovery_window_days" class="rounded-md border bg-background px-2 py-1 text-sm text-foreground" type="number" min="1" max="3650" /></label>
        <div class="flex gap-2"><label class="grid flex-1 gap-1 text-xs text-muted-foreground">批量<input v-model.number="runtimeConfig.batch_limit" class="rounded-md border bg-background px-2 py-1 text-sm text-foreground" type="number" min="1" max="1000" /></label><Button size="sm" :disabled="runtimeLoading || runtimeSaving" @click="saveRuntimeConfig">保存</Button></div>
      </CardContent>
    </Card>

    <Card class="min-h-0 flex-1 shadow-none">
      <CardHeader class="flex-row items-center justify-between gap-3 pb-3">
        <div>
          <CardTitle class="text-sm font-black uppercase tracking-tight">资格结果</CardTitle>
          <CardDescription class="text-xs">资格检查会重新执行依赖 hold、留存年限和 30 天恢复窗口判断。</CardDescription>
        </div>
        <span class="rounded-full bg-muted px-2.5 py-1 text-[10px] font-black text-muted-foreground">{{ eligibility.length }} 条</span>
      </CardHeader>
      <CardContent class="p-0">
        <div v-if="!eligibility.length" class="flex min-h-40 flex-col items-center justify-center gap-2 px-4 text-center text-muted-foreground">
          <Clock3 class="size-5 opacity-50" />
          <p class="text-xs font-bold">输入会话 ID 后预览 retention 资格。</p>
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full min-w-[760px] text-left text-xs">
            <thead class="border-y border-border/60 bg-muted/25 text-[10px] font-black uppercase tracking-wider text-muted-foreground">
              <tr>
                <th class="px-4 py-3">会话</th>
                <th class="px-4 py-3">生命周期</th>
                <th class="px-4 py-3">动作</th>
                <th class="px-4 py-3">恢复期</th>
                <th class="px-4 py-3">原因</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border/50">
              <tr v-for="item in eligibility" :key="item.ticket_id" class="align-top">
                <td class="px-4 py-3 font-black">#{{ item.ticket_id }}</td>
                <td class="px-4 py-3 font-bold text-muted-foreground">{{ item.lifecycle_status || '—' }}</td>
                <td class="px-4 py-3">
                  <span :class="['inline-flex rounded-full px-2 py-1 text-[10px] font-black', actionTone(item.action)]">
                    {{ actionLabel(item.action) }}
                  </span>
                </td>
                <td class="px-4 py-3 font-medium text-muted-foreground">
                  <span v-if="item.purge_after" class="inline-flex items-center gap-1">
                    <CalendarClock class="size-3" />
                    {{ formatDate(item.purge_after) }}
                  </span>
                  <span v-else>—</span>
                </td>
                <td class="max-w-[360px] px-4 py-3 font-medium text-muted-foreground">{{ item.reason || '符合当前动作条件' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>

    <AdminConfirmDialog
      v-model:open="confirmOpen"
      :title="pendingAction === 'purge' ? '确认永久清理会话？' : '确认软删除会话？'"
      :description="confirmDescription"
      :confirm-label="pendingAction === 'purge' ? '执行永久清理' : '进入 30 天恢复期'"
      destructive
      @confirm="confirmAction"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { isAxiosError } from 'axios'
import { toast } from 'vue-sonner'
import {
  ArchiveX,
  CalendarClock,
  Clock3,
  RefreshCw,
  ScanSearch,
  ShieldAlert,
  Trash2,
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import customerServiceApi, { type CustomerServiceRetentionEligibility } from '@/api/customerService'
import type { CustomerServiceRetentionRuntimeConfig } from '@/api/customerService'
import { useAuthStore } from '@/stores/auth'

type RetentionAction = 'soft-delete' | 'purge'

const authStore = useAuthStore()
const isAdmin = computed(() => authStore.hasRole('admin'))
const idsInput = ref('')
const reason = ref('')
const eligibility = ref<CustomerServiceRetentionEligibility[]>([])
const loading = ref(false)
const busy = ref(false)
const validationError = ref('')
const confirmOpen = ref(false)
const pendingAction = ref<RetentionAction | null>(null)
const runtimeLoading = ref(false)
const runtimeSaving = ref(false)
const runtimeConfig = ref<CustomerServiceRetentionRuntimeConfig>({ enabled: false, interval_seconds: 86400, minimum_retention_days: 730, recovery_window_days: 30, batch_limit: 100 })

const parseIDs = (): number[] => {
  const values = idsInput.value
    .split(/[\s,]+/)
    .map((value) => Number(value.trim()))
    .filter((value) => Number.isInteger(value) && value > 0)
  return [...new Set(values)]
}

const selectedIDs = computed(() => parseIDs())
const canSoftDelete = computed(() => eligibility.value.some((item) => item.action === 'soft_delete'))
const canPurge = computed(() => eligibility.value.some((item) => item.action === 'purge'))
const confirmDescription = computed(() => {
  const count = pendingAction.value === 'purge'
    ? eligibility.value.filter((item) => item.action === 'purge').length
    : eligibility.value.filter((item) => item.action === 'soft_delete').length
  return `${count} 个会话将按当前资格结果执行。运维原因会写入审计日志。`
})

const validateInput = (): number[] | null => {
  const ids = selectedIDs.value
  if (!ids.length) {
    validationError.value = '请输入至少一个有效的会话 ID。'
    return null
  }
  if (ids.length > 100) {
    validationError.value = '一次最多处理 100 个会话。'
    return null
  }
  if (!reason.value.trim()) {
    validationError.value = '执行软删除或 purge 前必须填写运维原因。'
    return null
  }
  if (reason.value.trim().length > 500) {
    validationError.value = '运维原因最多 500 个字符。'
    return null
  }
  validationError.value = ''
  return ids
}

const handleError = (error: unknown, fallback: string) => {
  const responseMessage = isAxiosError(error)
    ? String(error.response?.data?.error || error.response?.data?.message || '').trim()
    : ''
  toast.error(responseMessage || fallback)
}

const preview = async () => {
  const ids = selectedIDs.value
  if (!ids.length || ids.length > 100) {
    validationError.value = ids.length > 100 ? '一次最多处理 100 个会话。' : '请输入至少一个有效的会话 ID。'
    return
  }
  validationError.value = ''
  loading.value = true
  try {
    eligibility.value = await customerServiceApi.evaluateRetention(ids)
  } catch (error) {
    handleError(error, '资格检查失败')
  } finally {
    loading.value = false
  }
}

const requestAction = (action: RetentionAction) => {
  const ids = validateInput()
  if (!ids) return
  const eligibleAction = action === 'purge' ? 'purge' : 'soft_delete'
  if (!eligibility.value.some((item) => item.action === eligibleAction)) {
    validationError.value = '请先预览资格，并确保至少有一个会话符合当前动作。'
    return
  }
  pendingAction.value = action
  confirmOpen.value = true
}

const confirmAction = async () => {
  const action = pendingAction.value
  const ids = validateInput()
  if (!action || !ids) return
  const eligibleAction = action === 'purge' ? 'purge' : 'soft_delete'
  const targetIDs = eligibility.value.filter((item) => item.action === eligibleAction).map((item) => item.ticket_id)
  busy.value = true
  try {
    eligibility.value = action === 'purge'
      ? await customerServiceApi.purgeRetention(targetIDs, reason.value.trim())
      : await customerServiceApi.softDeleteRetention(targetIDs, reason.value.trim())
    toast.success(action === 'purge' ? '会话已完成 purge' : '会话已进入 30 天恢复期')
    confirmOpen.value = false
    await preview()
  } catch (error) {
    handleError(error, action === 'purge' ? 'purge 执行失败' : '软删除执行失败')
  } finally {
    busy.value = false
    pendingAction.value = null
  }
}

const actionLabel = (action: string): string => {
  switch (action) {
    case 'soft_delete': return '可软删除'
    case 'soft_deleted': return '恢复期中'
    case 'purge': return '可 purge'
    default: return '不可操作'
  }
}

const actionTone = (action: string): string => {
  switch (action) {
    case 'purge': return 'bg-destructive/10 text-destructive'
    case 'soft_delete': return 'bg-amber-500/10 text-amber-700 dark:text-amber-300'
    case 'soft_deleted': return 'bg-blue-500/10 text-blue-700 dark:text-blue-300'
    default: return 'bg-muted text-muted-foreground'
  }
}

const formatDate = (value?: string | null): string => {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(date)
}

const loadRuntimeConfig = async () => {
  if (!isAdmin.value) return
  runtimeLoading.value = true
  try { runtimeConfig.value = await customerServiceApi.getRetentionConfig() } catch (error) { handleError(error, '加载自动清理策略失败') } finally { runtimeLoading.value = false }
}

const saveRuntimeConfig = async () => {
  runtimeSaving.value = true
  try { runtimeConfig.value = await customerServiceApi.updateRetentionConfig(runtimeConfig.value); toast.success('自动清理策略已保存') } catch (error) { handleError(error, '保存自动清理策略失败') } finally { runtimeSaving.value = false }
}

onMounted(loadRuntimeConfig)
</script>
