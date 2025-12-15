import { useUnit } from 'effector-react'
import { Tabs } from '@shared/ui/tabs'
import type { TabItem } from '@shared/ui/tabs'
import { TrackFeed } from '@widgets/home'
import { $currentTrack, $isPlaying, playbackToggled, trackQueued } from '@features/player'
import type { Track } from '@entities/track'
import { routes } from '@shared/router'
import {
  $activeTab,
  $tracks,
  $trendingPending,
  $trendingError,
  $newPending,
  $newError,
  $recommendationsPending,
  $recommendationsError,
  tabChanged,
  trackLikedToggled,
  type FeedTab,
} from '@pages/home/model'

const tabItems: TabItem[] = [
  { value: 'trending', label: 'В тренде' },
  { value: 'new', label: 'Новое' },
  { value: 'recommended', label: 'Рекомендации' },
]

export const HomePage = () => {
  const {
    tracks,
    activeTab,
    trendingPending,
    trendingError,
    newPending,
    newError,
    recommendationsPending,
    recommendationsError,
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
    trendingPending: $trendingPending,
    trendingError: $trendingError,
    newPending: $newPending,
    newError: $newError,
    recommendationsPending: $recommendationsPending,
    recommendationsError: $recommendationsError,
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

      {(() => {
        // Loading states
        if (
          (activeTab === 'trending' && trendingPending) ||
          (activeTab === 'new' && newPending) ||
          (activeTab === 'recommended' && recommendationsPending)
        ) {
          return (
            <div className="rounded-2xl border border-border/60 bg-secondary/20 px-4 py-10 text-center text-sm text-muted-foreground">
              Загрузка...
            </div>
          )
        }

        // Error states
        if (activeTab === 'trending' && trendingError) {
          return (
            <div className="rounded-2xl border border-destructive/40 bg-destructive/10 px-4 py-10 text-center text-sm text-destructive">
              {trendingError}
            </div>
          )
        }
        if (activeTab === 'new' && newError) {
          return (
            <div className="rounded-2xl border border-destructive/40 bg-destructive/10 px-4 py-10 text-center text-sm text-destructive">
              {newError}
            </div>
          )
        }
        if (activeTab === 'recommended' && recommendationsError) {
          return (
            <div className="rounded-2xl border border-destructive/40 bg-destructive/10 px-4 py-10 text-center text-sm text-destructive">
              {recommendationsError}
            </div>
          )
        }

        // Empty states
        if (tracks.length === 0) {
          return (
            <div className="rounded-2xl border border-dashed border-border/60 bg-secondary/30 px-4 py-10 text-center text-sm text-muted-foreground">
              {activeTab === 'trending' && 'Пока нет данных для топ чартов'}
              {activeTab === 'new' && 'Пока нет новинок'}
              {activeTab === 'recommended' && 'Пока нет рекомендаций'}
            </div>
          )
        }

        // Content
        const titles = {
          trending: 'В тренде',
          new: 'Новое',
          recommended: 'Рекомендации для вас',
        }
        const subtitles = {
          trending: 'Самые популярные треки за месяц',
          new: 'Треки отсортированные в обратном порядке добавления',
          recommended: 'Персональные рекомендации на основе ваших предпочтений',
        }

        return (
          <TrackFeed
            title={titles[activeTab]}
            subtitle={subtitles[activeTab]}
            tracks={tracks}
            activeTrackId={currentTrack?.id}
            isPlaying={isPlaying && Boolean(currentTrack)}
            onPlayToggle={handlePlayToggle}
            onLike={handleLike}
            onShare={handleShare}
            onOpen={handleOpen}
          />
        )
      })()}
    </div>
  )
}
