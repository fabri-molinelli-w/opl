import { afterEach, describe, expect, it, vi } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { Server } from '@/api/gen/model'
import { ServerForm } from './ServerForm'

const jsonResponse = (status: number, body: unknown): Response =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })

const server: Server = {
  id: '0192f000-0000-7000-8000-000000000001',
  hostname: 'web-01.example.com',
  ip: '192.0.2.10',
  environment: 'prod',
  os: 'debian-12',
  location: 'onprem',
  owner: 'Plataforma',
  tags: ['web'],
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}

const renderForm = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ServerForm server={server} onClose={() => undefined} />
    </QueryClientProvider>,
  )
}

describe('ServerForm', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('muestra el detalle de validación junto al campo', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(409, {
          code: 'conflict',
          message: 'Ya existe un servidor con ese hostname',
          details: [
            { field: 'hostname', issue: 'hostname duplicado' },
          ],
        }),
      ),
    )

    renderForm()
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(
      await screen.findByText('hostname duplicado'),
    ).toBeInTheDocument()
    expect(screen.getByRole('alert')).toHaveTextContent(
      'Ya existe un servidor con ese hostname',
    )
  })
})
