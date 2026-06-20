import {
  createContext,
  useCallback,
  useContext,
  useState,
  type ReactNode
} from 'react'

interface Toast {
  id: number
  message: string
  type: 'error' | 'success' | 'info'
}

interface ToastContextType {
  toasts: Toast[]
  showToast: (message: string, type?: Toast['type']) => void
}

const ToastContext = createContext<ToastContextType | null>(null)

let nextId = 0

export const ToastProvider = ({ children }: { children: ReactNode }) => {
  const [toasts, setToasts] = useState<Toast[]>([])

  const showToast = useCallback(
    (message: string, type: Toast['type'] = 'error') => {
      const id = nextId++
      setToasts((prev: any) => [...prev, { id, message, type }])
      setTimeout(() => {
        setToasts((prev: any) => prev.filter((t: any) => t.id !== id))
      }, 4000)
    },
    []
  )

  const dismiss = (id: number) => {
    setToasts((prev: any) => prev.filter((t: any) => t.id !== id))
  }

  return (
    <ToastContext.Provider value={{ toasts, showToast }}>
      {children}
      <div className='toast-container'>
        {toasts.map((t: any) => (
          <div
            key={t.id}
            className={`toast toast-${t.type}`}
            onClick={() => dismiss(t.id)}
          >
            {t.message}
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  )
}

// eslint-disable-next-line react-refresh/only-export-components
export const useToast = () => {
  const ctx = useContext(ToastContext)
  if (!ctx) throw new Error('useToast must be used within ToastProvider')
  return ctx
}
