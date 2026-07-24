<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="保修管理"
      description="统一查看产品注册、保修申请和即将到期记录；当前只消费 Go 后端保修事实源。"
    >
      <template #actions>
        <Button variant="outline" :disabled="refreshing" @click="refreshCurrent">
          <RefreshCw :class="['size-3.5', { 'animate-spin': refreshing }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <Tabs v-model="activeTab" class="gap-4">
      <TabsList variant="line" class="h-10 w-full justify-start overflow-x-auto rounded-none border-b bg-transparent p-0">
        <TabsTrigger value="registrations" class="h-9 flex-none px-3">
          <ShieldCheck class="size-4" />
          注册记录
        </TabsTrigger>
        <TabsTrigger value="claims" class="h-9 flex-none px-3">
          <FileWarning class="size-4" />
          保修申请
        </TabsTrigger>
        <TabsTrigger value="expiring" class="h-9 flex-none px-3">
          <Clock3 class="size-4" />
          即将到期
        </TabsTrigger>
        <TabsTrigger value="boundary" class="h-9 flex-none px-3">
          <GitBranch class="size-4" />
          数据边界
        </TabsTrigger>
      </TabsList>

      <TabsContent value="registrations" class="space-y-3">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h2 class="text-sm font-black tracking-tighter italic uppercase">产品注册记录</h2>
            <p class="mt-1 text-xs text-muted-foreground">按序列号、用户、商品和保修到期时间管理注册状态。</p>
          </div>
          <div class="w-full sm:w-48">
            <Select v-model="registrationFilters.status">
              <SelectTrigger class="h-9 w-full rounded-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="active">有效</SelectItem>
                <SelectItem value="expired">已过期</SelectItem>
                <SelectItem value="claimed">已申请</SelectItem>
                <SelectItem value="cancelled">已取消</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <AdminTablePanel :loading="loading.registrations">
          <Table class="min-w-[1080px]">
            <TableHeader>
              <TableRow>
                <TableHead>注册商品</TableHead>
                <TableHead>序列号</TableHead>
                <TableHead>客户</TableHead>
                <TableHead>购买 / 到期</TableHead>
                <TableHead>凭证</TableHead>
                <TableHead class="w-40">状态</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="registrations.length === 0" :colspan="6">
                <div class="flex flex-col items-center text-muted-foreground">
                  <ShieldCheck class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无产品注册记录</span>
                </div>
              </TableEmpty>
              <TableRow v-for="registration in registrations" :key="registration.id">
                <TableCell>
                  <span class="block font-bold text-xs">{{ productName(registration.product) }}</span>
                  <span class="block font-mono text-[10px] text-muted-foreground/70">product_id={{ registration.product_id || '-' }}</span>
                </TableCell>
                <TableCell>
                  <span class="font-mono text-xs font-bold">{{ registration.serial_number || '-' }}</span>
                </TableCell>
                <TableCell>
                  <span class="block text-xs font-bold">{{ userName(registration.user) }}</span>
                  <span class="block max-w-56 truncate text-[10px] text-muted-foreground/70">{{ registration.user?.email || '-' }}</span>
                </TableCell>
                <TableCell>
                  <span class="block text-xs">购买：{{ formatDate(registration.purchase_date) }}</span>
                  <span class="block text-[10px] text-muted-foreground/70">到期：{{ formatDate(registration.warranty_expires) }}</span>
                </TableCell>
                <TableCell>
                  <Button
                    v-if="registration.purchase_proof"
                    variant="ghost"
                    size="sm"
                    class="h-7 rounded-full px-2 text-xs"
                    as-child
                  >
                    <a :href="registration.purchase_proof" target="_blank" rel="noopener noreferrer">查看凭证</a>
                  </Button>
                  <span v-else class="text-xs text-muted-foreground">无凭证</span>
                </TableCell>
                <TableCell>
                  <Select
                    :model-value="registration.status || 'active'"
                    :disabled="statusUpdating.registration === registration.id || !canEdit"
                    @update:model-value="updateRegistrationStatus(registration, $event)"
                  >
                    <SelectTrigger class="h-8 w-full rounded-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="active">有效</SelectItem>
                      <SelectItem value="expired">已过期</SelectItem>
                      <SelectItem value="claimed">已申请</SelectItem>
                      <SelectItem value="cancelled">已取消</SelectItem>
                    </SelectContent>
                  </Select>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
          <template #footer>
            <AdminPagination
              v-model:page="registrationPagination.page"
              v-model:page-size="registrationPagination.pageSize"
              :total="registrationPagination.total"
            />
          </template>
        </AdminTablePanel>
      </TabsContent>

      <TabsContent value="claims" class="space-y-3">
        <div class="flex flex-wrap items-end justify-between gap-3">
          <div>
            <h2 class="text-sm font-black tracking-tighter italic uppercase">保修申请处理</h2>
            <p class="mt-1 text-xs text-muted-foreground">订单型申请和注册型申请统一进入这里，关联注册记录为空时保持真实为空。</p>
          </div>
          <div class="w-full sm:w-48">
            <Select v-model="claimFilters.status">
              <SelectTrigger class="h-9 w-full rounded-full">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="submitted">已提交</SelectItem>
                <SelectItem value="reviewing">审核中</SelectItem>
                <SelectItem value="approved">已批准</SelectItem>
                <SelectItem value="rejected">已拒绝</SelectItem>
                <SelectItem value="completed">已完成</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </div>

        <AdminTablePanel :loading="loading.claims">
          <Table class="min-w-[1180px]">
            <TableHeader>
              <TableRow>
                <TableHead>申请来源</TableHead>
                <TableHead>关联商品</TableHead>
                <TableHead>问题说明</TableHead>
                <TableHead>证据</TableHead>
                <TableHead>提交时间</TableHead>
                <TableHead class="w-40">状态</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="claims.length === 0" :colspan="6">
                <div class="flex flex-col items-center text-muted-foreground">
                  <FileWarning class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无保修申请</span>
                </div>
              </TableEmpty>
              <TableRow v-for="claim in claims" :key="claim.id">
                <TableCell>
                  <span class="block font-mono text-xs font-bold">#{{ claim.id }}</span>
                  <span class="block max-w-60 truncate text-[10px] text-muted-foreground/70">
                    订单：{{ claim.order_number || '-' }}
                  </span>
                  <span class="block max-w-60 truncate text-[10px] text-muted-foreground/70">
                    邮箱：{{ claim.email || claim.registration?.user?.email || '-' }}
                  </span>
                </TableCell>
                <TableCell>
                  <span class="block text-xs font-bold">{{ registrationProductName(claim) }}</span>
                  <span class="block font-mono text-[10px] text-muted-foreground/70">
                    registration_id={{ claim.registration_id || '-' }}
                  </span>
                </TableCell>
                <TableCell>
                  <span class="block text-xs font-bold">{{ issueTypeLabel(claim.issue_type) }}</span>
                  <span class="block max-w-[26rem] truncate text-[10px] text-muted-foreground/70">
                    {{ claim.description || '-' }}
                  </span>
                  <span v-if="claim.tire_pressure || claim.is_tubeless" class="block text-[10px] text-muted-foreground/70">
                    胎压 {{ claim.tire_pressure || '-' }} / {{ claim.is_tubeless ? '真空胎' : '非真空胎' }}
                  </span>
                </TableCell>
                <TableCell>
                  <div class="flex flex-wrap gap-1.5">
                    <Badge v-for="(image, index) in claimImages(claim)" :key="`${claim.id}-image-${index}`" variant="outline" class="rounded-full">
                      <a :href="image" target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-1 no-underline">
                        <ImageIcon class="size-3" />
                        图 {{ index + 1 }}
                      </a>
                    </Badge>
                    <Badge v-if="claim.video_url" variant="outline" class="rounded-full">
                      <a :href="claim.video_url" target="_blank" rel="noopener noreferrer" class="inline-flex items-center gap-1 no-underline">
                        <Video class="size-3" />
                        视频
                      </a>
                    </Badge>
                    <span v-if="claimImages(claim).length === 0 && !claim.video_url" class="text-xs text-muted-foreground">无附件</span>
                  </div>
                </TableCell>
                <TableCell>
                  <span class="block text-xs">{{ formatDateTime(claim.created_at) }}</span>
                  <span class="block text-[10px] text-muted-foreground/70">处理：{{ formatDateTime(claim.processed_at) }}</span>
                </TableCell>
                <TableCell>
                  <Select
                    :model-value="claim.status || 'submitted'"
                    :disabled="statusUpdating.claim === claim.id || !canEdit"
                    @update:model-value="updateClaimStatus(claim, $event)"
                  >
                    <SelectTrigger class="h-8 w-full rounded-full">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="submitted">已提交</SelectItem>
                      <SelectItem value="reviewing">审核中</SelectItem>
                      <SelectItem value="approved">已批准</SelectItem>
                      <SelectItem value="rejected">已拒绝</SelectItem>
                      <SelectItem value="completed">已完成</SelectItem>
                    </SelectContent>
                  </Select>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
          <template #footer>
            <AdminPagination
              v-model:page="claimPagination.page"
              v-model:page-size="claimPagination.pageSize"
              :total="claimPagination.total"
            />
          </template>
        </AdminTablePanel>
      </TabsContent>

      <TabsContent value="expiring" class="space-y-3">
        <div class="rounded-[24px] border border-dashed bg-card p-4">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h2 class="text-sm font-black tracking-tighter italic uppercase">30 天内到期</h2>
              <p class="mt-1 text-xs text-muted-foreground">从后端 `/registrations/expiring` 读取，不在前端重新推算。</p>
            </div>
            <AdminStatusBadge tone="amber">{{ expiring.length }} 条</AdminStatusBadge>
          </div>
        </div>

        <AdminTablePanel :loading="loading.expiring">
          <Table class="min-w-[980px]">
            <TableHeader>
              <TableRow>
                <TableHead>商品</TableHead>
                <TableHead>序列号</TableHead>
                <TableHead>客户</TableHead>
                <TableHead>保修到期</TableHead>
                <TableHead>状态</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableEmpty v-if="expiring.length === 0" :colspan="5">
                <div class="flex flex-col items-center text-muted-foreground">
                  <Clock3 class="mb-2 size-7 opacity-55" />
                  <span class="text-xs">暂无即将到期记录</span>
                </div>
              </TableEmpty>
              <TableRow v-for="item in expiring" :key="item.id">
                <TableCell>{{ productName(item.product) }}</TableCell>
                <TableCell class="font-mono text-xs font-bold">{{ item.serial_number || '-' }}</TableCell>
                <TableCell>{{ userName(item.user) }}</TableCell>
                <TableCell>{{ formatDate(item.warranty_expires) }}</TableCell>
                <TableCell>
                  <AdminStatusBadge :tone="registrationStatusTone(item.status)">
                    {{ registrationStatusLabel(item.status) }}
                  </AdminStatusBadge>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </AdminTablePanel>
      </TabsContent>

      <TabsContent value="boundary" class="space-y-4">
        <section class="grid gap-4 lg:grid-cols-3">
          <div class="rounded-[24px] border border-dashed bg-card p-4">
            <ShieldCheck class="size-5 text-emerald-600" />
            <h2 class="mt-3 text-sm font-black tracking-tighter italic uppercase">单一事实源</h2>
            <p class="mt-2 text-xs leading-5 text-muted-foreground">
              后台只读取 Go 后端 `product_registrations` 和 `warranty_claims`，不接 WordPress，也不在前端保留影子数据。
            </p>
          </div>
          <div class="rounded-[24px] border border-dashed bg-card p-4">
            <GitBranch class="size-5 text-blue-600" />
            <h2 class="mt-3 text-sm font-black tracking-tighter italic uppercase">注册关联边界</h2>
            <p class="mt-2 text-xs leading-5 text-muted-foreground">
              注册型申请必须绑定 `registration_id`；订单型申请在产品/序列号链路确认前允许为空，后台如实显示为空，不用 0 兜底。
            </p>
          </div>
          <div class="rounded-[24px] border border-dashed bg-card p-4">
            <FileWarning class="size-5 text-amber-600" />
            <h2 class="mt-3 text-sm font-black tracking-tighter italic uppercase">下一阶段</h2>
            <p class="mt-2 text-xs leading-5 text-muted-foreground">
              下一步再加服务记录、申请详情富文本处理备注和订单行项目到注册记录的绑定，不把逻辑散到 Nuxt 页面里。
            </p>
          </div>
        </section>
      </TabsContent>
    </Tabs>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import {
  Clock3,
  FileWarning,
  GitBranch,
  Image as ImageIcon,
  RefreshCw,
  ShieldCheck,
  Video
} from '@lucide/vue'
import registrationApi from '@/api/registrations'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const activeTab = ref('registrations')
const refreshing = ref(false)
const stats = ref({})
const registrations = ref([])
const claims = ref([])
const expiring = ref([])

const loading = reactive({
  stats: false,
  registrations: false,
  claims: false,
  expiring: false,
})

const statusUpdating = reactive({
  registration: null,
  claim: null,
})

const registrationFilters = reactive({
  status: 'all',
})

const claimFilters = reactive({
  status: 'all',
})

const registrationPagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const claimPagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
})

const canEdit = computed(() => authStore.hasPermission('product:edit'))

const statItems = computed(() => [
  {
    key: 'total',
    label: '总注册',
    value: stats.value.total_count ?? registrations.value.length,
    icon: ShieldCheck,
    tone: 'blue',
  },
  {
    key: 'active',
    label: '有效保修',
    value: stats.value.active_count ?? 0,
    icon: ShieldCheck,
    tone: 'green',
  },
  {
    key: 'expired',
    label: '已过期',
    value: stats.value.expired_count ?? 0,
    icon: Clock3,
    tone: 'amber',
  },
  {
    key: 'claims',
    label: '当前申请',
    value: claimPagination.total,
    icon: FileWarning,
    tone: 'coral',
  },
])

const fetchStats = async () => {
  loading.stats = true
  try {
    stats.value = await registrationApi.getStats()
  } finally {
    loading.stats = false
  }
}

const fetchRegistrations = async () => {
  loading.registrations = true
  try {
    const response = await registrationApi.listRegistrations({
      page: registrationPagination.page,
      page_size: registrationPagination.pageSize,
      status: registrationFilters.status === 'all' ? undefined : registrationFilters.status,
    })
    registrations.value = response.data
    registrationPagination.total = response.pagination.total || 0
  } finally {
    loading.registrations = false
  }
}

const fetchClaims = async () => {
  loading.claims = true
  try {
    const response = await registrationApi.listWarrantyClaims({
      page: claimPagination.page,
      page_size: claimPagination.pageSize,
      status: claimFilters.status === 'all' ? undefined : claimFilters.status,
    })
    claims.value = response.data
    claimPagination.total = response.pagination.total || 0
  } finally {
    loading.claims = false
  }
}

const fetchExpiring = async () => {
  loading.expiring = true
  try {
    expiring.value = await registrationApi.listExpiringWarranties(30)
  } finally {
    loading.expiring = false
  }
}

const refreshCurrent = async () => {
  refreshing.value = true
  try {
    await fetchStats()
    if (activeTab.value === 'claims') await fetchClaims()
    else if (activeTab.value === 'expiring') await fetchExpiring()
    else await fetchRegistrations()
  } finally {
    refreshing.value = false
  }
}

const updateRegistrationStatus = async (registration, status) => {
  if (!registration?.id || registration.status === status) return
  statusUpdating.registration = registration.id
  try {
    await registrationApi.updateRegistrationStatus(registration.id, status)
    registration.status = status
    toast.success('注册状态已更新')
    await fetchStats()
  } finally {
    statusUpdating.registration = null
  }
}

const updateClaimStatus = async (claim, status) => {
  if (!claim?.id || claim.status === status) return
  statusUpdating.claim = claim.id
  try {
    await registrationApi.updateWarrantyClaimStatus(claim.id, status)
    claim.status = status
    claim.processed_at = new Date().toISOString()
    toast.success('保修申请状态已更新')
  } finally {
    statusUpdating.claim = null
  }
}

const productName = (product) => product?.name || product?.sku || '未关联商品'

const registrationProductName = (claim) => {
  if (claim?.registration?.product) return productName(claim.registration.product)
  if (claim?.registration_id) return `Registration #${claim.registration_id}`
  return '未绑定注册记录'
}

const userName = (user) => {
  if (!user) return '未关联用户'
  const fullName = [user.first_name, user.last_name].filter(Boolean).join(' ').trim()
  return fullName || user.username || user.email || `User #${user.id}`
}

const issueTypeLabel = (issueType) => {
  const labels = {
    warranty: '保修问题',
    defect: '质量缺陷',
    damage: '损坏',
    malfunction: '功能异常',
  }
  return labels[issueType] || issueType || '-'
}

const registrationStatusLabel = (status) => {
  const labels = {
    active: '有效',
    expired: '已过期',
    claimed: '已申请',
    cancelled: '已取消',
  }
  return labels[status] || status || '-'
}

const registrationStatusTone = (status) => {
  const tones = {
    active: 'green',
    expired: 'amber',
    claimed: 'blue',
    cancelled: 'gray',
  }
  return tones[status] || 'gray'
}

const claimImages = (claim) => {
  if (!claim?.images) return []
  try {
    const parsed = JSON.parse(claim.images)
    return Array.isArray(parsed) ? parsed.filter(Boolean) : []
  } catch {
    return []
  }
}

const formatDate = (value) => value ? new Date(value).toLocaleDateString('zh-CN') : '-'
const formatDateTime = (value) => value ? new Date(value).toLocaleString('zh-CN') : '-'

watch(
  () => [registrationFilters.status, registrationPagination.page, registrationPagination.pageSize],
  async ([, page], [, previousPage] = []) => {
    if (previousPage !== undefined && registrationPagination.page !== page) return
    await fetchRegistrations()
  }
)

watch(
  () => registrationFilters.status,
  () => {
    registrationPagination.page = 1
  }
)

watch(
  () => [claimFilters.status, claimPagination.page, claimPagination.pageSize],
  async ([, page], [, previousPage] = []) => {
    if (previousPage !== undefined && claimPagination.page !== page) return
    await fetchClaims()
  }
)

watch(
  () => claimFilters.status,
  () => {
    claimPagination.page = 1
  }
)

watch(activeTab, async (tab) => {
  if (tab === 'claims' && claims.value.length === 0) await fetchClaims()
  if (tab === 'expiring' && expiring.value.length === 0) await fetchExpiring()
})

onMounted(async () => {
  await Promise.all([
    fetchStats(),
    fetchRegistrations(),
    fetchClaims(),
    fetchExpiring(),
  ])
})
</script>
