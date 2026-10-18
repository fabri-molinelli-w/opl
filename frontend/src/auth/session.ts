import type { Role, User } from '@/api/gen/model'

const TOKEN_KEY = 'opl.token'
const USER_KEY = 'opl.user'

const ROLE_VALUES: readonly Role[] = ['admin', 'sysadmin', 'soporte']

const isRole = (value: unknown): value is Role =>
  typeof value === 'string' && (ROLE_VALUES as readonly string[]).includes(value)

const isUser = (value: unknown): value is User => {
  if (typeof value !== 'object' || value === null) {
    return false
  }
  const record = value as Record<string, unknown>
  return (
    typeof record.id === 'string' &&
    typeof record.email === 'string' &&
    typeof record.name === 'string' &&
    typeof record.active === 'boolean' &&
    isRole(record.role) &&
    typeof record.createdAt === 'string' &&
    typeof record.updatedAt === 'string'
  )
}

export const getToken = (): string | null => {
  try {
    return sessionStorage.getItem(TOKEN_KEY)
  } catch {
    return null
  }
}

export const setToken = (token: string | null): void => {
  try {
    if (token === null) {
      sessionStorage.removeItem(TOKEN_KEY)
    } else {
      sessionStorage.setItem(TOKEN_KEY, token)
    }
  } catch {
    // sessionStorage unavailable: the session simply stays in memory.
  }
}

export const getStoredUser = (): User | null => {
  try {
    const raw = sessionStorage.getItem(USER_KEY)
    if (raw === null) {
      return null
    }
    const parsed: unknown = JSON.parse(raw)
    return isUser(parsed) ? parsed : null
  } catch {
    return null
  }
}

export const setStoredUser = (user: User | null): void => {
  try {
    if (user === null) {
      sessionStorage.removeItem(USER_KEY)
    } else {
      sessionStorage.setItem(USER_KEY, JSON.stringify(user))
    }
  } catch {
    // sessionStorage unavailable: the session simply stays in memory.
  }
}

export const clearSession = (): void => {
  setToken(null)
  setStoredUser(null)
}
