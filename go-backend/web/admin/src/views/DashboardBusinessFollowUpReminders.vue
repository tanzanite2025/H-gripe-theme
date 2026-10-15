<template>
  <div class="space-y-4">
    <Button variant="outline" size="sm" @click="router.push('/')">返回首页</Button>
    <DashboardOutboxFailureActionCenterPanel
      list-mode
      :state="panelState"
      @refresh="refreshDashboardBusinessFollowUps"
      @update-processing-state="selectedProcessingState = $event"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import DashboardOutboxFailureActionCenterPanel from '@/components/admin/dashboard/DashboardOutboxFailureActionCenterPanel.vue'
import { useDashboardBusinessFollowUpDataLoadingAndTargetSelection } from '@/composables/dashboard/useDashboardBusinessFollowUpDataLoadingAndTargetSelection'
import { parsePositiveUint32RouteQueryParameter } from '@/lib/parsePositiveUint32RouteQueryParameter'
const router = useRouter()
const route = useRoute()
const selectedEventId = computed(() => parsePositiveUint32RouteQueryParameter(route.query.event_id))
const selectedEventQueryInvalid = computed(() => Boolean(route.query.event_id) && selectedEventId.value === null)
const { panelState, selectedProcessingState, refreshDashboardBusinessFollowUps } = useDashboardBusinessFollowUpDataLoadingAndTargetSelection({ mode: 'business-list', selectedEventId, selectedEventQueryInvalid })
let reminderPageUnmounted = false
let targetScrollSequence = 0
watch([() => panelState.value.loading, selectedEventId], async ([loading, eventId]) => {
  const currentScroll = ++targetScrollSequence
  if (loading || !eventId) return
  await nextTick()
  if (reminderPageUnmounted || currentScroll !== targetScrollSequence) return
  document.getElementById(`business-follow-up-${eventId}`)?.scrollIntoView({ block: 'start', behavior: 'smooth' })
})
onBeforeUnmount(() => {
  reminderPageUnmounted = true
  targetScrollSequence += 1
})
</script>
