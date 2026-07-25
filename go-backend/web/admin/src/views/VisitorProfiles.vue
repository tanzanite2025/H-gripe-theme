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

    <AdminFilterPanel>
      <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(240px,1.5fr)_120px_120px_120px_120px_120px_120px_auto]" @submit.prevent="applyFilters">
        <label class="space-y-1 block md:col-span-2 xl:col-span-1">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">SEARCH / 搜索</span>
          <div class="relative">
            <Search class="pointer-events-none absolute left-3 top-1/2 size-3.5 -translate-y-1/2 text-muted-foreground/60" />
            <Input v-model="filters.search" class="h-9 pl-9" placeholder="邮箱、用户 ID、visitor hash、cart session、地区" />
          </div>
        </label>

        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">IDENTITY / 身份</span>
          <Select v-model="filters.identity">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部</SelectItem>
              <SelectItem value="account">会员</SelectItem>
              <SelectItem value="anonymous">匿名</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <TriStateSelect v-model="filters.email" label="EMAIL / 邮箱" yes-label="已采集" no-label="未采集" />
        <TriStateSelect v-model="filters.cartSession" label="CART / 购物车" yes-label="已绑定" no-label="未绑定" />
        <TriStateSelect v-model="filters.customerServiceVisitor" label="CHAT / 聊天" yes-label="已绑定" no-label="未绑定" />

        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">LAST SEEN / 活跃</span>
          <Select v-model="filters.lastSeen">
            <SelectTrigger class="h-9 w-full"><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部</SelectItem>
              <SelectItem value="24h">24小时</SelectItem>
              <SelectItem value="7d">7天</SelectItem>
              <SelectItem value="30d">30天</SelectItem>
            </SelectContent>
          </Select>
        </label>

        <label class="space-y-1 block">
          <span class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block">COUNTRY / 国家</span>
          <Input v-model="filters.countryCode" class="h-9 font-mono uppercase" placeholder="US" maxlength="8" />
        </label>

        <label class="space-y-1 block">
          <span class="block text-[10px] font-black uppercase tracking-widest text-transparent select-none">ACTION</span>
          <div class="flex items-center gap-2">
            <Button type="submit" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" :disabled="loading">
              <Search class="size-3.5" />
              查询
            </Button>
            <Button type="button" variant="outline" class="h-9 rounded-full px-3 font-black text-xs uppercase tracking-wider" @click="resetFilters">
              <RotateCcw class="size-3.5" />
              重置
            </Button>
          </div>
        </label>
      </form>
    </AdminFilterPanel>

    <section class="grid gap-4 2xl:grid-cols-[minmax(0,1fr)_380px]">
      <AdminTablePanel :loading="loading">
        <Table class="min-w-[1280px]">
          <TableHeader>
            <TableRow>
              <TableHead class="w-20">ID</TableHead>
              <TableHead class="w-24">身份</TableHead>
              <TableHead>联系信息</TableHead>
              <TableHead class="w-56">地区/语言</TableHead>
              <TableHead class="w-72">绑定事实</TableHead>
              <TableHead class="w-44">指纹状态</TableHead>
              <TableHead class="w-44">最后活跃</TableHead>
              <TableHead class="w-44">创建时间</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableEmpty v-if="profiles.length === 0" :colspan="8">
              <div class="flex flex-col items-center text-muted-foreground">
                <Fingerprint class="mb-2 size-7 opacity-55" />
                <span class="text-xs">暂无访客画像</span>
              </div>
            </TableEmpty>

            <TableRow
              v-for="profile in profiles"
              :key="profile.id"
              class="cursor-pointer"
              :class="selectedProfile?.id === profile.id ? 'bg-primary/5' : ''"
              @click="selectedProfile = profile"
            >
              <TableCell class="font-mono text-xs text-muted-foreground">#{{ profile.id }}</TableCell>
              <TableCell>
                <AdminStatusBadge :tone="profile.identity === 'account' ? 'green' : 'amber'">
                  {{ profile.identity === 'account' ? '会员' : '匿名' }}
                </AdminStatusBadge>
                <span v-if="profile.user_id" class="mt-1 block font-mono text-[11px] text-muted-foreground">UID {{ profile.user_id }}</span>
              </TableCell>
              <TableCell>
                <div class="min-w-0">
                  <p class="truncate font-medium">{{ profile.email || '未采集邮箱' }}</p>
                  <p class="mt-1 text-[11px] text-muted-foreground">来源：{{ profile.email_source || 'not_captured' }}</p>
                </div>
              </TableCell>
              <TableCell>
                <p class="truncate text-xs font-bold">{{ profile.region_label || '未采集地区' }}</p>
                <p class="mt-1 font-mono text-[11px] text-muted-foreground">{{ profile.locale || 'no-locale' }}</p>
              </TableCell>
              <TableCell>
                <div class="flex flex-wrap gap-1.5">
                  <AdminStatusBadge :tone="profile.has_customer_service_visitor ? 'green' : 'gray'">Public Chat</AdminStatusBadge>
                  <AdminStatusBadge :tone="profile.has_cart_session ? 'blue' : 'gray'">Cart</AdminStatusBadge>
                  <AdminStatusBadge :tone="profile.has_email ? 'green' : 'gray'">Email</AdminStatusBadge>
                </div>
                <p class="mt-2 break-all font-mono text-[11px] text-muted-foreground">
                  {{ profile.customer_service_visitor_hash_preview || 'no-chat-visitor' }}
                </p>
              </TableCell>
              <TableCell>
                <div class="flex flex-wrap gap-1.5">
                  <AdminStatusBadge :tone="profile.has_ip_fingerprint ? 'green' : 'gray'">IP hash</AdminStatusBadge>
                  <AdminStatusBadge :tone="profile.has_user_agent_fingerprint ? 'green' : 'gray'">UA hash</AdminStatusBadge>
                </div>
              </TableCell>
              <TableCell class="text-xs text-muted-foreground">{{ formatDate(profile.last_seen_at) }}</TableCell>
              <TableCell class="text-xs text-muted-foreground">{{ formatDate(profile.created_at) }}</TableCell>
            </TableRow>
          </TableBody>
        </Table>

        <template #footer>
          <AdminPagination
            :page="pagination.page"
            :page-size="pagination.pageSize"
            :total="pagination.total"
            @update:page="updatePage"
            @update:page-size="updatePageSize"
          />
        </template>
      </AdminTablePanel>

      <Card class="overflow-hidden py-0">
        <CardHeader class="border-b bg-muted/30 px-4 py-3">
          <CardTitle class="flex items-center gap-2">
            <Fingerprint class="size-4 text-primary" />
            画像详情
          </CardTitle>
          <CardDescription>只展示已采集事实，不猜测邮箱、地区或身份</CardDescription>
        </CardHeader>
        <CardContent class="space-y-4 p-4">
          <div v-if="!selectedProfile" class="flex min-h-72 flex-col items-center justify-center text-center text-muted-foreground">
            <UserRound class="mb-2 size-8 opacity-55" />
            <p class="text-xs leading-6">从左侧列表选择一个访客画像查看详情。</p>
          </div>

          <template v-else>
            <section class="rounded-2xl border p-3">
              <div class="mb-3 flex items-center justify-between gap-2">
                <h3 class="text-xs font-black uppercase tracking-wider">身份</h3>
                <AdminStatusBadge :tone="selectedProfile.identity === 'account' ? 'green' : 'amber'">
                  {{ selectedProfile.identity === 'account' ? '会员账号' : '匿名访客' }}
                </AdminStatusBadge>
              </div>
              <dl class="grid grid-cols-2 gap-2 text-xs">
                <DetailItem label="Profile ID">#{{ selectedProfile.id }}</DetailItem>
                <DetailItem label="User ID">{{ selectedProfile.user_id || '-' }}</DetailItem>
                <DetailItem label="Email" class="col-span-2">{{ selectedProfile.email || '未采集' }}</DetailItem>
                <DetailItem label="Email Source" class="col-span-2">{{ selectedProfile.email_source || 'not_captured' }}</DetailItem>
              </dl>
            </section>

            <section class="rounded-2xl border p-3">
              <h3 class="mb-3 text-xs font-black uppercase tracking-wider">绑定事实</h3>
              <div class="space-y-2 text-xs">
                <FactRow label="Public Chat visitor" :active="selectedProfile.has_customer_service_visitor">
                  {{ selectedProfile.customer_service_visitor_hash_preview || '未绑定' }}
                </FactRow>
                <FactRow label="Cart session" :active="selectedProfile.has_cart_session">
                  <span class="break-all font-mono">{{ selectedProfile.cart_session_id || '未绑定' }}</span>
                </FactRow>
                <FactRow label="Locale" :active="Boolean(selectedProfile.locale)">
                  {{ selectedProfile.locale || '未采集' }} / {{ selectedProfile.locale_source || 'not_captured' }}
                </FactRow>
              </div>
            </section>

            <section class="rounded-2xl border p-3">
              <h3 class="mb-3 text-xs font-black uppercase tracking-wider">地区与采集状态</h3>
              <dl class="grid grid-cols-2 gap-2 text-xs">
                <DetailItem label="Country">{{ selectedProfile.country_code || '-' }}</DetailItem>
                <DetailItem label="Timezone">{{ selectedProfile.timezone || '-' }}</DetailItem>
                <DetailItem label="Region" class="col-span-2">{{ selectedProfile.region || '-' }}</DetailItem>
                <DetailItem label="City" class="col-span-2">{{ selectedProfile.city || '-' }}</DetailItem>
              </dl>
              <div class="mt-3 flex flex-wrap gap-1.5">
                <AdminStatusBadge :tone="selectedProfile.has_ip_fingerprint ? 'green' : 'gray'">IP fingerprint</AdminStatusBadge>
                <AdminStatusBadge :tone="selectedProfile.has_user_agent_fingerprint ? 'green' : 'gray'">User-Agent fingerprint</AdminStatusBadge>
              </div>
            </section>

            <section class="rounded-2xl border p-3">
              <h3 class="mb-3 text-xs font-black uppercase tracking-wider">时间</h3>
              <dl class="grid gap-2 text-xs">
                <DetailItem label="Last Seen">{{ formatDate(selectedProfile.last_seen_at) }}</DetailItem>
                <DetailItem label="Created">{{ formatDate(selectedProfile.created_at) }}</DetailItem>
                <DetailItem label="Updated">{{ formatDate(selectedProfile.updated_at) }}</DetailItem>
              </dl>
            </section>
          </template>
        </CardContent>
      </Card>
    </section>
  </div>
</template>

<script setup>
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import {
  Activity,
  Fingerprint,
  Mail,
  MapPin,
  RefreshCw,
  RotateCcw,
  Search,
  ShoppingCart,
  UserRound
} from '@lucide/vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminPagination from '@/components/admin/AdminPagination.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import AdminStatusBadge from '@/components/admin/AdminStatusBadge.vue'
import AdminTablePanel from '@/components/admin/AdminTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Table, TableBody, TableCell, TableEmpty, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import axios from '@/utils/axios'

const DetailItem = defineComponent({
  props: {
    label: { type: String, required: true }
  },
  setup(props, { slots, attrs }) {
    return () => h('div', { class: ['space-y-1 rounded-xl border p-2', attrs.class] }, [
      h('dt', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70' }, props.label),
      h('dd', { class: 'break-words font-bold text-foreground' }, slots.default ? slots.default() : '-')
    ])
  }
})

const FactRow = defineComponent({
  props: {
    label: { type: String, required: true },
    active: { type: Boolean, default: false }
  },
  setup(props, { slots }) {
    return () => h('div', { class: 'rounded-xl border p-2' }, [
      h('div', { class: 'mb-1 flex items-center justify-between gap-2' }, [
        h('span', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70' }, props.label),
        h(AdminStatusBadge, { tone: props.active ? 'green' : 'gray' }, { default: () => props.active ? 'captured' : 'missing' })
      ]),
      h('p', { class: 'break-words font-mono text-[11px] text-foreground' }, slots.default ? slots.default() : '-')
    ])
  }
})

const TriStateSelect = defineComponent({
  props: {
    modelValue: { type: String, required: true },
    label: { type: String, required: true },
    yesLabel: { type: String, default: '是' },
    noLabel: { type: String, default: '否' }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    return () => h('label', { class: 'space-y-1 block' }, [
      h('span', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70 block' }, props.label),
      h(Select, {
        modelValue: props.modelValue,
        'onUpdate:modelValue': (value) => emit('update:modelValue', value)
      }, {
        default: () => [
          h(SelectTrigger, { class: 'h-9 w-full' }, { default: () => h(SelectValue) }),
          h(SelectContent, {}, {
            default: () => [
              h(SelectItem, { value: 'all' }, { default: () => '全部' }),
              h(SelectItem, { value: 'yes' }, { default: () => props.yesLabel }),
              h(SelectItem, { value: 'no' }, { default: () => props.noLabel })
            ]
          })
        ]
      })
    ])
  }
})

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
