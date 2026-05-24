import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import Layout from '@/components/Layout'
import Card from '@/components/Card'
import Loading from '@/components/Loading'
import { apiFetch } from '@/api/client'
import type { Container, ContainerPrediction } from '@/types/api'

export default function Containers() {
  const [containers, setContainers] = useState<Container[]>([])
  const [predictions, setPredictions] = useState<Record<string, ContainerPrediction | null>>({})
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('')

  useEffect(() => {
    loadContainers()
    const interval = setInterval(loadContainers, 20000) // Polling less frequently
    return () => clearInterval(interval)
  }, [])

  async function loadContainers() {
    if (!loading) setLoading(true)
    try {
      const items = await apiFetch<Container[]>('/api/v1/containers')
      setContainers(items || [])

      // Fetch predictions in the background (throttled)
      const ids = (items || []).map((c) => c.id)
      ids.forEach((cid, idx) => {
        setTimeout(() => {
          apiFetch<ContainerPrediction>(`/api/v1/containers/${cid}/prediction`)
            .then((p) => setPredictions((prev) => ({ ...prev, [cid]: p })))
            .catch(() => setPredictions((prev) => ({ ...prev, [cid]: null })))
        }, idx * 150) // slightly faster throttle
      })
    } catch (err) {
      console.error(err)
    } finally {
      setLoading(false)
    }
  }

  const filtered = containers.filter((c) => c.name.toLowerCase().includes(filter.toLowerCase()) || c.image.toLowerCase().includes(filter.toLowerCase()))

  if (loading && containers.length === 0) {
    return (
      <Layout title="Containers">
        <Loading message="Loading containers..." />
      </Layout>
    )
  }

  return (
    <Layout title="Containers">
      <div style={{ marginBottom: 16, display: 'flex', gap: 12, alignItems: 'center' }}>
        <input
          type="text"
          placeholder="Search containers..."
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          style={{
            padding: '8px 12px',
            borderRadius: 8,
            background: '#1e293b',
            border: '1px solid #334155',
            color: '#e5e7eb',
            fontSize: 14,
            flex: 1,
            maxWidth: 400,
          }}
        />
        <button
          onClick={loadContainers}
          disabled={loading}
          style={{
            padding: '8px 16px',
            borderRadius: 8,
            background: '#3b82f6',
            color: '#fff',
            border: 'none',
            cursor: loading ? 'not-allowed' : 'pointer',
            fontSize: 14,
            fontWeight: 600,
            opacity: loading ? 0.6 : 1,
          }}
        >
          {loading ? 'Refreshing...' : 'Refresh'}
        </button>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: 16 }}>
        {filtered.map((c) => (
          <ContainerCard key={c.id} container={c} prediction={predictions[c.id]} />
        ))}
      </div>
    </Layout>
  )
}

function ContainerCard({ container, prediction }: { container: Container; prediction?: ContainerPrediction | null }) {
  return (
    <Link to={`/containers/${container.id}`} style={{ textDecoration: 'none', color: 'inherit' }}>
            <Card>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', marginBottom: 12 }}>
          <div style={{ fontWeight: 700, fontSize: 18, color: '#f1f5f9', wordBreak: 'break-all' }}>{container.name}</div>
          <StatusBadge status={container.status} />
        </div>
        <div style={{ fontSize: 12, color: '#94a3b8', marginBottom: 12, fontFamily: 'monospace', wordBreak: 'break-all' }}>
          {container.image}
              </div>

              <div style={{ display: 'flex', gap: 20, fontSize: 13, color: '#64748b', paddingTop: 12, borderTop: '1px solid #334155' }}>
                <div>
                  <div style={{ fontSize: 11, color: '#94a3b8', marginBottom: 4 }}>ID</div>
            <div style={{ fontSize: 14, fontWeight: 600, color: '#e5e7eb', fontFamily: 'monospace' }}>{container.id.substring(0, 12)}</div>
                </div>
                <div>
                  <div style={{ fontSize: 11, color: '#94a3b8', marginBottom: 4 }}>Uptime</div>
            <div style={{ fontSize: 16, fontWeight: 600, color: '#e5e7eb' }}>{formatUptime(container.uptime_sec)}</div>
          </div>
                </div>

        {prediction && (
          <div style={{ marginTop: 12, paddingTop: 12, borderTop: '1px solid #334155' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <RiskBadge level={prediction.risk_level} score={prediction.risk_score} />
              <div>
                <div style={{ fontSize: 11, color: '#94a3b8', textAlign: 'right', marginBottom: 4 }}>ETA to Failure</div>
                <div style={{ fontSize: 14, fontWeight: 600, color: '#e5e7eb' }}>{formatETA(prediction.eta_minutes)}</div>
              </div>
            </div>
          </div>
        )}
            </Card>
          </Link>
  )
}

function StatusBadge({ status }: { status: string }) {
  const s = status.toLowerCase()
  const color = s.includes('paused') ? '#f59e0b' : s.includes('up') ? '#16a34a' : s.includes('restart') ? '#f59e0b' : s.includes('unhealthy') ? '#ef4444' : '#64748b'
  return (
    <span
      style={{
        fontSize: 11,
        padding: '4px 10px',
        borderRadius: 999,
        background: `${color}22`,
        color,
        flexShrink: 0,
        marginLeft: 8,
      }}
    >
      {status}
    </span>
  )
}

function RiskBadge({ level, score }: { level: 'low' | 'medium' | 'high'; score: number }) {
  const colors = {
    low: '#10b981',
    medium: '#f59e0b',
    high: '#ef4444',
  }
  const color = colors[level] || '#64748b'
  return (
    <div>
      <div style={{ fontSize: 11, color: '#94a3b8', marginBottom: 4 }}>Risk</div>
      <span style={{ fontSize: 12, padding: '4px 10px', borderRadius: 999, background: `${color}22`, color, fontWeight: 600 }}>
        {level.toUpperCase()} ({score})
      </span>
    </div>
  )
}

function formatUptime(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  if (seconds < 86400) {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return `${h}h ${m}m`
  }
  const d = Math.floor(seconds / 86400)
  const h = Math.floor((seconds % 86400) / 3600)
  return `${d}d ${h}h`
}

function formatETA(minutes: number): string {
  if (minutes < 0) return 'N/A'
  if (minutes < 60) return `~${minutes}m`
  if (minutes < 1440) return `~${(minutes / 60).toFixed(1)}h`
  return `~${(minutes / 1440).toFixed(1)}d`
}
