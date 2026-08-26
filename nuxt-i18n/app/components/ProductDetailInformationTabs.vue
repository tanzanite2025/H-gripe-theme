<template>
  <section
    class="product-information"
    :aria-label="t('productInformationTabs.sectionLabel')"
    :data-hydrated="isHydrated"
  >
    <div
      class="product-information__tabs"
      role="tablist"
      aria-orientation="horizontal"
      :aria-label="t('productInformationTabs.tabListLabel')"
    >
      <button
        v-for="(tab, index) in tabs"
        :id="tabId(tab.id)"
        :key="tab.id"
        type="button"
        role="tab"
        class="product-information__tab"
        :class="{ 'product-information__tab--active': activeTab === tab.id }"
        :aria-selected="activeTab === tab.id"
        :aria-controls="panelId(tab.id)"
        :tabindex="activeTab === tab.id ? 0 : -1"
        @click="activeTab = tab.id"
        @keydown="handleTabKeydown($event, index)"
      >
        {{ tab.label }}
      </button>
    </div>

    <div
      v-for="tab in tabs"
      v-show="activeTab === tab.id"
      :id="panelId(tab.id)"
      :key="`${tab.id}-panel`"
      class="product-information__panel"
      role="tabpanel"
      :aria-labelledby="tabId(tab.id)"
      tabindex="0"
    >
      <SafeRichText
        v-if="contentByTab[tab.id]"
        as="article"
        class="product-information__content"
        :html="contentByTab[tab.id]"
      />
      <p v-else class="product-information__empty">
        {{ tab.emptyMessage }}
      </p>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'

const { t } = useI18n()

type ProductDetailInformationTab = 'details' | 'after-sales' | 'packaging' | 'product-shipping-info'

interface Props {
  detailsHtml?: string | null
  afterSalesHtml?: string | null
  packagingHtml?: string | null
  productShippingInfoHtml?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  detailsHtml: '',
  afterSalesHtml: '',
  packagingHtml: '',
  productShippingInfoHtml: '',
})

const tabs = computed<ReadonlyArray<{
  id: ProductDetailInformationTab
  label: string
  emptyMessage: string
}>>(() => [
  {
    id: 'details',
    label: t('productInformationTabs.tabs.details'),
    emptyMessage: t('productInformationTabs.empty.details'),
  },
  {
    id: 'after-sales',
    label: t('productInformationTabs.tabs.afterSales'),
    emptyMessage: t('productInformationTabs.empty.afterSales'),
  },
  {
    id: 'packaging',
    label: t('productInformationTabs.tabs.packaging'),
    emptyMessage: t('productInformationTabs.empty.packaging'),
  },
  {
    id: 'product-shipping-info',
    label: t('productInformationTabs.tabs.productShippingInfo'),
    emptyMessage: t('productInformationTabs.empty.productShippingInfo'),
  },
])

const activeTab = ref<ProductDetailInformationTab>('details')
const isHydrated = ref(false)

onMounted(() => {
  isHydrated.value = true
})

const contentByTab = computed<Record<ProductDetailInformationTab, string>>(() => ({
  details: props.detailsHtml?.trim() || '',
  'after-sales': props.afterSalesHtml?.trim() || '',
  packaging: props.packagingHtml?.trim() || '',
  'product-shipping-info': props.productShippingInfoHtml?.trim() || '',
}))

const tabId = (id: ProductDetailInformationTab) => `product-information-tab-${id}`
const panelId = (id: ProductDetailInformationTab) => `product-information-panel-${id}`

const activateAndFocusTab = (index: number) => {
  const tab = tabs.value[index]
  if (!tab) return

  activeTab.value = tab.id
  nextTick(() => {
    const tabElement = document.getElementById(tabId(tab.id))
    tabElement?.focus()
    tabElement?.scrollIntoView({ block: 'nearest', inline: 'nearest' })
  })
}

const handleTabKeydown = (event: KeyboardEvent, currentIndex: number) => {
  let nextIndex: number | null = null
  const direction = document.documentElement.dir === 'rtl' ? -1 : 1

  if (event.key === 'ArrowRight') {
    nextIndex = (currentIndex + direction + tabs.value.length) % tabs.value.length
  } else if (event.key === 'ArrowLeft') {
    nextIndex = (currentIndex - direction + tabs.value.length) % tabs.value.length
  } else if (event.key === 'Home') {
    nextIndex = 0
  } else if (event.key === 'End') {
    nextIndex = tabs.value.length - 1
  }

  if (nextIndex === null) return

  event.preventDefault()
  activateAndFocusTab(nextIndex)
}
</script>

<style scoped>
.product-information {
  --product-information-accent: #059669;
  --product-information-accent-soft: rgba(5, 150, 105, 0.72);
  color: var(--tz-text-primary);
}

.product-information__tabs {
  display: flex;
  justify-content: safe center;
  overflow-x: auto;
  border-bottom: 1px solid rgba(148, 163, 184, 0.28);
  scrollbar-width: thin;
  scrollbar-color: rgba(148, 163, 184, 0.35) transparent;
}

.product-information__tab {
  position: relative;
  flex: 0 0 auto;
  min-height: 3.25rem;
  border: 0;
  border-bottom: 2px solid transparent;
  background: transparent;
  color: var(--tz-text-secondary);
  cursor: pointer;
  font: inherit;
  font-size: 0.95rem;
  font-weight: 700;
  padding: 0.85rem 1.25rem;
  white-space: nowrap;
}

.product-information__tab:hover {
  color: var(--tz-text-primary);
}

.product-information__tab--active {
  border-bottom-color: var(--product-information-accent);
  color: var(--product-information-accent);
}

.product-information__tab:focus-visible {
  outline: 2px solid var(--product-information-accent);
  outline-offset: -3px;
}

.product-information__panel:focus-visible {
  outline: 2px solid var(--product-information-accent);
  outline-offset: 3px;
}

.product-information__panel {
  min-height: 9rem;
  padding: 1.5rem 0 0.5rem;
}

.product-information__content {
  color: var(--tz-text-secondary);
  line-height: 1.75;
}

.product-information__content :deep(p) {
  margin: 0 0 1rem;
}

.product-information__content :deep(ul),
.product-information__content :deep(ol) {
  margin: 0 0 1rem;
  padding-left: 1.35rem;
  list-style-position: outside;
}

.product-information__content :deep(ul) {
  list-style-type: disc;
}

.product-information__content :deep(ol) {
  list-style-type: decimal;
}

.product-information__content :deep(li) {
  margin: 0.35rem 0;
}

.product-information__content :deep(a) {
  color: var(--product-information-accent);
  text-decoration: underline;
  text-underline-offset: 0.2em;
}

.product-information__content :deep(h2),
.product-information__content :deep(h3),
.product-information__content :deep(h4),
.product-information__content :deep(strong) {
  color: var(--tz-text-primary);
}

.product-information__content :deep(blockquote) {
  margin: 1.25rem 0;
  border-left: 3px solid var(--product-information-accent-soft);
  color: var(--tz-text-primary);
  padding-left: 1rem;
}

.product-information__content :deep(figure) {
  margin: 1.4rem 0;
}

.product-information__content :deep(img),
.product-information__content :deep(video) {
  display: block;
  max-width: 100%;
  height: auto;
  border-radius: 0.75rem;
  background: var(--tz-image-loading-surface);
}

.product-information__content :deep(video) {
  width: 100%;
}

.product-information__content :deep(figcaption) {
  margin-top: 0.55rem;
  color: var(--tz-text-muted);
  font-size: 0.9rem;
}

.product-information__content :deep(pre),
.product-information__content :deep(table) {
  display: block;
  max-width: 100%;
  overflow-x: auto;
}

.product-information__content :deep(p),
.product-information__content :deep(li),
.product-information__content :deep(a) {
  overflow-wrap: anywhere;
}

.product-information__empty {
  display: flex;
  min-height: 7rem;
  align-items: center;
  margin: 0;
  border-block: 1px solid rgba(148, 163, 184, 0.16);
  background: var(--tz-surface-muted);
  color: var(--tz-text-secondary);
  padding: 1.25rem;
}

@media (max-width: 767px) {
  .product-information__tab {
    min-height: 3rem;
    padding-inline: 1rem;
  }

  .product-information__panel {
    min-height: 8rem;
  }
}
</style>
