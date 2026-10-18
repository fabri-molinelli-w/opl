import { Outlet } from 'react-router-dom'
import type { Role } from '@/api/gen/model'
import { hasRole } from './roles'
import { useAuth } from './useAuth'

export const RequireRole = ({ minRole }: { minRole: Role }) => {
  const { user } = useAuth()

  if (!hasRole(user?.role, minRole)) {
    return (
      <section className="space-y-2 rounded-lg border border-red-200 bg-red-50 p-6">
        <h1 className="text-xl font-semibold text-red-800">No tenés permisos</h1>
        <p className="text-sm text-red-700">
          Tu rol no permite acceder a esta sección.
        </p>
      </section>
    )
  }

  return <Outlet />
}
