export type ChatAttachmentActionId =
  | 'image_library'
  | 'camera_capture'
  | 'order_reference'
  | 'product_reference'

export interface ChatAttachmentAction {
  id: ChatAttachmentActionId
  labelKey: string
  icon: string
  requiresAuth: boolean
  mobilePreferred?: boolean
}

export const CHAT_ATTACHMENT_ACTIONS: readonly ChatAttachmentAction[] = [
  {
    id: 'image_library',
    labelKey: 'chatModal.attachments.photoLibrary',
    icon: 'lucide:image',
    requiresAuth: false,
  },
  {
    id: 'camera_capture',
    labelKey: 'chatModal.attachments.camera',
    icon: 'lucide:camera',
    requiresAuth: false,
    mobilePreferred: true,
  },
  {
    id: 'order_reference',
    labelKey: 'chatModal.attachments.order',
    icon: 'lucide:receipt-text',
    requiresAuth: true,
  },
  {
    id: 'product_reference',
    labelKey: 'chatModal.attachments.product',
    icon: 'lucide:shopping-bag',
    requiresAuth: false,
  },
] as const
