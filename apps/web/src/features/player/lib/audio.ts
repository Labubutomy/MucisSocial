import Hls from 'hls.js'

let audioElement: HTMLAudioElement | null = null
let hlsInstance: Hls | null = null

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

  if (hlsInstance) {
    hlsInstance.destroy()
    hlsInstance = null
  }

  audio.currentTime = 0

  if (audio.canPlayType('application/vnd.apple.mpegurl')) {
    audio.src = url
  } else if (Hls.isSupported()) {
    hlsInstance = new Hls({
      enableWorker: true,
      lowLatencyMode: true,
    })
    hlsInstance.loadSource(url)
    hlsInstance.attachMedia(audio)
  } else {
    audio.src = url
  }

  await audio.play()
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
    audio.currentTime = safeSeconds
  } catch (error) {
    console.warn('Failed to seek audio element', error)
  }
}
