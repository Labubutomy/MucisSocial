import { createApiClient } from '@shared/api/client'
import { API_CONFIG } from '@shared/config/api'

const gatewayClient = createApiClient(API_CONFIG.gateway)

export interface Friend {
  userId?: string
  user_id?: string
  friendId?: string
  friend_id?: string
  createdAt?: string
  created_at?: string
  friendUsername?: string
  friend_username?: string
  friendAvatarUrl?: string
  friend_avatar_url?: string | null
}

export interface FriendRequest {
  id: string
  fromUserId: string
  toUserId: string
  status: 'pending' | 'accepted' | 'declined' | 'cancelled'
  createdAt: string
}

export const fetchFriends = async (): Promise<Friend[]> => {
  const { data } = await gatewayClient.get<Friend[]>('/api/v1/friends')
  return data
}

export const fetchIncomingRequests = async (): Promise<FriendRequest[]> => {
  const { data } = await gatewayClient.get<FriendRequest[]>('/api/v1/friends/requests/incoming')
  return data
}

export const sendFriendRequest = async (toUserId: string): Promise<FriendRequest> => {
  const { data } = await gatewayClient.post<FriendRequest>('/api/v1/friends/requests', {
    to_user_id: toUserId,
  })
  return data
}

export const respondToFriendRequest = async (requestId: string, accept: boolean): Promise<void> => {
  await gatewayClient.post(`/api/v1/friends/requests/${requestId}/respond`, {
    accept,
  })
  // Если запрос принят, автоматически создаем диалог
  if (accept) {
    try {
      // Получаем информацию о запросе, чтобы узнать from_user_id
      const requests = await fetchIncomingRequests()
      const request = requests.find(r => r.id === requestId)
      if (request) {
        // Импортируем функцию создания диалога
        const { createDirectConversation } = await import('@features/messaging/api')
        await createDirectConversation(request.fromUserId)
      }
    } catch (error) {
      console.error('Failed to create conversation for new friend:', error)
      // Не критично, диалог можно создать позже
    }
  }
}

export interface UserSearchResult {
  id: string
  username: string
  avatarUrl?: string
}

export const searchUsers = async (query: string, limit = 10): Promise<UserSearchResult[]> => {
  const { data } = await gatewayClient.get<{ users: UserSearchResult[] }>('/api/v1/users/search', {
    params: {
      q: query,
      limit,
    },
  })
  return data.users || []
}
