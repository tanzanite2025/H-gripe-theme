<template>
  <div class="space-y-4">
    <AdminPageHeader title="Outbox 失败事件" description="查看事务通知及其他 Outbox 事件的重试状态；正文 payload 不在列表展示。">
      <template #actions>
        <Button variant="ghost" @click="router.push('/')"><ArrowLeft class="size-4" />返回仪表板</Button>
        <Button variant="outline" :disabled="loading" @click="load">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <p v-if="loadError" role="alert" class="rounded-2xl border border-dashed border-amber-500/40 bg-amber-500/10 p-4 text-sm text-amber-700 dark:text-amber-400">失败事件加载失败，请刷新重试。</p>
    <p v-else-if="route.query.event_id && selectedOutboxEventId === null" role="status" class="rounded-2xl border border-dashed border-amber-500/40 bg-amber-500/10 p-4 text-sm text-amber-700 dark:text-amber-400">事件编号格式无效，请从业务跟进提醒重新打开。下方仍可查看当前失败事件。</p>
    <p v-else-if="selectedEventMissing" role="status" class="rounded-2xl border border-dashed border-border/80 bg-muted/10 p-4 text-sm text-muted-foreground">当前失败队列中无法找到事件 #{{ selectedOutboxEventId }}，可能已离开清单、已删除或编号有误。请核查实际事件与业务记录；找不到事件不能证明投递成功。</p>

    <section class="flex flex-wrap items-end gap-3 rounded-2xl border border-dashed border-border/80 bg-muted/10 p-4">
      <label class="min-w-44 space-y-1 text-xs font-bold">
        <span>状态</span>
        <select v-model="status" class="h-9 w-full rounded-xl bg-muted/50 px-3 text-xs font-bold">
          <option value="">失败与死信</option>
          <option value="failed">失败</option>
          <option value="dead_letter">死信</option>
        </select>
      </label>
      <label class="min-w-64 space-y-1 text-xs font-bold">
        <span>事件类型</span>
        <Input v-model="eventType" placeholder="例如 order.payment_succeeded" @keyup.enter="load" />
      </label>
      <Button :disabled="loading" @click="load">应用筛选</Button>
    </section>

    <AdminTablePanel :loading="loading">
      <table class="w-full min-w-[980px] text-left text-xs">
        <thead class="border-b border-dashed border-border/70 text-muted-foreground">
          <tr>
            <th class="px-4 py-3">事件</th>
            <th class="px-4 py-3">类型</th>
            <th class="px-4 py-3">状态</th>
            <th class="px-4 py-3">尝试</th>
            <th class="px-4 py-3">最近错误</th>
            <th class="px-4 py-3">更新时间</th>
            <th class="px-4 py-3 text-right">操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in events" :key="item.id" class="border-b border-dashed border-border/50 align-top" :class="item.id === selectedOutboxEventId ? 'bg-primary/10' : ''">
            <td class="px-4 py-3 font-mono">#{{ item.id }}<br><span class="text-muted-foreground">{{ item.event_key }}</span></td>
            <td class="px-4 py-3 font-mono">{{ item.event_type }}</td>
            <td class="px-4 py-3"><span :class="item.status === 'dead_letter' ? 'text-destructive' : 'text-amber-700'">{{ item.status }}</span></td>
            <td class="px-4 py-3">{{ item.attempts }} / {{ item.max_attempts }}</td>
            <td class="max-w-[300px] whitespace-normal break-words px-4 py-3 text-muted-foreground">{{ item.last_error || '-' }}</td>
            <td class="px-4 py-3">{{ formatDate(item.updated_at) }}</td>
            <td class="px-4 py-3 text-right">
              <div v-if="authStore.hasPermission('system:manage')" class="flex justify-end gap-2">
                <Button variant="outline" size="sm" :disabled="busyId === item.id" @click="retry(item)"><RotateCcw class="size-3.5" />重试</Button>
                <Button variant="ghost" size="sm" :disabled="busyId === item.id" @click="ignore(item)"><Ban class="size-3.5" />忽略</Button>
              </div>
              <span v-else class="text-muted-foreground">仅查看</span>
            </td>
          </tr>
          <tr v-if="!loading && !loadError && events.length === 0"><td colspan="7" class="px-4 py-12 text-center text-muted-foreground">暂无失败事件</td></tr>
        </tbody>
      </table>
    </AdminTablePanel>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { isAxiosError } from 'axios'
import { toast } from 'vue-sonner'
import { ArrowLeft, Ban, RefreshCw, RotateCcw } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import opsApi, { type OutboxFailureEvent } from '@/api/ops'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const events = ref<OutboxFailureEvent[]>([])
const loading = ref(false)
const busyId = ref<number | null>(null)
const status = ref<'' | 'failed' | 'dead_letter'>('')
const eventType = ref('')
const loadError = ref(false)
const selectedEventMissing = ref(false)
let outboxFailureRequestSequence = 0
let outboxFailurePageUnmounted = false
const selectedOutboxEventId = computed(() => {
  const queryValue = route.query.event_id
  if (typeof queryValue !== 'string' || !/^[1-9]\d*$/.test(queryValue)) return null
  const eventId = Number(queryValue)
  return Number.isSafeInteger(eventId) && eventId <= 4294967295 ? eventId : null
})

const formatDate = (value: string): string => new Date(value).toLocaleString('zh-CN')

async function load(): Promise<void> {
  if (outboxFailurePageUnmounted) return
  const currentRequest = ++outboxFailureRequestSequence
  const selectedEventId = selectedOutboxEventId.value
  loading.value = true
  loadError.value = false
  selectedEventMissing.value = false
  try {
    const loadedEvents = await opsApi.listOutboxFailures({
      ...(status.value ? { status: status.value } : {}),
      ...(eventType.value.trim() ? { event_type: eventType.value.trim() } : {})
    })
    if (outboxFailurePageUnmounted || currentRequest !== outboxFailureRequestSequence) return
    if (selectedEventId !== null && !loadedEvents.some((item) => item.id === selectedEventId)) {
      try {
        const selectedEvent = await opsApi.getOutboxFailure(selectedEventId, true)
        if (outboxFailurePageUnmounted || currentRequest !== outboxFailureRequestSequence) return
        loadedEvents.unshift(selectedEvent)
      } catch (error) {
        if (outboxFailurePageUnmounted || currentRequest !== outboxFailureRequestSequence) return
        if (isAxiosError(error) && error.response?.status === 404) {
          selectedEventMissing.value = true
        } else {
          throw error
        }
      }
    }
    if (selectedEventId !== null) {
      loadedEvents.sort((firstEvent, secondEvent) => Number(secondEvent.id === selectedEventId) - Number(firstEvent.id === selectedEventId))
    }
    events.value = loadedEvents
  } catch {
    if (outboxFailurePageUnmounted || currentRequest !== outboxFailureRequestSequence) return
    events.value = []
    loadError.value = true
  } finally {
    if (!outboxFailurePageUnmounted && currentRequest === outboxFailureRequestSequence) loading.value = false
  }
}

async function retry(item: OutboxFailureEvent): Promise<void> {
  const note = window.prompt('请输入重试原因', 'SMTP 已恢复，人工重新入队')
  if (!note || note.trim().length < 3) return
  busyId.value = item.id
  try {
    await opsApi.retryOutboxFailure(item.id, note.trim())
    toast.success('事件已重新入队，投递结果需等待处理')
    await load()
  } catch {
    // The API interceptor reports the failed mutation; keep the event visible.
  } finally { busyId.value = null }
}

async function ignore(item: OutboxFailureEvent): Promise<void> {
  const note = window.prompt('请输入忽略原因', '确认无需再次投递')
  if (!note || note.trim().length < 3) return
  busyId.value = item.id
  try {
    await opsApi.ignoreOutboxFailure(item.id, note.trim())
    toast.success('事件已忽略')
    await load()
  } catch {
    // The API interceptor reports the failed mutation; keep the event visible.
  } finally { busyId.value = null }
}

watch(selectedOutboxEventId, () => { void load() }, { immediate: true })
onBeforeUnmount(() => {
  outboxFailurePageUnmounted = true
  outboxFailureRequestSequence += 1
})
</script>
