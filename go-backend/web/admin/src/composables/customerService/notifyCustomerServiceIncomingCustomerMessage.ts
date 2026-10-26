import { useCustomerServiceNotificationSettings, isCustomerServiceWithinQuietHours } from './useCustomerServiceNotificationSettingsAndQuietHours'
import { playCustomerServiceBrowserMessageNotificationSound, showCustomerServiceBrowserMessageDesktopNotification } from '@/modules/customer-service/customerServiceBrowserMessageNotificationEffects'
import { rememberCustomerServiceMessageNotificationWithinTab, runWithCustomerServiceCrossTabMessageNotificationClaim } from '@/modules/customer-service/customerServiceCrossTabMessageNotificationClaims'

const { soundEnabled, soundVolume, desktopEnabled } = useCustomerServiceNotificationSettings()

export const notifyCustomerServiceIncomingCustomerMessage = async (
  event: Record<string, any>,
  options: { shouldSilenceDesktopNotification?: (event: Record<string, any>) => boolean } = {},
): Promise<void> => {
  if (event.type !== 'conversation.message.created' || event.actor?.kind !== 'customer') return
  const eventId = String(event.event_id || '').trim()
  if (!rememberCustomerServiceMessageNotificationWithinTab(eventId)) return
  await runWithCustomerServiceCrossTabMessageNotificationClaim(eventId, () => {
    if (!isCustomerServiceWithinQuietHours() && soundEnabled.value) playCustomerServiceBrowserMessageNotificationSound(soundVolume.value)
    if (
      desktopEnabled.value
      && !isCustomerServiceWithinQuietHours()
      && !(options.shouldSilenceDesktopNotification?.(event) ?? false)
    ) {
      showCustomerServiceBrowserMessageDesktopNotification(event)
    }
  })
}
