import assert from 'node:assert/strict'
import {
  calculateTireRimEngineeringReference,
  TIRE_RIM_ENGINEERING_REFERENCE_MODEL_VERSION,
} from '../app/data/tireguides/tireRimEngineeringReferenceModel.js'

const hookless25MillimetreRimWith28MillimetreTire = calculateTireRimEngineeringReference(
  'hookless',
  25,
  28,
)
assert.equal(TIRE_RIM_ENGINEERING_REFERENCE_MODEL_VERSION, 'tire-rim-engineering-reference-v1')
assert.equal(hookless25MillimetreRimWith28MillimetreTire?.verdict, 'critical')
assert.equal(hookless25MillimetreRimWith28MillimetreTire?.verdictReason, 'below_minimum')

const hookless25MillimetreRimWith32MillimetreTire = calculateTireRimEngineeringReference(
  'hookless',
  25,
  32,
)
assert.equal(hookless25MillimetreRimWith32MillimetreTire?.verdict, 'recommended')
assert.equal(hookless25MillimetreRimWith32MillimetreTire?.inflatedTireWidthMm, 34.4)
assert.equal(hookless25MillimetreRimWith32MillimetreTire?.aeroTargetOuterWidthMm, 36.1)

const hookless19MillimetreTransitionalReference = calculateTireRimEngineeringReference(
  'hookless',
  19,
  28,
)
assert.equal(hookless19MillimetreTransitionalReference?.verdict, 'reference')
assert.equal(hookless19MillimetreTransitionalReference?.verdictReason, 'transitional_rim')

const invalidTireWidth = calculateTireRimEngineeringReference('hookless', 25, 28.5)
assert.equal(invalidTireWidth, null)

console.log('Tire/rim engineering reference model tests passed.')
