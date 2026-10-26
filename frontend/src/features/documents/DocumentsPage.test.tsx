import { afterEach, describe, expect, it, vi } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { render, screen } from '@testing-library/react'
import { DocumentsPage } from './DocumentsPage'

const jsonResponse = (status: number, body: unknown): Response =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })

const page = {
  items: [
    {
      id: '0192f000-0000-7000-8000-000000000101',
      title: 'Configurar VPN en notebook',
      category: 'guia',
      tags: ['vpn', 'notebooks'],
      version: 3,
      updatedByName: 'Usuario Demo',
      createdAt: '2026-01-01T00:00:00Z',
      updatedAt: '2026-01-02T00:00:00Z',
    },
  ],
  total: 1,
  page: 1,
  pageSize: 20,
}

const renderPage = (initialEntry = '/documentacion') => {
  const fetchMock = vi.fn((input: RequestInfo | URL) => {
    const url = String(input)
    if (url.includes('/api/documents/tags')) {
      return Promise.resolve(jsonResponse(200, ['notebooks', 'vpn']))
    }
    return Promise.resolve(jsonResponse(200, page))
  })
  vi.stubGlobal('fetch', fetchMock)
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[initialEntry]}>
        <DocumentsPage />
      </MemoryRouter>
    </QueryClientProvider>,
  )
  return fetchMock
}

describe('DocumentsPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('CA-1: lista los documentos con categoría, etiquetas y autor', async () => {
    renderPage()
    expect(await screen.findByRole('link', { name: 'Configurar VPN en notebook' })).toHaveAttribute(
      'href',
      '/documentacion/0192f000-0000-7000-8000-000000000101',
    )
    expect(screen.getAllByText('Guía')).toHaveLength(2) // filtro + tarjeta
    expect(screen.getByRole('button', { name: '#vpn' })).toBeInTheDocument()
    expect(screen.getByText(/editado por Usuario Demo/)).toBeInTheDocument()
  })

  it('envía la búsqueda y el orden al backend', async () => {
    const fetchMock = renderPage('/documentacion?q=vpn&sort=title-asc')
    await screen.findByText('Configurar VPN en notebook')
    const urls = fetchMock.mock.calls.map(([input]) => String(input))
    const listUrl = urls.find((u) => u.includes('/api/documents?')) ?? ''
    expect(listUrl).toContain('q=vpn')
    expect(listUrl).toContain('sort=title')
    expect(listUrl).toContain('order=asc')
  })
})
