import { createBrowserRouter } from 'react-router-dom'
import { RequireAuth } from '@/auth/RequireAuth'
import { RequireRole } from '@/auth/RequireRole'
import { AppLayout } from '@/layouts/AppLayout'
import { AccountPage } from '@/features/account/AccountPage'
import { LoginPage } from '@/features/auth/LoginPage'
import { DashboardPage } from '@/features/dashboard/DashboardPage'
import { DocumentsPage } from '@/features/documents/DocumentsPage'
import { InventoryPage } from '@/features/inventory/InventoryPage'
import { UsersPage } from '@/features/users/UsersPage'

export const router = createBrowserRouter([
  { path: '/login', element: <LoginPage /> },
  {
    element: <RequireAuth />,
    children: [
      {
        path: '/',
        element: <AppLayout />,
        children: [
          { index: true, element: <DashboardPage /> },
          { path: 'documentacion', element: <DocumentsPage /> },
          { path: 'inventario', element: <InventoryPage /> },
          { path: 'mi-cuenta', element: <AccountPage /> },
          {
            element: <RequireRole minRole="admin" />,
            children: [{ path: 'usuarios', element: <UsersPage /> }],
          },
        ],
      },
    ],
  },
])
