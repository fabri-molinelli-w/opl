import { useState } from 'react'
import ReactMarkdown, { type Components } from 'react-markdown'
import remarkGfm from 'remark-gfm'

// Seguridad (CU-05 RN-10): no se usa rehype-raw ni dangerouslySetInnerHTML, así que el
// HTML crudo del documento nunca se interpreta. El urlTransform por defecto de
// react-markdown neutraliza los links javascript:, data:, etc.

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null

/** Extrae el texto plano de un nodo hast (para copiar un bloque de código). */
const hastText = (node: unknown): string => {
  if (!isRecord(node)) return ''
  if (node.type === 'text' && typeof node.value === 'string') return node.value
  if (Array.isArray(node.children)) return node.children.map(hastText).join('')
  return ''
}

const CopyButton = ({ text }: { text: string }) => {
  const [copied, setCopied] = useState(false)
  const copy = async () => {
    try {
      await navigator.clipboard.writeText(text)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 1500)
    } catch {
      setCopied(false)
    }
  }
  return (
    <button
      type="button"
      onClick={() => void copy()}
      className="absolute right-2 top-2 rounded border border-slate-600 bg-slate-800 px-2 py-0.5 text-xs text-slate-200 hover:bg-slate-700"
    >
      {copied ? 'Copiado' : 'Copiar'}
    </button>
  )
}

const components: Components = {
  pre: ({ node, children }) => (
    <div className="relative">
      <CopyButton text={hastText(node).replace(/\n$/, '')} />
      <pre className="overflow-x-auto rounded-lg bg-slate-900 p-4 pr-20 text-sm text-slate-100">
        {children}
      </pre>
    </div>
  ),
  code: ({ className, children }) => (
    <code className={className ?? 'rounded bg-slate-100 px-1 py-0.5 text-sm'}>{children}</code>
  ),
  a: ({ href, children }) => (
    <a
      href={href}
      target="_blank"
      rel="noopener noreferrer"
      className="text-sky-700 underline hover:text-sky-900"
    >
      {children}
    </a>
  ),
  h1: ({ children }) => <h1 className="mt-6 text-2xl font-semibold">{children}</h1>,
  h2: ({ children }) => <h2 className="mt-5 text-xl font-semibold">{children}</h2>,
  h3: ({ children }) => <h3 className="mt-4 text-lg font-semibold">{children}</h3>,
  ul: ({ children }) => <ul className="list-disc space-y-1 pl-6">{children}</ul>,
  ol: ({ children }) => <ol className="list-decimal space-y-1 pl-6">{children}</ol>,
  table: ({ children }) => (
    <div className="overflow-x-auto">
      <table className="min-w-full border border-slate-200 text-sm">{children}</table>
    </div>
  ),
  th: ({ children }) => <th className="border border-slate-200 bg-slate-50 px-2 py-1 text-left">{children}</th>,
  td: ({ children }) => <td className="border border-slate-200 px-2 py-1">{children}</td>,
  blockquote: ({ children }) => (
    <blockquote className="border-l-4 border-slate-300 pl-4 text-slate-600">{children}</blockquote>
  ),
}

export const Markdown = ({ content }: { content: string }) => (
  <div className="space-y-3 leading-relaxed">
    <ReactMarkdown remarkPlugins={[remarkGfm]} components={components}>
      {content}
    </ReactMarkdown>
  </div>
)
