<template>
  <Tabs :model-value="activeTab" class="gap-4">
    <TabsContent value="coupons" class="space-y-3">
      <CouponTablePanel
        :loading="couponsLoading"
        :coupons="coupons"
        :filters="couponFilters"
        :pagination="couponPagination"
        :can-create="canCreate"
        :can-edit="canEdit"
        :can-delete="canDelete"
        :coupon-value="couponValue"
        :coupon-status="couponStatus"
        :format-money="formatMoney"
        :format-date="formatDate"
        @filter-change="emit('coupon-filter-change')"
        @create="emit('create-coupon')"
        @edit="emit('edit-coupon', $event)"
        @delete="emit('delete-coupon', $event)"
        @update-page="emit('update-coupon-page', $event)"
        @update-page-size="emit('update-coupon-page-size', $event)"
      />
    </TabsContent>

    <TabsContent value="giftcards" class="space-y-3">
      <GiftCardRedeemOptionsPanel
        :redeem-settings="redeemSettings"
        :loading="loyaltyProgramLoading"
        :saving="loyaltyProgramSaving"
        :payment-currency-options="paymentCurrencyOptions"
        :payment-currencies-loading="paymentCurrenciesLoading"
        :can-edit="canEdit"
        @refresh="emit('refresh-loyalty-program-config')"
        @save="emit('save-loyalty-program-config')"
      />

      <GiftCardTablePanel
        :loading="giftCardsLoading"
        :gift-cards="giftCards"
        :filters="giftCardFilters"
        :pagination="giftCardPagination"
        :format-currency="formatCurrency"
        :format-date="formatDate"
        :gift-card-status-name="giftCardStatusName"
        :gift-card-status-tone="giftCardStatusTone"
        @filter-change="emit('gift-card-filter-change')"
        @view="emit('view-gift-card', $event)"
        @update-page="emit('update-gift-card-page', $event)"
        @update-page-size="emit('update-gift-card-page-size', $event)"
      />
    </TabsContent>

    <TabsContent value="loyalty" class="space-y-3">
      <Tabs :model-value="activeSubTab" class="gap-3">
        <TabsContent value="transactions" class="space-y-3">
          <LoyaltyPanel
            :loading="loyaltyLoading"
            :transactions="loyaltyTransactions"
            :filters="loyaltyFilters"
            :pagination="loyaltyPagination"
            :form="loyaltyForm"
            :errors="loyaltyErrors"
            :submitting="loyaltySubmitting"
            :can-adjust="canCreate"
            :loyalty-type-name="loyaltyTypeName"
            :format-date="formatDate"
            @apply-filter="emit('loyalty-filter-change')"
            @update-page="emit('update-loyalty-page', $event)"
            @update-page-size="emit('update-loyalty-page-size', $event)"
            @submit="emit('submit-loyalty-adjustment')"
            @clear-error="emit('clear-loyalty-error', $event)"
          />
        </TabsContent>

        <TabsContent value="rules" class="space-y-3">
          <LoyaltyProgramSettingsPanel
            :loyalty-settings="loyaltySettings"
            :version="loyaltyProgramVersion"
            :loading="loyaltyProgramLoading"
            :saving="loyaltyProgramSaving"
            :can-edit="canEdit"
            :points-base-currency="pointsBaseCurrency"
            @refresh="emit('refresh-loyalty-program-config')"
            @save="emit('save-loyalty-program-config')"
          />
        </TabsContent>
      </Tabs>
    </TabsContent>

    <TabsContent value="levels" class="space-y-3">
      <div class="rounded-2xl border bg-card/75 px-4 py-3 text-xs text-muted-foreground shadow-sm">
        <p>系统内置普通、铜牌、银牌、金牌、铂金、钻石六个会员等级。这里维护积分区间、折扣率和权益说明，不再新建或删除等级。</p>
        <p v-if="levelsUsingFallback" class="mt-1 text-[11px] font-bold text-amber-600 dark:text-amber-300">
          接口暂未返回真实会员等级，当前显示默认预览；请确认后端迁移已执行并重启服务后再编辑。
        </p>
      </div>

      <MemberLevelTablePanel
        :loading="levelsLoading"
        :levels="levels"
        :can-edit="canEdit"
        :can-delete="false"
        :format-rate="formatRate"
        @edit="emit('edit-level', $event)"
      />
    </TabsContent>
  </Tabs>
</template>

<script setup>
import CouponTablePanel from '@/components/admin/marketing/CouponTablePanel.vue'
import GiftCardRedeemOptionsPanel from '@/components/admin/marketing/GiftCardRedeemOptionsPanel.vue'
import GiftCardTablePanel from '@/components/admin/marketing/GiftCardTablePanel.vue'
import LoyaltyPanel from '@/components/admin/marketing/LoyaltyPanel.vue'
import LoyaltyProgramSettingsPanel from '@/components/admin/marketing/LoyaltyProgramSettingsPanel.vue'
import MemberLevelTablePanel from '@/components/admin/marketing/MemberLevelTablePanel.vue'
import { Tabs, TabsContent } from '@/components/ui/tabs'

defineProps({
  activeTab: { type: String, default: 'coupons' },
  activeSubTab: { type: String, default: 'transactions' },
  canCreate: { type: Boolean, default: false },
  canEdit: { type: Boolean, default: false },
  canDelete: { type: Boolean, default: false },
  couponsLoading: { type: Boolean, default: false },
  coupons: { type: Array, default: () => [] },
  couponFilters: { type: Object, default: () => ({}) },
  couponPagination: { type: Object, default: () => ({}) },
  couponValue: { type: Function, required: true },
  couponStatus: { type: Function, required: true },
  formatMoney: { type: Function, required: true },
  formatDate: { type: Function, required: true },
  giftCardsLoading: { type: Boolean, default: false },
  giftCards: { type: Array, default: () => [] },
  giftCardFilters: { type: Object, default: () => ({}) },
  giftCardPagination: { type: Object, default: () => ({}) },
  formatCurrency: { type: Function, required: true },
  giftCardStatusName: { type: Function, required: true },
  giftCardStatusTone: { type: Function, required: true },
  loyaltyLoading: { type: Boolean, default: false },
  loyaltyTransactions: { type: Array, default: () => [] },
  loyaltyFilters: { type: Object, default: () => ({}) },
  loyaltyPagination: { type: Object, default: () => ({}) },
  loyaltyForm: { type: Object, default: () => ({}) },
  loyaltyErrors: { type: Object, default: () => ({}) },
  loyaltySubmitting: { type: Boolean, default: false },
  loyaltyTypeName: { type: Function, required: true },
  loyaltySettings: { type: Object, required: true },
  redeemSettings: { type: Object, required: true },
  pointsBaseCurrency: { type: String, default: 'USD' },
  loyaltyProgramVersion: { type: Number, default: 0 },
  loyaltyProgramLoading: { type: Boolean, default: false },
  loyaltyProgramSaving: { type: Boolean, default: false },
  paymentCurrencyOptions: { type: Array, default: () => [] },
  paymentCurrenciesLoading: { type: Boolean, default: false },
  levelsLoading: { type: Boolean, default: false },
  levels: { type: Array, default: () => [] },
  levelsUsingFallback: { type: Boolean, default: false },
  formatRate: { type: Function, required: true }
})

const emit = defineEmits([
  'coupon-filter-change',
  'create-coupon',
  'edit-coupon',
  'delete-coupon',
  'update-coupon-page',
  'update-coupon-page-size',
  'gift-card-filter-change',
  'view-gift-card',
  'update-gift-card-page',
  'update-gift-card-page-size',
  'loyalty-filter-change',
  'update-loyalty-page',
  'update-loyalty-page-size',
  'submit-loyalty-adjustment',
  'clear-loyalty-error',
  'refresh-loyalty-program-config',
  'save-loyalty-program-config',
  'create-level',
  'edit-level',
  'delete-level'
])
</script>
