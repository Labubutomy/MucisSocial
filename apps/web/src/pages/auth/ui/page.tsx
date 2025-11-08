import { useState } from 'react'
import { AuthForm, AuthTabs, type AuthFormValues, type AuthMode } from '@features/auth'
import { AuthCard } from '@widgets/auth'
import { authFormSubmitted } from '@pages/auth/model'

const initialValues: AuthFormValues = {
  email: '',
  password: '',
}

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
  const [mode, setMode] = useState<AuthMode>('signIn')
  const [values, setValues] = useState<AuthFormValues>(initialValues)
  const submitLabel = mode === 'signIn' ? 'Войти' : 'Создать аккаунт'

  const handleModeChange = (nextMode: AuthMode) => {
    setMode(nextMode)
    setValues(initialValues)
  }

  const handleSubmit: React.FormEventHandler<HTMLFormElement> = event => {
    event.preventDefault()
    authFormSubmitted({ mode, values })
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background via-background/95 to-sidebar/30 px-4 py-16">
      <AuthCard
        header={{
          title: 'Добро пожаловать в музыкальную соцсеть',
          subtitle: 'С возвращением',
        }}
        tabs={<AuthTabs mode={mode} onModeChange={handleModeChange} />}
        form={
          <AuthForm
            mode={mode}
            values={values}
            onChange={setValues}
            onSubmit={handleSubmit}
            submitLabel={submitLabel}
          />
        }
        illustration={illustration}
      />
    </div>
  )
}
