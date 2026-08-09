import { type ReactNode, forwardRef } from 'react'
import { motion, type HTMLMotionProps } from 'framer-motion'

export interface CardProps extends HTMLMotionProps<"div"> {
  children: ReactNode
  hoverable?: boolean
}

export const Card = forwardRef<HTMLDivElement, CardProps>(
  ({ children, className = '', hoverable = false, ...props }, ref) => {
    const hoverStyles = hoverable ? 'cursor-pointer transition-all duration-300 hover:border-accent-nature/50 hover:shadow-[0_4px_20px_rgba(16,185,129,0.1)] hover:-translate-y-1' : ''
    return (
      <motion.div
        ref={ref}
        className={`glass-panel rounded-xl p-5 ${hoverStyles} ${className}`}
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
