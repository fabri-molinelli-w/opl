import { useState, type FormEvent } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { getListUsersQueryKey, useCreateUser, useUpdateUser } from '@/api/gen'
import type { Role, User } from '@/api/gen/model'
import { ApiError } from '@/api/http'
import { ALL_ROLES, ROLE_LABELS } from '@/auth/roles'

interface UserFormProps {
  user?: User
  onClose: () => void
}

export const UserForm = ({ user, onClose }: UserFormProps) => {
  const isEdit = user !== undefined
  const queryClient = useQueryClient()
  const createUser = useCreateUser()
  const updateUser = useUpdateUser()

  const [name, setName] = useState(user?.name ?? '')
  const [email, setEmail] = useState(user?.email ?? '')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState<Role>(user?.role ?? 'soporte')
  const [active, setActive] = useState(user?.active ?? true)
  const [error, setError] = useState<string | null>(null)

  const isSaving = createUser.isPending || updateUser.isPending

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError(null)

    try {
      if (isEdit) {
        await updateUser.mutateAsync({
          id: user.id,
          data: { name, email, role, active },
        })
      } else {
        if (password.length < 8) {
          setError('La contraseña debe tener al menos 8 caracteres')
          return
        }
        await createUser.mutateAsync({
          data: { name, email, password, role },
        })
      }
      await queryClient.invalidateQueries({ queryKey: getListUsersQueryKey() })
      onClose()
    } catch (cause) {
      setError(
        cause instanceof ApiError
          ? cause.message
          : 'No se pudo guardar el usuario',
      )
    }
  }

  return (
    <form
      onSubmit={handleSubmit}
      className="space-y-4 rounded-xl border border-slate-200 bg-white p-6 shadow-sm"
    >
      <h2 className="text-lg font-medium">
        {isEdit ? 'Editar usuario' : 'Nuevo usuario'}
      </h2>

      <div className="space-y-1">
        <label htmlFor="user-name" className="text-sm font-medium">
          Nombre
        </label>
        <input
          id="user-name"
          required
          value={name}
          onChange={(event) => setName(event.target.value)}
          className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-500 focus:outline-none"
        />
      </div>

      <div className="space-y-1">
        <label htmlFor="user-email" className="text-sm font-medium">
          Email
        </label>
        <input
          id="user-email"
          type="email"
          required
          value={email}
          onChange={(event) => setEmail(event.target.value)}
          className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-500 focus:outline-none"
        />
      </div>

      {!isEdit && (
        <div className="space-y-1">
          <label htmlFor="user-password" className="text-sm font-medium">
            Contraseña
          </label>
          <input
            id="user-password"
            type="password"
            autoComplete="new-password"
            required
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-500 focus:outline-none"
          />
        </div>
      )}

      <div className="space-y-1">
        <label htmlFor="user-role" className="text-sm font-medium">
          Rol
        </label>
        <select
          id="user-role"
          value={role}
          onChange={(event) => setRole(event.target.value as Role)}
          className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm focus:border-slate-500 focus:outline-none"
        >
          {ALL_ROLES.map((value) => (
            <option key={value} value={value}>
              {ROLE_LABELS[value]}
            </option>
          ))}
        </select>
      </div>

      {isEdit && (
        <label className="flex items-center gap-2 text-sm font-medium">
          <input
            type="checkbox"
            checked={active}
            onChange={(event) => setActive(event.target.checked)}
          />
          Usuario activo
        </label>
      )}

      {error !== null && (
        <p role="alert" className="text-sm text-red-600">
          {error}
        </p>
      )}

      <div className="flex gap-3">
        <button
          type="submit"
          disabled={isSaving}
          className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:opacity-60"
        >
          {isSaving ? 'Guardando…' : 'Guardar'}
        </button>
        <button
          type="button"
          onClick={onClose}
          className="rounded-md border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
        >
          Cancelar
        </button>
      </div>
    </form>
  )
}
