import { useEffect, useRef } from 'react'
import { useUnit } from 'effector-react'
import { routes } from '@shared/router'
import { cn } from '@shared/lib/cn'
import { Avatar } from '@shared/ui/avatar'
import {
  $query,
  $showDropdown,
  $suggestionSeeds,
  $suggestions,
  focusChanged,
  hoverChanged,
  queryChanged,
  searchSubmitted,
  suggestionSelected,
} from './model'

export const AppHeader = () => {
  const blurTimeoutRef = useRef<number | null>(null)

  const {
    query,
    suggestions,
    suggestionSeeds,
    showDropdown,
    navigateToHome,
    navigateToProfile,
    changeQuery,
    setFocus,
    setHover,
    submitSearch,
    selectSuggestion,
  } = useUnit({
    query: $query,
    suggestions: $suggestions,
    suggestionSeeds: $suggestionSeeds,
    showDropdown: $showDropdown,
    navigateToHome: routes.home.navigate,
    navigateToProfile: routes.profile.navigate,
    changeQuery: queryChanged,
    setFocus: focusChanged,
    setHover: hoverChanged,
    submitSearch: searchSubmitted,
    selectSuggestion: suggestionSelected,
  })

  const navigateToSearch = useUnit(routes.search.navigate)

  useEffect(
    () => () => {
      if (blurTimeoutRef.current) {
        window.clearTimeout(blurTimeoutRef.current)
      }
    },
    []
  )

  const handleSubmit: React.FormEventHandler<HTMLFormElement> = event => {
    event.preventDefault()
    submitSearch()
  }

  const handleSuggestionClick = (value: string) => {
    selectSuggestion(value)
  }

  const handleInputFocus = () => {
    if (blurTimeoutRef.current) {
      window.clearTimeout(blurTimeoutRef.current)
    }
    setFocus(true)
  }

  const handleInputBlur = () => {
    blurTimeoutRef.current = window.setTimeout(() => {
      setFocus(false)
    }, 120)
  }

  return (
    <div className="sticky top-0 z-40 border-b border-border/60 bg-background/80 backdrop-blur">
      <div className="page-container flex items-center gap-4 py-4">
        <button
          type="button"
          onClick={() => navigateToHome({ params: {}, query: {} })}
          className="flex h-12 w-12 items-center justify-center rounded-full border border-border/60 bg-secondary/40 text-muted-foreground transition hover:border-primary hover:bg-secondary/60 hover:text-foreground md:hidden"
          aria-label="На главную"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            viewBox="0 0 24 24"
            className="h-5 w-5"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.5"
          >
            <path d="m3 9 9-6 9 6" />
            <path d="M4 10v9a1 1 0 0 0 1 1h4v-6h6v6h4a1 1 0 0 0 1-1v-9" />
          </svg>
        </button>
        <button
          type="button"
          onClick={() => navigateToHome({ params: {}, query: {} })}
          className="hidden items-center gap-2 rounded-full border border-transparent px-4 py-2 text-sm font-semibold uppercase tracking-[0.3em] text-muted-foreground transition hover:border-primary hover:text-foreground md:flex"
        >
          <span className="h-2 w-2 rounded-full bg-primary" />
          <span>Music Social</span>
        </button>

        <form onSubmit={handleSubmit} className="relative flex-1">
          <div className="flex items-center gap-3 rounded-full border border-border/60 bg-secondary/30 px-5 py-3 transition focus-within:border-primary focus-within:bg-secondary/50">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              className="h-5 w-5 text-muted-foreground"
              fill="none"
              stroke="currentColor"
              strokeWidth="1.5"
            >
              <path d="m21 21-4.35-4.35" />
              <circle cx={11} cy={11} r={8} />
            </svg>
            <input
              value={query}
              onChange={event => changeQuery(event.target.value)}
              onFocus={handleInputFocus}
              onBlur={handleInputBlur}
              placeholder="Поиск по трекам, артистам, плейлистам"
              className="flex-1 bg-transparent text-base text-foreground placeholder:text-muted-foreground focus:outline-none"
            />
          </div>
          {showDropdown && (
            <div
              className="absolute left-0 right-0 mt-3 rounded-2xl border border-border/60 bg-background shadow-xl shadow-black/20"
              onMouseEnter={() => setHover(true)}
              onMouseLeave={() => setHover(false)}
            >
              <ul className="flex flex-col divide-y divide-border/60">
                {suggestions.map(item => (
                  <li key={item}>
                    <button
                      type="button"
                      onClick={() => handleSuggestionClick(item)}
                      className="flex w-full items-center justify-between px-4 py-3 text-left text-sm text-muted-foreground transition hover:bg-secondary/40 hover:text-foreground"
                    >
                      <span>{item}</span>
                      <span className="text-xs uppercase text-muted-foreground/80">Найти</span>
                    </button>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </form>
        <button
          type="button"
          onClick={() => navigateToProfile({ params: {}, query: {} })}
          className={cn(
            'flex h-12 w-12 items-center justify-center rounded-full border border-border/60 bg-secondary/40 transition hover:border-primary hover:bg-secondary/60'
          )}
        >
          <Avatar fallback="AW" size="sm" />
        </button>
      </div>
    </div>
  )
}
