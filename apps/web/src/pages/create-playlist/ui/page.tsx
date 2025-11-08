import { useMemo, useState } from 'react'
import { useUnit } from 'effector-react'
import { Card } from '@shared/ui/card'
import { Input } from '@shared/ui/input'
import { Button } from '@shared/ui/button'
import { Chip } from '@shared/ui/chip'
import { cn } from '@shared/lib/cn'
import { routes } from '@shared/router'

const genreOptions = [
  'Синтвейв',
  'Инди-поп',
  'Лоуфай',
  'Нео-соул',
  'Хип-хоп',
  'Электроника',
  'Поп',
  'Акустика',
]

export const CreatePlaylistPage = () => {
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [genres, setGenres] = useState<string[]>([])
  const [isPrivate, setIsPrivate] = useState(false)

  const goBack = useUnit(routes.profilePlaylists.navigate)
  const goToAddTracks = useUnit(routes.playlistAddTracks.navigate)

  const suggestions = useMemo(
    () =>
      genreOptions.map(genre => ({
        value: genre,
        selected: genres.includes(genre),
      })),
    [genres]
  )

  const toggleGenre = (genre: string) => {
    setGenres(prev =>
      prev.includes(genre) ? prev.filter(item => item !== genre) : [...prev, genre]
    )
  }

  const handleSubmit: React.FormEventHandler<HTMLFormElement> = event => {
    event.preventDefault()
    const newPlaylistId = crypto.randomUUID()
    console.info('Create playlist', { id: newPlaylistId, title, description, genres, isPrivate })
    goToAddTracks({
      params: { playlistId: newPlaylistId },
      query: {},
    })
  }

  return (
    <div className="page-container space-y-8 pb-20 pt-10">
      <header className="space-y-3">
        <p className="text-xs uppercase tracking-[0.4em] text-primary">Создание плейлиста</p>
        <h1 className="text-3xl font-semibold md:text-4xl">
          Соберите новое музыкальное настроение
        </h1>
        <p className="max-w-2xl text-base text-muted-foreground md:text-lg">
          Добавьте детали — и мы предложим, чем наполнить свежий плейлист. Делитесь им с друзьями
          или оставьте только для себя.
        </p>
      </header>

      <form
        onSubmit={handleSubmit}
        className="grid gap-8 lg:grid-cols-[minmax(0,1.1fr),minmax(0,0.9fr)]"
      >
        <Card padding="lg" className="space-y-6 bg-secondary/20">
          <div className="space-y-4">
            <Input
              label="Название плейлиста"
              placeholder="Например, Ночной драйв"
              value={title}
              onChange={event => setTitle(event.target.value)}
              required
            />
            <div className="flex flex-col gap-2">
              <label className="text-sm font-medium text-muted-foreground">Описание</label>
              <textarea
                value={description}
                onChange={event => setDescription(event.target.value)}
                placeholder="Расскажите, для какого настроения этот плейлист"
                className="min-h-[140px] rounded-xl border border-input bg-secondary/30 px-4 py-3 text-base text-foreground placeholder:text-muted-foreground/70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
              />
            </div>
          </div>

          <div className="space-y-4">
            <p className="text-sm font-semibold text-muted-foreground">Жанры и настроение</p>
            <div className="flex flex-wrap gap-2">
              {suggestions.map(item => (
                <Chip
                  key={item.value}
                  selected={item.selected}
                  onClick={() => toggleGenre(item.value)}
                >
                  {item.value}
                </Chip>
              ))}
            </div>
          </div>

          <div className="flex flex-col gap-3 rounded-2xl border border-border/60 bg-secondary/30 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
            <div className="space-y-1">
              <p className="text-sm font-semibold text-foreground">Сделать плейлист приватным</p>
              <p className="text-xs text-muted-foreground">
                Плейлист будет виден только вам и тем, с кем вы поделитесь ссылкой.
              </p>
            </div>
            <button
              type="button"
              onClick={() => setIsPrivate(prev => !prev)}
              className={cn(
                'relative inline-flex h-6 w-12 items-center rounded-full border transition sm:flex-shrink-0',
                isPrivate ? 'border-primary bg-primary/40' : 'border-border/60 bg-secondary/30'
              )}
            >
              <span
                className={cn(
                  'inline-block h-5 w-5 transform rounded-full transition',
                  isPrivate
                    ? 'translate-x-[26px] bg-primary text-primary-foreground'
                    : 'translate-x-[2px] bg-background'
                )}
              />
            </button>
          </div>
        </Card>

        <Card padding="lg" className="flex flex-col gap-6 bg-secondary/20">
          <div className="space-y-2">
            <p className="text-xs uppercase tracking-[0.4em] text-primary">Подсказки</p>
            <h2 className="text-2xl font-semibold text-foreground">Привяжите треки позже</h2>
            <p className="text-sm text-muted-foreground">
              После создания плейлиста вы сможете добавить треки из поиска или коллекции.
            </p>
          </div>
          <div className="flex-1 rounded-2xl border border-dashed border-border/60 bg-secondary/30 px-4 py-6 text-sm text-muted-foreground">
            <p className="font-semibold text-foreground">Как начать:</p>
            <ul className="mt-3 list-disc space-y-2 pl-4">
              <li>Сначала задайте настроение — название и описание.</li>
              <li>Добавьте несколько жанров, чтобы облегчить подбор.</li>
              <li>Сохраните черновик, а затем добавьте треки из поиска.</li>
              <li>Поделитесь плейлистом с друзьями или оставьте приватным.</li>
            </ul>
          </div>
          <div className="flex flex-col gap-3 md:flex-row md:justify-end">
            <Button
              type="button"
              variant="outline"
              onClick={() => goBack({ params: {}, query: {} })}
              className="md:w-auto"
            >
              Отменить
            </Button>
            <Button type="submit" className="md:w-auto">
              Сохранить
            </Button>
          </div>
        </Card>
      </form>
    </div>
  )
}
