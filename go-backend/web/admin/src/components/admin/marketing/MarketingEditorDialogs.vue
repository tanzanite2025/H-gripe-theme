<template>
  <CouponEditorDialog
    :open="couponOpen"
    :mode="couponMode"
    :form="couponForm"
    :errors="couponErrors"
    :submitting="couponSubmitting"
    @update:open="emit('update:couponOpen', $event)"
    @submit="emit('submit-coupon')"
    @clear-error="emit('clear-coupon-error', $event)"
  />
  <MemberLevelEditorDialog
    :open="levelOpen"
    :mode="levelMode"
    :form="levelForm"
    :errors="levelErrors"
    :submitting="levelSubmitting"
    @update:open="emit('update:levelOpen', $event)"
    @submit="emit('submit-level')"
    @clear-error="emit('clear-level-error', $event)"
  />
</template>

<script setup lang="ts">
import CouponEditorDialog from '@/components/admin/marketing/CouponEditorDialog.vue'
import type { CouponErrors, CouponForm } from '@/components/admin/marketing/CouponEditorDialog.vue'
import MemberLevelEditorDialog from '@/components/admin/marketing/MemberLevelEditorDialog.vue'
import type { MemberLevelErrors, MemberLevelForm } from '@/components/admin/marketing/MemberLevelEditorDialog.vue'

const props = withDefaults(defineProps<{
  couponOpen?: boolean
  couponMode?: string
  couponForm: CouponForm
  couponErrors: CouponErrors
  couponSubmitting?: boolean
  levelOpen?: boolean
  levelMode?: string
  levelForm: MemberLevelForm
  levelErrors: MemberLevelErrors
  levelSubmitting?: boolean
}>(), {
  couponOpen: false,
  couponMode: 'create',
  couponSubmitting: false,
  levelOpen: false,
  levelMode: 'create',
  levelSubmitting: false,
})

const emit = defineEmits<{
  (event: 'update:couponOpen', value: boolean): void
  (event: 'update:levelOpen', value: boolean): void
  (event: 'submit-coupon'): void
  (event: 'submit-level'): void
  (event: 'clear-coupon-error', field: unknown): void
  (event: 'clear-level-error', field: unknown): void
}>()
</script>
