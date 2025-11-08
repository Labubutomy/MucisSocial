import type { UserProfile } from '@entities/user/model/types'
import { Avatar } from '@shared/ui/avatar'
import { cn } from '@shared/lib/cn'

export interface ProfileHeaderProps {
  user: UserProfile
  className?: string
}

export const ProfileHeader = ({ user, className }: ProfileHeaderProps) => (
  <section
    className={cn(
      'flex flex-col items-center gap-6 rounded-3xl border border-border/60 bg-secondary/30 p-6 text-center md:flex-row md:justify-center md:p-8',
      className
    )}
  >
    <Avatar
      src={user.avatarUrl}
      fallback={user.username}
      size="xl"
      className="shadow-xl shadow-black/30"
    />
    <div className="max-w-xl space-y-3 md:text-center">
      <p className="text-sm uppercase tracking-[0.35em] text-primary">Профиль</p>
      <h1 className="text-3xl font-semibold md:text-4xl">{user.username}</h1>
      <p className="text-sm text-muted-foreground md:text-base">
        {user.musicTasteSummary.topGenres.slice(0, 3).join(' • ')}
      </p>
    </div>
  </section>
)
