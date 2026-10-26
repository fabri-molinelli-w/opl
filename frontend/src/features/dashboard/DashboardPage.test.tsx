import { afterEach, describe, expect, it, vi } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { render, screen } from '@testing-library/react'
import { DashboardPage } from './DashboardPage'

const jsonResponse = (status: number, body: unknown): Response =>
  new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } })

const summary = {
  generatedAt: '2026-10-26T12:00:00Z',
  servers: { total: 6, prod: 3, staging: 1, dev: 2, onprem: 4, cloud: 2 },
  services: {
    total: 6,
    up: 4,
    down: 1,
    degraded: 1,
    unknown: 0,
    problems: [{ id: 'svc-1', name: 'lab-nginx', status: 'DOWN', serverId: 'srv-1', hostname: 'lab-01.lab.local' }],
  },
  documents: {
    total: 6,
    updatedLast7Days: 2,
    byCategory: [{ category: 'runbook', count: 2 }],
    recent: [{ id: 'doc-1', title: 'Reiniciar servicio Nginx', updatedByName: 'Admin', updatedAt: '2026-10-26T10:00:00Z' }],
  },
  incidents: {
    active: 1,
    sev1: 0,
    sev2: 1,
    sev3: 0,
    sev4: 0,
    recent: [
      { id: 'inc-1', title: 'Servicio lab-nginx caído', severity: 'SEV2', status: 'open', serviceName: 'lab-nginx', openedAt: '2026-10-26T11:00:00Z' },
    ],
  },
  availability: { checks24h: 0, series: [] },
}

describe('DashboardPage', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('muestra los contadores y sus links a los listados filtrados (CA-3)', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn((input: RequestInfo | URL) =>
        Promise.resolve(
          String(input).includes('/api/dashboard/summary')
            ? jsonResponse(200, summary)
            : jsonResponse(200, { status: 'ok', database: 'up' }),
        ),
      ),
    )
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } })
    render(
      <QueryClientProvider client={client}>
        <MemoryRouter>
          <DashboardPage />
        </MemoryRouter>
      </QueryClientProvider>,
    )

    expect(await screen.findByRole('link', { name: /En prod\s*3/ })).toHaveAttribute('href', '/inventario?environment=prod')
    expect(screen.getByRole('link', { name: /Servicios caídos\s*1/ })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Servicio lab-nginx caído' })).toHaveAttribute('href', '/incidentes/inc-1')
    expect(screen.getByRole('link', { name: 'Reiniciar servicio Nginx' })).toHaveAttribute('href', '/documentacion/doc-1')
    // CA-4: sin chequeos todavía, la sección se muestra vacía en lugar de fallar.
    expect(screen.getByText('Todavía no hay chequeos en las últimas 24 h.')).toBeInTheDocument()
  })
})
