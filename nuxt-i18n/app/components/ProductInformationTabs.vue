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
      <article
        v-if="contentByTab[tab.id]"
        class="product-information__content"
        v-html="contentByTab[tab.id]"
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

type ProductInformationTab = 'details' | 'after-sales' | 'packaging' | 'shipping'

interface Props {
  detailsHtml?: string | null
  afterSalesHtml?: string | null
  packagingHtml?: string | null
  shippingHtml?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  detailsHtml: '',
  afterSalesHtml: '',
  packagingHtml: '',
  shippingHtml: '',
})

const tabs = computed<ReadonlyArray<{
  id: ProductInformationTab
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
    id: 'shipping',
    label: t('productInformationTabs.tabs.shipping'),
    emptyMessage: t('productInformationTabs.empty.shipping'),
  },
])

const activeTab = ref<ProductInformationTab>('details')
const isHydrated = ref(false)

onMounted(() => {
  isHydrated.value = true
})

const contentByTab = computed<Record<ProductInformationTab, string>>(() => ({
  details: props.detailsHtml?.trim() || '',
  'after-sales': props.afterSalesHtml?.trim() || '',
  packaging: props.packagingHtml?.trim() || '',
  shipping: props.shippingHtml?.trim() || '',
}))

const tabId = (id: ProductInformationTab) => `product-information-tab-${id}`
const panelId = (id: ProductInformationTab) => `product-information-panel-${id}`

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
  color: var(--tz-text-primary);
}

.product-information__tabs {
  display: flex;
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
  border-bottom-color: #38bdf8;
  color: #f8fafc;
}

.product-information__tab:focus-visible {
  outline: 2px solid #38bdf8;
  outline-offset: -3px;
}

.product-information__panel:focus-visible {
  outline: 2px solid #38bdf8;
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

.product-information__content :deep(a) {
  color: #7dd3fc;
  text-decoration: underline;
  text-underline-offset: 0.2em;
}

.product-information__content :deep(h2),
.product-information__content :deep(h3),
.product-information__content :deep(h4),
.product-information__content :deep(strong) {
  color: #f8fafc;
}

.product-information__content :deep(img),
.product-information__content :deep(video),
.product-information__content :deep(iframe) {
  max-width: 100%;
  height: auto;
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
  background: rgba(15, 23, 42, 0.5);
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
