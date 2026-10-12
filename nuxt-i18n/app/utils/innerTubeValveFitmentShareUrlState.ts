import type { LocationQuery } from 'vue-router'
import type { InnerTubeValveFitmentMatrixMetadata, InnerTubeValveFitmentMode } from '~/types/innerTubeValveFitmentCalculationAndMatrixTypes'

export interface InnerTubeValveFitmentShareState {
  rimDepth: number
  fitmentMode: InnerTubeValveFitmentMode
  manualValveLength: number
  manualExtenderLength: number
  pumpHeadGripDepth: number
  rimDepthUncertainty: number
}

type InnerTubeValveFitmentShareInputMetadata = Pick<InnerTubeValveFitmentMatrixMetadata,
  | 'rim_depth_min_mm' | 'rim_depth_max_mm'
  | 'base_valve_length_options_mm' | 'extender_length_options_mm'
  | 'pump_head_grip_depth_min_mm' | 'pump_head_grip_depth_max_mm' | 'default_pump_head_grip_depth_mm'
  | 'rim_depth_uncertainty_max_mm' | 'default_rim_depth_uncertainty_mm'
>

const innerTubeValveFitmentShareQueryParameterNames = [
  'rim_depth',
  'fitment_mode',
  'valve_length',
  'extender_length',
  'pump_head_grip_depth',
  'rim_depth_uncertainty',
] as const

export function hasInnerTubeValveFitmentShareQueryParameters(query: LocationQuery): boolean {
  return innerTubeValveFitmentShareQueryParameterNames.some(parameterName => (
    Object.prototype.hasOwnProperty.call(query, parameterName)
  ))
}

export function readInnerTubeValveFitmentShareQueryParameterSignature(query: LocationQuery): string {
  // Include presence so a bare parameter (?rim_depth) differs from no parameter.
  return JSON.stringify(innerTubeValveFitmentShareQueryParameterNames.map(parameterName => [
    Object.prototype.hasOwnProperty.call(query, parameterName), query[parameterName],
  ]))
}

function readFirstInnerTubeValveFitmentShareQueryValue(query: LocationQuery, parameterName: string): string | null {
  const queryValue = query[parameterName]
  return (Array.isArray(queryValue) ? queryValue[0] : queryValue) ?? null
}

function parseInnerTubeValveFitmentShareInteger(value: string | null): number | null {
  if (value === null || !value.trim()) return null
  const parsedValue = Number(value)
  return Number.isFinite(parsedValue) ? Math.round(parsedValue) : null
}

function normalizeInnerTubeValveFitmentShareInteger(
  value: string | null, minimum: number, maximum: number, fallback: number,
): number {
  const parsedValue = parseInnerTubeValveFitmentShareInteger(value)
  return parsedValue === null ? fallback : Math.min(maximum, Math.max(minimum, parsedValue))
}

function normalizeSupportedInnerTubeValveFitmentShareLength(
  value: string | null, supportedLengths: number[], fallback: number,
): number {
  const parsedValue = parseInnerTubeValveFitmentShareInteger(value)
  return parsedValue !== null && supportedLengths.includes(parsedValue) ? parsedValue : fallback
}

export function restoreInnerTubeValveFitmentStateFromShareQuery(
  query: LocationQuery, metadata: InnerTubeValveFitmentShareInputMetadata,
): InnerTubeValveFitmentShareState {
  const supportedValveLengths = metadata.base_valve_length_options_mm
  const supportedExtenderLengths = metadata.extender_length_options_mm
  const sharedFitmentMode = readFirstInnerTubeValveFitmentShareQueryValue(query, 'fitment_mode')

  // Every URL starts from the same defaults, including when history removes
  // parameters. Never inherit values from the previous calculation.
  return {
    rimDepth: normalizeInnerTubeValveFitmentShareInteger(
      readFirstInnerTubeValveFitmentShareQueryValue(query, 'rim_depth'),
      metadata.rim_depth_min_mm, metadata.rim_depth_max_mm, 50,
    ),
    fitmentMode: sharedFitmentMode === 'manual' ? 'manual' : 'automatic',
    manualValveLength: normalizeSupportedInnerTubeValveFitmentShareLength(
      readFirstInnerTubeValveFitmentShareQueryValue(query, 'valve_length'),
      supportedValveLengths, supportedValveLengths.includes(60) ? 60 : (supportedValveLengths[0] ?? 60),
    ),
    manualExtenderLength: normalizeSupportedInnerTubeValveFitmentShareLength(
      readFirstInnerTubeValveFitmentShareQueryValue(query, 'extender_length'),
      supportedExtenderLengths, supportedExtenderLengths.includes(0) ? 0 : (supportedExtenderLengths[0] ?? 0),
    ),
    pumpHeadGripDepth: normalizeInnerTubeValveFitmentShareInteger(
      readFirstInnerTubeValveFitmentShareQueryValue(query, 'pump_head_grip_depth'),
      metadata.pump_head_grip_depth_min_mm, metadata.pump_head_grip_depth_max_mm,
      metadata.default_pump_head_grip_depth_mm,
    ),
    rimDepthUncertainty: normalizeInnerTubeValveFitmentShareInteger(
      readFirstInnerTubeValveFitmentShareQueryValue(query, 'rim_depth_uncertainty'),
      0, metadata.rim_depth_uncertainty_max_mm, metadata.default_rim_depth_uncertainty_mm,
    ),
  }
}

export function createInnerTubeValveFitmentShareQueryParameters(state: InnerTubeValveFitmentShareState): Record<string, string> {
  return {
    rim_depth: String(state.rimDepth),
    fitment_mode: state.fitmentMode,
    valve_length: String(state.manualValveLength),
    extender_length: String(state.manualExtenderLength),
    pump_head_grip_depth: String(state.pumpHeadGripDepth),
    rim_depth_uncertainty: String(state.rimDepthUncertainty),
  }
}
