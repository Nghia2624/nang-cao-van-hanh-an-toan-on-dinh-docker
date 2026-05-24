import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { format } from 'date-fns'
import Layout from '@/components/Layout'
import Card from '@/components/Card'
import { apiFetch } from '@/api/client'
import type { Alert } from '@/types/api'

export default function Alerts() {
  const { id } = useParams<{ id: string }>()
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [statusFilter, setStatusFilter] = useState<string>('')

  useEffect(() => {
    loadAlerts()
    const interval = setInterval(loadAlerts, 10000)
    return () => clearInterval(interval)
  }, [statusFilter])

  function loadAlerts() {
    const params = statusFilter ? `?status=${statusFilter}` : ''
    apiFetch<Alert[]>(`/api/v1/alerts${params}`)
      .then(setAlerts)
      .catch(console.error)
  }

  function updateAlertStatus(alertId: string, status: 'ACKNOWLEDGED' | 'RESOLVED') {
    apiFetch(`/api/v1/alerts/${alertId}/${status.toLowerCase()}`, { method: 'POST' })
      .then(() => loadAlerts())
      .catch(console.error)
  }

  const selectedAlert = id ? alerts.find((a) => a.id === id) : null

  if (selectedAlert) {
    return (
      <Layout title={`Alert: ${selectedAlert.title}`}>
        <div style={{ marginBottom: 16 }}>
          <Link to="/alerts" style={{ color: '#60a5fa', fontSize: 14 }}>
            ← Back to alerts
          </Link>
        </div>
        <Card>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', marginBottom: 16 }}>
            <div>
              <h2 style={{ margin: '0 0 8px 0', fontSize: 20 }}>{selectedAlert.title}</h2>
              <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
                <SeverityBadge severity={selectedAlert.severity} />
                <StatusBadge status={selectedAlert.status} />
              </div>
            </div>
            {(selectedAlert.status === 'NEW' || selectedAlert.status === 'ACKNOWLEDGED') && (
              <div style={{ display: 'flex', gap: 8 }}>
                {selectedAlert.status === 'NEW' && (
                <button
                  onClick={() => updateAlertStatus(selectedAlert.id, 'ACKNOWLEDGED')}
                  style={{
                    padding: '8px 16px',
                    borderRadius: 6,
                    background: '#3b82f6',
                    color: '#fff',
                    border: 'none',
                    cursor: 'pointer',
                    fontSize: 14,
                    fontWeight: 600,
                  }}
                >
                  Acknowledge
                </button>
                )}
                <button
                  onClick={() => updateAlertStatus(selectedAlert.id, 'RESOLVED')}
                  style={{
                    padding: '8px 16px',
                    borderRadius: 6,
                    background: '#10b981',
                    color: '#fff',
                    border: 'none',
                    cursor: 'pointer',
                    fontSize: 14,
                    fontWeight: 600,
                  }}
                >
                  Resolve
                </button>
              </div>
            )}
          </div>
          <div style={{ marginBottom: 16 }}>
            <div style={{ fontSize: 14, color: '#94a3b8', marginBottom: 8 }}>Description</div>
            <div style={{ color: '#e5e7eb' }}>{selectedAlert.description}</div>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
            <div>
              <div style={{ fontSize: 12, color: '#94a3b8' }}>Source</div>
              <div style={{ color: '#e5e7eb', marginTop: 4 }}>{selectedAlert.source}</div>
            </div>
            <div>
              <div style={{ fontSize: 12, color: '#94a3b8' }}>Created</div>
              <div style={{ color: '#e5e7eb', marginTop: 4 }}>
                {format(new Date(selectedAlert.created_at), 'PPpp')}
              </div>
            </div>
          </div>
        </Card>
      </Layout>
    )
  }

  // Backend already filters by status via query param, no need to double-filter
  const filtered = alerts

  return (
    <Layout title="Alerts">
      <div style={{ marginBottom: 16, display: 'flex', gap: 12, alignItems: 'center' }}>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          style={{
            padding: '8px 12px',
            borderRadius: 6,
            background: '#1e293b',
            border: '1px solid #334155',
            color: '#e5e7eb',
            fontSize: 14,
          }}
        >
          <option value="">All statuses</option>
          <option value="NEW">New</option>
          <option value="ACKNOWLEDGED">Acknowledged</option>
          <option value="RESOLVED">Resolved</option>
        </select>
        <button
          onClick={loadAlerts}
          style={{
            padding: '8px 16px',
            borderRadius: 6,
            background: '#3b82f6',
            color: '#fff',
            border: 'none',
            cursor: 'pointer',
            fontSize: 14,
            fontWeight: 600,
          }}
        >
          Refresh
        </button>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {filtered.length === 0 ? (
          <Card>
            <div style={{ color: '#94a3b8', textAlign: 'center', padding: 40 }}>
              <div style={{ fontSize: 48, marginBottom: 16 }}>🔔</div>
              <div style={{ fontSize: 16 }}>No alerts</div>
              <div style={{ fontSize: 12, color: '#64748b', marginTop: 8 }}>All systems operational</div>
            </div>
          </Card>
        ) : (
          filtered.map((a) => (
            <Link key={a.id} to={`/alerts/${a.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
              <Card>
                <div
                  style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'start',
                    transition: 'all 0.2s ease',
                  }}
                  onMouseEnter={(e) => {
                    e.currentTarget.style.transform = 'translateX(4px)'
                  }}
                  onMouseLeave={(e) => {
                    e.currentTarget.style.transform = 'translateX(0)'
                  }}
                >
                  <div style={{ flex: 1 }}>
                    <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 10 }}>
                      <h3 style={{ margin: 0, fontSize: 18, fontWeight: 700, color: '#f1f5f9' }}>{a.title}</h3>
                      <SeverityBadge severity={a.severity} />
                      <StatusBadge status={a.status} />
                    </div>
                    <div style={{ color: '#94a3b8', fontSize: 14, marginBottom: 10, lineHeight: 1.6 }}>{a.description}</div>
                    <div style={{ fontSize: 12, color: '#64748b', display: 'flex', gap: 16, alignItems: 'center' }}>
                      <span style={{ padding: '4px 8px', background: '#1e293b', borderRadius: 6 }}>{a.source}</span>
                      <span>{format(new Date(a.created_at), 'PPpp')}</span>
                    </div>
                  </div>
                </div>
              </Card>
            </Link>
          ))
        )}
      </div>
    </Layout>
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
    <span style={{ fontSize: 11, padding: '4px 10px', borderRadius: 999, background: `${colors[severity] || '#64748b'}22`, color: colors[severity] || '#64748b' }}>
      {severity}
    </span>
  )
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    NEW: '#3b82f6',
    ACKNOWLEDGED: '#f59e0b',
    RESOLVED: '#10b981',
  }
  return (
    <span style={{ fontSize: 11, padding: '4px 10px', borderRadius: 999, background: `${colors[status] || '#64748b'}22`, color: colors[status] || '#64748b' }}>
      {status}
    </span>
  )
}
