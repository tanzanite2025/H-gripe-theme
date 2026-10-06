import assert from 'node:assert/strict'
import {
  convertTireWidthInputBetweenUnits,
  parseTireWidthInputAsMillimeters,
} from '../app/data/tireguides/tireRimWidthInputConversion.js'

assert.equal(parseTireWidthInputAsMillimeters('28', 'mm'), 28)
assert.equal(parseTireWidthInputAsMillimeters('2.4', 'inch'), 61)
assert.equal(parseTireWidthInputAsMillimeters('2.1', 'inch'), 53)
assert.equal(parseTireWidthInputAsMillimeters('2.25', 'inch'), 57)
assert.equal(parseTireWidthInputAsMillimeters('', 'inch'), null)
assert.equal(parseTireWidthInputAsMillimeters('not a width', 'inch'), null)

assert.equal(convertTireWidthInputBetweenUnits('61', 'mm', 'inch'), '2.4')
assert.equal(convertTireWidthInputBetweenUnits('2.4', 'inch', 'mm'), '61')
assert.equal(convertTireWidthInputBetweenUnits('', 'mm', 'inch'), '')

console.log('Tire/rim width input conversion tests passed.')
