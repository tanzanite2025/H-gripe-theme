import assert from 'node:assert/strict'
import {
  clearStoredChatConversationId,
  countUnreadCustomerServiceMessagesForOwner,
  createEmptyChatRoom,
  getChatDraftStorageKey,
  getChatOwnerStorageScope,
  getChatSessionStorageKey,
  getChatStorageKey,
  getOrCreateAnonymousChatOwnerId,
  hasLocalChatHistory,
  loadChatConversationIdFromStorage,
  loadChatDraftFromStorage,
  loadChatRoomFromStorage,
  migrateChatDraftStorage,
  migrateChatRoomStorage,
  restoreChatRoomDraftFromStoredSources,
  saveChatConversationIdToStorage,
  saveChatDraftToStorage,
  saveChatRoomToStorage,
} from '../app/composables/chat/useChatStorage'

const originalWindowDescriptor = Object.getOwnPropertyDescriptor(globalThis, 'window')
const originalLocalStorageDescriptor = Object.getOwnPropertyDescriptor(globalThis, 'localStorage')
const storedValues = new Map<string, string>()
const browserLocalStorage = {
  getItem(key: string) {
    return storedValues.get(key) ?? null
  },
  key(index: number) {
    return Array.from(storedValues.keys())[index] ?? null
  },
  get length() {
    return storedValues.size
  },
  setItem(key: string, value: string) {
    storedValues.set(key, String(value))
  },
  removeItem(key: string) {
    storedValues.delete(key)
  },
  clear() {
    storedValues.clear()
  },
}

Object.defineProperty(globalThis, 'window', {
  configurable: true,
  value: { localStorage: browserLocalStorage },
})
Object.defineProperty(globalThis, 'localStorage', {
  configurable: true,
  value: browserLocalStorage,
})

const restoreBrowserStorageGlobals = () => {
  if (originalWindowDescriptor) {
    Object.defineProperty(globalThis, 'window', originalWindowDescriptor)
  } else {
    Reflect.deleteProperty(globalThis, 'window')
  }

  if (originalLocalStorageDescriptor) {
    Object.defineProperty(globalThis, 'localStorage', originalLocalStorageDescriptor)
  } else {
    Reflect.deleteProperty(globalThis, 'localStorage')
  }
}

const testAnonymousChatOwnerIdentityAndMemberStorageScopes = () => {
  const guestIdentity = getOrCreateAnonymousChatOwnerId()
  assert.ok(guestIdentity.startsWith('visitor_'))
  assert.equal(getOrCreateAnonymousChatOwnerId(), guestIdentity)
  assert.equal(getChatOwnerStorageScope(null, guestIdentity), `guest:${guestIdentity}`)
  assert.equal(getChatOwnerStorageScope(42, guestIdentity), 'member:42')
  assert.notEqual(getChatOwnerStorageScope(42, guestIdentity), getChatOwnerStorageScope(null, guestIdentity))
  assert.notEqual(
    getChatStorageKey('', 7, 'guest:owner-one'),
    getChatStorageKey('', 7, 'member:42'),
  )
  assert.notEqual(
    getChatDraftStorageKey('', 7, 'guest:owner-one'),
    getChatDraftStorageKey('', 7, 'member:42'),
  )
}

const testConversationIdStorageAndExpectedIdClearGuard = () => {
  const guestScope = 'guest:owner-one'
  const memberScope = 'member:42'

  saveChatConversationIdToStorage(guestScope, 'guest-conversation')
  saveChatConversationIdToStorage(memberScope, 'member-conversation')

  assert.equal(loadChatConversationIdFromStorage(guestScope), 'guest-conversation')
  assert.equal(loadChatConversationIdFromStorage(memberScope), 'member-conversation')

  clearStoredChatConversationId(memberScope, 'another-conversation')
  assert.equal(loadChatConversationIdFromStorage(memberScope), 'member-conversation')

  clearStoredChatConversationId(memberScope, 'member-conversation')
  assert.equal(loadChatConversationIdFromStorage(memberScope), '')
}

const testVersionedChatDraftSaveRestoreExpiryAndFutureVersionPreservation = () => {
  const draftStorageKey = getChatDraftStorageKey('conversation-one', 7)
  saveChatDraftToStorage(draftStorageKey, 'Unsent support question')
  assert.equal(loadChatDraftFromStorage(draftStorageKey), 'Unsent support question')

  saveChatDraftToStorage(draftStorageKey, '')
  assert.equal(loadChatDraftFromStorage(draftStorageKey), null)

  const expiredDraftStorageKey = getChatDraftStorageKey('conversation-expired', 7)
  browserLocalStorage.setItem(expiredDraftStorageKey, JSON.stringify({
    version: 1,
    draft: 'Old draft',
    updatedAt: new Date(Date.now() - 6 * 24 * 60 * 60 * 1000).toISOString(),
  }))
  assert.equal(loadChatDraftFromStorage(expiredDraftStorageKey), null)
  assert.equal(browserLocalStorage.getItem(expiredDraftStorageKey), null)

  const futureDraftStorageKey = getChatDraftStorageKey('conversation-future', 7)
  const futureVersionRecord = JSON.stringify({ version: 2, draft: 'Newer app draft', updatedAt: new Date().toISOString() })
  browserLocalStorage.setItem(futureDraftStorageKey, futureVersionRecord)
  assert.equal(loadChatDraftFromStorage(futureDraftStorageKey), null)
  assert.equal(browserLocalStorage.getItem(futureDraftStorageKey), futureVersionRecord)
}

const testVersionedChatDraftWinsOverRepeatedLegacyRoomRestore = () => {
  assert.equal(restoreChatRoomDraftFromStoredSources({
    currentDraft: '',
    legacyRoomDraft: '',
    versionedDraft: 'Versioned draft',
  }), 'Versioned draft')
  assert.equal(restoreChatRoomDraftFromStoredSources({
    currentDraft: 'New in-memory draft',
    legacyRoomDraft: 'Stale room snapshot',
    versionedDraft: null,
  }), 'New in-memory draft')
  assert.equal(restoreChatRoomDraftFromStoredSources({
    currentDraft: '',
    legacyRoomDraft: 'Legacy room draft',
    versionedDraft: null,
  }), 'Legacy room draft')
  assert.equal(restoreChatRoomDraftFromStoredSources({
    currentDraft: '',
    legacyRoomDraft: 'Stale room snapshot',
    versionedDraft: '',
  }), '')
}

const testPendingChatRoomMigrationMergesMessagesAndRetainsFailedLocalSend = () => {
  const ownerStorageScope = 'guest:owner-one'
  const pendingStorageKey = getChatStorageKey('', 7, ownerStorageScope)
  const conversationStorageKey = getChatStorageKey('conversation-one', 7, ownerStorageScope)
  const pendingRoom = createEmptyChatRoom()
  pendingRoom.messages = [{
    id: 'local-pending-message',
    client_message_id: 'local-1732450000000-a1b2c3',
    message: 'Offline send',
    created_at: new Date().toISOString(),
    sync_state: 'failed',
  }]
  pendingRoom.newMessage = 'Pending legacy draft'
  const conversationRoom = createEmptyChatRoom()
  conversationRoom.messages = [{
    id: 15,
    message: 'Server history',
    created_at: new Date(Date.now() - 1000).toISOString(),
    sync_state: 'persisted',
  }]

  saveChatRoomToStorage(pendingStorageKey, pendingRoom)
  saveChatRoomToStorage(conversationStorageKey, conversationRoom)

  assert.equal(migrateChatRoomStorage(pendingStorageKey, conversationStorageKey), true)
  const migratedRoom = loadChatRoomFromStorage(conversationStorageKey)
  assert.deepEqual(migratedRoom?.room.messages.map(message => message.id), [15, 'local-pending-message'])
  assert.equal(migratedRoom?.room.messages[1]?.sync_state, 'failed')
  assert.equal(migratedRoom?.room.messages[1]?.client_message_id, 'local-1732450000000-a1b2c3')
  assert.equal(migratedRoom?.room.newMessage, 'Pending legacy draft')
  assert.equal(browserLocalStorage.getItem(pendingStorageKey), null)
}

const testPendingDraftMigrationDoesNotReplaceNewerConversationDraft = () => {
  const ownerStorageScope = 'guest:owner-one'
  const pendingDraftStorageKey = getChatDraftStorageKey('', 7, ownerStorageScope)
  const conversationDraftStorageKey = getChatDraftStorageKey('conversation-one', 7, ownerStorageScope)
  saveChatDraftToStorage(pendingDraftStorageKey, 'Pending draft')
  saveChatDraftToStorage(conversationDraftStorageKey, 'Conversation draft')

  assert.equal(migrateChatDraftStorage(pendingDraftStorageKey, conversationDraftStorageKey), true)
  assert.equal(loadChatDraftFromStorage(conversationDraftStorageKey), 'Conversation draft')
  assert.equal(loadChatDraftFromStorage(pendingDraftStorageKey), null)
  assert.ok(getChatSessionStorageKey('member:42').includes('member%3A42'))
}

const testLocalHistoryIsLimitedToCurrentOwnerScope = () => {
  const guestRoom = createEmptyChatRoom()
  guestRoom.messages = [{ id: 1, message: 'Guest message', created_at: new Date().toISOString() }]
  const memberRoom = createEmptyChatRoom()
  memberRoom.messages = [{ id: 2, message: 'Member message', created_at: new Date().toISOString() }]

  saveChatRoomToStorage(getChatStorageKey('guest-conversation', 7, 'guest:owner-one'), guestRoom)
  saveChatRoomToStorage(getChatStorageKey('member-conversation', 7, 'member:42'), memberRoom)

  assert.equal(hasLocalChatHistory('guest:owner-one'), true)
  assert.equal(hasLocalChatHistory('member:42'), true)
  assert.equal(hasLocalChatHistory('member:84'), false)
}

const testUnreadChatCountIsLimitedToCurrentOwnerAndIgnoresNonRoomStorage = () => {
  storedValues.clear()
  const guestScope = 'guest:visitor_unread-count'
  const memberScope = 'member:42'
  browserLocalStorage.setItem('tz_chat_anonymous_owner_id_v1', 'visitor_unread-count')
  browserLocalStorage.setItem(getChatSessionStorageKey(guestScope), JSON.stringify({
    version: 1,
    conversationId: 'guest-conversation',
  }))
  browserLocalStorage.setItem(getChatStorageKey('guest-conversation', 7, guestScope), JSON.stringify({
    messages: [
      { is_agent: true, is_read: false },
      { is_agent: true, is_read: true },
      { is_agent: false, is_read: false },
    ],
  }))
  browserLocalStorage.setItem(getChatStorageKey('member-conversation', 7, memberScope), JSON.stringify({
    messages: [
      { is_agent: true, is_read: false },
      { is_agent: true, is_read: false },
    ],
  }))
  browserLocalStorage.setItem(getChatStorageKey('malformed-room', 7, guestScope), 'visitor_unread-count')

  assert.equal(countUnreadCustomerServiceMessagesForOwner(guestScope), 1)
  assert.equal(countUnreadCustomerServiceMessagesForOwner(memberScope), 2)
  assert.equal(countUnreadCustomerServiceMessagesForOwner('member:84'), 0)
}

try {
  testAnonymousChatOwnerIdentityAndMemberStorageScopes()
  testConversationIdStorageAndExpectedIdClearGuard()
  testVersionedChatDraftSaveRestoreExpiryAndFutureVersionPreservation()
  testVersionedChatDraftWinsOverRepeatedLegacyRoomRestore()
  testPendingChatRoomMigrationMergesMessagesAndRetainsFailedLocalSend()
  testPendingDraftMigrationDoesNotReplaceNewerConversationDraft()
  testLocalHistoryIsLimitedToCurrentOwnerScope()
  testUnreadChatCountIsLimitedToCurrentOwnerAndIgnoresNonRoomStorage()
  console.log('Customer-service chat owner-scoped session and versioned draft storage checks passed.')
} finally {
  restoreBrowserStorageGlobals()
}
