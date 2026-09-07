<template>
  <div
    ref="root"
    class="home-deferred-section"
    :style="placeholderStyle"
    @focusin="mountSection"
    @pointerenter="mountSection"
  >
    <component
      :is="deferredComponent"
      v-bind="$attrs"
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
  rootMargin?: string
  minHeight?: string
}>(), {
  rootMargin: '0px 0px 96px 0px',
  minHeight: '0px',
})

const root = ref<HTMLElement | null>(null)
const shouldMount = ref(false)
const deferredComponent = defineAsyncComponent({
  loader: props.loader,
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
  const nuxtApp = useNuxtApp()
  nuxtApp.hook('app:rendered', ({ ssrContext }) => {
    if (!ssrContext) return
    ssrContext.modules?.delete(props.moduleId)
    ssrContext['~lazyHydratedModules'] ||= new Set()
    ssrContext['~lazyHydratedModules'].add(props.moduleId)
  })
}

const placeholderStyle = computed(() => {
  if (shouldMount.value) return undefined
  if (!props.minHeight || props.minHeight === '0px') return undefined
  return { minHeight: props.minHeight }
})

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
