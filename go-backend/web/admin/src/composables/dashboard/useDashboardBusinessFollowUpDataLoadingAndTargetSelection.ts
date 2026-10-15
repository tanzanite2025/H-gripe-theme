import { computed, onMounted, onUnmounted, ref, shallowRef, watch, type Ref } from 'vue'
import { isAxiosError } from 'axios'
import {
  fetchDashboardCommandCenterOutboxFailures,
  fetchDashboardBusinessFollowUpList,
  fetchDashboardBusinessFollowUpDetail,
  type DashboardBusinessFollowUpItem,
  type DashboardCommandCenterOutboxFailures,
} from '@/api/dashboardCommandCenterOutboxFailures'
import type { DashboardBusinessFollowUpPanelState } from '@/modules/dashboard/dashboardBusinessFollowUpPresentationTypes'

interface DashboardBusinessFollowUpDataLoadingOptions {
  mode: 'home-preview' | 'business-list'
  selectedEventId?: Ref<number | null>
  selectedEventQueryInvalid?: Ref<boolean>
}

export function useDashboardBusinessFollowUpDataLoadingAndTargetSelection(options: DashboardBusinessFollowUpDataLoadingOptions) {
  const loading = ref(true)
  const loadError = ref(false)
  const outboxFailures = shallowRef<DashboardCommandCenterOutboxFailures | null>(null)
  const lastSuccessfulTimestamp = ref<string | null>(null)
  const selectedEventNotice = ref('')
  const selectedEventOutsidePreview = shallowRef<DashboardBusinessFollowUpItem | null>(null)
  const selectedProcessingState = ref('all')
  const selectedEventId = computed(() => options.selectedEventId?.value ?? null)
  let dataLoadingUnmounted = false
  let requestSequence = 0

  const visibleBusinessFollowUps = computed(() => {
    const items = [...(outboxFailures.value?.items || [])]
    if (selectedEventOutsidePreview.value) items.unshift(selectedEventOutsidePreview.value)
    const targetIndex = items.findIndex((item) => item.id === selectedEventId.value)
    if (targetIndex > 0) items.unshift(...items.splice(targetIndex, 1))
    return items.filter((item) => selectedProcessingState.value === 'all' || item.processing_state === selectedProcessingState.value)
  })
  const panelState = computed<DashboardBusinessFollowUpPanelState>(() => ({
    loading: loading.value,
    loadError: loadError.value,
    outboxFailures: outboxFailures.value,
    lastSuccessfulTimestamp: lastSuccessfulTimestamp.value,
    selectedEventNotice: selectedEventNotice.value,
    selectedEventOutsidePreview: selectedEventOutsidePreview.value,
    selectedEventId: selectedEventId.value,
    selectedProcessingState: selectedProcessingState.value,
    visibleBusinessFollowUps: visibleBusinessFollowUps.value,
  }))

  async function refreshDashboardBusinessFollowUps(): Promise<void> {
    if (dataLoadingUnmounted) return
    const currentRequest = ++requestSequence
    const requestedEventId = selectedEventId.value
    loading.value = true
    loadError.value = false
    selectedEventNotice.value = ''
    selectedEventOutsidePreview.value = null
    try {
      const response = await (options.mode === 'business-list' ? fetchDashboardBusinessFollowUpList() : fetchDashboardCommandCenterOutboxFailures())
      if (dataLoadingUnmounted || currentRequest !== requestSequence) return
      outboxFailures.value = response.action_center.outbox_failures
      lastSuccessfulTimestamp.value = outboxFailures.value.status === 'ready' ? response.timestamp : null
      if (options.selectedEventQueryInvalid?.value) selectedEventNotice.value = '事项编号格式无效，请从首页重新打开。'
      if (outboxFailures.value.status === 'ready' && requestedEventId && !outboxFailures.value.items.some((item) => item.id === requestedEventId)) {
        try {
          const item = await fetchDashboardBusinessFollowUpDetail(requestedEventId)
          if (dataLoadingUnmounted || currentRequest !== requestSequence) return
          selectedEventOutsidePreview.value = item
        } catch (error) {
          if (dataLoadingUnmounted || currentRequest !== requestSequence) return
          selectedEventNotice.value = isAxiosError(error) && error.response?.status === 404
            ? '这项提醒已离开出错清单，或当前账号无法查看。请到相关业务记录核实实际结果；提醒消失不代表业务已完成。'
            : '目标事项详情暂时无法读取，请刷新重试。下方其他事项仍可查看。'
        }
      }
    } catch {
      if (dataLoadingUnmounted || currentRequest !== requestSequence) return
      outboxFailures.value = null
      lastSuccessfulTimestamp.value = null
      loadError.value = true
    } finally {
      if (!dataLoadingUnmounted && currentRequest === requestSequence) loading.value = false
    }
  }

  onMounted(() => {
    void refreshDashboardBusinessFollowUps()
    window.addEventListener('focus', refreshDashboardBusinessFollowUps)
  })
  watch([selectedEventId, () => options.selectedEventQueryInvalid?.value], () => {
    if (options.mode !== 'business-list') return
    selectedProcessingState.value = 'all'
    void refreshDashboardBusinessFollowUps()
  })
  onUnmounted(() => {
    dataLoadingUnmounted = true
    requestSequence += 1
    window.removeEventListener('focus', refreshDashboardBusinessFollowUps)
  })

  return { panelState, selectedProcessingState, refreshDashboardBusinessFollowUps }
}
