import { defineNuxtPlugin, useRouter } from '#imports'
import { installOverlayBackStack } from '~/composables/useOverlayBackStack'

export default defineNuxtPlugin(() => {
  installOverlayBackStack(useRouter())
})
