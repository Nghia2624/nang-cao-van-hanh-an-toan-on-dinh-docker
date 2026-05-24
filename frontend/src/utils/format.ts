export function formatUptime(seconds: number): string {
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

export function formatETA(minutes: number): string {
  if (minutes < 0) return 'N/A'
  if (minutes < 60) return `~${minutes}m`
  if (minutes < 1440) return `~${(minutes / 60).toFixed(1)}h`
  return `~${(minutes / 1440).toFixed(1)}d`
}

/**
 * Clean log message by removing Docker header bytes and control characters.
 */
export function cleanLogMessage(msg: string): string {
  return msg
    .replace(/[\x00-\x08\x0B-\x0C\x0E-\x1F\x7F]/g, '')
    .replace(/\u0001/g, '')
    .replace(/\u0002/g, '')
    .replace(/\u0000/g, '')
    .trim()
}

/**
 * Parse a metric timestamp that could be a number (unix) or ISO string.
 */
export function parseMetricTimestamp(ts: number | string): number {
  if (typeof ts === 'number') {
    // If > 1e12, treat as milliseconds; otherwise seconds → milliseconds
    return ts > 1e12 ? ts : ts * 1000
  }
  return new Date(ts).getTime()
}
