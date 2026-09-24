<template>
  <div class="space-y-4">
    <AdminPageHeader title="Outbox 失败事件" description="查看事务通知及其他 Outbox 事件的重试状态；正文 payload 不在列表展示。">
      <template #actions>
        <Button variant="outline" :disabled="loading" @click="load">
          <RefreshCw :class="['size-4', loading ? 'animate-spin' : '']" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

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
          <tr v-for="item in events" :key="item.id" class="border-b border-dashed border-border/50 align-top">
            <td class="px-4 py-3 font-mono">#{{ item.id }}<br><span class="text-muted-foreground">{{ item.event_key }}</span></td>
            <td class="px-4 py-3 font-mono">{{ item.event_type }}</td>
            <td class="px-4 py-3"><span :class="item.status === 'dead_letter' ? 'text-destructive' : 'text-amber-700'">{{ item.status }}</span></td>
            <td class="px-4 py-3">{{ item.attempts }} / {{ item.max_attempts }}</td>
            <td class="max-w-[300px] whitespace-normal break-words px-4 py-3 text-muted-foreground">{{ item.last_error || '-' }}</td>
            <td class="px-4 py-3">{{ formatDate(item.updated_at) }}</td>
            <td class="px-4 py-3 text-right">
              <div class="flex justify-end gap-2">
                <Button variant="outline" size="sm" :disabled="busyId === item.id" @click="retry(item)"><RotateCcw class="size-3.5" />重试</Button>
                <Button variant="ghost" size="sm" :disabled="busyId === item.id" @click="ignore(item)"><Ban class="size-3.5" />忽略</Button>
              </div>
            </td>
          </tr>
          <tr v-if="!loading && events.length === 0"><td colspan="7" class="px-4 py-12 text-center text-muted-foreground">暂无失败事件</td></tr>
        </tbody>
      </table>
    </AdminTablePanel>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Ban, RefreshCw, RotateCcw } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import opsApi, { type OutboxFailureEvent } from '@/api/ops'

const events = ref<OutboxFailureEvent[]>([])
const loading = ref(false)
const busyId = ref<number | null>(null)
const status = ref<'' | 'failed' | 'dead_letter'>('')
const eventType = ref('')

const formatDate = (value: string): string => new Date(value).toLocaleString('zh-CN')

async function load(): Promise<void> {
  loading.value = true
  try {
    events.value = await opsApi.listOutboxFailures({
      ...(status.value ? { status: status.value } : {}),
      ...(eventType.value.trim() ? { event_type: eventType.value.trim() } : {})
    })
  } finally {
    loading.value = false
  }
}

async function retry(item: OutboxFailureEvent): Promise<void> {
  const note = window.prompt('请输入重试原因', 'SMTP 已恢复，人工重新入队')
  if (!note || note.trim().length < 3) return
  busyId.value = item.id
  try { await opsApi.retryOutboxFailure(item.id, note.trim()); await load() } finally { busyId.value = null }
}

async function ignore(item: OutboxFailureEvent): Promise<void> {
  const note = window.prompt('请输入忽略原因', '确认无需再次投递')
  if (!note || note.trim().length < 3) return
  busyId.value = item.id
  try { await opsApi.ignoreOutboxFailure(item.id, note.trim()); await load() } finally { busyId.value = null }
}

onMounted(() => { void load() })
</script>
