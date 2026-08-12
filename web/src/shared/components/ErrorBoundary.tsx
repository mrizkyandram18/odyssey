import { Component } from 'react'
import type { ErrorInfo, ReactNode } from 'react'

interface Props {
  children?: ReactNode
}

interface State {
  hasError: boolean
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  public static getDerivedStateFromError(_: Error): State {
    return { hasError: true }
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Uncaught error:', error, errorInfo)
  }

  public render() {
    if (this.state.hasError) {
      return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-bg-app p-4 text-center">
          <div className="bg-surface-elevated p-8 rounded-xl border border-border-subtle shadow-xl max-w-sm w-full">
            <h1 className="text-4xl mb-4">⚠️</h1>
            <h2 className="text-xl font-bold text-text-primary mb-2">Terjadi masalah saat membuka halaman.</h2>
            <p className="text-sm text-text-secondary mb-6">
              Sistem tidak dapat menampilkan halaman ini dengan benar.
            </p>
            <button
              onClick={() => window.location.reload()}
              className="w-full bg-accent-magic hover:bg-accent-magic/90 text-black font-bold py-3 px-4 rounded-lg transition-colors"
            >
              Muat Ulang
            </button>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
