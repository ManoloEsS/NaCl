import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider, useAuth } from './context/AuthContext'
import ErrorBoundary from './componets/ErrorBoundary'
import { LoginPage } from './pages/LoginPage'
import { ToastProvider } from './context/ToastContext'
import { Dashboard } from './pages/Dashboard'
import { ProtectedRoute } from './componets/ProtectedRoute'
import { RegistrationPage } from './pages/RegistrationPage'

function RootRedirect() {
  const { user } = useAuth()
  if (!user) return <Navigate to='/login' replace />
  return <Navigate to='/dash' replace />
}

export const App = () => {
  return (
    <AuthProvider>
      <BrowserRouter>
        <ToastProvider>
          <ErrorBoundary>
            <Routes>
              <Route path='/login' element={<LoginPage />} />
              <Route path='/register' element={<RegistrationPage />} />
              <Route path='/' element={<RootRedirect />} />
              <Route
                path='/dash'
                element={
                  <ProtectedRoute>
                    <Dashboard />
                  </ProtectedRoute>
                }
              />
              <Route path='*' element={<Navigate to='/' replace />} />
            </Routes>
          </ErrorBoundary>
        </ToastProvider>
      </BrowserRouter>
    </AuthProvider>
  )
}
