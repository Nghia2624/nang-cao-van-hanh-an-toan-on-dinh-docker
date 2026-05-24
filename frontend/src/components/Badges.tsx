

export function StatusBadge({ status }: { status: string }) {
  const s = status.toLowerCase()
  const color = s.includes('running') || s.includes('up')
    ? '#16a34a'
    : s.includes('restart')
      ? '#f59e0b'
      : s.includes('unhealthy') || s.includes('dead')
        ? '#ef4444'
        : s.includes('paused')
          ? '#8b5cf6'
          : '#64748b'
  return (
    <span
      style={{
        fontSize: 11,
        padding: '4px 10px',
        borderRadius: 999,
        background: `${color}22`,
        color,
        flexShrink: 0,
        fontWeight: 500,
      }}
    >
      {status}
    </span>
  )
}

export function SeverityBadge({ severity }: { severity: string }) {
  const colors: Record<string, string> = {
    CRITICAL: '#ef4444',
    HIGH: '#f59e0b',
    MEDIUM: '#3b82f6',
    LOW: '#94a3b8',
    INFO: '#60a5fa',
  }
  const color = colors[severity] || '#64748b'
  return (
    <span
      style={{
        fontSize: 11,
        padding: '4px 10px',
        borderRadius: 999,
        background: `${color}22`,
        color,
        fontWeight: 500,
      }}
    >
      {severity}
    </span>
  )
}

export function RiskBadge({ level, score }: { level: 'low' | 'medium' | 'high'; score: number }) {
  const colors = { low: '#10b981', medium: '#f59e0b', high: '#ef4444' }
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
