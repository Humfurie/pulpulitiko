import type { WSMessage, WSMessageType } from '~/types'

interface WebSocketState {
  connected: boolean
  connecting: boolean
  error: string | null
}

type MessageHandler = (message: WSMessage) => void

export function useWebSocket() {
  const config = useRuntimeConfig()
  const auth = useAuth()

  const socket = ref<WebSocket | null>(null)
  const state = reactive<WebSocketState>({
    connected: false,
    connecting: false,
    error: null
  })

  const messageHandlers = new Map<WSMessageType | 'all', Set<MessageHandler>>()
  const reconnectAttempts = ref(0)
  const maxReconnectAttempts = 5
  const reconnectDelay = ref(1000)

  // Get WebSocket URL (convert http to ws)
  function getWsUrl(): string {
    const apiUrl = config.public.apiUrl as string
    const wsProtocol = apiUrl.startsWith('https') ? 'wss' : 'ws'
    const wsUrl = apiUrl.replace(/^https?/, wsProtocol).replace('/api', '')
    return `${wsUrl}/ws?token=${auth.token.value}`
  }

  // Connect to WebSocket
  function connect() {
    // Only connect on client side and when authenticated
    if (import.meta.server || !auth.isAuthenticated.value || state.connecting || state.connected) {
      return
    }

    state.connecting = true
    state.error = null

    try {
      const wsUrl = getWsUrl()
      socket.value = new WebSocket(wsUrl)

      socket.value.onopen = () => {
        state.connected = true
        state.connecting = false
        state.error = null
        reconnectAttempts.value = 0
        reconnectDelay.value = 1000
        console.log('WebSocket connected')
      }

      socket.value.onclose = (event) => {
        state.connected = false
        state.connecting = false
        console.log('WebSocket closed:', event.code, event.reason)

        // Attempt to reconnect if not a normal closure
        if (event.code !== 1000 && reconnectAttempts.value < maxReconnectAttempts) {
          setTimeout(() => {
            reconnectAttempts.value++
            reconnectDelay.value = Math.min(reconnectDelay.value * 2, 30000)
            connect()
          }, reconnectDelay.value)
        }
      }

      socket.value.onerror = (error) => {
        state.error = 'WebSocket connection error'
        state.connecting = false
        console.error('WebSocket error:', error)
      }

      socket.value.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data)
          handleMessage(message)
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }
    } catch (error) {
      state.connecting = false
      state.error = 'Failed to create WebSocket connection'
      console.error('WebSocket connection error:', error)
    }
  }

  // Disconnect from WebSocket
  function disconnect() {
    if (socket.value) {
      socket.value.close(1000, 'User disconnected')
      socket.value = null
    }
    state.connected = false
    state.connecting = false
  }

  // Send a message through WebSocket
  function send(message: Partial<WSMessage>) {
    if (!socket.value || socket.value.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket is not connected')
      return false
    }

    try {
      socket.value.send(JSON.stringify({
        ...message,
        timestamp: new Date().toISOString()
      }))
      return true
    } catch (error) {
      console.error('Failed to send WebSocket message:', error)
      return false
    }
  }

  // Handle incoming messages
  function handleMessage(message: WSMessage) {
    // Call type-specific handlers
    const typeHandlers = messageHandlers.get(message.type)
    if (typeHandlers) {
      typeHandlers.forEach(handler => handler(message))
    }

    // Call 'all' handlers
    const allHandlers = messageHandlers.get('all')
    if (allHandlers) {
      allHandlers.forEach(handler => handler(message))
    }
  }

  // Register a message handler
  function onMessage(type: WSMessageType | 'all', handler: MessageHandler) {
    if (!messageHandlers.has(type)) {
      messageHandlers.set(type, new Set())
    }
    messageHandlers.get(type)!.add(handler)

    // Return cleanup function
    return () => {
      messageHandlers.get(type)?.delete(handler)
    }
  }

  // Send typing indicator
  function sendTyping(conversationId: string) {
    send({
      type: 'typing',
      conversation_id: conversationId
    })
  }

  // Send stop typing indicator
  function sendStopTyping(conversationId: string) {
    send({
      type: 'stop_typing',
      conversation_id: conversationId
    })
  }

  // Send message read notification
  function sendMessageRead(conversationId: string) {
    send({
      type: 'message_read',
      conversation_id: conversationId
    })
  }

  // Auto-connect when auth state changes
  watch(() => auth.isAuthenticated.value, (isAuthenticated) => {
    if (isAuthenticated && !state.connected && !state.connecting) {
      connect()
    } else if (!isAuthenticated && state.connected) {
      disconnect()
    }
  }, { immediate: true })

  // Cleanup on unmount
  onUnmounted(() => {
    disconnect()
  })

  return {
    // State
    connected: computed(() => state.connected),
    connecting: computed(() => state.connecting),
    error: computed(() => state.error),

    // Methods
    connect,
    disconnect,
    send,
    onMessage,
    sendTyping,
    sendStopTyping,
    sendMessageRead
  }
}
