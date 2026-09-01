<template>
  <div class="relative space-y-5">
    <div v-if="loading" class="absolute inset-0 z-10 flex items-center justify-center rounded-2xl bg-background/75">
      <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载积分规则" />
    </div>

    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <p class="text-xs font-black uppercase tracking-widest text-muted-foreground/60">Loyalty Program</p>
        <h3 class="mt-1 text-base font-black text-foreground">积分规则配置</h3>
        <p class="mt-1 text-xs leading-relaxed text-muted-foreground">
          保存会生成新的积分规则版本，前台积分说明、签到和推荐共用这份配置。
        </p>
      </div>
      <div class="flex gap-2">
        <Button type="button" variant="outline" size="sm" :disabled="loading || saving" @click="emit('refresh')">
 <RefreshCw :class="['size-3.5', loading ? 'animate-spin': '']" />
          刷新
        </Button>
        <Button v-if="canEdit" type="button" size="sm" :disabled="loading || saving" @click="emit('save')">
          <LoaderCircle v-if="saving" class="size-3.5 animate-spin" />
          <Save v-else class="size-3.5" />
          {{ saving ? '保存中' : '保存规则' }}
        </Button>
      </div>
    </div>

    <div v-if="version" class="rounded-xl border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
      当前生效版本：<span class="font-mono font-black text-foreground">v{{ version }}</span>
    </div>

    <ProgramSection title="订单完成返积分" description="真实发放规则；不是前台展示文案。">
      <div class="grid gap-4 lg:grid-cols-[minmax(0,1fr)_minmax(260px,0.8fr)]">
        <AdminFormField
          :label="`每 1 ${baseCurrency} 商品成交金额奖励积分`"
          description="订单从已发货变为已完成时发放。按商品小计减优惠后的 USD 成交金额计算，不含运费和税费；0 表示关闭订单返积分。"
        >
          <Input
            v-model.number="loyaltySettings.tz_loyalty_purchase_earn_points_per_currency_unit"
            type="number"
            min="0"
            step="1"
            :disabled="!canEdit"
          />
        </AdminFormField>
        <div class="space-y-1.5 text-xs leading-relaxed text-muted-foreground">
          <p><span class="font-bold text-foreground">触发时间：</span>后台把订单状态改为“已完成”后自动入账。</p>
          <p><span class="font-bold text-foreground">计算口径：</span>积分 = USD 商品成交金额 × 当前比例，向下取整。</p>
          <p><span class="font-bold text-foreground">币种基准：</span>积分规则只以 USD 为基准；礼品卡币种只影响兑换面额。</p>
        </div>
      </div>
    </ProgramSection>

    <ProgramSection title="推荐与签到积分" description="用于推荐奖励、每日签到和连续签到奖励。">
      <div class="grid gap-4 md:grid-cols-2">
        <AdminFormField label="推荐人奖励积分" description="被推荐用户完成首次购买后，推荐人获得的积分。">
          <Input
            v-model.number="loyaltySettings.tz_loyalty_referral_referrer_points"
            type="number"
            min="0"
            :disabled="!canEdit"
          />
        </AdminFormField>
        <AdminFormField label="被推荐人奖励积分" description="被推荐用户完成首次购买后，被推荐人获得的积分。">
          <Input
            v-model.number="loyaltySettings.tz_loyalty_referral_referee_points"
            type="number"
            min="0"
            :disabled="!canEdit"
          />
        </AdminFormField>
        <AdminFormField label="每日签到基础积分" description="会员每天第一次签到获得的基础积分。">
          <Input
            v-model.number="loyaltySettings.tz_loyalty_checkin_base_points"
            type="number"
            min="0"
            :disabled="!canEdit"
          />
        </AdminFormField>
        <AdminFormField label="连续签到奖励周期（天）" description="连续签到达到这个天数周期时，增加一次额外积分。">
          <Input
            v-model.number="loyaltySettings.tz_loyalty_checkin_streak_interval_days"
            type="number"
            min="1"
            :disabled="!canEdit"
          />
        </AdminFormField>
        <AdminFormField label="连续签到额外积分" description="每完成一个连续签到周期额外增加的积分。">
          <Input
            v-model.number="loyaltySettings.tz_loyalty_checkin_streak_bonus_points"
            type="number"
            min="0"
            :disabled="!canEdit"
          />
        </AdminFormField>
        <AdminFormField label="单次签到最高积分" description="基础积分加连续签到奖励后，单次签到最多发这么多。">
          <Input
            v-model.number="loyaltySettings.tz_loyalty_checkin_max_points"
            type="number"
            min="0"
            :disabled="!canEdit"
          />
        </AdminFormField>
      </div>
    </ProgramSection>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h } from 'vue'
import { LoaderCircle, RefreshCw, Save } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import type { LoyaltySettings } from '@/modules/marketing/marketingTypes'

const props = withDefaults(defineProps<{
  loyaltySettings: LoyaltySettings
  pointsBaseCurrency?: string
  version?: number
  loading?: boolean
  saving?: boolean
  canEdit?: boolean
}>(), {
  pointsBaseCurrency: 'USD',
  version: 0,
  loading: false,
  saving: false,
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'refresh'): void
  (event: 'save'): void
}>()

const baseCurrency = computed(() => String(props.pointsBaseCurrency || 'USD').trim().toUpperCase() || 'USD')

const ProgramSection = defineComponent({
  props: {
    title: { type: String, required: true },
    description: { type: String, default: '' },
  },
  setup(sectionProps, { slots }) {
 return () => h('section', { class: 'grid w-full max-w-none gap-5 rounded-2xl border bg-card/75 p-4 shadow-sm lg:grid-cols-[190px_minmax(0,1fr)]'}, [
      h('div', {}, [
 h('h4', { class: 'text-sm font-black tracking-tighter uppercase text-foreground'}, sectionProps.title),
 sectionProps.description ? h('p', { class: 'mt-1 text-[10px] font-bold leading-relaxed text-muted-foreground'}, sectionProps.description) : null,
      ]),
 h('div', { class: 'min-w-0'}, slots.default?.()),
    ])
  },
})
</script>

