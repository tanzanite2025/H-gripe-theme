import assert from 'node:assert/strict'
import {
  commonTireWidthInputPresets,
  convertTireWidthInputBetweenUnits,
  inferTireWidthInputUnit,
  isLikelyInchTireWidthInput,
  parseTireWidthInputAsMillimeters,
} from '../app/data/tireguides/tireRimWidthInputConversion.js'

assert.equal(parseTireWidthInputAsMillimeters('28', 'mm'), 28)
assert.equal(parseTireWidthInputAsMillimeters('2.4', 'inch'), 61)
assert.equal(parseTireWidthInputAsMillimeters('2.1', 'inch'), 53)
assert.equal(parseTireWidthInputAsMillimeters('2.25', 'inch'), 57)
assert.equal(parseTireWidthInputAsMillimeters(2.4, 'inch'), 61)
assert.equal(parseTireWidthInputAsMillimeters('25C', 'mm'), 25)
assert.equal(parseTireWidthInputAsMillimeters('2.4″', 'mm'), 61)
assert.equal(parseTireWidthInputAsMillimeters('', 'inch'), null)
assert.equal(parseTireWidthInputAsMillimeters('not a width', 'inch'), null)

assert.equal(convertTireWidthInputBetweenUnits('61', 'mm', 'inch'), '2.4')
assert.equal(convertTireWidthInputBetweenUnits('2.4', 'inch', 'mm'), '61')
assert.equal(convertTireWidthInputBetweenUnits(2.4, 'mm', 'inch'), '0.09')
assert.equal(convertTireWidthInputBetweenUnits('', 'mm', 'inch'), '')

assert.equal(isLikelyInchTireWidthInput('2.4', 'mm'), true)
assert.equal(isLikelyInchTireWidthInput(2.4, 'mm'), true)
assert.equal(isLikelyInchTireWidthInput('2.4″', 'mm'), true)
assert.equal(isLikelyInchTireWidthInput('22', 'mm'), false)
assert.equal(isLikelyInchTireWidthInput('2.4', 'inch'), false)

assert.equal(inferTireWidthInputUnit('25C'), 'mm')
assert.equal(inferTireWidthInputUnit('61 mm'), 'mm')
assert.equal(inferTireWidthInputUnit('2.4″'), 'inch')
assert.equal(inferTireWidthInputUnit('2.4'), 'inch')
assert.equal(inferTireWidthInputUnit('25'), 'mm')
assert.equal(commonTireWidthInputPresets.find(preset => preset.id === '25c')?.millimeters, 25)
assert.equal(commonTireWidthInputPresets.find(preset => preset.id === '2-4-inch')?.millimeters, 61)

console.log('Tire/rim width input conversion tests passed.')
