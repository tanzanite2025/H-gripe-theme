import type { TireRimReferenceRimSystem, TireRimWidthRange } from './tireRimWidthReferencePresentation'

export interface TireRimDtSwissReferenceTableRow {
  rim_system: TireRimReferenceRimSystem
  tire_width_mm: number
  inch: string
  recommended: TireRimWidthRange[]
  possible: TireRimWidthRange[]
}

export interface TireRimDtSwissReferenceData {
  model_version: string
  knowledge_as_of: string
  source: { name: string; provenance: string }
  rows: TireRimDtSwissReferenceTableRow[]
  methodology: { exact_rows: string; interpolated_rows: string; possible_rows: string; derived_metrics: string }
  limitations: string[]
}

export const tireRimDtSwissReferenceData: TireRimDtSwissReferenceData = {
  model_version: 'dt-swiss-tss-tc-v1',
  knowledge_as_of: '2026-09-28',
  source: {
    name: 'DT Swiss TSS/TC tire-width and rim-inner-width reference chart',
    provenance: 'Repository transcription of the published chart; not a model-specific certification',
  },
  rows: [
    { rim_system: 'hookless', tire_width_mm: 30, inch: '1.20', recommended: [], possible: [ { min: 23, max: 25 } ] },
    { rim_system: 'hookless', tire_width_mm: 32, inch: '1.25', recommended: [ { min: 23, max: 25 } ], possible: [] },
    { rim_system: 'hookless', tire_width_mm: 34, inch: '1.35', recommended: [ { min: 23, max: 25 } ], possible: [] },
    { rim_system: 'hookless', tire_width_mm: 36, inch: '1.40', recommended: [ { min: 23, max: 25 } ], possible: [ { min: 26, max: 27 } ] },
    { rim_system: 'hookless', tire_width_mm: 38, inch: '1.50', recommended: [ { min: 23, max: 25 } ], possible: [ { min: 26, max: 27 } ] },
    { rim_system: 'hookless', tire_width_mm: 41, inch: '1.60', recommended: [ { min: 23, max: 25 } ], possible: [ { min: 26, max: 27 } ] },
    { rim_system: 'hookless', tire_width_mm: 43, inch: '1.70', recommended: [ { min: 23, max: 25 }, { min: 26, max: 27 } ], possible: [] },
    { rim_system: 'hookless', tire_width_mm: 47, inch: '1.85', recommended: [ { min: 23, max: 25 }, { min: 26, max: 27 } ], possible: [ { min: 28, max: 30 } ] },
    { rim_system: 'hookless', tire_width_mm: 50, inch: '2.00', recommended: [ { min: 23, max: 25 }, { min: 26, max: 27 } ], possible: [ { min: 28, max: 30 } ] },
    { rim_system: 'hookless', tire_width_mm: 52, inch: '2.00', recommended: [ { min: 26, max: 27 } ], possible: [ { min: 23, max: 25 }, { min: 28, max: 30 } ] },
    { rim_system: 'hookless', tire_width_mm: 53, inch: '2.10', recommended: [ { min: 26, max: 27 }, { min: 28, max: 30 } ], possible: [ { min: 23, max: 25 } ] },
    { rim_system: 'hookless', tire_width_mm: 56, inch: '2.20', recommended: [ { min: 26, max: 27 }, { min: 28, max: 30 } ], possible: [ { min: 23, max: 25 } ] },
    { rim_system: 'hookless', tire_width_mm: 60, inch: '2.40', recommended: [ { min: 26, max: 27 }, { min: 28, max: 30 } ], possible: [ { min: 23, max: 25 }, { min: 31, max: 35 } ] },
    { rim_system: 'hookless', tire_width_mm: 64, inch: '2.50', recommended: [ { min: 28, max: 30 }, { min: 31, max: 35 } ], possible: [ { min: 23, max: 25 }, { min: 26, max: 27 }, { min: 36, max: 40 } ] },
    { rim_system: 'hookless', tire_width_mm: 66, inch: '2.60', recommended: [ { min: 28, max: 30 }, { min: 31, max: 35 } ], possible: [ { min: 23, max: 25 }, { min: 26, max: 27 }, { min: 36, max: 40 } ] },
    { rim_system: 'hookless', tire_width_mm: 69, inch: '2.70', recommended: [ { min: 28, max: 30 }, { min: 31, max: 35 } ], possible: [ { min: 26, max: 27 }, { min: 36, max: 40 } ] },
    { rim_system: 'hookless', tire_width_mm: 71, inch: '2.80', recommended: [ { min: 31, max: 35 }, { min: 36, max: 40 } ], possible: [ { min: 26, max: 27 }, { min: 28, max: 30 } ] },
    { rim_system: 'hookless', tire_width_mm: 75, inch: '3.00', recommended: [ { min: 36, max: 40 } ], possible: [ { min: 31, max: 35 } ] },
    { rim_system: 'hookless', tire_width_mm: 85, inch: '3.30', recommended: [ { min: 36, max: 40 } ], possible: [ { min: 31, max: 35 } ] },
    { rim_system: 'hookless', tire_width_mm: 96, inch: '3.70', recommended: [], possible: [ { min: 76, max: 76 } ] },
    { rim_system: 'hookless', tire_width_mm: 102, inch: '4.00', recommended: [ { min: 76, max: 76 } ], possible: [] },
    { rim_system: 'hookless', tire_width_mm: 110, inch: '4.30', recommended: [ { min: 76, max: 76 } ], possible: [] },
    { rim_system: 'hookless', tire_width_mm: 115, inch: '4.50', recommended: [ { min: 76, max: 76 } ], possible: [] },
    { rim_system: 'hookless', tire_width_mm: 122, inch: '4.80', recommended: [ { min: 76, max: 76 } ], possible: [] },
    { rim_system: 'hookless', tire_width_mm: 127, inch: '5.00', recommended: [ { min: 76, max: 76 } ], possible: [] },
    { rim_system: 'hooked', tire_width_mm: 20, inch: '0.80', recommended: [], possible: [ { min: 15, max: 17 } ] },
    { rim_system: 'hooked', tire_width_mm: 23, inch: '0.90', recommended: [ { min: 15, max: 17 } ], possible: [ { min: 18, max: 20 } ] },
    { rim_system: 'hooked', tire_width_mm: 25, inch: '1.00', recommended: [ { min: 15, max: 17 }, { min: 18, max: 20 } ], possible: [ { min: 21, max: 22 } ] },
    { rim_system: 'hooked', tire_width_mm: 26, inch: '1.00', recommended: [ { min: 18, max: 20 }, { min: 21, max: 22 } ], possible: [ { min: 15, max: 17 } ] },
    { rim_system: 'hooked', tire_width_mm: 28, inch: '1.10', recommended: [ { min: 18, max: 20 }, { min: 21, max: 22 } ], possible: [ { min: 15, max: 17 } ] },
    { rim_system: 'hooked', tire_width_mm: 29, inch: '1.15', recommended: [ { min: 18, max: 20 }, { min: 21, max: 22 } ], possible: [ { min: 15, max: 17 }, { min: 23, max: 25 } ] },
    { rim_system: 'hooked', tire_width_mm: 30, inch: '1.20', recommended: [ { min: 18, max: 20 }, { min: 21, max: 22 } ], possible: [ { min: 15, max: 17 }, { min: 23, max: 25 } ] },
    { rim_system: 'hooked', tire_width_mm: 32, inch: '1.25', recommended: [ { min: 21, max: 22 }, { min: 23, max: 25 } ], possible: [ { min: 15, max: 17 }, { min: 18, max: 20 } ] },
    { rim_system: 'hooked', tire_width_mm: 34, inch: '1.35', recommended: [ { min: 21, max: 22 }, { min: 23, max: 25 } ], possible: [ { min: 15, max: 17 }, { min: 18, max: 20 } ] },
    { rim_system: 'hooked', tire_width_mm: 35, inch: '1.40', recommended: [ { min: 21, max: 22 }, { min: 23, max: 25 } ], possible: [ { min: 15, max: 17 }, { min: 18, max: 20 }, { min: 26, max: 27 } ] },
    { rim_system: 'hooked', tire_width_mm: 36, inch: '1.40', recommended: [ { min: 23, max: 25 } ], possible: [ { min: 18, max: 20 }, { min: 21, max: 22 }, { min: 26, max: 27 } ] },
    { rim_system: 'hooked', tire_width_mm: 38, inch: '1.50', recommended: [ { min: 23, max: 25 } ], possible: [ { min: 18, max: 20 }, { min: 21, max: 22 }, { min: 26, max: 27 } ] },
    { rim_system: 'hooked', tire_width_mm: 41, inch: '1.60', recommended: [ { min: 23, max: 25 } ], possible: [ { min: 18, max: 20 }, { min: 21, max: 22 }, { min: 26, max: 27 } ] },
    { rim_system: 'hooked', tire_width_mm: 43, inch: '1.70', recommended: [ { min: 23, max: 25 }, { min: 26, max: 27 } ], possible: [ { min: 18, max: 20 }, { min: 21, max: 22 } ] },
    { rim_system: 'hooked', tire_width_mm: 47, inch: '1.85', recommended: [ { min: 23, max: 25 }, { min: 26, max: 27 } ], possible: [ { min: 18, max: 20 }, { min: 21, max: 22 }, { min: 28, max: 30 } ] },
    { rim_system: 'hooked', tire_width_mm: 50, inch: '2.00', recommended: [ { min: 23, max: 25 }, { min: 26, max: 27 } ], possible: [ { min: 18, max: 20 }, { min: 21, max: 22 }, { min: 28, max: 30 } ] },
    { rim_system: 'hooked', tire_width_mm: 52, inch: '2.00', recommended: [ { min: 26, max: 27 } ], possible: [ { min: 18, max: 20 }, { min: 21, max: 22 }, { min: 23, max: 25 }, { min: 28, max: 30 } ] },
    { rim_system: 'hooked', tire_width_mm: 53, inch: '2.10', recommended: [ { min: 26, max: 27 }, { min: 28, max: 30 } ], possible: [ { min: 18, max: 20 }, { min: 21, max: 22 }, { min: 23, max: 25 } ] },
    { rim_system: 'hooked', tire_width_mm: 56, inch: '2.20', recommended: [ { min: 26, max: 27 }, { min: 28, max: 30 } ], possible: [ { min: 18, max: 20 }, { min: 21, max: 22 }, { min: 23, max: 25 } ] },
    { rim_system: 'hooked', tire_width_mm: 60, inch: '2.40', recommended: [ { min: 26, max: 27 }, { min: 28, max: 30 } ], possible: [ { min: 21, max: 22 }, { min: 23, max: 25 }, { min: 31, max: 35 } ] },
    { rim_system: 'hooked', tire_width_mm: 64, inch: '2.50', recommended: [ { min: 28, max: 30 }, { min: 31, max: 35 } ], possible: [ { min: 21, max: 22 }, { min: 23, max: 25 }, { min: 26, max: 27 } ] },
    { rim_system: 'hooked', tire_width_mm: 66, inch: '2.60', recommended: [ { min: 28, max: 30 }, { min: 31, max: 35 } ], possible: [ { min: 23, max: 25 }, { min: 26, max: 27 }, { min: 36, max: 40 } ] },
    { rim_system: 'hooked', tire_width_mm: 69, inch: '2.70', recommended: [ { min: 28, max: 30 }, { min: 31, max: 35 } ], possible: [ { min: 26, max: 27 }, { min: 36, max: 40 } ] },
    { rim_system: 'hooked', tire_width_mm: 71, inch: '2.80', recommended: [ { min: 31, max: 35 }, { min: 36, max: 40 } ], possible: [ { min: 26, max: 27 }, { min: 28, max: 30 } ] },
    { rim_system: 'hooked', tire_width_mm: 75, inch: '3.00', recommended: [ { min: 36, max: 40 } ], possible: [ { min: 31, max: 35 } ] },
    { rim_system: 'hooked', tire_width_mm: 85, inch: '3.30', recommended: [ { min: 36, max: 40 } ], possible: [ { min: 31, max: 35 } ] },
    { rim_system: 'hooked', tire_width_mm: 96, inch: '3.70', recommended: [], possible: [ { min: 76, max: 76 } ] },
    { rim_system: 'hooked', tire_width_mm: 102, inch: '4.00', recommended: [ { min: 76, max: 76 } ], possible: [] },
    { rim_system: 'hooked', tire_width_mm: 110, inch: '4.30', recommended: [ { min: 76, max: 76 } ], possible: [] },
    { rim_system: 'hooked', tire_width_mm: 115, inch: '4.50', recommended: [ { min: 76, max: 76 } ], possible: [] },
    { rim_system: 'hooked', tire_width_mm: 122, inch: '4.80', recommended: [ { min: 76, max: 76 } ], possible: [] },
    { rim_system: 'hooked', tire_width_mm: 127, inch: '5.00', recommended: [ { min: 76, max: 76 } ], possible: [] },
  ],
  methodology: {
    exact_rows: 'Published DT Swiss chart ranges transcribed for display as source references.',
    interpolated_rows: 'This table does not interpolate widths or calculate compatibility.',
    possible_rows: 'The source chart possible ranges remain labeled as possible and are not engineering recommendations.',
    derived_metrics: 'No tire-width, aerodynamic, or safety calculation is derived from this comparison table.',
  },
  limitations: [
    'This DT Swiss chart is a comparison reference, not a model-specific compatibility certification.',
    'Possible ranges preserve the source chart labels and do not establish approval for a specific tire and rim.',
    'Use the exact tire, rim, wheelset, pressure, and frame specifications for a real installation.',
  ],
}
