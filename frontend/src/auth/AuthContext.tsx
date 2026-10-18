import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import { getCurrentUser, login as loginRequest } from '@/api/gen'
import type { User } from '@/api/gen/model'
import { AuthContext, type AuthContextValue, type AuthStatus } from './context'
import {
  clearSession,
  getStoredUser,
  getToken,
  setStoredUser,
  setToken,
} from './session'

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<User | null>(() =>
    getToken() === null ? null : getStoredUser(),
  )
  const [status, setStatus] = useState<AuthStatus>(() =>
    getToken() === null ? 'anonymous' : 'loading',
  )

  useEffect(() => {
    if (getToken() === null) {
      return
    }

    let active = true

    const bootstrap = async () => {
      try {
        const response = await getCurrentUser()
        if (!active) {
          return
        }
        if (response.status === 200) {
          setUser(response.data)
          setStoredUser(response.data)
          setStatus('authenticated')
          return
        }
      } catch {
        if (!active) {
          return
        }
      }
      clearSession()
      setUser(null)
      setStatus('anonymous')
    }

    void bootstrap()

    return () => {
      active = false
    }
  }, [])

  const login = useCallback(async (email: string, password: string) => {
    const response = await loginRequest({ email, password })
    if (response.status !== 200) {
      throw new Error(response.data.message)
    }
    setToken(response.data.token)
    setStoredUser(response.data.user)
    setUser(response.data.user)
    setStatus('authenticated')
  }, [])

  const logout = useCallback(() => {
    clearSession()
    setUser(null)
    setStatus('anonymous')
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({ user, status, login, logout }),
    [user, status, login, logout],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
