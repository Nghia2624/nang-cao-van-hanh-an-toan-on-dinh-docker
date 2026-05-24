import { Component, ErrorInfo, ReactNode } from 'react'

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error?: Error
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      return (
        this.props.fallback || (
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: 'center',
              justifyContent: 'center',
              minHeight: '400px',
              padding: 24,
              color: '#e5e7eb',
            }}
          >
            <h2 style={{ color: '#ef4444', marginBottom: 16 }}>Something went wrong</h2>
            <p style={{ color: '#94a3b8', marginBottom: 24, textAlign: 'center' }}>
              {this.state.error?.message || 'An unexpected error occurred'}
            </p>
            <button
              onClick={() => window.location.reload()}
              style={{
                padding: '10px 20px',
                borderRadius: 8,
                background: '#3b82f6',
                color: '#fff',
                border: 'none',
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: 600,
              }}
            >
              Reload Page
            </button>
          </div>
        )
      )
    }

    return this.props.children
  }
}
