import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { getToken, setToken } from '@/auth/session'
import { redirectToLogin } from '@/auth/redirect'
import { ApiError, customInstance } from './http'

vi.mock('@/auth/redirect', () => ({ redirectToLogin: vi.fn() }))

const jsonResponse = (status: number, body: unknown): Response =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })

describe('customInstance', () => {
  beforeEach(() => {
    sessionStorage.clear()
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('limpia la sesión y redirige al login ante 401 con token presente', async () => {
    setToken('jwt')
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(401, { code: 'unauthorized', message: 'Token inválido' }),
      ),
    )

    await expect(customInstance('/api/users')).rejects.toBeInstanceOf(ApiError)

    expect(getToken()).toBeNull()
    expect(redirectToLogin).toHaveBeenCalledTimes(1)
  })

  it('no redirige en 401 si no había token (login fallido)', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(401, {
          code: 'unauthorized',
          message: 'Credenciales inválidas',
        }),
      ),
    )

    await expect(customInstance('/api/auth/login')).rejects.toMatchObject({
      status: 401,
    })

    expect(redirectToLogin).not.toHaveBeenCalled()
  })

  it('traduce el 403 a un mensaje de permisos', async () => {
    setToken('jwt')
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        jsonResponse(403, { code: 'forbidden', message: 'No autorizado' }),
      ),
    )

    await expect(customInstance('/api/users')).rejects.toMatchObject({
      status: 403,
      message: 'No tenés permisos',
    })
  })
})
