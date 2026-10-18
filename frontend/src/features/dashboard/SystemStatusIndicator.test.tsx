import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { SystemStatusIndicator } from './SystemStatusIndicator'

const jsonResponse = (status: number, body: unknown): Response =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })

const renderIndicator = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <SystemStatusIndicator />
    </QueryClientProvider>,
  )
}

describe('SystemStatusIndicator', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('muestra verde cuando el sistema está operativo', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(jsonResponse(200, { status: 'ok', database: 'up' })),
    )

    renderIndicator()

    expect(await screen.findByText('Operativo')).toBeInTheDocument()
  })

  it('muestra rojo cuando el backend responde 503 (degradado)', async () => {
    vi.stubGlobal(
      'fetch',
      vi
        .fn()
        .mockResolvedValue(
          jsonResponse(503, { status: 'degraded', database: 'down' }),
        ),
    )

    renderIndicator()

    expect(await screen.findByText('Degradado')).toBeInTheDocument()
  })

  it('muestra gris cuando no hay conexión', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockRejectedValue(new TypeError('Failed to fetch')),
    )

    renderIndicator()

    expect(await screen.findByText('Sin conexión')).toBeInTheDocument()
  })
})
