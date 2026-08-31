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
  const base = 'inline-flex items-center justify-center rounded-xl font-bold transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-background disabled:opacity-50 disabled:pointer-events-none cursor-pointer border'

  const variants = {
    primary: 'bg-accent-magic text-white border-transparent hover:brightness-110 shadow-sm focus:ring-accent-magic',
    secondary: 'bg-surface-elevated text-text-primary border-border-subtle hover:bg-surface focus:ring-border-subtle',
    ghost: 'bg-transparent text-text-secondary border-transparent hover:text-text-primary hover:bg-black/5 focus:ring-border-subtle',
    danger: 'bg-accent-danger/10 text-accent-danger border-transparent hover:bg-accent-danger hover:text-white focus:ring-accent-danger',
  }

  const sizes = {
    sm: 'h-9 px-4 text-xs',
    md: 'h-11 px-5 text-sm',
    lg: 'h-12 px-6 text-sm',
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
