import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import type { Role, User } from '@/api/gen/model'
import { AuthProvider } from '@/auth/AuthContext'
import { setStoredUser, setToken } from '@/auth/session'
import { InventoryPage } from './InventoryPage'

const jsonResponse = (status: number, body: unknown): Response =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })

const userWithRole = (role: Role): User => ({
  id: '0192f000-0000-7000-8000-00000000000a',
  email: `${role}@example.com`,
  name: 'Usuario Demo',
  role,
  active: true,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
})

const server = {
  id: '0192f000-0000-7000-8000-000000000001',
  hostname: 'web-01.example.com',
  ip: '192.0.2.10',
  environment: 'prod',
  os: 'debian-12',
  location: 'onprem',
  owner: 'Plataforma',
  tags: ['web', 'critical'],
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}

const renderPage = (role: Role) => {
  setToken('jwt')
  setStoredUser(userWithRole(role))
  const fetchMock = vi.fn((input: RequestInfo | URL) => {
    const url = String(input)
    if (url.includes('/api/auth/me')) {
      return Promise.resolve(jsonResponse(200, userWithRole(role)))
    }
    return Promise.resolve(
      jsonResponse(200, { items: [server], total: 1, page: 1, pageSize: 20 }),
    )
  })
  vi.stubGlobal('fetch', fetchMock)

  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return {
    fetchMock,
    ...render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={['/inventario?environment=prod']}>
          <AuthProvider>
            <InventoryPage />
          </AuthProvider>
        </MemoryRouter>
      </QueryClientProvider>,
    ),
  }
}

describe('InventoryPage', () => {
  beforeEach(() => {
    sessionStorage.clear()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('muestra los servidores y la ficha enlazada', async () => {
    renderPage('soporte')

    expect(await screen.findByText('web-01.example.com')).toBeInTheDocument()
    expect(screen.getByText('192.0.2.10')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Ver ficha' })).toHaveAttribute(
      'href',
      `/inventario/${server.id}`,
    )
  })

  it('soporte no ve acciones de escritura', async () => {
    renderPage('soporte')

    await screen.findByText('web-01.example.com')
    expect(
      screen.queryByRole('button', { name: 'Nuevo servidor' }),
    ).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Editar' })).not.toBeInTheDocument()
  })

  it('sysadmin ve acciones de escritura', async () => {
    renderPage('sysadmin')

    await screen.findByText('web-01.example.com')
    expect(
      screen.getByRole('button', { name: 'Nuevo servidor' }),
    ).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Editar' })).toBeInTheDocument()
  })

  it('actualiza los filtros en la URL y reinicia la página', async () => {
    const { fetchMock } = renderPage('soporte')
    await screen.findByText('web-01.example.com')

    await userEvent.selectOptions(
      screen.getByLabelText('Environment'),
      'dev',
    )

    await waitFor(() => {
      const urls = fetchMock.mock.calls.map(([input]) => String(input))
      expect(
        urls.some(
          (url) =>
            url.includes('/api/servers?') &&
            url.includes('environment=dev') &&
            url.includes('page=1'),
        ),
      ).toBe(true)
    })
  })
})
