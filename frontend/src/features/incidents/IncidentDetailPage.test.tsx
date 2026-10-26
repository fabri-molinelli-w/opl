import { afterEach, describe, expect, it, vi } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { User } from '@/api/gen/model'
import { AuthProvider } from '@/auth/AuthContext'
import { setStoredUser, setToken } from '@/auth/session'
import { IncidentDetailPage } from './IncidentDetailPage'

const jsonResponse = (status: number, body: unknown): Response =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })

const user: User = {
  id: 'user-1',
  email: 'soporte@example.com',
  name: 'Soporte Demo',
  role: 'soporte',
  active: true,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}

const incident = {
  id: 'inc-1',
  title: 'Servicio lab-nginx caído',
  description: 'Abierto automáticamente tras 3 chequeos fallidos consecutivos.',
  serviceId: 'svc-1',
  serviceName: 'lab-nginx',
  severity: 'SEV2',
  status: 'in_progress',
  assigneeName: '',
  rootCause: '',
  autoCreated: true,
  openedAt: '2026-10-26T10:00:00Z',
  createdAt: '2026-10-26T10:00:00Z',
  updatedAt: '2026-10-26T10:00:00Z',
  events: [
    { id: 'e1', kind: 'system', body: 'Incidente abierto automáticamente', authorName: 'Sistema', createdAt: '2026-10-26T10:00:00Z' },
  ],
  allowedTransitions: ['resolved'],
}

describe('IncidentDetailPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('CA-1: resolver sin causa raíz muestra el 409 del backend', async () => {
    setToken('jwt')
    setStoredUser(user)
    const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input)
      if (url.endsWith('/transition') && init?.method === 'POST') {
        return Promise.resolve(
          jsonResponse(409, {
            code: 'conflict',
            message: 'Para resolver el incidente hay que indicar la causa raíz',
            details: [{ field: 'rootCause', issue: 'obligatoria' }],
          }),
        )
      }
      if (url.includes('/api/auth/me')) return Promise.resolve(jsonResponse(200, user))
      return Promise.resolve(jsonResponse(200, incident))
    })
    vi.stubGlobal('fetch', fetchMock)
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={client}>
        <MemoryRouter initialEntries={['/incidentes/inc-1']}>
          <AuthProvider>
            <Routes>
              <Route path="/incidentes/:id" element={<IncidentDetailPage />} />
            </Routes>
          </AuthProvider>
        </MemoryRouter>
      </QueryClientProvider>,
    )

    expect(await screen.findByRole('heading', { name: 'Servicio lab-nginx caído' })).toBeInTheDocument()
    expect(screen.getByText('Incidente abierto automáticamente')).toBeInTheDocument()
    // Solo se ofrecen las transiciones válidas (no "Cerrar" desde en curso).
    expect(screen.queryByRole('button', { name: 'Cerrar' })).not.toBeInTheDocument()

    await userEvent.setup().click(screen.getByRole('button', { name: 'Resolver' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('Para resolver el incidente hay que indicar la causa raíz')
  })
})
