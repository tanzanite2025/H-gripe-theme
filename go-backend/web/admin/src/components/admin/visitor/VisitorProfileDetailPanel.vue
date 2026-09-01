<template>
  <Card class="h-full min-h-0 overflow-hidden py-0">
    <CardHeader class="shrink-0 border-b bg-muted/30 px-4 py-3">
      <CardTitle class="flex items-center gap-2">
        <Fingerprint class="size-4 text-primary" />
        画像详情
      </CardTitle>
      <CardDescription>只展示已采集事实，不猜测邮箱、地区或身份</CardDescription>
    </CardHeader>
    <CardContent class="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
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
            <DetailItem label="Status">
              <AdminStatusBadge :tone="statusTone(selectedProfile.profile_status)">
                {{ statusLabel(selectedProfile.profile_status) }}
              </AdminStatusBadge>
            </DetailItem>
            <DetailItem label="Quality">Q{{ selectedProfile.profile_quality_score || 0 }}</DetailItem>
 <DetailItem label="Email" class="col-span-2">{{ selectedProfile.email || '未采集'}}</DetailItem>
 <DetailItem label="Email Source" class="col-span-2">{{ selectedProfile.email_source || 'not_captured'}}</DetailItem>
          </dl>
        </section>

        <section class="rounded-2xl border p-3">
          <h3 class="mb-3 text-xs font-black uppercase tracking-wider">绑定事实</h3>
          <div class="space-y-2 text-xs">
            <FactRow label="Public Chat visitor" :active="selectedProfile.has_customer_service_visitor">
              {{ selectedProfile.customer_service_visitor_hash_preview || '未绑定' }}
            </FactRow>
            <FactRow label="Cart session" :active="selectedProfile.has_cart_session">
 <span class="break-all font-mono">{{ selectedProfile.cart_session_id || '未绑定'}}</span>
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
 <DetailItem label="Region" class="col-span-2">{{ selectedProfile.region || '-'}}</DetailItem>
 <DetailItem label="City" class="col-span-2">{{ selectedProfile.city || '-'}}</DetailItem>
          </dl>
          <div class="mt-3 flex flex-wrap gap-1.5">
            <AdminStatusBadge :tone="selectedProfile.has_ip_fingerprint ? 'green' : 'gray'">IP fingerprint</AdminStatusBadge>
            <AdminStatusBadge :tone="selectedProfile.has_user_agent_fingerprint ? 'green' : 'gray'">User-Agent fingerprint</AdminStatusBadge>
          </div>
        </section>

        <section class="rounded-2xl border border-rose-500/20 p-3">
          <div class="mb-3 flex items-center justify-between gap-2">
            <h3 class="text-xs font-black uppercase tracking-wider">访问控制</h3>
            <AdminStatusBadge :tone="selectedProfile.ip_block_match ? 'coral' : selectedProfile.ip_address ? 'green' : 'gray'">
              {{ selectedProfile.ip_block_match ? 'IP 实际已拦截' : selectedProfile.ip_address ? 'IP 当前未拦截' : '当前 IP 未保留' }}
            </AdminStatusBadge>
          </div>
          <dl class="grid grid-cols-2 gap-2 text-xs">
            <DetailItem label="IP（已脱敏）" class="col-span-2">
              <span class="break-all font-mono">{{ selectedProfile.ip_address || '未保留可操作 IP' }}</span>
            </DetailItem>
          </dl>
          <p class="mt-3 text-xs leading-5 text-amber-800/80 dark:text-amber-100/80">
            画像中的 IP 仅展示脱敏值。这里的封禁是全局 IP 封禁，会影响使用同一出口 IP 的其他用户。
          </p>

          <div v-if="profileBlockRules.length" class="mt-3 rounded-xl border border-rose-500/20 bg-rose-500/10 p-3 text-xs">
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <p class="font-black text-rose-700 dark:text-rose-200">
                  画像归属规则 · {{ profileBlockRules.length }} 条有效规则
                </p>
                <p class="mt-1 text-[11px] text-muted-foreground">
                  规则按最近一次变更排序；IP 漂移后旧规则仍会保留，解除会批量解除此画像的有效规则。
                </p>
              </div>
              <ShieldAlert class="size-4 shrink-0 text-rose-500" />
            </div>
            <div class="mt-2 space-y-2">
              <div
                v-for="rule in profileBlockRules"
                :key="rule.id"
                class="rounded-lg border border-rose-500/15 bg-background/50 p-2"
              >
                <div class="flex items-start justify-between gap-2">
                  <span class="break-all font-mono text-[11px] font-bold text-rose-800 dark:text-rose-100">
                    #{{ rule.id }} · {{ rule.cidr || 'unknown-cidr' }}
                  </span>
                  <span class="shrink-0 text-[10px] text-muted-foreground">
                    {{ rule.expires_at ? `到期：${formatDate(rule.expires_at)}` : '持续有效' }}
                  </span>
                </div>
                <p class="mt-1 break-words text-[11px] text-muted-foreground">{{ rule.reason || '未填写原因' }}</p>
              </div>
            </div>
          </div>

          <div v-if="selectedProfile.ip_block_match" class="mt-3 rounded-xl border border-amber-500/25 bg-amber-500/10 p-3 text-xs">
            <div class="flex items-start justify-between gap-3">
              <div class="min-w-0">
                <p class="font-black text-amber-800 dark:text-amber-200">当前 IP 实际命中规则 #{{ selectedProfile.ip_block_match.id }}</p>
                <p class="mt-1 break-all font-mono text-[11px] text-amber-900/80 dark:text-amber-100/80">
                  {{ selectedProfile.ip_block_match.cidr || 'unknown-cidr' }}
                </p>
                <p class="mt-1 text-[11px] text-muted-foreground">
                  来源：{{ sourceLabel(selectedProfile.ip_block_match.source) }}
                  <span v-if="selectedProfile.ip_block_match.expires_at"> · 到期：{{ formatDate(selectedProfile.ip_block_match.expires_at) }}</span>
                </p>
              </div>
              <ShieldCheck class="size-4 shrink-0 text-amber-600" />
            </div>
            <p v-if="profileOwnsIPBlockMatch && !profileRuleIsLatest" class="mt-2 leading-5 text-amber-800/80 dark:text-amber-100/80">
              当前 IP 命中的是该画像的历史归属规则；解除画像封禁会批量解除此画像的有效规则。
            </p>
            <p v-else-if="!profileOwnsIPBlockMatch" class="mt-2 leading-5 text-amber-800/80 dark:text-amber-100/80">
              当前 IP 命中的是其他来源的全局规则，解除画像规则不会解除这条拦截。
            </p>
          </div>

          <p v-else-if="!selectedProfile.ip_address" class="mt-3 text-xs leading-5 text-muted-foreground">
            该历史画像没有保留原始 IP，或原始 IP 已按保留策略清除，无法从这里创建封禁规则。
          </p>
          <p v-else class="mt-3 text-xs leading-5 text-muted-foreground">
            当前 IP 未命中任何有效的全局封禁规则。
          </p>

          <div v-if="selectedProfile.ip_address || profileBlockRules.length" class="mt-3 flex flex-wrap items-center gap-2">
            <Button
              v-if="blockIp && selectedProfile.ip_address && !profileOwnsIPBlockMatch"
              type="button"
              variant="destructive"
              size="sm"
              :disabled="blockSubmitting || unblockSubmitting || !blockIp"
              @click="openBlockDialog"
            >
              <ShieldAlert class="size-3.5" />
              {{ profileBlockRules.length ? '新增画像封禁规则' : '创建画像封禁规则' }}
            </Button>
            <Button
              v-if="unblockIp && profileBlockRules.length"
              type="button"
              variant="outline"
              size="sm"
              :disabled="blockSubmitting || unblockSubmitting || !unblockIp"
              @click="unblockSelectedIP"
            >
              <ShieldCheck class="size-3.5" />
              {{ unblockSubmitting ? '解除画像规则中' : '解除画像封禁' }}
            </Button>
          </div>
        </section>

        <section class="rounded-2xl border p-3">
          <h3 class="mb-3 text-xs font-black uppercase tracking-wider">时间</h3>
          <dl class="grid gap-2 text-xs">
            <DetailItem label="Last Seen">{{ formatDate(selectedProfile.last_seen_at) }}</DetailItem>
            <DetailItem label="Last Meaningful">{{ formatDate(selectedProfile.last_meaningful_seen_at) }}</DetailItem>
            <DetailItem label="First Meaningful">{{ formatDate(selectedProfile.first_meaningful_seen_at) }}</DetailItem>
            <DetailItem label="Last Action">{{ actionLabel(selectedProfile.last_meaningful_action) }}</DetailItem>
            <DetailItem label="Retention Until">{{ formatDate(selectedProfile.retention_until) }}</DetailItem>
            <DetailItem label="Created">{{ formatDate(selectedProfile.created_at) }}</DetailItem>
            <DetailItem label="Updated">{{ formatDate(selectedProfile.updated_at) }}</DetailItem>
          </dl>
        </section>
      </template>
    </CardContent>
  </Card>

  <Dialog v-model:open="blockDialogOpen">
    <DialogContent size="md" class="max-h-[90dvh] overflow-y-auto" @open-auto-focus.prevent>
      <form class="space-y-5" @submit.prevent="submitBlock">
        <DialogHeader>
          <DialogTitle>封禁访客 IP</DialogTitle>
          <DialogDescription>
            这会创建全局 IP 规则，命中后 API 请求直接返回 HTTP 403；同一出口 IP 的其他用户也可能受影响。
          </DialogDescription>
        </DialogHeader>

        <div v-if="selectedProfile" class="rounded-xl border border-border/70 bg-muted/30 p-3 text-xs">
          <p class="text-[10px] font-black uppercase tracking-widest text-muted-foreground/70">TARGET</p>
          <p class="mt-1 break-all font-mono font-bold text-foreground">{{ selectedProfile.ip_address }}</p>
        </div>

        <div class="space-y-4">
          <AdminFormField label="封禁原因" required description="原因会保留在规则和审计中，但审计日志只记录长度与是否填写。">
            <Textarea v-model="blockForm.reason" class="min-h-24" maxlength="500" placeholder="例如：同一 IP 连续探测库存接口。" />
          </AdminFormField>

          <AdminFormField
            label="到期时间"
            description="留空表示持续有效；填写后会自动在该时间解除命中。"
          >
            <Input v-model="blockForm.expiresAt" type="datetime-local" />
          </AdminFormField>
        </div>

        <p v-if="blockError" class="rounded-xl border border-red-500/30 bg-red-500/10 px-3 py-2 text-xs font-bold text-red-300">
          {{ blockError }}
        </p>

        <DialogFooter>
          <Button type="button" variant="outline" @click="blockDialogOpen = false">取消</Button>
          <Button type="submit" variant="destructive" :disabled="blockSubmitting">
            <LoaderCircle v-if="blockSubmitting" class="size-4 animate-spin" />
            {{ blockSubmitting ? '封禁中' : '确认封禁' }}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, reactive, ref } from 'vue'
import { Fingerprint, LoaderCircle, ShieldAlert, ShieldCheck, UserRound } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import type { VisitorIPBlockPayload, VisitorProfile } from '@/modules/visitor/visitorTypes'

const props = withDefaults(defineProps<{
  selectedProfile?: VisitorProfile | null
  formatDate: (value: unknown) => string
  blockIp?: (profile: VisitorProfile, payload: VisitorIPBlockPayload) => Promise<void>
  unblockIp?: (profile: VisitorProfile) => Promise<void>
}>(), {
  selectedProfile: null,
  blockIp: undefined,
  unblockIp: undefined,
})

const blockDialogOpen = ref(false)
const blockSubmitting = ref(false)
const unblockSubmitting = ref(false)
const blockError = ref('')
const blockForm = reactive({
  reason: '',
  expiresAt: '',
})
const profileBlockRules = computed(() => {
  const rules = props.selectedProfile?.ip_block_rules
  if (rules?.length) return rules
  const legacyRule = props.selectedProfile?.ip_block
  return legacyRule ? [legacyRule] : []
})
const profileOwnsIPBlockMatch = computed(() => {
  const profile = props.selectedProfile
  const match = profile?.ip_block_match
  return Boolean(
    profile &&
    match &&
    match.source === 'visitor_profile' &&
    String(match.source_reference || '') === String(profile.id),
  )
})
const profileRuleIsLatest = computed(() => {
  const latestRule = profileBlockRules.value[0]
  const matchedRule = props.selectedProfile?.ip_block_match
  return Boolean(
    latestRule &&
    matchedRule &&
    Number(latestRule.id) === Number(matchedRule.id),
  )
})

const openBlockDialog = (): void => {
  blockError.value = ''
  blockForm.reason = ''
  blockForm.expiresAt = ''
  blockDialogOpen.value = true
}

interface ErrorLike {
  response?: { data?: { error?: string; message?: string } }
}

const submitBlock = async (): Promise<void> => {
  if (!props.selectedProfile || !props.blockIp) return
  const reason = blockForm.reason.trim()
  if (!reason) {
    blockError.value = '请填写封禁原因。'
    return
  }
  if (!window.confirm('确认创建这条全局 IP 封禁规则吗？使用同一出口 IP 的其他用户也可能被拦截。')) {
    return
  }

  blockSubmitting.value = true
  blockError.value = ''
  try {
    await props.blockIp(props.selectedProfile, {
      reason,
      expires_at: blockForm.expiresAt ? new Date(blockForm.expiresAt).toISOString() : null,
    })
    blockDialogOpen.value = false
  } catch (error) {
    const typedError = error as ErrorLike
    blockError.value = typedError.response?.data?.error || typedError.response?.data?.message || '封禁失败，请稍后重试。'
  } finally {
    blockSubmitting.value = false
  }
}

const unblockSelectedIP = async (): Promise<void> => {
  if (!props.selectedProfile || !props.unblockIp) return
  if (!window.confirm(`确定解除访客 #${props.selectedProfile.id} 的 IP 封禁吗？`)) return

  unblockSubmitting.value = true
  blockError.value = ''
  try {
    await props.unblockIp(props.selectedProfile)
  } catch (error) {
    const typedError = error as ErrorLike
    blockError.value = typedError.response?.data?.error || typedError.response?.data?.message || '解除失败，请稍后重试。'
  } finally {
    unblockSubmitting.value = false
  }
}

const statusLabel = (status?: string): string => ({
  active: '有效',
  candidate: '候选',
  archived: '归档',
  suppressed: '抑制',
} as Record<string, string>)[status || ''] || '有效'

const statusTone = (status?: string): AdminStatusTone => ({
  active: 'green',
  candidate: 'amber',
  archived: 'gray',
  suppressed: 'coral',
} as Record<string, AdminStatusTone>)[status || ''] || 'green'

const actionLabel = (action?: string): string => ({
  cart_action: '购物车动作',
  customer_service: '客服会话',
  email_capture: '邮箱捕获',
  account: '账号绑定',
  identity_bind: '身份绑定',
} as Record<string, string>)[action || ''] || '无有效动作'

const sourceLabel = (source?: string): string => ({
  manual: '手动',
  visitor_profile: '访客画像',
  commercial_crawler: '商业爬虫',
  risk_automation: '风险自动化',
} as Record<string, string>)[source || ''] || source || '未知'

const DetailItem = defineComponent({
  props: {
    label: { type: String, required: true }
  },
  setup(props, { slots, attrs }) {
    return () => h('div', { class: ['space-y-1 rounded-xl border p-2', attrs.class] }, [
 h('dt', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70'}, props.label),
 h('dd', { class: 'break-words font-bold text-foreground'}, slots.default ? slots.default() : '-')
    ])
  }
})

const FactRow = defineComponent({
  props: {
    label: { type: String, required: true },
    active: { type: Boolean, default: false }
  },
  setup(props, { slots }) {
 return () => h('div', { class: 'rounded-xl border p-2'}, [
 h('div', { class: 'mb-1 flex items-center justify-between gap-2'}, [
 h('span', { class: 'text-[10px] font-black uppercase tracking-widest text-muted-foreground/70'}, props.label),
        h(AdminStatusBadge, { tone: props.active ? 'green' : 'gray' }, { default: () => props.active ? 'captured' : 'missing' })
      ]),
 h('p', { class: 'break-words font-mono text-[11px] text-foreground'}, slots.default ? slots.default() : '-')
    ])
  }
})
</script>

