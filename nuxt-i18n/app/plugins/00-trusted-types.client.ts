import { initializeTrustedTypes } from '~/utils/security/trustedScriptUrl'

export default defineNuxtPlugin(() => {
  initializeTrustedTypes()
})
