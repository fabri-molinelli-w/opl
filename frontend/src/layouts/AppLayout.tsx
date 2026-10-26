import { useState } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import type { Role } from '@/api/gen/model'
import { hasRole, ROLE_LABELS } from '@/auth/roles'
import { useAuth } from '@/auth/useAuth'
import { ChatPanel } from '@/features/chat/ChatPanel'

interface NavItem {
  to: string
  label: string
  end: boolean
  minRole: Role
}

const NAV_ITEMS: readonly NavItem[] = [
  { to: '/', label: 'Dashboard', end: true, minRole: 'soporte' },
  { to: '/documentacion', label: 'Documentación', end: false, minRole: 'soporte' },
  { to: '/incidentes', label: 'Incidentes', end: false, minRole: 'soporte' },
  { to: '/inventario', label: 'Inventario', end: false, minRole: 'soporte' },
  { to: '/usuarios', label: 'Usuarios', end: false, minRole: 'admin' },
  { to: '/mi-cuenta', label: 'Mi cuenta', end: false, minRole: 'soporte' },
]

export const AppLayout = () => {
  const { user, logout } = useAuth()
  const items = NAV_ITEMS.filter((item) => hasRole(user?.role, item.minRole))
  const [chatOpen, setChatOpen] = useState(false)

  return (
    <div className="flex min-h-screen bg-slate-100 text-slate-900">
      <aside className="flex w-60 flex-col bg-slate-900 text-slate-100">
        <div className="px-5 py-6 text-lg font-semibold tracking-wide">OPL</div>
        <nav className="flex flex-1 flex-col gap-1 px-3">
          {items.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                `rounded-md px-3 py-2 text-sm transition-colors ${
                  isActive
                    ? 'bg-slate-700 text-white'
                    : 'text-slate-300 hover:bg-slate-800 hover:text-white'
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <div className="flex flex-1 flex-col">
        <header className="flex items-center justify-end gap-4 border-b border-slate-200 bg-white px-8 py-4">
          <button
            type="button"
            onClick={() => setChatOpen((open) => !open)}
            aria-expanded={chatOpen}
            className="rounded-md bg-sky-700 px-3 py-1.5 text-sm font-medium text-white hover:bg-sky-800"
          >
            OPSBot
          </button>
          {user !== null && (
            <div className="text-right">
              <p className="text-sm font-medium">{user.name}</p>
              <p className="text-xs text-slate-500">{ROLE_LABELS[user.role]}</p>
            </div>
          )}
          <button
            type="button"
            onClick={logout}
            className="rounded-md border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-100"
          >
            Cerrar sesión
          </button>
        </header>
        <main className="flex-1 p-8">
          <Outlet />
        </main>
        {chatOpen && <ChatPanel onClose={() => setChatOpen(false)} />}
      </div>
    </div>
  )
}
