import { STOREFRONT_PRIMARY_NAVIGATION_RESOURCE_WARMUP } from './storefrontLoadingPolicy'

// Imports cannot be aborted. Stop waiting after the budget so one stalled
// namespace does not prevent other menus from warming; its original promise
// remains reusable by intent loading and still merges messages if it finishes.
export async function warmStorefrontPrimaryNavigationMessagesWithBoundedNamespaceWait(
  namespaces: readonly string[],
  loadNavigationNamespace: (namespace: string) => Promise<void>,
  shouldContinueWarmup: () => boolean,
): Promise<void> {
  for (const namespace of namespaces) {
    if (!shouldContinueWarmup()) return
    let namespaceWaitTimer: ReturnType<typeof setTimeout> | undefined
    try {
      await Promise.race([
        loadNavigationNamespace(namespace),
        new Promise<void>(resolve => {
          namespaceWaitTimer = setTimeout(resolve, STOREFRONT_PRIMARY_NAVIGATION_RESOURCE_WARMUP.maximumNamespaceWaitMs)
        }),
      ])
    } catch {
      // Missing translations retain the menu's existing fallback and links.
    } finally {
      clearTimeout(namespaceWaitTimer)
    }
  }
}
