import { useState } from 'react'
import { useUnit } from 'effector-react'
import { Card } from '@shared/ui/card'
import { Input } from '@shared/ui/input'
import { Button } from '@shared/ui/button'
import { Chip } from '@shared/ui/chip'
import { cn } from '@shared/lib/cn'
import { routes } from '@shared/router'
import {
  $form,
  titleChanged,
  descriptionChanged,
  cityChanged,
  countryChanged,
  privacyToggled,
  linearToggled,
  tagAdded,
  tagRemoved,
  pointAdded,
  pointRemoved,
  pointUpdated,
  formSubmitted,
  createRouteFx,
} from '@pages/create-route/model'
import { RouteMap } from './route-map'
import { PointEditor } from './point-editor'
import { TrackSelector } from './track-selector'

const tagSuggestions = [
  'Экскурсия',
  'Романтика',
  'Спорт',
  'История',
  'Архитектура',
  'Природа',
  'Ночная жизнь',
  'Еда',
]

export const CreateRoutePage = () => {
  const {
    form,
    goBack,
    changeTitle,
    changeDescription,
    changeCity,
    changeCountry,
    togglePrivacy,
    toggleLinear,
    addTag,
    removeTag,
    addPoint,
    removePoint,
    updatePoint,
    submitForm,
    creating,
  } = useUnit({
    form: $form,
    goBack: routes.home.navigate,
    changeTitle: titleChanged,
    changeDescription: descriptionChanged,
    changeCity: cityChanged,
    changeCountry: countryChanged,
    togglePrivacy: privacyToggled,
    toggleLinear: linearToggled,
    addTag: tagAdded,
    removeTag: tagRemoved,
    addPoint: pointAdded,
    removePoint: pointRemoved,
    updatePoint: pointUpdated,
    submitForm: formSubmitted,
    creating: createRouteFx.pending,
  })

  const [selectedPointIndex, setSelectedPointIndex] = useState<number | null>(null)
  const [trackSelectorOpen, setTrackSelectorOpen] = useState(false)
  const [trackSelectorPointIndex, setTrackSelectorPointIndex] = useState<number | null>(null)

  const handleMapClick = (lat: number, lng: number) => {
    addPoint({
      latitude: lat,
      longitude: lng,
      radius_meters: 50,
      track_id: '',
      track_start_offset_sec: 0,
    })
  }

  const handlePointClick = (index: number) => {
    setSelectedPointIndex(index === selectedPointIndex ? null : index)
  }

  const handleSelectTrack = (index: number) => {
    setTrackSelectorPointIndex(index)
    setTrackSelectorOpen(true)
  }

  const handleTrackSelected = (trackId: string) => {
    if (trackSelectorPointIndex !== null) {
      updatePoint({ index: trackSelectorPointIndex, point: { track_id: trackId } })
    }
    setTrackSelectorOpen(false)
    setTrackSelectorPointIndex(null)
  }

  const handleSubmit: React.FormEventHandler<HTMLFormElement> = event => {
    event.preventDefault()
    if (form.title.trim() && form.points.length > 0) {
      submitForm()
    }
  }

  const canSubmit =
    form.title.trim().length > 0 && form.points.length > 0 && form.points.every(p => p.track_id)

  return (
    <div className="page-container space-y-8 pb-20 pt-10">
      <header className="space-y-3">
        <p className="text-xs uppercase tracking-[0.4em] text-primary">Создание маршрута</p>
        <h1 className="text-3xl font-semibold md:text-4xl">Создайте музыкальный маршрут</h1>
        <p className="max-w-2xl text-base text-muted-foreground md:text-lg">
          Добавьте точки на карте и привяжите к ним треки. Другие пользователи смогут проходить ваш
          маршрут, слушая музыку при приближении к точкам.
        </p>
      </header>

      <form onSubmit={handleSubmit} className="space-y-6">
        <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr),minmax(0,1fr)]">
          {/* Левая колонка - Форма */}
          <Card padding="lg" className="space-y-6 bg-secondary/20">
            <div className="space-y-4">
              <Input
                label="Название маршрута"
                placeholder="Например, Музыкальная экскурсия по центру"
                value={form.title}
                onChange={event => changeTitle(event.target.value)}
                required
              />
              <div className="flex flex-col gap-2">
                <label className="text-sm font-medium text-muted-foreground">Описание</label>
                <textarea
                  value={form.description}
                  onChange={event => changeDescription(event.target.value)}
                  placeholder="Расскажите о маршруте"
                  className="min-h-[100px] rounded-xl border border-input bg-secondary/30 px-4 py-3 text-base text-foreground placeholder:text-muted-foreground/70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                />
              </div>
              <div className="grid grid-cols-2 gap-4">
                <Input
                  label="Город"
                  placeholder="Москва"
                  value={form.city}
                  onChange={event => changeCity(event.target.value)}
                />
                <Input
                  label="Страна"
                  placeholder="Россия"
                  value={form.country}
                  onChange={event => changeCountry(event.target.value)}
                />
              </div>
            </div>

            <div className="space-y-4">
              <p className="text-sm font-semibold text-muted-foreground">Теги</p>
              <div className="flex flex-wrap gap-2">
                {tagSuggestions.map(tag => (
                  <Chip
                    key={tag}
                    selected={form.tags.includes(tag)}
                    onClick={() => (form.tags.includes(tag) ? removeTag(tag) : addTag(tag))}
                  >
                    {tag}
                  </Chip>
                ))}
              </div>
            </div>

            <div className="flex flex-col gap-3 rounded-2xl border border-border/60 bg-secondary/30 px-4 py-3">
              <div className="flex items-center justify-between">
                <div className="space-y-1">
                  <p className="text-sm font-semibold text-foreground">Публичный маршрут</p>
                  <p className="text-xs text-muted-foreground">
                    Маршрут будет виден всем пользователям и доступен для прохождения.
                  </p>
                </div>
                <button
                  type="button"
                  onClick={() => togglePrivacy()}
                  className={cn(
                    'relative inline-flex h-6 w-12 items-center rounded-full border transition',
                    form.is_public
                      ? 'border-primary bg-primary/40'
                      : 'border-border/60 bg-secondary/30'
                  )}
                >
                  <span
                    className={cn(
                      'inline-block h-5 w-5 transform rounded-full transition',
                      form.is_public
                        ? 'translate-x-[26px] bg-primary text-primary-foreground'
                        : 'translate-x-[2px] bg-background'
                    )}
                  />
                </button>
              </div>
              <div className="flex items-center justify-between border-t border-border/60 pt-3">
                <div className="space-y-1">
                  <p className="text-sm font-semibold text-foreground">Линейный маршрут</p>
                  <p className="text-xs text-muted-foreground">
                    Точки должны посещаться в порядке добавления.
                  </p>
                </div>
                <button
                  type="button"
                  onClick={() => toggleLinear()}
                  className={cn(
                    'relative inline-flex h-6 w-12 items-center rounded-full border transition',
                    form.is_linear
                      ? 'border-primary bg-primary/40'
                      : 'border-border/60 bg-secondary/30'
                  )}
                >
                  <span
                    className={cn(
                      'inline-block h-5 w-5 transform rounded-full transition',
                      form.is_linear
                        ? 'translate-x-[26px] bg-primary text-primary-foreground'
                        : 'translate-x-[2px] bg-background'
                    )}
                  />
                </button>
              </div>
            </div>
          </Card>

          {/* Правая колонка - Карта */}
          <Card padding="lg" className="bg-secondary/20">
            <div className="space-y-4">
              <div>
                <p className="text-sm font-semibold text-muted-foreground mb-2">Карта маршрута</p>
                <p className="text-xs text-muted-foreground">
                  Кликните на карте, чтобы добавить точку. Всего точек: {form.points.length}
                </p>
              </div>
              <div className="h-[400px] w-full">
                <RouteMap
                  points={form.points}
                  onPointAdd={handleMapClick}
                  onPointClick={handlePointClick}
                  selectedPointIndex={selectedPointIndex}
                />
              </div>
            </div>
          </Card>
        </div>

        {/* Список точек */}
        {form.points.length > 0 && (
          <Card padding="lg" className="space-y-4 bg-secondary/20">
            <div>
              <p className="text-sm font-semibold text-muted-foreground mb-2">Точки маршрута</p>
              <p className="text-xs text-muted-foreground">
                Настройте каждую точку: добавьте название, описание и выберите трек.
              </p>
            </div>
            <div className="space-y-3">
              {form.points.map((point, index) => (
                <PointEditor
                  key={index}
                  point={point}
                  index={index}
                  onUpdate={point => updatePoint({ index, point })}
                  onRemove={() => removePoint(index)}
                  onSelectTrack={() => handleSelectTrack(index)}
                />
              ))}
            </div>
          </Card>
        )}

        {/* Кнопки действий */}
        <div className="flex flex-col gap-3 md:flex-row md:justify-end">
          <Button
            type="button"
            variant="outline"
            onClick={() => goBack({ params: {}, query: {} })}
            className="md:w-auto"
          >
            Отменить
          </Button>
          <Button type="submit" className="md:w-auto" disabled={!canSubmit || creating}>
            {creating ? 'Сохранение...' : 'Сохранить маршрут'}
          </Button>
        </div>
      </form>

      {trackSelectorOpen && (
        <TrackSelector
          onSelect={handleTrackSelected}
          onClose={() => {
            setTrackSelectorOpen(false)
            setTrackSelectorPointIndex(null)
          }}
          currentTrackId={
            trackSelectorPointIndex !== null
              ? form.points[trackSelectorPointIndex]?.track_id
              : undefined
          }
        />
      )}
    </div>
  )
}
