<template>
  <aside
    v-if="isVisible && !hasActiveOverlay"
    class="embedded-mobile-browser-compatibility-banner"
    data-embedded-browser-notice
    :aria-label="t('common.embeddedBrowserCompatibility.title')"
  >
    <p class="embedded-mobile-browser-compatibility-banner__title">
      {{ t('common.embeddedBrowserCompatibility.title') }}
    </p>
    <p class="embedded-mobile-browser-compatibility-banner__description">
      {{ t('common.embeddedBrowserCompatibility.description') }}
    </p>
    <input
      ref="currentPageLinkInput"
      class="embedded-mobile-browser-compatibility-banner__url"
      :value="securePageUrl"
      :aria-label="t('common.embeddedBrowserCompatibility.linkLabel')"
      dir="ltr"
      readonly
      @click="selectCurrentPageLink"
    />
    <button
      class="embedded-mobile-browser-compatibility-banner__copy"
      type="button"
      :disabled="isCopying"
      @click="copySecurePageUrl"
    >
      {{
        t(
          `common.embeddedBrowserCompatibility.${copied ? 'copied' : 'copyLink'}`,
        )
      }}
    </button>
    <p
      v-if="copyFailed"
      class="embedded-mobile-browser-compatibility-banner__description"
      role="status"
    >
      {{ t('common.embeddedBrowserCompatibility.copyFailed') }}
    </p>
    <button
      class="embedded-mobile-browser-compatibility-banner__dismiss"
      type="button"
      :aria-label="t('common.embeddedBrowserCompatibility.dismiss')"
      @click="dismissBanner"
    >
      ×
    </button>
  </aside>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n, useRouter } from '#imports'
import { useOverlayBackStack } from '~/composables/useOverlayBackStack'
import {
  buildSecureCurrentPageUrl,
  detectEmbeddedMobileBrowser,
} from '~/utils/embeddedMobileBrowserCompatibility'

const embeddedBrowserBannerDismissedStorageKey =
  'embedded-mobile-browser-compatibility-banner-dismissed'

const { t } = useI18n()
const router = useRouter()
const { activeLayers } = useOverlayBackStack()
const hasActiveOverlay = computed(() => activeLayers.value.length > 0)
const isVisible = ref(false)
const copied = ref(false)
const copyFailed = ref(false)
const isCopying = ref(false)
const securePageUrl = ref('')
const currentPageLinkInput = ref<HTMLInputElement | null>(null)
let removeRouteChangeListener: (() => void) | undefined

const readSessionDismissalState = (): boolean => {
  try {
    return (
      window.sessionStorage.getItem(
        embeddedBrowserBannerDismissedStorageKey,
      ) === '1'
    )
  } catch {
    return false
  }
}

const refreshSecureCurrentPageLink = (): void => {
  const latestPageUrl = buildSecureCurrentPageUrl(window.location.href)
  if (latestPageUrl !== securePageUrl.value) {
    securePageUrl.value = latestPageUrl
    copied.value = false
    copyFailed.value = false
  }
}

const selectCurrentPageLink = (): void => {
  currentPageLinkInput.value?.select()
  currentPageLinkInput.value?.setSelectionRange(0, securePageUrl.value.length)
}

const copyPageUrlWithLegacySelection = (pageUrlToCopy: string): boolean => {
  const previouslyFocusedElement = document.activeElement
  const selectionTarget = document.createElement('textarea')
  try {
    selectionTarget.value = pageUrlToCopy
    selectionTarget.setAttribute('readonly', '')
    selectionTarget.style.position = 'fixed'
    selectionTarget.style.opacity = '0'
    document.body.appendChild(selectionTarget)
    selectionTarget.select()
    selectionTarget.setSelectionRange(0, pageUrlToCopy.length)
    return document.execCommand('copy')
  } catch {
    return false
  } finally {
    selectionTarget.remove()
    if (previouslyFocusedElement instanceof HTMLElement)
      previouslyFocusedElement.focus()
  }
}

const copySecurePageUrl = async (): Promise<void> => {
  if (isCopying.value) return
  refreshSecureCurrentPageLink()
  const pageUrlToCopy = securePageUrl.value
  if (!pageUrlToCopy) return
  isCopying.value = true
  copied.value = false
  copyFailed.value = false
  let copySucceeded = false
  try {
    await navigator.clipboard.writeText(pageUrlToCopy)
    copySucceeded = true
  } catch {
    // Embedded browsers can reject clipboard permissions even on HTTPS.
    copySucceeded = copyPageUrlWithLegacySelection(pageUrlToCopy)
  } finally {
    refreshSecureCurrentPageLink()
    if (securePageUrl.value === pageUrlToCopy) {
      copied.value = copySucceeded
      copyFailed.value = !copySucceeded
    }
    isCopying.value = false
  }
}

const dismissBanner = (): void => {
  isVisible.value = false
  try {
    window.sessionStorage.setItem(embeddedBrowserBannerDismissedStorageKey, '1')
  } catch {
    // Dismissal still works for this page if the browser blocks storage.
  }
}

onMounted(() => {
  if (
    !detectEmbeddedMobileBrowser(navigator.userAgent).isEmbeddedMobileBrowser ||
    readSessionDismissalState()
  )
    return
  refreshSecureCurrentPageLink()
  isVisible.value = Boolean(securePageUrl.value)
  removeRouteChangeListener = router.afterEach(refreshSecureCurrentPageLink)
  window.addEventListener('hashchange', refreshSecureCurrentPageLink)
  window.addEventListener('popstate', refreshSecureCurrentPageLink)
  window.addEventListener('pageshow', refreshSecureCurrentPageLink)
})

onBeforeUnmount(() => {
  removeRouteChangeListener?.()
  window.removeEventListener('hashchange', refreshSecureCurrentPageLink)
  window.removeEventListener('popstate', refreshSecureCurrentPageLink)
  window.removeEventListener('pageshow', refreshSecureCurrentPageLink)
})
</script>

<style scoped>
.embedded-mobile-browser-compatibility-banner {
  position: fixed;
  inset-inline: max(
    0.5rem,
    env(safe-area-inset-left),
    env(safe-area-inset-right)
  );
  top: calc(var(--site-header-overlay-offset, 7rem) + 0.5rem);
  z-index: 800;
  box-sizing: border-box;
  width: min(34rem, calc(100% - 1rem));
  max-height: 50vh;
  max-height: 50dvh;
  margin-inline: auto;
  overflow: auto;
  padding: 0.85rem;
  border: 1px solid rgba(5, 150, 105, 0.28);
  border-radius: 0.85rem;
  background: #fff;
  box-shadow: 0 0.5rem 2rem rgba(15, 23, 42, 0.18);
  color: var(--tz-text-primary, #0f172a);
}

.embedded-mobile-browser-compatibility-banner__title {
  margin: 0;
  padding-inline-end: 2.25rem;
  font-size: 0.9rem;
  font-weight: 750;
}

.embedded-mobile-browser-compatibility-banner__description {
  margin: 0.4rem 0;
  color: var(--tz-text-secondary, #475569);
  font-size: 0.8rem;
  line-height: 1.45;
}

.embedded-mobile-browser-compatibility-banner__url {
  display: block;
  box-sizing: border-box;
  width: 100%;
  min-height: 2.75rem;
  padding: 0.5rem;
  border: 1px solid #cbd5e1;
  border-radius: 0.5rem;
  background: #f8fafc;
  color: #334155;
  font-size: 16px;
}

.embedded-mobile-browser-compatibility-banner__copy {
  min-height: 2.75rem;
  margin-top: 0.5rem;
  padding: 0.5rem 0.8rem;
  border: 0;
  border-radius: 0.5rem;
  background: #047857;
  color: #fff;
  cursor: pointer;
  font: inherit;
  font-size: 0.8rem;
}

.embedded-mobile-browser-compatibility-banner__copy:disabled {
  cursor: progress;
  opacity: 0.65;
}

.embedded-mobile-browser-compatibility-banner__dismiss {
  position: absolute;
  top: 0.25rem;
  inset-inline-end: 0.25rem;
  width: 2.75rem;
  height: 2.75rem;
  border: 0;
  background: transparent;
  color: #475569;
  cursor: pointer;
  font-size: 1.5rem;
}

@media (min-width: 768px) {
  .embedded-mobile-browser-compatibility-banner {
    inset-inline-start: auto;
    inset-inline-end: max(1rem, env(safe-area-inset-right));
  }
}
</style>
