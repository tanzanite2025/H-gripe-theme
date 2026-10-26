<template>
  <div
    ref="root"
    :id="!preserveServerRenderedSectionBeforeHydration && !hasMountedSectionContent ? sectionAnchorId : undefined"
    class="home-deferred-section"
    :style="placeholderStyle"
    @focusin="mountSection"
    @pointerenter="mountSection"
  >
    <component
      v-if="preserveServerRenderedSectionBeforeHydration || shouldMount"
      :is="deferredComponent"
      v-bind="$attrs"
      @vue:mounted="markDeferredSectionContentAsMounted"
    />
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  defineAsyncComponent,
  onBeforeUnmount,
  onMounted,
  ref,
  type Component,
  watch,
} from 'vue'
import { useNuxtApp } from '#app/nuxt'

defineOptions({
  inheritAttrs: false,
})

type DeferredSectionLoader = () => Promise<{ default: Component }>

const props = withDefaults(defineProps<{
  loader: DeferredSectionLoader
  moduleId: string
  /** Keep in-page destinations available before client navigation mounts the section. */
  sectionAnchorId?: string
  rootMargin?: string
  minHeight?: string
}>(), {
  rootMargin: '0px 0px 96px 0px',
  minHeight: '0px',
})

const root = ref<HTMLElement | null>(null)
const shouldMount = ref(false)
const hasMountedSectionContent = ref(false)
const nuxtApp = useNuxtApp()
const preserveServerRenderedSectionBeforeHydration = import.meta.server || nuxtApp.isHydrating
const deferredSectionContentComponent = defineAsyncComponent(props.loader)
const deferredComponent = defineAsyncComponent({
  // Vue resolves this wrapper before running its hydration strategy. Resolve
  // only a local component here; fetch the section through the inner async
  // component once the strategy activates, as Nuxt's lazy wrappers do.
  loader: () => Promise.resolve(deferredSectionContentComponent),
  hydrate: (hydrate) => {
    if (shouldMount.value) {
      hydrate()
      return
    }

    const stop = watch(
      shouldMount,
      (ready) => {
        if (!ready) return
        stop()
        hydrate()
      },
    )
    return stop
  },
})
let observer: IntersectionObserver | null = null

if (import.meta.server) {
  nuxtApp.hook('app:rendered', ({ ssrContext }) => {
    if (!ssrContext) return
    ssrContext.modules?.delete(props.moduleId)
    ssrContext['~lazyHydratedModules'] ||= new Set()
    ssrContext['~lazyHydratedModules'].add(props.moduleId)
  })
}

const placeholderStyle = computed(() => {
  if (hasMountedSectionContent.value) return undefined
  if (!props.minHeight || props.minHeight === '0px') return undefined
  return { minHeight: props.minHeight }
})

const markDeferredSectionContentAsMounted = () => {
  hasMountedSectionContent.value = true
}

const cleanupObserver = () => {
  observer?.disconnect()
  observer = null
}

const mountSection = () => {
  if (shouldMount.value) return
  shouldMount.value = true
  cleanupObserver()
}

onMounted(() => {
  if (shouldMount.value) return
  if (!root.value || typeof IntersectionObserver === 'undefined') {
    mountSection()
    return
  }

  observer = new IntersectionObserver((entries) => {
    if (entries.some((entry) => entry.isIntersecting || entry.intersectionRatio > 0)) {
      mountSection()
    }
  }, {
    rootMargin: props.rootMargin,
    threshold: 0,
  })
  observer.observe(root.value)
})

onBeforeUnmount(cleanupObserver)
</script>

<style scoped>
.home-deferred-section {
  width: 100%;
  scroll-margin-top: calc(var(--tz-site-header-spacer-height) + 1rem);
}
</style>
