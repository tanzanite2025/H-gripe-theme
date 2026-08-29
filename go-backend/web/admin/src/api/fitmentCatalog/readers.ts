import {
  requireApiArray,
  requireApiBooleanField,
  requireApiNumberField,
  requireApiObject,
  requireApiStringField,
} from '@/utils/apiResponse'
import type {
  FitmentYearMode,
  ForkFitmentEntry,
  FrameFitmentEntry,
  HubSpecification,
  HubSpecificationAxleType,
  HubSpecificationPosition,
} from './types'

export const readNullableNumber = (value: Record<string, unknown>, field: string, endpoint: string): number | null => {
  if (!(field in value)) {
    throw new Error(`[CRITICAL] Invalid API response for ${endpoint}: required field "${field}" is missing`)
  }
  const fieldValue = value[field]
  if (fieldValue === null) return null
  if (typeof fieldValue !== 'number' || !Number.isFinite(fieldValue)) {
    throw new Error(`[CRITICAL] Invalid API response for ${endpoint}: field "${field}" must be a number or null`)
  }
  return fieldValue
}

export const readYearMode = (value: unknown, endpoint: string): FitmentYearMode => {
  if (value === 'single' || value === 'range' || value === 'all' || value === 'unknown') return value
  throw new Error(`[CRITICAL] Invalid API response for ${endpoint}: unsupported year_mode`)
}

export const readHubPosition = (value: unknown, endpoint: string): HubSpecificationPosition => {
  if (value === 'front' || value === 'rear') return value
  throw new Error(`[CRITICAL] Invalid API response for ${endpoint}: unsupported position`)
}

export const readHubAxleType = (value: unknown, endpoint: string): HubSpecificationAxleType => {
  if (value === 'quick_release' || value === 'thru_axle' || value === 'bolt_on' || value === 'other') return value
  throw new Error(`[CRITICAL] Invalid API response for ${endpoint}: unsupported axle_type`)
}

export const readHubSpecification = (value: unknown, endpoint: string): HubSpecification => {
  const specification = requireApiObject(value, endpoint, 'hub specification')
  return {
    id: requireApiNumberField(specification, 'id', endpoint),
    spec_code: requireApiStringField(specification, 'spec_code', endpoint),
    display_name: requireApiStringField(specification, 'display_name', endpoint),
    position: readHubPosition(requireApiStringField(specification, 'position', endpoint), endpoint),
    axle_type: readHubAxleType(requireApiStringField(specification, 'axle_type', endpoint), endpoint),
    axle_spacing_mm: requireApiNumberField(specification, 'axle_spacing_mm', endpoint),
    wr_mm: readNullableNumber(specification, 'wr_mm', endpoint),
    wl_mm: readNullableNumber(specification, 'wl_mm', endpoint),
    pcdr_mm: readNullableNumber(specification, 'pcdr_mm', endpoint),
    pcdl_mm: readNullableNumber(specification, 'pcdl_mm', endpoint),
    notes: requireApiStringField(specification, 'notes', endpoint),
    is_enabled: requireApiBooleanField(specification, 'is_enabled', endpoint),
    sort_order: requireApiNumberField(specification, 'sort_order', endpoint),
    frame_reference_count: requireApiNumberField(specification, 'frame_reference_count', endpoint),
    fork_reference_count: requireApiNumberField(specification, 'fork_reference_count', endpoint),
    created_at: specification.created_at === undefined ? undefined : String(specification.created_at),
    updated_at: specification.updated_at === undefined ? undefined : String(specification.updated_at),
  }
}

export const readForkFitmentEntry = (value: unknown, endpoint: string): ForkFitmentEntry => {
  const entry = requireApiObject(value, endpoint, 'entry')
  const hubSpecifications = entry.hub_specifications === null
    ? []
    : requireApiArray<HubSpecification>(entry.hub_specifications, endpoint, 'field "hub_specifications"')
      .map((specification) => readHubSpecification(specification, endpoint))
  return {
    id: requireApiNumberField(entry, 'id', endpoint),
    brand_name: requireApiStringField(entry, 'brand_name', endpoint),
    model_name: requireApiStringField(entry, 'model_name', endpoint),
    series_name: requireApiStringField(entry, 'series_name', endpoint),
    generation_name: requireApiStringField(entry, 'generation_name', endpoint),
    year_mode: readYearMode(requireApiStringField(entry, 'year_mode', endpoint), endpoint),
    year_from: readNullableNumber(entry, 'year_from', endpoint),
    year_to: readNullableNumber(entry, 'year_to', endpoint),
    market_code: requireApiStringField(entry, 'market_code', endpoint),
    notes: requireApiStringField(entry, 'notes', endpoint),
    is_enabled: requireApiBooleanField(entry, 'is_enabled', endpoint),
    sort_order: requireApiNumberField(entry, 'sort_order', endpoint),
    hub_specifications: hubSpecifications,
    hub_specification_count: requireApiNumberField(entry, 'hub_specification_count', endpoint),
    created_at: entry.created_at === undefined ? undefined : String(entry.created_at),
    updated_at: entry.updated_at === undefined ? undefined : String(entry.updated_at),
  }
}

export const readFrameFitmentEntry = (value: unknown, endpoint: string): FrameFitmentEntry => {
  const entry = requireApiObject(value, endpoint, 'entry')
  const hubSpecifications = entry.hub_specifications === null
    ? []
    : requireApiArray<HubSpecification>(entry.hub_specifications, endpoint, 'field "hub_specifications"')
      .map((specification) => readHubSpecification(specification, endpoint))
  return {
    id: requireApiNumberField(entry, 'id', endpoint),
    brand_name: requireApiStringField(entry, 'brand_name', endpoint),
    model_name: requireApiStringField(entry, 'model_name', endpoint),
    series_name: requireApiStringField(entry, 'series_name', endpoint),
    generation_name: requireApiStringField(entry, 'generation_name', endpoint),
    year_mode: readYearMode(requireApiStringField(entry, 'year_mode', endpoint), endpoint),
    year_from: readNullableNumber(entry, 'year_from', endpoint),
    year_to: readNullableNumber(entry, 'year_to', endpoint),
    market_code: requireApiStringField(entry, 'market_code', endpoint),
    notes: requireApiStringField(entry, 'notes', endpoint),
    is_enabled: requireApiBooleanField(entry, 'is_enabled', endpoint),
    sort_order: requireApiNumberField(entry, 'sort_order', endpoint),
    hub_specifications: hubSpecifications,
    hub_specification_count: requireApiNumberField(entry, 'hub_specification_count', endpoint),
    created_at: entry.created_at === undefined ? undefined : String(entry.created_at),
    updated_at: entry.updated_at === undefined ? undefined : String(entry.updated_at),
  }
}
