import type { TireRimWidthRange } from './tireRimWidthReferencePresentation'

export interface SchwalbeEtrto2024TireRimWidthReferenceRow {
  tireWidthRangeMm: TireRimWidthRange
  possibleRimInnerWidthRangeMm: TireRimWidthRange
}

/**
 * Display-only transcription of Schwalbe's 05/2024 PDF, not a fitment solver.
 * Every marked column in each source row is contiguous, so the two endpoints
 * preserve the complete set of possible widths without interpolation.
 */
export const schwalbeEtrto2024TireRimWidthReferenceData = {
  publisher: 'Schwalbe',
  standardEdition: 'ETRTO 2024',
  publicationMonth: '2024-05',
  sourceTitle: 'POSSIBLE COMBINATIONS OF TIRE WIDTHS AND INNER RIM WIDTHS',
  sourcePdfFilename: 'ETRTO_Standard_2024_Tire_Rim_Combinations.pdf',
  sourcePdfSha256: 'efcd9f55a1fdc6ae1900195f9a875406e5d77f4d1470eb17f5231066fb44c1bc',
  sourceImage: {
    path: '/images/guides/tireguides/choose/schwalbe-etrto-2024-published-tire-rim-width-reference-page.webp',
    width: 2800,
    height: 1980,
  },
  rows: [
    { tireWidthRangeMm: { min: 20, max: 21 }, possibleRimInnerWidthRangeMm: { min: 15, max: 17 } },
    { tireWidthRangeMm: { min: 22, max: 24 }, possibleRimInnerWidthRangeMm: { min: 15, max: 20 } },
    { tireWidthRangeMm: { min: 25, max: 27 }, possibleRimInnerWidthRangeMm: { min: 15, max: 22 } },
    { tireWidthRangeMm: { min: 28, max: 28 }, possibleRimInnerWidthRangeMm: { min: 16, max: 23 } },
    { tireWidthRangeMm: { min: 29, max: 34 }, possibleRimInnerWidthRangeMm: { min: 16, max: 25 } },
    { tireWidthRangeMm: { min: 35, max: 46 }, possibleRimInnerWidthRangeMm: { min: 17, max: 27 } },
    { tireWidthRangeMm: { min: 47, max: 57 }, possibleRimInnerWidthRangeMm: { min: 17, max: 30 } },
    { tireWidthRangeMm: { min: 58, max: 65 }, possibleRimInnerWidthRangeMm: { min: 21, max: 35 } },
    { tireWidthRangeMm: { min: 66, max: 71 }, possibleRimInnerWidthRangeMm: { min: 25, max: 43 } },
    { tireWidthRangeMm: { min: 72, max: 83 }, possibleRimInnerWidthRangeMm: { min: 31, max: 53 } },
    { tireWidthRangeMm: { min: 84, max: 95 }, possibleRimInnerWidthRangeMm: { min: 41, max: 64 } },
    { tireWidthRangeMm: { min: 96, max: 113 }, possibleRimInnerWidthRangeMm: { min: 59, max: 89 } },
    { tireWidthRangeMm: { min: 114, max: 132 }, possibleRimInnerWidthRangeMm: { min: 72, max: 100 } },
  ] satisfies SchwalbeEtrto2024TireRimWidthReferenceRow[],
}
