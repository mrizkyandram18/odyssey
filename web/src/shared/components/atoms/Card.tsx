import { type ReactNode, forwardRef } from 'react'
import { motion, type HTMLMotionProps } from 'framer-motion'

export interface CardProps extends HTMLMotionProps<"div"> {
  children: ReactNode
  hoverable?: boolean
}

export const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ children, className = '', hoverable = false, ...props }, ref) => {
    const hoverStyles = hoverable ? 'cursor-pointer transition-all duration-300 hover:border-accent-reward/50 hover:shadow-lg hover:-translate-y-1' : ''
    return (
      <motion.div
        ref={ref}
        className={`bg-surface rounded-[24px] p-5 shadow-sm border border-border-subtle ${hoverStyles} ${className}`}
        whileHover={hoverable ? { y: -4 } : {}}
        whileTap={hoverable ? { scale: 0.98 } : {}}
        {...props}
      >
        {children}
      </motion.div>
    )
  }
)
Card.displayName = 'Card'
