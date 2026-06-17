import axios from 'axios'

// TODO: Add token refresh and retry pattern
// NOTE: Maybe add response unwrapping

let showToast: ((message: string) => void) | null = null

export function setToastHandler(handler: (message: string) => void) {
  showToast = handler
}

export const client = axios.create({ baseURL: '/api' })

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

client.interceptors.response.use(
  (response) => response,
  (error) => {
    const msg = error.response?.data?.error || error.message
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
      return Promise.reject(error)
    }
    showToast?.(msg)
    return Promise.reject(error)
  }
)
