import { useUnit } from 'effector-react'
import { Tabs } from '@shared/ui/tabs'
import type { TabItem } from '@shared/ui/tabs'
import { TrackFeed } from '@widgets/home'
import { $currentTrack, $isPlaying, playbackToggled, trackQueued } from '@features/player'
import type { Track } from '@entities/track'
import { routes } from '@shared/router'
import { $activeTab, $tracks, tabChanged, trackLikedToggled, type FeedTab } from '@pages/home/model'

const tabItems: TabItem[] = [
  { value: 'trending', label: 'В тренде' },
  { value: 'popular', label: 'Популярное' },
  { value: 'new', label: 'Новое' },
]

export const HomePage = () => {
  const {
    tracks,
    activeTab,
    currentTrack,
    isPlaying,
    enqueueTrack,
    togglePlayback,
    navigateToTrack,
    toggleLike,
    changeTab,
  } = useUnit({
    tracks: $tracks,
    activeTab: $activeTab,
    currentTrack: $currentTrack,
    isPlaying: $isPlaying,
    enqueueTrack: trackQueued,
    togglePlayback: playbackToggled,
    navigateToTrack: routes.track.navigate,
    toggleLike: trackLikedToggled,
    changeTab: tabChanged,
  })

  const handlePlayToggle = (track: Track) => {
    if (!currentTrack || currentTrack.id !== track.id) {
      enqueueTrack(track)
      return
    }
    togglePlayback()
  }

  const handleLike = (track: Track) => {
    toggleLike(track.id)
  }

  const handleShare = (track: Track) => {
    console.info('Поделиться треком', track.id)
  }

  const handleOpen = (track: Track) => {
    navigateToTrack({
      params: { trackId: track.id },
      query: {},
    })
  }

  return (
    <div className="page-container space-y-12 pb-24">
      <header className="space-y-6 pt-6">
        <div className="flex flex-col gap-3">
          <p className="text-xs uppercase tracking-[0.4em] text-primary">Для вас</p>
          <h1 className="text-3xl font-semibold md:text-5xl">
            Почувствуйте пульс музыкального сообщества
          </h1>
          <p className="max-w-2xl text-base text-muted-foreground md:text-lg">
            Узнавайте, что слушают прямо сейчас, к чему до сих пор возвращаются и какие релизы
            появились совсем недавно — всё в одном потоке.
          </p>
        </div>
        <Tabs value={activeTab} onChange={value => changeTab(value as FeedTab)} items={tabItems} />
      </header>

      <TrackFeed
        title="Треки"
        subtitle="Свежие композиции под ваше настроение"
        tracks={tracks}
        activeTrackId={currentTrack?.id}
        isPlaying={isPlaying && Boolean(currentTrack)}
        onPlayToggle={handlePlayToggle}
        onLike={handleLike}
        onShare={handleShare}
        onOpen={handleOpen}
      />
    </div>
  )
}
