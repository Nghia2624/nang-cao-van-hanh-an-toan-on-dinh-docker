import { useEffect, useState } from 'react'
import { format } from 'date-fns'
import Layout from '@/components/Layout'
import Card from '@/components/Card'
import { apiFetch } from '@/api/client'
import type { AIAnalysis } from '@/types/api'

export default function AIInsights() {
  const [analyses, setAnalyses] = useState<AIAnalysis[]>([])
  const [selected, setSelected] = useState<AIAnalysis | null>(null)

  useEffect(() => {
    loadAnalyses()
  }, [])

  function loadAnalyses() {
    apiFetch<AIAnalysis[]>('/api/v1/ai/analyses')
      .then(setAnalyses)
      .catch(console.error)
  }

  const selectedAnalysis = selected || (analyses && analyses.length > 0 ? analyses[0] : null)

  return (
    <Layout title="AI Insights">
      <div style={{ display: 'grid', gridTemplateColumns: '380px 1fr', gap: 20 }}>
        <Card title="📊 AI Analyses">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 10, maxHeight: 650, overflow: 'auto' }}>
            {analyses.length === 0 ? (
              <div style={{ color: '#94a3b8', textAlign: 'center', padding: 40, fontSize: 14 }}>
                <div style={{ fontSize: 48, marginBottom: 16 }}>🤖</div>
                <div>No AI analyses yet</div>
                <div style={{ fontSize: 12, color: '#64748b', marginTop: 8 }}>Analyses will appear when error logs are detected</div>
              </div>
            ) : (
              analyses.map((a) => (
                <div
                  key={a.fingerprint}
                  onClick={() => setSelected(a)}
                  style={{
                    padding: 16,
                    borderRadius: 12,
                    background: selected?.fingerprint === a.fingerprint ? 'linear-gradient(135deg, #1e293b 0%, #0f172a 100%)' : 'transparent',
                    border: `2px solid ${selected?.fingerprint === a.fingerprint ? '#3b82f6' : 'transparent'}`,
                    cursor: 'pointer',
                    transition: 'all 0.2s ease',
                  }}
                  onMouseEnter={(e) => {
                    if (selected?.fingerprint !== a.fingerprint) {
                      e.currentTarget.style.background = '#1e293b'
                      e.currentTarget.style.borderColor = '#334155'
                    }
                  }}
                  onMouseLeave={(e) => {
                    if (selected?.fingerprint !== a.fingerprint) {
                      e.currentTarget.style.background = 'transparent'
                      e.currentTarget.style.borderColor = 'transparent'
                    }
                  }}
                >
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', marginBottom: 8 }}>
                    <SeverityBadge severity={a.severity_label} />
                    <div style={{ fontSize: 11, color: '#64748b' }}>
                      {format(new Date(a.created_at), 'MMM d, HH:mm')}
                    </div>
                  </div>
                  <div style={{ fontSize: 13, color: '#e5e7eb', marginBottom: 4 }}>
                    {a.root_cause ? (a.root_cause.length > 80 ? a.root_cause.substring(0, 80) + '...' : a.root_cause) : 'No root cause'}
                  </div>
                  <div style={{ fontSize: 11, color: '#94a3b8' }}>
                    Confidence: {a.confidence_score ? (a.confidence_score * 100).toFixed(0) : 0}%
                  </div>
                </div>
              ))
            )}
          </div>
        </Card>

        {selectedAnalysis && (
          <Card title="Analysis Details">
            <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
              <div>
                <div style={{ fontSize: 12, color: '#94a3b8', marginBottom: 8 }}>Severity</div>
                <SeverityBadge severity={selectedAnalysis.severity_label || 'UNKNOWN'} />
              </div>

              <div>
                <div style={{ fontSize: 12, color: '#94a3b8', marginBottom: 12, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                  Root Cause
                </div>
                <div style={{ color: '#e5e7eb', lineHeight: 1.7, fontSize: 14, padding: 16, borderRadius: 10, background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)', border: '1px solid #334155' }}>
                  {selectedAnalysis.root_cause || 'No root cause analysis available'}
                </div>
              </div>

              <div>
                <div style={{ fontSize: 12, color: '#94a3b8', marginBottom: 12, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                  Impact Analysis
                </div>
                <div style={{ color: '#e5e7eb', lineHeight: 1.7, fontSize: 14, padding: 16, borderRadius: 10, background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)', border: '1px solid #334155' }}>
                  {selectedAnalysis.impact_analysis || 'No impact analysis available'}
                </div>
              </div>

              {selectedAnalysis.affected_components && selectedAnalysis.affected_components.length > 0 && (
                <div>
                  <div style={{ fontSize: 12, color: '#94a3b8', marginBottom: 12, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                    Affected Components
                  </div>
                  <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
                    {selectedAnalysis.affected_components.map((comp, i) => (
                      <span
                        key={i}
                        style={{
                          padding: '6px 14px',
                          borderRadius: 8,
                          background: 'linear-gradient(135deg, #1e293b 0%, #0f172a 100%)',
                          border: '1px solid #334155',
                          fontSize: 13,
                          color: '#cbd5e1',
                          fontWeight: 500,
                          transition: 'all 0.2s ease',
                        }}
                        onMouseEnter={(e) => {
                          e.currentTarget.style.borderColor = '#475569'
                          e.currentTarget.style.transform = 'translateY(-2px)'
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.borderColor = '#334155'
                          e.currentTarget.style.transform = 'translateY(0)'
                        }}
                      >
                        {comp}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {selectedAnalysis.recommended_actions && selectedAnalysis.recommended_actions.length > 0 && (
                <div>
                  <div style={{ fontSize: 12, color: '#94a3b8', marginBottom: 12, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                    Recommended Actions
                  </div>
                  <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                    {selectedAnalysis.recommended_actions
                      .sort((a, b) => {
                        const aPriority = typeof a.priority === 'number' ? a.priority : typeof a.priority === 'string' ? parseInt(a.priority) || 0 : 0
                        const bPriority = typeof b.priority === 'number' ? b.priority : typeof b.priority === 'string' ? parseInt(b.priority) || 0 : 0
                        return aPriority - bPriority
                      })
                      .map((action, i) => (
                        <div
                          key={i}
                          style={{
                            padding: 16,
                            borderRadius: 12,
                            background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)',
                            border: '1px solid #334155',
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
                          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', marginBottom: 8 }}>
                            <div style={{ fontWeight: 600, color: '#e5e7eb', fontSize: 15, lineHeight: 1.4 }}>{action.action}</div>
                            <span
                              style={{
                                fontSize: 11,
                                padding: '4px 10px',
                                borderRadius: 999,
                                background: '#3b82f622',
                                color: '#60a5fa',
                                fontWeight: 600,
                                whiteSpace: 'nowrap',
                                marginLeft: 12,
                              }}
                            >
                              Priority {action.priority}
                            </span>
                          </div>
                          <div style={{ fontSize: 13, color: '#94a3b8', marginTop: 8, display: 'flex', gap: 12, alignItems: 'center' }}>
                            {action.impact && (
                              <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                                <span style={{ fontSize: 16 }}>📊</span>
                                {action.impact}
                              </span>
                            )}
                            {action.estimated_effort && (
                              <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                                <span style={{ fontSize: 16 }}>⏱️</span>
                                {action.estimated_effort}
                              </span>
                            )}
                          </div>
                          {action.steps && action.steps.length > 0 && (
                            <div style={{ marginTop: 12, paddingTop: 12, borderTop: '1px solid #334155' }}>
                              <div style={{ fontSize: 12, color: '#64748b', marginBottom: 8, fontWeight: 600 }}>Steps:</div>
                              <ol style={{ margin: 0, paddingLeft: 20, color: '#cbd5e1', lineHeight: 1.8 }}>
                                {action.steps.map((step, stepIdx) => (
                                  <li key={stepIdx} style={{ marginBottom: 6, fontSize: 13 }}>
                                    {step}
                                  </li>
                                ))}
                              </ol>
                            </div>
                          )}
                        </div>
                      ))}
                  </div>
                </div>
              )}

              {selectedAnalysis.prevention_measures && selectedAnalysis.prevention_measures.length > 0 && (
                <div>
                  <div style={{ fontSize: 12, color: '#94a3b8', marginBottom: 12, fontWeight: 600, textTransform: 'uppercase', letterSpacing: '0.5px' }}>
                    Prevention Measures
                  </div>
                  <div style={{ padding: 16, borderRadius: 10, background: 'linear-gradient(135deg, #0f172a 0%, #1e293b 100%)', border: '1px solid #334155' }}>
                    <ul style={{ margin: 0, paddingLeft: 20, color: '#e5e7eb', lineHeight: 1.9 }}>
                    {selectedAnalysis.prevention_measures.map((measure, i) => (
                        <li key={i} style={{ marginBottom: 8, fontSize: 14 }}>{measure}</li>
                    ))}
                  </ul>
                  </div>
                </div>
              )}

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, paddingTop: 16, borderTop: '1px solid #334155' }}>
                <div>
                  <div style={{ fontSize: 12, color: '#94a3b8' }}>Confidence Score</div>
                  <div style={{ fontSize: 20, fontWeight: 600, color: '#60a5fa', marginTop: 4 }}>
                    {selectedAnalysis.confidence_score ? (selectedAnalysis.confidence_score * 100).toFixed(1) : 0}%
                  </div>
                </div>
                <div>
                  <div style={{ fontSize: 12, color: '#94a3b8' }}>Created</div>
                  <div style={{ color: '#e5e7eb', marginTop: 4 }}>
                    {format(new Date(selectedAnalysis.created_at), 'PPpp')}
                  </div>
                </div>
              </div>
            </div>
          </Card>
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
