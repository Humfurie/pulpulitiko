// Composable for handling notification sounds

export function useNotificationSound() {
  const soundEnabled = useState('notification-sound-enabled', () => true)
  const audioContext = ref<AudioContext | null>(null)

  // Initialize audio context on user interaction
  function initAudio() {
    if (typeof window === 'undefined') return
    if (!audioContext.value) {
      // Type-safe way to access webkit prefix
      const AudioContextClass = window.AudioContext || (window as Window & { webkitAudioContext?: typeof AudioContext }).webkitAudioContext
      if (AudioContextClass) {
        audioContext.value = new AudioContextClass()
      }
    }
  }

  // Generate a pleasant notification sound
  function playNotificationSound() {
    if (!soundEnabled.value) return
    if (typeof window === 'undefined') return

    try {
      initAudio()
      if (!audioContext.value) return

      const ctx = audioContext.value

      // Create oscillator for a pleasant chime
      const oscillator = ctx.createOscillator()
      const gainNode = ctx.createGain()

      oscillator.connect(gainNode)
      gainNode.connect(ctx.destination)

      // Pleasant notification tone
      oscillator.frequency.setValueAtTime(880, ctx.currentTime) // A5
      oscillator.frequency.setValueAtTime(1100, ctx.currentTime + 0.1) // C#6

      oscillator.type = 'sine'

      // Fade in and out
      gainNode.gain.setValueAtTime(0, ctx.currentTime)
      gainNode.gain.linearRampToValueAtTime(0.3, ctx.currentTime + 0.05)
      gainNode.gain.linearRampToValueAtTime(0, ctx.currentTime + 0.3)

      oscillator.start(ctx.currentTime)
      oscillator.stop(ctx.currentTime + 0.3)
    } catch (error) {
      console.warn('Could not play notification sound:', error)
    }
  }

  // Play a softer sound for message sent confirmation
  function playMessageSentSound() {
    if (!soundEnabled.value) return
    if (typeof window === 'undefined') return

    try {
      initAudio()
      if (!audioContext.value) return

      const ctx = audioContext.value

      const oscillator = ctx.createOscillator()
      const gainNode = ctx.createGain()

      oscillator.connect(gainNode)
      gainNode.connect(ctx.destination)

      oscillator.frequency.setValueAtTime(600, ctx.currentTime)
      oscillator.type = 'sine'

      gainNode.gain.setValueAtTime(0, ctx.currentTime)
      gainNode.gain.linearRampToValueAtTime(0.15, ctx.currentTime + 0.02)
      gainNode.gain.linearRampToValueAtTime(0, ctx.currentTime + 0.1)

      oscillator.start(ctx.currentTime)
      oscillator.stop(ctx.currentTime + 0.1)
    } catch (error) {
      console.warn('Could not play sound:', error)
    }
  }

  function toggleSound() {
    soundEnabled.value = !soundEnabled.value
    // Save preference
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('notification-sound', soundEnabled.value.toString())
    }
  }

  // Load preference
  onMounted(() => {
    if (typeof localStorage !== 'undefined') {
      const saved = localStorage.getItem('notification-sound')
      if (saved !== null) {
        soundEnabled.value = saved === 'true'
      }
    }
  })

  return {
    soundEnabled: readonly(soundEnabled),
    playNotificationSound,
    playMessageSentSound,
    toggleSound
  }
}
