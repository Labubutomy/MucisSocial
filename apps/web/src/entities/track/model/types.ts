export interface ArtistRef {
  id: string
  name: string
}

export interface TrackStreamInfo {
  masterUrl: string
  qualities: string[]
}

export interface Track {
  id: string
  title: string
  artist: ArtistRef
  coverUrl: string
  duration: number
  liked?: boolean
  stream?: TrackStreamInfo
}
