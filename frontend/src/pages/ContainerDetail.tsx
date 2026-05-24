import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Area, AreaChart } from 'recharts'
import { format } from 'date-fns'
import Layout from '@/components/Layout'
import Card from '@/components/Card'
import { apiFetch, sseUrl } from '@/api/client'
import { useSSE } from '@/hooks/useSSE'
import type { Container, LogEntry, MetricPoint } from '@/types/api'

export default function ContainerDetail() {
  const { id } = useParams<{ id: string }>()
  const [container, setContainer] = useState<Container | null>(null)
  const [cpuMetrics, setCpuMetrics] = useState<MetricPoint[]>([])
  const [memMetrics, setMemMetrics] = useState<MetricPoint[]>([])
  const [logs, setLogs] = useState<LogEntry[]>([])

  useEffect(() => {
    if (!id) return
    loadContainer()
    loadMetrics()
    loadRecentLogs()
  }, [id])

  function loadContainer() {
    if (!id) return
    apiFetch<Container>(`/api/v1/containers/${id}`)
      .then(setContainer)
      .catch(console.error)
  }

  function loadMetrics() {
    if (!id) return
    const now = Date.now()
    const from = now - 600000 // 10 minutes
    Promise.all([
      apiFetch<MetricPoint[]>(`/api/v1/metrics/container/${id}/cpu?from=${from}&to=${now}`),
      apiFetch<MetricPoint[]>(`/api/v1/metrics/container/${id}/memory?from=${from}&to=${now}`),
    ])
      .then(([cpu, mem]) => {
        setCpuMetrics(cpu || [])
        setMemMetrics(mem || [])
      })
      .catch(console.error)
  }

  function loadRecentLogs() {
    if (!id) return
    const now = Date.now()
    const from = now - 3600000 // 1 hour
    apiFetch<LogEntry[]>(`/api/v1/logs?container=${id}&limit=200&from=${from}&to=${now}`)
      .then((items) => setLogs(items || []))
      .catch(console.error)
  }

  useSSE(
    id ? sseUrl('/api/v1/logs/stream', { container: id }) : null,
    {
      enabled: !!id,
      onMessage: (data) => {
        try {
          const entry: LogEntry = JSON.parse(data)
          setLogs((prev) => {
            // Dedup: skip if same timestamp+message already exists
            const isDup = prev.some(
              (p) => p.timestamp === entry.timestamp && p.message === entry.message
            )
            if (isDup) return prev
            return [entry, ...prev].slice(0, 500)
          })
        } catch {
          // ignore invalid JSON
        }
      },
    }
  )

  if (!container) {
    return (
      <Layout title="Container Detail">
        <div>Loading...</div>
      </Layout>
    )
  }

  function parseMetricTimestamp(ts: string | number): number {
    if (typeof ts === 'number') {
      return ts > 1000000000000 ? ts : ts * 1000
    }
    return new Date(ts).getTime()
  }

  const cpuChartData = cpuMetrics.map((p) => ({
    time: format(new Date(parseMetricTimestamp(p.timestamp)), 'HH:mm:ss'),
    cpu: p.value || 0,
  }))

  const memChartData = memMetrics.map((p) => ({
    time: format(new Date(parseMetricTimestamp(p.timestamp)), 'HH:mm:ss'),
    memory: p.value || 0,
  }))

  return (
    <Layout title={`Container: ${container.name}`}>
      <div style={{ marginBottom: 16 }}>
        <Link to="/containers" style={{ color: '#60a5fa', fontSize: 14 }}>
          ← Back to containers
        </Link>
      </div>

      <div className="grid-stats">
        <Card>
          <div style={{ fontSize: 12, color: '#94a3b8' }}>Status</div>
          <div style={{ fontSize: 20, fontWeight: 600, marginTop: 8 }}>{container.status}</div>
        </Card>
        <Card>
          <div style={{ fontSize: 12, color: '#94a3b8' }}>Restarts</div>
          <div style={{ fontSize: 20, fontWeight: 600, marginTop: 8 }}>{container.restart_count}</div>
        </Card>
        <Card>
          <div style={{ fontSize: 12, color: '#94a3b8' }}>Uptime</div>
          <div style={{ fontSize: 20, fontWeight: 600, marginTop: 8 }}>{formatUptime(container.uptime_sec)}</div>
        </Card>
      </div>

      <div className="grid-charts">
        <Card title="CPU Usage (10m)">
          <ResponsiveContainer width="100%" height={200}>
            <AreaChart data={cpuChartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="time" stroke="#94a3b8" fontSize={12} />
              <YAxis stroke="#94a3b8" fontSize={12} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: 8 }} />
              <Area type="monotone" dataKey="cpu" stroke="#3b82f6" fill="#3b82f6" fillOpacity={0.3} />
            </AreaChart>
          </ResponsiveContainer>
        </Card>

        <Card title="Memory Usage (10m)">
          <ResponsiveContainer width="100%" height={200}>
            <AreaChart data={memChartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="time" stroke="#94a3b8" fontSize={12} />
              <YAxis stroke="#94a3b8" fontSize={12} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: 8 }} />
              <Area type="monotone" dataKey="memory" stroke="#10b981" fill="#10b981" fillOpacity={0.3} />
            </AreaChart>
          </ResponsiveContainer>
        </Card>
      </div>

      <Card title={`Logs (${logs.length}) - Real-time`}>
        <div style={{ maxHeight: 500, overflow: 'auto', background: '#020617', borderRadius: 8, padding: 12 }}>
          {logs.length === 0 ? (
            <div style={{ color: '#94a3b8', textAlign: 'center', padding: 24 }}>No logs yet</div>
          ) : (
            [...logs].sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()).map((log, i) => (
              <LogLine key={i} log={log} />
            ))
          )}
        </div>
      </Card>
    </Layout>
  )
}

function LogLine({ log }: { log: LogEntry }) {
  const levelColors: Record<string, string> = {
    ERROR: '#ef4444',
    FATAL: '#dc2626',
    WARN: '#f59e0b',
    INFO: '#60a5fa',
    DEBUG: '#94a3b8',
  }
  const color = levelColors[log.level] || '#94a3b8'
  
  // Clean message - remove control characters
  const cleanMessage = (msg: string): string => {
    return msg
      .replace(/[\x00-\x08\x0B-\x0C\x0E-\x1F\x7F]/g, '')
      .replace(/\u0001/g, '')
      .replace(/\u0002/g, '')
      .replace(/\u0000/g, '')
      .trim()
  }
  
  const timestamp = log.timestamp ? format(new Date(log.timestamp), 'HH:mm:ss.SSS') : ''
  const message = cleanMessage(log.message || '')
  
  return (
    <div style={{ fontSize: 12, fontFamily: 'monospace', marginBottom: 6, padding: '4px 0', borderLeft: `3px solid ${color}`, paddingLeft: 8 }}>
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: 8, flexWrap: 'wrap' }}>
        <span style={{ color: '#64748b', minWidth: 90 }}>{timestamp}</span>
        <span style={{ color, fontWeight: 600, minWidth: 60, textTransform: 'uppercase' }}>[{log.level}]</span>
        <span style={{ color: '#e5e7eb', flex: 1, wordBreak: 'break-word' }}>{message || '(empty message)'}</span>
      </div>
    </div>
  )
}

function formatUptime(seconds: number): string {
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  return `${h}h ${m}m`
}
