import type {
  ApiResponse,
  Conversation,
  CreateConversationRequest,
  CreateMessageRequest,
  Message,
  PaginatedConversations,
  PaginatedMessages,
  UnreadCounts,
  UpdateConversationRequest,
  WSMessage
} from '~/types'

export function useMessaging() {
  const config = useRuntimeConfig()
  const auth = useAuth()
  const ws = useWebSocket()

  const baseUrl = config.public.apiUrl as string

  // State
  const conversations = ref<Conversation[]>([])
  const currentConversation = ref<Conversation | null>(null)
  const messages = ref<Message[]>([])
  const unreadCounts = ref<UnreadCounts>({ total: 0, conversations: 0 })
  const loading = ref(false)
  const error = ref<string | null>(null)
  const typingUsers = ref<Set<string>>(new Set())

  // Typing indicator timeout
  let typingTimeout: ReturnType<typeof setTimeout> | null = null

  type HttpMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH'

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  type RequestBody = Record<string, any> | null | undefined

  // Helper to make authenticated requests
  async function fetchApi<T>(endpoint: string, options?: {
    method?: HttpMethod
    body?: RequestBody
  }): Promise<T> {
    const headers = auth.getAuthHeaders()

    const response = await $fetch<ApiResponse<T>>(`${baseUrl}${endpoint}`, {
      method: options?.method || 'GET',
      headers: {
        'Content-Type': 'application/json',
        ...headers
      },
      body: options?.body
    })

    if (!response.success) {
      throw new Error((response as unknown as { error: string }).error || 'Request failed')
    }

    return response.data
  }

  // Fetch unread counts
  async function fetchUnreadCounts() {
    try {
      unreadCounts.value = await fetchApi<UnreadCounts>('/messages/unread')
    } catch (err) {
      console.error('Failed to fetch unread counts:', err)
    }
  }

  // Fetch user's conversations
  async function fetchConversations() {
    loading.value = true
    error.value = null
    try {
      conversations.value = await fetchApi<Conversation[]>('/messages/conversations')
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch conversations'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Fetch admin conversations list (paginated)
  async function fetchAdminConversations(page = 1, perPage = 20, status?: string) {
    loading.value = true
    error.value = null
    try {
      let url = `/admin/messages/conversations?page=${page}&per_page=${perPage}`
      if (status) {
        url += `&status=${status}`
      }
      return await fetchApi<PaginatedConversations>(url)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch conversations'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Fetch a single conversation
  async function fetchConversation(id: string) {
    loading.value = true
    error.value = null
    try {
      const isAdmin = auth.isAdmin.value
      const endpoint = isAdmin
        ? `/admin/messages/conversations/${id}`
        : `/messages/conversations/${id}`
      currentConversation.value = await fetchApi<Conversation>(endpoint)
      return currentConversation.value
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch conversation'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Fetch messages for a conversation
  async function fetchMessages(conversationId: string, page = 1, perPage = 50) {
    loading.value = true
    error.value = null
    try {
      const isAdmin = auth.isAdmin.value
      const endpoint = isAdmin
        ? `/admin/messages/conversations/${conversationId}/messages`
        : `/messages/conversations/${conversationId}/messages`
      const result = await fetchApi<PaginatedMessages>(`${endpoint}?page=${page}&per_page=${perPage}`)
      messages.value = result.messages || []
      return result
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch messages'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Create a new conversation (user only)
  async function createConversation(data: CreateConversationRequest) {
    loading.value = true
    error.value = null
    try {
      const result = await fetchApi<{ conversation: Conversation; message: Message }>(
        '/messages/conversations',
        { method: 'POST', body: data }
      )
      conversations.value.unshift(result.conversation)
      currentConversation.value = result.conversation
      messages.value = [result.message]
      return result
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create conversation'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Send a message
  async function sendMessage(conversationId: string, data: CreateMessageRequest) {
    error.value = null
    try {
      const isAdmin = auth.isAdmin.value
      const endpoint = isAdmin
        ? `/admin/messages/conversations/${conversationId}/messages`
        : `/messages/conversations/${conversationId}/messages`
      const message = await fetchApi<Message>(endpoint, { method: 'POST', body: data })
      messages.value.push(message)

      // Update conversation's last message
      const conversation = conversations.value.find(c => c.id === conversationId)
      if (conversation) {
        conversation.last_message = message
        conversation.last_message_at = message.created_at
      }

      return message
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to send message'
      throw err
    }
  }

  // Mark messages as read
  async function markAsRead(conversationId: string) {
    try {
      const isAdmin = auth.isAdmin.value
      const endpoint = isAdmin
        ? `/admin/messages/conversations/${conversationId}/read`
        : `/messages/conversations/${conversationId}/read`
      await fetchApi<{ success: boolean }>(endpoint, { method: 'POST' })

      // Update local state
      messages.value.forEach(msg => {
        if (!msg.is_read && msg.sender_id !== auth.user.value?.id) {
          msg.is_read = true
        }
      })

      // Update unread counts
      const conversation = conversations.value.find(c => c.id === conversationId)
      if (conversation && conversation.unread_count) {
        unreadCounts.value.total -= conversation.unread_count
        unreadCounts.value.conversations = Math.max(0, unreadCounts.value.conversations - 1)
        conversation.unread_count = 0
      }

      // Send WebSocket notification
      ws.sendMessageRead(conversationId)
    } catch (err) {
      console.error('Failed to mark messages as read:', err)
    }
  }

  // Update conversation status (admin only)
  async function updateConversationStatus(conversationId: string, data: UpdateConversationRequest) {
    error.value = null
    try {
      await fetchApi<{ success: boolean }>(
        `/admin/messages/conversations/${conversationId}/status`,
        { method: 'PATCH', body: data }
      )

      // Update local state
      const conversation = conversations.value.find(c => c.id === conversationId)
      if (conversation) {
        conversation.status = data.status
      }
      if (currentConversation.value?.id === conversationId) {
        currentConversation.value.status = data.status
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update conversation'
      throw err
    }
  }

  // Handle typing indicator
  function startTyping(conversationId: string) {
    ws.sendTyping(conversationId)

    // Clear existing timeout
    if (typingTimeout) {
      clearTimeout(typingTimeout)
    }

    // Stop typing after 3 seconds of inactivity
    typingTimeout = setTimeout(() => {
      ws.sendStopTyping(conversationId)
    }, 3000)
  }

  // WebSocket message handlers
  ws.onMessage('new_message', (wsMsg: WSMessage) => {
    if (wsMsg.message) {
      // Add message if we're viewing this conversation
      if (currentConversation.value?.id === wsMsg.conversation_id) {
        messages.value.push(wsMsg.message)
      }

      // Update conversation in list
      const conversation = conversations.value.find(c => c.id === wsMsg.conversation_id)
      if (conversation) {
        conversation.last_message = wsMsg.message
        conversation.last_message_at = wsMsg.message.created_at
        if (wsMsg.message.sender_id !== auth.user.value?.id) {
          conversation.unread_count = (conversation.unread_count || 0) + 1
          unreadCounts.value.total++
        }
      }
    }
  })

  ws.onMessage('typing', (wsMsg: WSMessage) => {
    if (wsMsg.user_id && wsMsg.conversation_id === currentConversation.value?.id) {
      typingUsers.value.add(wsMsg.user_id)
    }
  })

  ws.onMessage('stop_typing', (wsMsg: WSMessage) => {
    if (wsMsg.user_id) {
      typingUsers.value.delete(wsMsg.user_id)
    }
  })

  // Cleanup typing timeout on unmount
  onUnmounted(() => {
    if (typingTimeout) {
      clearTimeout(typingTimeout)
    }
  })

  return {
    // State
    conversations,
    currentConversation,
    messages,
    unreadCounts,
    loading,
    error,
    typingUsers: computed(() => Array.from(typingUsers.value)),

    // Methods
    fetchUnreadCounts,
    fetchConversations,
    fetchAdminConversations,
    fetchConversation,
    fetchMessages,
    createConversation,
    sendMessage,
    markAsRead,
    updateConversationStatus,
    startTyping
  }
}
