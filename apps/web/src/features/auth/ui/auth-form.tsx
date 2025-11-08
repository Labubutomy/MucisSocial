import type { FormEventHandler } from 'react'
import type { AuthFormValues } from '@features/auth/model/types'
import { Input } from '@shared/ui/input'
import { Button } from '@shared/ui/button'

export interface AuthFormProps {
  values: AuthFormValues
  onChange: (values: AuthFormValues) => void
  onSubmit: FormEventHandler<HTMLFormElement>
  loading?: boolean
  submitLabel: string
  mode: 'signIn' | 'signUp'
}

export const AuthForm = ({ values, onChange, onSubmit, loading, submitLabel }: AuthFormProps) => {
  const updateField = (field: keyof AuthFormValues) => (value: string) =>
    onChange({ ...values, [field]: value })

  return (
    <form className="space-y-6" onSubmit={onSubmit}>
      <Input
        label="Электронная почта"
        type="email"
        name="email"
        placeholder="you@music.social"
        value={values.email}
        onChange={event => updateField('email')(event.target.value)}
        required
      />
      <Input
        label="Пароль"
        type="password"
        name="password"
        placeholder="Введите пароль"
        value={values.password}
        onChange={event => updateField('password')(event.target.value)}
        required
      />
      <div className="flex flex-col gap-3">
        <Button type="submit" size="lg" fullWidth loading={loading}>
          {submitLabel}
        </Button>
        <button
          type="button"
          className="self-center text-sm font-medium text-primary transition hover:text-primary/80"
        >
          Забыли пароль?
        </button>
      </div>
    </form>
  )
}
