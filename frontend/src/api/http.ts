const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

let authToken: string | null = null

export const setAuthToken = (token: string | null): void => {
  authToken = token
}

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
  const response = await fetch(`${BASE_URL}${url}`, {
    ...options,
    headers: {
      Accept: 'application/json',
      ...(options.body != null ? { 'Content-Type': 'application/json' } : {}),
      ...(authToken !== null ? { Authorization: `Bearer ${authToken}` } : {}),
      ...options.headers,
    },
  })

  const text = await response.text()
  const data: unknown = text.length > 0 ? JSON.parse(text) : undefined

  if (!response.ok && response.status !== 503) {
    throw new ApiError(response.status, extractMessage(data), data)
  }

  return { data, status: response.status, headers: response.headers } as T
}
