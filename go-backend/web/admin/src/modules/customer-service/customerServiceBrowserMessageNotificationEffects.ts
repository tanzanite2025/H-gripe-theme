let notificationAudioContext: AudioContext | null = null

const getCustomerServiceNotificationAudioContext = (): AudioContext | null => {
  if (typeof window === 'undefined') return null
  const AudioContextConstructor = window.AudioContext || (window as typeof window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext
  if (!AudioContextConstructor) return null
  if (!notificationAudioContext) notificationAudioContext = new AudioContextConstructor()
  return notificationAudioContext
}

const unlockCustomerServiceNotificationAudio = (): void => {
  const context = getCustomerServiceNotificationAudioContext()
  if (context?.state === 'suspended') void context.resume().catch(() => undefined)
}

if (typeof window !== 'undefined') {
  window.addEventListener('pointerdown', unlockCustomerServiceNotificationAudio, { passive: true })
  window.addEventListener('keydown', unlockCustomerServiceNotificationAudio, { passive: true })
}

export const playCustomerServiceBrowserMessageNotificationSound = (volume: number): void => {
  const context = getCustomerServiceNotificationAudioContext()
  if (!context || context.state === 'suspended' || volume <= 0) return

  const now = context.currentTime
  const gain = context.createGain()
  const oscillator = context.createOscillator()
  oscillator.type = 'sine'
  oscillator.frequency.setValueAtTime(880, now)
  oscillator.frequency.setValueAtTime(660, now + 0.12)
  gain.gain.setValueAtTime(0.0001, now)
  gain.gain.exponentialRampToValueAtTime(volume, now + 0.012)
  gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.3)
  oscillator.connect(gain)
  gain.connect(context.destination)
  oscillator.start(now)
  oscillator.stop(now + 0.3)
}

export const showCustomerServiceBrowserMessageDesktopNotification = (event: Record<string, any>): void => {
  if (typeof window === 'undefined' || !('Notification' in window)) return

  const showGrantedCustomerServiceDesktopNotification = () => {
    if (Notification.permission !== 'granted') return
    try {
      const notification = new Notification('收到新的客户消息', {
        body: '客服会话中有一条新消息，请及时处理。',
        tag: `customer-service-${event.event_id || event.ticket_id || 'message'}`,
      })
      notification.onclick = () => {
        window.focus()
        notification.close()
      }
    } catch {
      // Notification can still be unavailable in restricted browser contexts.
    }
  }

  if (Notification.permission === 'granted') {
    showGrantedCustomerServiceDesktopNotification()
    return
  }
  // Permission prompts must follow a deliberate user gesture. The first
  // realtime event only plays the sound; the UI can call requestPermission
  // from its notification-settings control when the agent opts in.
}
