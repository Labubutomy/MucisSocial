import { IconButton } from '@shared/ui/icon-button'
import { trackAddedToQueue } from '@features/queue'
import { useUnit } from 'effector-react'

export interface AddToQueueButtonProps {
  trackId: string
  className?: string
}

export const AddToQueueButton = ({ trackId, className }: AddToQueueButtonProps) => {
  const addToQueue = useUnit(trackAddedToQueue)

  return (
    <IconButton
      aria-label="Добавить в очередь"
      onClick={() => addToQueue(trackId)}
      variant="ghost"
      size="sm"
      className={className}
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        className="h-5 w-5"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
      >
        <path d="M12 5v14M5 12h14" />
      </svg>
    </IconButton>
  )
}
