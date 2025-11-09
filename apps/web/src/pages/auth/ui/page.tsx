import { AuthForm, AuthTabs } from '@features/auth'
import { AuthCard } from '@widgets/auth'
import { $mode, $values, modeChanged, submitClicked, valuesChanged } from '@pages/auth/model'
import { $authError, $authPending } from '@features/auth/model'
import { useUnit } from 'effector-react'

const illustration = (
  <div className="flex flex-1 flex-col justify-between gap-6 rounded-3xl bg-gradient-to-br from-primary/60 via-accent/60 to-sidebar/60 p-8 text-primary-foreground shadow-2xl">
    <div className="space-y-4">
      <h2 className="text-3xl font-semibold">Почувствуйте ритм</h2>
      <p className="text-sm text-primary-foreground/80">
        Откройте подборки, подключайтесь к совместным плейлистам и следите, что слушают друзья прямо
        сейчас.
      </p>
    </div>
    <ul className="space-y-4 text-sm text-primary-foreground/90">
      <li className="flex items-start gap-3">
        <span className="mt-1 inline-flex h-2 w-2 flex-shrink-0 rounded-full bg-primary-foreground" />
        <span>Получайте еженедельные новинки, собранные специально для вашей компании.</span>
      </li>
      <li className="flex items-start gap-3">
        <span className="mt-1 inline-flex h-2 w-2 flex-shrink-0 rounded-full bg-primary-foreground" />
        <span>Проводите совместные прослушивания с мгновенными реакциями.</span>
      </li>
      <li className="flex items-start gap-3">
        <span className="mt-1 inline-flex h-2 w-2 flex-shrink-0 rounded-full bg-primary-foreground" />
        <span>Следите за кураторами и создавайте плейлисты под любое настроение.</span>
      </li>
    </ul>
  </div>
)

export const AuthPage = () => {
  const { mode, values, changeMode, changeValues, submit, pending, error } = useUnit({
    mode: $mode,
    values: $values,
    changeMode: modeChanged,
    changeValues: valuesChanged,
    submit: submitClicked,
    pending: $authPending,
    error: $authError,
  })
  const submitLabel = mode === 'signIn' ? 'Войти' : 'Создать аккаунт'

  const handleSubmit: React.FormEventHandler<HTMLFormElement> = event => {
    event.preventDefault()
    submit()
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background via-background/95 to-sidebar/30 px-4 py-16">
      <AuthCard
        header={{
          title: 'Добро пожаловать в музыкальную соцсеть',
          subtitle: 'С возвращением',
        }}
        tabs={<AuthTabs mode={mode} onModeChange={changeMode} />}
        form={
          <div className="space-y-4">
            <AuthForm
              mode={mode}
              values={values}
              onChange={changeValues}
              onSubmit={handleSubmit}
              submitLabel={submitLabel}
              loading={pending}
            />
            {error && (
              <p className="text-sm text-destructive" role="alert">
                {error}
              </p>
            )}
          </div>
        }
        illustration={illustration}
      />
    </div>
  )
}
