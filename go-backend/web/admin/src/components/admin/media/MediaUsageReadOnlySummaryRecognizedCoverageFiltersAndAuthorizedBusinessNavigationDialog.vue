<template>
  <Dialog :open="summary.open.value" @update:open="summary.setMediaUsageSummaryDialogOpen">
    <DialogContent size="lg" @open-auto-focus.prevent>
      <DialogHeader>
        <DialogTitle>媒体使用汇总</DialogTitle>
        <DialogDescription>核对媒体库当前筛选范围，最多 200 个资源。仅统计已接入的引用来源。</DialogDescription>
      </DialogHeader>
      <div class="max-h-[65vh] min-w-0 space-y-3 overflow-y-auto text-xs">
        <p v-if="summary.loading.value" role="status">正在核对使用位置，关闭窗口可取消请求。</p>
        <p v-if="summary.error.value" role="alert" class="text-destructive">{{ summary.error.value }}</p>
        <template v-if="summary.report.value">
          <p class="text-muted-foreground">核对时间：{{ new Date(summary.report.value.completed_at).toLocaleString() }}。业务修改后请重新核对。</p>
          <div class="flex flex-wrap gap-2" aria-label="使用状态筛选">
            <Button v-for="option in mediaUsageSummaryFilterOptions" :key="option.value" size="xs" :variant="selectedFilter === option.value ? 'default' : 'outline'" :aria-pressed="selectedFilter === option.value" @click="selectMediaUsageSummaryFilter(option.value)">
              {{ option.label }} {{ summary.report.value.counts[option.value] }}
            </Button>
          </div>
          <p class="text-muted-foreground">各项可重叠。跨业务、保留和异常仅按获授权引用统计；受限引用的来源、数量与内容不展示。</p>
          <p v-if="!filteredRows.length" class="py-4 text-center text-muted-foreground">当前汇总中没有符合此项的资源</p>
          <article v-for="row in pagedRows" :key="row.asset_id" class="min-w-0 space-y-2 rounded-lg border p-3">
            <p class="break-all font-bold">{{ row.title }} <span class="font-normal text-muted-foreground">#{{ row.asset_id }} · {{ mediaLibraryPartitionLabel(row.library_partition) }}</span></p>
            <p v-if="row.no_recognized_references" class="text-amber-700">本次未识别到引用，仍需核查未接入的业务与历史来源。</p>
            <p v-else>已识别可见引用 {{ row.visible_reference_total }} 条<span v-if="row.has_restricted_references">；另含受限引用，文件仍受保护</span></p>
            <p v-if="row.pending_business_draft" class="text-amber-700">尚未保存到业务的草稿上传，等待绑定或到期核查。</p>
            <p v-if="row.anomalous" class="text-amber-700">发现关联异常，请核查业务记录或媒体登记；核实前保留文件。</p>
            <p v-if="row.retained" class="text-amber-700">含历史保留、停用内容或待绑定引用。</p>
            <Button v-if="row.anomalous || row.retained" size="xs" variant="outline" @click="summary.setMediaUsageSummaryDialogOpen(false); emit('diagnose', row.asset_id)">异常关联诊断</Button>
            <details v-if="row.references.length">
              <summary class="cursor-pointer">查看获授权使用位置</summary>
              <div v-for="(reference, index) in row.references" :key="index" class="mt-2 space-y-1 border-t pt-2">
                <p class="break-all">{{ reference.label || mediaReferenceTypeLabel(reference.resource_type) }}</p>
                <p class="break-all text-muted-foreground">{{ mediaReferenceCategoryLabel(reference.category) }} · {{ mediaReferenceTypeLabel(reference.resource_type) }} #{{ reference.resource_id }}<span v-if="reference.parent_resource_id !== '0'"> · 所属记录 #{{ reference.parent_resource_id }}</span> · {{ mediaUsageReferenceStateLabel(reference.usage_state) }}</p>
                <RouterLink v-if="mediaUsageReferenceAuthorizedBusinessListDestination(reference, auth.hasPermission)" :to="mediaUsageReferenceAuthorizedBusinessListDestination(reference, auth.hasPermission)!" class="inline-block underline" @click="summary.setMediaUsageSummaryDialogOpen(false)">打开业务列表核查</RouterLink>
                <p v-else-if="reference.usage_state === 'missing_parent'" class="text-amber-700">业务父记录缺失，需人工核查；此处保留关联。</p>
              </div>
              <p v-if="row.reference_preview_limited" class="mt-2 text-muted-foreground">仅展示前 20 条可见引用，统计包含其余可见引用。</p>
            </details>
          </article>
          <div v-if="filteredRows.length > 20" class="flex flex-wrap items-center justify-between gap-2">
            <Button size="xs" variant="outline" :disabled="page === 1" @click="page--">上一页</Button>
            <span>第 {{ page }} / {{ Math.ceil(filteredRows.length / 20) }} 页 · {{ filteredRows.length }} 个资源</span>
            <Button size="xs" variant="outline" :disabled="page * 20 >= filteredRows.length" @click="page++">下一页</Button>
          </div>
        </template>
        <p class="text-muted-foreground">未识别到引用不代表可直接删除。清理时仍会重新核查真实引用、保留策略与文件所有权。</p>
      </div>
      <DialogFooter>
        <Button variant="outline" :disabled="summary.loading.value" @click="summary.rebuildMediaUsageSummary">重新核对</Button>
        <Button variant="outline" @click="summary.setMediaUsageSummaryDialogOpen(false)">关闭</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { Button } from '@/components/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { mediaUsageSummaryFilterOptions, mediaUsageSummaryRowMatchesVisibleFilter, type MediaUsageSummaryFilter } from '@/api/mediaUsageReadOnlyRebuildableSummaryWithValidatedCoverageAndReferenceStates'
import type { useReadOnlyMediaUsageSummaryWithScopeInvalidationCancellationAndPermissionRevocation } from '@/composables/media/useReadOnlyMediaUsageSummaryWithScopeInvalidationCancellationAndPermissionRevocation'
import { mediaUsageReferenceAuthorizedBusinessListDestination } from '@/lib/mediaUsageReferencePermissionCheckedBusinessListNavigation'
import { mediaReferenceCategoryLabel, mediaReferenceTypeLabel } from '@/lib/mediaReferencePresentation'
import { mediaLibraryPartitionLabel } from '@/modules/media/mediaLibraryBusinessPartitionUploadOriginAndCustomerServicePurposePresentation'
import { useAuthStore } from '@/stores/auth'

const props = defineProps<{ summary: ReturnType<typeof useReadOnlyMediaUsageSummaryWithScopeInvalidationCancellationAndPermissionRevocation> }>()
const emit = defineEmits<{ (event: 'diagnose', assetID: string): void }>()
const auth = useAuthStore()
const selectedFilter = ref<MediaUsageSummaryFilter>('all')
const page = ref(1)
const filteredRows = computed(() => (props.summary.report.value?.rows ?? []).filter((row) => mediaUsageSummaryRowMatchesVisibleFilter(row, selectedFilter.value)))
const pagedRows = computed(() => filteredRows.value.slice((page.value - 1) * 20, page.value * 20))
function selectMediaUsageSummaryFilter(value: MediaUsageSummaryFilter) { selectedFilter.value = value; page.value = 1 }
watch(() => props.summary.report.value, () => { selectedFilter.value = 'all'; page.value = 1 })
function mediaUsageReferenceStateLabel(state: string): string {
  const labels: Record<string, string> = { active: '使用中', missing_parent: '父记录缺失', retained_deleted_parent: '已删除内容保留', retained_deleted_placement: '已删除关联保留', retained_unpublished_parent: '未发布内容保留', retained_upload_receipt: '已绑定收据保留', disabled: '已停用保留', draft: '草稿保留', archived: '归档保留', pending_upload: '待绑定保留', immutable_history: '不可变历史保留' }
  return labels[state] || '保留引用'
}
</script>
