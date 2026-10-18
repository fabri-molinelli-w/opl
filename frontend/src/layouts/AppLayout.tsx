import { NavLink, Outlet } from 'react-router-dom'

const navItems = [
  { to: '/', label: 'Dashboard', end: true },
  { to: '/documentacion', label: 'Documentación', end: false },
  { to: '/inventario', label: 'Inventario', end: false },
  { to: '/usuarios', label: 'Usuarios', end: false },
] as const

export const AppLayout = () => (
  <div className="flex min-h-screen bg-slate-100 text-slate-900">
    <aside className="flex w-60 flex-col bg-slate-900 text-slate-100">
      <div className="px-5 py-6 text-lg font-semibold tracking-wide">OPL</div>
      <nav className="flex flex-1 flex-col gap-1 px-3">
        {navItems.map((item) => (
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
    <main className="flex-1 p-8">
      <Outlet />
    </main>
  </div>
)
