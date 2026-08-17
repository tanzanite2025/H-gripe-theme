<template>
 <div class="space-y-4">
    <AdminPageHeader
      title="URL 管理 / 同步与检查"
      description="更新路由快照并执行前台可用性检查"
    >
      <template #actions>
        <Button variant="outline" :disabled="loading || syncing || checking" @click="refreshAll">
 <RefreshCw :class="['size-4', loading ? 'animate-spin': '']" />
          刷新
        </Button>
        <Button variant="outline" :disabled="syncing || !canEdit" @click="syncCatalog">
 <RefreshCw :class="['size-4', syncing ? 'animate-spin': '']" />
          同步 URL
        </Button>
        <Button :disabled="checking || !canEdit || pagination.total === 0" @click="checkCatalog">
 <CircleCheck :class="['size-4', checking ? 'animate-spin': '']" />
          检查路由
        </Button>
      </template>
    </AdminPageHeader>

    <AdminStatsGrid :items="statItems" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { CircleCheck, RefreshCw, TriangleAlert } from '@lucide/vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import AdminStatsGrid from '@/components/admin/AdminStatsGrid.vue'
import { Button } from '@/components/ui/button'
import { useStorefrontRouteCatalog } from '@/composables/url-management/useStorefrontRouteCatalog'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const canEdit = authStore.hasPermission('url:edit')
const {
  stats,
  loading,
  syncing,
  checking,
  pagination,
  refreshAll,
  syncCatalog,
  checkCatalog,
} = useStorefrontRouteCatalog(canEdit)

const statItems = computed(() => [
  { key: 'checked', label: '已检查', value: stats.value.checked, icon: CircleCheck, tone: 'green' },
  { key: 'unchecked', label: '未检查', value: stats.value.unchecked, icon: RefreshCw, tone: stats.value.unchecked ? 'amber' : 'gray' },
  { key: 'attention', label: '待处理', value: stats.value.needs_attention, icon: TriangleAlert, tone: stats.value.needs_attention ? 'coral' : 'gray' },
  { key: 'stale', label: '失效快照', value: stats.value.stale, icon: TriangleAlert, tone: stats.value.stale ? 'amber' : 'gray' },
])

onMounted(() => {
  void refreshAll()
})
</script>
