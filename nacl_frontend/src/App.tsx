import { useState, useEffect } from 'react'
import axios from 'axios'
import './App.css'

function App() {
  const [newUsername, setNewUsername] = useState('')
  const [newPassword, setNewPassword] = useState('')
  // const [loading, setLoading] = useState(true)

  useEffect(() => {}, [])

  const handleRegisterUser = async (e: any) => {
    e.preventDefault()
    if (!newUsername.trim()) {
      return
    }

    try {
      await axios.post('/api/users', {
        username: newUsername,
        password: newPassword,
      })
      setNewUsername('')
      setNewPassword('')
    } catch (err) {
      console.error('Failed to add item:', err)
    }
  }

  return (
    <div className="container">
      <h1>NaCl</h1>

      <form onSubmit={handleRegisterUser} className="add-form">
        <input
          type="text"
          placeholder="Username"
          value={newUsername}
          onChange={(e) => setNewUsername(e.target.value)}
        />
        <input
          type="text"
          placeholder="Password"
          value={newPassword}
          onChange={(e) =>
            setNewPassword(() => {
              return '*'.repeat(e.target.value.length)
            })
          }
        />
        <button type="submit">Register</button>
      </form>
    </div>
  )
}

export default App
