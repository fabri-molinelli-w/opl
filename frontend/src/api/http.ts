import { clearSession, getToken } from '@/auth/session'
import { redirectToLogin } from '@/auth/redirect'

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

export class ApiError extends Error {
  readonly status: number
  readonly data: unknown

  constructor(status: number, message: string, data?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.data = data
  }
}

export const isApiError = (value: unknown): value is ApiError =>
  value instanceof ApiError

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === 'object' && value !== null

const extractMessage = (data: unknown): string => {
  if (isRecord(data) && typeof data.message === 'string') {
    return data.message
  }
  return 'Error inesperado del servidor'
}

export const customInstance = async <T>(
  url: string,
  options: RequestInit = {},
): Promise<T> => {
  const token = getToken()

  const response = await fetch(`${BASE_URL}${url}`, {
    ...options,
    headers: {
      Accept: 'application/json',
      ...(options.body != null ? { 'Content-Type': 'application/json' } : {}),
      ...(token !== null ? { Authorization: `Bearer ${token}` } : {}),
      ...options.headers,
    },
  })

  const text = await response.text()
  const data: unknown = text.length > 0 ? JSON.parse(text) : undefined

  if (!response.ok && response.status !== 503) {
    if (response.status === 401 && token !== null) {
      clearSession()
      redirectToLogin()
    }
    const message =
      response.status === 403 ? 'No tenés permisos' : extractMessage(data)
    throw new ApiError(response.status, message, data)
  }

  return { data, status: response.status, headers: response.headers } as T
}
