import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { render, screen } from '@testing-library/react'
import type { Role, User } from '@/api/gen/model'
import { setStoredUser, setToken } from '@/auth/session'
import { AuthProvider } from '@/auth/AuthContext'
import { AppLayout } from './AppLayout'

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

const renderLayout = (role: Role) => {
  setToken('jwt')
  setStoredUser(userWithRole(role))
  vi.stubGlobal(
    'fetch',
    vi.fn().mockResolvedValue(jsonResponse(200, userWithRole(role))),
  )

  return render(
    <MemoryRouter initialEntries={['/']}>
      <AuthProvider>
        <Routes>
          <Route element={<AppLayout />}>
            <Route index element={<div>Inicio</div>} />
            <Route path="usuarios" element={<div>UsuariosPage</div>} />
          </Route>
        </Routes>
      </AuthProvider>
    </MemoryRouter>,
  )
}

describe('AppLayout', () => {
  beforeEach(() => {
    sessionStorage.clear()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('un admin ve la sección Usuarios', async () => {
    renderLayout('admin')

    expect(
      await screen.findByRole('link', { name: 'Usuarios' }),
    ).toBeInTheDocument()
  })

  it('un usuario soporte no ve la sección Usuarios', async () => {
    renderLayout('soporte')

    expect(
      await screen.findByRole('link', { name: 'Dashboard' }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('link', { name: 'Usuarios' }),
    ).not.toBeInTheDocument()
  })

  it('muestra el nombre y el rol del usuario autenticado', async () => {
    renderLayout('sysadmin')

    expect(await screen.findByText('Usuario Demo')).toBeInTheDocument()
    expect(screen.getByText('SysAdmin')).toBeInTheDocument()
  })
})
