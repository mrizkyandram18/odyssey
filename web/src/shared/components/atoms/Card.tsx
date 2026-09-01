import { type ReactNode, forwardRef } from 'react'
import { motion, type HTMLMotionProps } from 'framer-motion'

export interface CardProps extends HTMLMotionProps<"div"> {
  children: ReactNode
  hoverable?: boolean
}

export const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ children, className = '', hoverable = false, ...props }, ref) => {
    const hoverStyles = hoverable ? 'cursor-pointer transition-all duration-200 hover:border-accent-magic/25 hover:shadow-sm' : ''
    return (
      <motion.div
        ref={ref}
        className={`bg-surface rounded-2xl p-5 shadow-sm border border-border-subtle ${hoverStyles} ${className}`}
        whileHover={hoverable ? { y: -1 } : {}}
        whileTap={hoverable ? { scale: 0.99 } : {}}
        {...props}
      >
        {children}
      </motion.div>
    )
  }
)
Card.displayName = 'Card'
