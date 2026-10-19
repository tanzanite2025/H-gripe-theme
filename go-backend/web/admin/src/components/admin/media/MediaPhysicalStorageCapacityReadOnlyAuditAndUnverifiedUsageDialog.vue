<template>
  <Dialog :open="audit.open.value" @update:open="audit.setPhysicalStorageCapacityDialogOpen">
    <DialogContent size="lg" @open-auto-focus.prevent>
      <DialogHeader>
        <DialogTitle>存储容量核对</DialogTitle>
        <DialogDescription>核对当前配置的公有、私有及 Logo 存储中的文件字节。</DialogDescription>
      </DialogHeader>
      <div class="max-h-[65vh] space-y-4 overflow-y-auto text-xs">
        <p v-if="audit.loading.value" role="status">正在核对，关闭窗口可取消本次请求。</p>
        <p v-if="audit.error.value" role="alert" class="text-destructive">{{ audit.error.value }}</p>
        <template v-if="audit.report.value">
          <p :class="audit.report.value.physical_scan_complete ? 'text-muted-foreground' : 'text-amber-700'">
            {{ audit.report.value.physical_scan_complete ? '已完成当前存储扫描' : '扫描未完成，以下为已观察到的部分容量' }}
            · {{ audit.report.value.locations_scanned }} 个空间完成，{{ audit.report.value.locations_incomplete }} 个未完成
          </p>
          <dl class="grid grid-cols-[1fr_auto] gap-x-4 gap-y-2">
            <template v-for="row in observedCapacityRows" :key="row.label">
              <dt>{{ row.label }}</dt>
              <dd class="text-right" :title="`${row.metric.bytes} 字节`">{{ formatPhysicalCapacityBytes(row.metric.bytes) }} · {{ row.metric.objects }} 个</dd>
            </template>
          </dl>
          <p class="text-muted-foreground">原图、衍生图、仅有清理收据和未匹配登记四项合计为已观察容量。同一对象复用只计一次；未匹配登记可能包含专用文件或历史文件，仍需核实使用位置。</p>
          <div class="space-y-2 rounded-lg border p-3">
            <p>其中，带清理收据的对象：{{ formatPhysicalCapacityBytes(audit.report.value.with_cleanup_receipt.bytes) }}（{{ audit.report.value.with_cleanup_receipt.objects }} 个）</p>
            <p>其中，带未消费上传收据的对象：{{ formatPhysicalCapacityBytes(audit.report.value.with_upload_receipt.bytes) }}（{{ audit.report.value.with_upload_receipt.objects }} 个）</p>
            <p>其中，含软删登记的对象：{{ formatPhysicalCapacityBytes(audit.report.value.with_soft_deleted_registration.bytes) }}（{{ audit.report.value.with_soft_deleted_registration.objects }} 个）</p>
            <p>未观察到的已知上传计划：{{ formatPhysicalCapacityBytes(audit.report.value.unobserved_upload_reservations.bytes) }}（{{ audit.report.value.unobserved_upload_reservations.objects }} 个）</p>
            <p>预计大小未知的上传计划：{{ audit.report.value.unknown_upload_reservation_objects }} 个</p>
            <p class="text-muted-foreground">上传计划单列，不与文件字节重复相加。扫描未完成时，尚未观察到的计划也可能已经落盘。</p>
          </div>
          <div class="space-y-2 rounded-lg border p-3">
            <p>已完成扫描的空间中未观察到登记对象：{{ audit.report.value.registered_not_observed.objects }} 个</p>
            <p>未能核实登记对象是否存在：{{ audit.report.value.registered_absence_unverified }} 个</p>
            <p>登记与观察大小不一致：{{ audit.report.value.registered_size_mismatches }} 个；登记大小互相冲突：{{ audit.report.value.conflicting_recorded_sizes }} 个</p>
            <p>登记大小未知：{{ audit.report.value.registered_unknown_recorded_size_objects }} 个</p>
            <p>未能确定存储位置的记录：{{ audit.report.value.unresolved_record_locations }} 条</p>
            <p v-for="message in capacityAuditIssueMessages" :key="message" class="text-amber-700">{{ message }}</p>
          </div>
          <p class="text-muted-foreground">核对区间：{{ new Date(audit.report.value.started_at).toLocaleString() }} — {{ new Date(audit.report.value.completed_at).toLocaleString() }}。上传或清理可能在此期间发生，结果为观察值。</p>
        </template>
        <p class="text-muted-foreground">统计当前对象内容字节，版本历史、未完成的分片上传和磁盘分配空间需另行核对。历史存储位置尚未纳入。此结果用于容量核查，账户上传额度仍按现有规则执行。</p>
      </div>
      <DialogFooter>
        <Button variant="outline" :disabled="audit.loading.value" @click="audit.runPhysicalStorageCapacityAudit">重新核对</Button>
        <Button variant="outline" @click="audit.setPhysicalStorageCapacityDialogOpen(false)">关闭</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import type { useReadOnlyMediaPhysicalStorageCapacityAuditWithCancellationAndPermissionRevocation } from '@/composables/media/useReadOnlyMediaPhysicalStorageCapacityAuditWithCancellationAndPermissionRevocation'

const props = defineProps<{ audit: ReturnType<typeof useReadOnlyMediaPhysicalStorageCapacityAuditWithCancellationAndPermissionRevocation> }>()
const capacityAuditIssueMessages = computed(() => {
  const messages: Record<string, string> = {
    provider_metadata_listing_unavailable: '部分存储暂不支持容量扫描。',
    storage_location_identity_unavailable: '部分存储位置未能确定。',
    storage_metadata_scan_incomplete: '部分空间扫描失败或未完成，容量仅为已观察到的部分。',
    duplicate_object_in_provider_listing: '存储返回了重复对象，本次扫描未能完整核实。',
    object_changed_during_scan: '扫描期间有对象大小变化，请稍后重新核对。',
    object_observation_limit_reached: '已达到单次扫描上限，容量结果尚不完整。',
    record_object_limit_reached: '已达到登记对象核对上限。',
    audit_cancelled_or_timed_out: '核对已取消或超时。',
    cleanup_receipt_metadata_invalid: '部分清理收据资料无法识别。',
    cleanup_receipt_has_no_exact_object_identity: '部分清理收据没有精确文件位置。',
    dedicated_logo_location_unavailable: '独立 Logo 存储位置未能确定。',
    draft_upload_metadata_invalid: '部分草稿上传计划无法识别。',
    some_record_locations_unverified: '部分记录的存储位置尚未核实。',
    some_upload_reservation_sizes_unknown: '部分历史上传、衍生图或复制计划的预计大小未知。',
    object_has_original_and_derivative_registration: '有对象同时登记为原图和衍生图，已按原图计入一次。',
    registration_location_or_key_conflict: '部分登记的文件键与访问地址不一致，或无法确认其属于当前存储位置。',
  }
  return [...new Set((props.audit.report.value?.issues ?? []).map((issue) => messages[issue] ?? '有待核实的存储资料。'))]
})
const observedCapacityRows = computed(() => {
  const report = props.audit.report.value
  if (!report) return []
  return [
    { label: '已观察文件总计', metric: report.observed },
    { label: '登记原图', metric: report.registered_originals },
    { label: '登记衍生图', metric: report.registered_derivatives },
    { label: '仅有清理收据', metric: report.cleanup_receipt_only },
    { label: '未匹配媒体登记', metric: report.unmatched_registration },
  ]
})
function formatPhysicalCapacityBytes(bytes: string): string {
  const value = BigInt(bytes)
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB']
  let divisor = 1n
  let index = 0
  while (value >= divisor * 1024n && index < units.length - 1) { divisor *= 1024n; index++ }
  return index === 0 ? `${value} B` : `${value / divisor}.${((value % divisor) * 100n / divisor).toString().padStart(2, '0')} ${units[index]}`
}
</script>
