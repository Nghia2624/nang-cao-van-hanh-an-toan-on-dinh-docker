import type { ReactNode } from 'react'

interface CardProps {
  title?: ReactNode
  children: ReactNode
  className?: string
}

export default function Card({ title, children, className }: CardProps) {
  return (
    <div
      style={{
        background: '#1e293b',
        border: '1px solid #334155',
        borderRadius: 12,
        padding: 20,
        boxShadow: '0 1px 3px rgba(0, 0, 0, 0.3)',
        transition: 'all 0.2s ease',
      }}
      className={className}
      onMouseEnter={(e) => {
        e.currentTarget.style.borderColor = '#475569'
        e.currentTarget.style.boxShadow = '0 4px 6px rgba(0, 0, 0, 0.4)'
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.borderColor = '#334155'
        e.currentTarget.style.boxShadow = '0 1px 3px rgba(0, 0, 0, 0.3)'
      }}
    >
      {title && (
        <h3 style={{ margin: '0 0 16px 0', fontSize: 18, fontWeight: 600, color: '#f1f5f9', borderBottom: '1px solid #334155', paddingBottom: 12 }}>
          {title}
        </h3>
      )}
      {children}
    </div>
  )
}
