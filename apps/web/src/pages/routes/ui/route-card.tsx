import { Card } from '@shared/ui/card'
import { routes } from '@shared/router'
import type { Route } from '@features/routes/api'

interface RouteCardProps {
  route: Route
}

export const RouteCard = ({ route }: RouteCardProps) => {
  const handleClick = () => {
    routes.routeView.navigate({
      params: { routeId: route.id },
      query: {},
    })
  }

  return (
    <Card
      padding="md"
      className="cursor-pointer transition hover:bg-secondary/50"
      onClick={handleClick}
    >
      <div className="space-y-3">
        <div className="flex items-start justify-between gap-3">
          <div className="flex-1 min-w-0">
            <h3 className="text-lg font-semibold truncate">{route.title}</h3>
            {route.description && (
              <p className="text-sm text-muted-foreground line-clamp-2 mt-1">{route.description}</p>
            )}
          </div>
          {route.cover_image_url && (
            <div className="h-20 w-20 flex-shrink-0 rounded-lg overflow-hidden bg-secondary/50">
              <img
                src={route.cover_image_url}
                alt={route.title}
                className="w-full h-full object-cover"
              />
            </div>
          )}
        </div>

        <div className="flex items-center gap-4 text-xs text-muted-foreground">
          {route.city && (
            <span className="flex items-center gap-1">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                className="h-4 w-4"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7z" />
                <circle cx="12" cy="9" r="2.5" />
              </svg>
              {route.city}
            </span>
          )}
          {route.total_distance_km && (
            <span className="flex items-center gap-1">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                className="h-4 w-4"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <path d="M12 2L2 7l10 5 10-5-10-5z" />
                <path d="M2 17l10 5 10-5M2 12l10 5 10-5" />
              </svg>
              {route.total_distance_km.toFixed(1)} км
            </span>
          )}
          {route.estimated_minutes && (
            <span className="flex items-center gap-1">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                viewBox="0 0 24 24"
                className="h-4 w-4"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
              >
                <circle cx="12" cy="12" r="10" />
                <path d="M12 6v6l4 2" />
              </svg>
              {route.estimated_minutes} мин
            </span>
          )}
        </div>

        {route.tags && route.tags.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {route.tags.slice(0, 3).map(tag => (
              <span
                key={tag}
                className="px-2 py-0.5 text-xs rounded-full bg-primary/20 text-primary"
              >
                {tag}
              </span>
            ))}
            {route.tags.length > 3 && (
              <span className="px-2 py-0.5 text-xs rounded-full bg-secondary/50 text-muted-foreground">
                +{route.tags.length - 3}
              </span>
            )}
          </div>
        )}
      </div>
    </Card>
  )
}
