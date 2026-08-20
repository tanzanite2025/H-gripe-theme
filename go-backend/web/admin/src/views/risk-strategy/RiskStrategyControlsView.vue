<template>
  <div class="flex h-full min-h-0 flex-col gap-3 overflow-hidden">
    <AdminPageHeader
      class="shrink-0"
      title="人工保护"
      description="临时强制 3DS 或暂停新支付，所有保护规则都带有过期时间和审计记录。"
    />
    <RiskStrategyControlPanel />
  </div>
</template>

<script setup lang="ts">
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import RiskStrategyControlPanel from '@/components/admin/payment/RiskStrategyControlPanel.vue'
</script>

