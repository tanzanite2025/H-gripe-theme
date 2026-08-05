# Chat Attachment Hub Design

Last drafted: 2026-07-29

> Goal: the `+` button in the customer chat input opens one attachment/action layer, then routes the customer into the correct flow: phone image, camera photo, order, or product. This must extend the existing structured customer-service message contract instead of adding one-off text links.

## 1. Product Decision

The `+` button is not an image-only upload button anymore. It becomes the chat Attachment Hub.

Top-level actions:

- `Send phone image`
- `Take photo`
- `Send order`
- `Send product`

The panel is only an action router. It should not own product search, order loading, upload persistence, or message normalization. Those responsibilities stay in dedicated composables and picker components.

## 2. Current Architecture Fit

Existing facts already support this direction:

- Durable message source: `tickets` with `category = customer_service` and `ticket_messages`.
- Public send route: `POST /api/v1/customer-service/messages`.
- Structured message columns: `ticket_messages.message_type`, `ticket_messages.metadata`, and `ticket_messages.attachments`.
- Existing message types: `text`, `product`, `order`, `image`, `config_confirm`.
- Existing customer routes:
  - `GET /api/v1/customer-service/products`
  - `GET /api/v1/customer-service/orders`
- Existing frontend builders:
  - `app/composables/chat/useProductConfigConfirmPayload.ts`
  - `app/composables/chat/useOrderChatPayload.ts`

The new hub must route into these same facts. It must not create a second chat table, local-storage-only message payload, or product/order link format that Admin cannot render after reload.

## 3. UX Contract

### 3.1 Entry

- The `+` button lives in `app/components/whatsapp/ChatTab.vue`, beside the text input.
- Click/tap toggles the Attachment Hub layer.
- Selecting one action closes the hub immediately and opens the matching picker/input flow.
- The hub closes on outside click, `Esc`, tab switch, chat close, or successful action selection.

### 3.2 Layout

- Mobile: bottom action sheet inside the chat modal shell, above the input bar.
- Desktop: compact popover anchored above the `+` button.
- Both layouts must render the same four actions in the same order.
- The visual style should follow the current chat theme: black surface, white text, theme green selected/hover state, no blue gradient.
- Each action should use an icon plus a short label. Do not use long explanatory copy inside the hub.

### 3.3 Action Labels

Add i18n keys under `chatModal.attachments`:

```json
{
  "photoLibrary": "Send phone image",
  "camera": "Take photo",
  "order": "Send order",
  "product": "Send product"
}
```

Chinese copy can map to:

```json
{
  "photoLibrary": "发送手机图片",
  "camera": "拍照",
  "order": "发送订单",
  "product": "发送商品"
}
```

## 4. Action Routing

```ts
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
```

Routing table:

| Action | Immediate UI | Data source | Final message |
| --- | --- | --- | --- |
| `image_library` | Hidden file input, no `capture` attribute | Customer-selected local image | `message_type = image` |
| `camera_capture` | Hidden file input with `accept="image/*"` and `capture="environment"` | Customer camera photo when supported | `message_type = image` |
| `order_reference` | Order picker drawer | `GET /api/v1/customer-service/orders` | `message_type = order` |
| `product_reference` | Dedicated customer-service product search modal | `GET /api/v1/customer-service/products` | `message_type = product` |

Important behavior:

- `image_library` and `camera_capture` use separate inputs. A single input with `capture` cannot reliably support both normal gallery selection and camera capture.
- `camera_capture` should gracefully fall back to a normal image picker when a browser ignores `capture`.
- `order_reference` requires login. If the customer is not authenticated, open the existing chat-embedded `AuthModal` in login mode, then continue to the order picker after success.
- `product_reference` does not require login.

## 5. Message Composer Boundary

Add a dedicated composer layer instead of growing `useWhatsAppState.ts` with more ad hoc send functions.

Recommended file:

- `app/composables/chat/useChatMessageComposer.ts`

Responsibilities:

- Build a normalized local optimistic message.
- Persist through `sendMessageToAPI`.
- Replace local optimistic messages with server messages.
- Mark failed messages consistently.
- Call shared payload builders for product, order, and config-confirm messages.

Suggested public API:

```ts
export interface ChatMessageDraft {
  message: string
  message_type: 'text' | 'image' | 'product' | 'order' | 'config_confirm'
  metadata?: unknown
  attachment_url?: string
}

export const useChatMessageComposer = () => ({
  sendTextMessage,
  sendImageMessage,
  sendProductMessage,
  sendProductConfigConfirmMessage,
  sendOrderMessage,
})
```

`useWhatsAppState.ts` should orchestrate UI state and pass the selected entity into the composer. It should not keep separate copies of optimistic-send logic for every new action type.

## 6. Image And Camera Flow

Current risk: the existing image send path reads the file as base64 and places it into `attachment_url`. That works for demos but is not a good long-term storage model.

Final target:

1. Ensure a customer-service conversation exists.
2. Upload the selected image to a managed backend endpoint.
3. Receive a managed URL and optional asset id.
4. Send a persisted chat message with:

```json
{
  "message": "[image]",
  "message_type": "image",
  "attachment_url": "/uploads/customer-service/...",
  "metadata": {
    "kind": "image",
    "source": "library",
    "asset_id": 123,
    "file_name": "photo.jpg",
    "mime_type": "image/jpeg",
    "size": 123456
  }
}
```

For camera capture, only `metadata.source` changes:

```json
{
  "source": "camera"
}
```

Recommended new backend route:

- `POST /api/v1/customer-service/attachments`

Rules:

- Use `OptionalAuthMiddleware` and the signed Public Chat visitor cookie, matching other customer-service routes.
- Require `conversation_id`.
- Check that the conversation belongs to the current customer owner before storing the file.
- Validate file type, size, and dimensions through the existing upload package.
- Store a durable URL, never a base64 data URL.
- Return display-safe metadata only.

## 7. Order Flow

Entry: `+` -> `Send order`.

If not logged in:

1. Close the hub.
2. Open chat-embedded login.
3. After successful login, refresh membership/user state.
4. Load the order picker.

If logged in:

1. Close the hub.
2. Open an order picker drawer.
3. Fetch `GET /api/v1/customer-service/orders`.
4. Customer explicitly clicks a send/confirm button on one order.
5. Composer sends `message_type = order` using `buildOrderChatMetadata(order)`.

Do not make the whole order card a hidden send action. The existing explicit "confirm order with staff" button rule stays.

## 8. Product Flow

Entry: `+` -> `Send product`.

Final picker behavior:

1. Close the hub.
2. Open the dedicated `CustomerServiceProductSearchModal`.
3. Fetch `GET /api/v1/customer-service/products` from this modal, not from the Products tab and not by switching the active tab.
4. Let the customer explicitly choose `Send product`, which sends `message_type = product`.

The `+` product flow must stay independent from the built-in Products tab. The modal is the reusable boundary: future chat entries, order-service tools, or other storefront placements can open the same component directly without dragging along the Products tab's cart, wishlist, history, or configuration-confirm UI. Function and event names should be intentionally explicit, for example `openCustomerServiceProductSearchModal` and `handleSelectCustomerServiceProductFromSearchModal`.

The Products tab can keep its richer commerce workflow and existing product result drawer. If a future feature needs configuration confirmation from the `+` product flow, add it deliberately to the dedicated modal contract instead of silently routing customers into the tab.

Recommended `product` metadata should include more than a URL:

```json
{
  "kind": "product_reference",
  "product_id": 123,
  "variant_id": 456,
  "title": "Product title",
  "slug": "product-slug",
  "sku": "SKU-001",
  "url": "/shop/product-slug",
  "thumbnail": "/uploads/...",
  "price": "$199.00",
  "price_value": 199
}
```

The URL is display support only. Product id, variant id, SKU, and persisted metadata are the operational facts.

## 9. Component Plan

Recommended additions:

- `app/components/whatsapp/ChatAttachmentHub.vue`
  - Pure UI for the four actions.
  - Emits `select(actionId)`.
- `app/composables/chat/useChatAttachmentActions.ts`
  - Builds action availability and labels.
  - Handles auth gating metadata, not message sending.
- `app/composables/chat/useChatMessageComposer.ts`
  - Owns normalized message sending for all message types.
- `app/components/CustomerServiceProductSearchModal.vue`
  - Dedicated reusable modal for `+` -> `Send product`.
  - Owns a simple search input, product result cards, and an explicit send action.
  - Emits the selected product to the chat orchestration layer.
  - Must not switch the chat active tab or import Products-tab-only workflows.
- `app/components/whatsapp/ChatOrderPickerDrawer.vue`
  - Reuse the order card/list display from `OrderTab.vue`.

Recommended refactors:

- Keep `ChatTab.vue` as message list/input plus Attachment Hub entry.
- Keep `UserChatBody.vue` as tab/body coordinator.
- Keep `WhatsAppChatModal.vue` as drawer mounting and high-level modal shell.
- Move repeated optimistic send code out of `useWhatsAppState.ts`.

## 10. Acceptance Criteria

Functional:

- Clicking `+` opens a four-action layer.
- `Send phone image` opens the photo library/file picker.
- `Take photo` prefers camera capture on mobile and falls back cleanly.
- `Send order` opens login first when logged out, otherwise opens the order picker.
- `Send product` opens the dedicated customer-service product search modal while the chat tab stays active.
- Product/order/image messages persist through `ticket_messages` and re-render after reload.
- Admin renders the same persisted product/order/image facts without relying on customer local storage.

Technical:

- No base64 image payload is stored as the final chat attachment model.
- No product/order message is stored as plain text only.
- All final messages use `POST /api/v1/customer-service/messages`.
- Realtime remains HTTP/SSE refresh over persisted facts, not a second message source.
- Product and order pickers are shared between tab entry and `+` entry.

Visual:

- The hub uses black/card surfaces and theme green state.
- No blue gradient is introduced.
- Mobile action sheet does not cover the text input in a way that prevents closing or typing.
- Desktop popover remains visually subordinate to the chat header and message body.

## 11. Implementation Order

1. Add Attachment Hub UI and action ids without changing message persistence.
2. Add `useChatMessageComposer.ts` and migrate text/product/order/config-confirm/image send paths into it.
3. Split image library and camera capture inputs.
4. Add or confirm a customer-service attachment upload route, then replace base64 image messages.
5. Add the dedicated reusable customer-service product search modal for `+` product references.
6. Refactor order picker so the Orders tab and `+` order action share one list/card renderer.
7. Add Playwright coverage for mobile and desktop:
   - open chat;
   - click `+`;
   - verify four actions;
   - route into product picker;
   - route into order login/picker;
   - route into library/camera input where browser support allows.

## 12. Non-Goals

- Do not add staff reply tools to the storefront chat.
- Do not add another chat message table.
- Do not make the `+` menu a marketing or help menu.
- Do not duplicate product/order card renderers across unrelated components.
- Do not make order/product cards auto-send when tapped; explicit send/confirm remains required.
