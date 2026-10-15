<template>
  <DashboardOutboxFailureActionCenterPanel
    :state="panelState"
    @refresh="refreshDashboardBusinessFollowUps"
    @open-follow-up-list="router.push('/dashboard/business-follow-ups')"
  />
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import DashboardOutboxFailureActionCenterPanel from './DashboardOutboxFailureActionCenterPanel.vue'
import { useDashboardBusinessFollowUpDataLoadingAndTargetSelection } from '@/composables/dashboard/useDashboardBusinessFollowUpDataLoadingAndTargetSelection'

const router = useRouter()
const { panelState, refreshDashboardBusinessFollowUps } = useDashboardBusinessFollowUpDataLoadingAndTargetSelection({ mode: 'home-preview' })
</script>
