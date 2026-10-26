import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import {
  getCountMyUnreadNotificationsQueryKey,
  getListMyNotificationsQueryKey,
  useCountMyUnreadNotifications,
  useListMyNotifications,
  useMarkAllMyNotificationsRead,
  useMarkMyNotificationRead,
} from '@/api/gen'
import type { Notification } from '@/api/gen/model'

const formatWhen = (value: string): string =>
  new Intl.DateTimeFormat('es-AR', { dateStyle: 'short', timeStyle: 'short' }).format(new Date(value))

/** Campana de notificaciones internas (CU-08): contador cada 30 s y panel con las últimas 20. */
export const NotificationBell = () => {
  const [open, setOpen] = useState(false)
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const count = useCountMyUnreadNotifications({ query: { refetchInterval: 30_000 } })
  const list = useListMyNotifications(undefined, { query: { enabled: open } })
  const markRead = useMarkMyNotificationRead()
  const markAll = useMarkAllMyNotificationsRead()

  const unread = count.data?.status === 200 ? count.data.data.count : 0
  const items = list.data?.status === 200 ? list.data.data : []

  const refresh = async () => {
    await queryClient.invalidateQueries({ queryKey: getCountMyUnreadNotificationsQueryKey() })
    await queryClient.invalidateQueries({ queryKey: getListMyNotificationsQueryKey() })
  }

  const openNotification = async (n: Notification) => {
    if (!n.read) {
      try {
        await markRead.mutateAsync({ id: n.id })
      } catch {
        // Si falla, igual se navega: marcarla leída no es crítico.
      }
      await refresh()
    }
    setOpen(false)
    if (n.link !== '') void navigate(n.link)
  }

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        aria-expanded={open}
        aria-label={`Notificaciones (${unread} sin leer)`}
        className="relative rounded-md border border-slate-300 px-3 py-1.5 text-sm hover:bg-slate-100"
      >
        <span aria-hidden="true">🔔</span>
        {unread > 0 && (
          <span className="absolute -right-2 -top-2 min-w-5 rounded-full bg-red-600 px-1.5 text-center text-xs font-semibold text-white">
            {unread > 99 ? '99+' : unread}
          </span>
        )}
      </button>
      {open && (
        <div className="absolute right-0 z-30 mt-2 w-96 rounded-xl border border-slate-200 bg-white shadow-xl">
          <div className="flex items-center justify-between border-b border-slate-100 px-4 py-2">
            <span className="text-sm font-semibold">Notificaciones</span>
            <button
              type="button"
              disabled={unread === 0}
              onClick={() =>
                void (async () => {
                  await markAll.mutateAsync()
                  await refresh()
                })()
              }
              className="text-xs text-sky-700 hover:underline disabled:text-slate-400 disabled:no-underline"
            >
              Marcar todas como leídas
            </button>
          </div>
          <ul className="max-h-96 overflow-y-auto">
            {items.length === 0 && <li className="px-4 py-6 text-center text-sm text-slate-500">No hay notificaciones.</li>}
            {items.map((n) => (
              <li key={n.id}>
                <button
                  type="button"
                  onClick={() => void openNotification(n)}
                  className={`block w-full border-b border-slate-50 px-4 py-2 text-left text-sm hover:bg-slate-50 ${
                    n.read ? 'text-slate-500' : 'bg-sky-50/60 font-medium'
                  }`}
                >
                  {n.title}
                  {n.body !== '' && <span className="block truncate text-xs font-normal text-slate-500">{n.body}</span>}
                  <span className="block text-xs font-normal text-slate-400">{formatWhen(n.createdAt)}</span>
                </button>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}
