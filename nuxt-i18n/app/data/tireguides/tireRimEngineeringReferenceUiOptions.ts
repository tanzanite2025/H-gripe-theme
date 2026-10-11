/**
 * UI options accepted by the engineering workbench.
 *
 * These values only control the buttons and presets shown in the browser.
 * Verdicts, ranges, pressure, tube guidance, and projections are returned
 * by the Go tire/rim engineering endpoint.
 */
export type TireRimEngineeringRimSystem = 'hookless' | 'hooked'

export const TIRE_RIM_ENGINEERING_MIN_RIM_INNER_WIDTH_MM = 13
export const TIRE_RIM_ENGINEERING_MAX_RIM_INNER_WIDTH_MM = 100

/** Every integer inner width accepted by the Go engineering model. */
export const TIRE_RIM_ENGINEERING_SUPPORTED_RIM_INNER_WIDTHS_MM = Array.from(
  {
    length: TIRE_RIM_ENGINEERING_MAX_RIM_INNER_WIDTH_MM - TIRE_RIM_ENGINEERING_MIN_RIM_INNER_WIDTH_MM + 1,
  },
  (_, index) => TIRE_RIM_ENGINEERING_MIN_RIM_INNER_WIDTH_MM + index,
)

export const TIRE_RIM_ENGINEERING_TIRE_WIDTH_PRESETS_MM = [18, 20, 23, 25, 28, 30, 32, 35, 38, 40, 45, 50, 55, 60, 70, 80, 90, 100, 110, 115, 120, 127, 132] as const
