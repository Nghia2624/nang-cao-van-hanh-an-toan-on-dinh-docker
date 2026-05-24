import { useEffect, useMemo, useRef, useState } from 'react'
import { apiFetch } from '@/api/client'
import type { Container } from '@/types/api'

type ChatSession = {
  id: string
  title: string
  created_at: string
  updated_at: string
  messages: Array<{
    role: 'user' | 'assistant'
    content: string
    created_at: string
    container_id?: string
    from?: number
    to?: number
    evidence?: Array<{ timestamp: string; level: string; message: string; container_id?: string }>
  }>
}

type ChatSendResp = {
  session?: ChatSession
  answer: string
  evidence: Array<{ timestamp: string; level: string; message: string; container_id?: string }>
  insight?: any
}

function toInputDatetimeLocalValue(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0')
  const yyyy = d.getFullYear()
  const mm = pad(d.getMonth() + 1)
  const dd = pad(d.getDate())
  const hh = pad(d.getHours())
  const mi = pad(d.getMinutes())
  return `${yyyy}-${mm}-${dd}T${hh}:${mi}`
}

function parseInputDatetimeLocalToMillis(s: string): number {
  const ms = new Date(s).getTime()
  return Number.isFinite(ms) ? ms : 0
}

export default function ChatWidget() {
  const [open, setOpen] = useState(false)

  const [sessions, setSessions] = useState<ChatSession[]>([])
  const [sessionId, setSessionId] = useState<string>('')
  const [session, setSession] = useState<ChatSession | null>(null)

  const [containers, setContainers] = useState<Container[]>([])
  const [selectedContainer, setSelectedContainer] = useState<string>('')

  const [fromLocal, setFromLocal] = useState<string>(() => toInputDatetimeLocalValue(new Date(Date.now() - 60 * 60 * 1000)))
  const [toLocal, setToLocal] = useState<string>(() => toInputDatetimeLocalValue(new Date()))
  const [maxLogs, setMaxLogs] = useState<number>(120)

  const [pasteLogsOpen, setPasteLogsOpen] = useState(false)
  const [logsText, setLogsText] = useState('')

  const [message, setMessage] = useState('')
  const [sending, setSending] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const scrollerRef = useRef<HTMLDivElement | null>(null)

  // load containers once
  useEffect(() => {
    apiFetch<Container[]>('/api/v1/containers')
      .then((c) => setContainers(Array.isArray(c) ? c : []))
      .catch(() => setContainers([]))
  }, [])

  // lazy load sessions when open
  useEffect(() => {
    if (!open) return
    refreshSessions()
  }, [open])

  useEffect(() => {
    if (!open) return
    if (!sessionId) {
      setSession(null)
      return
    }
    apiFetch<ChatSession>(`/api/v1/ai/chat/sessions/${sessionId}`)
      .then((s) => setSession(s || null))
      .catch(() => setSession(null))
  }, [open, sessionId])

  useEffect(() => {
    if (!open) return
    // keep view pinned to bottom when session changes
    const t = setTimeout(() => {
      if (scrollerRef.current) {
        scrollerRef.current.scrollTop = scrollerRef.current.scrollHeight
      }
    }, 50)
    return () => clearTimeout(t)
  }, [open, session?.messages?.length])

  async function refreshSessions() {
    try {
      const resp = await apiFetch<{ sessions: ChatSession[] }>('/api/v1/ai/chat/sessions?limit=50')
      const list = Array.isArray(resp?.sessions) ? resp.sessions : []
      setSessions(list)
      if (!sessionId && list.length > 0) {
        setSessionId(list[0].id)
      }
    } catch {
      setSessions([])
    }
  }

  async function newChat(): Promise<string> {
    setError(null)
    try {
      const resp = await apiFetch<{ session?: ChatSession }>('/api/v1/ai/chat/sessions', { method: 'POST', body: JSON.stringify({ title: 'New chat' }) })
      const s = resp?.session
      await refreshSessions()
      if (s?.id) {
        setSessionId(s.id)
        return s.id
      }
    } catch (e: any) {
      setError(e?.message || 'Failed to create session')
    }
    return ''
  }

  const fromMillis = useMemo(() => parseInputDatetimeLocalToMillis(fromLocal), [fromLocal])
  const toMillis = useMemo(() => parseInputDatetimeLocalToMillis(toLocal), [toLocal])

  async function send() {
    setError(null)
    const q = message.trim()
    if (!q) {
      setError('Bạn chưa nhập câu hỏi')
      return
    }

    let sid = sessionId
    if (!sid) {
      sid = await newChat()
    }
    if (!sid) {
      setError('Không tạo được chat session')
      return
    }

    const hasPastedLogs = logsText.trim().length > 0
    const hasContainerRange = selectedContainer && fromMillis > 0 && toMillis > 0

    if (!hasPastedLogs && !hasContainerRange) {
      setError('Chọn container + khoảng thời gian hoặc paste logs')
      return
    }
    if (fromMillis && toMillis && toMillis < fromMillis) {
      setError('Thời gian To phải >= From')
      return
    }

    const payload: any = {
      message: q,
      max_logs: Math.max(10, Math.min(300, maxLogs || 120)),
    }

    if (hasPastedLogs) {
      payload.logs_text = logsText
    }
    if (hasContainerRange) {
      payload.container_id = selectedContainer
      payload.from = fromMillis
      payload.to = toMillis
    }

    setSending(true)
    try {
      const resp = await apiFetch<ChatSendResp>(`/api/v1/ai/chat/sessions/${sid}/messages`, {
        method: 'POST',
        body: JSON.stringify(payload),
      })

      // refresh session from server to get full history/titles
      if (resp?.session?.id) {
        setSession(resp.session)
        setSessionId(resp.session.id)
      } else {
        // fallback: reload
        const s = await apiFetch<ChatSession>(`/api/v1/ai/chat/sessions/${sid}`)
        setSession(s || null)
      }
      setMessage('')

      // if title updated server-side, refresh list
      refreshSessions()
    } catch (e: any) {
      setError(e?.message || 'Chat failed')
    } finally {
      setSending(false)
      setTimeout(() => {
        if (scrollerRef.current) {
          scrollerRef.current.scrollTop = scrollerRef.current.scrollHeight
        }
      }, 50)
    }
  }

  const panelStyle: React.CSSProperties = {
    position: 'fixed',
    right: 18,
    bottom: 18,
    zIndex: 9999,
  }

  const bubbleStyle: React.CSSProperties = {
    width: 56,
    height: 56,
    borderRadius: 999,
    border: '1px solid #334155',
    background: 'linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)',
    color: '#fff',
    cursor: 'pointer',
    boxShadow: '0 10px 30px rgba(0,0,0,0.35)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontWeight: 900,
    userSelect: 'none',
  }

  return (
    <div style={panelStyle}>
      {!open ? (
        <button type="button" onClick={() => setOpen(true)} style={bubbleStyle} aria-label="Open AI chat">
          AI
        </button>
      ) : (
        <div
          style={{
            width: 380,
            height: 560,
            borderRadius: 14,
            background: 'linear-gradient(135deg, #020617 0%, #0b1220 100%)',
            border: '1px solid #334155',
            boxShadow: '0 18px 50px rgba(0,0,0,0.5)',
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <div
            style={{
              padding: '10px 12px',
              borderBottom: '1px solid #1e293b',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              gap: 10,
            }}
          >
            <div style={{ display: 'flex', flexDirection: 'column', gap: 2 }}>
              <div style={{ fontWeight: 800, fontSize: 13, color: '#e5e7eb' }}>AI Chat (Logs + Containers)</div>
              <div style={{ fontSize: 11, color: '#94a3b8' }}>Hỏi đáp chỉ dựa trên logs/container bạn cung cấp</div>
            </div>
            <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
              <button
                type="button"
                onClick={newChat}
                style={{
                  padding: '6px 10px',
                  borderRadius: 10,
                  background: '#0f172a',
                  color: '#e5e7eb',
                  border: '1px solid #334155',
                  cursor: 'pointer',
                  fontSize: 12,
                  fontWeight: 800,
                }}
              >
                New
              </button>
              <button
                type="button"
                onClick={() => setOpen(false)}
                style={{
                  padding: '6px 10px',
                  borderRadius: 10,
                  background: '#0f172a',
                  color: '#e5e7eb',
                  border: '1px solid #334155',
                  cursor: 'pointer',
                  fontSize: 12,
                  fontWeight: 800,
                }}
              >
                Close
              </button>
            </div>
          </div>

          <div style={{ padding: 12, borderBottom: '1px solid #1e293b', display: 'flex', flexDirection: 'column', gap: 10 }}>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr auto', gap: 8, alignItems: 'center' }}>
              <select
                value={sessionId}
                onChange={(e) => setSessionId(e.target.value)}
                style={{
                  width: '100%',
                  padding: '8px 10px',
                  borderRadius: 10,
                  background: '#0f172a',
                  border: '1px solid #334155',
                  color: '#e5e7eb',
                  fontSize: 12,
                }}
              >
                {sessions.length === 0 ? (
                  <option value="">No sessions</option>
                ) : (
                  sessions.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.title || 'New chat'}
                    </option>
                  ))
                )}
              </select>
              <button
                type="button"
                onClick={refreshSessions}
                style={{
                  padding: '8px 10px',
                  borderRadius: 10,
                  background: '#0f172a',
                  color: '#e5e7eb',
                  border: '1px solid #334155',
                  cursor: 'pointer',
                  fontSize: 12,
                  fontWeight: 800,
                }}
              >
                ↻
              </button>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
                <div>
                  <div style={{ fontSize: 11, color: '#94a3b8', marginBottom: 6 }}>Container (optional)</div>
                  <select
                    value={selectedContainer}
                    onChange={(e) => setSelectedContainer(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px 10px',
                      borderRadius: 10,
                      background: '#0f172a',
                      border: '1px solid #334155',
                      color: '#e5e7eb',
                      fontSize: 12,
                    }}
                  >
                    <option value="">(none)</option>
                    {containers.map((c) => (
                      <option key={c.id} value={c.id}>
                        {c.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <div style={{ fontSize: 11, color: '#94a3b8', marginBottom: 6 }}>Max logs</div>
                  <input
                    type="number"
                    value={maxLogs}
                    min={10}
                    max={300}
                    onChange={(e) => setMaxLogs(Number(e.target.value))}
                    style={{
                      width: '100%',
                      padding: '8px 10px',
                      borderRadius: 10,
                      background: '#0f172a',
                      border: '1px solid #334155',
                      color: '#e5e7eb',
                      fontSize: 12,
                    }}
                  />
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
                <div>
                  <div style={{ fontSize: 11, color: '#94a3b8', marginBottom: 6 }}>From</div>
                  <input
                    type="datetime-local"
                    value={fromLocal}
                    onChange={(e) => setFromLocal(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px 10px',
                      borderRadius: 10,
                      background: '#0f172a',
                      border: '1px solid #334155',
                      color: '#e5e7eb',
                      fontSize: 12,
                    }}
                  />
                </div>
                <div>
                  <div style={{ fontSize: 11, color: '#94a3b8', marginBottom: 6 }}>To</div>
                  <input
                    type="datetime-local"
                    value={toLocal}
                    onChange={(e) => setToLocal(e.target.value)}
                    style={{
                      width: '100%',
                      padding: '8px 10px',
                      borderRadius: 10,
                      background: '#0f172a',
                      border: '1px solid #334155',
                      color: '#e5e7eb',
                      fontSize: 12,
                    }}
                  />
                </div>
              </div>

              <button
                type="button"
                onClick={() => setPasteLogsOpen((v) => !v)}
                style={{
                  padding: '8px 10px',
                  borderRadius: 10,
                  background: pasteLogsOpen ? '#1e293b' : '#0f172a',
                  color: '#e5e7eb',
                  border: '1px solid #334155',
                  cursor: 'pointer',
                  fontSize: 12,
                  fontWeight: 800,
                  textAlign: 'left',
                }}
              >
                Paste logs (optional)
              </button>

              {pasteLogsOpen && (
                <textarea
                  value={logsText}
                  onChange={(e) => setLogsText(e.target.value)}
                  placeholder="Paste logs here..."
                  rows={6}
                  style={{
                    width: '100%',
                    padding: '10px 12px',
                    borderRadius: 10,
                    background: '#020617',
                    border: '1px solid #334155',
                    color: '#e5e7eb',
                    fontSize: 12,
                    fontFamily: 'SF Mono, Monaco, Inconsolata, monospace',
                  }}
                />
              )}
            </div>
          </div>

          <div
            ref={scrollerRef}
            style={{
              flex: 1,
              padding: 12,
              overflow: 'auto',
              background: '#020617',
            }}
          >
            {session?.messages?.length ? (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                {session.messages.map((m, idx) => (
                  <ChatBubble key={idx} role={m.role} content={m.content} evidence={m.evidence} />
                ))}
              </div>
            ) : (
              <div style={{ color: '#94a3b8', fontSize: 12, lineHeight: 1.6 }}>
                - Chọn container + thời gian rồi đặt câu hỏi
                <br />- Hoặc paste logs vào ô trên
              </div>
            )}
          </div>

          <div style={{ padding: 12, borderTop: '1px solid #1e293b', background: '#0b1220' }}>
            {error && <div style={{ color: '#ef4444', fontSize: 12, marginBottom: 8 }}>{error}</div>}
            <div style={{ display: 'flex', gap: 8 }}>
              <input
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault()
                    if (!sending) send()
                  }
                }}
                placeholder="Hỏi về logs/container..."
                disabled={sending}
                style={{
                  flex: 1,
                  padding: '10px 12px',
                  borderRadius: 12,
                  background: '#0f172a',
                  border: '1px solid #334155',
                  color: '#e5e7eb',
                  fontSize: 13,
                }}
              />
              <button
                type="button"
                onClick={send}
                disabled={sending}
                style={{
                  padding: '10px 12px',
                  borderRadius: 12,
                  background: sending ? '#334155' : '#3b82f6',
                  color: '#fff',
                  border: 'none',
                  cursor: sending ? 'not-allowed' : 'pointer',
                  fontSize: 13,
                  fontWeight: 900,
                  width: 72,
                }}
              >
                {sending ? '...' : 'Send'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

function ChatBubble({
  role,
  content,
  evidence,
}: {
  role: 'user' | 'assistant'
  content: string
  evidence?: Array<{ timestamp: string; level: string; message: string; container_id?: string }>
}) {
  const isUser = role === 'user'
  return (
    <div style={{ display: 'flex', justifyContent: isUser ? 'flex-end' : 'flex-start' }}>
      <div
        style={{
          maxWidth: '85%',
          padding: '10px 12px',
          borderRadius: 14,
          background: isUser ? 'linear-gradient(135deg, #2563eb 0%, #3b82f6 100%)' : '#0f172a',
          border: isUser ? '1px solid #1d4ed8' : '1px solid #334155',
          color: '#e5e7eb',
        }}
      >
        <div style={{ fontSize: 12, whiteSpace: 'pre-wrap', lineHeight: 1.55 }}>{content}</div>
        {!isUser && evidence && evidence.length > 0 && (
          <div style={{ marginTop: 10, borderTop: '1px solid #1e293b', paddingTop: 8 }}>
            <div style={{ color: '#94a3b8', fontSize: 11, fontWeight: 800, marginBottom: 6 }}>Evidence</div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              {evidence.slice(0, 3).map((e, idx) => (
                <EvidenceLine key={idx} e={e} />
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

function EvidenceLine({ e }: { e: { timestamp: string; level: string; message: string; container_id?: string } }) {
  const levelColors: Record<string, string> = {
    ERROR: '#ef4444',
    FATAL: '#dc2626',
    WARN: '#f59e0b',
    INFO: '#60a5fa',
    DEBUG: '#94a3b8',
  }
  const c = levelColors[(e.level || '').toUpperCase()] || '#94a3b8'

  return (
    <div
      style={{
        fontFamily: 'SF Mono, Monaco, Inconsolata, monospace',
        fontSize: 11,
        background: '#020617',
        border: '1px solid #334155',
        borderRadius: 10,
        padding: 8,
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', gap: 10, marginBottom: 6 }}>
        <span style={{ color: '#64748b' }}>{String(e.timestamp || '').slice(0, 19).replace('T', ' ')}</span>
        <span style={{ color: c, fontWeight: 800 }}>{String(e.level || '').toUpperCase()}</span>
      </div>
      <div style={{ color: '#e5e7eb', whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>{e.message}</div>
    </div>
  )
}

