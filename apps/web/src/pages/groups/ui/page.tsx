import { useState } from 'react'
import { useUnit } from 'effector-react'
import { $user } from '@features/auth'
import { QueueList } from '@widgets/queue'
import { $queueContext, queueLoaded, queueContextSet } from '@features/queue'
import { Button } from '@shared/ui/button'
import { Input } from '@shared/ui/input'

export const GroupsPage = () => {
  const user = useUnit($user)
  const queueContext = useUnit($queueContext)
  const loadQueue = useUnit(queueLoaded)
  const [groupId, setGroupId] = useState('')
  const [currentGroupId, setCurrentGroupId] = useState<string | null>(null)

  const handleJoinGroup = () => {
    if (!groupId.trim() || !user) return

    const newContext = {
      type: 'group' as const,
      groupId: groupId.trim(),
    }

    // Set queue context
    queueContextSet(newContext)
    setCurrentGroupId(groupId.trim())
    loadQueue()
  }

  const handleLeaveGroup = () => {
    if (!user) return

    const newContext = {
      type: 'user' as const,
      userId: user.id,
    }

    queueContextSet(newContext)
    setCurrentGroupId(null)
    setGroupId('')
    loadQueue()
  }

  const isInGroup = currentGroupId !== null || queueContext?.type === 'group'

  return (
    <div className="page-container space-y-8 pb-16 pt-10">
      <div className="space-y-4">
        <h1 className="text-3xl font-bold">Группы</h1>
        <p className="text-muted-foreground">
          Создавайте группы для совместного прослушивания музыки
        </p>
      </div>

      <div className="rounded-3xl border border-border/60 bg-secondary/20 p-6 space-y-4">
        <div>
          <h2 className="text-lg font-semibold mb-2">
            {isInGroup ? 'Текущая группа' : 'Присоединиться к группе'}
          </h2>
          {isInGroup ? (
            <div className="space-y-4">
              <div className="rounded-2xl border border-border/60 bg-background/50 p-4">
                <p className="text-sm font-medium mb-1">
                  ID группы:{' '}
                  {currentGroupId || (queueContext?.type === 'group' ? queueContext.groupId : '')}
                </p>
                <p className="text-xs text-muted-foreground">
                  Вы можете поделиться этим ID с друзьями, чтобы они присоединились к группе
                </p>
              </div>
              <Button onClick={handleLeaveGroup} variant="outline">
                Покинуть группу
              </Button>
            </div>
          ) : (
            <div className="space-y-3">
              <Input
                type="text"
                placeholder="Введите ID группы"
                value={groupId}
                onChange={e => setGroupId(e.target.value)}
                onKeyDown={e => {
                  if (e.key === 'Enter') {
                    handleJoinGroup()
                  }
                }}
              />
              <Button onClick={handleJoinGroup} disabled={!groupId.trim() || !user}>
                Присоединиться
              </Button>
            </div>
          )}
        </div>
      </div>

      {isInGroup && (
        <div className="rounded-3xl border border-border/60 bg-secondary/20 p-6">
          <h2 className="text-lg font-semibold mb-4">Очередь группы</h2>
          <QueueList />
        </div>
      )}

      <div className="rounded-3xl border border-border/60 bg-secondary/20 p-6 space-y-4">
        <div>
          <h2 className="text-lg font-semibold mb-2">Создать новую группу</h2>
          <p className="text-sm text-muted-foreground mb-4">
            Создайте новую группу и поделитесь ID с друзьями
          </p>
          <Button
            onClick={() => {
              const newGroupId = `group-${Math.random().toString(36).slice(2, 10)}`
              setGroupId(newGroupId)
              handleJoinGroup()
            }}
            disabled={!user}
          >
            Создать группу
          </Button>
        </div>
      </div>
    </div>
  )
}
