import { useEffect, useMemo, useRef, useState } from 'react'
import { useUnit } from 'effector-react'
import {
  $sessionRoomId,
  $sessionState,
  $sessionConnected,
  $sessionError,
  $sessionShareLink,
  $sessionHasRoom,
  sessionRoomJoinRequested,
  sessionRoomLeaveRequested,
  sessionPlayTriggered,
  sessionPauseTriggered,
  sessionSeekTriggered,
  sessionTrackSelected,
} from '@features/session'
import { $isAuthenticated, $user } from '@features/auth'
import { routes } from '@shared/router'
import { searchTracks, fetchTrackDetail } from '@entities/track/api'
import type { Track } from '@entities/track'
import type { TrackDetail } from '@widgets/track'
import { fetchStreamMetadata } from '@features/player/api'
import { TrackHero } from '@widgets/track'
import { getSessionAudio } from '@features/session/lib/audio'

const generateRoomId = () => `room-${Math.random().toString(36).slice(2, 8)}`

const DEFAULT_COVER_URL =
  'https://mir-s3-cdn-cf.behance.net/projects/202/e2ba0e187042211.Y3JvcCw4MDgsNjMyLDAsMA.png'

export const SessionPage = () => {
  const [manualRoomId, setManualRoomId] = useState('')
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<Track[]>([])
  const [isSearching, setIsSearching] = useState(false)
  const [trackDetail, setTrackDetail] = useState<TrackDetail | null>(null)
  const lastRequestedRoomRef = useRef<string | null>(null)

  const {
    roomParams,
    roomRouteOpened,
    sessionRouteOpened,
    navigateToRoom,
    navigateToSessionRoot,
    isAuthenticated,
    user,
    roomId,
    sessionState,
    connected,
    shareLink,
    error,
    hasRoom,
    joinRoom,
    leaveRoom,
    playAction,
    pauseAction,
    seekAction,
    selectTrackAction,
  } = useUnit({
    roomParams: routes.sessionRoom.$params,
    roomRouteOpened: routes.sessionRoom.$isOpened,
    sessionRouteOpened: routes.session.$isOpened,
    navigateToRoom: routes.sessionRoom.navigate,
    navigateToSessionRoot: routes.session.navigate,
    isAuthenticated: $isAuthenticated,
    user: $user,
    roomId: $sessionRoomId,
    sessionState: $sessionState,
    connected: $sessionConnected,
    shareLink: $sessionShareLink,
    error: $sessionError,
    hasRoom: $sessionHasRoom,
    joinRoom: sessionRoomJoinRequested,
    leaveRoom: sessionRoomLeaveRequested,
    playAction: sessionPlayTriggered,
    pauseAction: sessionPauseTriggered,
    seekAction: sessionSeekTriggered,
    selectTrackAction: sessionTrackSelected,
  })

  const showLobby = sessionRouteOpened && !roomRouteOpened
  const showRoom = roomRouteOpened

  // Auto-connect to saved room on mount
  useEffect(() => {
    if (!isAuthenticated || !user) {
      return
    }

    // Wait for user to be loaded before joining room
    if (!roomId || connected) {
      return
    }

    if (lastRequestedRoomRef.current === roomId) {
      return
    }
    lastRequestedRoomRef.current = roomId
    joinRoom({ roomId })
  }, [isAuthenticated, user, roomId, connected, joinRoom])

  // Handle route changes
  useEffect(() => {
    if (!isAuthenticated) {
      return
    }

    // If route has roomId, use it
    if (roomRouteOpened && roomParams?.roomId) {
      if (roomParams.roomId === roomId || lastRequestedRoomRef.current === roomParams.roomId) {
        return
      }
      lastRequestedRoomRef.current = roomParams.roomId
      joinRoom({ roomId: roomParams.roomId })
      return
    }

    // If route opened without roomId but we have a saved roomId, navigate to it
    if (roomRouteOpened && !roomParams?.roomId && roomId) {
      navigateToRoom({ params: { roomId }, query: {} })
      return
    }
  }, [roomRouteOpened, roomParams?.roomId, isAuthenticated, roomId, joinRoom, navigateToRoom])

  // Don't disconnect on unmount - keep connection alive in background
  // Only disconnect when explicitly leaving via handleLeaveRoom

  useEffect(() => {
    if (roomId === null) {
      lastRequestedRoomRef.current = null
    }
  }, [roomId])

  const handleCreateRoom = () => {
    if (!isAuthenticated) {
      navigateToSessionRoot({ params: {}, query: {} })
      return
    }
    const newRoomId = generateRoomId()
    lastRequestedRoomRef.current = newRoomId
    joinRoom({ roomId: newRoomId, createIfMissing: true })
    navigateToRoom({ params: { roomId: newRoomId }, query: {} })
  }

  const handleJoinRoom = (event: React.FormEvent) => {
    event.preventDefault()
    if (!manualRoomId.trim()) return
    navigateToRoom({ params: { roomId: manualRoomId.trim() }, query: {} })
  }

  const handleCopyLink = async () => {
    if (!shareLink) return
    try {
      await navigator.clipboard.writeText(shareLink)
    } catch (copyError) {
      console.warn('Failed to copy link', copyError)
    }
  }

  const handleLeaveRoom = () => {
    navigateToSessionRoot({ params: {}, query: {} })
    leaveRoom()
  }

  // Filter out offline participants (those who left the room)
  const participants = (sessionState?.participants ?? []).filter(p => p.isOnline)
  const currentTrack = sessionState?.currentTrack ?? null

  // Load full track detail when current track changes
  useEffect(() => {
    if (currentTrack?.trackId) {
      fetchTrackDetail(currentTrack.trackId)
        .then(detail => setTrackDetail(detail))
        .catch(error => {
          console.warn('Failed to fetch track detail, using minimal info:', error)
          // Fallback to minimal TrackDetail
          setTrackDetail({
            id: currentTrack.trackId,
            title: currentTrack.title,
            artist: {
              id: '',
              name: currentTrack.artist,
            },
            album: {
              id: `album-${currentTrack.trackId}`,
              title: 'Неизвестный альбом',
            },
            coverUrl: DEFAULT_COVER_URL,
            duration: currentTrack.duration,
            liked: false,
          })
        })
    } else {
      setTrackDetail(null)
    }
    // Only reload when trackId changes, not when other track properties change
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentTrack?.trackId])

  // Get current playback time from audio element
  const [currentTime, setCurrentTime] = useState(0)
  useEffect(() => {
    if (!currentTrack) return

    const audio = getSessionAudio()
    if (!audio) return

    const updateTime = () => setCurrentTime(audio.currentTime)
    audio.addEventListener('timeupdate', updateTime)

    return () => {
      audio.removeEventListener('timeupdate', updateTime)
    }
  }, [currentTrack])

  const connectionLabel = connected
    ? 'Подключено'
    : hasRoom && !connected
      ? 'Переподключение...'
      : 'Отключено'

  const connectionClass = connected
    ? 'text-emerald-400'
    : hasRoom
      ? 'text-amber-400'
      : 'text-muted-foreground'

  const shareMessage = useMemo(() => {
    if (!shareLink || !roomId) return 'Создайте комнату, чтобы поделиться ссылкой'
    return shareLink
  }, [shareLink, roomId])

  if (!isAuthenticated) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center text-center">
        <div className="space-y-4 rounded-3xl border border-border/60 bg-secondary/20 px-8 py-10">
          <p className="text-lg font-semibold text-foreground">
            Авторизуйтесь, чтобы создать комнату совместного прослушивания
          </p>
          <button
            type="button"
            onClick={() => routes.auth.navigate({ params: {}, query: {} })}
            className="rounded-full bg-primary px-5 py-2 text-sm font-medium text-primary-foreground transition hover:opacity-90"
          >
            Войти
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="page-container space-y-10 pb-16 pt-10">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <p className="text-sm uppercase tracking-[0.3em] text-muted-foreground">
            Совместное прослушивание
          </p>
          <h1 className="text-3xl font-semibold text-foreground">Музыкальная комната</h1>
          <p className="text-muted-foreground">
            Создайте комнату, отправьте ссылку друзьям и управляйте воспроизведением вместе.
          </p>
        </div>
        {roomId && (
          <div className="flex items-center gap-3 rounded-2xl border border-border/60 bg-secondary/30 px-4 py-3">
            <span className={`text-sm font-medium ${connectionClass}`}>{connectionLabel}</span>
            <button
              type="button"
              onClick={handleLeaveRoom}
              className="rounded-full border border-border/60 px-3 py-1 text-xs font-medium text-muted-foreground transition hover:border-red-400 hover:text-red-400"
            >
              Выйти
            </button>
          </div>
        )}
      </div>

      {showLobby && (
        <div className="grid gap-6 lg:grid-cols-2">
          <div className="rounded-3xl border border-border/60 bg-secondary/20 p-6 sm:p-8 space-y-6">
            <div>
              <h2 className="text-xl font-semibold">Создать комнату</h2>
              <p className="text-sm text-muted-foreground">
                Мы сгенерируем уникальный идентификатор и сразу подключим вас как хозяина комнаты.
              </p>
            </div>
            <button
              type="button"
              onClick={handleCreateRoom}
              className="w-full rounded-full bg-primary px-4 py-3 text-sm font-semibold text-primary-foreground transition hover:opacity-90"
            >
              Создать комнату
            </button>
          </div>

          <div className="rounded-3xl border border-border/60 bg-secondary/20 p-6 sm:p-8 space-y-6">
            <div>
              <h2 className="text-xl font-semibold">Присоединиться по ссылке</h2>
              <p className="text-sm text-muted-foreground">
                Введите идентификатор комнаты, чтобы присоединиться к друзьям.
              </p>
            </div>
            <form onSubmit={handleJoinRoom} className="space-y-4">
              <input
                value={manualRoomId}
                onChange={event => setManualRoomId(event.target.value)}
                placeholder="room-abc123"
                className="w-full rounded-2xl border border-border/60 bg-background px-4 py-3 text-sm"
              />
              <button
                type="submit"
                className="w-full rounded-full border border-border/60 px-4 py-3 text-sm font-semibold text-foreground transition hover:border-primary hover:text-primary"
              >
                Присоединиться
              </button>
            </form>
          </div>
        </div>
      )}

      {showRoom && (
        <div className="grid gap-6 xl:grid-cols-[2fr,1fr]">
          <div className="space-y-6">
            <div className="rounded-3xl border border-border/60 bg-secondary/20 p-6 sm:p-8 space-y-4">
              <div className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Идентификатор комнаты</p>
                  <p className="text-2xl font-semibold text-foreground">{roomId}</p>
                </div>
                <div className="flex flex-col gap-2 md:items-end">
                  <p className="text-xs uppercase tracking-[0.3em] text-muted-foreground">Ссылка</p>
                  <div className="flex flex-col gap-2 md:flex-row">
                    <p className="text-sm text-muted-foreground break-all">{shareMessage}</p>
                    {shareLink && (
                      <button
                        type="button"
                        onClick={handleCopyLink}
                        className="rounded-full border border-border/60 px-3 py-1 text-xs font-medium text-muted-foreground transition hover:border-primary hover:text-primary"
                      >
                        Скопировать
                      </button>
                    )}
                  </div>
                </div>
              </div>
              {error && (
                <p className="text-sm text-red-400">
                  {error}. Попробуйте переподключиться или обновите страницу.
                </p>
              )}
            </div>

            {trackDetail ? (
              <TrackHero
                track={trackDetail}
                isPlaying={sessionState?.isPlaying ?? false}
                onTogglePlay={() => {
                  if (sessionState?.isPlaying) {
                    pauseAction()
                  } else {
                    playAction()
                  }
                }}
                onToggleLike={() => console.info('Toggle like for track', trackDetail.id)}
                onShare={() => console.info('Share track', trackDetail.id)}
                onAddToPlaylist={() => console.info('Add track to playlist', trackDetail.id)}
                onGoToArtist={() => console.info('Go to artist', trackDetail.artist.id)}
                onGoToAlbum={() => console.info('Go to album', trackDetail.album.id)}
                currentTime={currentTime}
                duration={trackDetail.duration ?? 0}
                isBuffering={false}
                isSeekEnabled={connected}
                onSeek={seekAction}
              />
            ) : (
              <div className="rounded-3xl border border-border/60 bg-secondary/20 p-6 sm:p-8 space-y-5">
                <p className="text-muted-foreground">Комната ожидает выбора трека.</p>
              </div>
            )}
          </div>

          <div className="space-y-6">
            <div className="rounded-3xl border border-border/60 bg-secondary/20 p-6 space-y-4">
              <div>
                <h3 className="text-lg font-semibold">Участники</h3>
                <p className="text-sm text-muted-foreground">
                  В комнате {participants.length}{' '}
                  {participants.length === 1 ? 'человек' : 'человека'}.
                </p>
              </div>
              <ul className="space-y-2">
                {participants.map(participant => (
                  <li
                    key={participant.userId}
                    className="flex items-center justify-between rounded-2xl border border-border/60 px-3 py-2"
                  >
                    <div>
                      <p className="text-sm font-medium">{participant.username}</p>
                      <p className="text-xs text-muted-foreground">
                        С {new Date(participant.joinedAt).toLocaleTimeString()}
                      </p>
                    </div>
                    <span
                      className={`text-xs ${
                        participant.isOnline ? 'text-emerald-400' : 'text-muted-foreground'
                      }`}
                    >
                      {participant.isOnline ? 'online' : 'offline'}
                    </span>
                  </li>
                ))}
                {participants.length === 0 && (
                  <li className="rounded-2xl border border-dashed border-border/60 px-3 py-4 text-center text-sm text-muted-foreground">
                    Пока никого нет. Отправьте ссылку друзьям!
                  </li>
                )}
              </ul>
            </div>

            <div className="rounded-3xl border border-border/60 bg-secondary/20 p-6 space-y-4">
              <div>
                <h3 className="text-lg font-semibold">Выбор трека</h3>
                <p className="text-sm text-muted-foreground">
                  Найдите трек и нажмите "Включить" для всех участников
                </p>
              </div>
              <div className="space-y-3">
                <input
                  type="text"
                  value={searchQuery}
                  onChange={e => setSearchQuery(e.target.value)}
                  onKeyDown={async e => {
                    if (e.key === 'Enter' && searchQuery.trim()) {
                      e.preventDefault()
                      setIsSearching(true)
                      try {
                        const tracks = await searchTracks(searchQuery.trim(), 10)
                        setSearchResults(tracks)
                      } catch (error) {
                        console.error('Failed to search tracks', error)
                        setSearchResults([])
                      } finally {
                        setIsSearching(false)
                      }
                    }
                  }}
                  placeholder="Поиск треков..."
                  className="w-full rounded-2xl border border-border/60 bg-background px-4 py-3 text-sm"
                  disabled={!connected}
                />
                {isSearching && <p className="text-sm text-muted-foreground">Поиск...</p>}
                {searchResults.length > 0 && (
                  <ul className="space-y-2 max-h-96 overflow-y-auto">
                    {searchResults.map(track => (
                      <li
                        key={track.id}
                        className="flex items-center justify-between rounded-2xl border border-border/60 px-3 py-2"
                      >
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium truncate">{track.title}</p>
                          <p className="text-xs text-muted-foreground truncate">
                            {track.artist.name}
                          </p>
                        </div>
                        <button
                          type="button"
                          onClick={async () => {
                            if (!connected) return
                            try {
                              const streamMetadata = await fetchStreamMetadata({
                                trackId: track.id,
                                artistId: track.artist.id,
                              })
                              selectTrackAction({
                                trackId: track.id,
                                title: track.title,
                                artist: track.artist.name,
                                duration: track.duration ?? 0,
                                cdnUrl: streamMetadata.masterUrl,
                              })
                              setSearchQuery('')
                              setSearchResults([])
                            } catch (error) {
                              console.error('Failed to get stream URL', error)
                            }
                          }}
                          disabled={!connected}
                          className="rounded-full bg-primary/90 px-4 py-1.5 text-xs font-semibold text-primary-foreground transition hover:bg-primary disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                          Включить
                        </button>
                      </li>
                    ))}
                  </ul>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
