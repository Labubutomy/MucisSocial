export interface UserProfile {
  id: string
  username: string
  avatarUrl?: string
  musicTasteSummary?: {
    topGenres?: string[]
    topArtists?: string[]
  }
}
