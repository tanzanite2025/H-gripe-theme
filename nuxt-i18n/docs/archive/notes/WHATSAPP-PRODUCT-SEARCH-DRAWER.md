# WhatsApp Product Search Drawer Plan

**Last updated:** 2025-11-20

This document describes the implementation plan for a new product search result drawer used together with `WhatsAppChatModal.vue`.

---

## 1. Goal

- When the user is in **Share Products** tab (Technology / Pre-sale / After-sales) and clicks **Search**:
  - A new component (Product Search Drawer) slides up from the **bottom of the page**.
  - The drawer covers the main content area of the WhatsApp chat modal (roughly the red framed area in the design screenshot).
  - The drawer shows **loading first**, then shows product search results.
- The existing `WhatsAppChatModal`:
  - **Stays open** and unchanged in the background (same backdrop, same position).
  - Only the new drawer is stacked above it.
- Search data continues to come from the **Tanzanite Setting** REST API:
  - `GET /wp-json/tanzanite/v1/products?keyword=...&per_page=20&status=publish`.

---

## 2. Current Implementation Summary

File: `app/components/WhatsAppChatModal.vue`

- Per-agent chat state is stored in `chatRooms[agentId]` with:
  - `searchQuery`, `searchResults`, `isSearching`, etc.
- Computed helpers:
  - `searchQuery`, `searchResults`, `isSearching` are computed wrappers on the current agent's `ChatRoomState`.
- Search function (already working):
  - `searchProducts()` calls `$fetch('/wp-json/tanzanite/v1/products', { params: { keyword, per_page, status } })`.
  - Maps `response.items` to `{ id, title, url, thumbnail, price }`.
- UI:
  - **Mobile**: in `activeTab === 'share'` section, shows search input + button + list of `searchResults`.
  - **Desktop**: in `activeTab === 'share'` (desktop area), also shows search input + button + grid of `searchResults`.

The new drawer will **reuse** this API and most of this state, but will move the result rendering into a separate component.

---

## 3. New Drawer Component

Planned file: `app/components/WhatsAppProductSearchDrawer.vue`

### 3.1. Props

- `modelValue: boolean` – controls drawer visibility (for `v-model`).
- `loading: boolean` – bind to `isSearching` from parent.
- `results: any[]` – bind to `searchResults`.
- `error?: string | null` – optional error message.
- `agent?: any | null` – the currently selected agent (Technology / Pre-sale / After-sales).
- `query?: string` – current search keyword (for display only).

### 3.2. Emits

- `update:modelValue(boolean)` – close drawer when user clicks the X button.
- `select(product)` – emitted when user clicks a product in the drawer. Parent will forward to existing `shareProductToChat(product)`.

### 3.3. Layout & Animation

- Use `<Teleport to="body">` so the drawer is fixed to the viewport bottom.
- Use a custom transition (e.g. `whatsapp-product-drawer`) with:
  - `enter-from`: `translateY(100%)`, `opacity: 0`.
  - `enter-to`: `translateY(0)`, `opacity: 1`.
  - `leave-from`: `translateY(0)`, `opacity: 1`.
  - `leave-to`: `translateY(100%)`, `opacity: 0`.
- Wrapper structure:

  ```vue
  <Teleport to="body">
    <Transition name="whatsapp-product-drawer">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-[10001] flex items-end md:items-center justify-center pointer-events-none"
      >
        <div
          class="pointer-events-auto w-full max-w-[1400px] h-[90vh] md:h-[700px] max-h-[85vh]
                 rounded-2xl border-2 border-[#6b73ff] bg-black shadow-[0_0_30px_rgba(107,115,255,0.6)]
                 flex flex-col overflow-hidden"
        >
          <!-- header + content -->
        </div>
      </div>
    </Transition>
  </Teleport>
  ```

- **Header**:
  - Left: title like `Search results` + current agent name + optional keyword.
  - Right: circular X close button styled similar to WhatsAppChatModal close button.

- **Content**:
  - If `loading`: show loading spinner / text.
  - Else if `error`: show error text.
  - Else if `results` empty: show "No products found" / "Search products to share".
  - Else: list/grid of product cards reusing styles from desktop Share Products section.

No additional global background overlay is added; we rely on the existing WhatsAppChatModal backdrop.

---

## 4. New State in WhatsAppChatModal.vue

Add state inside `<script setup lang="ts">`:

- `const productDrawerVisible = ref(false)`
- `const productDrawerError = ref<string | null>(null)`
- `const productDrawerQuery = ref('')`

We will **reuse** existing computed values:

- `searchQuery`
- `searchResults`
- `isSearching`
- `selectedAgent`

No extra per-agent storage is needed beyond what already exists.

---

## 5. Updated searchProducts Flow

Replace current `searchProducts` with the following logical steps:

1. If there is **no selected agent**, early return.
2. If `searchQuery` is empty:
   - Clear `searchResults` and **do not** open the drawer.
   - Return.
3. Before making the API request:
   - `productDrawerQuery.value = searchQuery.value`
   - `productDrawerError.value = null`
   - `productDrawerVisible.value = true` (open drawer first)
4. Perform API request as before:
   - `isSearching.value = true`
   - On success: map `response.items` to the product view model and assign to `searchResults`.
   - On failure: set `productDrawerError.value` and clear `searchResults`.
   - In `finally`: `isSearching.value = false`.

This ensures the drawer **slides up immediately**, then shows loading and results inside.

---

## 6. Drawer Close Behaviour

Implement a handler in `WhatsAppChatModal.vue` (e.g. `handleProductDrawerClose()`):

- `productDrawerVisible.value = false`
- `productDrawerError.value = null`
- `productDrawerQuery.value = ''`
- Clear search state for the current agent:
  - `searchQuery.value = ''`
  - `searchResults.value = []`
  - `isSearching.value = false`

This matches the requirement: **after closing the new component, the Share Products input and history should not retain previous search state**.

---

## 7. Integrating the Drawer Component

At the bottom of `WhatsAppChatModal.vue` template (near other Teleport modals like FAQ), mount the drawer:

```vue
<WhatsAppProductSearchDrawer
  v-model="productDrawerVisible"
  :loading="isSearching"
  :results="searchResults"
  :error="productDrawerError"
  :agent="selectedAgent"
  :query="productDrawerQuery"
  @select="shareProductToChat"
  @update:modelValue="val => { if (!val) handleProductDrawerClose() }"
/>
```

This way:

- When Search is clicked, the drawer opens and uses the same data as current Share Products.
- When a product card is clicked in the drawer, `shareProductToChat(product)` is reused.
- When the drawer is closed (X button), `handleProductDrawerClose` resets all search-related state.

---

## 8. Hiding Old Inline Results

To avoid duplicate result lists:

- In **mobile** Share Products tab, change the product list block so that it only renders when `!productDrawerVisible` (or later, remove it entirely once the drawer is verified):
  - Example: `v-if="searchResults.length > 0 && !productDrawerVisible"`.
- In **desktop** Share Products tab, apply the same condition.
- Because our new flow always opens the drawer when there is a non-empty query, `productDrawerVisible` should be `true` whenever there are active results, so the original inline lists will effectively stay hidden.

`BrowsingHistoryDark` remains in the original position and behaviour, independent of the new drawer.

---

## 9. Later Enhancements (Optional)

- Allow searching again *inside* the drawer (second input box), calling the same `searchProducts` or a lighter wrapper.
- Smarter empty-state copy per agent (e.g. "No products found for Technology agent").
- Dedicated mobile layout for the drawer (full-screen sheet style).
