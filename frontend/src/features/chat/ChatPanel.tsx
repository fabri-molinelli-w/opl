import { useEffect, useRef, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { useChatWithOpsbot } from '@/api/gen'
import type { ChatSource } from '@/api/gen/model'
import { isApiError } from '@/api/http'
import { Markdown } from '@/components/Markdown'

type ChatMessage =
  | { role: 'user'; content: string }
  | { role: 'assistant'; content: string; sources: ChatSource[] }
  | { role: 'error'; content: string }

const SUGGESTIONS = [
  '¿Qué servicios están caídos en producción?',
  '¿Cómo reinicio Nginx?',
  '¿Cómo restauro un backup de PostgreSQL?',
]

const newSessionId = (): string =>
  typeof crypto !== 'undefined' && 'randomUUID' in crypto
    ? crypto.randomUUID()
    : `s-${Date.now()}-${Math.random().toString(16).slice(2)}`

const sourceLink = (source: ChatSource): string | null => {
  switch (source.type) {
    case 'document':
      return `/documentacion/${source.id}`
    case 'server':
      return `/inventario/${source.id}`
    default:
      return null
  }
}

const SOURCE_LABEL: Record<ChatSource['type'], string> = {
  document: 'Documento',
  server: 'Servidor',
  service: 'Servicio',
}

export const ChatPanel = ({ onClose }: { onClose: () => void }) => {
  const [sessionId, setSessionId] = useState(newSessionId)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState('')
  const chat = useChatWithOpsbot()
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView?.({ behavior: 'smooth' })
  }, [messages, chat.isPending])

  const send = async (text: string) => {
    const message = text.trim()
    if (message === '' || chat.isPending) return
    setInput('')
    setMessages((prev) => [...prev, { role: 'user', content: message }])
    try {
      const result = await chat.mutateAsync({ data: { sessionId, message } })
      if (result.status === 200) {
        const { answer, sources } = result.data
        setMessages((prev) => [...prev, { role: 'assistant', content: answer, sources }])
      } else {
        // 503: el mutator no lanza para este código; el cuerpo es un Error.
        const data: unknown = result.data
        const text =
          typeof data === 'object' && data !== null && 'message' in data && typeof data.message === 'string'
            ? data.message
            : 'OPSBot no está disponible en este momento.'
        setMessages((prev) => [...prev, { role: 'error', content: text }])
      }
    } catch (cause) {
      setMessages((prev) => [
        ...prev,
        { role: 'error', content: isApiError(cause) ? cause.message : 'No se pudo contactar a OPSBot.' },
      ])
    }
  }

  const submit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    void send(input)
  }

  return (
    <aside
      aria-label="OPSBot"
      className="fixed inset-y-0 right-0 z-40 flex w-full max-w-md flex-col border-l border-slate-200 bg-white shadow-xl"
    >
      <header className="flex items-center justify-between border-b border-slate-200 px-4 py-3">
        <div>
          <h2 className="font-semibold">OPSBot</h2>
          <p className="text-xs text-slate-500">Responde con la documentación y el inventario. Solo lectura.</p>
        </div>
        <div className="flex gap-2">
          <button
            type="button"
            onClick={() => {
              setMessages([])
              setSessionId(newSessionId())
            }}
            className="rounded-md border border-slate-300 px-2 py-1 text-xs hover:bg-slate-100"
          >
            Nueva conversación
          </button>
          <button
            type="button"
            onClick={onClose}
            aria-label="Cerrar OPSBot"
            className="rounded-md px-2 py-1 text-slate-500 hover:bg-slate-100"
          >
            ✕
          </button>
        </div>
      </header>

      <div className="flex-1 space-y-4 overflow-y-auto px-4 py-4 text-sm">
        {messages.length === 0 && (
          <div className="space-y-2">
            <p className="text-slate-600">Probá con alguna de estas preguntas:</p>
            {SUGGESTIONS.map((s) => (
              <button
                key={s}
                type="button"
                onClick={() => void send(s)}
                className="block w-full rounded-md border border-slate-200 px-3 py-2 text-left hover:bg-slate-50"
              >
                {s}
              </button>
            ))}
          </div>
        )}

        {messages.map((m, i) => {
          if (m.role === 'user') {
            return (
              <div key={i} className="ml-8 rounded-lg bg-slate-900 px-3 py-2 text-white">
                {m.content}
              </div>
            )
          }
          if (m.role === 'error') {
            return (
              <div key={i} role="alert" className="rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-red-800">
                {m.content}
              </div>
            )
          }
          return (
            <div key={i} className="mr-4 space-y-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">
              <Markdown content={m.content} />
              {m.sources.length > 0 && (
                <div className="border-t border-slate-200 pt-2 text-xs">
                  <p className="mb-1 font-medium text-slate-500">Fuentes</p>
                  <ul className="flex flex-wrap gap-1.5">
                    {m.sources.map((s) => {
                      const to = sourceLink(s)
                      const label = `${SOURCE_LABEL[s.type]}: ${s.title}`
                      return (
                        <li key={`${s.type}-${s.id}`}>
                          {to === null ? (
                            <span className="rounded bg-white px-1.5 py-0.5">{label}</span>
                          ) : (
                            <Link to={to} className="rounded bg-white px-1.5 py-0.5 text-sky-800 hover:underline">
                              {label}
                            </Link>
                          )}
                        </li>
                      )
                    })}
                  </ul>
                </div>
              )}
            </div>
          )
        })}
        {chat.isPending && <p className="text-slate-500">OPSBot está pensando…</p>}
        <div ref={bottomRef} />
      </div>

      <form onSubmit={submit} className="flex gap-2 border-t border-slate-200 p-3">
        <input
          aria-label="Mensaje para OPSBot"
          value={input}
          onChange={(event) => setInput(event.target.value)}
          maxLength={4000}
          placeholder="Escribí tu consulta…"
          className="flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm"
        />
        <button
          type="submit"
          disabled={chat.isPending || input.trim() === ''}
          className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white disabled:opacity-40"
        >
          Enviar
        </button>
      </form>
    </aside>
  )
}
