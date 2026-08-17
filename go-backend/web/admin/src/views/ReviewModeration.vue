<template>
  <div class="flex h-full min-h-0 flex-col gap-4 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="评价审核"
      description="审核通过后，评价才会进入商品详情、评分摘要和搜索引擎结构化数据"
    >
      <template #actions>
        <Button variant="outline" size="sm" :disabled="loading" @click="fetchReviews">
          <RefreshCw :class="['size-3.5', loading && 'animate-spin']" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminFilterPanel class="shrink-0">
      <form class="grid grid-cols-1 gap-3 md:grid-cols-[180px_minmax(0,1fr)_auto]" @submit.prevent="applyFilters">
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">STATUS / 状态</span>
          <select v-model="status" class="h-9 w-full rounded-md border border-dashed border-border bg-background px-3 text-sm">
            <option value="pending">待审核</option>
            <option value="approved">已通过</option>
            <option value="rejected">已拒绝</option>
            <option value="all">全部</option>
          </select>
        </label>
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SEARCH / 搜索</span>
          <Input v-model="searchInput" placeholder="商品、SKU、用户、标题或内容" />
        </label>
        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9 px-3 text-xs font-black uppercase tracking-wider" :disabled="loading">
            <Search class="size-3.5" />
            查询
          </Button>
          <Button type="button" variant="outline" class="h-9 px-3 text-xs font-black uppercase tracking-wider" @click="resetFilters">
            重置
          </Button>
        </div>
      </form>
    </AdminFilterPanel>

    <section class="grid min-h-0 flex-1 grid-cols-1 gap-3 overflow-hidden xl:grid-cols-[minmax(0,1fr)_380px]">
      <AdminTablePanel :loading="loading" scroll-body>
        <Table class="min-w-[860px]">
          <TableHeader>
            <TableRow>
              <TableHead class="w-20">ID</TableHead>
              <TableHead class="w-28">评分</TableHead>
              <TableHead class="w-48">商品</TableHead>
              <TableHead class="w-36">用户</TableHead>
              <TableHead>评价</TableHead>
              <TableHead class="w-28">状态</TableHead>
              <TableHead class="w-40">提交时间</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow v-if="!loading && reviews.length === 0">
              <TableCell colspan="7" class="h-40 text-center text-sm text-muted-foreground">
                当前筛选下没有评价
              </TableCell>
            </TableRow>
            <TableRow
              v-for="item in reviews"
              :key="item.id"
              class="cursor-pointer"
 :class="selected?.id === item.id ? 'bg-primary/5': ''"
              @click="selectReview(item)"
            >
              <TableCell class="font-mono text-xs">#{{ item.id }}</TableCell>
              <TableCell>
                <div class="flex items-center gap-0.5 text-amber-500">
                  <Star
                    v-for="star in 5"
                    :key="star"
 :class="['size-3', star <= item.rating ? 'fill-current': 'text-muted-foreground/30']"
                  />
                </div>
              </TableCell>
              <TableCell>
 <p class="truncate text-xs font-semibold">{{ item.product?.name || '商品已删除'}}</p>
 <p class="font-mono text-[10px] text-muted-foreground">{{ item.product?.sku || '-'}}</p>
              </TableCell>
              <TableCell>
 <p class="truncate text-xs font-semibold">{{ item.user?.display_name || '未知用户'}}</p>
 <p class="truncate text-[10px] text-muted-foreground">{{ item.user?.email || '-'}}</p>
              </TableCell>
              <TableCell>
 <p class="truncate text-xs font-semibold">{{ item.title || '无标题'}}</p>
 <p class="max-w-[300px] truncate text-xs text-muted-foreground">{{ item.content || '无内容'}}</p>
              </TableCell>
              <TableCell>
                <Badge :variant="statusBadgeVariant(item.status)">
                  {{ statusLabel(item.status) }}
                </Badge>
              </TableCell>
              <TableCell class="text-xs text-muted-foreground">
                {{ formatDate(item.created_at) }}
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

      <aside class="min-h-0 overflow-auto rounded-[24px] border border-dashed border-border/80 bg-card p-4">
        <div v-if="selected" class="space-y-5">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">REVIEW #{{ selected.id }}</p>
 <h2 class="mt-1 truncate text-base font-black">{{ selected.title || '无标题评价'}}</h2>
            </div>
            <Badge :variant="statusBadgeVariant(selected.status)">
              {{ statusLabel(selected.status) }}
            </Badge>
          </div>

          <div class="grid grid-cols-2 gap-2 text-xs">
            <div class="rounded-xl bg-muted/40 p-3">
              <p class="font-black uppercase text-muted-foreground">商品</p>
 <p class="mt-1 truncate">{{ selected.product?.name || '商品已删除'}}</p>
 <p class="font-mono text-[10px] text-muted-foreground">{{ selected.product?.sku || '-'}}</p>
            </div>
            <div class="rounded-xl bg-muted/40 p-3">
              <p class="font-black uppercase text-muted-foreground">用户</p>
 <p class="mt-1 truncate">{{ selected.user?.display_name || '未知用户'}}</p>
 <p class="truncate text-[10px] text-muted-foreground">{{ selected.user?.email || '-'}}</p>
            </div>
          </div>

          <div class="flex items-center gap-1 text-amber-500">
            <Star
              v-for="star in 5"
              :key="star"
 :class="['size-4', star <= selected.rating ? 'fill-current': 'text-muted-foreground/30']"
            />
            <span class="ml-1 text-xs font-semibold text-foreground">{{ selected.rating }}/5</span>
          </div>

          <div class="space-y-2 text-sm">
 <p class="whitespace-pre-wrap leading-6">{{ selected.content || '无评价内容'}}</p>
            <div v-if="selected.pros" class="border-l-2 border-emerald-500/50 pl-3 text-xs text-muted-foreground">
              优点：{{ selected.pros }}
            </div>
            <div v-if="selected.cons" class="border-l-2 border-rose-500/50 pl-3 text-xs text-muted-foreground">
              缺点：{{ selected.cons }}
            </div>
          </div>

          <div v-if="selected.images.length" class="grid grid-cols-3 gap-2">
            <img
              v-for="(image, index) in selected.images"
              :key="`${selected.id}-${index}`"
              :src="image"
              :alt="`评价 ${selected.id} 图片 ${index + 1}`"
              class="aspect-square rounded-md border object-cover"
              loading="lazy"
            />
          </div>

          <div class="border-y border-dashed py-3 text-xs text-muted-foreground">
            <p>提交：{{ formatDate(selected.created_at) }}</p>
            <p v-if="selected.moderated_at">审核：{{ formatDate(selected.moderated_at) }}</p>
            <p v-if="selected.moderation_reason" class="mt-1 text-rose-600">
              审核备注：{{ selected.moderation_reason }}
            </p>
          </div>

          <div v-if="canModerate" class="space-y-2">
            <label for="review-moderation-reason" class="text-xs font-semibold">审核备注</label>
            <Textarea
              id="review-moderation-reason"
              v-model="moderationReason"
              class="min-h-24 resize-y"
              maxlength="1000"
              placeholder="拒绝或退回待审核时建议填写原因"
            />
            <div class="flex flex-wrap gap-2">
              <Button
                v-if="selected.status === 'pending'"
                :disabled="submitting"
                @click="updateSelectedStatus('approved')"
              >
                <CircleCheck class="size-4" />
                通过并发布
              </Button>
              <Button
                v-if="selected.status === 'pending'"
                variant="outline"
                :disabled="submitting"
                @click="updateSelectedStatus('rejected')"
              >
                <XCircle class="size-4" />
                拒绝
              </Button>
              <Button
                v-if="selected.status !== 'pending'"
                variant="outline"
                :disabled="submitting"
                @click="updateSelectedStatus('pending')"
              >
                <RotateCcw class="size-4" />
                退回待审核
              </Button>
            </div>
          </div>
        </div>
        <div v-else class="flex h-full min-h-64 items-center justify-center text-center text-sm text-muted-foreground">
          选择一条评价查看详情
        </div>
      </aside>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { CircleCheck, RefreshCw, RotateCcw, Search, Star, XCircle } from '@lucide/vue'
import { toast } from 'vue-sonner'
import reviewApi, { type AdminReview, type ReviewStatus, type ReviewStatusFilter } from '@/api/review'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const reviews = ref<AdminReview[]>([])
const selected = ref<AdminReview | null>(null)
const status = ref<ReviewStatusFilter>('pending')
const searchInput = ref('')
const search = ref('')
const loading = ref(false)
const submitting = ref(false)
const moderationReason = ref('')
const pagination = reactive({
  page: 1,
  page_size: 20,
  total: 0,
  total_pages: 0,
})

const canModerate = computed(() => authStore.hasPermission('review:moderate'))

const fetchReviews = async (): Promise<void> => {
  loading.value = true
  try {
    const result = await reviewApi.list({
      status: status.value,
      search: search.value || undefined,
      page: pagination.page,
      page_size: pagination.page_size,
    })
    reviews.value = result.data
    Object.assign(pagination, result.pagination)
    if (selected.value) {
      selected.value = reviews.value.find((item) => item.id === selected.value?.id) || null
    }
  } catch (error) {
    console.error('Failed to load reviews:', error)
    reviews.value = []
    toast.error('评价列表加载失败')
  } finally {
    loading.value = false
  }
}

const applyFilters = (): void => {
  search.value = searchInput.value.trim()
  pagination.page = 1
  void fetchReviews()
}

const resetFilters = (): void => {
  status.value = 'pending'
  searchInput.value = ''
  search.value = ''
  pagination.page = 1
  void fetchReviews()
}

const selectReview = (item: AdminReview): void => {
  selected.value = item
  moderationReason.value = item.moderation_reason || ''
}

const updateSelectedStatus = async (nextStatus: ReviewStatus): Promise<void> => {
  if (!selected.value || submitting.value) return
  if (nextStatus === 'rejected' && !moderationReason.value.trim()) {
    toast.error('拒绝时请填写审核备注')
    return
  }
  submitting.value = true
  try {
    const updated = await reviewApi.updateStatus(
      selected.value.id,
      nextStatus,
      moderationReason.value.trim(),
    )
    selected.value = updated
    moderationReason.value = updated.moderation_reason || ''
    toast.success(nextStatus === 'approved' ? '评价已通过，评分摘要已刷新' : '评价状态已更新，评分摘要已刷新')
    await fetchReviews()
  } catch (error: any) {
    toast.error(error?.response?.data?.error || '评价审核失败')
  } finally {
    submitting.value = false
  }
}

const updatePage = (page: number): void => {
  pagination.page = page
}

const updatePageSize = (pageSize: number): void => {
  pagination.page = 1
  pagination.page_size = pageSize
}

const formatDate = (value?: string | null): string =>
  value ? new Date(value).toLocaleString('zh-CN') : '-'

const statusLabel = (value: ReviewStatus): string => ({
  pending: '待审核',
  approved: '已通过',
  rejected: '已拒绝',
}[value] || value)

const statusBadgeVariant = (value: ReviewStatus): 'default' | 'secondary' | 'destructive' =>
  value === 'approved' ? 'default' : value === 'rejected' ? 'destructive' : 'secondary'

watch([status, () => pagination.page, () => pagination.page_size], () => {
  void fetchReviews()
})

onMounted(() => {
  void fetchReviews()
})
</script>
