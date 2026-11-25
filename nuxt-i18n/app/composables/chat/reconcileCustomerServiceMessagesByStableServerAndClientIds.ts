export type CustomerServiceChatMessageRecord = Record<string, any>

export const compareCustomerServiceMessageIDs = (left: number | string, right: number | string) => {
  const leftID = String(left)
  const rightID = String(right)
  if (/^\d+$/.test(leftID) && /^\d+$/.test(rightID)) {
    if (leftID.length !== rightID.length) return leftID.length - rightID.length
    return leftID.localeCompare(rightID)
  }
  return leftID.localeCompare(rightID)
}

export const reconcileCustomerServiceMessagesByStableServerAndClientIds = (
  currentMessages: CustomerServiceChatMessageRecord[],
  serverMessages: CustomerServiceChatMessageRecord[],
  normalizeServerMessage: (message: CustomerServiceChatMessageRecord) => CustomerServiceChatMessageRecord | null,
) => {
  const localByID = new Map(currentMessages.map((message) => [String(message.id), message]))
  const localByClientMessageID = new Map(
    currentMessages
      .filter((message) => message?.client_message_id && message?.sync_state !== 'persisted')
      .map((message) => [String(message.client_message_id), message]),
  )
  const reconciledLocalMessageIDs = new Set<string>()
  const normalizedServerMessages = serverMessages
    .map((serverMessage) => {
      const normalized = normalizeServerMessage(serverMessage)
      if (!normalized) return null

      const local = (normalized.client_message_id
        ? localByClientMessageID.get(String(normalized.client_message_id))
        : undefined)
        || localByID.get(String(normalized.id))
      if (!local) return normalized
      reconciledLocalMessageIDs.add(String(local.id))

      return {
        ...normalized,
        metadata: normalized.metadata || local.metadata || null,
        message_type: normalized.message_type || local.message_type || 'text',
        attachment_url: normalized.attachment_url || local.attachment_url || '',
        attachments: normalized.attachments?.length ? normalized.attachments : (local.attachments || []),
        source: normalized.source || local.source || '',
        type: local.type || normalized.type,
        title: local.title || normalized.title,
        url: local.url || normalized.url,
        thumbnail: local.thumbnail || normalized.thumbnail,
      }
    })
    .filter((message): message is CustomerServiceChatMessageRecord => message !== null)

  const uniqueMessagesByServerID = new Map<string, CustomerServiceChatMessageRecord>()
  currentMessages.forEach((message) => {
    if (reconciledLocalMessageIDs.has(String(message.id))) return
    uniqueMessagesByServerID.set(String(message.id), message)
  })
  normalizedServerMessages.forEach((message) => {
    uniqueMessagesByServerID.set(String(message.id), message)
  })

  return Array.from(uniqueMessagesByServerID.values()).sort((left, right) => {
    const timeDifference = new Date(left.created_at || 0).getTime() - new Date(right.created_at || 0).getTime()
    if (timeDifference !== 0) return timeDifference
    return compareCustomerServiceMessageIDs(left.cursor_id || left.id, right.cursor_id || right.id)
  })
}
