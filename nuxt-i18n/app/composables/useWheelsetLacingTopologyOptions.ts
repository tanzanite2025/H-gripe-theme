import { computed } from 'vue'
import { useState } from '#imports'
import { useApiRequest } from '~/composables/useApiRequest'
import type {
  SpokeCalculatorSelectOption,
  SpokeCalculatorTopologyDistribution,
  SpokeCalculatorTopologySummary,
} from '~/types/spokeCalculator'

interface WheelsetLacingTopologyOptionsState {
  topologies: SpokeCalculatorTopologySummary[]
  loading: boolean
  error: string | null
  source: 'empty' | 'api' | 'error'
}

interface UnknownRecord {
  [key: string]: unknown
}

const inFlightWheelsetLacingTopologyLoads: Record<string, Promise<SpokeCalculatorTopologySummary[]> | undefined> = {}

const isRecord = (value: unknown): value is UnknownRecord => (
  typeof value === 'object' && value !== null && !Array.isArray(value)
)

const requireFiniteInteger = (value: unknown, field: string): number => {
  if (typeof value !== 'number' || !Number.isInteger(value) || value < 0) {
    throw new Error(`Invalid wheelset lacing topology ${field}`)
  }
  return value
}

const requireString = (value: unknown, field: string): string => {
  if (typeof value !== 'string' || value.trim() === '') {
    throw new Error(`Invalid wheelset lacing topology ${field}`)
  }
  return value
}

const requireDistribution = (value: unknown): SpokeCalculatorTopologyDistribution => {
  if (value === 'symmetric_1to1' || value === 'uniform_2to1' || value === 'g3_2to1') {
    return value
  }
  throw new Error('Invalid wheelset lacing topology distribution')
}

const arrayLength = (value: unknown, field: string): number => {
  if (!Array.isArray(value)) {
    throw new Error(`Invalid wheelset lacing topology ${field}`)
  }
  return value.length
}

const normalizeWheelsetLacingTopologySummaries = (payload: unknown): SpokeCalculatorTopologySummary[] => {
  if (!isRecord(payload) || !isRecord(payload.data) || !Array.isArray(payload.data.topologies)) {
    throw new Error('Wheelset lacing topology response is malformed')
  }

  return payload.data.topologies.map((value, index) => {
    if (!isRecord(value)) {
      throw new Error(`Wheelset lacing topology ${index} is malformed`)
    }
    const topologyId = requireString(value.topology_id, `${index}.topology_id`)
    const holeCount = requireFiniteInteger(value.hole_count, `${topologyId}.hole_count`)
    const cross = requireFiniteInteger(value.cross, `${topologyId}.cross`)
    const distribution = requireDistribution(value.distribution)
    const sideACount = arrayLength(value.hub_holes_a, `${topologyId}.hub_holes_a`)
    const sideBCount = arrayLength(value.hub_holes_b, `${topologyId}.hub_holes_b`)
    const expectedSideCount = distribution === 'symmetric_1to1'
      ? holeCount / 2
      : distribution === 'g3_2to1'
        ? Math.round(holeCount * 2 / 3)
        : Math.round(holeCount * 2 / 3)
    if (!Number.isInteger(expectedSideCount) || sideACount !== expectedSideCount || sideBCount !== holeCount - expectedSideCount) {
      throw new Error(`Wheelset lacing topology ${topologyId} has inconsistent side counts`)
    }
    return {
      topologyId,
      selection: requireString(value.selection, `${topologyId}.selection`),
      holeCount,
      cross,
      distribution,
      displayLayout: requireString(value.display_layout, `${topologyId}.display_layout`),
      sideACount,
      sideBCount,
    }
  })
}

const topologyRatioLabel = (distribution: SpokeCalculatorTopologyDistribution) => {
  if (distribution === 'g3_2to1') return 'G3 · 2:1'
  if (distribution === 'uniform_2to1') return '2:1'
  return '1:1'
}

const createWheelsetLacingTopologySelectOptions = (
  summaries: SpokeCalculatorTopologySummary[],
): SpokeCalculatorSelectOption[] => summaries.map(summary => ({
  value: summary.topologyId,
  label: `${summary.holeCount}H · ${topologyRatioLabel(summary.distribution)} · ${summary.cross}X`,
}))

export const useWheelsetLacingTopologyOptions = () => {
  const { baseURL, request } = useApiRequest()
  const state = useState<WheelsetLacingTopologyOptionsState>(
    `wheelset-lacing-topology-options:${baseURL}`,
    () => ({ topologies: [], loading: false, error: null, source: 'empty' }),
  )

  const loadTopologies = async (): Promise<SpokeCalculatorTopologySummary[]> => {
    if (state.value.loading && inFlightWheelsetLacingTopologyLoads[baseURL]) {
      return inFlightWheelsetLacingTopologyLoads[baseURL]!
    }

    state.value.loading = true
    state.value.error = null
    const loadRequest = request<unknown>(
      '/wheelset-lacing/topologies',
      {},
      'Failed to load wheelset lacing topologies',
    )
      .then(normalizeWheelsetLacingTopologySummaries)
      .then((topologies) => {
        state.value.topologies = topologies
        state.value.source = 'api'
        return topologies
      })
      .catch((error: unknown) => {
        state.value.topologies = []
        state.value.source = 'error'
        state.value.error = error instanceof Error ? error.message : 'Failed to load wheelset lacing topologies'
        return []
      })
      .finally(() => {
        state.value.loading = false
        inFlightWheelsetLacingTopologyLoads[baseURL] = undefined
      })

    inFlightWheelsetLacingTopologyLoads[baseURL] = loadRequest
    return loadRequest
  }

  // A server render can serialize `loading: true` before its fire-and-forget
  // request settles. The source marker is the durable state, so retry on the
  // client whenever no topology payload has been received yet.
  if (state.value.source === 'empty' && !inFlightWheelsetLacingTopologyLoads[baseURL]) {
    void loadTopologies()
  }

  return {
    topologies: computed(() => state.value.topologies),
    options: computed(() => createWheelsetLacingTopologySelectOptions(state.value.topologies)),
    loading: computed(() => state.value.loading),
    error: computed(() => state.value.error),
    source: computed(() => state.value.source),
    loadTopologies,
  }
}
