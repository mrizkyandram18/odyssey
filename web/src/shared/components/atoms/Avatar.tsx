import { createAvatar } from '@dicebear/core'
import { adventurer } from '@dicebear/collection'
import { useMemo } from 'react'
import { motion } from 'framer-motion'

export interface AvatarProps {
  seed: string
  style?: string // defaults to adventurer for MVP
  /** Slice 2.2: equipped frame (e.g. gold). Free default is none. */
  frame?: string
  /** Slice 4.5: equipped explorer effect (e.g. sparkle). */
  effect?: string
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

const frameClasses: Record<string, string> = {
  none: 'border border-border-subtle',
  gold: 'border-2 border-amber-400 shadow-[0_0_12px_rgba(251,191,36,0.55)] ring-2 ring-amber-300/40',
}

const effectVariants: Record<string, any> = {
  none: {},
  sparkle: {
    animate: { scale: [1, 1.06, 1] },
    transition: { repeat: Infinity, duration: 1.5, ease: 'easeInOut' },
  },
  float: {
    animate: { y: [-4, 4, -4] },
    transition: { repeat: Infinity, duration: 2, ease: 'easeInOut' },
  },
  trail: {
    animate: { rotate: [-2, 2, -2], y: [-2, 2, -2] },
    transition: { repeat: Infinity, duration: 2.5, ease: 'easeInOut' },
  },
}

export function Avatar({ seed, style = 'adventurer', frame = 'none', effect = 'none', size = 'md', className = '' }: AvatarProps) {
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

  const frameClass = frameClasses[frame] || frameClasses.none
  const effectVariant = effectVariants[effect] || effectVariants.none
  const hasEffect = effect !== 'none' && effectVariant.animate !== undefined

  return (
    <motion.span
      className={`inline-block ${sizeClasses[size]} ${className}`}
      {...(hasEffect ? effectVariant : {})}
    >
      <img
        src={dataUri}
        alt={`Avatar ${seed}`}
        className={`rounded-full shadow-sm object-cover bg-surface ${frameClass}`}
        data-avatar-frame={frame}
        data-avatar-effect={effect}
      />
    </motion.span>
  )
}
