import { createEffect, createEvent, createStore, sample } from 'effector'
import type { RoomState } from '../api'
import { addParticipant, createRoom, getRoom, removeParticipant } from '../api'
import { SessionWebSocket, type PlayerAction } from '../lib/websocket'
import {
  applyPlaybackState,
  stopSessionAudio,
  startSessionPlayback,
  pauseSessionPlayback,
  getSessionAudio,
  setTrackEndedCallback,
  seekSessionAudio,
} from '../lib/audio'
import { $user } from '@features/auth'
import { playbackStopped, playbackToggled, seekRequested } from '@features/player'
import {
  queueContextSet,
  $queueTrackIds,
  loadQueueFx,
  removeTrackFromQueueFx,
  trackFromQueueSelectedForSession,
} from '@features/queue'
import { fetchTrackDetail } from '@entities/track/api'
import { fetchStreamMetadata } from '@features/player/api'

type SessionRole = 'host' | 'guest'

interface JoinParams {
  roomId: string
  createIfMissing?: boolean
}

const wsClient = typeof window !== 'undefined' ? new SessionWebSocket() : null

const connectionStatusChanged = createEvent<boolean>()
const syncStateReceived = createEvent<RoomState>()
const websocketErrorReceived = createEvent<string>()

if (wsClient) {
  wsClient.onConnectionChange(status => connectionStatusChanged(status))
  wsClient.onMessage('sync_state', message => {
    if (message.state) {
      console.log('[Session] Received sync_state:', message.state)

      // The state from WebSocket may come in snake_case (from Kotlin/Jackson)
      // or camelCase (from Go gateway), so we need to normalize it
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const rawState: any = message.state

      console.log('[Session] Current track:', rawState.currentTrack ?? rawState.current_track)
      if (rawState.currentTrack || rawState.current_track) {
        const track = rawState.currentTrack ?? rawState.current_track
        console.log('[Session] Track URL:', track.cdnUrl ?? track.cdn_url)
      }

      // Normalize state: convert snake_case to camelCase and handle missing fields
      // Jackson/Kotlin may serialize fields in snake_case, but frontend expects camelCase
      const normalizedState: RoomState = {
        roomId: rawState.roomId ?? rawState.room_id ?? '',
        currentTrack:
          rawState.currentTrack || rawState.current_track
            ? {
                trackId: rawState.currentTrack?.trackId ?? rawState.current_track?.track_id ?? '',
                title: rawState.currentTrack?.title ?? rawState.current_track?.title ?? '',
                artist: rawState.currentTrack?.artist ?? rawState.current_track?.artist ?? '',
                duration: rawState.currentTrack?.duration ?? rawState.current_track?.duration ?? 0,
                cdnUrl: rawState.currentTrack?.cdnUrl ?? rawState.current_track?.cdn_url ?? '',
              }
            : null,
        position: rawState.position ?? 0,
        isPlaying: rawState.isPlaying ?? rawState.is_playing ?? rawState.playing ?? false,
        participants: (rawState.participants ?? []).map(
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          (p: any) => ({
            userId: p.userId ?? p.user_id ?? '',
            username: p.username ?? '',
            isOnline: p.isOnline ?? p.is_online ?? true,
            joinedAt: p.joinedAt ?? p.joined_at ?? new Date().toISOString(),
          })
        ),
        queue: (rawState.queue ?? []).map(
          // eslint-disable-next-line @typescript-eslint/no-explicit-any
          (t: any) => ({
            trackId: t.trackId ?? t.track_id ?? '',
            title: t.title ?? '',
            artist: t.artist ?? '',
            duration: t.duration ?? 0,
            cdnUrl: t.cdnUrl ?? t.cdn_url ?? '',
          })
        ),
        lastAction: rawState.lastAction ?? rawState.last_action ?? null,
        createdAt: rawState.createdAt ?? rawState.created_at ?? new Date().toISOString(),
        updatedAt: rawState.updatedAt ?? rawState.updated_at ?? new Date().toISOString(),
      }

      console.log('[Session] Normalized state:', normalizedState)
      console.log('[Session] isPlaying:', normalizedState.isPlaying)
      syncStateReceived(normalizedState)
    }
  })
  wsClient.onMessage('error', message => {
    if (message.error) {
      console.error('[Session] WebSocket error:', message.error)
      websocketErrorReceived(message.error)
    }
  })
}

export const sessionRoomJoinRequested = createEvent<JoinParams>()
export const sessionRoomLeaveRequested = createEvent()
export const sessionPlayTriggered = createEvent()
export const sessionPauseTriggered = createEvent()
export const sessionSeekTriggered = createEvent<number>()
export const sessionSkipTrackTriggered = createEvent()
export const sessionTrackSelected = createEvent<{
  trackId: string
  title: string
  artist: string
  duration: number
  cdnUrl: string
}>()

const connectToRoomFx = createEffect(
  async ({
    roomId,
    createIfMissing,
    userId,
    username,
  }: {
    roomId: string
    createIfMissing?: boolean
    userId: string
    username: string
  }) => {
    if (!wsClient) {
      throw new Error('WebSocket недоступен в текущем окружении')
    }

    let room = await getRoom(roomId)
    if (!room && createIfMissing) {
      room = await createRoom(roomId)
    }

    if (!room) {
      throw new Error('Комната не найдена')
    }

    const updatedRoom = await addParticipant(roomId, userId, username)
    await wsClient.connect(roomId, userId)

    return {
      room: updatedRoom,
      roomId,
      role: createIfMissing ? ('host' as const) : ('guest' as const),
    }
  }
)

const leaveRoomFx = createEffect(async ({ roomId, userId }: { roomId: string; userId: string }) => {
  if (roomId && userId) {
    try {
      await removeParticipant(roomId, userId)
    } catch (error) {
      console.warn('Failed to remove participant', error)
    }
  }

  if (wsClient) {
    wsClient.disconnect()
  }
  stopSessionAudio()
})

const applySyncFx = createEffect(async (state: RoomState) => {
  // Stop regular player when session sync is active
  playbackStopped()
  try {
    await applyPlaybackState(state)
  } catch (error) {
    console.error('[Session] Failed to apply playback state:', error)
    // Don't throw - just log the error and continue
  }
  return state
})

type SessionActionPayload = {
  action: PlayerAction
  payload?: Record<string, unknown>
}

const sessionActionRequested = createEvent<SessionActionPayload>()

const sendPlayerActionFx = createEffect((payload: SessionActionPayload) => {
  if (!wsClient) {
    throw new Error('WebSocket недоступен')
  }
  console.log('[Session] Sending player action:', payload)
  wsClient.sendAction(payload.action, payload.payload)
  console.log('[Session] Player action sent successfully')
})

// Load roomId from localStorage on init
const loadRoomIdFromStorage = () => {
  if (typeof window === 'undefined') return null
  return localStorage.getItem('session_room_id')
}

// Save roomId to localStorage
const saveRoomIdToStorage = (roomId: string | null) => {
  if (typeof window === 'undefined') return
  if (roomId) {
    localStorage.setItem('session_room_id', roomId)
  } else {
    localStorage.removeItem('session_room_id')
  }
}

export const $sessionRoomId = createStore<string | null>(loadRoomIdFromStorage())
  .on(connectToRoomFx.doneData, (_, payload) => {
    saveRoomIdToStorage(payload.roomId)
    return payload.roomId
  })
  .on(sessionRoomLeaveRequested, () => {
    saveRoomIdToStorage(null)
    return null
  })
  .reset(connectToRoomFx.fail)

export const $sessionRole = createStore<SessionRole | null>(null)
  .on(connectToRoomFx.doneData, (_, payload) => payload.role)
  .reset([sessionRoomLeaveRequested, connectToRoomFx.fail])

export const $sessionState = createStore<RoomState | null>(null)
  .on(connectToRoomFx.doneData, (_, payload) => payload.room)
  .on(syncStateReceived, (_, state) => state)
  .reset([sessionRoomLeaveRequested, connectToRoomFx.fail])

export const $sessionConnected = createStore(false)
  .on(connectionStatusChanged, (_, status) => status)
  .reset(sessionRoomLeaveRequested)

export const $sessionError = createStore<string | null>(null)
  .on(connectToRoomFx.failData, (_, error) =>
    error instanceof Error ? error.message : 'Не удалось подключиться к комнате'
  )
  .on(websocketErrorReceived, (_, message) => message)
  .reset([connectToRoomFx.done, sessionRoomLeaveRequested])

export const $sessionShareLink = createStore<string | null>(null)
  .on($sessionRoomId, (_, roomId) => {
    if (!roomId || typeof window === 'undefined') return null
    return `${window.location.origin}/listen/${roomId}`
  })
  .reset(sessionRoomLeaveRequested)

export const $sessionHasRoom = $sessionRoomId.map(Boolean)

// Short polling for queue updates in session
const queuePollInterval = 2000 // 2 seconds
let queuePollIntervalId: ReturnType<typeof setInterval> | null = null

const startQueuePolling = () => {
  if (queuePollIntervalId) {
    return // Already polling
  }

  const pollQueue = () => {
    const roomId = $sessionRoomId.getState()

    // Only poll if we're in a session
    if (roomId) {
      loadQueueFx({ type: 'session' as const, roomId })
    }
  }

  // Poll immediately
  pollQueue()

  // Then poll periodically
  queuePollIntervalId = setInterval(pollQueue, queuePollInterval)
  console.log('[Session] Started queue polling')
}

const stopQueuePolling = () => {
  if (queuePollIntervalId) {
    clearInterval(queuePollIntervalId)
    queuePollIntervalId = null
    console.log('[Session] Stopped queue polling')
  }
}

// Start polling when connected to a room
if (typeof window !== 'undefined') {
  // Watch for connection changes
  $sessionConnected.watch(connected => {
    const roomId = $sessionRoomId.getState()
    if (connected && roomId) {
      startQueuePolling()
    } else {
      stopQueuePolling()
    }
  })

  // Also watch for room ID changes
  $sessionRoomId.watch(roomId => {
    const connected = $sessionConnected.getState()
    if (connected && roomId) {
      startQueuePolling()
    } else {
      stopQueuePolling()
    }
  })

  // Stop polling when leaving a room
  sessionRoomLeaveRequested.watch(() => {
    stopQueuePolling()
  })
}

// Also reload queue when sync_state is received (in case queue changed on server)
sample({
  clock: syncStateReceived,
  source: $sessionRoomId,
  filter: Boolean,
  fn: roomId => ({
    type: 'session' as const,
    roomId: roomId!,
  }),
  target: loadQueueFx,
})

sample({
  clock: sessionRoomJoinRequested,
  source: $user,
  filter: user => Boolean(user), // Only proceed if user is loaded
  fn: (user, params) => {
    return {
      roomId: params.roomId.trim(),
      createIfMissing: params.createIfMissing,
      userId: user!.id,
      username: user!.username,
    }
  },
  target: connectToRoomFx,
})

// Stop regular player when joining a room - session will handle playback
sample({
  clock: connectToRoomFx.done,
  target: playbackStopped,
})

// Set queue context to session when joining a room
sample({
  clock: connectToRoomFx.doneData,
  fn: ({ roomId }) => ({
    type: 'session' as const,
    roomId,
  }),
  target: queueContextSet,
})

// Load queue when joining a room
sample({
  clock: connectToRoomFx.doneData,
  fn: ({ roomId }) => ({
    type: 'session' as const,
    roomId,
  }),
  target: loadQueueFx,
})

// Event for track ending in session
const sessionTrackEnded = createEvent()

// Set up track ended callback
if (typeof window !== 'undefined') {
  setTrackEndedCallback(() => {
    sessionTrackEnded()
  })
}

// Effect to load track by ID and get stream metadata
const loadTrackByIdFx = createEffect(async (trackId: string) => {
  const track = await fetchTrackDetail(trackId)
  // Fetch stream metadata to get CDN URL
  if (track.artist?.id) {
    try {
      const parseBitrates = (qualities?: string[]): number[] | undefined => {
        if (!qualities) return undefined
        const values = Array.from(
          new Set(
            qualities
              .map(item => {
                const digits = item.replace(/\D/g, '')
                if (!digits) return null
                const numeric = Number(digits)
                if (!numeric) return null
                return numeric < 1000 ? numeric * 1000 : numeric
              })
              .filter((value): value is number => Boolean(value))
          )
        )
        return values.length ? values : undefined
      }

      // Type assertion for track.stream since TrackDetail extends Track
      const streamInfo = track.stream as { qualities?: string[] } | undefined
      const streamQualities = streamInfo?.qualities
      const bitrates = parseBitrates(streamQualities)
      const streamMetadata = await fetchStreamMetadata({
        trackId: track.id,
        artistId: track.artist.id,
        bitrates,
      })
      return { track, cdnUrl: streamMetadata.masterUrl }
    } catch (error) {
      console.error('[Session] Failed to fetch stream metadata:', error)
      return { track, cdnUrl: '' }
    }
  }
  return { track, cdnUrl: '' }
})

// Effect to remove tracks until selected one in session
const removeTracksUntilSessionFx = createEffect(
  async ({ trackIds, roomId }: { trackIds: string[]; roomId: string }) => {
    const context = { type: 'session' as const, roomId }
    await Promise.all(trackIds.map(trackId => removeTrackFromQueueFx({ trackId, context })))
    return { trackIds, roomId }
  }
)

// Flag to track if we should auto-play next track after queue reload
const $shouldAutoPlayNext = createStore(false)
  .on(sessionTrackEnded, () => true)
  .on(sessionSkipTrackTriggered, () => true)
  .reset([sessionTrackSelected])

// When track ends, reload queue (track will be removed after next track starts)
sample({
  clock: sessionTrackEnded,
  source: { sessionState: $sessionState, roomId: $sessionRoomId },
  filter: ({ sessionState, roomId }) => Boolean(sessionState?.currentTrack) && Boolean(roomId),
  fn: ({ roomId }) => ({
    type: 'session' as const,
    roomId: roomId!,
  }),
  target: loadQueueFx,
})

// When skip track is triggered, first remove current track from queue, then reload queue
sample({
  clock: sessionSkipTrackTriggered,
  source: { sessionState: $sessionState, roomId: $sessionRoomId },
  filter: ({ sessionState, roomId }) => Boolean(sessionState?.currentTrack) && Boolean(roomId),
  fn: ({ sessionState, roomId }) => ({
    trackId: sessionState!.currentTrack!.trackId,
    context: { type: 'session' as const, roomId: roomId! },
  }),
  target: removeTrackFromQueueFx,
})

// After removing skipped track, reload queue
sample({
  clock: removeTrackFromQueueFx.done,
  source: { roomId: $sessionRoomId, shouldAutoPlay: $shouldAutoPlayNext },
  filter: ({ roomId, shouldAutoPlay }) => Boolean(roomId) && shouldAutoPlay,
  fn: ({ roomId }) => ({
    type: 'session' as const,
    roomId: roomId!,
  }),
  target: loadQueueFx,
})

// After queue reload (after track ended/skipped), play first track if available
// BUT only if we should auto-play (track ended or skipped)
sample({
  clock: loadQueueFx.doneData,
  source: {
    queue: $queueTrackIds,
    sessionState: $sessionState,
    shouldAutoPlay: $shouldAutoPlayNext,
  },
  filter: ({ queue, sessionState, shouldAutoPlay }) => {
    // Filter out current track from queue if it's still there
    const currentTrackId = sessionState?.currentTrack?.trackId
    const filteredQueue = currentTrackId ? queue.filter(id => id !== currentTrackId) : queue
    return filteredQueue.length > 0 && Boolean(sessionState) && shouldAutoPlay
  },
  fn: ({ queue, sessionState }) => {
    // Get first track that is not the current one
    const currentTrackId = sessionState?.currentTrack?.trackId
    const filteredQueue = currentTrackId ? queue.filter(id => id !== currentTrackId) : queue
    return filteredQueue[0] || queue[0]
  },
  target: loadTrackByIdFx,
})

// After loading track from queue, select it to play
sample({
  clock: loadTrackByIdFx.doneData,
  fn: ({ track, cdnUrl }) => ({
    trackId: track.id,
    title: track.title,
    artist: track.artist.name,
    duration: track.duration ?? 0,
    cdnUrl,
  }),
  target: sessionTrackSelected,
})

// After track is selected from queue (auto-play scenario), remove it from queue
// This happens when track ends or is skipped, and next track is taken from queue
sample({
  clock: sessionTrackSelected,
  source: { queue: $queueTrackIds, roomId: $sessionRoomId, shouldAutoPlay: $shouldAutoPlayNext },
  filter: ({ queue, roomId, shouldAutoPlay }, track) => {
    // Remove track from queue if:
    // 1. It's an auto-play scenario (track ended or skipped)
    // 2. Track is in the queue
    // 3. We're in a session
    return shouldAutoPlay && queue.includes(track.trackId) && Boolean(roomId)
  },
  fn: ({ roomId }, track) => ({
    trackId: track.trackId,
    context: { type: 'session' as const, roomId: roomId! },
  }),
  target: removeTrackFromQueueFx,
})

// Reload queue after removing track (but don't auto-play unless it was ended/skipped)
// This is handled by the flag $shouldAutoPlayNext
sample({
  clock: removeTrackFromQueueFx.done,
  source: { roomId: $sessionRoomId, shouldAutoPlay: $shouldAutoPlayNext },
  filter: ({ roomId, shouldAutoPlay }) => Boolean(roomId) && !shouldAutoPlay,
  fn: ({ roomId }) => ({
    type: 'session' as const,
    roomId: roomId!,
  }),
  target: loadQueueFx,
})

// When track is selected from queue in session, remove tracks before it and play it
sample({
  clock: trackFromQueueSelectedForSession,
  source: { queue: $queueTrackIds, roomId: $sessionRoomId },
  filter: ({ queue, roomId }, trackId: string) => queue.includes(trackId) && Boolean(roomId),
  fn: ({ queue, roomId }, trackId: string) => {
    const trackIndex = queue.findIndex((id: string) => id === trackId)
    const tracksToRemove = queue.slice(0, trackIndex + 1)
    return { trackIds: tracksToRemove, roomId: roomId! }
  },
  target: removeTracksUntilSessionFx,
})

// After removing tracks, load the selected track and play it
sample({
  clock: removeTracksUntilSessionFx.doneData,
  fn: ({ trackIds }) => trackIds[trackIds.length - 1], // Get the last track (the selected one)
  target: loadTrackByIdFx,
})

// Reload queue after removing tracks for selected track
sample({
  clock: removeTracksUntilSessionFx.done,
  source: $sessionRoomId,
  filter: Boolean,
  fn: roomId => ({
    type: 'session' as const,
    roomId: roomId!,
  }),
  target: loadQueueFx,
})

// When track is added to queue in session, DO NOT auto-play it
// Tracks should only start playing when:
// 1. User explicitly selects a track from queue
// 2. Track ends and next track is taken from queue
// 3. User clicks play on a track

// Reset queue context when leaving a room
sample({
  clock: sessionRoomLeaveRequested,
  source: $user,
  filter: Boolean,
  fn: user => ({
    type: 'user' as const,
    userId: user.id,
  }),
  target: queueContextSet,
})

sample({
  clock: sessionRoomLeaveRequested,
  source: { roomId: $sessionRoomId, user: $user },
  filter: ({ roomId, user }) => Boolean(roomId && user),
  fn: ({ roomId, user }) => ({
    roomId: roomId as string,
    userId: user!.id,
  }),
  target: leaveRoomFx,
})

sample({
  clock: syncStateReceived,
  filter: state => {
    // Only apply playback state if there's a track to play
    // State itself should always be updated in $sessionState store
    return Boolean(state.currentTrack && state.currentTrack.cdnUrl)
  },
  fn: state => {
    console.log('[Session] Applying sync state:', state)
    return state
  },
  target: applySyncFx,
})

// When sync_state arrives with currentTrack = null, stop playback
sample({
  clock: syncStateReceived,
  filter: state => !state.currentTrack || !state.currentTrack.cdnUrl,
  fn: () => {
    console.log('[Session] Sync state received with no track, stopping playback')
    stopSessionAudio()
  },
})

// When connecting to room, apply current state to restore playback (only if track exists)
sample({
  clock: connectToRoomFx.doneData,
  filter: payload => Boolean(payload.room?.currentTrack),
  fn: payload => {
    console.log('[Session] Room connected, applying initial state:', payload.room)
    return payload.room
  },
  target: applySyncFx,
})

// When connecting to room without a track, stop any existing playback
sample({
  clock: connectToRoomFx.doneData,
  filter: payload => !payload.room?.currentTrack,
  fn: () => {
    console.log('[Session] Room connected without track, stopping playback')
    stopSessionAudio()
  },
})

// When user clicks Play button, start playback locally and send action to server
const startLocalPlaybackFx = createEffect(async () => {
  try {
    await startSessionPlayback()
  } catch (error) {
    console.error('[Session] Failed to start local playback:', error)
    // Continue anyway - we still want to send action to server
  }
})

const pauseLocalPlaybackFx = createEffect(() => {
  pauseSessionPlayback()
})

sample({
  clock: sessionPlayTriggered,
  target: [
    startLocalPlaybackFx,
    sessionActionRequested.prepend(() => ({ action: 'play' as const })),
  ],
})

sample({
  clock: sessionPauseTriggered,
  fn: () => {
    // Get current position before pausing
    const audio = getSessionAudio()
    const currentPosition = audio?.currentTime ?? 0
    return {
      action: 'pause' as const,
      payload: { position: currentPosition },
    }
  },
  target: [pauseLocalPlaybackFx, sessionActionRequested],
})

sample({
  clock: sessionSeekTriggered,
  fn: seconds => {
    // Apply seek locally first
    seekSessionAudio(seconds)
    return { action: 'seek' as const, payload: { position: seconds } }
  },
  target: sessionActionRequested,
})

// When track is selected, apply it locally first, then send to server
const applyTrackLocallyFx = createEffect(
  async (track: {
    trackId: string
    title: string
    artist: string
    duration: number
    cdnUrl: string
  }) => {
    // Apply track state locally to start playback immediately
    const roomState: RoomState = {
      roomId: $sessionRoomId.getState() || '',
      currentTrack: {
        trackId: track.trackId,
        title: track.title,
        artist: track.artist,
        duration: track.duration,
        cdnUrl: track.cdnUrl,
      },
      position: 0,
      isPlaying: true, // Start playing immediately
      participants: [],
      queue: [],
      lastAction: null,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }
    await applyPlaybackState(roomState)
    return track
  }
)

sample({
  clock: sessionTrackSelected,
  filter: track => {
    // Don't send change_track if cdnUrl is empty - track won't play
    if (!track.cdnUrl || track.cdnUrl.trim() === '') {
      console.warn('[Session] Cannot select track without CDN URL:', track.trackId)
      return false
    }
    return true
  },
  target: applyTrackLocallyFx,
})

sample({
  clock: applyTrackLocallyFx.doneData,
  fn: track => {
    console.log('[Session] Track selected, preparing action:', {
      trackId: track.trackId,
      title: track.title,
      artist: track.artist,
      cdnUrl: track.cdnUrl,
    })
    // Server automatically sets isPlaying = true when change_track is received
    return {
      action: 'change_track' as const,
      payload: {
        track_id: track.trackId,
        title: track.title,
        artist: track.artist,
        duration: track.duration,
        cdn_url: track.cdnUrl,
      },
    }
  },
  target: sessionActionRequested,
})

sample({
  clock: sessionActionRequested,
  source: $sessionConnected,
  filter: connected => connected,
  fn: (_, payload) => payload,
  target: sendPlayerActionFx,
})

// When user toggles playback and we're in a session, send action to room
// Don't use local player state - session controls it
sample({
  clock: playbackToggled,
  source: { connected: $sessionConnected, sessionState: $sessionState },
  filter: ({ connected }) => connected,
  fn: ({ sessionState }) => {
    // Use session state, not player state
    const shouldPlay = !sessionState?.isPlaying
    return { action: (shouldPlay ? 'play' : 'pause') as PlayerAction }
  },
  target: sessionActionRequested,
})

sample({
  clock: seekRequested,
  source: $sessionConnected,
  filter: connected => connected,
  fn: (_, seconds) => ({
    action: 'seek' as const,
    payload: { position: seconds },
  }),
  target: sessionActionRequested,
})

// Don't auto-send tracks from global player to session
// Tracks should be selected directly in the room
