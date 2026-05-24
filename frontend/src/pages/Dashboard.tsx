 import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { format } from 'date-fns'
import Layout from '@/components/Layout'
import Card from '@/components/Card'
import Loading from '@/components/Loading'
import { apiFetch } from '@/api/client'
import type { Container, SystemMetrics, Alert, AIAnalysis } from '@/types/api'

export default function Dashboard() {
  const [containers, setContainers] = useState<Container[]>([])
  const [system, setSystem] = useState<SystemMetrics | null>(null)
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [latestAI, setLatestAI] = useState<AIAnalysis | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadData()
    const interval = setInterval(loadData, 20000)
    return () => clearInterval(interval)
  }, [])

  function loadData() {
    setError(null)
    Promise.allSettled([
      apiFetch<Container[]>('/api/v1/containers'),
      apiFetch<SystemMetrics>('/api/v1/metrics/system'),
      apiFetch<Alert[]>('/api/v1/alerts?limit=10'),
      apiFetch<AIAnalysis[]>('/api/v1/ai/analyses?limit=1'),
    ])
      .then(([c, s, a, ai]) => {
        setContainers(c.status === 'fulfilled' ? (c.value || []) : [])
        setSystem(s.status === 'fulfilled' ? s.value : null)
        setAlerts(a.status === 'fulfilled' ? (a.value || []) : [])
        setLatestAI(ai.status === 'fulfilled' && ai.value && ai.value.length > 0 ? ai.value[0] : null)
        setLoading(false)
      })
      .catch((err) => {
        setError(err.message || 'Failed to load data')
        setLoading(false)
        console.error(err)
      })
  }

  if (loading) {
    return (
      <Layout title="Dashboard Overview">
        <Loading message="Loading dashboard..." />
      </Layout>
    )
  }

  if (error) {
    return (
      <Layout title="Dashboard Overview">
        <Card>
          <div style={{ color: '#ef4444', textAlign: 'center', padding: 24 }}>
            <p>{error}</p>
            <button
              onClick={loadData}
              style={{
                marginTop: 16,
                padding: '8px 16px',
                borderRadius: 6,
                background: '#3b82f6',
                color: '#fff',
                border: 'none',
                cursor: 'pointer',
              }}
            >
              Retry
            </button>
          </div>
        </Card>
      </Layout>
    )
  }

  const runningCount = containers.filter((c) => {
    const s = c.status.toLowerCase()
    return s.includes('up') && !s.includes('paused')
  }).length
  const criticalAlerts = alerts.filter((a) => a.severity === 'CRITICAL' || a.severity === 'HIGH').length

  return (
    <Layout title="Dashboard Overview">
      <div className="grid-metrics">
        <MetricCard title="CPU Usage" value={system?.cpu_percent} unit="%" color="#3b82f6" />
        <MetricCard title="Memory Usage" value={system?.memory_percent} unit="%" color="#10b981" />
        <MetricCard title="Running Containers" value={runningCount} color="#f59e0b" />
        <MetricCard title="Active Alerts" value={alerts.length} critical={criticalAlerts} color="#ef4444" />
      </div>

      <div className="grid-sections">
        <Card title="Recent Containers">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {containers.slice(0, 5).map((c) => (
              <Link
                key={c.id}
                to={`/containers/${c.id}`}
                style={{
                  padding: 14,
                  borderRadius: 10,
                  background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)',
                  border: '1px solid #334155',
                  textDecoration: 'none',
                  color: 'inherit',
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  transition: 'all 0.2s ease',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = '#475569'
                  e.currentTarget.style.transform = 'translateY(-2px)'
                  e.currentTarget.style.boxShadow = '0 4px 12px rgba(0, 0, 0, 0.3)'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = '#334155'
                  e.currentTarget.style.transform = 'translateY(0)'
                  e.currentTarget.style.boxShadow = 'none'
                }}
              >
                <div>
                  <div style={{ fontWeight: 600 }}>{c.name}</div>
                  <div style={{ fontSize: 12, color: '#94a3b8', marginTop: 4 }}>{c.image}</div>
                </div>
                <StatusBadge status={c.status} />
              </Link>
            ))}
          </div>
          <div style={{ marginTop: 12 }}>
            <Link to="/containers" style={{ color: '#60a5fa', fontSize: 14 }}>
              View all containers →
            </Link>
          </div>
        </Card>

        <Card title="Critical Alerts">
          {alerts.filter((a) => a.severity === 'CRITICAL' || a.severity === 'HIGH').length === 0 ? (
            <div style={{ color: '#94a3b8', textAlign: 'center', padding: 24 }}>No critical alerts</div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              {alerts
                .filter((a) => a.severity === 'CRITICAL' || a.severity === 'HIGH')
                .slice(0, 5)
                .map((a) => (
                  <Link
                    key={a.id}
                    to={`/alerts/${a.id}`}
                    style={{
                      padding: 14,
                      borderRadius: 10,
                      background: a.severity === 'CRITICAL' ? 'linear-gradient(135deg, #7f1d1d 0%, #991b1b 100%)' : 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)',
                      border: `1px solid ${a.severity === 'CRITICAL' ? '#ef4444' : '#334155'}`,
                      textDecoration: 'none',
                      color: 'inherit',
                      transition: 'all 0.2s ease',
                    }}
                    onMouseEnter={(e) => {
                      e.currentTarget.style.transform = 'translateX(4px)'
                      e.currentTarget.style.boxShadow = '0 4px 12px rgba(0, 0, 0, 0.3)'
                    }}
                    onMouseLeave={(e) => {
                      e.currentTarget.style.transform = 'translateX(0)'
                      e.currentTarget.style.boxShadow = 'none'
                    }}
                  >
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
                      <div>
                        <div style={{ fontWeight: 600, color: a.severity === 'CRITICAL' ? '#ef4444' : '#f59e0b' }}>
                          {a.title}
                        </div>
                        <div style={{ fontSize: 12, color: '#94a3b8', marginTop: 4 }}>{a.description}</div>
                      </div>
                      <SeverityBadge severity={a.severity} />
                    </div>
                  </Link>
                ))}
            </div>
          )}
          <div style={{ marginTop: 12 }}>
            <Link to="/alerts" style={{ color: '#60a5fa', fontSize: 14 }}>
              View all alerts →
            </Link>
          </div>
        </Card>

        <Card title="Latest AI Insight">
          {latestAI ? (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start' }}>
                <div>
                  <div style={{ fontWeight: 600, color: '#f1f5f9' }}>Severity: {latestAI.severity_label}</div>
                  <div style={{ fontSize: 12, color: '#94a3b8', marginTop: 4 }}>
                    {format(new Date(latestAI.created_at), 'PPpp')}
                  </div>
                </div>
                <SeverityBadge severity={latestAI.severity_label} />
              </div>
              <div style={{ fontSize: 13, color: '#e5e7eb', marginBottom: 12, lineHeight: 1.6 }}>
                <strong>Root cause:</strong> {latestAI.root_cause}
              </div>
              <div style={{ fontSize: 13, color: '#94a3b8', marginBottom: 12 }}>
                <strong>Impact:</strong> {latestAI.impact_analysis}
              </div>
              <div style={{ marginTop: 12 }}>
                <Link
                  to="/ai-insights"
                  style={{
                    color: '#60a5fa',
                    fontSize: 14,
                    fontWeight: 600,
                    textDecoration: 'none',
                  }}
                >
                  View all insights →
                </Link>
              </div>
            </div>
          ) : (
            <div style={{ color: '#94a3b8', textAlign: 'center', padding: 24 }}>
              <div style={{ fontSize: 48, marginBottom: 16 }}>🤖</div>
              <div style={{ fontSize: 16 }}>No AI insights yet</div>
              <div style={{ fontSize: 12, color: '#64748b', marginTop: 8 }}>Start analyzing logs to see insights</div>
            </div>
          )}
        </Card>
      </div>
    </Layout>
  )
}

function MetricCard({ title, value, unit, color, critical }: { title: string; value?: number; unit?: string; color: string; critical?: number }) {
  const percentage = value != null ? Math.min(100, Math.max(0, value)) : 0
  return (
    <Card>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', marginBottom: 12 }}>
        <div style={{ fontSize: 13, color: '#94a3b8', fontWeight: 500, textTransform: 'uppercase', letterSpacing: '0.5px' }}>
          {title}
        </div>
        {critical != null && critical > 0 && (
          <span style={{ fontSize: 11, padding: '4px 8px', borderRadius: 12, background: '#ef444422', color: '#ef4444', fontWeight: 600 }}>
            {critical} Critical
          </span>
        )}
      </div>
      <div style={{ fontSize: 42, fontWeight: 800, color, marginBottom: 8, lineHeight: 1 }}>
        {value != null ? (Number.isInteger(value) ? value : value.toFixed(1)) : '--'}
        {unit && <span style={{ fontSize: 24, fontWeight: 600, marginLeft: 4 }}>{unit}</span>}
      </div>
      {value != null && unit === '%' && (
        <div style={{ marginTop: 12 }}>
          <div style={{ height: 6, background: '#1e293b', borderRadius: 3, overflow: 'hidden' }}>
            <div
              style={{
                height: '100%',
                width: `${percentage}%`,
                background: `linear-gradient(90deg, ${color} 0%, ${color}dd 100%)`,
                borderRadius: 3,
                transition: 'width 0.3s ease',
              }}
            />
          </div>
        </div>
      )}
    </Card>
  )
}

function StatusBadge({ status }: { status: string }) {
  const s = status.toLowerCase()
  const color = s.includes('paused') ? '#f59e0b' : s.includes('up') ? '#16a34a' : s.includes('restart') ? '#f59e0b' : s.includes('unhealthy') ? '#ef4444' : '#64748b'
  return (
    <span style={{ fontSize: 12, padding: '4px 10px', borderRadius: 999, background: `${color}22`, color }}>
      {status}
    </span>
  )
}

function SeverityBadge({ severity }: { severity: string }) {
  const colors: Record<string, string> = {
    CRITICAL: '#ef4444',
    HIGH: '#f59e0b',
    MEDIUM: '#3b82f6',
    LOW: '#94a3b8',
  }
  return (
    <span style={{ fontSize: 11, padding: '4px 8px', borderRadius: 999, background: `${colors[severity] || '#64748b'}22`, color: colors[severity] || '#64748b' }}>
      {severity}
    </span>
  )
}
