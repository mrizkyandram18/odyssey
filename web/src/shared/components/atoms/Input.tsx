import type { InputHTMLAttributes } from 'react'

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

export function Input({ label, error, className = '', ...props }: InputProps) {
  return (
    <div className="flex flex-col gap-1.5 w-full">
      {label && <label className="text-sm font-medium text-text-secondary pl-1">{label}</label>}
      <input
        className={`w-full rounded-md border bg-surface-elevated px-4 py-3 text-base text-text-primary placeholder:text-text-secondary/50 transition-colors focus:outline-none focus:ring-2 focus:ring-accent-magic ${error ? 'border-accent-danger focus:ring-accent-danger' : 'border-border-subtle hover:border-text-secondary/50'} ${className}`}
        {...props}
      />
      {error && <span className="text-sm text-accent-danger pl-1 font-medium">{error}</span>}
    </div>
  )
}
