import assert from 'node:assert/strict'
import { warrantyResultFromResponse } from '../app/utils/warrantyResponse.js'

const warranty = { order_number: 'TZ-WARRANTY-123', status: 'valid' }
assert.deepEqual(warrantyResultFromResponse({
  code: 0,
  data: { success: true, data: warranty },
}), warranty)

assert.equal(warrantyResultFromResponse({ code: 0, data: { success: false, data: warranty } }), null)
assert.equal(warrantyResultFromResponse({ code: 0, data: { success: true } }), null)
assert.equal(warrantyResultFromResponse({ code: 1, data: { success: true, data: warranty } }), null)

console.log('Warranty response contract checks passed.')
