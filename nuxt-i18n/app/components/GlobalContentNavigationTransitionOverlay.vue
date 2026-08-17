<template>
  <Teleport to="body">
    <Transition
      enter-active-class="transition-opacity duration-200 ease-out"
      leave-active-class="transition-opacity duration-150 ease-in"
      enter-from-class="opacity-0"
      leave-to-class="opacity-0"
    >
      <section
        v-if="open"
        class="global-content-navigation-transition-overlay"
        :aria-label="$t(
          'header.globalNavigationTransition.ariaLabel',
          'Global content navigation',
        )"
      >
        <div
          ref="panelRef"
          class="global-content-navigation-transition-overlay__inner"
          :style="panelStyle"
        >
          <div
            class="global-content-navigation-transition-overlay__options"
            role="group"
            :aria-label="$t(
              'header.globalNavigationTransition.optionsLabel',
              'Choose a destination',
            )"
          >
            <button
              v-for="option in options"
              :key="option.id"
              type="button"
              class="global-content-navigation-transition-overlay__option"
              @click="emit('select', option.id)"
            >
              <span class="global-content-navigation-transition-overlay__icon" aria-hidden="true">
                <Icon :name="option.icon" />
              </span>
              <span class="global-content-navigation-transition-overlay__label">
                {{ $t(option.labelKey, option.fallback) }}
              </span>
            </button>
          </div>
        </div>
      </section>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'

type GlobalContentNavigationOptionId = 'products' | 'faq' | 'pages' | 'blog'

interface GlobalContentNavigationOption {
  id: GlobalContentNavigationOptionId
  icon: string
  labelKey: string
  fallback: string
}

const props = defineProps<{
  open: boolean
  desktopAnchor?: HTMLElement | null
  mobileAnchor?: HTMLElement | null
}>()

const emit = defineEmits<{
  (event: 'select', option: GlobalContentNavigationOptionId): void
}>()

const panelRef = ref<HTMLElement | null>(null)
const panelStyle = ref<Record<string, string>>({
  left: 'auto',
  top: 'calc(var(--site-header-overlay-offset, 0px) + 0.75rem)',
  '--global-navigation-arrow-left': '50%',
})

const options: GlobalContentNavigationOption[] = [
  {
    id: 'products',
    icon: 'lucide:package-search',
    labelKey: 'header.globalNavigationTransition.options.products',
    fallback: 'Products',
  },
  {
    id: 'faq',
    icon: 'lucide:circle-help',
    labelKey: 'header.globalNavigationTransition.options.faq',
    fallback: 'FAQ',
  },
  {
    id: 'pages',
    icon: 'lucide:file-text',
    labelKey: 'header.globalNavigationTransition.options.pages',
    fallback: 'Pages',
  },
  {
    id: 'blog',
    icon: 'lucide:notebook-tabs',
    labelKey: 'header.globalNavigationTransition.options.blog',
    fallback: 'Blog',
  },
]

let positionFrame: number | null = null

const isVisibleAnchor = (element?: HTMLElement | null) => {
  if (!element) return false

  const rect = element.getBoundingClientRect()
  const styles = getComputedStyle(element)
  return (
    rect.width > 0
    && rect.height > 0
    && styles.display !== 'none'
    && styles.visibility !== 'hidden'
  )
}

const getActiveAnchor = (isMobileLayout: boolean) => {
  const preferredAnchor = isMobileLayout
    ? props.mobileAnchor
    : props.desktopAnchor
  const fallbackAnchor = isMobileLayout
    ? props.desktopAnchor
    : props.mobileAnchor

  if (isVisibleAnchor(preferredAnchor)) return preferredAnchor
  if (isVisibleAnchor(fallbackAnchor)) return fallbackAnchor
  return null
}

const updatePanelPosition = () => {
  if (typeof window === 'undefined' || !props.open) return

  const panel = panelRef.value
  const viewportWidth = document.documentElement.clientWidth || window.innerWidth
  const isMobileLayout = window.matchMedia('(max-width: 767px)').matches
  const anchor = getActiveAnchor(isMobileLayout)
  const panelRect = panel?.getBoundingClientRect()
  const panelWidth = panelRect?.width || Math.min(
    384,
    viewportWidth - (isMobileLayout ? 2 : 24),
  )
  const panelHeight = panelRect?.height || 0
  const viewportPadding = 12
  const horizontalPadding = isMobileLayout ? 0 : viewportPadding
  const gap = 12

  if (!anchor) {
    const fallbackTop = Number.parseFloat(
      getComputedStyle(document.documentElement)
        .getPropertyValue('--site-header-overlay-offset'),
    ) || 0
    const fallbackLeft = Math.max(
      horizontalPadding,
      viewportWidth - panelWidth - horizontalPadding,
    )

    panelStyle.value = {
      left: `${fallbackLeft}px`,
      top: `${fallbackTop + gap}px`,
      '--global-navigation-arrow-left': `${Math.min(
        Math.max(panelWidth - 42, 24),
        panelWidth - 24,
      )}px`,
    }
    return
  }

  const anchorRect = anchor.getBoundingClientRect()
  const preferredLeft = isMobileLayout
    ? (viewportWidth - panelWidth) / 2
    : anchorRect.right + gap
  const maxLeft = Math.max(
    horizontalPadding,
    viewportWidth - panelWidth - horizontalPadding,
  )
  const left = Math.min(
    Math.max(preferredLeft, horizontalPadding),
    maxLeft,
  )
  const preferredTop = anchorRect.bottom + gap
  const maxTop = window.innerHeight - panelHeight - viewportPadding
  const top = Math.min(
    Math.max(preferredTop, viewportPadding),
    Math.max(viewportPadding, maxTop),
  )
  const arrowLeft = anchorRect.left + anchorRect.width / 2 - left

  panelStyle.value = {
    left: `${left}px`,
    top: `${top}px`,
    '--global-navigation-arrow-left': `${Math.min(
      Math.max(arrowLeft, 24),
      Math.max(24, panelWidth - 24),
    )}px`,
  }
}

const schedulePanelPositionUpdate = () => {
  if (typeof window === 'undefined') return

  if (positionFrame !== null) {
    window.cancelAnimationFrame(positionFrame)
  }

  positionFrame = window.requestAnimationFrame(() => {
    positionFrame = null
    updatePanelPosition()
  })
}

const handleViewportChange = () => {
  schedulePanelPositionUpdate()
}

watch(
  () => [props.open, props.desktopAnchor, props.mobileAnchor],
  () => {
    schedulePanelPositionUpdate()
  },
  { flush: 'post' },
)

onMounted(() => {
  window.addEventListener('resize', handleViewportChange)
  window.addEventListener('scroll', handleViewportChange, true)
  schedulePanelPositionUpdate()
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleViewportChange)
  window.removeEventListener('scroll', handleViewportChange, true)

  if (positionFrame !== null) {
    window.cancelAnimationFrame(positionFrame)
    positionFrame = null
  }
})
</script>

<style scoped>
.global-content-navigation-transition-overlay {
  position: fixed;
  inset: 0;
  z-index: 1100;
  width: 100%;
  pointer-events: none;
}

.global-content-navigation-transition-overlay__inner {
  --global-navigation-shell-surface: #050505;
  --global-navigation-panel-surface: var(--tz-card-surface, #111116);
  --global-navigation-panel-surface-raised: #17171b;
  --global-navigation-accent-edge: color-mix(
    in srgb,
    var(--tz-brand-primary, #b5ff6d) 74%,
    transparent
  );
  width: min(24rem, calc(100vw - 1.5rem));
  position: absolute;
  padding: 0.75rem;
  border: 1px solid var(--global-navigation-accent-edge);
  border-radius: 0.9rem;
  background:
    linear-gradient(
      180deg,
      #080808 0%,
      var(--global-navigation-shell-surface) 100%
    );
  box-shadow:
    0 20px 54px rgba(0, 0, 0, 0.72),
    0 0 0 4px rgba(181, 255, 109, 0.055),
    inset 0 1px 0 rgba(255, 255, 255, 0.025),
    inset 0 0 0 1px rgba(0, 0, 0, 0.72);
  pointer-events: auto;
}

.global-content-navigation-transition-overlay__inner::after {
  position: absolute;
  top: -0.5rem;
  left: var(--global-navigation-arrow-left, 50%);
  width: 1rem;
  height: 1rem;
  border-top: 1px solid var(--global-navigation-accent-edge);
  border-left: 1px solid var(--global-navigation-accent-edge);
  background: var(--global-navigation-shell-surface);
  box-shadow: -4px -4px 12px rgba(0, 0, 0, 0.12);
  content: "";
  transform: translateX(-50%) rotate(45deg);
}

.global-content-navigation-transition-overlay__options {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 0.5rem;
}

.global-content-navigation-transition-overlay__option {
  display: flex;
  min-width: 0;
  min-height: 4.25rem;
  align-items: center;
  justify-content: flex-start;
  gap: 0.65rem;
  overflow: hidden;
  border: 0;
  border-radius: 0.75rem;
  background:
    linear-gradient(
      180deg,
      var(--global-navigation-panel-surface-raised),
      var(--global-navigation-panel-surface)
    );
  box-shadow:
    inset 0 1px 0 rgba(255, 255, 255, 0.03),
    inset 0 0 0 1px rgba(0, 0, 0, 0.56),
    0 12px 28px rgba(0, 0, 0, 0.26);
  color: #ffffff;
  padding: 0.75rem;
  font-size: 0.82rem;
  font-weight: 800;
  line-height: 1.25;
  text-align: left;
  transition:
    background-color 0.18s ease,
    background 0.18s ease,
    transform 0.18s ease;
}

.global-content-navigation-transition-overlay__option:hover,
.global-content-navigation-transition-overlay__option:focus-visible {
  background:
    linear-gradient(
      180deg,
      #202026,
      var(--global-navigation-panel-surface-raised)
    );
  transform: translateY(-1px);
}

.global-content-navigation-transition-overlay__icon {
  display: grid;
  width: 2rem;
  height: 2rem;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 999px;
  background: #ffffff;
  color: #050505;
  box-shadow: 0 8px 18px rgba(0, 0, 0, 0.28);
}

.global-content-navigation-transition-overlay__icon :deep(svg) {
  width: 1rem;
  height: 1rem;
}

.global-content-navigation-transition-overlay__label {
  min-width: 0;
  overflow-wrap: anywhere;
}

@media (max-width: 767px) {
  .global-content-navigation-transition-overlay__inner {
    width: calc(100% - 2px);
    padding: 0.65rem;
  }

  .global-content-navigation-transition-overlay__options {
    grid-template-columns: minmax(0, 1fr);
    gap: 0.45rem;
  }

  .global-content-navigation-transition-overlay__option {
    gap: 0.25rem;
    min-height: 3.6rem;
    align-items: center;
    padding: 0.6rem;
    font-size: 0.72rem;
  }

  .global-content-navigation-transition-overlay__icon {
    width: 1.75rem;
    height: 1.75rem;
  }

  .global-content-navigation-transition-overlay__icon :deep(svg) {
    width: 0.9rem;
    height: 0.9rem;
  }
}
</style>
