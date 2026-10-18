import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from 'react-router-dom'
import { queryClient } from '@/api/queryClient'
import { router } from '@/routes'
import '@/index.css'

const rootElement = document.getElementById('root')

if (rootElement === null) {
  throw new Error('No se encontró el elemento #root')
}

createRoot(rootElement).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </StrictMode>,
)
