<template>
  <div class="flex h-full min-h-0 flex-col gap-4 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="买家秀审批"
      description="审核 Picture Warehouse 用户投稿，只有通过后的图片才会进入公开展示"
    >
      <template #actions>
        <Button variant="outline" size="sm" :disabled="loading" @click="fetchItems">
          <RefreshCw :class="['size-3.5', loading && 'animate-spin']" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <div class="flex shrink-0 flex-wrap items-center justify-between gap-3 border-y py-3">
      <div class="flex flex-wrap items-center gap-2">
        <Button
          v-for="option in statusOptions"
          :key="option.value"
          size="sm"
          :variant="status === option.value ? 'default' : 'outline'"
          @click="setStatus(option.value)"
        >
          {{ option.label }}
        </Button>
      </div>
      <span class="text-xs text-muted-foreground">
        共 {{ pagination.total }} 条
      </span>
    </div>

    <AdminTablePanel class="min-h-0 flex-1" :loading="loading" :batch-visible="false" scroll-body>
      <Table class="min-w-[1080px]">
        <TableHeader>
          <TableRow>
            <TableHead class="w-20">ID</TableHead>
            <TableHead class="w-24">图片</TableHead>
            <TableHead class="w-32">投稿用户</TableHead>
            <TableHead class="w-56">地区</TableHead>
            <TableHead>投稿说明</TableHead>
            <TableHead class="w-32">状态</TableHead>
            <TableHead class="w-44">提交时间</TableHead>
            <TableHead class="w-28 text-right">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="!loading && items.length === 0">
            <TableCell colspan="8" class="h-40 text-center text-sm text-muted-foreground">
              当前筛选下没有买家秀投稿
            </TableCell>
          </TableRow>
          <TableRow
            v-for="item in items"
            :key="item.id"
            class="cursor-pointer"
 :class="selected?.id === item.id ? 'bg-muted/50': ''"
            @click="selectItem(item)"
          >
            <TableCell class="font-mono text-xs">#{{ item.id }}</TableCell>
            <TableCell>
              <div class="relative size-16 overflow-hidden rounded-md border bg-muted">
                <img
                  v-if="item.image_files?.[0]?.file_url"
                  :src="item.image_files[0].file_url"
                  :alt="item.title || `Showcase ${item.id}`"
                  class="size-full object-cover"
                />
                <div v-else class="flex size-full items-center justify-center text-muted-foreground">
                  <ImageOff class="size-4" />
                </div>
                <span class="absolute bottom-1 right-1 rounded bg-black/70 px-1 text-[9px] text-white">
                  {{ item.image_count }}
                </span>
              </div>
            </TableCell>
            <TableCell>
              <div class="space-y-1">
 <p class="text-xs font-semibold">{{ item.nickname || '未填写昵称'}}</p>
                <p class="font-mono text-[10px] text-muted-foreground">USER #{{ item.user_id }}</p>
              </div>
            </TableCell>
            <TableCell>
              <div class="space-y-1 text-xs">
                <p>{{ locationLabel(item) }}</p>
                <p class="text-muted-foreground">
                  订单 {{ item.order_id ? `#${item.order_id}` : '未关联订单' }}
                </p>
              </div>
            </TableCell>
            <TableCell>
              <p class="max-w-xl whitespace-pre-wrap text-xs leading-5 text-muted-foreground">
                {{ item.notes || '未填写投稿说明' }}
              </p>
              <p v-if="item.rejected_reason" class="mt-1 text-xs text-rose-600">
                拒绝原因：{{ item.rejected_reason }}
              </p>
            </TableCell>
            <TableCell>
              <Badge :variant="statusBadgeVariant(item.status)">
                {{ statusLabel(item.status) }}
              </Badge>
            </TableCell>
            <TableCell class="text-xs text-muted-foreground">
              {{ formatDate(item.created_at) }}
            </TableCell>
            <TableCell class="text-right">
              <Button variant="ghost" size="icon" aria-label="查看投稿" @click.stop="selectItem(item)">
                <Eye class="size-4" />
              </Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>

      <template #footer>
        <AdminPagination
          :page="pagination.page"
          :page-size="pagination.page_size"
          :total="pagination.total"
          :page-sizes="[10, 20, 50]"
          @update:page="updatePage"
          @update:page-size="updatePageSize"
        />
      </template>
    </AdminTablePanel>
  </div>

  <Dialog :open="Boolean(selected)" @update:open="handleDialogOpen">
    <DialogScrollContent class="max-w-4xl">
      <DialogHeader>
        <DialogTitle>买家秀 #{{ selected?.id }}</DialogTitle>
        <DialogDescription>
          用户 #{{ selected?.user_id }} · {{ selected ? locationLabel(selected) : '' }} ·
          {{ selected ? formatDate(selected.created_at) : '' }}
        </DialogDescription>
      </DialogHeader>

      <div v-if="selected" class="space-y-5">
        <div class="grid grid-cols-2 gap-3 sm:grid-cols-3">
          <div
            v-for="image in selected.image_files"
            :key="image.index"
            class="aspect-square overflow-hidden rounded-md border bg-muted"
          >
            <img
              :src="image.file_url"
              :alt="`Showcase ${selected.id} image ${image.index + 1}`"
              class="size-full object-contain"
            />
          </div>
        </div>

        <div class="grid gap-4 border-y py-4 sm:grid-cols-2">
          <div>
            <p class="text-[10px] font-bold uppercase text-muted-foreground">投稿信息</p>
 <p class="mt-1 text-sm">{{ selected.nickname || '未填写昵称'}}</p>
            <p class="text-xs text-muted-foreground">{{ locationLabel(selected) }}</p>
            <p class="text-xs text-muted-foreground">
              订单 {{ selected.order_id ? `#${selected.order_id}` : '未关联订单' }}
            </p>
          </div>
          <div>
            <p class="text-[10px] font-bold uppercase text-muted-foreground">当前状态</p>
            <Badge class="mt-1" :variant="statusBadgeVariant(selected.status)">
              {{ statusLabel(selected.status) }}
            </Badge>
          </div>
        </div>

        <div>
          <p class="text-[10px] font-bold uppercase text-muted-foreground">投稿说明</p>
          <p class="mt-1 whitespace-pre-wrap text-sm leading-6">
            {{ selected.notes || '未填写投稿说明' }}
          </p>
        </div>

 <div v-if="selected.status === 'pending'&& canModerate" class="space-y-2">
          <label for="showcase-reject-reason" class="text-xs font-semibold">拒绝原因</label>
          <Textarea
            id="showcase-reject-reason"
            v-model="rejectReason"
            class="min-h-24 resize-y"
            maxlength="1000"
            placeholder="拒绝时必须填写，用户内容、图片质量或合规原因均应留痕"
          />
        </div>
      </div>

      <DialogFooter v-if="selected?.status === 'pending' && canModerate">
        <Button variant="outline" :disabled="submitting" @click="rejectSelected">
          <XCircle class="size-4" />
          拒绝
        </Button>
        <Button :disabled="submitting" @click="approveSelected">
          <CircleCheck class="size-4" />
          通过
        </Button>
      </DialogFooter>
    </DialogScrollContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { CircleCheck, Eye, ImageOff, RefreshCw, XCircle } from '@lucide/vue'
import { toast } from 'vue-sonner'
import showcaseApi, {
  type ShowcaseRecord,
  type ShowcaseStatus,
  type ShowcaseStatusFilter,
} from '@/api/showcase'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogScrollContent,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Textarea } from '@/components/ui/textarea'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const items = ref<ShowcaseRecord[]>([])
const selected = ref<ShowcaseRecord | null>(null)
const status = ref<ShowcaseStatusFilter>('pending')
const loading = ref(false)
const submitting = ref(false)
const rejectReason = ref('')
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  total_pages: 0,
})

const statusOptions: Array<{ value: ShowcaseStatusFilter; label: string }> = [
  { value: 'pending', label: '待审核' },
  { value: 'approved', label: '已通过' },
  { value: 'rejected', label: '已拒绝' },
  { value: 'all', label: '全部' },
]

const canModerate = computed(() => authStore.hasPermission('gallery:edit'))

const fetchItems = async (): Promise<void> => {
  loading.value = true
  try {
    const result = await showcaseApi.list({
      status: status.value,
      page: pagination.page,
      page_size: pagination.page_size,
    })
    items.value = result.items
    Object.assign(pagination, result.pagination)
  } catch (error) {
    console.error('Failed to load showcase submissions:', error)
    items.value = []
    toast.error('买家秀投稿加载失败')
  } finally {
    loading.value = false
  }
}

const selectItem = (item: ShowcaseRecord): void => {
  selected.value = item
  rejectReason.value = item.rejected_reason || ''
}

const handleDialogOpen = (open: boolean): void => {
  if (open) return
  selected.value = null
  rejectReason.value = ''
}

const approveSelected = async (): Promise<void> => {
  if (!selected.value || submitting.value) return
  submitting.value = true
  try {
    await showcaseApi.approve(selected.value.id)
    toast.success('买家秀已通过并发布')
    handleDialogOpen(false)
    await fetchItems()
  } catch (error: any) {
    toast.error(error?.response?.data?.error || '审批失败')
  } finally {
    submitting.value = false
  }
}

const rejectSelected = async (): Promise<void> => {
  if (!selected.value || submitting.value) return
  const reason = rejectReason.value.trim()
  if (!reason) {
    toast.error('拒绝时必须填写原因')
    return
  }
  submitting.value = true
  try {
    await showcaseApi.reject(selected.value.id, reason)
    toast.success('买家秀已拒绝，待审图片已清理')
    handleDialogOpen(false)
    await fetchItems()
  } catch (error: any) {
    toast.error(error?.response?.data?.error || '拒绝失败')
  } finally {
    submitting.value = false
  }
}

const setStatus = (nextStatus: ShowcaseStatusFilter): void => {
  if (status.value === nextStatus) return
  status.value = nextStatus
  pagination.page = 1
}

const updatePage = (page: number): void => {
  pagination.page = page
}

const updatePageSize = (pageSize: number): void => {
  pagination.page = 1
  pagination.page_size = pageSize
}

const locationLabel = (item: ShowcaseRecord): string =>
  [item.region, item.location].filter(Boolean).join(' / ') || '未填写地区'

const formatDate = (value?: string | null): string =>
  value ? new Date(value).toLocaleString('zh-CN') : '-'

const statusLabel = (value: ShowcaseStatus): string => ({
  pending: '待审核',
  approved: '已通过',
  rejected: '已拒绝',
}[value] || value)

const statusBadgeVariant = (value: ShowcaseStatus): 'default' | 'secondary' | 'destructive' =>
  value === 'approved' ? 'default' : value === 'rejected' ? 'destructive' : 'secondary'

watch([status, () => pagination.page, () => pagination.page_size], () => {
  void fetchItems()
})

onMounted(() => {
  void fetchItems()
})

</script>
