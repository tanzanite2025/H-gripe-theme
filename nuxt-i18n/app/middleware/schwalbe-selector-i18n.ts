import { useNuxtApp } from '#imports'
import { usePageMessages } from '~/composables/usePageMessages'

export default defineNuxtRouteMiddleware(async () => {
  const locale = useNuxtApp().$i18n.locale.value

  await Promise.all([
    usePageMessages('guidesTireguides').loadPageMessages(locale),
    usePageMessages('guidesSchwalbeTireSelector').loadPageMessages(locale),
  ])
})
