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

<script setup>
import CouponEditorDialog from '@/components/admin/marketing/CouponEditorDialog.vue'
import MemberLevelEditorDialog from '@/components/admin/marketing/MemberLevelEditorDialog.vue'

defineProps({
  couponOpen: { type: Boolean, default: false },
  couponMode: { type: String, default: 'create' },
  couponForm: { type: Object, required: true },
  couponErrors: { type: Object, required: true },
  couponSubmitting: { type: Boolean, default: false },
  levelOpen: { type: Boolean, default: false },
  levelMode: { type: String, default: 'create' },
  levelForm: { type: Object, required: true },
  levelErrors: { type: Object, required: true },
  levelSubmitting: { type: Boolean, default: false },
})

const emit = defineEmits([
  'update:couponOpen',
  'update:levelOpen',
  'submit-coupon',
  'submit-level',
  'clear-coupon-error',
  'clear-level-error',
])
</script>
