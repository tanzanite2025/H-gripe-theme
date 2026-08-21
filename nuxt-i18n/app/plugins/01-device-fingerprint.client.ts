import { defineNuxtPlugin } from '#imports'
import { attachDeviceFingerprintHeader, resolveDeviceFingerprint } from '~/utils/deviceFingerprint'

export default defineNuxtPlugin(() => {
  void resolveDeviceFingerprint()

  const currentFetch = (globalThis as { $fetch?: any }).$fetch
  if (!currentFetch?.create) {
    return
  }

  ;(globalThis as { $fetch?: any }).$fetch = currentFetch.create({
    async onRequest({ options }: { options: { headers?: HeadersInit } }) {
      options.headers = await attachDeviceFingerprintHeader(options.headers)
    },
  })
})
