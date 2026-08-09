import { createAvatar } from '@dicebear/core'
import { adventurer } from '@dicebear/collection'
import { useMemo } from 'react'

export interface AvatarProps {
  seed: string
  style?: string // defaults to adventurer for MVP
  size?: 'sm' | 'md' | 'lg' | 'xl' | '2xl'
  className?: string
}

const sizeClasses = {
  sm: 'h-8 w-8',
  md: 'h-12 w-12',
  lg: 'h-16 w-16',
  xl: 'h-24 w-24',
  '2xl': 'h-32 w-32'
}

export function Avatar({ seed, style = 'adventurer', size = 'md', className = '' }: AvatarProps) {
  // Use useMemo to avoid re-generating the SVG on every render if the seed hasn't changed.
  const dataUri = useMemo(() => {
    // We only support 'adventurer' style for now
    void style
    
    const avatar = createAvatar(adventurer, {
      seed: seed,
      backgroundColor: ['b6e3f4', 'c0aede', 'd1d4f9', 'ffd5dc', 'ffdfbf'],
      radius: 50, // slightly rounded if not fully rounded by parent, though parent uses rounded-full
    })
    
    return avatar.toDataUri()
  }, [seed, style])

  return (
    <img 
      src={dataUri} 
      alt={`Avatar ${seed}`} 
      className={`rounded-full shadow-sm object-cover bg-surface border border-border-subtle ${sizeClasses[size]} ${className}`} 
    />
  )
}
