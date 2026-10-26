<template>
  <AdminGlobalSearchDialogContent
    v-model:query="query"
    :result-groups="resultGroups"
    :loading="loading"
    :failed-source-labels="failedSourceLabels"
    :navigation-failed="navigationFailed"
    @activate-link="closeAdminGlobalSearchAfterUnmodifiedLinkActivation"
    @open-first-result="openFirstAdminGlobalSearchResultFromKeyboard"
    @retry-search="retryFailedAdminGlobalSearchSources"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { isNavigationFailure, NavigationFailureType, useRouter, type RouteLocationRaw } from 'vue-router'
import AdminGlobalSearchDialogContent from './AdminGlobalSearchDialogContent.vue'
import { useAdminGlobalSearchController } from '@/composables/search/useAdminGlobalSearchController'
import { createAdminNavigationPageSearchSource } from '@/modules/search/adminNavigationPageSearchSource'
import type { AdminNavigationItem } from '@/lib/adminNavigation'
import { useAdminI18n } from '@/i18n'

const props = defineProps<{ navigationItems: AdminNavigationItem[] }>()
const emit = defineEmits<{ (event: 'request-close'): void }>()
const { t } = useAdminI18n()
const router = useRouter()
// Compose sources here; presentation and query control stay independent of page navigation.
const sources = computed(() => [createAdminNavigationPageSearchSource(props.navigationItems, t)])

async function navigateToAdminGlobalSearchResultRoute(target: RouteLocationRaw): Promise<boolean> {
  const failure = await router.push(target)
  return !failure || isNavigationFailure(failure, NavigationFailureType.duplicated)
}

const {
  query,
  resultGroups,
  loading,
  failedSourceLabels,
  navigationFailed,
  retryFailedAdminGlobalSearchSources,
  closeAdminGlobalSearchAfterUnmodifiedLinkActivation,
  openFirstAdminGlobalSearchResultFromKeyboard,
} = useAdminGlobalSearchController({
  sources,
  navigateToResult: navigateToAdminGlobalSearchResultRoute,
  closeSearchDialog: () => emit('request-close'),
})
</script>
