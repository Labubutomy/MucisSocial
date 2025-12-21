import { useUnit } from 'effector-react'
import { QRCodeSVG } from 'qrcode.react'
import { Card } from '@shared/ui/card'
import { Button } from '@shared/ui/button'
import { $user } from '@features/auth/model'
import {
  $activeSession,
  $pendingRequests,
  $isLoading,
  createSessionClicked,
  stopSessionClicked,
  acceptRequestClicked,
  declineRequestClicked,
  refreshRequestsClicked,
  fetchIncomingRequestsFx,
} from '../model'

export const MusicRequestsPage = () => {
  const {
    user,
    activeSession,
    pendingRequests,
    isLoading,
    requestsPending,
    handleCreateSession,
    handleStopSession,
    handleAccept,
    handleDecline,
    handleRefresh,
  } = useUnit({
    user: $user,
    activeSession: $activeSession,
    pendingRequests: $pendingRequests,
    isLoading: $isLoading,
    requestsPending: fetchIncomingRequestsFx.pending,
    handleCreateSession: createSessionClicked,
    handleStopSession: stopSessionClicked,
    handleAccept: acceptRequestClicked,
    handleDecline: declineRequestClicked,
    handleRefresh: refreshRequestsClicked,
  })

  if (!user) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-20 pt-10">
        <p className="text-sm text-muted-foreground">Пожалуйста, войдите в систему</p>
      </div>
    )
  }

  const requestUrl = activeSession
    ? `${window.location.origin}/request/${activeSession.session_code}`
    : ''

  return (
    <div className="page-container space-y-8 pb-20 pt-10">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Заказы музыки</h1>
          <p className="text-muted-foreground">
            Создайте QR-код, чтобы получать заказы треков от слушателей
          </p>
        </div>
        <div className="flex items-center gap-2 rounded-full bg-secondary/50 px-4 py-2">
          <span className="text-sm text-muted-foreground">Монеты:</span>
          <span className="font-bold text-primary">{user.coins ?? 0}</span>
        </div>
      </div>

      {/* QR Code Section */}
      <Card padding="lg" className="space-y-6 bg-secondary/20">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold">
            {activeSession ? 'Активная сессия' : 'Создать сессию'}
          </h2>
          {activeSession ? (
            <div className="flex items-center gap-2">
              <span className="flex items-center gap-1 text-sm text-green-500">
                <span className="h-2 w-2 animate-pulse rounded-full bg-green-500" />
                Активна
              </span>
              <Button variant="destructive" size="sm" onClick={() => handleStopSession()}>
                Остановить
              </Button>
            </div>
          ) : null}
        </div>

        {activeSession ? (
          <div className="flex flex-col items-center gap-6 md:flex-row md:items-start">
            <div className="rounded-lg bg-white p-4">
              <QRCodeSVG value={requestUrl} size={200} level="H" includeMargin />
            </div>
            <div className="flex-1 space-y-4">
              <div>
                <p className="text-sm text-muted-foreground">Ссылка для заказа:</p>
                <code className="mt-1 block break-all rounded bg-secondary p-2 text-sm">
                  {requestUrl}
                </code>
              </div>
              <Button
                variant="outline"
                onClick={() => navigator.clipboard.writeText(requestUrl)}
              >
                Скопировать ссылку
              </Button>
              <p className="text-sm text-muted-foreground">
                Покажите этот QR-код или поделитесь ссылкой, чтобы слушатели могли заказать треки.
                Каждый заказ стоит 1 монету.
              </p>
            </div>
          </div>
        ) : (
          <div className="flex flex-col items-center gap-4 py-8">
            <p className="text-center text-muted-foreground">
              Создайте сессию, чтобы начать принимать заказы треков
            </p>
            <Button onClick={() => handleCreateSession()} disabled={isLoading}>
              {isLoading ? 'Создание...' : 'Создать сессию'}
            </Button>
          </div>
        )}
      </Card>

      {/* Incoming Requests */}
      <Card padding="lg" className="space-y-4 bg-secondary/20">
        <div className="flex items-center justify-between">
          <h2 className="text-xl font-semibold">
            Входящие заказы{' '}
            {pendingRequests.length > 0 && (
              <span className="ml-2 rounded-full bg-primary px-2 py-0.5 text-sm text-primary-foreground">
                {pendingRequests.length}
              </span>
            )}
          </h2>
          <Button variant="ghost" size="sm" onClick={() => handleRefresh()} disabled={requestsPending}>
            {requestsPending ? 'Обновление...' : 'Обновить'}
          </Button>
        </div>

        {pendingRequests.length === 0 ? (
          <p className="py-8 text-center text-muted-foreground">
            {activeSession
              ? 'Пока нет заказов. Поделитесь QR-кодом!'
              : 'Создайте сессию, чтобы начать принимать заказы'}
          </p>
        ) : (
          <div className="space-y-3">
            {pendingRequests.map(request => (
              <div
                key={request.id}
                className="flex items-center justify-between rounded-lg border border-border bg-card p-4"
              >
                <div className="space-y-1">
                  <p className="font-medium">Трек: {request.track_id}</p>
                  {request.message && (
                    <p className="text-sm text-muted-foreground">"{request.message}"</p>
                  )}
                  <p className="text-xs text-muted-foreground">
                    {new Date(request.created_at).toLocaleString('ru')}
                  </p>
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    className="text-red-500 hover:text-red-600"
                    onClick={() => handleDecline(request.id)}
                  >
                    Отклонить
                  </Button>
                  <Button
                    size="sm"
                    onClick={() => handleAccept(request.id)}
                  >
                    Принять (+1 🪙)
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}
      </Card>
    </div>
  )
}
