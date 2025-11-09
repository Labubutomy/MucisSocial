import { useUnit } from 'effector-react'
import { Card } from '@shared/ui/card'
import { Button } from '@shared/ui/button'
import { SearchBar } from '@features/search'
import { TrackRow } from '@entities/track'
import type { Track } from '@entities/track'
import { routes } from '@shared/router'
import {
  $search,
  $selectedTracks,
  $suggestions,
  saveRequested,
  searchChanged,
  trackSelected,
  saveTracksFx,
} from '@pages/playlist-add-tracks/model'

export const PlaylistAddTracksPage = () => {
  const {
    search,
    selectedTracks,
    suggestions,
    params,
    changeSearch,
    addTrack,
    saveSelection,
    saving,
    goToTrack,
    goBack,
  } = useUnit({
    search: $search,
    selectedTracks: $selectedTracks,
    suggestions: $suggestions,
    params: routes.playlistAddTracks.$params,
    changeSearch: searchChanged,
    addTrack: trackSelected,
    saveSelection: saveRequested,
    saving: saveTracksFx.pending,
    goToTrack: routes.track.navigate,
    goBack: routes.profilePlaylists.navigate,
  })

  const handleAddTrack = (track: Track) => {
    addTrack(track)
  }

  return (
    <div className="page-container space-y-8 pb-20 pt-10">
      <header className="space-y-3">
        <p className="text-xs uppercase tracking-[0.4em] text-primary">
          Добавление треков в плейлист
        </p>
        <h1 className="text-3xl font-semibold md:text-4xl">Подберите треки для свежей подборки</h1>
        <p className="max-w-2xl text-base text-muted-foreground md:text-lg">
          Найдите нужные композиции или воспользуйтесь рекомендациями. Сохраните изменения, когда
          будете готовы поделиться плейлистом.
        </p>
      </header>

      <div className="grid gap-8 lg:grid-cols-[minmax(0,1fr),minmax(0,0.9fr)]">
        <Card padding="lg" className="space-y-6 bg-secondary/20">
          <div className="space-y-4">
            <p className="text-sm font-semibold text-muted-foreground">Выбранные треки</p>
            {selectedTracks.length === 0 ? (
              <div className="rounded-2xl border border-dashed border-border/60 bg-secondary/30 px-4 py-10 text-center text-sm text-muted-foreground">
                Пока ничего нет. Добавьте треки из подсказок или найдите их в поиске.
              </div>
            ) : (
              <div className="space-y-2 max-h-[360px] overflow-y-auto pr-2">
                {selectedTracks.map((track, index) => (
                  <TrackRow
                    key={track.id}
                    track={track}
                    index={index}
                    onPlayToggle={() => console.info('Играть трек', track.id)}
                    onLike={() => console.info('Лайкнуть трек', track.id)}
                    onAddToPlaylist={() => console.info('Добавить в другой плейлист', track.id)}
                    onShare={() => console.info('Поделиться треком', track.id)}
                    onOpen={() =>
                      goToTrack({
                        params: { trackId: track.id },
                        query: {},
                      })
                    }
                    className="cursor-pointer"
                  />
                ))}
              </div>
            )}
          </div>
          <div className="flex flex-col gap-3 md:flex-row md:justify-end">
            <Button variant="outline" onClick={() => goBack({ params: {}, query: {} })}>
              Отменить
            </Button>
            <Button
              onClick={() =>
                saveSelection({
                  playlistId: params.playlistId,
                })
              }
              disabled={selectedTracks.length === 0 || saving}
            >
              Сохранить
            </Button>
          </div>
        </Card>

        <Card padding="lg" className="space-y-6 bg-secondary/20">
          <div className="space-y-3">
            <p className="text-xs uppercase tracking-[0.4em] text-primary">Поиск треков</p>
            <SearchBar
              value={search}
              onChange={event => changeSearch(event.target.value)}
              onSubmit={event => {
                event.preventDefault()
                console.info('Найти трек', search)
              }}
            />
          </div>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <p className="text-sm font-semibold text-foreground">Рекомендации</p>
            </div>
            <div className="space-y-2 max-h-[360px] overflow-y-auto pr-2">
              {suggestions.length === 0 ? (
                <div className="rounded-2xl border border-dashed border-border/60 bg-secondary/30 px-4 py-10 text-center text-sm text-muted-foreground">
                  Пока нет рекомендаций. Попробуйте воспользоваться поиском.
                </div>
              ) : (
                suggestions.map((track, index) => (
                  <TrackRow
                    key={track.id}
                    track={track}
                    index={index}
                    onPlayToggle={() => console.info('Играть рекомендованный трек', track.id)}
                    onLike={() => console.info('Лайкнуть рекомендованный трек', track.id)}
                    onAddToPlaylist={() => console.info('Добавить в другой плейлист', track.id)}
                    onShare={() => console.info('Поделиться треком', track.id)}
                    onOpen={() => handleAddTrack(track)}
                    className="cursor-pointer"
                  />
                ))
              )}
            </div>
            <p className="text-xs text-muted-foreground">
              Нажмите на трек, чтобы добавить его в плейлист. Используйте иконку воспроизведения,
              чтобы послушать перед добавлением.
            </p>
          </div>
        </Card>
      </div>
    </div>
  )
}
