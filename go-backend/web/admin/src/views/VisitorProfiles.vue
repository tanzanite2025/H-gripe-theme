<template>
  <div class="space-y-4">
    <AdminPageHeader
      title="访客画像"
      description="只读查看 Public Chat、购物车会话、邮箱、语言和粗粒度地区的统一事实源"
    >
      <template #actions>
        <Button variant="outline" size="sm" class="rounded-full font-black uppercase tracking-wider" :disabled="loading" @click="refreshProfiles">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
          刷新
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />

    <VisitorProfileFilterPanel
      :filters="filters"
      :loading="loading"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <section class="grid gap-4 2xl:grid-cols-[minmax(0,1fr)_380px]">
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
      />
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import {
  Activity,
  Fingerprint,
  Mail,
  MapPin,
  RefreshCw,
  ShoppingCart,
  UserRound
} from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import VisitorProfileDetailPanel from '@/components/admin/visitor/VisitorProfileDetailPanel.vue'
import VisitorProfileFilterPanel from '@/components/admin/visitor/VisitorProfileFilterPanel.vue'
import VisitorProfileTablePanel from '@/components/admin/visitor/VisitorProfileTablePanel.vue'
import { Button } from '@/components/ui/button'
import axios from '@/utils/axios'

const loading = ref(false)
const profiles = ref([])
const selectedProfile = ref(null)
const stats = ref({})
const filters = reactive({
  search: '',
  identity: 'all',
  email: 'all',
  cartSession: 'all',
  customerServiceVisitor: 'all',
  lastSeen: 'all',
  countryCode: '',
  locale: ''
})
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })

const apiData = (response) => response.data?.data ?? response.data ?? {}
const formatDate = (dateString) => dateString ? new Date(dateString).toLocaleString('zh-CN') : '-'

const statItems = computed(() => [
  { key: 'total', label: '画像总数', value: stats.value.total || 0, icon: Fingerprint, tone: 'gray' },
  { key: 'email', label: '已采集邮箱', value: stats.value.email_count || 0, icon: Mail, tone: 'green' },
  { key: 'cart', label: '已绑定购物车', value: stats.value.cart_linked_count || 0, icon: ShoppingCart, tone: 'blue' },
  { key: 'chat', label: '已绑定聊天', value: stats.value.customer_service_count || 0, icon: UserRound, tone: 'amber' },
  { key: 'region', label: '已采集地区', value: stats.value.region_count || 0, icon: MapPin, tone: 'green' },
  { key: 'recent', label: '24小时活跃', value: stats.value.recent_24h_count || 0, icon: Activity, tone: 'blue' }
])

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

const refreshProfiles = () => Promise.all([fetchProfiles(), fetchStats()])
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
    countryCode: '',
    locale: ''
  })
  pagination.page = 1
  fetchProfiles()
}
const updatePage = (page) => {
  pagination.page = page
  fetchProfiles()
}
const updatePageSize = (pageSize) => {
  pagination.pageSize = pageSize
  pagination.page = 1
  fetchProfiles()
}

onMounted(refreshProfiles)
</script>
