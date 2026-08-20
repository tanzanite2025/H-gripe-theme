<template>
  <div class="flex h-full min-h-0 flex-col gap-4 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="收款方式"
      description="管理前台可用的收款方式、费用规则和展示顺序。"
    />

    <div class="min-h-0 flex-1 overflow-auto">
      <PaymentMethodsSettingsPanel :can-edit="canEdit" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import PaymentMethodsSettingsPanel from '@/components/admin/settings/PaymentMethodsSettingsPanel.vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const canEdit = computed(() => authStore.hasPermission('settings:edit'))
</script>
