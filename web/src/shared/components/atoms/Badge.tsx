import type { ReactNode } from 'react'

export interface BadgeProps {
  children?: ReactNode
  variant?: 'default' | 'primary' | 'secondary' | 'success' | 'warning' | 'error'
  size?: 'sm' | 'md'
}

export function Badge({ children, variant = 'default', size = 'md' }: BadgeProps) {
  const variants = {
    default: 'bg-surface text-muted-foreground',
    primary: 'bg-primary text-black',
    secondary: 'bg-secondary text-white',
    success: 'bg-success text-white',
    warning: 'bg-accent text-black',
    error: 'bg-error text-white',
  }

  const sizes = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-2.5 py-1 text-xs',
  }

  return (
    <span className={`inline-flex items-center rounded-full font-semibold ${variants[variant]} ${sizes[size]}`}>
      {children}
    </span>
  )
}
