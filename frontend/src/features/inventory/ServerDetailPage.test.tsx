import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { render, screen } from '@testing-library/react'
import type { User } from '@/api/gen/model'
import { AuthProvider } from '@/auth/AuthContext'
import { setStoredUser, setToken } from '@/auth/session'
import { ServerDetailPage } from './ServerDetailPage'

const jsonResponse = (status: number, body: unknown): Response =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })

const user: User = {
  id: '0192f000-0000-7000-8000-00000000000a',
  email: 'sysadmin@example.com',
  name: 'Usuario Demo',
  role: 'sysadmin',
  active: true,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}

const server = {
  id: '0192f000-0000-7000-8000-000000000001',
  hostname: 'web-01.example.com',
  ip: '192.0.2.10',
  environment: 'prod',
  os: 'debian-12',
  location: 'onprem',
  owner: 'Plataforma',
  tags: ['web', 'critical'],
  services: [
    {
      id: '0192f000-0000-7000-8000-000000000002',
      serverId: '0192f000-0000-7000-8000-000000000001',
      name: 'Portal',
      type: 'web',
      port: 443,
      checkType: 'http',
      checkTarget: 'https://web-01.example.com/health',
      expectedStatus: 200,
      intervalSeconds: 30,
      latencyThresholdMs: 1000,
      currentStatus: 'UP',
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-01T00:00:00Z',
    },
  ],
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}

const renderDetail = () => {
  setToken('jwt')
  setStoredUser(user)
  vi.stubGlobal(
    'fetch',
    vi.fn((input: RequestInfo | URL) => {
      const url = String(input)
      if (url.includes('/api/auth/me')) {
        return Promise.resolve(jsonResponse(200, user))
      }
      return Promise.resolve(jsonResponse(200, server))
    }),
  )

  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[`/inventario/${server.id}`]}>
        <AuthProvider>
          <Routes>
            <Route path="/inventario/:id" element={<ServerDetailPage />} />
          </Routes>
        </AuthProvider>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('ServerDetailPage', () => {
  beforeEach(() => {
    sessionStorage.clear()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('muestra la ficha y sus servicios', async () => {
    renderDetail()

    expect(await screen.findByText('web-01.example.com')).toBeInTheDocument()
    expect(screen.getByText('Portal')).toBeInTheDocument()
    expect(screen.getByText('UP')).toBeInTheDocument()
    expect(screen.getByText('https://web-01.example.com/health')).toBeInTheDocument()
  })

  it('sysadmin ve las acciones de servicios', async () => {
    renderDetail()

    await screen.findByText('Portal')
    expect(
      screen.getByRole('button', { name: 'Nuevo servicio' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Editar' })).toBeInTheDocument()
  })
})
