import type { ReactNode } from 'react'
import { useAuth } from '../context/AuthContext'
import { Navigate } from 'react-router-dom'

interface Props {
  children: ReactNode
}

export const ProtectedRoute = ({ children }: Props) => {
  const { user } = useAuth()

  if (!user) {
    return <Navigate to='/login' replace />
  }
  return <>{children}</>
}
