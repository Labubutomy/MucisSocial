import { useState, useEffect, useRef } from 'react'
import { useUnit } from 'effector-react'
import { Card } from '@shared/ui/card'
import { Button } from '@shared/ui/button'
import { Input } from '@shared/ui/input'
import { Avatar } from '@shared/ui/avatar'
import { $user } from '@features/auth'
import {
  $activeConversationId,
  $conversations,
  $messages,
  conversationOpened,
  loadConversationsFx,
  loadMessagesFx,
  messageSendRequested,
  sendMessageFx,
  createDirectConversationFx,
} from '@features/messaging/model'
import { TrackCard } from '@entities/track/ui/track-card'
import { routes } from '@shared/router'
import { $friends, $searchResults, loadFriendsFx, searchUsersFx } from '@features/friends/model'
import { sendFriendRequest } from '@features/friends/api'
import { fetchTrackDetail } from '@entities/track/api'
import type { Track } from '@entities/track'
import { TrackSelector } from '@pages/create-route/ui/track-selector'
import { createRoom } from '@features/session/api'
import { fetchUserById } from '@entities/user/api'
import { trackQueued } from '@features/player'
import { ShareTrackPopup } from '@shared/ui/share-track-popup'

export const MessagesPage = () => {
  const {
    user,
    conversations,
    messages,
    activeConversationId,
    openConversation,
    conversationsLoading,
    messagesLoading,
    sendPending,
    sendMessage,
    createDirectConversationFxEffect,
    friends,
    searchResults,
    searching,
    loadFriends,
    searchUsers,
    navigateToFriends,
  } = useUnit({
    user: $user,
    conversations: $conversations,
    messages: $messages,
    activeConversationId: $activeConversationId,
    openConversation: conversationOpened,
    conversationsLoading: loadConversationsFx.pending,
    messagesLoading: loadMessagesFx.pending,
    sendPending: sendMessageFx.pending,
    sendMessage: messageSendRequested,
    friends: $friends,
    searchResults: $searchResults,
    searching: searchUsersFx.pending,
    loadFriends: loadFriendsFx,
    searchUsers: searchUsersFx,
    navigateToFriends: routes.friends.navigate,
    createDirectConversationFxEffect: createDirectConversationFx,
  })

  const [draft, setDraft] = useState('')
  const [showFriendsSearch, setShowFriendsSearch] = useState(false)
  const [friendsSearchQuery, setFriendsSearchQuery] = useState('')
  const [showTrackSelector, setShowTrackSelector] = useState(false)
  const [selectedTrackId, setSelectedTrackId] = useState<string | null>(null)
  const [trackCache, setTrackCache] = useState<Record<string, Track>>({})
  const [userCache, setUserCache] = useState<
    Record<string, { username: string; avatarUrl?: string }>
  >({})
  const [shareTrackPopup, setShareTrackPopup] = useState<{ track: Track } | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const messagesContainerRef = useRef<HTMLDivElement>(null)

  // Загружаем информацию о треках и пользователях из сообщений
  useEffect(() => {
    const loadTracks = async () => {
      const trackIds = messages
        .filter(m => (m.trackId || m.track_id) && !trackCache[m.trackId || m.track_id || ''])
        .map(m => m.trackId || m.track_id || '')
        .filter(id => id !== '')

      if (trackIds.length === 0) return

      for (const trackId of trackIds) {
        try {
          const trackDetail = await fetchTrackDetail(trackId)
          setTrackCache(prev => ({ ...prev, [trackId]: trackDetail }))
        } catch (error) {
          console.error(`Failed to load track ${trackId}:`, error)
        }
      }
    }

    const loadUsers = async () => {
      // Собираем уникальные sender_id из сообщений
      // ВАЖНО: нормализуем ID в строки для корректного сравнения
      const userIds = Array.from(
        new Set(
          messages
            .map(m => String(m.senderId || ''))
            .filter(id => id && id !== '' && !userCache[id] && id !== String(user?.id || ''))
        )
      )

      if (userIds.length === 0) return

      // Загружаем информацию о пользователях параллельно
      await Promise.all(
        userIds.map(async userId => {
          try {
            const userInfo = await fetchUserById(userId)
            setUserCache((prev: Record<string, { username: string; avatarUrl?: string }>) => ({
              ...prev,
              [userId]: {
                username: userInfo.username,
                avatarUrl: userInfo.avatarUrl,
              },
            }))
          } catch (error) {
            console.error(`Failed to load user ${userId}:`, error)
            // Добавляем в кэш с дефолтным значением, чтобы не пытаться загружать снова
            setUserCache((prev: Record<string, { username: string; avatarUrl?: string }>) => ({
              ...prev,
              [userId]: { username: 'Друг', avatarUrl: undefined },
            }))
          }
        })
      )
    }

    loadTracks()
    loadUsers()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [messages])

  // Автоматическая прокрутка вниз при изменении сообщений или открытии диалога
  useEffect(() => {
    if (messagesContainerRef.current && messages.length > 0) {
      // Используем setTimeout для гарантии, что DOM обновлен
      setTimeout(() => {
        if (messagesContainerRef.current) {
          messagesContainerRef.current.scrollTop = messagesContainerRef.current.scrollHeight
        }
      }, 100)
    }
  }, [messages, activeConversationId])

  if (!user) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-20 pt-10">
        <p className="text-sm text-muted-foreground">Загрузка профиля...</p>
      </div>
    )
  }

  const handleSend = async () => {
    if (!activeConversationId || sendPending) return
    const content = draft.trim()
    if (!content && !selectedTrackId) return

    sendMessage({
      content: content || '', // Пустая строка разрешена, если есть trackId
      trackId: selectedTrackId || undefined,
    })
    setDraft('')
    setSelectedTrackId(null)
  }

  const handleShareTrackToFriend = async (friendId: string, track: Track) => {
    try {
      // Создаем или получаем диалог с другом
      const conversation = await createDirectConversationFxEffect(friendId)

      // Отправляем трек в диалог (используем sendMessageFx напрямую через API)
      const { sendMessage: sendMessageApi } = await import('@features/messaging/api')
      await sendMessageApi({
        conversationId: conversation.id,
        trackId: track.id,
        content: `🎵 ${track.title} - ${track.artist.name}`,
      })
    } catch (error) {
      console.error('Failed to share track to friend:', error)
    }
  }

  const handleInviteToSession = async () => {
    if (!activeConversationId || !user) return
    try {
      // Создаем комнату для совместного прослушивания
      const roomId = `room-${Math.random().toString(36).slice(2, 8)}`
      await createRoom(roomId)
      const sessionLink = `${window.location.origin}/listen/${roomId}`

      // Отправляем сообщение с приглашением
      sendMessage({
        content: `🎧 Приглашаю в совместное прослушивание: ${sessionLink}`,
      })
    } catch (error) {
      console.error('Не удалось создать сессию:', error)
    }
  }

  return (
    <>
      {showTrackSelector && (
        <TrackSelector
          onSelect={async trackId => {
            setSelectedTrackId(trackId)
            // Загружаем информацию о треке
            try {
              const trackDetail = await fetchTrackDetail(trackId)
              setTrackCache(prev => ({ ...prev, [trackId]: trackDetail }))
            } catch (error) {
              console.error('Failed to load track:', error)
            }
          }}
          onClose={() => setShowTrackSelector(false)}
        />
      )}
      {shareTrackPopup && (
        <ShareTrackPopup
          track={shareTrackPopup.track}
          friends={friends}
          onSelectFriend={handleShareTrackToFriend}
          onClose={() => setShareTrackPopup(null)}
        />
      )}
      <div className="page-container space-y-8 pb-20 pt-10">
        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-semibold">Сообщения</h1>
              <p className="text-sm text-muted-foreground">
                Общайтесь с друзьями и обсуждайте музыку. Через этот экран позже можно будет
                делиться треками и сессиями.
              </p>
            </div>
            <Button
              variant="outline"
              onClick={() => {
                setShowFriendsSearch(!showFriendsSearch)
                if (!showFriendsSearch) {
                  loadFriends()
                }
              }}
            >
              {showFriendsSearch ? 'Скрыть' : 'Найти друзей'}
            </Button>
          </div>
        </div>

        {showFriendsSearch && (
          <Card padding="lg" className="space-y-4 bg-secondary/20">
            <div className="space-y-3">
              <p className="text-xs uppercase tracking-[0.4em] text-primary">Поиск друзей</p>
              <div className="flex gap-2">
                <Input
                  placeholder="Поиск по имени пользователя"
                  value={friendsSearchQuery}
                  onChange={e => {
                    const query = e.target.value.trim()
                    setFriendsSearchQuery(query)
                    if (query.length >= 2) {
                      searchUsers(query)
                    }
                  }}
                />
                <Button
                  variant="outline"
                  onClick={() => navigateToFriends({ params: {}, query: {} })}
                >
                  Все друзья
                </Button>
              </div>
              {searching && <p className="text-xs text-muted-foreground">Поиск...</p>}
              {searchResults.length > 0 && (
                <div className="space-y-2">
                  {searchResults.map(user => {
                    const isFriend = friends.some(
                      f => f.friendId === user.id || f.userId === user.id
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
                            <p className="text-xs text-muted-foreground">ID: {user.id}</p>
                          </div>
                        </div>
                        {!isFriend && (
                          <Button
                            size="sm"
                            variant="outline"
                            onClick={async () => {
                              try {
                                await sendFriendRequest(user.id)
                                setFriendsSearchQuery('')
                                loadFriends()
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
              {friendsSearchQuery.length >= 2 && searchResults.length === 0 && !searching && (
                <p className="text-xs text-muted-foreground">Пользователи не найдены</p>
              )}
            </div>
          </Card>
        )}

        <div className="grid gap-6 lg:grid-cols-[minmax(0,0.8fr),minmax(0,1.2fr)]">
          <Card padding="lg" className="space-y-4 bg-secondary/20">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-xs uppercase tracking-[0.4em] text-primary">Диалоги</p>
                <h2 className="text-xl font-semibold">
                  {conversationsLoading ? 'Загрузка...' : `${conversations.length} диалогов`}
                </h2>
              </div>
            </div>

            <div className="space-y-2">
              {conversations.length === 0 && !conversationsLoading && (
                <p className="text-sm text-muted-foreground">
                  Диалоги появятся, когда вы начнете переписку с друзьями.
                </p>
              )}

              {conversations.map(conv => (
                <button
                  key={conv.id}
                  type="button"
                  onClick={() => openConversation({ conversationId: conv.id })}
                  className={`flex w-full items-center justify-between rounded-lg px-3 py-2 text-left text-sm transition hover:bg-background/70 ${
                    activeConversationId === conv.id ? 'bg-background' : 'bg-background/40'
                  }`}
                >
                  <div className="space-y-1">
                    <p className="font-medium">
                      {conv.title ||
                        (conv.conversationType === 'group'
                          ? 'Групповой чат'
                          : (() => {
                              // Для прямого диалога находим имя друга
                              const otherParticipantId =
                                conv.otherParticipantId || conv.other_participant_id
                              if (otherParticipantId) {
                                // В списке друзей user_id - это текущий пользователь, friend_id - это друг
                                // Ищем запись, где friend_id совпадает с otherParticipantId
                                const friend = friends.find(f => {
                                  const friendId = f.friendId || f.friend_id
                                  return String(friendId) === String(otherParticipantId)
                                })
                                if (friend) {
                                  const username = friend.friendUsername || friend.friend_username
                                  if (username) return username
                                }
                              }
                              // Fallback: ищем в сообщениях
                              const otherParticipant = messages.find(
                                m => m.senderId !== user.id && m.conversationId === conv.id
                              )?.senderId
                              if (otherParticipant) {
                                const friend = friends.find(f => {
                                  const friendId = f.friendId || f.friend_id
                                  return String(friendId) === String(otherParticipant)
                                })
                                if (friend) {
                                  const username = friend.friendUsername || friend.friend_username
                                  if (username) return username
                                }
                              }
                              // Если не нашли друга, показываем "Друг"
                              return 'Друг'
                            })())}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {conv.updatedAt && !isNaN(new Date(conv.updatedAt).getTime())
                        ? `Обновлено: ${new Date(conv.updatedAt).toLocaleString('ru-RU')}`
                        : ''}
                    </p>
                  </div>
                  {conv.unreadCount > 0 && (
                    <span className="inline-flex h-6 min-w-[1.5rem] items-center justify-center rounded-full bg-primary px-1 text-xs font-semibold text-primary-foreground">
                      {conv.unreadCount}
                    </span>
                  )}
                </button>
              ))}
            </div>
          </Card>

          <Card padding="lg" className="flex min-h-[320px] max-h-[70vh] flex-col bg-secondary/20">
            {activeConversationId ? (
              <>
                <div className="mb-3 flex items-center justify-between">
                  <div>
                    <p className="text-xs uppercase tracking-[0.4em] text-primary">Чат</p>
                    <h2 className="text-lg font-semibold">
                      {(() => {
                        const conv = conversations.find(c => c.id === activeConversationId)
                        if (!conv) return 'Диалог без названия'
                        if (conv.title) return conv.title
                        if (conv.conversationType === 'group') return 'Групповой чат'
                        // Для прямого диалога находим имя друга
                        const otherParticipantId =
                          conv.otherParticipantId || conv.other_participant_id
                        if (otherParticipantId) {
                          // Сначала проверяем кэш пользователей
                          const cachedUser = userCache[String(otherParticipantId)]
                          if (cachedUser) {
                            return cachedUser.username
                          }
                          // Затем в списке друзей
                          const friend = friends.find(f => {
                            const friendId = f.friendId || f.friend_id
                            return String(friendId) === String(otherParticipantId)
                          })
                          if (friend) {
                            const username = friend.friendUsername || friend.friend_username
                            if (username) return username
                          }
                        }
                        // Fallback: ищем в сообщениях
                        const otherParticipant = messages.find(
                          m =>
                            String(m.senderId) !== String(user?.id) && m.conversationId === conv.id
                        )?.senderId
                        if (otherParticipant) {
                          const cachedUser = userCache[String(otherParticipant)]
                          if (cachedUser) {
                            return cachedUser.username
                          }
                          const friend = friends.find(f => {
                            const friendId = f.friendId || f.friend_id
                            return String(friendId) === String(otherParticipant)
                          })
                          if (friend) {
                            const username = friend.friendUsername || friend.friend_username
                            if (username) return username
                          }
                        }
                        return 'Диалог'
                      })()}
                    </h2>
                  </div>
                </div>

                <div
                  ref={messagesContainerRef}
                  className="flex-1 space-y-2 overflow-y-auto rounded-lg p-3 text-sm max-h-[calc(70vh-120px)] [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]"
                >
                  {messagesLoading && (
                    <p className="text-xs text-muted-foreground">Загрузка сообщений...</p>
                  )}
                  {messages.length === 0 && !messagesLoading && (
                    <p className="text-xs text-muted-foreground">
                      Напишите первое сообщение, чтобы начать диалог.
                    </p>
                  )}
                  {messages.map((msg, index) => {
                    // Добавляем ref для последнего сообщения
                    const isLastMessage = index === messages.length - 1
                    // Нормализуем ID для сравнения (на случай разных форматов)
                    const msgSenderId = String(msg.senderId || '')
                    const currentUserId = String(user?.id || '')
                    const isMine = msgSenderId === currentUserId && currentUserId !== ''
                    const isTrackMsg = Boolean(msg.trackId)
                    const prevMsg = index > 0 ? messages[index - 1] : null
                    const showAvatar = !prevMsg || String(prevMsg.senderId || '') !== msgSenderId

                    // Находим информацию об отправителе
                    let senderInfo: { id: string; username: string; avatarUrl?: string }
                    if (isMine && user) {
                      senderInfo = {
                        id: user.id,
                        username: user.username,
                        avatarUrl: user.avatarUrl,
                      }
                    } else {
                      // Если сообщение не от текущего пользователя, получаем информацию из кэша пользователей
                      const senderId = String(msg.senderId || '')
                      const cachedUser = userCache[senderId]
                      if (cachedUser) {
                        senderInfo = {
                          id: senderId,
                          username: cachedUser.username,
                          avatarUrl: cachedUser.avatarUrl,
                        }
                      } else {
                        // Fallback: пытаемся найти в списке друзей (пока пользователь загружается)
                        const friend = friends.find(f => {
                          const friendId = f.friendId || f.friend_id
                          return String(friendId) === senderId
                        })
                        if (friend) {
                          senderInfo = {
                            id: senderId,
                            username: friend.friendUsername || friend.friend_username || 'Друг',
                            avatarUrl:
                              friend.friendAvatarUrl || friend.friend_avatar_url || undefined,
                          }
                        } else {
                          // Если не нашли, показываем "Друг" (пользователь будет загружен в useEffect)
                          senderInfo = {
                            id: senderId,
                            username: 'Друг',
                            avatarUrl: undefined,
                          }
                        }
                      }
                    }

                    // Поддерживаем оба формата: snake_case и camelCase
                    const createdAt = msg.createdAt || msg.created_at
                    const messageDate =
                      createdAt && !isNaN(new Date(createdAt).getTime())
                        ? new Date(createdAt)
                        : null

                    return (
                      <div
                        key={msg.id}
                        ref={isLastMessage ? messagesEndRef : null}
                        className={`flex items-end gap-2 ${isMine ? 'flex-row-reverse' : 'flex-row'}`}
                      >
                        {showAvatar && (
                          <Avatar
                            src={senderInfo.avatarUrl || undefined}
                            fallback={senderInfo.username}
                            size="sm"
                            className="flex-shrink-0"
                          />
                        )}
                        {!showAvatar && <div className="w-10 flex-shrink-0" />}
                        <div
                          className={`flex flex-col ${isMine ? 'items-end' : 'items-start'} max-w-[70%]`}
                        >
                          {showAvatar && !isMine && (
                            <p className="mb-1 text-xs text-muted-foreground">
                              {senderInfo.username}
                            </p>
                          )}
                          <div
                            className={`rounded-2xl px-4 py-2 ${
                              isMine
                                ? 'bg-primary text-primary-foreground rounded-br-sm'
                                : 'bg-secondary/40 rounded-bl-sm'
                            }`}
                          >
                            {isTrackMsg && (msg.trackId || msg.track_id) && (
                              <div className="mb-2">
                                {trackCache[msg.trackId || msg.track_id || ''] ? (
                                  <TrackCard
                                    track={trackCache[msg.trackId || msg.track_id || '']}
                                    isPlaying={false}
                                    onPlayToggle={() => {
                                      const track = trackCache[msg.trackId || msg.track_id || '']
                                      if (track) {
                                        trackQueued(track)
                                      }
                                    }}
                                    onLike={() => {}}
                                    onShare={() => {
                                      const track = trackCache[msg.trackId || msg.track_id || '']
                                      if (track) {
                                        setShareTrackPopup({ track })
                                      }
                                    }}
                                    onOpen={() => {
                                      routes.track.navigate({
                                        params: { trackId: msg.trackId || msg.track_id || '' },
                                        query: {},
                                      })
                                    }}
                                  />
                                ) : (
                                  <div className="rounded-lg bg-secondary/50 p-3">
                                    <p className="text-sm">Загрузка трека...</p>
                                  </div>
                                )}
                              </div>
                            )}
                            {msg.content && (
                              <div className="whitespace-pre-wrap break-words text-sm">
                                {msg.content.split('\n').map((line, i, arr) => {
                                  // Проверяем, является ли строка ссылкой на совместное прослушивание
                                  const sessionLinkMatch = line.match(/\/listen\/([a-z0-9-]+)/i)
                                  if (sessionLinkMatch) {
                                    const roomId = sessionLinkMatch[1]
                                    return (
                                      <div key={i} className="mt-2">
                                        <Button
                                          size="sm"
                                          variant="outline"
                                          onClick={() => {
                                            routes.sessionRoom.navigate({
                                              params: { roomId },
                                              query: {},
                                            })
                                          }}
                                          className="w-full"
                                        >
                                          🎧 Присоединиться
                                        </Button>
                                      </div>
                                    )
                                  }
                                  return (
                                    <span key={i}>
                                      {line}
                                      {i < arr.length - 1 && <br />}
                                    </span>
                                  )
                                })}
                              </div>
                            )}
                            {messageDate && (
                              <p
                                className={`mt-1 text-[10px] ${isMine ? 'opacity-70' : 'text-muted-foreground'}`}
                              >
                                {messageDate.toLocaleDateString('ru-RU', {
                                  day: 'numeric',
                                  month: 'short',
                                  year:
                                    messageDate.getFullYear() !== new Date().getFullYear()
                                      ? 'numeric'
                                      : undefined,
                                })}{' '}
                                {messageDate.toLocaleTimeString('ru-RU', {
                                  hour: '2-digit',
                                  minute: '2-digit',
                                })}
                              </p>
                            )}
                          </div>
                        </div>
                      </div>
                    )
                  })}
                </div>

                <div className="mt-4 space-y-2">
                  {selectedTrackId && trackCache[selectedTrackId] && (
                    <div className="flex items-center justify-between rounded-lg bg-secondary/50 p-2">
                      <div className="flex items-center gap-2">
                        <div className="h-10 w-10 rounded overflow-hidden">
                          <img
                            src={trackCache[selectedTrackId].coverUrl}
                            alt={trackCache[selectedTrackId].title}
                            className="w-full h-full object-cover"
                          />
                        </div>
                        <div>
                          <p className="text-sm font-medium">{trackCache[selectedTrackId].title}</p>
                          <p className="text-xs text-muted-foreground">
                            {trackCache[selectedTrackId].artist.name}
                          </p>
                        </div>
                      </div>
                      <Button variant="ghost" size="sm" onClick={() => setSelectedTrackId(null)}>
                        ✕
                      </Button>
                    </div>
                  )}
                  <div className="flex gap-2">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => setShowTrackSelector(true)}
                      title="Поделиться треком"
                    >
                      🎵
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={handleInviteToSession}
                      title="Пригласить в совместное прослушивание"
                    >
                      🎧
                    </Button>
                    <Input
                      value={draft}
                      onChange={e => setDraft(e.target.value)}
                      placeholder="Напишите сообщение..."
                      onKeyDown={e => {
                        if (e.key === 'Enter' && !e.shiftKey) {
                          e.preventDefault()
                          handleSend()
                        }
                      }}
                      className="flex-1"
                    />
                    <Button
                      onClick={handleSend}
                      disabled={(!draft.trim() && !selectedTrackId) || sendPending}
                    >
                      Отправить
                    </Button>
                  </div>
                </div>
              </>
            ) : (
              <div className="flex flex-1 flex-col items-center justify-center gap-2 text-center">
                <p className="text-sm font-medium">Выберите диалог слева, чтобы начать общение.</p>
                <p className="max-w-sm text-xs text-muted-foreground">
                  В следующих итерациях здесь появится выбор друга и возможность отправлять треки,
                  плейлисты и приглашения в сессии.
                </p>
              </div>
            )}
          </Card>
        </div>
      </div>
    </>
  )
}
