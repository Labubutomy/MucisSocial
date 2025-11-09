import type { AuthMode } from '@features/auth/model/types'
import { Tabs } from '@shared/ui/tabs'

export interface AuthTabsProps {
  mode: AuthMode
  onModeChange: (mode: AuthMode) => void
}

export const AuthTabs = ({ mode, onModeChange }: AuthTabsProps) => (
  <Tabs
    value={mode}
    onChange={value => onModeChange(value as AuthMode)}
    items={[
      { value: 'signIn', label: 'Вход' },
      { value: 'signUp', label: 'Регистрация' },
    ]}
    className="w-full"
  />
)
