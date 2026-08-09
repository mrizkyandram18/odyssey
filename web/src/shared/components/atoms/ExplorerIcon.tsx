import type { SVGProps } from 'react'
import type { Role } from '../../types'
import { Search, Hammer, Compass, HelpCircle } from 'lucide-react'

export interface ExplorerIconProps {
  role: Role
  size?: 'sm' | 'md' | 'lg'
}

export function ExplorerIcon({ role, size = 'md' }: ExplorerIconProps) {
  const sizes = { sm: 'h-6 w-6', md: 'h-8 w-8', lg: 'h-12 w-12' }
  const iconSizes = { sm: 14, md: 18, lg: 24 }
  
  const renderIcon = () => {
    const s = iconSizes[size]
    switch (role) {
      case 'SEEKER': return <Search size={s} className="text-accent-magic" />
      case 'BUILDER': return <Hammer size={s} className="text-accent-reward" />
      case 'GUIDE': return <Compass size={s} className="text-accent-nature" />
      default: return <HelpCircle size={s} />
    }
  }

  return (
    <div className={`flex items-center justify-center rounded-full bg-surface border border-border-subtle ${sizes[size]}`}>
      {renderIcon()}
    </div>
  )
}

export function Icon({ name, size = 16, ...props }: { name: string; size?: number } & SVGProps<SVGSVGElement>) {
  void name
  return <svg width={size} height={size} {...props} />
}
