import type { Ref, WritableComputedRef } from 'vue'
import { buildOrderChatMetadata } from '~/composables/chat/useOrderChatPayload'
import { buildProductChatMetadata } from '~/composables/chat/useProductChatPayload'
import { buildProductConfigConfirmMetadata } from '~/composables/chat/useProductConfigConfirmPayload'

type ChatMessageType = 'text' | 'image' | 'product' | 'order' | 'faq' | 'config_confirm'

interface ComposerOptions {
  conversationId: Ref<string>
  selectedAgent: Ref<any>
  messages: WritableComputedRef<any[]>
  user: Ref<any>
  isSending: Ref<boolean>
  currentSenderEmail: () => string
  saveMessagesToStorage: () => void
  scrollToBottom: () => void
  sendMessageToAPI: (messageData: any) => Promise<any>
  replaceLocalMessageWithServerMessage: (localId: number | string, payload: any) => void
  markLocalMessageFailed: (localId: number | string) => void
}

interface ChatMessageDraft {
  message: string
  message_type: ChatMessageType
  metadata?: unknown
  attachment_url?: string
  attachments?: string[]
  type?: string
  title?: string
  url?: string
  thumbnail?: string
}

const createLocalMessageId = () => {
  return `local-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`
}

export const useChatMessageComposer = (options: ComposerOptions) => {
  const sendDraft = async (
    draft: ChatMessageDraft,
    errorLabel: string,
    afterSuccess?: () => void | Promise<void>
  ) => {
    if (!options.selectedAgent.value || options.isSending.value) return false

    options.isSending.value = true
    const localId = createLocalMessageId()
    const messageData = {
      id: localId,
      conversation_id: options.conversationId.value,
      sender_id: options.user.value?.id || 0,
      sender_name: options.user.value?.display_name || options.user.value?.username || '访客',
      sender_email: options.currentSenderEmail(),
      message: draft.message,
      message_type: draft.message_type,
      metadata: draft.metadata || null,
      attachment_url: draft.attachment_url || '',
      attachments: Array.isArray(draft.attachments) ? draft.attachments : (draft.attachment_url ? [draft.attachment_url] : []),
      type: draft.type,
      title: draft.title,
      url: draft.url,
      thumbnail: draft.thumbnail,
      created_at: new Date().toISOString(),
      is_agent: false,
      sync_state: 'sending',
    }

    try {
      options.messages.value.push(messageData)
      options.saveMessagesToStorage()
      options.scrollToBottom()

      const response = await options.sendMessageToAPI(messageData)
      options.replaceLocalMessageWithServerMessage(localId, response)
      await afterSuccess?.()
      return true
    } catch (error) {
      options.markLocalMessageFailed(localId)
      console.error(errorLabel, error)
      return false
    } finally {
      options.isSending.value = false
    }
  }

  const sendTextMessage = (
    message: string,
    afterSuccess?: () => void | Promise<void>
  ) => {
    return sendDraft(
      {
        message,
        message_type: 'text',
      },
      '发送消息失败:',
      afterSuccess
    )
  }

  const sendImageMessage = (
    attachmentUrls: string | string[],
    metadata: Record<string, unknown>,
    afterSuccess?: () => void | Promise<void>
  ) => {
    const attachments = Array.isArray(attachmentUrls)
      ? attachmentUrls.filter(Boolean)
      : [attachmentUrls].filter(Boolean)
    const attachmentUrl = attachments[0] || ''

    return sendDraft(
      {
        message: '[图片]',
        message_type: 'image',
        attachment_url: attachmentUrl,
        attachments,
        metadata,
      },
      '发送图片失败:',
      afterSuccess
    )
  }

  const sendProductMessage = (
    product: Record<string, any>,
    afterSuccess?: () => void | Promise<void>
  ) => {
    const metadata = buildProductChatMetadata(product)
    return sendDraft(
      {
        message: metadata.title,
        message_type: 'product',
        metadata: {
          ...metadata,
          title: metadata.title,
          url: metadata.url,
          thumbnail: metadata.thumbnail,
          price: metadata.price,
        },
        type: 'card',
        title: metadata.title,
        url: metadata.url,
        thumbnail: metadata.thumbnail,
      },
      '分享商品失败:',
      afterSuccess
    )
  }

  const sendProductConfigConfirmMessage = (
    product: Record<string, any>,
    variant: Record<string, any> | null = null,
    afterSuccess?: () => void | Promise<void>
  ) => {
    const metadata = buildProductConfigConfirmMetadata(product, variant)
    const productTitle = metadata.product.title || 'Product'
    return sendDraft(
      {
        message: `Configuration confirmation request: ${productTitle}`,
        message_type: 'config_confirm',
        metadata,
      },
      '发送配置确认失败:',
      afterSuccess
    )
  }

  const sendOrderMessage = (
    order: Record<string, any>,
    afterSuccess?: () => void | Promise<void>
  ) => {
    const metadata = buildOrderChatMetadata(order)
    return sendDraft(
      {
        message: `Order confirmation request: ${metadata.order_number || 'Order'}`,
        message_type: 'order',
        metadata,
      },
      '分享订单失败:',
      afterSuccess
    )
  }

  return {
    sendTextMessage,
    sendImageMessage,
    sendProductMessage,
    sendProductConfigConfirmMessage,
    sendOrderMessage,
  }
}
