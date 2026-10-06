export type WheelsetLacingTopologySelection =
  | 16
  | 20
  | 24
  | 28
  | 32
  | 36
  | '21_g3'
  | '18_2to1'
  | '24_2to1'

const SUPPORTED_WHEELSET_LACING_CROSS_COUNTS_BY_SYMMETRIC_HOLE_COUNT: Record<number, readonly number[]> = {
  16: [0, 1],
  20: [0, 1, 2],
  24: [0, 1, 2, 3],
  28: [0, 1, 2, 3],
  32: [0, 1, 2, 3, 4],
  36: [0, 1, 2, 3, 4],
}

/**
 * Returns supported cross counts for one explicit topology selection. Topology
 * generation, coordinates, and projection metrics are owned by the Go API;
 * this contract only controls which buttons the presentation can enable.
 */
export const getSupportedWheelsetLacingCrossCounts = (
  topologySelection: WheelsetLacingTopologySelection,
): readonly number[] => {
  if (topologySelection === '21_g3' || topologySelection === '18_2to1' || topologySelection === '24_2to1') return [2]
  return SUPPORTED_WHEELSET_LACING_CROSS_COUNTS_BY_SYMMETRIC_HOLE_COUNT[topologySelection] || []
}
