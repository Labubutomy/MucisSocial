export interface PlaylistSummary {
  id: string
  title: string
  coverUrl: string
  itemsCount: number
  description?: string
}

export interface PlaylistDetail extends PlaylistSummary {
  owner: {
    id: string
    name: string
  }
  type: 'playlist' | 'album'
  liked?: boolean
  totalDuration?: number
}
