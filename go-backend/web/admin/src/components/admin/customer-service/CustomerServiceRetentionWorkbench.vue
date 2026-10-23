<template>
  <div class="flex h-full min-h-0 flex-col gap-3 overflow-auto">
    <AdminPageHeader
      title="客服会话保留运维"
      description="仅管理员可执行；所有操作都需要显式会话 ID 和运维原因"
    >
      <template #actions>
        <Button variant="outline" size="sm" :disabled="loading || busy || confirmOpen" @click="preview">
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
            :disabled="loading || busy || confirmOpen"
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
            :disabled="loading || busy || confirmOpen"
            class="rounded-xl text-xs normal-case tracking-normal"
          />
          <span class="font-medium normal-case tracking-normal text-muted-foreground/70">原因会写入 retention 审计记录。</span>
        </label>
        <div class="flex flex-wrap gap-2 lg:justify-end">
          <Button size="sm" :disabled="loading || busy || confirmOpen" @click="preview">
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
      <CardContent class="space-y-3 p-4 pt-0">
        <div v-if="runtimeError" role="alert" class="text-xs font-bold text-destructive">
          {{ runtimeError }}
          <Button variant="outline" size="sm" :disabled="runtimeLoading" @click="loadRuntimeConfig">重新加载策略</Button>
        </div>
        <p v-else-if="runtimeLoading" class="text-xs text-muted-foreground">正在读取已保存的保留策略……</p>
        <fieldset v-if="appliedRuntimeConfig" :disabled="runtimeLoading || runtimeSaving || busy || confirmOpen" class="grid gap-4 md:grid-cols-5 md:items-end">
          <label class="flex items-center gap-2 text-xs font-bold"><input v-model="runtimeConfig.enabled" type="checkbox" />启用自动 purge</label>
          <label class="grid gap-1 text-xs text-muted-foreground">运行间隔（秒）<input v-model.number="runtimeConfig.interval_seconds" class="rounded-md border bg-background px-2 py-1 text-sm text-foreground" type="number" min="60" max="604800" /></label>
          <label class="grid gap-1 text-xs text-muted-foreground">最小保留（天）<input v-model.number="runtimeConfig.minimum_retention_days" class="rounded-md border bg-background px-2 py-1 text-sm text-foreground" type="number" min="1" max="36500" /></label>
          <label class="grid gap-1 text-xs text-muted-foreground">恢复窗口（天）<input v-model.number="runtimeConfig.recovery_window_days" class="rounded-md border bg-background px-2 py-1 text-sm text-foreground" type="number" min="1" max="3650" /></label>
          <div class="flex gap-2"><label class="grid flex-1 gap-1 text-xs text-muted-foreground">批量<input v-model.number="runtimeConfig.batch_limit" class="rounded-md border bg-background px-2 py-1 text-sm text-foreground" type="number" min="1" max="1000" /></label><Button size="sm" :disabled="runtimeLoading || runtimeSaving || busy || confirmOpen" @click="saveRuntimeConfig">保存</Button></div>
        </fieldset>
      </CardContent>
    </Card>

    <Card v-if="isAdmin" class="shadow-none">
      <CardHeader class="flex-row items-start justify-between gap-3 pb-3">
        <div>
          <CardTitle class="text-sm font-black uppercase tracking-tight">历史会话检索与附件检查</CardTitle>
          <CardDescription class="text-xs">从这里按客户、邮箱、会话编号或消息内容查找历史记录；软删除恢复期内的记录也会保留在清单中。</CardDescription>
        </div>
        <Button variant="outline" size="sm" :disabled="historicalLoading" @click="loadHistoricalRecords">
          <RefreshCw :class="['size-3.5', { 'animate-spin': historicalLoading }]" />
          刷新历史
        </Button>
      </CardHeader>
      <CardContent class="space-y-4 p-4 pt-0">
        <form class="grid gap-3 md:grid-cols-[minmax(0,1fr)_180px_auto]" @submit.prevent="applyHistoricalSearch">
          <label class="grid gap-1 text-xs font-bold text-muted-foreground">
            搜索历史会话
            <div class="relative">
              <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
              <Input v-model="historicalSearch" class="h-9 pl-9" placeholder="客户、邮箱、会话 ID、消息内容" />
            </div>
          </label>
          <label class="grid gap-1 text-xs font-bold text-muted-foreground">
            生命周期
            <Select v-model="historicalStatus">
              <SelectTrigger class="h-9"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="closed">已关闭 / 已解决</SelectItem>
                <SelectItem value="all">全部生命周期</SelectItem>
                <SelectItem value="open">待处理</SelectItem>
                <SelectItem value="in_progress">处理中</SelectItem>
              </SelectContent>
            </Select>
          </label>
          <Button type="submit" class="h-9 self-end rounded-full" :disabled="historicalLoading">
            <Search class="size-3.5" />
            查询
          </Button>
        </form>

        <div class="grid gap-2 rounded-xl border bg-muted/25 p-3 text-xs text-muted-foreground md:grid-cols-3">
          <span>当前最小保留：<strong class="text-foreground">{{ appliedRuntimeConfig ? `${appliedRuntimeConfig.minimum_retention_days} 天` : '策略尚未读取' }}</strong></span>
          <span>软删除恢复窗口：<strong class="text-foreground">{{ appliedRuntimeConfig ? `${appliedRuntimeConfig.recovery_window_days} 天` : '策略尚未读取' }}</strong></span>
          <span>备份：<strong class="text-foreground">需单独配置数据库与附件备份</strong></span>
          <span class="md:col-span-3">会话正文和附件在 purge 前继续保留；purge 后应用没有在线恢复入口。备份用于灾难恢复和经批准的维护恢复，不会被页面上的归档或软删除动作替代。</span>
          <span class="md:col-span-3">附件检查统计旧附件字段的引用与格式异常；新附件按其绑定关系执行保留和清理。</span>
        </div>

        <div v-if="historicalError" class="rounded-lg border border-destructive/20 bg-destructive/5 p-3 text-xs font-bold text-destructive">
          {{ historicalError }}
        </div>
        <div v-else-if="!historicalRecords.length && !historicalLoading" class="flex min-h-28 items-center justify-center rounded-lg border border-dashed px-4 text-center text-xs font-bold text-muted-foreground">
          没有符合条件的历史会话。
        </div>
        <div v-else class="overflow-x-auto rounded-lg border">
          <table class="w-full min-w-[1060px] text-left text-xs">
            <thead class="border-b bg-muted/25 text-[10px] font-black uppercase tracking-wider text-muted-foreground">
              <tr>
                <th class="px-3 py-3">会话 / 客户</th>
                <th class="px-3 py-3">生命周期</th>
                <th class="px-3 py-3">保留动作</th>
                <th class="px-3 py-3">最近活动</th>
                <th class="px-3 py-3">附件检查</th>
                <th class="px-3 py-3">操作</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border/50">
              <tr v-for="record in historicalRecords" :key="record.id" class="align-top">
                <td class="px-3 py-3">
                  <div class="font-black">#{{ record.id }} · {{ record.ticket_number || record.conversation_id || '无编号' }}</div>
                  <div class="mt-1 max-w-[260px] truncate text-muted-foreground">{{ record.customer_name || '匿名客户' }}</div>
                </td>
                <td class="px-3 py-3 font-bold text-muted-foreground">{{ lifecycleLabel(record.retention.lifecycle_status) }}</td>
                <td class="px-3 py-3">
                  <span :class="['inline-flex rounded-full px-2 py-1 text-[10px] font-black', actionTone(record.retention.action)]">
                    {{ retentionActionLabel(record.retention) }}
                  </span>
                  <div class="mt-1 max-w-[240px] text-[11px] text-muted-foreground">{{ retentionTimingLabel(record.retention) }}</div>
                </td>
                <td class="px-3 py-3 whitespace-nowrap text-muted-foreground">{{ formatDate(record.retention.last_activity_at) }}</td>
                <td class="px-3 py-3">
                  <span v-if="record.retention.malformed_attachment_payloads" class="inline-flex items-center gap-1 rounded-full bg-amber-500/10 px-2 py-1 font-black text-amber-700 dark:text-amber-300">
                    <AlertTriangle class="size-3" />
                    {{ record.retention.malformed_attachment_payloads }} 条异常
                  </span>
                  <span v-else class="text-muted-foreground">旧附件字段 {{ record.retention.attachment_reference_count || 0 }} 个引用，未发现格式异常</span>
                </td>
                <td class="px-3 py-3">
                  <Button v-if="!record.retention.soft_deleted_at" variant="outline" size="sm" class="rounded-full" @click="openHistoricalConversationHistory(record)">
                    查看会话
                  </Button>
                  <span v-else class="text-[11px] text-muted-foreground">需维护恢复</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="flex flex-wrap items-center justify-between gap-2 text-xs text-muted-foreground">
          <span>共 {{ historicalTotal }} 条，当前第 {{ historicalPage }} / {{ historicalTotalPages }} 页</span>
          <div class="flex gap-2">
            <Button variant="outline" size="sm" :disabled="historicalLoading || historicalPage <= 1" @click="changeHistoricalPage(historicalPage - 1)">上一页</Button>
            <Button variant="outline" size="sm" :disabled="historicalLoading || historicalPage >= historicalTotalPages" @click="changeHistoricalPage(historicalPage + 1)">下一页</Button>
          </div>
        </div>
      </CardContent>
    </Card>

    <Card class="min-h-0 flex-1 shadow-none">
      <CardHeader class="flex-row items-center justify-between gap-3 pb-3">
        <div>
          <CardTitle class="text-sm font-black uppercase tracking-tight">资格结果</CardTitle>
          <CardDescription class="text-xs">资格检查会重新执行依赖 hold、最小保留期限和已保存的恢复窗口判断。</CardDescription>
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
                <th class="px-4 py-3">附件检查</th>
                <th class="px-4 py-3">原因</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-border/50">
              <tr v-for="item in eligibility" :key="item.ticket_id" class="align-top">
                <td class="px-4 py-3 font-black">#{{ item.ticket_id }}</td>
                <td class="px-4 py-3 font-bold text-muted-foreground">{{ item.lifecycle_status || '—' }}</td>
                <td class="px-4 py-3">
                  <span :class="['inline-flex rounded-full px-2 py-1 text-[10px] font-black', actionTone(item.action)]">
                    {{ retentionActionLabel(item) }}
                  </span>
                </td>
                <td class="px-4 py-3 font-medium text-muted-foreground">
                  <span v-if="item.purge_after" class="inline-flex items-center gap-1">
                    <CalendarClock class="size-3" />
                    {{ formatDate(item.purge_after) }}
                  </span>
                  <span v-else>—</span>
                </td>
                <td class="px-4 py-3">
                  <span v-if="item.malformed_attachment_payloads" class="font-black text-amber-700 dark:text-amber-300">异常 {{ item.malformed_attachment_payloads }} 条</span>
                  <span v-else class="text-muted-foreground">{{ item.attachment_reference_count || 0 }} 个引用</span>
                </td>
                <td class="max-w-[360px] px-4 py-3 font-medium text-muted-foreground">{{ item.reason || '符合当前动作条件' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </CardContent>
    </Card>

    <CustomerServiceReadOnlyConversationHistoryDialog
      v-if="isAdmin"
      :conversation="selectedHistoricalConversation"
      @close="selectedHistoricalConversation = null"
    />

    <AdminConfirmDialog
      v-model:open="confirmOpen"
      :title="pendingAction === 'purge' ? '确认永久清理会话？' : '确认软删除会话？'"
      :description="confirmDescription"
      :confirm-label="pendingAction === 'purge' ? '执行永久清理' : '进入软删除恢复期'"
      destructive
      @confirm="confirmAction"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { isAxiosError } from 'axios'
import { toast } from 'vue-sonner'
import {
  AlertTriangle,
  ArchiveX,
  CalendarClock,
  Clock3,
  RefreshCw,
  Search,
  ScanSearch,
  ShieldAlert,
  Trash2,
} from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import CustomerServiceReadOnlyConversationHistoryDialog from '@/components/admin/customer-service/CustomerServiceReadOnlyConversationHistoryDialog.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import customerServiceApi, { type CustomerServiceRetentionEligibility, type CustomerServiceRetentionRecord } from '@/api/customerService'
import type { CustomerServiceRetentionRuntimeConfig } from '@/api/customerService'
import { useAuthStore } from '@/stores/auth'

type RetentionAction = 'soft-delete' | 'purge'

const authStore = useAuthStore()
const isAdmin = computed(() => authStore.hasRole('admin'))
const selectedHistoricalConversation = ref<CustomerServiceRetentionRecord | null>(null)
watch(isAdmin, (allowed) => {
  if (!allowed) selectedHistoricalConversation.value = null
}, { flush: 'sync' })
const idsInput = ref('')
const reason = ref('')
const eligibility = ref<CustomerServiceRetentionEligibility[]>([])
const loading = ref(false)
const busy = ref(false)
const validationError = ref('')
const confirmOpen = ref(false)
const pendingAction = ref<RetentionAction | null>(null)
const pendingConversationIDs = ref<number[]>([])
const previewConversationIDsKey = ref('')
let eligibilityRequestVersion = 0
let historicalRequestVersion = 0
let disposed = false
const runtimeLoading = ref(false)
const runtimeSaving = ref(false)
const runtimeError = ref('')
const appliedRuntimeConfig = ref<CustomerServiceRetentionRuntimeConfig | null>(null)
const runtimeConfig = ref<CustomerServiceRetentionRuntimeConfig>({ enabled: false, interval_seconds: 86400, minimum_retention_days: 730, recovery_window_days: 30, batch_limit: 100 })
const historicalSearch = ref('')
const historicalStatus = ref('closed')
const appliedHistoricalSearch = ref('')
const appliedHistoricalStatus = ref('closed')
const historicalPage = ref(1)
const historicalTotal = ref(0)
const historicalRecords = ref<CustomerServiceRetentionRecord[]>([])
const historicalLoading = ref(false)
const historicalError = ref('')
const historicalPageSize = 20
const historicalTotalPages = computed(() => Math.max(1, Math.ceil(historicalTotal.value / historicalPageSize)))

const parseIDs = (): number[] => {
  const values = idsInput.value
    .split(/[\s,]+/)
    .map((value) => Number(value.trim()))
    .filter((value) => Number.isInteger(value) && value > 0)
  return [...new Set(values)]
}

const selectedIDs = computed(() => parseIDs())
const selectedConversationIDsKey = computed(() => [...selectedIDs.value].sort((a, b) => a - b).join(','))
const hasInvalidConversationIDs = computed(() => idsInput.value.split(/[\s,]+/).filter(Boolean).some((value) => {
  const id = Number(value)
  return !/^\d+$/.test(value) || !Number.isSafeInteger(id) || id <= 0 || id > 4294967295
}))
const previewMatchesCurrentConversationIDs = computed(() => Boolean(appliedRuntimeConfig.value) &&
  !hasInvalidConversationIDs.value && Boolean(previewConversationIDsKey.value) &&
  previewConversationIDsKey.value === selectedConversationIDsKey.value)
const canSoftDelete = computed(() => previewMatchesCurrentConversationIDs.value && eligibility.value.some((item) => item.eligible && item.action === 'soft_delete'))
const canPurge = computed(() => previewMatchesCurrentConversationIDs.value && eligibility.value.some((item) => item.eligible && item.action === 'purge'))
const confirmDescription = computed(() => {
  return `${pendingConversationIDs.value.length} 个会话将按当前资格结果执行，ID：${pendingConversationIDs.value.join(', ')}。运维原因会写入审计日志。`
})

const invalidateRetentionEligibilityPreview = () => {
  eligibilityRequestVersion++
  eligibility.value = []
  previewConversationIDsKey.value = ''
  loading.value = false
  confirmOpen.value = false
  pendingAction.value = null
  pendingConversationIDs.value = []
}

watch(idsInput, invalidateRetentionEligibilityPreview, { flush: 'sync' })

const validateInput = (): number[] | null => {
  const ids = selectedIDs.value
  if (hasInvalidConversationIDs.value) {
    validationError.value = '会话 ID 必须全部为有效的正整数，请修正后重新预览。'
    return null
  }
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
  invalidateRetentionEligibilityPreview()
  if (hasInvalidConversationIDs.value || !ids.length || ids.length > 100) {
    validationError.value = hasInvalidConversationIDs.value ? '会话 ID 必须全部为有效的正整数，请修正后重新预览。'
      : ids.length > 100 ? '一次最多处理 100 个会话。' : '请输入至少一个有效的会话 ID。'
    return
  }
  const requestVersion = ++eligibilityRequestVersion
  const requestedIDsKey = selectedConversationIDsKey.value
  validationError.value = ''
  loading.value = true
  try {
    const result = await customerServiceApi.evaluateRetention(ids)
    if (disposed || requestVersion !== eligibilityRequestVersion) return
    const responseIDsKey = result.map((item) => item.ticket_id).sort((a, b) => a - b).join(',')
    if (responseIDsKey !== requestedIDsKey) throw new Error('Retention eligibility response does not match the requested conversation IDs')
    eligibility.value = result
    previewConversationIDsKey.value = requestedIDsKey
  } catch (error) {
    if (disposed || requestVersion !== eligibilityRequestVersion) return
    handleError(error, '资格检查失败')
  } finally {
    if (!disposed && requestVersion === eligibilityRequestVersion) loading.value = false
  }
}

const requestAction = (action: RetentionAction) => {
  if (loading.value || busy.value || runtimeSaving.value) return
  const ids = validateInput()
  if (!ids) return
  const eligibleAction = action === 'purge' ? 'purge' : 'soft_delete'
  const targetIDs = eligibility.value.filter((item) => item.eligible && item.action === eligibleAction).map((item) => item.ticket_id)
  if (!previewMatchesCurrentConversationIDs.value || !targetIDs.length) {
    validationError.value = '请先预览资格，并确保至少有一个会话符合当前动作。'
    return
  }
  pendingConversationIDs.value = targetIDs
  pendingAction.value = action
  confirmOpen.value = true
}

const confirmAction = async () => {
  if (busy.value) return
  const action = pendingAction.value
  const ids = validateInput()
  if (!action || !ids) return
  const targetIDs = [...pendingConversationIDs.value]
  if (!previewMatchesCurrentConversationIDs.value || !targetIDs.length || targetIDs.some((id) => !ids.includes(id))) {
    invalidateRetentionEligibilityPreview()
    validationError.value = '会话目标已改变，请重新预览资格后确认。'
    return
  }
  busy.value = true
  try {
    eligibility.value = action === 'purge'
      ? await customerServiceApi.purgeRetention(targetIDs, reason.value.trim())
      : await customerServiceApi.softDeleteRetention(targetIDs, reason.value.trim())
    toast.success(action === 'purge' ? '会话已完成 purge' : '会话已进入软删除恢复期，截止时间见资格结果')
    confirmOpen.value = false
    await Promise.all([preview(), loadHistoricalRecords()])
  } catch (error) {
    invalidateRetentionEligibilityPreview()
    handleError(error, action === 'purge' ? 'purge 执行失败' : '软删除执行失败')
    await loadHistoricalRecords()
  } finally {
    busy.value = false
    pendingAction.value = null
    pendingConversationIDs.value = []
  }
}

const retentionActionLabel = (item: CustomerServiceRetentionEligibility): string => {
  if (item.soft_deleted_at && item.action !== 'purge') return item.reason === 'soft-delete recovery window has not elapsed' ? '恢复期中' : '清理受阻'
  switch (item.action) {
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

const lifecycleLabel = (status?: string): string => {
  switch (status) {
    case 'open': return '待处理'
    case 'in_progress': return '处理中'
    case 'resolved': return '已解决'
    case 'closed': return '已关闭'
    default: return status || '未知'
  }
}

const retentionTimingLabel = (item: CustomerServiceRetentionEligibility): string => {
  if (item.soft_deleted_at && item.reason === 'soft-delete recovery window has not elapsed') return `恢复窗口至 ${formatDate(item.purge_after)}`
  if (item.reason === 'minimum retention age has not elapsed') return `需保留至 ${formatDate(item.eligible_after)}`
  if (item.reason === 'conversation is not resolved or closed') return '会话尚未关闭或解决，继续保留'
  if (item.reason) return item.reason
  if (item.action === 'soft_deleted') return `恢复窗口至 ${formatDate(item.purge_after)}`
  if (item.action === 'purge') return '已过恢复窗口，可人工确认 purge'
  if (item.action === 'soft_delete') return `达到保留期限后可软删除：${formatDate(item.eligible_after)}`
  return `可在 ${formatDate(item.eligible_after)} 后重新检查`
}

const formatHistoricalConversationLoadError = (error: unknown): string => {
  if (!isAxiosError(error)) return '历史会话返回的数据格式异常，请确认前后端版本一致后重试。'
  const status = error.response?.status
  if (status === 404) return '历史会话查询接口不存在（HTTP 404），请更新并重启后端服务后重试。'
  if (status === 401) return '登录已过期，请重新登录后查询历史会话。'
  if (status === 403) return '当前账号没有查询历史会话的权限，请使用管理员账号。'
  if (!error.response) return '无法连接历史会话查询服务，请检查后端服务和网络后重试。'
  const responseMessage = error.response.data?.error || error.response.data?.message
  const detail = typeof responseMessage === 'string' ? responseMessage.trim() : ''
  return `历史会话读取失败（HTTP ${status}）${detail ? `：${detail}` : '，请稍后重试。'}`
}

const loadHistoricalRecords = async () => {
  if (!isAdmin.value || disposed) return
  const requestVersion = ++historicalRequestVersion
  historicalLoading.value = true
  historicalError.value = ''
  try {
    const data = await customerServiceApi.listRetentionRecords({
      page: historicalPage.value,
      page_size: historicalPageSize,
      search: appliedHistoricalSearch.value || undefined,
      status: appliedHistoricalStatus.value === 'all' ? undefined : appliedHistoricalStatus.value,
    })
    if (disposed || requestVersion !== historicalRequestVersion) return
    historicalRecords.value = data.records as CustomerServiceRetentionRecord[]
    historicalTotal.value = data.pagination.total
    if (historicalPage.value > historicalTotalPages.value) {
      historicalPage.value = historicalTotalPages.value
      await loadHistoricalRecords()
    }
  } catch (error) {
    if (disposed || requestVersion !== historicalRequestVersion) return
    historicalRecords.value = []
    historicalTotal.value = 0
    historicalError.value = formatHistoricalConversationLoadError(error)
  } finally {
    if (!disposed && requestVersion === historicalRequestVersion) historicalLoading.value = false
  }
}

const applyHistoricalSearch = () => {
  appliedHistoricalSearch.value = historicalSearch.value.trim()
  appliedHistoricalStatus.value = historicalStatus.value
  historicalPage.value = 1
  void loadHistoricalRecords()
}

const changeHistoricalPage = (page: number) => {
  if (page < 1 || page > historicalTotalPages.value || page === historicalPage.value) return
  historicalPage.value = page
  void loadHistoricalRecords()
}

const openHistoricalConversationHistory = (record: CustomerServiceRetentionRecord) => {
  if (!isAdmin.value || record.retention.soft_deleted_at) return
  selectedHistoricalConversation.value = record
}

const loadRuntimeConfig = async () => {
  if (!isAdmin.value) return
  runtimeLoading.value = true
  runtimeError.value = ''
  try {
    const config = await customerServiceApi.getRetentionConfig()
    if (disposed) return
    runtimeConfig.value = { ...config }
    appliedRuntimeConfig.value = { ...config }
  } catch (error) {
    if (disposed) return
    appliedRuntimeConfig.value = null
    invalidateRetentionEligibilityPreview()
    runtimeError.value = '保留策略读取失败，当前天数未知。重新加载成功前无法保存策略或执行清理。'
  } finally { if (!disposed) runtimeLoading.value = false }
}

const saveRuntimeConfig = async () => {
  if (!appliedRuntimeConfig.value || runtimeLoading.value || runtimeSaving.value || busy.value) return
  runtimeSaving.value = true
  try {
    runtimeConfig.value = await customerServiceApi.updateRetentionConfig(runtimeConfig.value)
    appliedRuntimeConfig.value = { ...runtimeConfig.value }
    invalidateRetentionEligibilityPreview()
    toast.success('自动清理策略已保存')
    await loadHistoricalRecords()
  } catch (error) { handleError(error, '保存自动清理策略失败') } finally { runtimeSaving.value = false }
}

onMounted(async () => {
  await loadRuntimeConfig()
  if (!disposed) await loadHistoricalRecords()
})

onBeforeUnmount(() => {
  disposed = true
  eligibilityRequestVersion++
  historicalRequestVersion++
})
</script>
