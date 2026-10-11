import { tireRimDtSwissReferenceData } from '../app/data/tireguides/tireRimDtSwissReferenceData'

const rows = tireRimDtSwissReferenceData.rows
const hooklessRows = rows.filter(row => row.rim_system === 'hookless')
const hookedRows = rows.filter(row => row.rim_system === 'hooked')

if (rows.length !== 57 || hooklessRows.length !== 25 || hookedRows.length !== 32) {
  throw new Error(`Unexpected DT Swiss comparison row counts: total=${rows.length}, hookless=${hooklessRows.length}, hooked=${hookedRows.length}`)
}

const hookless60MillimeterRow = hooklessRows.find(row => row.tire_width_mm === 60)
if (
  !hookless60MillimeterRow
  || hookless60MillimeterRow.recommended.length !== 2
  || hookless60MillimeterRow.recommended[0]?.min !== 26
  || hookless60MillimeterRow.recommended[1]?.min !== 28
  || hookless60MillimeterRow.possible.length !== 2
) {
  throw new Error('The hookless 60 mm DT Swiss source row was not preserved exactly.')
}

if (
  !tireRimDtSwissReferenceData.source.name.includes('DT Swiss')
  || !tireRimDtSwissReferenceData.methodology.interpolated_rows.includes('does not interpolate')
) {
  throw new Error('The DT Swiss data must retain its source attribution and display-only boundary.')
}

console.info('DT Swiss tire/rim reference rows and source metadata are stable.')
