import { useUnit, useGate } from 'effector-react'
import { useParams } from 'atomic-router-react'
import { Card } from '@shared/ui/card'
import { Button } from '@shared/ui/button'
import { Input } from '@shared/ui/input'
import { $user } from '@features/auth/model'
import {
  RequestSessionGate,
  $session,
  $userCoins,
  $selectedTrackId,
  $message,
  $isLoading,
  $isSubmitting,
  $submitSuccess,
  $submitError,
  $sessionError,
  trackSelected,
  messageChanged,
  submitRequestClicked,
  resetForm,
} from '../model'

export const MusicRequestSessionPage = () => {
  const params = useParams<{ sessionCode: string }>()
  useGate(RequestSessionGate, { sessionCode: params?.sessionCode || '' })

  const {
    user,
    session,
    userCoins,
    selectedTrackId,
    message,
    isLoading,
    isSubmitting,
    submitSuccess,
    submitError,
    sessionError,
    handleTrackSelected,
    handleMessageChanged,
    handleSubmit,
    handleReset,
  } = useUnit({
    user: $user,
    session: $session,
    userCoins: $userCoins,
    selectedTrackId: $selectedTrackId,
    message: $message,
    isLoading: $isLoading,
    isSubmitting: $isSubmitting,
    submitSuccess: $submitSuccess,
    submitError: $submitError,
    sessionError: $sessionError,
    handleTrackSelected: trackSelected,
    handleMessageChanged: messageChanged,
    handleSubmit: submitRequestClicked,
    handleReset: resetForm,
  })

  if (!user) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-20 pt-10">
        <Card padding="lg" className="max-w-md bg-secondary/20 text-center">
          <h2 className="text-xl font-semibold">Войдите в систему</h2>
          <p className="mt-2 text-muted-foreground">
            Чтобы заказать трек, необходимо войти в аккаунт
          </p>
        </Card>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-20 pt-10">
        <p className="text-muted-foreground">Загрузка...</p>
      </div>
    )
  }

  if (sessionError || !session) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-20 pt-10">
        <Card padding="lg" className="max-w-md bg-secondary/20 text-center">
          <h2 className="text-xl font-semibold text-destructive">Сессия не найдена</h2>
          <p className="mt-2 text-muted-foreground">
            Возможно, сессия была закрыта или ссылка неверная
          </p>
        </Card>
      </div>
    )
  }

  if (submitSuccess) {
    return (
      <div className="page-container flex min-h-[60vh] items-center justify-center pb-20 pt-10">
        <Card padding="lg" className="max-w-md space-y-4 bg-secondary/20 text-center">
          <div className="text-5xl">🎵</div>
          <h2 className="text-xl font-semibold text-green-500">Заказ отправлен!</h2>
          <p className="text-muted-foreground">
            Ваш запрос на трек отправлен артисту. Ожидайте ответа!
          </p>
          <Button onClick={() => handleReset()}>Заказать ещё</Button>
        </Card>
      </div>
    )
  }

  const hasEnoughCoins = userCoins && userCoins.coins >= 1

  return (
    <div className="page-container mx-auto max-w-lg space-y-6 pb-20 pt-10">
      <div className="text-center">
        <h1 className="text-2xl font-bold">Заказать трек</h1>
        <p className="text-muted-foreground">Отправьте артисту заказ на любимую песню</p>
      </div>

      <Card padding="lg" className="bg-secondary/20">
        <div className="flex items-center justify-between">
          <span className="text-muted-foreground">Стоимость заказа:</span>
          <span className="font-bold">1 🪙</span>
        </div>
        <div className="mt-2 flex items-center justify-between">
          <span className="text-muted-foreground">Ваш баланс:</span>
          <span className={`font-bold ${hasEnoughCoins ? 'text-green-500' : 'text-red-500'}`}>
            {userCoins?.coins ?? 0} 🪙
          </span>
        </div>
      </Card>

      {!hasEnoughCoins && (
        <Card padding="lg" className="border-destructive bg-destructive/10">
          <p className="text-center text-destructive">
            Недостаточно монет для заказа. Пополните баланс!
          </p>
        </Card>
      )}

      <Card padding="lg" className="space-y-4 bg-secondary/20">
        <div>
          <label htmlFor="track" className="mb-2 block text-sm font-medium">
            ID трека *
          </label>
          <Input
            id="track"
            placeholder="Введите ID трека"
            value={selectedTrackId}
            onChange={e => handleTrackSelected(e.target.value)}
            disabled={isSubmitting || !hasEnoughCoins}
          />
          <p className="mt-1 text-xs text-muted-foreground">
            Найдите трек в приложении и скопируйте его ID
          </p>
        </div>

        <div>
          <label htmlFor="message" className="mb-2 block text-sm font-medium">
            Сообщение для артиста (необязательно)
          </label>
          <Input
            id="message"
            placeholder="Например: С днём рождения!"
            value={message}
            onChange={e => handleMessageChanged(e.target.value)}
            disabled={isSubmitting || !hasEnoughCoins}
            maxLength={200}
          />
        </div>

        {submitError && (
          <p className="text-center text-sm text-destructive">{submitError}</p>
        )}

        <Button
          className="w-full"
          onClick={() => handleSubmit()}
          disabled={isSubmitting || !selectedTrackId || !hasEnoughCoins}
        >
          {isSubmitting ? 'Отправка...' : 'Отправить заказ (1 🪙)'}
        </Button>
      </Card>
    </div>
  )
}
