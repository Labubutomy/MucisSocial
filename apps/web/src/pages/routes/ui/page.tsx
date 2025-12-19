import { useEffect } from 'react'
import { useUnit } from 'effector-react'
import { Card } from '@shared/ui/card'
import { Input } from '@shared/ui/input'
import { Button } from '@shared/ui/button'
import { routes } from '@shared/router'
import {
  $routes,
  $loading,
  $total,
  $filters,
  filtersChanged,
  loadRoutes,
  loadMoreRoutes,
} from '@pages/routes/model'
import { RouteCard } from './route-card'

export const RoutesPage = () => {
  const {
    routes: routesList,
    loading,
    total,
    filters,
    changeFilters,
    load,
    loadMore,
    navigateToCreate,
  } = useUnit({
    routes: $routes,
    loading: $loading,
    total: $total,
    filters: $filters,
    changeFilters: filtersChanged,
    load: loadRoutes,
    loadMore: loadMoreRoutes,
    navigateToCreate: routes.routeCreate.navigate,
  })

  useEffect(() => {
    load()
  }, [load])

  const handleCityChange = (city: string) => {
    changeFilters({ city: city || undefined, offset: 0 })
    load()
  }

  const hasMore = routesList.length < total

  return (
    <div className="page-container space-y-8 pb-20 pt-10">
      <header className="flex items-center justify-between">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold md:text-4xl">Маршруты</h1>
          <p className="text-muted-foreground">
            Откройте для себя музыкальные маршруты других пользователей
          </p>
        </div>
        <Button onClick={() => navigateToCreate({ params: {}, query: {} })}>Создать маршрут</Button>
      </header>

      <Card padding="lg" className="bg-secondary/20">
        <div className="space-y-4">
          <div className="flex gap-4">
            <Input
              placeholder="Фильтр по городу..."
              value={filters.city || ''}
              onChange={e => handleCityChange(e.target.value)}
              className="flex-1"
            />
            <Button
              variant="outline"
              onClick={() => {
                changeFilters({ is_public: true, offset: 0 })
                load()
              }}
            >
              Публичные
            </Button>
            <Button
              variant="outline"
              onClick={() => {
                changeFilters({ is_public: undefined, offset: 0 })
                load()
              }}
            >
              Все
            </Button>
          </div>
        </div>
      </Card>

      {loading && routesList.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">Загрузка маршрутов...</div>
      ) : routesList.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">
          <p className="text-lg font-semibold mb-2">Маршруты не найдены</p>
          <p className="text-sm">Попробуйте изменить фильтры или создайте свой первый маршрут</p>
        </div>
      ) : (
        <>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {routesList.map(route => (
              <RouteCard key={route.id} route={route} />
            ))}
          </div>

          {hasMore && (
            <div className="flex justify-center">
              <Button variant="outline" onClick={loadMore} disabled={loading}>
                {loading ? 'Загрузка...' : 'Загрузить еще'}
              </Button>
            </div>
          )}

          <div className="text-center text-sm text-muted-foreground">
            Показано {routesList.length} из {total} маршрутов
          </div>
        </>
      )}
    </div>
  )
}
