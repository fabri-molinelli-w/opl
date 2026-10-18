import { useState, type FormEvent } from 'react'
import { Navigate, useLocation, useNavigate } from 'react-router-dom'
import { ApiError } from '@/api/http'
import { useAuth } from '@/auth/useAuth'

const resolveRedirect = (state: unknown): string => {
  if (typeof state === 'object' && state !== null && 'from' in state) {
    const from = (state as { from?: unknown }).from
    if (typeof from === 'string' && from.length > 0) {
      return from
    }
  }
  return '/'
}

export const LoginPage = () => {
  const { login, status } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  if (status === 'authenticated') {
    return <Navigate to="/" replace />
  }

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await login(email, password)
      navigate(resolveRedirect(location.state), { replace: true })
    } catch (cause) {
      setError(
        cause instanceof ApiError ? cause.message : 'No se pudo iniciar sesión',
      )
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-100 p-6">
      <form
        onSubmit={handleSubmit}
        className="w-full max-w-sm space-y-4 rounded-xl bg-white p-8 shadow-sm"
      >
        <div className="space-y-1">
          <h1 className="text-xl font-semibold">OPL</h1>
          <p className="text-sm text-slate-500">Iniciá sesión para continuar</p>
        </div>

        <div className="space-y-1">
          <label htmlFor="email" className="text-sm font-medium">
            Email
          </label>
          <input
            id="email"
            type="email"
            autoComplete="username"
            required
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-500 focus:outline-none"
          />
        </div>

        <div className="space-y-1">
          <label htmlFor="password" className="text-sm font-medium">
            Contraseña
          </label>
          <input
            id="password"
            type="password"
            autoComplete="current-password"
            required
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-500 focus:outline-none"
          />
        </div>

        {error !== null && (
          <p role="alert" className="text-sm text-red-600">
            {error}
          </p>
        )}

        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:opacity-60"
        >
          {submitting ? 'Ingresando…' : 'Ingresar'}
        </button>
      </form>
    </div>
  )
}
