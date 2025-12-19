import Hls from 'hls.js'
import { $user } from '@features/auth'

let audioElement: HTMLAudioElement | null = null
let hlsInstance: Hls | null = null
let currentUserId: string | null = null

// Subscribe to user changes to keep currentUserId updated
if (typeof window !== 'undefined') {
  $user.watch(user => {
    currentUserId = user?.id || null
  })
}

/**
 * Adds user_id to URL query parameters if user is authenticated
 */
const addUserIdToUrl = (url: string): string => {
  try {
    if (!currentUserId) {
      return url
    }

    const urlObj = new URL(url)
    // Only add if not already present
    if (!urlObj.searchParams.has('user_id')) {
      urlObj.searchParams.set('user_id', currentUserId)
    }
    return urlObj.toString()
  } catch {
    // If URL parsing fails, return original URL
    return url
  }
}

const ensureAudioElement = () => {
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

const getExistingAudioElement = () => audioElement

export const attachListener = (event: keyof HTMLMediaElementEventMap, handler: () => void) => {
  const audio = ensureAudioElement()
  audio.addEventListener(event, handler)

  return () => audio.removeEventListener(event, handler)
}

export const playStream = async (url: string) => {
  const audio = ensureAudioElement()

  // Останавливаем текущее воспроизведение перед загрузкой нового источника
  try {
    audio.pause()
  } catch {
    // Игнорируем ошибки при паузе
  }

  if (hlsInstance) {
    hlsInstance.destroy()
    hlsInstance = null
  }

  audio.currentTime = 0

  // Ждем, пока текущий запрос play() завершится перед загрузкой нового источника
  await new Promise<void>(resolve => {
    if (audio.readyState === HTMLMediaElement.HAVE_NOTHING) {
      resolve()
      return
    }
    // Даем время для завершения текущих операций
    setTimeout(resolve, 50)
  })

  // Add user_id to URL for CDN tracking
  const urlWithUserId = addUserIdToUrl(url)

  if (audio.canPlayType('application/vnd.apple.mpegurl')) {
    audio.src = urlWithUserId
    audio.load()
  } else if (Hls.isSupported()) {
    hlsInstance = new Hls({
      enableWorker: true,
      lowLatencyMode: true,
      xhrSetup: (xhr, url) => {
        // Add user_id to all HLS requests (playlists and segments)
        const urlWithUserId = addUserIdToUrl(url)
        xhr.open('GET', urlWithUserId, true)
      },
    })
    hlsInstance.loadSource(urlWithUserId)
    hlsInstance.attachMedia(audio)
  } else {
    audio.src = urlWithUserId
    audio.load()
  }

  // Ждем готовности перед воспроизведением
  await new Promise<void>((resolve, reject) => {
    const timeout = setTimeout(() => {
      audio.removeEventListener('canplay', onCanPlay)
      audio.removeEventListener('error', onError)
      reject(new Error('Audio loading timeout'))
    }, 10000)

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

    if (audio.readyState >= HTMLMediaElement.HAVE_CURRENT_DATA) {
      clearTimeout(timeout)
      resolve()
    } else {
      audio.addEventListener('canplay', onCanPlay)
      audio.addEventListener('error', onError)
    }
  })

  try {
    await audio.play()
  } catch (error) {
    // Игнорируем ошибки play() если они связаны с прерыванием
    if (error instanceof Error && error.name !== 'NotAllowedError') {
      console.warn('Play interrupted, will retry:', error)
      // Повторяем попытку через небольшую задержку
      setTimeout(async () => {
        try {
          await audio.play()
        } catch (retryError) {
          console.error('Failed to play after retry:', retryError)
        }
      }, 100)
    } else {
      throw error
    }
  }
}

export const pauseStream = async () => {
  if (!audioElement) return
  await audioElement.pause()
}

export const resumeStream = async () => {
  const audio = ensureAudioElement()
  await audio.play()
}

export const stopStream = () => {
  if (!audioElement) return
  audioElement.pause()
  audioElement.currentTime = 0
  if (hlsInstance) {
    hlsInstance.destroy()
    hlsInstance = null
  }
}

export const isPlaying = () => {
  if (!audioElement) return false
  return !audioElement.paused && !audioElement.ended
}

export const getCurrentTime = () => {
  const audio = getExistingAudioElement()
  if (!audio) return 0
  return Number.isFinite(audio.currentTime) ? audio.currentTime : 0
}

export const getDuration = () => {
  const audio = getExistingAudioElement()
  if (!audio) return 0
  const { duration } = audio
  if (Number.isFinite(duration) && duration > 0) {
    return duration
  }
  if (audio.buffered.length > 0) {
    return audio.buffered.end(audio.buffered.length - 1)
  }
  return 0
}

export const seekTo = async (seconds: number) => {
  const audio = ensureAudioElement()
  const safeSeconds = Math.max(0, seconds)
  try {
    // Для HLS.js достаточно установить currentTime - HLS автоматически обработает это
    audio.currentTime = safeSeconds
  } catch (error) {
    console.warn('Failed to seek audio element', error)
  }
}
