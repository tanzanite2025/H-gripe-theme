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
  <GiftCardEditorDialog
    :open="giftCardOpen"
    :form="giftCardForm"
    :errors="giftCardErrors"
    :submitting="giftCardSubmitting"
    @update:open="emit('update:giftCardOpen', $event)"
    @submit="emit('submit-gift-card')"
    @clear-error="emit('clear-gift-card-error', $event)"
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

<script setup>
import CouponEditorDialog from '@/components/admin/marketing/CouponEditorDialog.vue'
import GiftCardEditorDialog from '@/components/admin/marketing/GiftCardEditorDialog.vue'
import MemberLevelEditorDialog from '@/components/admin/marketing/MemberLevelEditorDialog.vue'

defineProps({
  couponOpen: { type: Boolean, default: false },
  couponMode: { type: String, default: 'create' },
  couponForm: { type: Object, required: true },
  couponErrors: { type: Object, required: true },
  couponSubmitting: { type: Boolean, default: false },
  giftCardOpen: { type: Boolean, default: false },
  giftCardForm: { type: Object, required: true },
  giftCardErrors: { type: Object, required: true },
  giftCardSubmitting: { type: Boolean, default: false },
  levelOpen: { type: Boolean, default: false },
  levelMode: { type: String, default: 'create' },
  levelForm: { type: Object, required: true },
  levelErrors: { type: Object, required: true },
  levelSubmitting: { type: Boolean, default: false },
})

const emit = defineEmits([
  'update:couponOpen',
  'update:giftCardOpen',
  'update:levelOpen',
  'submit-coupon',
  'submit-gift-card',
  'submit-level',
  'clear-coupon-error',
  'clear-gift-card-error',
  'clear-level-error',
])
</script>
