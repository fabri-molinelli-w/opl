import { createBrowserRouter } from 'react-router-dom'
import { AppLayout } from '@/layouts/AppLayout'
import { DashboardPage } from '@/features/dashboard/DashboardPage'
import { DocumentsPage } from '@/features/documents/DocumentsPage'
import { InventoryPage } from '@/features/inventory/InventoryPage'
import { UsersPage } from '@/features/users/UsersPage'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { index: true, element: <DashboardPage /> },
      { path: 'documentacion', element: <DocumentsPage /> },
      { path: 'inventario', element: <InventoryPage /> },
      { path: 'usuarios', element: <UsersPage /> },
    ],
  },
])
