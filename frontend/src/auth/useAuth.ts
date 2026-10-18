import { useContext } from 'react'
import { AuthContext, type AuthContextValue } from './context'

export const useAuth = (): AuthContextValue => {
  const context = useContext(AuthContext)
  if (context === null) {
    throw new Error('useAuth debe usarse dentro de un AuthProvider')
  }
  return context
}
