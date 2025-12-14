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
} from '../lib/audio'
import { $user } from '@features/auth'
import { playbackStopped, playbackToggled, seekRequested } from '@features/player'

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
  fn: state => {
    console.log('[Session] Applying sync state:', state)
    return state
  },
  target: applySyncFx,
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
  fn: seconds => ({ action: 'seek' as const, payload: { position: seconds } }),
  target: sessionActionRequested,
})

sample({
  clock: sessionTrackSelected,
  fn: track => {
    console.log('[Session] Track selected, preparing action:', {
      trackId: track.trackId,
      title: track.title,
      artist: track.artist,
      cdnUrl: track.cdnUrl,
    })
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
