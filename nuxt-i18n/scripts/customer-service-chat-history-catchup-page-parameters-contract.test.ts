import assert from 'node:assert/strict'
import { getCustomerServiceMessageCatchupPageParameters } from '../app/composables/chat/getCustomerServiceMessageCatchupPageParameters'

assert.deepEqual(
  getCustomerServiceMessageCatchupPageParameters(100, null),
  { limit: 100, latest: 'true' },
)

assert.deepEqual(
  getCustomerServiceMessageCatchupPageParameters(100, {
    created_at: '2026-11-24T12:00:00.000Z',
    id: '9223372036854775807',
  }),
  {
    limit: 100,
    after_created_at: '2026-11-24T12:00:00.000Z',
    after_id: '9223372036854775807',
  },
)

console.log('Customer-service history catch-up page parameter checks passed.')
