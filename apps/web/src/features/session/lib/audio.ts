import type { RoomState } from '../api'
import Hls from 'hls.js'

let audioElement: HTMLAudioElement | null = null
let hlsInstance: Hls | null = null
// Track the base URL of the current source (without query params) to avoid unnecessary reloads
let currentSourceBaseUrl: string | null = null
// Track the current track ID to detect track changes
let currentTrackId: string | null = null

// Track when user manually initiated playback to ignore stale sync states
let userInitiatedPlayback = false
let userInitiatedPlaybackTimeout: ReturnType<typeof setTimeout> | null = null

const ensureAudio = () => {
  if (audioElement) {
    return audioElement
  }

  if (typeof document === 'undefined') {
    audioElement = new Audio()
    return audioElement
  }

  audioElement = document.createElement('audio')
  audioElement.preload = 'auto'
  audioElement.crossOrigin = 'anonymous'
  audioElement.style.display = 'none'
  document.body.appendChild(audioElement)
  return audioElement
}

const isHlsUrl = (url: string) => {
  try {
    const urlObj = new URL(url)
    const pathname = urlObj.pathname
    return pathname.includes('.m3u8') || url.includes('.m3u8')
  } catch {
    return url.includes('.m3u8') || url.includes('/master.m3u8')
  }
}

const normalizeUrl = (url: string) => {
  try {
    const urlObj = new URL(url, window.location.href)
    return `${urlObj.protocol}//${urlObj.host}${urlObj.pathname}`
  } catch {
    return url.split('?')[0]
  }
}

const syncPosition = (audio: HTMLAudioElement, targetPosition: number) => {
  const drift = Math.abs(audio.currentTime - targetPosition)
  if (drift > 0.5) {
    console.log(
      `[Session Audio] Syncing position: ${audio.currentTime.toFixed(2)} -> ${targetPosition.toFixed(2)}`
    )
    audio.currentTime = targetPosition
  }
}

export const applyPlaybackState = async (state: RoomState) => {
  if (!state.currentTrack || !state.currentTrack.cdnUrl) {
    console.warn('[Session Audio] No track or URL to play - skipping playback')
    return
  }

  const audio = ensureAudio()
  const nextSource = state.currentTrack.cdnUrl
  const nextTrackId = state.currentTrack?.trackId || null

  // Check if user just initiated playback and state says paused - ignore stale sync
  const stateIsPaused = state.isPlaying === false
  if (stateIsPaused && userInitiatedPlayback) {
    console.log('[Session Audio] Ignoring stale pause state - user just started playback')
    if (typeof state.position === 'number') {
      syncPosition(audio, state.position)
    }
    if (audio.paused) {
      try {
        await audio.play()
      } catch (error) {
        console.error('[Session Audio] Failed to start playback:', error)
      }
    }
    return
  }

  // Check if source changed (track ID or URL changed)
  const normalizedNextSource = normalizeUrl(nextSource)
  const normalizedCurrentSrc =
    currentSourceBaseUrl ||
    (audio.src && !audio.src.startsWith('blob:') ? normalizeUrl(audio.src) : null)
  const urlChanged = normalizedCurrentSrc !== normalizedNextSource && normalizedNextSource !== null
  const trackChanged = currentTrackId !== nextTrackId && nextTrackId !== null
  const sourceChanged = urlChanged || trackChanged

  console.log('[Session Audio] Applying state:', {
    trackId: nextTrackId,
    isPlaying: state.isPlaying,
    position: state.position,
    sourceChanged,
  })

  // Reload source if it changed
  if (sourceChanged || !audio.src || !currentSourceBaseUrl || !currentTrackId) {
    console.log('[Session Audio] Loading new source')
    currentSourceBaseUrl = normalizedNextSource
    currentTrackId = nextTrackId

    // Clean up existing HLS instance
    if (hlsInstance) {
      hlsInstance.destroy()
      hlsInstance = null
    }

    const isHls = isHlsUrl(nextSource)

    if (isHls) {
      if (Hls.isSupported()) {
        hlsInstance = new Hls({
          enableWorker: true,
          lowLatencyMode: false,
          debug: false,
        })
        hlsInstance.loadSource(nextSource)
        hlsInstance.attachMedia(audio)

        await new Promise<void>((resolve, reject) => {
          if (!hlsInstance) {
            reject(new Error('HLS instance not created'))
            return
          }

          const timeout = setTimeout(() => {
            hlsInstance?.off(Hls.Events.MANIFEST_PARSED, onReady)
            hlsInstance?.off(Hls.Events.ERROR, onError)
            reject(new Error('HLS manifest loading timeout'))
          }, 30000)

          const onReady = () => {
            clearTimeout(timeout)
            hlsInstance?.off(Hls.Events.MANIFEST_PARSED, onReady)
            hlsInstance?.off(Hls.Events.ERROR, onError)
            resolve()
          }

          const onError = (_event: string, data: { details?: string; fatal?: boolean }) => {
            clearTimeout(timeout)
            hlsInstance?.off(Hls.Events.MANIFEST_PARSED, onReady)
            hlsInstance?.off(Hls.Events.ERROR, onError)
            reject(new Error(`HLS error: ${data?.details || 'Unknown error'}`))
          }

          hlsInstance.on(Hls.Events.MANIFEST_PARSED, onReady)
          hlsInstance.on(Hls.Events.ERROR, onError)
        })
      } else if (audio.canPlayType('application/vnd.apple.mpegurl')) {
        // Native HLS support (Safari)
        console.log('[Session Audio] Using native HLS support (Safari)')
        audio.src = nextSource
        audio.load() // Ensure Safari loads the source
        await new Promise<void>((resolve, reject) => {
          const timeout = setTimeout(() => {
            audio.removeEventListener('canplay', onCanPlay)
            audio.removeEventListener('error', onError)
            reject(new Error('Native HLS loading timeout'))
          }, 30000)

          const onCanPlay = () => {
            clearTimeout(timeout)
            audio.removeEventListener('canplay', onCanPlay)
            audio.removeEventListener('error', onError)
            console.log('[Session Audio] Native HLS ready (canplay event)')
            resolve()
          }
          const onError = () => {
            clearTimeout(timeout)
            audio.removeEventListener('canplay', onCanPlay)
            audio.removeEventListener('error', onError)
            console.error('[Session Audio] Native HLS loading error')
            reject(new Error('Failed to load HLS stream'))
          }
          if (audio.readyState >= HTMLMediaElement.HAVE_FUTURE_DATA) {
            clearTimeout(timeout)
            console.log('[Session Audio] Native HLS already ready')
            resolve()
          } else {
            audio.addEventListener('canplay', onCanPlay)
            audio.addEventListener('error', onError)
          }
        })
      } else {
        console.error('[Session Audio] HLS is not supported')
        return
      }
    } else {
      audio.src = nextSource
      audio.load()
      await new Promise<void>((resolve, reject) => {
        const timeout = setTimeout(() => {
          audio.removeEventListener('canplay', onCanPlay)
          audio.removeEventListener('error', onError)
          reject(new Error('Audio loading timeout'))
        }, 30000)

        const onCanPlay = () => {
          clearTimeout(timeout)
          audio.removeEventListener('canplay', onCanPlay)
          audio.removeEventListener('error', onError)
          resolve()
        }
        const onError = () => {
          clearTimeout(timeout)
          audio.removeEventListener('canplay', onCanPlay)
          audio.removeEventListener('error', onError)
          reject(new Error('Failed to load audio'))
        }
        if (audio.readyState >= HTMLMediaElement.HAVE_FUTURE_DATA) {
          clearTimeout(timeout)
          resolve()
        } else {
          audio.addEventListener('canplay', onCanPlay)
          audio.addEventListener('error', onError)
        }
      })
    }
  }

  // Control playback state
  const isPlaying = state.isPlaying === true
  const isPaused = state.isPlaying === false

  if (isPlaying) {
    // State says playing - start playback if paused
    if (audio.paused) {
      // For Safari native HLS, audio.src will be set directly
      // For HLS.js, hlsInstance will be set
      // For regular audio, audio.src will be set
      const hasSource =
        (audio.src && audio.src.trim() !== '') || hlsInstance || currentSourceBaseUrl

      console.log('[Session Audio] Checking source for playback:', {
        hasSrc: !!audio.src && audio.src.trim() !== '',
        srcType: audio.src?.substring(0, 30) || 'none',
        hasHlsInstance: !!hlsInstance,
        hasStoredUrl: !!currentSourceBaseUrl,
        hasSource,
        readyState: audio.readyState,
      })

      if (hasSource) {
        try {
          // Wait for audio to be ready if needed
          // For HLS streams, we need more time and can use HAVE_METADATA instead of HAVE_CURRENT_DATA
          const isHls = hlsInstance || (audio.src && isHlsUrl(audio.src))
          const minReadyState = isHls
            ? HTMLMediaElement.HAVE_METADATA // For HLS, metadata is enough to start playback
            : HTMLMediaElement.HAVE_CURRENT_DATA
          const timeoutMs = isHls ? 15000 : 10000 // Longer timeout for HLS

          if (audio.readyState < minReadyState) {
            console.log(
              `[Session Audio] Waiting for audio to be ready (isHls: ${isHls}, readyState: ${audio.readyState})...`
            )
            try {
              await new Promise<void>(resolve => {
                const timeout = setTimeout(() => {
                  audio.removeEventListener('canplay', onCanPlay)
                  audio.removeEventListener('canplaythrough', onCanPlayThrough)
                  audio.removeEventListener('loadedmetadata', onLoadedMetadata)
                  audio.removeEventListener('error', onError)
                  // Don't reject - just resolve and try to play anyway
                  console.warn('[Session Audio] Audio ready timeout, but will try to play anyway')
                  resolve()
                }, timeoutMs)

                const onCanPlay = () => {
                  clearTimeout(timeout)
                  audio.removeEventListener('canplay', onCanPlay)
                  audio.removeEventListener('canplaythrough', onCanPlayThrough)
                  audio.removeEventListener('loadedmetadata', onLoadedMetadata)
                  audio.removeEventListener('error', onError)
                  console.log('[Session Audio] Audio is ready (canplay event)')
                  resolve()
                }

                const onCanPlayThrough = () => {
                  clearTimeout(timeout)
                  audio.removeEventListener('canplay', onCanPlay)
                  audio.removeEventListener('canplaythrough', onCanPlayThrough)
                  audio.removeEventListener('loadedmetadata', onLoadedMetadata)
                  audio.removeEventListener('error', onError)
                  console.log('[Session Audio] Audio is ready (canplaythrough event)')
                  resolve()
                }

                const onLoadedMetadata = () => {
                  // For HLS, loadedmetadata is often enough
                  if (isHls && audio.readyState >= HTMLMediaElement.HAVE_METADATA) {
                    clearTimeout(timeout)
                    audio.removeEventListener('canplay', onCanPlay)
                    audio.removeEventListener('canplaythrough', onCanPlayThrough)
                    audio.removeEventListener('loadedmetadata', onLoadedMetadata)
                    audio.removeEventListener('error', onError)
                    console.log('[Session Audio] Audio metadata loaded (HLS)')
                    resolve()
                  }
                }

                const onError = () => {
                  clearTimeout(timeout)
                  audio.removeEventListener('canplay', onCanPlay)
                  audio.removeEventListener('canplaythrough', onCanPlayThrough)
                  audio.removeEventListener('loadedmetadata', onLoadedMetadata)
                  audio.removeEventListener('error', onError)
                  // Don't reject - just resolve and let play() handle the error
                  console.warn('[Session Audio] Audio error event, but will try to play anyway')
                  resolve()
                }

                // Check if already ready
                if (audio.readyState >= minReadyState) {
                  clearTimeout(timeout)
                  resolve()
                } else {
                  audio.addEventListener('canplay', onCanPlay)
                  audio.addEventListener('canplaythrough', onCanPlayThrough)
                  if (isHls) {
                    audio.addEventListener('loadedmetadata', onLoadedMetadata)
                  }
                  audio.addEventListener('error', onError)
                }
              })
            } catch (error) {
              console.warn('[Session Audio] Error waiting for audio ready:', error)
              // Continue anyway - try to play
            }
          }

          console.log(
            '[Session Audio] Starting playback, readyState:',
            audio.readyState,
            'paused:',
            audio.paused
          )

          // Sync position before starting playback (important for pause/resume)
          if (typeof state.position === 'number' && state.position > 0) {
            syncPosition(audio, state.position)
          }

          await audio.play()
          console.log('[Session Audio] Playback started successfully')
        } catch (error) {
          console.warn('[Session Audio] Failed to start playback:', error)
          if (error instanceof Error) {
            console.warn('[Session Audio] Error name:', error.name, 'message:', error.message)
          }
        }
      } else {
        console.log('[Session Audio] Cannot start playback - source not loaded yet')
      }
    } else {
      // Already playing - sync position if needed
      if (typeof state.position === 'number') {
        syncPosition(audio, state.position)
      }
      console.log('[Session Audio] Already playing, position synced')
    }
  } else if (isPaused) {
    // State says paused - pause playback
    // But don't pause if track hasn't started (normal state for new track)
    const trackHasNotStarted = audio.paused && audio.currentTime === 0
    if (!trackHasNotStarted && !audio.paused) {
      // Sync position before pausing (so we can resume from correct position)
      // Only sync if position is provided and is meaningful (not 0 or stale)
      if (typeof state.position === 'number' && state.position > 0) {
        syncPosition(audio, state.position)
      }
      console.log('[Session Audio] Pausing playback')
      audio.pause()
    } else if (audio.paused) {
      // Track is already paused - sync position if provided and meaningful
      // Don't sync if:
      // 1. Position is 0 but current position > 0 (stale sync state)
      // 2. Position is very close to current position (within 1 second)
      if (typeof state.position === 'number' && state.position > 0) {
        const drift = Math.abs(audio.currentTime - state.position)
        // Only sync if there's a significant difference (more than 1 second)
        // This prevents resetting position to 0 when stale sync_state arrives
        if (drift > 1.0 && audio.currentTime > 0) {
          console.log(
            `[Session Audio] Syncing paused position: ${audio.currentTime.toFixed(2)} -> ${state.position.toFixed(2)}`
          )
          audio.currentTime = state.position
        } else {
          console.log(
            `[Session Audio] Skipping position sync for paused track (drift: ${drift.toFixed(2)}s, current: ${audio.currentTime.toFixed(2)}s, state: ${state.position.toFixed(2)}s)`
          )
        }
      } else if (
        typeof state.position === 'number' &&
        state.position === 0 &&
        audio.currentTime > 1.0
      ) {
        // Don't reset to 0 if we're already at a meaningful position
        console.log(
          `[Session Audio] Ignoring position sync to 0 (current position: ${audio.currentTime.toFixed(2)}s)`
        )
      }
    }
  }
}

export const stopSessionAudio = () => {
  if (hlsInstance) {
    hlsInstance.destroy()
    hlsInstance = null
  }
  currentSourceBaseUrl = null
  currentTrackId = null
  if (!audioElement) return
  audioElement.pause()
  audioElement.src = ''
  audioElement.load()
}

export const getSessionAudio = () => audioElement

export const startSessionPlayback = async () => {
  const audio = getSessionAudio()
  if (!audio) {
    console.warn('[Session Audio] No audio element available')
    return
  }

  // Mark user initiated playback
  userInitiatedPlayback = true
  if (userInitiatedPlaybackTimeout) {
    clearTimeout(userInitiatedPlaybackTimeout)
  }
  userInitiatedPlaybackTimeout = setTimeout(() => {
    userInitiatedPlayback = false
    userInitiatedPlaybackTimeout = null
  }, 3000)

  try {
    // Check if we have a source
    const hasSource = (audio.src && audio.src.trim() !== '') || hlsInstance || currentSourceBaseUrl
    if (!hasSource) {
      throw new Error('No audio source available')
    }

    // For HLS streams, use HAVE_METADATA and longer timeout
    const isHlsStream = hlsInstance || (audio.src && isHlsUrl(audio.src))
    const minReadyState = isHlsStream
      ? HTMLMediaElement.HAVE_METADATA
      : HTMLMediaElement.HAVE_CURRENT_DATA
    const timeoutMs = isHlsStream ? 15000 : 10000

    if (audio.readyState < minReadyState) {
      console.log(
        `[Session Audio] Waiting for audio ready (isHls: ${isHlsStream}, readyState: ${audio.readyState})...`
      )
      try {
        await new Promise<void>(resolve => {
          const timeout = setTimeout(() => {
            audio.removeEventListener('canplay', onCanPlay)
            audio.removeEventListener('canplaythrough', onCanPlayThrough)
            audio.removeEventListener('loadedmetadata', onLoadedMetadata)
            audio.removeEventListener('error', onError)
            // Don't reject - just resolve and try to play
            console.warn(
              '[Session Audio] Audio ready timeout in startSessionPlayback, but will try to play anyway'
            )
            resolve()
          }, timeoutMs)

          const onCanPlay = () => {
            clearTimeout(timeout)
            audio.removeEventListener('canplay', onCanPlay)
            audio.removeEventListener('canplaythrough', onCanPlayThrough)
            audio.removeEventListener('loadedmetadata', onLoadedMetadata)
            audio.removeEventListener('error', onError)
            resolve()
          }

          const onCanPlayThrough = () => {
            clearTimeout(timeout)
            audio.removeEventListener('canplay', onCanPlay)
            audio.removeEventListener('canplaythrough', onCanPlayThrough)
            audio.removeEventListener('loadedmetadata', onLoadedMetadata)
            audio.removeEventListener('error', onError)
            resolve()
          }

          const onLoadedMetadata = () => {
            if (isHlsStream && audio.readyState >= HTMLMediaElement.HAVE_METADATA) {
              clearTimeout(timeout)
              audio.removeEventListener('canplay', onCanPlay)
              audio.removeEventListener('canplaythrough', onCanPlayThrough)
              audio.removeEventListener('loadedmetadata', onLoadedMetadata)
              audio.removeEventListener('error', onError)
              resolve()
            }
          }

          const onError = () => {
            clearTimeout(timeout)
            audio.removeEventListener('canplay', onCanPlay)
            audio.removeEventListener('canplaythrough', onCanPlayThrough)
            audio.removeEventListener('loadedmetadata', onLoadedMetadata)
            audio.removeEventListener('error', onError)
            // Don't reject - just resolve and let play() handle the error
            console.warn(
              '[Session Audio] Audio error in startSessionPlayback, but will try to play anyway'
            )
            resolve()
          }

          if (audio.readyState >= minReadyState) {
            clearTimeout(timeout)
            resolve()
          } else {
            audio.addEventListener('canplay', onCanPlay)
            audio.addEventListener('canplaythrough', onCanPlayThrough)
            if (isHlsStream) {
              audio.addEventListener('loadedmetadata', onLoadedMetadata)
            }
            audio.addEventListener('error', onError)
          }
        })
      } catch (error) {
        console.warn(
          '[Session Audio] Error waiting for audio ready in startSessionPlayback:',
          error
        )
        // Continue anyway - try to play
      }
    }

    await audio.play()
    console.log('[Session Audio] User started playback')
  } catch (error) {
    console.error('[Session Audio] Failed to start playback:', error)
    userInitiatedPlayback = false
    if (userInitiatedPlaybackTimeout) {
      clearTimeout(userInitiatedPlaybackTimeout)
      userInitiatedPlaybackTimeout = null
    }
    throw error
  }
}

export const pauseSessionPlayback = () => {
  const audio = getSessionAudio()
  if (!audio) {
    return
  }
  console.log('[Session Audio] User paused playback')
  audio.pause()
}
