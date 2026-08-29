<template>
  <div class="flex min-h-full flex-col gap-3">
    <AdminPageHeader
      class="shrink-0"
      title="访客画像"
      description="只读查看 Public Chat、购物车会话、邮箱、语言和粗粒度地区的统一事实源"
    >
      <template #actions>
        <Button
          v-if="canViewGlobalIPBlocks"
          variant="outline"
          size="sm"
          class="font-black uppercase tracking-wider"
          @click="openGlobalIPBlockDialog"
        >
          <ShieldAlert class="size-3.5" />
          IP 封禁规则
        </Button>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" :disabled="currentLoading" @click="cleanupCurrent">
          <Trash2 class="size-3.5" />
          {{ cleanupLabel }}
        </Button>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" :disabled="currentLoading" @click="refreshCurrent">
          <RefreshCw :class="['size-3.5', { 'animate-spin': currentLoading }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <Tabs :model-value="activeTab" class="min-h-0 flex-1">
      <TabsContent value="profiles" class="flex flex-none flex-col gap-3 overflow-visible">
        <AdminStatsGrid class="shrink-0" :items="statItems" />

        <div class="shrink-0">
          <VisitorProfileFilterPanel
            :filters="filters"
            :loading="loading"
            @apply="applyFilters"
            @reset="resetFilters"
          />
        </div>

        <section class="grid min-h-[520px] flex-none grid-rows-[minmax(0,1.15fr)_minmax(0,0.85fr)] gap-4 overflow-visible 2xl:grid-cols-[minmax(0,1fr)_380px] 2xl:grid-rows-[minmax(0,1fr)]">
          <VisitorProfileTablePanel
            :loading="loading"
            :profiles="profiles"
            :selected-profile="selectedProfile"
            :pagination="pagination"
            :format-date="formatDate"
            @select-profile="selectedProfile = $event"
            @update-page="updatePage"
            @update-page-size="updatePageSize"
          />

          <VisitorProfileDetailPanel
            :selected-profile="selectedProfile"
            :format-date="formatDate"
            :block-ip="canManageGlobalIPBlocks ? blockProfileIP : undefined"
            :unblock-ip="canManageGlobalIPBlocks ? unblockProfileIP : undefined"
          />
        </section>
      </TabsContent>

      <TabsContent value="risk" class="min-h-0 flex flex-col gap-3 overflow-hidden">
        <AdminStatsGrid class="shrink-0" :items="riskStatItems" />

        <AdminFilterPanel class="shrink-0">
          <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(260px,1fr)_140px_140px_140px_auto]" @submit.prevent="applyRiskFilters">
            <label class="block space-y-1 md:col-span-2 xl:col-span-1">
              <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">SEARCH / 搜索</span>
              <div class="relative">
                <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
                <Input v-model="riskFilters.search" class="h-9 pl-9" placeholder="IP hash、UA hash、国家、ID" />
              </div>
            </label>

            <AdminFilterSelect v-model="riskFilters.riskLevel" label="LEVEL / 等级" :options="riskLevelOptions" />
            <AdminFilterSelect v-model="riskFilters.dayRange" label="RANGE / 时间" :options="riskDayOptions" />

            <label class="block space-y-1">
              <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">MIN SCORE / 最低分</span>
              <Input v-model="riskFilters.minRiskScore" class="h-9 font-mono" inputmode="numeric" placeholder="0" />
            </label>

            <label class="block space-y-1">
              <span class="block select-none text-[10px] font-black uppercase tracking-widest text-transparent">ACTION</span>
              <div class="flex items-center gap-2">
                <Button type="submit" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" :disabled="riskLoading">
                  <Search class="size-3.5" />
                  查询
                </Button>
                <Button type="button" variant="outline" class="h-9 rounded-full px-3 text-xs font-black uppercase tracking-wider" @click="resetRiskFilters">
                  <RotateCcw class="size-3.5" />
                  重置
                </Button>
              </div>
            </label>
          </form>
        </AdminFilterPanel>

        <VisitorRiskFactsPanel
          class="min-h-0 flex-1"
          :loading="riskLoading"
          :facts="riskFacts"
          :pagination="riskPagination"
          :format-date="formatDate"
          :save-decision="submitRiskDecision"
          @update-page="updateRiskPage"
          @update-page-size="updateRiskPageSize"
        />
      </TabsContent>

    </Tabs>

    <VisitorIPBlockRulesDialog
      v-model:open="globalIPBlockDialogOpen"
      :rules="globalIPBlockRules"
      :pagination="globalIPBlockPagination"
      :loading="globalIPBlockLoading"
      :saving="globalIPBlockSaving"
      :error="globalIPBlockError"
      :can-create="canManageGlobalIPBlocks"
      :can-manage="canManageGlobalIPBlocks"
      :format-date="formatDate"
      @create="createGlobalIPBlock"
      @disable="disableGlobalIPBlock"
      @update:page="updateGlobalIPBlockPage"
      @update:page-size="updateGlobalIPBlockPageSize"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { toast } from 'vue-sonner'
import {
  Activity,
  AlertTriangle,
  Bot,
  Cookie,
  Fingerprint,
  Mail,
  MapPin,
  RefreshCw,
  RotateCcw,
  Search,
  ShieldAlert,
  ShoppingCart,
  Trash2,
  UserRound
} from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminFilterSelect from '@/components/admin/AdminFilterSelect.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import VisitorIPBlockRulesDialog from '@/components/admin/visitor/VisitorIPBlockRulesDialog.vue'
import VisitorProfileDetailPanel from '@/components/admin/visitor/VisitorProfileDetailPanel.vue'
import VisitorProfileFilterPanel from '@/components/admin/visitor/VisitorProfileFilterPanel.vue'
import VisitorProfileTablePanel from '@/components/admin/visitor/VisitorProfileTablePanel.vue'
import VisitorRiskFactsPanel from '@/components/admin/visitor/VisitorRiskFactsPanel.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import { useRouteTab } from '@/composables/useRouteTab'
import axios from '@/utils/axios'
import { useAuthStore } from '@/stores/auth'
import type {
  VisitorGlobalIPBlockPayload,
  VisitorIPBlockPayload,
  VisitorIPBlockRule,
  VisitorRiskDecisionPayload,
  VisitorRiskFact,
  VisitorProfile,
} from '@/components/admin/visitor/visitorTypes'

interface VisitorProfileStats {
  total?: number
  active_count?: number
  candidate_count?: number
  email_count?: number
  cart_linked_count?: number
  customer_service_count?: number
  region_count?: number
  recent_24h_count?: number
}

interface VisitorRiskStats {
  total_facts?: number
  watch_count?: number
  suspicious_count?: number
  request_count?: number
  invalid_request_count?: number
  bot_like_user_agent_count?: number
  no_cookie_request_count?: number
  meaningful_action_count?: number
}

const activeTab = useRouteTab({
  defaultValue: 'profiles',
  values: ['profiles', 'risk'],
  routes: {
    profiles: 'VisitorProfilesProfiles',
    risk: 'VisitorProfilesRisk',
  },
})
const authStore = useAuthStore()
const loading = ref(false)
const riskLoading = ref(false)
const globalIPBlockDialogOpen = ref(false)
const globalIPBlockLoading = ref(false)
const globalIPBlockSaving = ref(false)
const globalIPBlockError = ref('')
const profiles = ref<VisitorProfile[]>([])
const globalIPBlockRules = ref<VisitorIPBlockRule[]>([])
const globalIPBlockPagination = reactive({ page: 1, pageSize: 20, total: 0 })
const riskFacts = ref<VisitorRiskFact[]>([])
const selectedProfile = ref<VisitorProfile | null>(null)
const stats = ref<VisitorProfileStats>({})
const riskStats = ref<VisitorRiskStats>({})
const filters = reactive({
  search: '',
  identity: 'all',
  email: 'all',
  cartSession: 'all',
  customerServiceVisitor: 'all',
  lastSeen: 'all',
  lastMeaningful: 'all',
  status: 'active',
  countryCode: '',
  locale: ''
})
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const riskFilters = reactive({
  search: '',
  riskLevel: 'all',
  dayRange: '7d',
  minRiskScore: ''
})
const riskPagination = reactive({ page: 1, pageSize: 20, total: 0 })

const apiData = (response: any) => response.data?.data ?? response.data ?? {}
const formatDate = (dateString?: string | number | Date | null) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'
const currentLoading = computed(() => activeTab.value === 'risk' ? riskLoading.value : loading.value)
const cleanupLabel = computed(() => activeTab.value === 'risk' ? '清理风险数据' : '清理过期画像')
const canViewGlobalIPBlocks = computed(() => (
  authStore.hasRole('admin') ||
  String(authStore.user?.role || '').trim().toLowerCase() === 'admin' ||
  authStore.hasPermission('ticket:view') ||
  authStore.hasPermission('settings:view')
))
const canManageGlobalIPBlocks = computed(() => (
  authStore.hasRole('admin') ||
  String(authStore.user?.role || '').trim().toLowerCase() === 'admin'
))

const statItems = computed(() => [
  { key: 'total', label: '画像总数', value: stats.value.total || 0, icon: Fingerprint, tone: 'gray' },
  { key: 'active', label: '有效画像', value: stats.value.active_count || 0, icon: Activity, tone: 'green' },
  { key: 'candidate', label: '候选画像', value: stats.value.candidate_count || 0, icon: Fingerprint, tone: 'amber' },
  { key: 'email', label: '已采集邮箱', value: stats.value.email_count || 0, icon: Mail, tone: 'green' },
  { key: 'cart', label: '已绑定购物车', value: stats.value.cart_linked_count || 0, icon: ShoppingCart, tone: 'blue' },
  { key: 'chat', label: '已绑定聊天', value: stats.value.customer_service_count || 0, icon: UserRound, tone: 'amber' },
  { key: 'region', label: '已采集地区', value: stats.value.region_count || 0, icon: MapPin, tone: 'green' },
  { key: 'recent', label: '24小时活跃', value: stats.value.recent_24h_count || 0, icon: Activity, tone: 'blue' }
])

const riskStatItems = computed(() => [
  { key: 'facts', label: '风险事实', value: riskStats.value.total_facts || 0, icon: ShieldAlert, tone: 'gray' },
  { key: 'watch', label: '观察', value: riskStats.value.watch_count || 0, icon: AlertTriangle, tone: 'amber' },
  { key: 'suspicious', label: '可疑', value: riskStats.value.suspicious_count || 0, icon: ShieldAlert, tone: 'coral' },
  { key: 'requests', label: '请求量', value: riskStats.value.request_count || 0, icon: Activity, tone: 'blue' },
  { key: 'invalid', label: '异常请求', value: riskStats.value.invalid_request_count || 0, icon: AlertTriangle, tone: 'amber' },
  { key: 'bot', label: 'Bot UA', value: riskStats.value.bot_like_user_agent_count || 0, icon: Bot, tone: 'coral' },
  { key: 'cookie', label: '无 Cookie', value: riskStats.value.no_cookie_request_count || 0, icon: Cookie, tone: 'gray' },
  { key: 'meaningful', label: '有效动作', value: riskStats.value.meaningful_action_count || 0, icon: Activity, tone: 'green' }
])

const riskLevelOptions = [
  { label: '全部', value: 'all' },
  { label: '正常', value: 'normal' },
  { label: '观察', value: 'watch' },
  { label: '可疑', value: 'suspicious' },
  { label: '封禁候选', value: 'block' },
]
const riskDayOptions = [
  { label: '全部', value: 'all' },
  { label: '24小时', value: '24h' },
  { label: '7天', value: '7d' },
  { label: '30天', value: '30d' },
  { label: '90天', value: '90d' },
]

const fetchProfiles = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/admin/customer-service/visitor-profiles', {
      params: {
        page: pagination.page,
        page_size: pagination.pageSize,
        search: filters.search.trim() || undefined,
        identity: filters.identity !== 'all' ? filters.identity : undefined,
        email: filters.email !== 'all' ? filters.email : undefined,
        cart_session: filters.cartSession !== 'all' ? filters.cartSession : undefined,
        customer_service_visitor: filters.customerServiceVisitor !== 'all' ? filters.customerServiceVisitor : undefined,
        last_seen: filters.lastSeen !== 'all' ? filters.lastSeen : undefined,
        last_meaningful: filters.lastMeaningful !== 'all' ? filters.lastMeaningful : undefined,
        status: filters.status !== 'active' ? filters.status : undefined,
        country_code: filters.countryCode.trim() || undefined,
        locale: filters.locale.trim() || undefined
      }
    })
    const data = apiData(response)
    profiles.value = data.profiles || []
    pagination.total = data.pagination?.total ?? profiles.value.length

    if (selectedProfile.value) {
      const refreshed = profiles.value.find((item) => Number(item.id) === Number(selectedProfile.value.id))
      selectedProfile.value = refreshed || null
    }
  } catch (error) {
    console.error('Failed to fetch visitor profiles:', error)
  } finally {
    loading.value = false
  }
}

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/customer-service/visitor-profiles/stats')
    const data = apiData(response)
    stats.value = data.stats || {}
  } catch (error) {
    console.error('Failed to fetch visitor profile stats:', error)
  }
}

const fetchGlobalIPBlockRules = async () => {
  if (!canViewGlobalIPBlocks.value) return
  globalIPBlockLoading.value = true
  globalIPBlockError.value = ''
  try {
    const response = await axios.get('/api/admin/security/ip-blocks', {
      params: {
        page: globalIPBlockPagination.page,
        page_size: globalIPBlockPagination.pageSize,
        status: 'all',
      },
    })
    const data = apiData(response)
    globalIPBlockRules.value = data.rules || []
    const responsePagination = data.pagination || {}
    globalIPBlockPagination.page = Number(responsePagination.page || globalIPBlockPagination.page)
    globalIPBlockPagination.pageSize = Number(responsePagination.page_size || globalIPBlockPagination.pageSize)
    globalIPBlockPagination.total = Number(responsePagination.total ?? globalIPBlockRules.value.length)

    const totalPages = Math.max(1, Math.ceil(globalIPBlockPagination.total / globalIPBlockPagination.pageSize))
    if (globalIPBlockRules.value.length === 0 && globalIPBlockPagination.total > 0 && globalIPBlockPagination.page > totalPages) {
      globalIPBlockPagination.page = totalPages
      await fetchGlobalIPBlockRules()
    }
  } catch (error) {
    globalIPBlockError.value = extractVisitorError(error, '全局 IP 封禁规则读取失败')
  } finally {
    globalIPBlockLoading.value = false
  }
}

const openGlobalIPBlockDialog = async () => {
  globalIPBlockDialogOpen.value = true
  globalIPBlockPagination.page = 1
  await fetchGlobalIPBlockRules()
}

const createGlobalIPBlock = async (payload: VisitorGlobalIPBlockPayload) => {
  if (!canManageGlobalIPBlocks.value) {
    globalIPBlockError.value = '当前账号没有管理员权限。'
    return
  }
  globalIPBlockSaving.value = true
  globalIPBlockError.value = ''
  try {
    const response = await axios.post('/api/admin/security/ip-blocks', payload)
    globalIPBlockPagination.page = 1
    await fetchGlobalIPBlockRules()
    globalIPBlockDialogOpen.value = false
    if (response.status === 202) {
      toast.warning(response.data?.message || '规则已落库，但当前实例的封禁缓存仍在刷新，暂不可视为已生效')
    } else {
      toast.success('全局 IP 封禁规则已创建')
    }
  } catch (error) {
    globalIPBlockError.value = extractVisitorError(error, '全局 IP 封禁失败')
    toast.error(globalIPBlockError.value)
  } finally {
    globalIPBlockSaving.value = false
  }
}

const disableGlobalIPBlock = async (rule: VisitorIPBlockRule) => {
  if (!canManageGlobalIPBlocks.value) {
    globalIPBlockError.value = '当前账号没有管理员权限。'
    return
  }
  if (!window.confirm(`确定解除全局 IP/CIDR 规则“${rule.cidr || rule.id}”吗？这会改变所有 API 请求的拦截范围。`)) return

  globalIPBlockSaving.value = true
  globalIPBlockError.value = ''
  try {
    const response = await axios.delete(`/api/admin/security/ip-blocks/${rule.id}`)
    await fetchGlobalIPBlockRules()
    if (response.status === 202) {
      toast.warning(response.data?.message || '规则已解除，但当前实例的封禁缓存仍在刷新')
    } else {
      toast.success('全局 IP 封禁规则已解除')
    }
  } catch (error) {
    globalIPBlockError.value = extractVisitorError(error, '全局 IP 封禁解除失败')
    toast.error(globalIPBlockError.value)
  } finally {
    globalIPBlockSaving.value = false
  }
}

const fetchRiskFacts = async () => {
  riskLoading.value = true
  try {
    const response = await axios.get('/api/admin/customer-service/visitor-risk-facts', {
      params: {
        page: riskPagination.page,
        page_size: riskPagination.pageSize,
        search: riskFilters.search.trim() || undefined,
        risk_level: riskFilters.riskLevel !== 'all' ? riskFilters.riskLevel : undefined,
        day_range: riskFilters.dayRange !== 'all' ? riskFilters.dayRange : undefined,
        min_risk_score: riskFilters.minRiskScore.trim() || undefined
      }
    })
    const data = apiData(response)
    riskFacts.value = data.facts || []
    riskPagination.total = data.pagination?.total ?? riskFacts.value.length
  } catch (error) {
    console.error('Failed to fetch visitor risk facts:', error)
  } finally {
    riskLoading.value = false
  }
}

const fetchRiskStats = async () => {
  try {
    const response = await axios.get('/api/admin/customer-service/visitor-risk-facts/stats', {
      params: {
        day_range: riskFilters.dayRange !== 'all' ? riskFilters.dayRange : undefined
      }
    })
    const data = apiData(response)
    riskStats.value = data.stats || {}
  } catch (error) {
    console.error('Failed to fetch visitor risk stats:', error)
  }
}

const submitRiskDecision = async (fact: VisitorRiskFact, payload: VisitorRiskDecisionPayload) => {
  await axios.post(`/api/admin/customer-service/visitor-risk-facts/${fact.id}/decision`, payload)
  await refreshRiskFacts()
}

const blockProfileIP = async (profile: VisitorProfile, payload: VisitorIPBlockPayload) => {
  const response = await axios.post(`/api/admin/customer-service/visitor-profiles/${profile.id}/ip-block`, payload)
  await refreshProfiles()
  if (response.status === 202) {
    toast.warning(response.data?.message || '画像规则已落库，但当前实例的封禁缓存仍在刷新，暂不可视为已生效')
  }
}

const unblockProfileIP = async (profile: VisitorProfile) => {
  const response = await axios.delete(`/api/admin/customer-service/visitor-profiles/${profile.id}/ip-block`)
  await refreshProfiles()
  if (response.status === 202) {
    toast.warning(response.data?.message || '画像规则已解除，但当前实例的封禁缓存仍在刷新')
  }
}

const refreshProfiles = () => Promise.all([fetchProfiles(), fetchStats()])
const refreshRiskFacts = () => Promise.all([fetchRiskFacts(), fetchRiskStats()])
const refreshCurrent = () => activeTab.value === 'risk' ? refreshRiskFacts() : refreshProfiles()
const cleanupExpiredProfiles = async () => {
  loading.value = true
  try {
    await axios.post('/api/admin/customer-service/visitor-profiles/cleanup')
    await Promise.all([fetchProfiles(), fetchStats()])
  } catch (error) {
    console.error('Failed to cleanup visitor profiles:', error)
  } finally {
    loading.value = false
  }
}
const cleanupExpiredRiskFacts = async () => {
  riskLoading.value = true
  try {
    await axios.post('/api/admin/customer-service/visitor-risk-facts/cleanup')
    await Promise.all([fetchRiskFacts(), fetchRiskStats()])
  } catch (error) {
    console.error('Failed to cleanup visitor risk facts:', error)
  } finally {
    riskLoading.value = false
  }
}
const cleanupCurrent = () => activeTab.value === 'risk' ? cleanupExpiredRiskFacts() : cleanupExpiredProfiles()
const applyFilters = () => {
  pagination.page = 1
  fetchProfiles()
}
const resetFilters = () => {
  Object.assign(filters, {
    search: '',
    identity: 'all',
    email: 'all',
    cartSession: 'all',
    customerServiceVisitor: 'all',
    lastSeen: 'all',
    lastMeaningful: 'all',
    status: 'active',
    countryCode: '',
    locale: ''
  })
  pagination.page = 1
  fetchProfiles()
}
const updatePage = (page: number) => {
  pagination.page = page
  fetchProfiles()
}
const updatePageSize = (pageSize: number) => {
  pagination.pageSize = pageSize
  pagination.page = 1
  fetchProfiles()
}
const updateGlobalIPBlockPage = (page: number) => {
  globalIPBlockPagination.page = page
  fetchGlobalIPBlockRules()
}
const updateGlobalIPBlockPageSize = (pageSize: number) => {
  globalIPBlockPagination.pageSize = pageSize
  globalIPBlockPagination.page = 1
  fetchGlobalIPBlockRules()
}
const applyRiskFilters = () => {
  riskPagination.page = 1
  refreshRiskFacts()
}
const resetRiskFilters = () => {
  Object.assign(riskFilters, {
    search: '',
    riskLevel: 'all',
    dayRange: '7d',
    minRiskScore: ''
  })
  riskPagination.page = 1
  refreshRiskFacts()
}
const updateRiskPage = (page: number) => {
  riskPagination.page = page
  fetchRiskFacts()
}
const updateRiskPageSize = (pageSize: number) => {
  riskPagination.pageSize = pageSize
  riskPagination.page = 1
  fetchRiskFacts()
}

interface VisitorErrorLike {
  message?: string
  response?: {
    data?: {
      error?: string
      message?: string
    }
  }
}

const extractVisitorError = (error: unknown, fallback: string): string => {
  const details = error as VisitorErrorLike
  return details.response?.data?.error || details.response?.data?.message || details.message || fallback
}

onMounted(() => {
  if (activeTab.value === 'risk') refreshRiskFacts()
  else refreshProfiles()
})
watch(activeTab, (tab) => {
  if (tab === 'risk' && riskFacts.value.length === 0) {
    refreshRiskFacts()
  }
})
</script>
