import { useState } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import {
  getListUsersQueryKey,
  useListUsers,
  useUpdateUser,
} from '@/api/gen'
import type { User } from '@/api/gen/model'
import { ApiError, isApiError } from '@/api/http'
import { ROLE_LABELS } from '@/auth/roles'
import { UserForm } from './UserForm'

export const UsersPage = () => {
  const queryClient = useQueryClient()
  const listResult = useListUsers()
  const updateUser = useUpdateUser()

  const [editing, setEditing] = useState<User | null>(null)
  const [creating, setCreating] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)

  const response = listResult.data
  const users = response !== undefined && response.status === 200 ? response.data : []

  const refresh = () =>
    queryClient.invalidateQueries({ queryKey: getListUsersQueryKey() })

  const closeForm = () => {
    setCreating(false)
    setEditing(null)
  }

  const toggleActive = async (user: User) => {
    setActionError(null)
    try {
      await updateUser.mutateAsync({
        id: user.id,
        data: { active: !user.active },
      })
      await refresh()
    } catch (cause) {
      setActionError(
        cause instanceof ApiError
          ? cause.message
          : 'No se pudo actualizar el usuario',
      )
    }
  }

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Usuarios</h1>
        <button
          type="button"
          onClick={() => {
            setEditing(null)
            setCreating(true)
          }}
          className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800"
        >
          Nuevo usuario
        </button>
      </div>

      {(creating || editing !== null) && (
        <UserForm user={editing ?? undefined} onClose={closeForm} />
      )}

      {actionError !== null && (
        <p role="alert" className="text-sm text-red-600">
          {actionError}
        </p>
      )}

      {listResult.isPending && (
        <p className="text-sm text-slate-500">Cargando usuarios…</p>
      )}

      {listResult.isError && (
        <p role="alert" className="text-sm text-red-600">
          {isApiError(listResult.error)
            ? listResult.error.message
            : 'No se pudieron cargar los usuarios'}
        </p>
      )}

      {!listResult.isPending && !listResult.isError && (
        <div className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-50 text-slate-500">
              <tr>
                <th className="px-4 py-3 font-medium">Nombre</th>
                <th className="px-4 py-3 font-medium">Email</th>
                <th className="px-4 py-3 font-medium">Rol</th>
                <th className="px-4 py-3 font-medium">Estado</th>
                <th className="px-4 py-3 font-medium">Acciones</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id} className="border-t border-slate-100">
                  <td className="px-4 py-3">{user.name}</td>
                  <td className="px-4 py-3">{user.email}</td>
                  <td className="px-4 py-3">{ROLE_LABELS[user.role]}</td>
                  <td className="px-4 py-3">
                    <span
                      className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                        user.active
                          ? 'bg-green-100 text-green-700'
                          : 'bg-slate-200 text-slate-600'
                      }`}
                    >
                      {user.active ? 'Activo' : 'Inactivo'}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-3">
                      <button
                        type="button"
                        onClick={() => {
                          setCreating(false)
                          setEditing(user)
                        }}
                        className="text-sm font-medium text-slate-700 hover:underline"
                      >
                        Editar
                      </button>
                      <button
                        type="button"
                        onClick={() => void toggleActive(user)}
                        className="text-sm font-medium text-slate-700 hover:underline"
                      >
                        {user.active ? 'Desactivar' : 'Activar'}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {users.length === 0 && (
            <p className="px-4 py-6 text-sm text-slate-500">
              No hay usuarios para mostrar.
            </p>
          )}
        </div>
      )}
    </section>
  )
}
