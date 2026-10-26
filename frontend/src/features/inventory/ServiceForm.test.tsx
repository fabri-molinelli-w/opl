import { afterEach, describe, expect, it, vi } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ServiceForm } from './ServiceForm'

const jsonResponse = (status: number, body: unknown): Response =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })

const renderForm = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ServiceForm
        serverId="0192f000-0000-7000-8000-000000000001"
        onClose={() => undefined}
      />
    </QueryClientProvider>,
  )
}

describe('ServiceForm', () => {
  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('muestra el detalle de validación del checkTarget', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(400, {
          code: 'validation_error',
          message: 'Datos inválidos',
          details: [
            {
              field: 'checkTarget',
              issue: 'El checkTarget debe ser una URL http o https válida',
            },
          ],
        }),
      ),
    )

    renderForm()
    await userEvent.type(screen.getByLabelText('Nombre'), 'API')
    await userEvent.type(
      screen.getByLabelText('Check target'),
      'ftp://api.example.com',
    )
    await userEvent.click(screen.getByRole('button', { name: 'Guardar' }))

    expect(
      await screen.findByText(
        'El checkTarget debe ser una URL http o https válida',
      ),
    ).toBeInTheDocument()
  })
})
