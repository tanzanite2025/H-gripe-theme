<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <DialogHeader>
        <DialogTitle>日志详情</DialogTitle>
        <DialogDescription v-if="currentLog">审计日志 #{{ currentLog.id }}</DialogDescription>
      </DialogHeader>

      <div v-if="detailLoading" class="flex h-52 items-center justify-center">
        <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载日志详情" />
      </div>
      <div v-else-if="currentLog" class="space-y-6">
        <dl class="grid overflow-hidden rounded-lg border sm:grid-cols-2">
          <DetailItem label="用户">{{ currentLog.username || '-' }}（ID {{ currentLog.user_id || '-' }}）</DetailItem>
          <DetailItem label="状态"><AdminStatusBadge :tone="currentLog.status === 'success' ? 'green' : 'coral'">{{ currentLog.status === 'success' ? '成功' : '失败' }}</AdminStatusBadge></DetailItem>
          <DetailItem label="操作"><AdminStatusBadge :tone="actionTone(currentLog.action)">{{ actionName(currentLog.action) }}</AdminStatusBadge></DetailItem>
          <DetailItem label="资源">{{ resourceName(currentLog.resource) }} / {{ currentLog.resource_id || '-' }}</DetailItem>
          <DetailItem label="请求方法"><AdminStatusBadge :tone="methodTone(currentLog.method)">{{ currentLog.method || '-' }}</AdminStatusBadge></DetailItem>
          <DetailItem label="耗时">{{ currentLog.duration || 0 }} ms</DetailItem>
 <DetailItem label="IP 地址"><span class="font-mono text-xs">{{ currentLog.ip_address || '-'}}</span></DetailItem>
          <DetailItem label="时间">{{ formatDate(currentLog.created_at) }}</DetailItem>
 <DetailItem label="请求路径" class="sm:col-span-2"><span class="break-all font-mono text-xs">{{ currentLog.path || '-'}}</span></DetailItem>
 <DetailItem label="User Agent" class="sm:col-span-2"><span class="break-all text-xs">{{ currentLog.user_agent || '-'}}</span></DetailItem>
        </dl>

        <Alert v-if="currentLog.error_message" variant="destructive">
          <CircleAlert class="size-4" />
          <AlertTitle>错误信息</AlertTitle>
          <AlertDescription class="whitespace-pre-wrap break-words">{{ currentLog.error_message }}</AlertDescription>
        </Alert>

        <JsonSection v-if="currentLog.changes" title="变更内容" :value="currentLog.changes" />
        <div v-if="currentLog.old_value || currentLog.new_value" class="grid gap-4 md:grid-cols-2">
          <JsonSection v-if="currentLog.old_value" title="变更前" :value="currentLog.old_value" />
          <JsonSection v-if="currentLog.new_value" title="变更后" :value="currentLog.new_value" />
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { defineComponent, h } from 'vue'
import type { PropType } from 'vue'
import { CircleAlert, LoaderCircle } from '@lucide/vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import type {
  AuditLogDateFormatter,
  AuditLogJsonValue,
  AuditLogLabelResolver,
  AuditLogRecord,
  AuditLogToneResolver
} from './auditLogTypes'

withDefaults(defineProps<{
  open?: boolean
  detailLoading?: boolean
  currentLog?: AuditLogRecord | null
  actionName: AuditLogLabelResolver
  actionTone: AuditLogToneResolver
  resourceName: AuditLogLabelResolver
  methodTone: AuditLogToneResolver
  formatDate: AuditLogDateFormatter
}>(), {
  open: false,
  detailLoading: false,
  currentLog: null
})

const emit = defineEmits<{
  (event: 'update:open', value: boolean): void
}>()

const DetailItem = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: [String, Number], default: '' }
  },
  setup(props, { slots }) {
 return () => h('div', { class: 'space-y-1'}, [
 h('dt', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block'}, props.label),
 h('dd', { class: 'text-xs font-bold'}, slots.default ? slots.default() : (props.value || '-'))
    ])
  }
})

const JsonSection = defineComponent({
  props: {
    title: { type: String, required: true },
    value: { type: null as unknown as PropType<AuditLogJsonValue>, required: true }
  },
  setup(props) {
 return () => h('section', { class: 'min-w-0 space-y-2'}, [
 h('h3', { class: 'text-sm font-black tracking-tighter uppercase text-foreground'}, props.title),
 h('pre', { class: 'max-h-80 overflow-auto rounded-lg border border-dashed bg-muted/40 p-3 font-mono text-xs leading-5 whitespace-pre-wrap break-words'}, formatJSON(props.value))
    ])
  }
})

const formatJSON = (value: AuditLogJsonValue): string => {
  if (typeof value !== 'string') return JSON.stringify(value, null, 2)
  try { return JSON.stringify(JSON.parse(value), null, 2) } catch { return value }
}
</script>
