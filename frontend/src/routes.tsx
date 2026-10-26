import { createBrowserRouter } from 'react-router-dom'
import { RequireAuth } from '@/auth/RequireAuth'
import { RequireRole } from '@/auth/RequireRole'
import { AppLayout } from '@/layouts/AppLayout'
import { AccountPage } from '@/features/account/AccountPage'
import { LoginPage } from '@/features/auth/LoginPage'
import { DashboardPage } from '@/features/dashboard/DashboardPage'
import { DocumentDetailPage } from '@/features/documents/DocumentDetailPage'
import { DocumentEditorPage } from '@/features/documents/DocumentEditorPage'
import { DocumentsPage } from '@/features/documents/DocumentsPage'
import { IncidentDetailPage } from '@/features/incidents/IncidentDetailPage'
import { IncidentsPage } from '@/features/incidents/IncidentsPage'
import { InventoryPage } from '@/features/inventory/InventoryPage'
import { ServerDetailPage } from '@/features/inventory/ServerDetailPage'
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
          { path: 'documentacion/nuevo', element: <DocumentEditorPage /> },
          { path: 'documentacion/:id', element: <DocumentDetailPage /> },
          { path: 'documentacion/:id/editar', element: <DocumentEditorPage /> },
          { path: 'incidentes', element: <IncidentsPage /> },
          { path: 'incidentes/:id', element: <IncidentDetailPage /> },
          { path: 'inventario', element: <InventoryPage /> },
          { path: 'inventario/:id', element: <ServerDetailPage /> },
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
