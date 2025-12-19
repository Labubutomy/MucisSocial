import { useState, useEffect } from 'react'
import { useUnit } from 'effector-react'
import { Card } from '@shared/ui/card'
import { Button } from '@shared/ui/button'
import { Input } from '@shared/ui/input'
import { Avatar } from '@shared/ui/avatar'
import { $user } from '@features/auth'
import {
  $friends,
  $incomingRequests,
  $searchResults,
  loadFriendsFx,
  loadIncomingRequestsFx,
  friendRequestResponded,
  searchUsersFx,
} from '@features/friends/model'
import { sendFriendRequest } from '@features/friends/api'
import { routes } from '@shared/router'
import { createDirectConversationFx } from '@features/messaging/model'

export const FriendsPage = () => {
  const [searchQuery, setSearchQuery] = useState('')
  const {
    user,
    friends,
    incomingRequests,
    searchResults,
    loadingFriends,
    loadingRequests,
    searching,
    respond,
    search,
  } = useUnit({
    user: $user,
    friends: $friends,
    incomingRequests: $incomingRequests,
    searchResults: $searchResults,
    loadingFriends: loadFriendsFx.pending,
    loadingRequests: loadIncomingRequestsFx.pending,
    searching: searchUsersFx.pending,
    respond: friendRequestResponded,
    search: searchUsersFx,
  })

  // Debounce для поиска друзей (1 секунда)
  useEffect(() => {
    if (searchQuery.trim().length < 2) {
      return
    }

    const timeoutId = setTimeout(() => {
      search(searchQuery.trim())
    }, 1000)

    return () => clearTimeout(timeoutId)
  }, [searchQuery, search])

  if (!user) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-20 pt-10">
        <p className="text-sm text-muted-foreground">Загрузка профиля...</p>
      </div>
    )
  }

  return (
    <div className="page-container space-y-8 pb-20 pt-10">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold">Друзья</h1>
        <p className="text-sm text-muted-foreground">
          Добавляйте друзей, делитесь музыкой и приглашайте их в совместные сессии.
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-[minmax(0,0.9fr),minmax(0,1.1fr)]">
        <Card padding="lg" className="space-y-4 bg-secondary/20">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs uppercase tracking-[0.4em] text-primary">Входящие заявки</p>
              <h2 className="text-xl font-semibold">
                {loadingRequests ? 'Загрузка...' : `${incomingRequests.length} заявок`}
              </h2>
            </div>
          </div>

          {incomingRequests.length === 0 && !loadingRequests && (
            <p className="text-sm text-muted-foreground">Пока нет новых заявок в друзья.</p>
          )}

          <div className="space-y-3">
            {incomingRequests.map(request => (
              <div
                key={request.id}
                className="flex items-center justify-between rounded-lg bg-background/60 p-3"
              >
                <div>
                  <p className="text-sm font-medium">
                    Запрос от пользователя <span className="font-mono">{request.fromUserId}</span>
                  </p>
                  <p className="text-xs text-muted-foreground">
                    Статус: {request.status === 'pending' ? 'ожидает ответа' : request.status}
                  </p>
                </div>
                {request.status === 'pending' && (
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => respond({ requestId: request.id, accept: true })}
                    >
                      Принять
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => respond({ requestId: request.id, accept: false })}
                    >
                      Отклонить
                    </Button>
                  </div>
                )}
              </div>
            ))}
          </div>
        </Card>

        <Card padding="lg" className="space-y-4 bg-secondary/20">
          <div className="space-y-3">
            <p className="text-xs uppercase tracking-[0.4em] text-primary">Найти друзей</p>
            <div className="flex gap-2">
              <Input
                placeholder="Поиск по имени пользователя"
                value={searchQuery}
                onChange={e => {
                  setSearchQuery(e.target.value)
                  // Поиск будет выполнен автоматически через debounce
                }}
              />
            </div>
            {searching && <p className="text-xs text-muted-foreground">Поиск...</p>}
            {searchResults.length > 0 && (
              <div className="space-y-2">
                {searchResults.map(user => {
                  const isFriend = friends.some(
                    f =>
                      (f.friendId || f.friend_id) === user.id || (f.userId || f.user_id) === user.id
                  )
                  return (
                    <div
                      key={user.id}
                      className="flex items-center justify-between rounded-lg bg-background/60 p-3"
                    >
                      <div className="flex items-center gap-3">
                        <Avatar src={user.avatarUrl} fallback={user.username} size="sm" />
                        <div>
                          <p className="text-sm font-medium">{user.username}</p>
                        </div>
                      </div>
                      {isFriend ? (
                        <span className="text-xs text-muted-foreground">Уже в друзьях</span>
                      ) : (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={async () => {
                            try {
                              await sendFriendRequest(user.id)
                              setSearchQuery('')
                              loadIncomingRequestsFx()
                            } catch (error) {
                              console.error('Не удалось отправить заявку', error)
                            }
                          }}
                        >
                          Добавить
                        </Button>
                      )}
                    </div>
                  )
                })}
              </div>
            )}
            {searchQuery.length >= 2 && searchResults.length === 0 && !searching && (
              <p className="text-xs text-muted-foreground">Пользователи не найдены</p>
            )}
          </div>

          <div className="flex items-center justify-between">
            <div>
              <p className="text-xs uppercase tracking-[0.4em] text-primary">Мои друзья</p>
              <h2 className="text-xl font-semibold">
                {loadingFriends ? 'Загрузка...' : `${friends.length} друзей`}
              </h2>
            </div>
          </div>

          {friends.length === 0 && !loadingFriends && (
            <p className="text-sm text-muted-foreground">
              Здесь появятся пользователи, с которыми вы подружились.
            </p>
          )}

          <div className="space-y-3">
            {friends.map(friend => {
              const friendId = friend.friendId || friend.friend_id || ''
              const friendUsername = friend.friendUsername || friend.friend_username
              const friendAvatarUrl = friend.friendAvatarUrl || friend.friend_avatar_url
              const createdAt = friend.createdAt || friend.created_at

              return (
                <div
                  key={`${friend.userId || friend.user_id}-${friendId}`}
                  className="flex items-center justify-between rounded-lg bg-background/60 p-3 cursor-pointer hover:bg-background/80 transition-colors"
                  onClick={async () => {
                    try {
                      // Создаем или получаем диалог с другом (бекенд сам найдет существующий или создаст новый)
                      await createDirectConversationFx(friendId)
                      // Переходим на страницу сообщений (диалог откроется автоматически через эффект в модели)
                      routes.messages.navigate({ params: {}, query: {} })
                    } catch (error) {
                      console.error('Не удалось открыть чат с другом', error)
                      // В случае ошибки все равно переходим на страницу сообщений
                      routes.messages.navigate({ params: {}, query: {} })
                    }
                  }}
                >
                  <div className="flex items-center gap-3 flex-1">
                    <Avatar
                      src={friendAvatarUrl || undefined}
                      fallback={friendUsername || friendId || 'Д'}
                      size="sm"
                    />
                    <div className="flex-1">
                      <p className="text-sm font-medium">
                        {friendUsername ||
                          (friendId ? `Друг ${String(friendId).slice(0, 8)}...` : 'Друг')}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {createdAt
                          ? `Добавлен: ${new Date(createdAt).toLocaleDateString('ru-RU', {
                              year: 'numeric',
                              month: 'long',
                              day: 'numeric',
                            })}`
                          : ''}
                      </p>
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </Card>
      </div>
    </div>
  )
}
