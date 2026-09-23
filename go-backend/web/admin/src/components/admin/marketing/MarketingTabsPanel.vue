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
        :format-currency="formatCurrency"
        :format-date="formatDate"
        @filter-change="emit('coupon-filter-change')"
        @create="emit('create-coupon')"
        @edit="emit('edit-coupon', $event)"
        @delete="emit('delete-coupon', $event)"
        @update-page="emit('update-coupon-page', $event)"
        @update-page-size="emit('update-coupon-page-size', $event)"
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

    <TabsContent value="risk" class="space-y-3">
      <PromotionRiskAnalysisPanel
        :loading="promotionRiskLoading"
        :analysis="promotionRiskAnalysis"
        :error="promotionRiskError"
        :format-currency="formatCurrency"
        :format-date="formatDate"
        :format-rate="formatRate"
        @refresh="emit('refresh-promotion-risk-analysis')"
      />
    </TabsContent>
  </Tabs>
</template>

<script setup lang="ts">
import CouponTablePanel from '@/components/admin/marketing/CouponTablePanel.vue'
import LoyaltyPanel from '@/components/admin/marketing/LoyaltyPanel.vue'
import LoyaltyProgramSettingsPanel from '@/components/admin/marketing/LoyaltyProgramSettingsPanel.vue'
import MemberLevelTablePanel from '@/components/admin/marketing/MemberLevelTablePanel.vue'
import PromotionRiskAnalysisPanel from '@/components/admin/marketing/PromotionRiskAnalysisPanel.vue'
import { Tabs, TabsContent } from '@/components/ui/tabs'
import type {
  CouponFilters,
  CouponRecord,
  LoyaltyAdjustmentForm,
  LoyaltyErrors,
  LoyaltyFilters,
  LoyaltySettings,
  LoyaltyTransaction,
  MarketingPagination,
  MarketingStatusDisplay,
  MemberLevel,
  PromotionRiskAnalysis,
} from '@/modules/marketing/marketingTypes'

withDefaults(defineProps<{
  activeTab?: string
  activeSubTab?: string
  canCreate?: boolean
  canEdit?: boolean
  canDelete?: boolean
  couponsLoading?: boolean
  coupons?: CouponRecord[]
  couponFilters: CouponFilters
  couponPagination: MarketingPagination
  couponValue: (coupon: CouponRecord) => string
  couponStatus: (coupon: CouponRecord) => MarketingStatusDisplay
  formatMoney: (value: unknown) => string
  formatCurrency: (value: unknown, currency?: string) => string
  formatDate: (value: unknown) => string
  loyaltyLoading?: boolean
  loyaltyTransactions?: LoyaltyTransaction[]
  loyaltyFilters: LoyaltyFilters
  loyaltyPagination: MarketingPagination
  loyaltyForm: LoyaltyAdjustmentForm
  loyaltyErrors: LoyaltyErrors
  loyaltySubmitting?: boolean
  loyaltyTypeName: (type?: string | null) => string
  loyaltySettings: LoyaltySettings
  pointsBaseCurrency?: string
  loyaltyProgramVersion?: number
  loyaltyProgramLoading?: boolean
  loyaltyProgramSaving?: boolean
  levelsLoading?: boolean
  levels?: MemberLevel[]
  levelsUsingFallback?: boolean
  promotionRiskLoading?: boolean
  promotionRiskAnalysis?: PromotionRiskAnalysis | null
  promotionRiskError?: string | null
  formatRate: (value: unknown) => string
}>(), {
  activeTab: 'coupons',
  activeSubTab: 'transactions',
  canCreate: false,
  canEdit: false,
  canDelete: false,
  couponsLoading: false,
  coupons: () => [],
  loyaltyLoading: false,
  loyaltyTransactions: () => [],
  loyaltySubmitting: false,
  pointsBaseCurrency: 'USD',
  loyaltyProgramVersion: 0,
  loyaltyProgramLoading: false,
  loyaltyProgramSaving: false,
  levelsLoading: false,
  levels: () => [],
  levelsUsingFallback: false,
  promotionRiskLoading: false,
  promotionRiskAnalysis: null,
  promotionRiskError: null,
})

const emit = defineEmits<{
  (event: 'coupon-filter-change'): void
  (event: 'create-coupon'): void
  (event: 'edit-coupon', coupon: CouponRecord): void
  (event: 'delete-coupon', coupon: CouponRecord): void
  (event: 'update-coupon-page', page: number): void
  (event: 'update-coupon-page-size', pageSize: number): void
  (event: 'loyalty-filter-change'): void
  (event: 'update-loyalty-page', page: number): void
  (event: 'update-loyalty-page-size', pageSize: number): void
  (event: 'submit-loyalty-adjustment'): void
  (event: 'clear-loyalty-error', field: keyof LoyaltyAdjustmentForm): void
  (event: 'refresh-loyalty-program-config'): void
  (event: 'save-loyalty-program-config'): void
  (event: 'create-level'): void
  (event: 'edit-level', level: MemberLevel): void
  (event: 'delete-level', level: MemberLevel): void
  (event: 'refresh-promotion-risk-analysis'): void
}>()
</script>

