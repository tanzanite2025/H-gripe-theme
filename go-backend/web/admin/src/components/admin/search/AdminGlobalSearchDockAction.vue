<template>
  <Dialog v-model:open="globalSearchOpen">
    <Tooltip>
      <TooltipTrigger as-child>
        <DialogTrigger as-child>
          <AdminBottomUtilityDockActionButton :aria-label="t('adminGlobalSearch.open')">
            <Search class="size-6" aria-hidden="true" />
          </AdminBottomUtilityDockActionButton>
        </DialogTrigger>
      </TooltipTrigger>
      <TooltipContent side="top" :side-offset="12">{{ t('adminGlobalSearch.title') }}</TooltipContent>
    </Tooltip>

    <AdminGlobalSearchDialog
      v-if="globalSearchOpen"
      :navigation-items="navigationItems"
      @request-close="globalSearchOpen = false"
    />
  </Dialog>
</template>

<script setup lang="ts">
import { defineAsyncComponent, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Search } from '@lucide/vue'
import AdminBottomUtilityDockActionButton from '@/components/admin/AdminBottomUtilityDockActionButton.vue'
import { Dialog, DialogTrigger } from '@/components/ui/dialog'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import type { AdminNavigationItem } from '@/lib/adminNavigation'
import { useAdminI18n } from '@/i18n'

defineProps<{ navigationItems: AdminNavigationItem[] }>()

const AdminGlobalSearchDialog = defineAsyncComponent(
  () => import('@/components/admin/search/AdminGlobalSearchDialog.vue'),
)
const { t } = useAdminI18n()
const route = useRoute()
const globalSearchOpen = ref(false)

watch(() => route.fullPath, () => { globalSearchOpen.value = false })
</script>
