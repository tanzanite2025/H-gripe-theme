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
      :hydrate-when="shouldMount"
      v-bind="$attrs"
    />
  </div>
</template>

<script setup lang="ts">
import {
  computed,
  defineAsyncComponent,
  defineComponent,
  h,
  mergeProps,
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

const createSsrDeferredComponent = (loader: DeferredSectionLoader) => {
  const child = defineAsyncComponent({ loader })

  return defineComponent({
    inheritAttrs: false,
    props: {
      hydrateWhen: {
        type: Boolean,
        default: false,
      },
    },
    setup(props, context) {
      const deferredChild = defineAsyncComponent({
        hydrate: (hydrate) => {
          if (props.hydrateWhen) {
            hydrate()
            return
          }

          const stop = watch(
            () => props.hydrateWhen,
            (ready) => {
              if (!ready) return
              stop()
              hydrate()
            },
          )
          return stop
        },
        loader: () => Promise.resolve(child),
      })

      return () => h(
        deferredChild,
        mergeProps(context.attrs),
        context.slots,
      )
    },
  })
}

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
const deferredComponent = createSsrDeferredComponent(props.loader)
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
