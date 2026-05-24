import { Link, useLocation } from 'react-router-dom'
import type { ReactNode } from 'react'

interface LayoutProps {
  children: ReactNode
  title?: string
}

const navItems = [
  { path: '/', label: 'Dashboard' },
  { path: '/containers', label: 'Containers' },
  { path: '/logs', label: 'Logs Viewer' },
  { path: '/alerts', label: 'Alerts' },
  { path: '/ai-insights', label: 'AI Insights' },
]

export default function Layout({ children, title }: LayoutProps) {
  const location = useLocation()

  return (
    <div style={{ minHeight: '100vh', background: '#0f172a', color: '#e5e7eb', fontFamily: 'system-ui, sans-serif' }}>
      <nav style={{ background: 'linear-gradient(135deg, #020617 0%, #0f172a 100%)', borderBottom: '1px solid #1e293b', padding: '16px 24px', boxShadow: '0 2px 8px rgba(0, 0, 0, 0.3)' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', maxWidth: 1600, margin: '0 auto' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 32 }}>
            <h1 style={{ margin: 0, fontSize: 22, fontWeight: 700, background: 'linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%)', WebkitBackgroundClip: 'text', WebkitTextFillColor: 'transparent', backgroundClip: 'text' }}>
              🐳 DockerAI Monitor
            </h1>
            <div className="nav-links">
              {navItems.map((item) => (
                <Link
                  key={item.path}
                  to={item.path}
                  style={{
                    padding: '10px 16px',
                    borderRadius: 8,
                    textDecoration: 'none',
                    color: location.pathname === item.path ? '#ffffff' : '#94a3b8',
                    background: location.pathname === item.path ? 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)' : 'transparent',
                    fontSize: 14,
                    fontWeight: location.pathname === item.path ? 600 : 400,
                    transition: 'all 0.2s ease',
                    border: location.pathname === item.path ? 'none' : '1px solid transparent',
                  }}
                  onMouseEnter={(e) => {
                    if (location.pathname !== item.path) {
                      e.currentTarget.style.background = '#1e293b'
                      e.currentTarget.style.borderColor = '#334155'
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (location.pathname !== item.path) {
                      e.currentTarget.style.background = 'transparent'
                      e.currentTarget.style.borderColor = 'transparent'
                    }
                  }}
                >
                  {item.label}
                </Link>
              ))}
            </div>
          </div>
        </div>
      </nav>
      <main style={{ padding: 24, maxWidth: 1600, margin: '0 auto' }}>
        {title && <h2 style={{ margin: '0 0 24px 0', fontSize: 28, fontWeight: 700, color: '#f1f5f9', letterSpacing: '-0.5px' }}>{title}</h2>}
        {children}
      </main>
    </div>
  )
}
