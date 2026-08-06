import type { SVGProps } from 'react'
import type { Role } from '../../types'

const roleIcons: Record<Role, string> = {
  SEEKER: '🔍',
  BUILDER: '🛠️',
  GUIDE: '🧭',
}

export interface ExplorerIconProps {
  role: Role
  size?: 'sm' | 'md' | 'lg'
}

export function ExplorerIcon({ role, size = 'md' }: ExplorerIconProps) {
  const sizes = { sm: 'h-6 w-6', md: 'h-8 w-8', lg: 'h-12 w-12' }
  return (
    <div className={`flex items-center justify-center rounded-full bg-surface ${sizes[size]}`}>
      <span className="text-xl">{roleIcons[role] || '❓'}</span>
    </div>
  )
}

export function Icon({ name, size = 16, ...props }: { name: string; size?: number } & SVGProps<SVGSVGElement>) {
  void name
  return <svg width={size} height={size} {...props} />
}
