import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from './useAuth'

export const RequireAuth = () => {
  const { status } = useAuth()
  const location = useLocation()

  if (status === 'loading') {
    return <div className="p-8 text-sm text-slate-500">Cargando sesión…</div>
  }

  if (status !== 'authenticated') {
    return (
      <Navigate to="/login" replace state={{ from: location.pathname }} />
    )
  }

  return <Outlet />
}
