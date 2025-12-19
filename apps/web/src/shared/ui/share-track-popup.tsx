import { useState } from 'react'
import { Card } from './card'
import { Button } from './button'
import { Input } from './input'
import { Avatar } from './avatar'
import type { Track } from '@entities/track'
import type { Friend } from '@features/friends/api'

interface ShareTrackPopupProps {
  track: Track
  friends: Friend[]
  onSelectFriend: (friendId: string, track: Track) => void
  onClose: () => void
}

export const ShareTrackPopup = ({
  track,
  friends,
  onSelectFriend,
  onClose,
}: ShareTrackPopupProps) => {
  const [searchQuery, setSearchQuery] = useState('')

  const filteredFriends = friends.filter(friend => {
    const username = friend.friendUsername || friend.friend_username || ''
    return username.toLowerCase().includes(searchQuery.toLowerCase())
  })

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <Card padding="lg" className="w-full max-w-md">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-xl font-semibold">Поделиться треком</h2>
          <Button type="button" variant="ghost" size="sm" onClick={onClose}>
            ✕
          </Button>
        </div>

        <div className="mb-4 flex items-center gap-3 rounded-lg bg-secondary/20 p-3">
          <div className="h-12 w-12 overflow-hidden rounded">
            <img src={track.coverUrl} alt={track.title} className="h-full w-full object-cover" />
          </div>
          <div className="flex-1 min-w-0">
            <p className="truncate font-semibold">{track.title}</p>
            <p className="truncate text-sm text-muted-foreground">{track.artist.name}</p>
          </div>
        </div>

        <Input
          placeholder="Поиск друзей..."
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
          className="mb-4"
        />

        <div className="max-h-[400px] space-y-2 overflow-y-auto">
          {filteredFriends.length === 0 ? (
            <p className="py-8 text-center text-sm text-muted-foreground">
              {searchQuery ? 'Друзья не найдены' : 'У вас пока нет друзей'}
            </p>
          ) : (
            filteredFriends.map(friend => {
              const friendId = friend.friendId || friend.friend_id || ''
              const username = friend.friendUsername || friend.friend_username || 'Друг'
              const avatarUrl = friend.friendAvatarUrl || friend.friend_avatar_url

              return (
                <button
                  key={friendId}
                  type="button"
                  onClick={() => {
                    onSelectFriend(friendId, track)
                    onClose()
                  }}
                  className="flex w-full items-center gap-3 rounded-lg border border-border/60 bg-background/60 p-3 text-left transition hover:bg-secondary/50"
                >
                  <Avatar src={avatarUrl || undefined} fallback={username} size="sm" />
                  <div className="flex-1 min-w-0">
                    <p className="truncate font-medium">{username}</p>
                  </div>
                </button>
              )
            })
          )}
        </div>
      </Card>
    </div>
  )
}
