import React, { useEffect, useState, useMemo } from 'react'
import { format, subDays } from 'date-fns'
import Layout from '@/components/Layout'
import Card from '@/components/Card'
import { apiFetch, sseUrl } from '@/api/client'
import { useSSE } from '@/hooks/useSSE'
import type { Container, LogEntry } from '@/types/api'

export default function LogsViewer() {
  const [containers, setContainers] = useState<Container[]>([])
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [selectedContainer, setSelectedContainer] = useState<string>('')
  const [levelFilter, setLevelFilter] = useState<string>('')
  const [searchQuery, setSearchQuery] = useState('')
  const [autoScroll, setAutoScroll] = useState(true)
  const [rangePreset, setRangePreset] = useState<'1d' | '3d' | '7d' | 'custom'>('1d')
  const [customFrom, setCustomFrom] = useState('')
  const [customTo, setCustomTo] = useState('')

  useEffect(() => {
    apiFetch<Container[]>('/api/v1/containers').then(setContainers).catch(console.error)
  }, [])

  useEffect(() => {
    loadLogs()
  }, [selectedContainer, levelFilter, rangePreset, customFrom, customTo])

  function loadLogs() {
    const now = new Date()
    let from: Date
    let to = now
    if (rangePreset === 'custom') {
      if (!customFrom || !customTo) return
      from = new Date(customFrom)
      to = new Date(customTo)
    } else {
      const days = rangePreset === '1d' ? 1 : rangePreset === '3d' ? 3 : 7
      from = subDays(now, days)
    }
    const params = new URLSearchParams({
      from: String(from.getTime()),
      to: String(to.getTime()),
      limit: '500',
    })
    if (selectedContainer) params.set('container', selectedContainer)
    if (levelFilter) params.set('level', levelFilter)
    apiFetch<LogEntry[]>(`/api/v1/logs?${params}`)
      .then((items) => setLogs(items || []))
      .catch(console.error)
  }

  useSSE(
    selectedContainer ? sseUrl('/api/v1/logs/stream', { container: selectedContainer }) : null,
    {
      enabled: !!selectedContainer,
      onMessage: (data) => {
        try {
          const entry: LogEntry = JSON.parse(data)
          setLogs((prev) => {
            const filtered = levelFilter ? entry.level === levelFilter : true
            return filtered ? [entry, ...prev].slice(0, 1000) : prev
          })
        } catch {
          // ignore
        }
      },
    }
  )

  const filteredLogs = useMemo(() => {
    let result = logs
    if (searchQuery) {
      const q = searchQuery.toLowerCase()
      result = result.filter((log) => log.message.toLowerCase().includes(q) || log.container?.toLowerCase().includes(q))
    }
    // Sort descending by timestamp so newest is always at the top
    return result.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
  }, [logs, searchQuery])

  const logContainerRef = (node: HTMLDivElement | null) => {
    if (node && autoScroll) {
      node.scrollTop = 0
    }
  }

  return (
    <Layout title="Logs Viewer">
      <div style={{ display: 'grid', gridTemplateColumns: '250px 1fr', gap: 16 }}>
        <Card title="Filters">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <div>
              <label style={{ fontSize: 12, color: '#94a3b8', display: 'block', marginBottom: 6 }}>Container</label>
              <select
                value={selectedContainer}
                onChange={(e) => setSelectedContainer(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px',
                  borderRadius: 6,
                  background: '#0f172a',
                  border: '1px solid #334155',
                  color: '#e5e7eb',
                  fontSize: 14,
                }}
              >
                <option value="">All containers</option>
                {containers.map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label style={{ fontSize: 12, color: '#94a3b8', display: 'block', marginBottom: 6 }}>Time Range</label>
              <select
                value={rangePreset}
                onChange={(e) => setRangePreset(e.target.value as any)}
                style={{
                  width: '100%',
                  padding: '8px',
                  borderRadius: 6,
                  background: '#0f172a',
                  border: '1px solid #334155',
                  color: '#e5e7eb',
                  fontSize: 14,
                }}
              >
                <option value="1d">Last 1 day</option>
                <option value="3d">Last 3 days</option>
                <option value="7d">Last 7 days</option>
                <option value="custom">Custom</option>
              </select>
              {rangePreset === 'custom' && (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginTop: 10 }}>
                  <input
                    type="datetime-local"
                    value={customFrom}
                    onChange={(e) => setCustomFrom(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px',
                      borderRadius: 6,
                      background: '#0f172a',
                      border: '1px solid #334155',
                      color: '#e5e7eb',
                      fontSize: 14,
                    }}
                  />
                  <input
                    type="datetime-local"
                    value={customTo}
                    onChange={(e) => setCustomTo(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px',
                      borderRadius: 6,
                      background: '#0f172a',
                      border: '1px solid #334155',
                      color: '#e5e7eb',
                      fontSize: 14,
                    }}
                  />
                </div>
              )}
            </div>

            <div>
              <label style={{ fontSize: 12, color: '#94a3b8', display: 'block', marginBottom: 6 }}>Level</label>
              <select
                value={levelFilter}
                onChange={(e) => setLevelFilter(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px',
                  borderRadius: 6,
                  background: '#0f172a',
                  border: '1px solid #334155',
                  color: '#e5e7eb',
                  fontSize: 14,
                }}
              >
                <option value="">All levels</option>
                <option value="ERROR">ERROR</option>
                <option value="WARN">WARN</option>
                <option value="INFO">INFO</option>
                <option value="DEBUG">DEBUG</option>
                <option value="FATAL">FATAL</option>
              </select>
            </div>

            <div>
              <label style={{ fontSize: 12, color: '#94a3b8', display: 'block', marginBottom: 6 }}>Search</label>
              <input
                type="text"
                placeholder="Search logs..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px',
                  borderRadius: 6,
                  background: '#0f172a',
                  border: '1px solid #334155',
                  color: '#e5e7eb',
                  fontSize: 14,
                }}
              />
            </div>

            <div>
              <label style={{ fontSize: 12, color: '#94a3b8', display: 'flex', alignItems: 'center', gap: 8 }}>
                <input
                  type="checkbox"
                  checked={autoScroll}
                  onChange={(e) => setAutoScroll(e.target.checked)}
                  style={{ cursor: 'pointer' }}
                />
                Auto-scroll
              </label>
            </div>

            <button
              onClick={loadLogs}
              style={{
                padding: '8px',
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
        </Card>

        <Card title={
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span>📋 Logs ({filteredLogs.length})</span>
            {selectedContainer ? (
              <span style={{ fontSize: 12, padding: '4px 8px', borderRadius: 12, background: '#10b98122', color: '#10b981', display: 'flex', alignItems: 'center', gap: 6 }}>
                <span style={{ width: 8, height: 8, borderRadius: '50%', background: '#10b981', animation: 'pulse 2s infinite' }} />
                Live
              </span>
            ) : (
              <span style={{ fontSize: 12, padding: '4px 8px', borderRadius: 12, background: '#64748b22', color: '#94a3b8', display: 'flex', alignItems: 'center', gap: 6 }}>
                <span style={{ width: 8, height: 8, borderRadius: '50%', background: '#64748b' }} />
                Streaming paused (select container)
              </span>
            )}
          </div>
        }>
          <div
            ref={logContainerRef}
            style={{
              maxHeight: 650,
              overflow: 'auto',
              background: '#020617',
              borderRadius: 10,
              padding: 16,
              fontFamily: 'SF Mono, Monaco, Inconsolata, monospace',
              fontSize: 13,
              border: '1px solid #334155',
            }}
          >
            {filteredLogs.length === 0 ? (
              <div style={{ color: '#94a3b8', textAlign: 'center', padding: 24 }}>No logs</div>
            ) : (
              filteredLogs.map((log, i) => <LogLine key={i} log={log} searchQuery={searchQuery} />)
            )}
          </div>
        </Card>
      </div>
    </Layout>
  )
}

function LogLine({ log, searchQuery }: { log: LogEntry; searchQuery: string }) {
  const levelColors: Record<string, string> = {
    ERROR: '#ef4444',
    FATAL: '#dc2626',
    WARN: '#f59e0b',
    INFO: '#60a5fa',
    DEBUG: '#94a3b8',
  }
  const color = levelColors[log.level] || '#94a3b8'
  const timestamp = log.timestamp ? format(new Date(log.timestamp), 'HH:mm:ss.SSS') : ''

  // Clean message - remove control characters and special bytes
  const cleanMessage = (msg: string): string => {
    return msg
      .replace(/[\x00-\x08\x0B-\x0C\x0E-\x1F\x7F]/g, '') // Remove control chars except \n, \t
      .replace(/\u0001/g, '') // Remove SOH
      .replace(/\u0002/g, '') // Remove STX
      .replace(/\u0000/g, '') // Remove null bytes
      .trim()
  }

  const msgStr = cleanMessage(log.message || '')
  let message: React.ReactNode = msgStr
  if (searchQuery && msgStr) {
    const parts = msgStr.split(new RegExp(`(${searchQuery.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi'))
    message = parts.map((part: string, i: number) =>
      part.toLowerCase() === searchQuery.toLowerCase() ? (
        <mark key={i} style={{ background: '#f59e0b', color: '#000', padding: '2px 4px', borderRadius: 3 }}>
          {part}
        </mark>
      ) : (
        part
      )
    )
  }

  return (
    <div style={{ marginBottom: 6, lineHeight: 1.6, padding: '4px 0', borderLeft: `3px solid ${color}`, paddingLeft: 8 }}>
      <div style={{ display: 'flex', alignItems: 'flex-start', gap: 8, flexWrap: 'wrap' }}>
        <span style={{ color: '#64748b', fontSize: 11, fontFamily: 'monospace', minWidth: 90 }}>{timestamp}</span>
        <span style={{ color, fontWeight: 600, fontSize: 11, minWidth: 60, textTransform: 'uppercase' }}>[{log.level}]</span>
        {log.container && (
          <span style={{ color: '#94a3b8', fontSize: 11, fontFamily: 'monospace' }}>{log.container.substring(0, 12)}</span>
        )}
        <span style={{ color: '#e5e7eb', fontSize: 12, flex: 1, wordBreak: 'break-word' }}>{message || '(empty message)'}</span>
      </div>
    </div>
  )
}
