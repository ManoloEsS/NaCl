import {
  createContext,
  useContext,
  useState,
  useEffect,
  type ReactNode
} from 'react'
import { client } from '../api/client'
import { type UserData } from '../lib/responseValidation'
import { loginService } from '../services/authServices'

interface AuthContextType {
  user: { username: string; id: string } | null
  token: string | null
  login: (username: string, password: string) => Promise<UserData>
  logout: () => void
  loading: boolean
}

const AuthContext = createContext<AuthContextType | null>(null)

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [user, setUser] = useState<{ username: string; id: string } | null>(
    () => {
      const stored = localStorage.getItem('user')
      return stored ? JSON.parse(stored) : null
    }
  )
  const [token, setToken] = useState<string | null>(() =>
    localStorage.getItem('token')
  )
  const [loading, setLoading] = useState(false)

  const logout = () => {
    setToken(null)
    setUser(null)
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  const login = async (
    username: string,
    password: string
  ): Promise<UserData> => {
    setLoading(true)
    try {
      const user: UserData = await loginService(username, password)
      const { token, ...browserUser } = user
      setToken(token)
      setUser(browserUser)
      localStorage.setItem('token', token)
      localStorage.setItem('user', JSON.stringify(browserUser))
      return user
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    const fetchUser = async () => {
      const token = localStorage.getItem('token')
      const user = localStorage.getItem('user')
      if (token && !user) {
        try {
          const res = await client.get('/me')
          setUser(res.data)
          localStorage.setItem('user', JSON.stringify(res.data))
        } catch {
          logout()
        }
      }
    }
    fetchUser()
  }, [])

  return (
    <AuthContext.Provider value={{ user, token, login, logout, loading }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
