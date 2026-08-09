import type { ReactNode } from 'react'
import { motion, type HTMLMotionProps } from 'framer-motion'

export interface ButtonProps extends HTMLMotionProps<"button"> {
  children: ReactNode
  variant?: 'primary' | 'secondary' | 'ghost' | 'danger'
  size?: 'sm' | 'md' | 'lg'
  isLoading?: boolean
}

export function Button({
  children,
  variant = 'primary',
  size = 'md',
  isLoading,
  className = '',
  type = 'button',
  ...props
}: ButtonProps) {
  const base = 'inline-flex items-center justify-center rounded-md font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-background disabled:opacity-50 disabled:pointer-events-none cursor-pointer border'

  const variants = {
    primary: 'bg-accent-magic text-black border-accent-magic hover:bg-accent-magic/90 hover:shadow-[0_0_15px_rgba(6,182,222,0.4)] focus:ring-accent-magic',
    secondary: 'bg-surface-elevated text-text-primary border-border-subtle hover:bg-surface-glass hover:border-accent-nature focus:ring-accent-nature',
    ghost: 'bg-transparent text-text-secondary border-transparent hover:text-text-primary hover:bg-surface-elevated focus:ring-border-subtle',
    danger: 'bg-accent-danger/20 text-accent-danger border-accent-danger hover:bg-accent-danger hover:text-white focus:ring-accent-danger',
  }

  const sizes = {
    sm: 'h-8 px-3 text-xs',
    md: 'h-11 px-5 py-2 text-sm',
    lg: 'h-14 px-8 text-base font-semibold',
  }

  return (
    <motion.button
      type={type}
      className={`${base} ${variants[variant]} ${sizes[size]} ${className}`}
      disabled={isLoading || props.disabled}
      whileHover={props.disabled || isLoading ? {} : { scale: 1.02 }}
      whileTap={props.disabled || isLoading ? {} : { scale: 0.95 }}
      {...props}
    >
      {isLoading && (
        <svg className="animate-spin -ml-1 mr-2 h-4 w-4" viewBox="0 0 24 24">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
        </svg>
      )}
      {children}
    </motion.button>
  )
}
