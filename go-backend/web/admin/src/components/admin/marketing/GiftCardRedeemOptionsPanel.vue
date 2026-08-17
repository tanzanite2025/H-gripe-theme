<template>
  <section class="relative rounded-2xl border bg-card/75 p-4 shadow-sm">
    <div v-if="loading" class="absolute inset-0 z-10 flex items-center justify-center rounded-2xl bg-background/75">
      <LoaderCircle class="size-5 animate-spin text-primary" aria-label="正在加载礼品卡兑换设置" />
    </div>

    <div class="flex flex-wrap items-start justify-between gap-3">
      <div>
        <p class="text-xs font-black uppercase tracking-widest text-muted-foreground/60">Gift Card Redemption</p>
        <h3 class="mt-1 text-base font-black text-foreground">积分兑换礼品卡</h3>
        <p class="mt-1 text-xs leading-relaxed text-muted-foreground">
          这里维护会员用积分可兑换的礼品卡面额、币种和可兑换张数。
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
          {{ saving ? '保存中' : '保存兑换设置' }}
        </Button>
      </div>
    </div>

    <div class="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <AdminFormField label="启用积分兑换">
        <Switch v-model="redeemSettings.tz_redeem_enabled" :disabled="!canEdit" aria-label="启用积分兑换" />
      </AdminFormField>
      <AdminFormField label="默认币种" description="新增面额默认使用该币种；每个面额仍可单独设置币种。">
        <Select v-model="redeemSettings.tz_redeem_currency" :disabled="!canEdit || redeemCurrenciesLoading || redeemCurrencyOptions.length === 0">
          <SelectTrigger class="w-full font-mono uppercase">
            <SelectValue :placeholder="redeemCurrenciesLoading ? '正在读取币种目录' : '选择币种'" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem v-for="currency in redeemCurrencyOptions" :key="currency" :value="currency">
              {{ currency }}
            </SelectItem>
          </SelectContent>
        </Select>
      </AdminFormField>
      <AdminFormField label="兑换比例" description="例如 100 表示 100 积分兑换 1 个币种单位。">
        <Input
          v-model.number="redeemSettings.tz_redeem_exchange_rate"
          type="number"
          min="1"
          :disabled="!canEdit"
        />
      </AdminFormField>
      <AdminFormField label="最低兑换积分">
        <Input
          v-model.number="redeemSettings.tz_redeem_min_points"
          type="number"
          min="0"
          :disabled="!canEdit"
        />
      </AdminFormField>
      <AdminFormField label="每日最高兑换金额">
        <Input
          v-model.number="redeemSettings.tz_redeem_max_value_per_day"
          type="number"
          min="0"
          step="0.01"
          :disabled="!canEdit"
        />
      </AdminFormField>
      <AdminFormField label="礼品卡有效期（天）">
        <Input
          v-model.number="redeemSettings.tz_redeem_card_expiry_days"
          type="number"
          min="0"
          :disabled="!canEdit"
        />
      </AdminFormField>
    </div>

    <div class="mt-5 space-y-3">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <h4 class="text-xs font-black text-foreground">兑换面额与库存</h4>
        <Button
          v-if="canEdit"
          type="button"
          variant="outline"
          size="sm"
          :disabled="redeemCurrenciesLoading || redeemCurrencyOptions.length === 0"
          @click="addRedeemOption"
        >
          <Plus class="size-3.5" />
          添加面额
        </Button>
      </div>

      <div v-if="redeemSettings.options?.length" class="space-y-2">
        <div class="hidden grid-cols-[minmax(0,1fr)_140px_150px_40px] gap-2 px-2 text-[11px] font-bold uppercase text-muted-foreground md:grid">
          <span>面额</span>
          <span>币种</span>
          <span>张数</span>
          <span />
        </div>
        <div
          v-for="(option, index) in redeemSettings.options"
          :key="option.key || index"
          class="grid gap-2 rounded-lg border bg-background/60 p-2 md:grid-cols-[minmax(0,1fr)_140px_150px_40px] md:items-end"
        >
          <AdminFormField label="面额">
            <Input v-model.number="option.value" type="number" min="0.01" step="0.01" :disabled="!canEdit" />
          </AdminFormField>
          <AdminFormField label="币种">
            <Select v-model="option.currency" :disabled="!canEdit || redeemCurrenciesLoading || redeemCurrencyOptions.length === 0">
              <SelectTrigger class="font-mono uppercase">
                <SelectValue placeholder="币种" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem v-for="currency in redeemCurrencyOptions" :key="`${index}-${currency}`" :value="currency">
                  {{ currency }}
                </SelectItem>
              </SelectContent>
            </Select>
          </AdminFormField>
          <AdminFormField label="可兑换张数">
            <Input
              v-model.number="option.stock_quantity"
              type="number"
              min="0"
              step="1"
              :disabled="!canEdit"
            />
          </AdminFormField>
          <div class="flex items-end justify-end md:pb-0.5">
            <Button
              v-if="canEdit"
              type="button"
              variant="ghost"
              size="icon"
              title="删除兑换面额"
              :aria-label="`删除兑换面额 ${index + 1}`"
              @click="removeRedeemOption(index)"
            >
              <Trash2 class="size-4" />
            </Button>
          </div>
        </div>
      </div>
      <div v-else class="rounded-xl border border-dashed px-3 py-5 text-center text-xs text-muted-foreground">
        暂无兑换面额。添加面额并设置可兑换张数后，前台才会显示礼品卡。
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { LoaderCircle, Plus, RefreshCw, Save, Trash2 } from '@lucide/vue'
import AdminFormField from '@/components/admin/AdminFormField.vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'

interface GiftCardRedeemOption {
  key?: string
  value: number | string
  currency: string
  stock_quantity: number | string
}

interface GiftCardRedeemSettings {
  tz_redeem_enabled: boolean
  tz_redeem_currency: string
  tz_redeem_exchange_rate: number | string
  tz_redeem_min_points: number | string
  tz_redeem_max_value_per_day: number | string
  tz_redeem_card_expiry_days: number | string
  options?: GiftCardRedeemOption[]
}

const props = withDefaults(defineProps<{
  redeemSettings: GiftCardRedeemSettings
  loading?: boolean
  saving?: boolean
  redeemCurrencyOptions?: string[]
  redeemCurrenciesLoading?: boolean
  canEdit?: boolean
}>(), {
  loading: false,
  saving: false,
  redeemCurrencyOptions: () => [],
  redeemCurrenciesLoading: false,
  canEdit: false,
})

const emit = defineEmits<{
  (event: 'refresh'): void
  (event: 'save'): void
}>()

const addRedeemOption = (): void => {
  const currency = props.redeemSettings.tz_redeem_currency || props.redeemCurrencyOptions?.[0] || ''
  props.redeemSettings.options = [
    ...(props.redeemSettings.options || []),
    {
      key: `new-${Date.now()}-${Math.random().toString(36).slice(2)}`,
      value: 0,
      currency,
      stock_quantity: 0,
    },
  ]
}

const removeRedeemOption = (index: number): void => {
  props.redeemSettings.options = (props.redeemSettings.options || []).filter((_, optionIndex) => optionIndex !== index)
}
</script>
