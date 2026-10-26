import { afterEach, describe, expect, it, vi } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { NotificationBell } from './NotificationBell'

const jsonResponse = (status: number, body?: unknown): Response =>
  new Response(body === undefined ? null : JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })

describe('NotificationBell', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('muestra el contador y al abrir una notificación la marca leída y navega', async () => {
    let unread = 1
    const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input)
      if (url.endsWith('/unread-count')) return Promise.resolve(jsonResponse(200, { count: unread }))
      if (url.endsWith('/read') && init?.method === 'POST') {
        unread = 0
        return Promise.resolve(jsonResponse(204))
      }
      return Promise.resolve(
        jsonResponse(200, [
          {
            id: 'n1',
            eventType: 'service.down',
            title: 'Servicio lab-nginx caído',
            body: 'connection refused',
            link: '/inventario/srv-1',
            read: unread === 0,
            createdAt: '2026-10-26T10:00:00Z',
          },
        ]),
      )
    })
    vi.stubGlobal('fetch', fetchMock)
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={client}>
        <MemoryRouter initialEntries={['/']}>
          <Routes>
            <Route path="/" element={<NotificationBell />} />
            <Route path="/inventario/:id" element={<p>Ficha del servidor</p>} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>,
    )
    const user = userEvent.setup()

    await user.click(await screen.findByRole('button', { name: 'Notificaciones (1 sin leer)' }))
    await user.click(await screen.findByRole('button', { name: /Servicio lab-nginx caído/ }))

    expect(await screen.findByText('Ficha del servidor')).toBeInTheDocument()
    expect(fetchMock.mock.calls.some(([u, i]) => String(u).endsWith('/n1/read') && i?.method === 'POST')).toBe(true)
  })
})
