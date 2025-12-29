<template>
  <div class="h-full overflow-y-auto px-1 pt-1 pb-3 md:p-6">
    <div class="warranty-check md:max-w-[640px] md:mx-auto">
      <WarrantyCheckPanel
        :is-logged-in="isLoggedIn"
        @login-request="$emit('loginRequest')"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import WarrantyCheckPanel from '~/components/WarrantyCheckPanel.vue'

defineProps<{
  isLoggedIn: boolean
}>()

defineEmits<{
  'loginRequest': []
}>()
</script>
