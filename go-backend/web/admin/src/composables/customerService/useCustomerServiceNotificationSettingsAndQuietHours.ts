import { ref } from 'vue'
import { playCustomerServiceBrowserMessageNotificationSound } from '@/modules/customer-service/customerServiceBrowserMessageNotificationEffects'

let notificationPermissionRequestStarted = false
const notificationSoundEnabled = ref(true)
const notificationSoundVolume = ref(0.12)
const desktopNotificationEnabled = ref(true)
const suppressDesktopNotificationWhenFocused = ref(true)
const quietHoursEnabled = ref(false)
const quietHoursStart = ref('22:00')
const quietHoursEnd = ref('08:00')
const notificationPermission = ref<NotificationPermission | null>(null)
const notificationSettingsStorageKey = 'customer-service-notification-settings'

const loadCustomerServiceNotificationPreferencesFromStorage = (): void => {
  if (typeof window === 'undefined') return
  try {
    const stored = JSON.parse(window.localStorage.getItem(notificationSettingsStorageKey) || '{}') as Record<string, unknown>
    if (typeof stored.soundEnabled === 'boolean') notificationSoundEnabled.value = stored.soundEnabled
    if (typeof stored.soundVolume === 'number' && Number.isFinite(stored.soundVolume)) {
      notificationSoundVolume.value = Math.min(1, Math.max(0, stored.soundVolume))
    }
    if (typeof stored.desktopEnabled === 'boolean') desktopNotificationEnabled.value = stored.desktopEnabled
    if (typeof stored.suppressDesktopWhenFocused === 'boolean') {
      suppressDesktopNotificationWhenFocused.value = stored.suppressDesktopWhenFocused
    }
    if (typeof stored.quietHoursEnabled === 'boolean') quietHoursEnabled.value = stored.quietHoursEnabled
    if (typeof stored.quietHoursStart === 'string' && /^\d{2}:\d{2}$/.test(stored.quietHoursStart)) quietHoursStart.value = stored.quietHoursStart
    if (typeof stored.quietHoursEnd === 'string' && /^\d{2}:\d{2}$/.test(stored.quietHoursEnd)) quietHoursEnd.value = stored.quietHoursEnd
  } catch {
    // Ignore malformed local preferences and use defaults.
  }
  if ('Notification' in window) notificationPermission.value = Notification.permission
}

const persistCustomerServiceNotificationPreferencesToStorage = (): void => {
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(notificationSettingsStorageKey, JSON.stringify({
      soundEnabled: notificationSoundEnabled.value,
      soundVolume: notificationSoundVolume.value,
      desktopEnabled: desktopNotificationEnabled.value,
      suppressDesktopWhenFocused: suppressDesktopNotificationWhenFocused.value,
      quietHoursEnabled: quietHoursEnabled.value,
      quietHoursStart: quietHoursStart.value,
      quietHoursEnd: quietHoursEnd.value,
    }))
  } catch {
    // Session preferences still apply when browser storage is unavailable.
  }
}

loadCustomerServiceNotificationPreferencesFromStorage()

export const isCustomerServiceWithinQuietHours = (
  now: Date = new Date(),
  enabled = quietHoursEnabled.value,
  startValue = quietHoursStart.value,
  endValue = quietHoursEnd.value,
): boolean => {
  if (!enabled) return false
  if (!/^\d{2}:\d{2}$/.test(startValue) || !/^\d{2}:\d{2}$/.test(endValue)) return false
  const [startHour, startMinute] = startValue.split(':').map(Number)
  const [endHour, endMinute] = endValue.split(':').map(Number)
  if (startHour > 23 || endHour > 23 || startMinute > 59 || endMinute > 59) return false
  const start = startHour * 60 + startMinute
  const end = endHour * 60 + endMinute
  const current = now.getHours() * 60 + now.getMinutes()
  if (start === end) return true
  return start < end ? current >= start && current < end : current >= start || current < end
}

export const requestCustomerServiceNotificationPermission = async (): Promise<NotificationPermission | null> => {
  if (typeof window === 'undefined' || !('Notification' in window)) return null
  notificationPermission.value = Notification.permission
  if (Notification.permission !== 'default' || notificationPermissionRequestStarted) return notificationPermission.value
  notificationPermissionRequestStarted = true
  try {
    notificationPermission.value = await Notification.requestPermission()
    return notificationPermission.value
  } catch {
    notificationPermission.value = Notification.permission
    return notificationPermission.value
  }
}

export const useCustomerServiceNotificationSettings = () => {
  const setSoundEnabled = (enabled: boolean): void => {
    notificationSoundEnabled.value = enabled
    persistCustomerServiceNotificationPreferencesToStorage()
  }

  const setSoundVolume = (volume: number): void => {
    if (!Number.isFinite(volume)) return
    notificationSoundVolume.value = Math.min(1, Math.max(0, volume))
    persistCustomerServiceNotificationPreferencesToStorage()
  }

  const setDesktopEnabled = (enabled: boolean): void => {
    desktopNotificationEnabled.value = enabled
    persistCustomerServiceNotificationPreferencesToStorage()
  }

  const setSuppressDesktopWhenFocused = (enabled: boolean): void => {
    suppressDesktopNotificationWhenFocused.value = enabled
    persistCustomerServiceNotificationPreferencesToStorage()
  }

  const setQuietHoursEnabled = (enabled: boolean): void => {
    quietHoursEnabled.value = enabled
    persistCustomerServiceNotificationPreferencesToStorage()
  }

  const setQuietHoursStart = (value: string): void => {
    if (!/^\d{2}:\d{2}$/.test(value)) return
    quietHoursStart.value = value
    persistCustomerServiceNotificationPreferencesToStorage()
  }

  const setQuietHoursEnd = (value: string): void => {
    if (!/^\d{2}:\d{2}$/.test(value)) return
    quietHoursEnd.value = value
    persistCustomerServiceNotificationPreferencesToStorage()
  }

  const testNotificationSound = (): void => {
    playCustomerServiceBrowserMessageNotificationSound(notificationSoundVolume.value)
  }

  return {
    soundEnabled: notificationSoundEnabled,
    soundVolume: notificationSoundVolume,
    desktopEnabled: desktopNotificationEnabled,
    suppressDesktopWhenFocused: suppressDesktopNotificationWhenFocused,
    quietHoursEnabled,
    quietHoursStart,
    quietHoursEnd,
    notificationPermission,
    setSoundEnabled,
    setSoundVolume,
    setDesktopEnabled,
    setSuppressDesktopWhenFocused,
    setQuietHoursEnabled,
    setQuietHoursStart,
    setQuietHoursEnd,
    testNotificationSound,
    requestPermission: requestCustomerServiceNotificationPermission,
  }
}
