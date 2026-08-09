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
</template>

<script setup lang="ts">
import { defineComponent, h } from 'vue'
import { Fingerprint, UserRound } from '@lucide/vue'
import AdminStatusBadge, { type AdminStatusTone } from '@/components/admin/AdminStatusBadge.vue'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import type { VisitorProfile } from './visitorTypes'

withDefaults(defineProps<{
  selectedProfile?: VisitorProfile | null
  formatDate: (value: unknown) => string
}>(), {
  selectedProfile: null,
})

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
</script>
