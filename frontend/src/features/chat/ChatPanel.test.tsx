import { afterEach, describe, expect, it, vi } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ChatPanel } from './ChatPanel'

const jsonResponse = (status: number, body: unknown): Response =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })

const renderPanel = (response: Response) => {
  const fetchMock = vi.fn().mockResolvedValue(response)
  vi.stubGlobal('fetch', fetchMock)
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } })
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <ChatPanel onClose={() => undefined} />
      </MemoryRouter>
    </QueryClientProvider>,
  )
  return fetchMock
}

describe('ChatPanel', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('CA-1: muestra la respuesta y sus fuentes con links', async () => {
    const fetchMock = renderPanel(
      jsonResponse(200, {
        answer: 'Según **Reiniciar servicio Nginx**, ejecutá `sudo nginx -t`.',
        sources: [{ type: 'document', id: 'doc-1', title: 'Reiniciar servicio Nginx' }],
      }),
    )
    const user = userEvent.setup()

    await user.type(screen.getByLabelText('Mensaje para OPSBot'), '¿Cómo reinicio nginx?')
    await user.click(screen.getByRole('button', { name: 'Enviar' }))

    expect(await screen.findByRole('link', { name: 'Documento: Reiniciar servicio Nginx' })).toHaveAttribute(
      'href',
      '/documentacion/doc-1',
    )
    expect(screen.getByText('¿Cómo reinicio nginx?')).toBeInTheDocument()
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit]
    expect(url).toContain('/api/chat')
    const body = JSON.parse(String(init.body)) as { sessionId: string; message: string }
    expect(body.message).toBe('¿Cómo reinicio nginx?')
    expect(body.sessionId.length).toBeGreaterThan(0)
  })

  it('muestra un error claro si OPSBot no está disponible (503)', async () => {
    renderPanel(jsonResponse(503, { code: 'unavailable', message: 'OPSBot no está configurado en este servidor.' }))
    const user = userEvent.setup()

    await user.click(screen.getByRole('button', { name: '¿Qué servicios están caídos en producción?' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('OPSBot no está configurado en este servidor.')
  })
})
