import type { ReactNode } from 'react'

export interface BadgeProps {
  children?: ReactNode
  variant?: 'default' | 'primary' | 'secondary' | 'success' | 'warning' | 'error'
  size?: 'sm' | 'md'
}

export function Badge({ children, variant = 'default', size = 'md' }: BadgeProps) {
  const variants = {
    default: 'bg-surface-elevated border border-border-subtle text-text-secondary',
    primary: 'bg-accent-magic text-white',
    secondary: 'bg-accent-rare text-white',
    success: 'bg-status-success text-white',
    warning: 'bg-accent-gold text-white',
    error: 'bg-accent-danger text-white',
  }

  const sizes = {
    sm: 'px-2 py-0.5 text-[11px]',
    md: 'px-2.5 py-1 text-xs',
  }

  return (
    <span className={`inline-flex items-center rounded-full font-bold ${variants[variant]} ${sizes[size]}`}>
      {children}
    </span>
  )
}
