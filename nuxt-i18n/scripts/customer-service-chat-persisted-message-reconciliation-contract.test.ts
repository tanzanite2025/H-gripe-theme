import assert from 'node:assert/strict'
import { reconcileCustomerServiceMessagesByStableServerAndClientIds } from '../app/composables/chat/reconcileCustomerServiceMessagesByStableServerAndClientIds'

const normalizeServerMessage = (message: Record<string, any>) => ({
  ...message,
  id: message.cursor_id || message.id,
  sync_state: 'persisted',
})

const localFailedMessage = {
  id: 'local-1732450000000-a1b2c3',
  client_message_id: 'local-1732450000000-a1b2c3',
  message: 'Sent before the response was lost',
  created_at: '2026-11-24T10:00:00.000Z',
  sync_state: 'failed',
}
const persistedServerMessage = {
  id: 84,
  cursor_id: '84',
  client_message_id: localFailedMessage.client_message_id,
  message: localFailedMessage.message,
  created_at: localFailedMessage.created_at,
}

const reconciledHistoryMessage = reconcileCustomerServiceMessagesByStableServerAndClientIds(
  [localFailedMessage],
  [persistedServerMessage],
  normalizeServerMessage,
)
assert.equal(reconciledHistoryMessage.length, 1)
assert.equal(reconciledHistoryMessage[0]?.id, '84')
assert.equal(reconciledHistoryMessage[0]?.sync_state, 'persisted')
assert.equal(reconciledHistoryMessage[0]?.client_message_id, localFailedMessage.client_message_id)

const reconciledRetryResponse = reconcileCustomerServiceMessagesByStableServerAndClientIds(
  [persistedServerMessage, localFailedMessage],
  [persistedServerMessage],
  normalizeServerMessage,
)
assert.equal(reconciledRetryResponse.length, 1)
assert.equal(reconciledRetryResponse[0]?.id, '84')

const unrelatedFailedMessage = {
  ...localFailedMessage,
  id: 'local-unrelated-message',
  client_message_id: 'local-unrelated-message',
}
const preservesUnrelatedLocalMessage = reconcileCustomerServiceMessagesByStableServerAndClientIds(
  [localFailedMessage, unrelatedFailedMessage],
  [persistedServerMessage],
  normalizeServerMessage,
)
assert.deepEqual(preservesUnrelatedLocalMessage.map((message) => message.id), ['84', 'local-unrelated-message'])

console.log('Customer-service persisted-message reconciliation checks passed.')
