<template>
  <Tabs :model-value="activeTab" class="gap-4" @update:model-value="emit('update:activeTab', $event)">
    <TabsList variant="line" class="h-10 w-full justify-start overflow-x-auto rounded-none border-b bg-transparent p-0">
      <TabsTrigger value="coupons" class="h-9 flex-none px-4">
        <BadgePercent class="size-4" />
        优惠券
      </TabsTrigger>
      <TabsTrigger value="giftcards" class="h-9 flex-none px-4">
        <Gift class="size-4" />
        礼品卡
      </TabsTrigger>
      <TabsTrigger value="loyalty" class="h-9 flex-none px-4">
        <Coins class="size-4" />
        积分
      </TabsTrigger>
      <TabsTrigger value="levels" class="h-9 flex-none px-4">
        <Crown class="size-4" />
        会员等级
      </TabsTrigger>
    </TabsList>

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
      <GiftCardTablePanel
        :loading="giftCardsLoading"
        :gift-cards="giftCards"
        :filters="giftCardFilters"
        :pagination="giftCardPagination"
        :can-create="canCreate"
        :format-currency="formatCurrency"
        :format-date="formatDate"
        :gift-card-status-name="giftCardStatusName"
        :gift-card-status-tone="giftCardStatusTone"
        @filter-change="emit('gift-card-filter-change')"
        @create="emit('create-gift-card')"
        @view="emit('view-gift-card', $event)"
        @update-page="emit('update-gift-card-page', $event)"
        @update-page-size="emit('update-gift-card-page-size', $event)"
      />
    </TabsContent>

    <TabsContent value="loyalty" class="space-y-3">
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

    <TabsContent value="levels" class="space-y-3">
      <div class="flex justify-end">
        <Button v-if="canCreate" size="sm" @click="emit('create-level')">
          <Plus class="size-3.5" />
          创建等级
        </Button>
      </div>

      <MemberLevelTablePanel
        :loading="levelsLoading"
        :levels="levels"
        :can-edit="canEdit"
        :can-delete="canDelete"
        :format-rate="formatRate"
        @edit="emit('edit-level', $event)"
        @delete="emit('delete-level', $event)"
      />
    </TabsContent>
  </Tabs>
</template>

<script setup>
import { BadgePercent, Coins, Crown, Gift, Plus } from '@lucide/vue'
import CouponTablePanel from '@/components/admin/marketing/CouponTablePanel.vue'
import GiftCardTablePanel from '@/components/admin/marketing/GiftCardTablePanel.vue'
import LoyaltyPanel from '@/components/admin/marketing/LoyaltyPanel.vue'
import MemberLevelTablePanel from '@/components/admin/marketing/MemberLevelTablePanel.vue'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

defineProps({
  activeTab: { type: String, default: 'coupons' },
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
  levelsLoading: { type: Boolean, default: false },
  levels: { type: Array, default: () => [] },
  formatRate: { type: Function, required: true }
})

const emit = defineEmits([
  'update:activeTab',
  'coupon-filter-change',
  'create-coupon',
  'edit-coupon',
  'delete-coupon',
  'update-coupon-page',
  'update-coupon-page-size',
  'gift-card-filter-change',
  'create-gift-card',
  'view-gift-card',
  'update-gift-card-page',
  'update-gift-card-page-size',
  'loyalty-filter-change',
  'update-loyalty-page',
  'update-loyalty-page-size',
  'submit-loyalty-adjustment',
  'clear-loyalty-error',
  'create-level',
  'edit-level',
  'delete-level'
])
</script>
