export type ReferralInvitationShareResult = 'shared' | 'copied' | 'cancelled' | 'unavailable'

export interface ReferralInvitationSharePlatform {
  share?: (data: { url: string }) => Promise<void>
  copy: (url: string) => Promise<boolean>
}

// The URL comes from the referral dashboard. Do not substitute the current
// page URL: this action invites an account registration, even on a product page.
export async function shareReferralInvitationURL(
  invitationURL: string,
  platform: ReferralInvitationSharePlatform,
): Promise<ReferralInvitationShareResult> {
  try {
    const parsed = new URL(invitationURL)
    if (!['https:', 'http:'].includes(parsed.protocol) || parsed.username || parsed.password) return 'unavailable'
  } catch {
    return 'unavailable'
  }
  if (platform.share) {
    try {
      await platform.share({ url: invitationURL })
      return 'shared'
    } catch (error) {
      if (error && typeof error === 'object' && 'name' in error && error.name === 'AbortError') return 'cancelled'
    }
  }
  try {
    return await platform.copy(invitationURL) ? 'copied' : 'unavailable'
  } catch {
    return 'unavailable'
  }
}

export async function copyReferralInvitationURL(invitationURL: string): Promise<boolean> {
  if (typeof navigator === 'undefined' || typeof document === 'undefined') return false
  try {
    if (navigator.clipboard?.writeText) {
      await navigator.clipboard.writeText(invitationURL)
      return true
    }
  } catch {
    // Clipboard permissions can be denied; try the selection-based fallback.
  }
  const previouslyFocused = document.activeElement
  const textarea = document.createElement('textarea')
  textarea.value = invitationURL
  textarea.readOnly = true
  textarea.style.position = 'fixed'
  textarea.style.opacity = '0'
  try {
    document.body.appendChild(textarea)
    textarea.select()
    return document.execCommand('copy')
  } catch {
    return false
  } finally {
    textarea.remove()
    if (previouslyFocused instanceof HTMLElement) previouslyFocused.focus({ preventScroll: true })
  }
}
