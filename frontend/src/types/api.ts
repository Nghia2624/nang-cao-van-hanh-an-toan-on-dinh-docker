export interface Container {
  id: string
  name: string
  image: string
  status: string
  created_at: string
  uptime_sec: number
  restart_count: number
  labels?: Record<string, string>
}

export interface SystemMetrics {
  cpu_percent: number
  memory_percent: number
  container_running: number
  container_total: number
}

export interface MetricPoint {
  timestamp: string | number
  value: number
}

export interface LogEntry {
  container_id: string
  container?: string
  container_name?: string
  image?: string
  level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR' | 'FATAL'
  message: string
  timestamp: string
  stream?: 'stdout' | 'stderr'
  fingerprint?: string
  labels?: Record<string, string>
}

export interface Alert {
  id: string
  source: string
  severity: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL'
  title: string
  description: string
  status: 'NEW' | 'ACKNOWLEDGED' | 'RESOLVED'
  container_id?: string
  fingerprint?: string
  created_at: string
  updated_at: string
}

export interface AIAnalysis {
  fingerprint: string
  root_cause: string
  severity: number
  severity_label: 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL'
  impact_analysis: string
  affected_components: string[]
  recommended_actions: Array<{
    priority?: number | string
    action: string
    estimated_effort?: string
    impact?: string
    steps?: string[]
  }>
  prevention_measures: string[]
  related_issues: string[]
  confidence_score: number
  created_at: string
}

export interface ContainerPrediction {
  id?: string
  container_id: string
  name: string
  status: string
  predicted_at: string
  risk_level: 'low' | 'medium' | 'high'
  risk_score: number
  risk_factors: string[]
  recommendations: string[]
  confidence_score: number
  eta_minutes: number
  signals: {
    cpu_avg: number
    cpu_max: number
    mem_avg: number
    mem_max: number
    error_count_1h: number
    restart_count: number
    uptime_sec: number
  }
  summary: string
}

export interface APIError {
  error: string
  message?: string
}
