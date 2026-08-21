<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n, useLocalePath } from '#imports'
import { useChatWidget } from '~/composables/useChatWidget'
import { isSimplifiedChineseStorefrontLocale } from '~/utils/storefrontLocales'
import type { ShopProduct } from '~/composables/useShopProducts'

type ProductShareTarget = {
  id: string
  label: string
  mode: 'chat' | 'open' | 'copy'
  url?: string
  iconName?: string
  iconPath?: string
}

const props = defineProps<{
  product: ShopProduct
  anchorEl?: HTMLElement | null
}>()

const emit = defineEmits<{
  close: []
}>()

const socialIconPaths: Record<string, string> = {
  x: 'M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z',
  facebook: 'M22 12c0-5.523-4.477-10-10-10S2 6.477 2 12c0 4.991 3.657 9.128 8.438 9.878v-6.987h-2.54V12h2.54V9.797c0-2.506 1.492-3.89 3.777-3.89 1.094 0 2.238.195 2.238.195v2.463h-1.261c-1.243 0-1.631.771-1.631 1.562V12h2.773l-.443 2.891h-2.33v6.987C18.343 21.128 22 16.991 22 12z',
  instagram: 'M7 2C4.2 2 2 4.2 2 7v10c0 2.8 2.2 5 5 5h10c2.8 0 5-2.2 5-5V7c0-2.8-2.2-5-5-5H7zm10 2c1.7 0 3 1.3 3 3v10c0 1.7-1.3 3-3 3H7c-1.7 0-3-1.3-3-3V7c0-1.7 1.3-3 3-3h10zm-5 3a5 5 0 100 10 5 5 0 000-10zm0 2.2a2.8 2.8 0 110 5.6 2.8 2.8 0 010-5.6zM17.8 6.2a1 1 0 110 2 1 1 0 010-2z',
  wechat: 'M8.5 4C4.91 4 2 6.46 2 9.5c0 1.7.94 3.24 2.45 4.27l-.62 2.23 2.31-1.16c.73.23 1.53.36 2.36.36.2 0 .4-.01.6-.02A5.96 5.96 0 0011 11.5c0-3.04-2.91-5.5-6.5-5.5zm-2.25 4.25a.9.9 0 110 1.8.9.9 0 010-1.8zm4.5 0a.9.9 0 110 1.8.9.9 0 010-1.8zM15.5 8c-3.59 0-6.5 2.24-6.5 5s2.91 5 6.5 5c.83 0 1.63-.13 2.36-.36l2.31 1.16-.62-2.23C21.06 15.54 22 14 22 12.5c0-2.49-2.07-4.38-4.93-4.85A7.85 7.85 0 0015.5 8zm-2.25 4a.9.9 0 110 1.8.9.9 0 010-1.8zm4.5 0a.9.9 0 110 1.8.9.9 0 010-1.8z',
}

const { locale } = useI18n()
const localePath = useLocalePath()
const { openChat } = useChatWidget()
const popoverElement = ref<HTMLElement | null>(null)
const copiedTargetId = ref('')
const popoverPosition = ref({ top: -9999, left: -9999 })
const popoverArrowLeft = ref(0)
const popoverPlacement = ref<'bottom' | 'top'>('bottom')
let copiedStateTimer: number | null = null

const productUrl = computed(() => {
  const rawUrl = String(props.product.url || (props.product.slug ? `/shop/${props.product.slug}` : '')).trim()
  if (!rawUrl) return ''
  if (/^https?:\/\//i.test(rawUrl)) return rawUrl
  if (/^\/[a-z]{2}(?:[_-][a-z]{2})?\/shop\/[^/?#]+(?:[?#].*)?$/i.test(rawUrl)) return import.meta.client
    ? new URL(rawUrl, window.location.origin).toString()
    : rawUrl
  const localizedPath = localePath(rawUrl)
  if (!import.meta.client) return localizedPath
  return new URL(localizedPath, window.location.origin).toString()
})

const encodedProductUrl = computed(() => encodeURIComponent(productUrl.value))
const encodedProductTitle = computed(() => encodeURIComponent(props.product.title || ''))
const isChineseLocale = computed(() => isSimplifiedChineseStorefrontLocale(locale.value))
const shareLabel = computed(() => (isChineseLocale.value ? '分享商品' : 'Share product'))
const shareToChatLabel = computed(() => (isChineseLocale.value ? '发送到聊天' : 'Send to chat'))
const copiedMessage = computed(() => (isChineseLocale.value ? '链接已复制' : 'Link copied'))

const shareTargets = computed<ProductShareTarget[]>(() => [
  {
    id: 'chat',
    label: shareToChatLabel.value,
    mode: 'chat',
    iconName: 'lucide:message-square-share',
  },
  {
    id: 'x',
    label: 'X',
    mode: 'open',
    iconPath: socialIconPaths.x,
    url: `https://twitter.com/intent/tweet?url=${encodedProductUrl.value}&text=${encodedProductTitle.value}`,
  },
  {
    id: 'facebook',
    label: 'Facebook',
    mode: 'open',
    iconPath: socialIconPaths.facebook,
    url: `https://www.facebook.com/sharer/sharer.php?u=${encodedProductUrl.value}`,
  },
  {
    id: 'instagram',
    label: 'Instagram',
    mode: 'copy',
    iconPath: socialIconPaths.instagram,
  },
  {
    id: 'rabbit',
    label: 'Rabbit',
    mode: 'copy',
    iconName: 'lucide:rabbit',
  },
  {
    id: 'wechat',
    label: '微信',
    mode: 'copy',
    iconPath: socialIconPaths.wechat,
  },
])

const copyProductUrl = async (targetId: string) => {
  if (!import.meta.client || !productUrl.value) return

  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(productUrl.value)
    } else {
      const textArea = document.createElement('textarea')
      textArea.value = productUrl.value
      textArea.setAttribute('readonly', 'true')
      textArea.style.position = 'fixed'
      textArea.style.opacity = '0'
      document.body.appendChild(textArea)
      textArea.select()
      document.execCommand('copy')
      document.body.removeChild(textArea)
    }
    copiedTargetId.value = targetId
    if (copiedStateTimer) {
      window.clearTimeout(copiedStateTimer)
    }
    copiedStateTimer = window.setTimeout(() => {
      copiedTargetId.value = ''
      copiedStateTimer = null
    }, 1600)
  } catch {
    copiedTargetId.value = ''
  }
}

const handleShareTarget = async (target: ProductShareTarget) => {
  if (target.mode === 'chat') {
    openChat({
      showAgentList: false,
      pendingProductReference: {
        ...props.product,
        name: props.product.title,
        price: props.product.priceLabel,
        priceValue: props.product.priceNumber,
        url: productUrl.value,
      },
    })
    emit('close')
    return
  }

  if (target.mode === 'open' && target.url && import.meta.client) {
    window.open(target.url, '_blank', 'noopener,noreferrer')
    return
  }

  await copyProductUrl(target.id)
}

const handleDocumentClick = (event: MouseEvent) => {
  const popover = popoverElement.value
  if (!popover || !(event.target instanceof Node) || popover.contains(event.target)) return
  emit('close')
}

const updatePopoverPosition = () => {
  if (!import.meta.client || !props.anchorEl) return
  const anchorRect = props.anchorEl.getBoundingClientRect()
  const viewportWidth = window.innerWidth
  const viewportHeight = window.innerHeight
  const popoverWidth = popoverElement.value?.offsetWidth || 0
  const popoverHeight = popoverElement.value?.offsetHeight || 0
  const mobileCentered = viewportWidth <= 767
  const viewportPadding = 8
  const anchorGap = 10
  const dockRect = document.querySelector('.dock-bar')?.getBoundingClientRect()
  const dockTop = dockRect && dockRect.height > 0 ? dockRect.top : viewportHeight
  const bottomBoundary = Math.max(viewportPadding, Math.min(viewportHeight - viewportPadding, dockTop - viewportPadding))
  const topPosition = anchorRect.top - popoverHeight - anchorGap
  const bottomPosition = anchorRect.bottom + anchorGap
  const shouldPlaceAbove = popoverHeight > 0 && bottomPosition + popoverHeight > bottomBoundary && topPosition >= viewportPadding
  const nextTop = shouldPlaceAbove
    ? topPosition
    : Math.min(bottomPosition, bottomBoundary - popoverHeight)
  const minLeft = popoverWidth ? popoverWidth / 2 + 8 : 8
  const maxLeft = popoverWidth ? viewportWidth - popoverWidth / 2 - 8 : viewportWidth - 8
  const anchorLeft = anchorRect.left + anchorRect.width / 2
  const nextLeft = mobileCentered ? viewportWidth / 2 : Math.min(Math.max(anchorLeft, minLeft), maxLeft)

  popoverPosition.value = {
    top: Math.max(viewportPadding, nextTop),
    left: nextLeft,
  }
  popoverPlacement.value = shouldPlaceAbove ? 'top' : 'bottom'
  if (popoverWidth > 0) {
    popoverArrowLeft.value = Math.min(Math.max(anchorLeft - nextLeft + popoverWidth / 2, 18), popoverWidth - 18)
  }
}

const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape') {
    emit('close')
  }
}

onMounted(() => {
  document.addEventListener('click', handleDocumentClick)
  document.addEventListener('keydown', handleKeydown)
  window.addEventListener('resize', updatePopoverPosition)
  window.addEventListener('scroll', updatePopoverPosition, true)
  void nextTick(updatePopoverPosition)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', handleDocumentClick)
  document.removeEventListener('keydown', handleKeydown)
  window.removeEventListener('resize', updatePopoverPosition)
  window.removeEventListener('scroll', updatePopoverPosition, true)
  if (copiedStateTimer) {
    window.clearTimeout(copiedStateTimer)
  }
})

watch(
  () => props.anchorEl,
  () => {
    void nextTick(updatePopoverPosition)
  }
)
</script>

<template>
  <Teleport to="body">
    <section
      ref="popoverElement"
      class="product-share-popover"
      :class="{ 'product-share-popover--above': popoverPlacement === 'top' }"
      role="group"
      :aria-label="shareLabel"
      :style="{
        top: `${popoverPosition.top}px`,
        left: `${popoverPosition.left}px`,
        '--share-popover-arrow-left': `${popoverArrowLeft}px`,
      }"
      @click.stop
    >
      <button
        v-for="target in shareTargets"
        :key="target.id"
        type="button"
        class="product-share-popover__target"
        :class="[
          `product-share-popover__target--${target.id}`,
          { 'product-share-popover__target--copied': copiedTargetId === target.id },
        ]"
        :aria-label="`${shareLabel}: ${target.label}`"
        :title="`${shareLabel}: ${target.label}`"
        @click="handleShareTarget(target)"
      >
        <Icon
          v-if="target.iconName"
          :name="target.iconName"
          class="product-share-popover__target-icon"
          aria-hidden="true"
        />
        <svg
          v-else
          class="product-share-popover__target-icon"
          viewBox="0 0 24 24"
          fill="currentColor"
          focusable="false"
          aria-hidden="true"
        >
          <path :d="target.iconPath" />
        </svg>
      </button>
      <span v-if="copiedTargetId" class="product-share-popover__status" role="status">
        {{ copiedMessage }}
      </span>
    </section>
  </Teleport>
</template>

<style scoped>
.product-share-popover {
  position: fixed;
  z-index: 100;
  display: inline-flex;
  width: max-content;
  max-width: min(16rem, calc(100vw - 1rem));
  align-items: center;
  justify-content: center;
  flex-wrap: nowrap;
  gap: 0.28rem;
  padding: 0.34rem;
  border: 1px solid var(--tz-brand-primary, #b5ff6d);
  border-radius: 999px;
  color: rgba(248, 250, 252, 0.86);
  background: var(--tz-card-surface, #111116);
  box-shadow:
    0 8px 22px rgba(0, 0, 0, 0.36),
    0 0 0 1px rgba(181, 255, 109, 0.16);
  transform: translateX(-50%);
}

.product-share-popover--above {
  align-items: flex-start;
}

.product-share-popover::before,
.product-share-popover::after {
  content: '';
  position: absolute;
  left: var(--share-popover-arrow-left, 50%);
  width: 0;
  height: 0;
  transform: translateX(-50%);
}

.product-share-popover::before {
  top: -0.55rem;
  border-right: 0.5rem solid transparent;
  border-bottom: 0.5rem solid var(--tz-brand-primary, #b5ff6d);
  border-left: 0.5rem solid transparent;
}

.product-share-popover::after {
  top: -0.44rem;
  border-right: 0.42rem solid transparent;
  border-bottom: 0.42rem solid var(--tz-card-surface, #111116);
  border-left: 0.42rem solid transparent;
}

.product-share-popover--above::before {
  top: auto;
  bottom: -0.55rem;
  border-top: 0.5rem solid var(--tz-brand-primary, #b5ff6d);
  border-bottom: 0;
}

.product-share-popover--above::after {
  top: auto;
  bottom: -0.44rem;
  border-top: 0.42rem solid var(--tz-card-surface, #111116);
  border-bottom: 0;
}

.product-share-popover__target {
  display: grid;
  width: 2.08rem;
  height: 2.08rem;
  flex: 0 0 auto;
  place-items: center;
  border: 0;
  border-radius: 999px;
  color: rgba(255, 255, 255, 0.78);
  background: transparent;
  transition:
    background-color 160ms ease,
    color 160ms ease,
    transform 160ms ease;
}

.product-share-popover__target:hover,
.product-share-popover__target:focus-visible,
.product-share-popover__target--copied {
  color: #0b0b0e;
  background: #b5ff6d;
  transform: translateY(-1px);
}

.product-share-popover__target:focus-visible {
  outline: 2px solid rgba(181, 255, 109, 0.72);
  outline-offset: 2px;
}

.product-share-popover__target-icon {
  display: block;
  width: 1.12rem;
  height: 1.12rem;
}

.product-share-popover__target--facebook .product-share-popover__target-icon,
.product-share-popover__target--instagram .product-share-popover__target-icon,
.product-share-popover__target--wechat .product-share-popover__target-icon {
  width: 1.2rem;
  height: 1.2rem;
}

.product-share-popover__status {
  position: absolute;
  width: 1px;
  height: 1px;
  overflow: hidden;
  clip: rect(0 0 0 0);
  clip-path: inset(50%);
  white-space: nowrap;
}
</style>
