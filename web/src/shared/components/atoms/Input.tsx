import type { InputHTMLAttributes } from 'react'

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

export function Input({ label, error, className = '', ...props }: InputProps) {
  return (
    <div className="flex flex-col gap-1.5 w-full">
      {label && <label className="text-xs font-bold tracking-wide text-text-secondary pl-0.5">{label}</label>}
      <input
        className={`w-full rounded-xl border bg-surface-elevated px-4 py-3 text-sm text-text-primary placeholder:text-text-secondary/50 transition-colors focus:outline-none focus:ring-2 focus:ring-accent-magic focus:border-accent-magic ${error ? 'border-accent-danger focus:ring-accent-danger' : 'border-border-subtle hover:border-text-secondary/30'} ${className}`}
        {...props}
      />
      {error && <span className="text-xs text-accent-danger pl-0.5 font-medium">{error}</span>}
    </div>
  )
}
