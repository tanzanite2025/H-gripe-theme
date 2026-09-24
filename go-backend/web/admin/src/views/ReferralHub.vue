<template>
  <div class="flex min-h-full flex-col gap-3">
    <AdminPageHeader title="推荐裂变台账" description="查看 v2 推荐归因、履约冷静期和影子风控信号">
      <template #actions>
        <Button v-if="activeTab === 'ledger'" variant="outline" size="sm" :disabled="loading" @click="fetchLedger">
          <RefreshCw :class="['size-3.5', { 'animate-spin': loading }]" />
          刷新
        </Button>
        <Button v-if="activeTab === 'ledger'" variant="outline" size="sm" :disabled="loading || exportLoading" @click="exportLedger">
          <Download class="size-3.5" /> 导出 CSV
        </Button>
        <Button v-else variant="outline" size="sm" :disabled="configLoading" @click="fetchConfig">
          <RefreshCw :class="['size-3.5', { 'animate-spin': configLoading }]" />
          刷新规则
        </Button>
      </template>
    </AdminPageHeader>

    <Tabs v-model="activeTab" class="min-h-0 gap-3">
      <TabsList class="h-11 w-full justify-start rounded-2xl bg-muted/50 p-1.5">
        <TabsTrigger value="ledger" class="max-w-48 gap-2 px-4 text-xs font-black normal-case tracking-normal">
          <ClipboardList class="size-3.5" /> 台账与风控审计
        </TabsTrigger>
        <TabsTrigger value="config" class="max-w-48 gap-2 px-4 text-xs font-black normal-case tracking-normal">
          <SlidersHorizontal class="size-3.5" /> 规则配置
        </TabsTrigger>
      </TabsList>

      <TabsContent value="ledger" class="min-h-0 space-y-3">
    <AdminStatsGrid :items="statItems" compact />

    <AdminFilterPanel>
      <form class="grid grid-cols-1 gap-3 md:grid-cols-2 xl:grid-cols-[minmax(220px,1fr)_160px_160px_160px_auto]" @submit.prevent="applyFilters">
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">搜索</span>
          <Input v-model="filters.keyword" class="h-9" placeholder="邀请码、邮箱或订单号" />
        </label>
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">状态</span>
          <select v-model="filters.status" class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm">
            <option value="">全部</option>
            <option v-for="option in statusOptions" :key="option.value" :value="option.value">{{ option.label }}</option>
          </select>
        </label>
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">开始日期（UTC）</span>
          <Input v-model="filters.from" type="date" :max="filters.to || undefined" class="h-9" />
        </label>
        <label class="block space-y-1">
          <span class="block text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">结束日期（UTC）</span>
          <Input v-model="filters.to" type="date" :min="filters.from || undefined" class="h-9" />
        </label>
        <div class="flex items-end gap-2">
          <Button type="submit" class="h-9">查询</Button>
          <Button type="button" variant="outline" class="h-9" @click="resetFilters">重置</Button>
        </div>
      </form>
    </AdminFilterPanel>

    <section class="min-h-0 overflow-hidden rounded-lg border border-dashed border-border/80 bg-card">
      <div v-if="error" class="p-6 text-sm text-destructive">{{ error }}</div>
      <div v-else-if="loading" class="p-6 text-sm text-muted-foreground">正在加载台账...</div>
      <div v-else-if="items.length === 0" class="p-10 text-center text-sm text-muted-foreground">暂无推荐记录</div>
      <div v-else class="overflow-x-auto">
        <table class="w-full min-w-[900px] text-sm">
          <thead class="border-b border-border/70 bg-muted/30 text-left text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">
            <tr>
              <th class="px-4 py-3">邀请码</th>
              <th class="px-4 py-3">推荐关系</th>
              <th class="px-4 py-3">首单</th>
              <th class="px-4 py-3">状态 / 冷静期</th>
              <th class="px-4 py-3">风控</th>
              <th class="px-4 py-3 text-right">奖励</th>
              <th class="px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-border/60">
            <tr v-for="item in items" :key="item.id" class="align-top">
              <td class="px-4 py-3 font-mono font-bold">{{ item.referral_code }}</td>
              <td class="px-4 py-3">
                <div class="font-medium">{{ item.referrer?.name || '—' }} → {{ item.referee?.name || '—' }}</div>
                <div class="mt-1 text-xs text-muted-foreground">{{ item.referee?.email_masked || item.referrer?.email_masked || '' }}</div>
              </td>
              <td class="px-4 py-3">
                <div class="font-medium">{{ item.order?.order_number || '等待首单' }}</div>
                <div v-if="item.order" class="mt-1 text-xs text-muted-foreground">{{ formatMoney(item.order.amount_minor, item.order.currency) }}</div>
              </td>
              <td class="px-4 py-3">
                <span :class="['inline-flex rounded-full px-2 py-0.5 text-[10px] font-black uppercase', statusClass(item.status)]">{{ statusLabel(item.status) }}</span>
                <div v-if="item.vesting_until" class="mt-1 text-xs text-muted-foreground">
                  {{ item.days_remaining > 0 ? `剩余 ${item.days_remaining} 天` : '已到期候选' }}
                </div>
              </td>
              <td class="px-4 py-3">
                <span v-if="item.risk_flags?.length" class="font-bold text-rose-700">{{ item.risk_flags.length }} 个标记</span>
                <span v-else class="text-emerald-700">无异常</span>
              </td>
              <td class="px-4 py-3 text-right font-mono font-bold">{{ item.reward_points }} pts</td>
              <td class="px-4 py-3 text-right">
                <Button variant="outline" size="sm" class="rounded-full" @click="openDetail(item.id)">
                  <Eye class="size-3.5" /> 查看
                </Button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="flex items-center justify-between border-t border-border/70 px-4 py-3 text-xs text-muted-foreground">
        <span>共 {{ pagination.total }} 条</span>
        <div class="flex items-center gap-2">
          <Button variant="outline" size="sm" :disabled="pagination.page <= 1 || loading" @click="updatePage(pagination.page - 1)">上一页</Button>
          <span>{{ pagination.page }}</span>
          <Button variant="outline" size="sm" :disabled="pagination.page * pagination.pageSize >= pagination.total || loading" @click="updatePage(pagination.page + 1)">下一页</Button>
        </div>
      </div>
    </section>
      </TabsContent>

      <TabsContent value="config" class="min-h-0">
        <section class="space-y-5 rounded-2xl border border-dashed border-border/80 bg-card p-5">
          <div class="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">Referral Incentive Policy</p>
              <h2 class="mt-1 text-base font-black tracking-tight">推荐返利经济模型与全局规则</h2>
              <p class="mt-1 text-xs leading-relaxed text-muted-foreground">发布会生成不可变的新版本。被推荐人积分在注册绑定时进入统一积分余额；推荐人订单奖励仍按订单履约规则结算。</p>
            </div>
            <div class="rounded-full border border-border/70 bg-muted/40 px-3 py-1 text-xs text-muted-foreground">
              当前版本 <span class="font-mono font-black text-foreground">v{{ config.version || '—' }}</span>
            </div>
          </div>
          <div v-if="configError" class="rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-700">{{ configError }}</div>
          <div v-if="configLoading" class="py-8 text-center text-sm text-muted-foreground">正在加载规则...</div>
          <form v-else class="space-y-5" @submit.prevent="saveConfig">
            <div class="grid gap-4 rounded-xl border border-border/70 p-4 md:grid-cols-3">
              <label class="flex items-center gap-3 text-sm font-bold md:col-span-3">
                <Switch v-model:checked="config.enabled" :disabled="!canEdit || configSaving" aria-label="启用推荐返利系统" />
                启用推荐返利系统
              </label>
              <label class="space-y-1 text-xs font-bold"><span>首单最低实付（分）</span><Input v-model.number="config.min_order_amount_minor" type="number" min="0" :disabled="!canEdit || configSaving" /></label>
              <label class="space-y-1 text-xs font-bold"><span>单用户月度上限</span><Input v-model.number="config.monthly_cap_per_referrer" type="number" min="1" :disabled="!canEdit || configSaving" /></label>
            </div>
            <div class="grid gap-4 rounded-xl border border-border/70 p-4 md:grid-cols-2">
              <label class="space-y-1 text-xs font-bold"><span>推荐人奖励积分</span><Input v-model.number="config.referrer_reward_points" type="number" min="0" :disabled="!canEdit || configSaving" /></label>
              <label class="space-y-1 text-xs font-bold"><span>被推荐人注册积分</span><Input v-model.number="config.referee_benefit_value" type="number" min="1" :disabled="!canEdit || configSaving" /></label>
              <p class="text-xs leading-relaxed text-muted-foreground md:col-span-2">推荐积分直接进入统一积分余额，消费时与账户内其他积分使用同一套规则。</p>
            </div>
            <div class="grid gap-4 rounded-xl border border-border/70 p-4 md:grid-cols-2 lg:grid-cols-4">
              <label class="space-y-1 text-xs font-bold"><span>妥投后冷静期（天）</span><Input v-model.number="config.vesting_period_days" type="number" min="1" :disabled="!canEdit || configSaving" /></label>
              <label class="space-y-1 text-xs font-bold"><span>未妥投兜底（天）</span><Input v-model.number="config.undelivered_fallback_days" type="number" min="1" :disabled="!canEdit || configSaving" /></label>
              <label class="space-y-1 text-xs font-bold"><span>归因 Cookie（天）</span><Input v-model.number="config.attribution_ttl_days" type="number" min="1" max="90" :disabled="!canEdit || configSaving" /></label>
              <label class="space-y-1 text-xs font-bold"><span>反作弊模式</span><select v-model="config.anti_fraud_mode" :disabled="!canEdit || configSaving" class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm"><option value="monitor">影子监控</option><option value="strict">严格阻断</option></select></label>
            </div>
            <div class="flex flex-wrap items-center justify-between gap-3 border-t border-border/70 pt-4">
              <p class="text-xs text-muted-foreground">发布时会校验当前版本 v{{ config.version }}，避免覆盖其他管理员刚发布的规则。</p>
              <Button type="submit" :disabled="!canEdit || configSaving || !config.version" class="rounded-full">
                <LoaderCircle v-if="configSaving" class="size-3.5 animate-spin" /><Save v-else class="size-3.5" />
                {{ configSaving ? '发布中' : '保存并发布新版本' }}
              </Button>
            </div>
          </form>
        </section>
      </TabsContent>
    </Tabs>

    <Dialog v-model:open="detailOpen">
      <DialogContent size="lg" class="max-h-[90dvh] overflow-y-auto">
        <DialogHeader><DialogTitle>推荐记录详情</DialogTitle><DialogDescription v-if="detail">记录 #{{ detail.item.id }} · {{ detail.item.referral_code }}</DialogDescription></DialogHeader>
        <div v-if="detailLoading" class="py-8 text-center text-sm text-muted-foreground">正在加载详情...</div>
        <div v-else-if="detail" class="space-y-4">
          <div class="grid gap-3 rounded-xl border border-border/70 p-4 text-sm md:grid-cols-3"><div><span class="text-xs text-muted-foreground">状态</span><p class="mt-1 font-bold">{{ statusLabel(detail.item.status) }}</p></div><div><span class="text-xs text-muted-foreground">推荐人</span><p class="mt-1 font-bold">{{ detail.item.referrer?.name || '—' }}</p></div><div><span class="text-xs text-muted-foreground">被推荐人</span><p class="mt-1 font-bold">{{ detail.item.referee?.name || '—' }}</p></div></div>
          <div><h3 class="mb-2 text-xs font-black uppercase tracking-widest text-muted-foreground/70">风控信号与证据维度</h3><div v-if="detail.item.risk_flags?.length" class="space-y-2"> <div v-for="(flag, index) in detail.item.risk_flags" :key="index" class="rounded-lg border border-rose-200 bg-rose-50 px-3 py-2 text-sm text-rose-800"><div>{{ flag.type || 'risk_signal' }} · {{ flag.level || 'unknown' }}<span v-if="flag.source" class="ml-2 text-xs opacity-70">{{ flag.source }}</span></div><div v-if="flag.dimensions?.length" class="mt-1 text-xs text-rose-700">命中维度：{{ flag.dimensions.join('、') }}</div></div></div><p v-else class="text-sm text-emerald-700">未发现风控标记</p><p class="mt-2 text-[11px] text-muted-foreground">敏感原始值仅保留哈希；此处展示命中的证据维度，不暴露地址、手机号、设备或支付凭据。</p></div>
          <div><h3 class="mb-2 text-xs font-black uppercase tracking-widest text-muted-foreground/70">状态迁移审计</h3><div v-if="detail.transitions.length" class="divide-y rounded-lg border border-border/70"> <div v-for="transition in detail.transitions" :key="transition.id" class="grid gap-1 px-3 py-2 text-xs md:grid-cols-[140px_1fr]"> <span class="font-mono text-muted-foreground">{{ formatDate(transition.created_at) }}</span><span><b>{{ transition.from_status }} → {{ transition.to_status }}</b> · {{ transition.trigger }}<span v-if="transition.reason" class="ml-1 text-muted-foreground">{{ transition.reason }}</span></span></div></div><p v-else class="text-sm text-muted-foreground">暂无迁移记录</p></div>
          <div v-if="canEdit && ['vesting', 'pending', 'ordered'].includes(detail.item.status)" class="space-y-3 border-t border-border/70 pt-4">
            <label class="block space-y-1 text-xs font-bold"><span>操作原因（必填）</span><Textarea v-model="actionReason" rows="2" placeholder="填写可审计的业务原因" /></label>
            <div class="flex flex-wrap justify-end gap-2">
              <Button v-if="detail.item.status === 'vesting'" variant="outline" class="rounded-full" :disabled="actionLoading" @click="requestAction('settle')"><Unlock class="size-3.5" /> 提前结算</Button>
              <Button variant="destructive" class="rounded-full" :disabled="actionLoading" @click="requestAction('revoke')"><Ban class="size-3.5" /> 风控作废</Button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
    <AdminConfirmDialog
      v-model:open="actionConfirm.open"
      :title="actionConfirm.action === 'revoke' ? '确认风控作废？' : '确认提前结算？'"
      :description="actionConfirm.action === 'revoke' ? '该操作会终止推荐记录并作废尚未发放的奖励，且会写入审计日志。' : '该操作会立即发放推荐人积分并结束冷静期，请确认原因已填写。'"
      :confirm-label="actionConfirm.action === 'revoke' ? '确认作废' : '确认结算'"
      :destructive="actionConfirm.action === 'revoke'"
      @confirm="executeAction"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { Ban, ClipboardList, Download, Eye, LoaderCircle, RefreshCw, Save, ShieldAlert, ShoppingCart, SlidersHorizontal, Unlock, UsersRound, WalletCards } from '@lucide/vue'
import { toast } from 'vue-sonner'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminFilterPanel from '@/components/admin/AdminFilterPanel.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

interface ReferralItem {
  id: number
  referral_code: string
  referrer?: { name?: string; email_masked?: string }
  referee?: { name?: string; email_masked?: string }
  order?: { order_number?: string; amount_minor?: number; currency?: string }
  status: string
  vesting_until?: string
  days_remaining?: number
  reward_points?: number
  risk_flags?: Array<Record<string, any> & { dimensions?: string[] }>
}

interface ReferralConfig {
  version: number
  enabled: boolean
  min_order_amount_minor: number
  referrer_reward_points: number
  referee_benefit_value: number
  vesting_period_days: number
  undelivered_fallback_days: number
  attribution_ttl_days: number
  monthly_cap_per_referrer: number
  anti_fraud_mode: string
}

interface ReferralDetail {
  item: ReferralItem
  transitions: Array<{ id: number; from_status: string; to_status: string; trigger: string; reason?: string; created_at: string }>
  rewards: Array<Record<string, unknown>>
}

const filters = reactive({ keyword: '', status: '', from: '', to: '' })
const items = ref<ReferralItem[]>([])
const overview = reactive({ total_referrals: 0, converted_orders: 0, attributed_gmv_minor: 0, pending_vesting_points: 0, settled_points: 0, fraud_blocked_count: 0 })
const pagination = reactive({ page: 1, pageSize: 20, total: 0 })
const loading = ref(false)
const exportLoading = ref(false)
const error = ref('')
const activeTab = ref<'ledger' | 'config'>('ledger')
const configLoading = ref(false)
const configSaving = ref(false)
const configError = ref('')
const config = reactive<ReferralConfig>({
  version: 0, enabled: false, min_order_amount_minor: 20000, referrer_reward_points: 1000,
  referee_benefit_value: 50,
  vesting_period_days: 30, undelivered_fallback_days: 45, attribution_ttl_days: 30, monthly_cap_per_referrer: 10, anti_fraud_mode: 'monitor'
})
const detailOpen = ref(false)
const detailLoading = ref(false)
const detail = ref<ReferralDetail | null>(null)
const actionReason = ref('')
const actionLoading = ref(false)
const actionConfirm = reactive<{ open: boolean; action: 'settle' | 'revoke' }>({ open: false, action: 'revoke' })
const authStore = useAuthStore()
const canEdit = computed(() => authStore.hasPermission('marketing:edit'))
const statusOptions = [
  { value: 'pending', label: '待首单' }, { value: 'ordered', label: '已支付' }, { value: 'vesting', label: '冷静期' },
  { value: 'settled', label: '已结算' }, { value: 'expired', label: '已过期' }, { value: 'revoked', label: '已作废' }, { value: 'reversed', label: '已冲正' },
]

const statItems = computed(() => [
  { key: 'total', label: '推荐记录', value: overview.total_referrals, icon: UsersRound, tone: 'blue' },
  { key: 'converted', label: '转化订单', value: overview.converted_orders, icon: ShoppingCart, tone: 'green' },
  { key: 'gmv', label: '裂变 GMV', value: formatMoney(overview.attributed_gmv_minor), icon: WalletCards, tone: 'blue' },
  { key: 'pending', label: '在途积分', value: overview.pending_vesting_points, icon: WalletCards, tone: 'amber' },
  { key: 'settled', label: '已结算积分', value: overview.settled_points, icon: WalletCards, tone: 'green' },
  { key: 'risk', label: '风控标记', value: overview.fraud_blocked_count, icon: ShieldAlert, tone: 'coral' },
])

const apiData = (response: any) => response?.data?.data ?? response?.data ?? {}
const fetchLedger = async () => {
  loading.value = true; error.value = ''
  try {
    const response = await axios.get('/api/admin/marketing/referrals', { params: { ...filters, page: pagination.page, page_size: pagination.pageSize } })
    const data = apiData(response)
    items.value = Array.isArray(data.items) ? data.items : []
    Object.assign(overview, data.overview || {})
    pagination.total = Number(response.data?.pagination?.total ?? data.total ?? 0)
  } catch (requestError) {
    error.value = requestError instanceof Error ? requestError.message : '推荐台账加载失败'
    toast.error(error.value)
  } finally { loading.value = false }
}
const exportLedger = async () => {
  exportLoading.value = true
  try {
    const response = await axios.get('/api/admin/marketing/referrals/export', { params: { ...filters }, responseType: 'blob' })
    const blob = new Blob([response.data], { type: 'text/csv;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url; link.download = 'referral-ledger.csv'; link.click(); URL.revokeObjectURL(url)
    toast.success('推荐台账已导出')
  } catch (requestError) {
    toast.error(requestError instanceof Error ? requestError.message : '推荐台账导出失败')
  } finally { exportLoading.value = false }
}
const fetchConfig = async () => {
  configLoading.value = true; configError.value = ''
  try {
    const response = await axios.get('/api/admin/marketing/referral-config')
    const data = apiData(response)
    if (!data.config) throw new Error('推荐规则接口未返回有效配置')
    Object.assign(config, data.config)
  } catch (requestError) {
    configError.value = requestError instanceof Error ? requestError.message : '推荐规则加载失败'
    toast.error(configError.value)
  } finally { configLoading.value = false }
}
const saveConfig = async () => {
  if (!canEdit.value || !config.version) return
  configSaving.value = true; configError.value = ''
  try {
    const response = await axios.put('/api/admin/marketing/referral-config', {
      enabled: config.enabled,
      min_order_amount_minor: config.min_order_amount_minor,
      referrer_reward_points: config.referrer_reward_points,
      referee_benefit_type: 'points',
      referee_benefit_value: config.referee_benefit_value,
      vesting_period_days: config.vesting_period_days,
      undelivered_fallback_days: config.undelivered_fallback_days,
      attribution_ttl_days: config.attribution_ttl_days,
      monthly_cap_per_referrer: config.monthly_cap_per_referrer,
      anti_fraud_mode: config.anti_fraud_mode,
      expected_version: config.version,
    })
    const data = apiData(response)
    if (!data.config) throw new Error('推荐规则接口未返回新版本')
    Object.assign(config, data.config)
    toast.success(`推荐规则已发布 v${config.version}`)
  } catch (requestError: any) {
    const status = requestError?.response?.status
    configError.value = status === 409 ? '规则版本已被其他管理员更新，请刷新后重试' : (requestError instanceof Error ? requestError.message : '推荐规则保存失败')
    toast.error(configError.value)
  } finally { configSaving.value = false }
}
const openDetail = async (id: number) => {
  detailOpen.value = true; detailLoading.value = true; detail.value = null; actionReason.value = ''
  try {
    const response = await axios.get(`/api/admin/marketing/referrals/${id}`)
    const data = apiData(response)
    detail.value = data as ReferralDetail
  } catch (requestError) {
    toast.error(requestError instanceof Error ? requestError.message : '推荐详情加载失败')
    detailOpen.value = false
  } finally { detailLoading.value = false }
}
const requestAction = (action: 'settle' | 'revoke') => {
  if (!actionReason.value.trim()) { toast.error('请填写操作原因'); return }
  actionConfirm.action = action
  actionConfirm.open = true
}
const executeAction = async () => {
  const action = actionConfirm.action
  if (!detail.value || !actionReason.value.trim()) { toast.error('请填写操作原因'); return }
  actionConfirm.open = false
  actionLoading.value = true
  try {
    const response = await axios.post(`/api/admin/marketing/referrals/${detail.value.item.id}/${action}`, { reason: actionReason.value.trim() })
    const data = apiData(response)
    toast.success(action === 'settle' ? '推荐奖励已结算' : '推荐记录已作废')
    actionReason.value = ''
    await fetchLedger()
    if (data.record) detail.value.item.status = data.record.status
    await openDetail(detail.value.item.id)
  } catch (requestError) {
    toast.error(requestError instanceof Error ? requestError.message : '推荐操作失败')
  } finally { actionLoading.value = false }
}
const applyFilters = () => { pagination.page = 1; fetchLedger() }
const resetFilters = () => { filters.keyword = ''; filters.status = ''; filters.from = ''; filters.to = ''; applyFilters() }
const updatePage = (page: number) => { pagination.page = page; fetchLedger() }
const statusLabel = (status: string) => statusOptions.find(option => option.value === status)?.label || status
const statusClass = (status: string) => ({ pending: 'bg-amber-100 text-amber-800', ordered: 'bg-blue-100 text-blue-800', vesting: 'bg-amber-100 text-amber-800', settled: 'bg-emerald-100 text-emerald-800', expired: 'bg-slate-100 text-slate-700', revoked: 'bg-rose-100 text-rose-800', reversed: 'bg-rose-100 text-rose-800' }[status] || 'bg-muted text-muted-foreground')
const formatMoney = (minor = 0, currency = 'USD') => new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(Number(minor) / 100)
const formatDate = (value?: string) => value ? new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value)) : '—'
watch(activeTab, (tab) => { if (tab === 'config' && !config.version) fetchConfig() })
onMounted(fetchLedger)
</script>
