export type WheelsetLacingHoleSelection =
  | 16
  | 20
  | 21
  | 24
  | 28
  | 32
  | 36
  | '18_2to1'
  | '24_2to1'

const SUPPORTED_WHEELSET_LACING_CROSS_COUNTS_BY_HOLE_SELECTION: Record<number, readonly number[]> = {
  16: [0, 1],
  20: [0, 1, 2],
  24: [0, 1, 2, 3],
  28: [0, 1, 2, 3],
  32: [0, 1, 2, 3, 4],
  36: [0, 1, 2, 3, 4],
}

/**
 * Returns the supported selector values for one topology choice. Topology
 * generation, coordinates, and projection metrics are owned by the Go API;
 * this contract only controls which buttons the presentation can enable.
 */
export const getSupportedWheelsetLacingCrossCounts = (
  holeSelection: WheelsetLacingHoleSelection,
): readonly number[] => {
  if (holeSelection === 21 || holeSelection === '18_2to1' || holeSelection === '24_2to1') return [2]
  return SUPPORTED_WHEELSET_LACING_CROSS_COUNTS_BY_HOLE_SELECTION[holeSelection] || []
}
