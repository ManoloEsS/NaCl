import { useState, useEffect } from 'react'
import axios from 'axios'
import './App.css'

interface Item {
  id: string;
  name: string;
  description?: string;
}

function App() {
  const [items, setItems] = useState<Item[]>([])
  const [newName, setNewName] = useState('')
  const [newDesc, setNewDesc] = useState('')
  // const [loading, setLoading] = useState(true)

  useEffect(() => {}, [])

  // const fetchItems = async () => {
  //   try {
  //     const res = await axios.get<Item[]>('/api/items')
  //     setItems(res.data)
  //   } catch (err) {
  //     console.error('Failed to fetch items:', err)
  //   } finally {
  //     setLoading(false)
  //   }
  // }

  const handleAddItem = async (e: any) => {
    e.preventDefault()
    if (!newName.trim()) return

    try {
      const res = await axios.post<Item>('/api/items', {
        name: newName,
        description: newDesc,
      })
      setItems([...items, res.data])
      setNewName('')
      setNewDesc('')
    } catch (err) {
      console.error('Failed to add item:', err)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await axios.delete(`/api/items/${id}`)
      setItems(items.filter((i) => i.id !== id))
    } catch (err) {
      console.error('Failed to delete item:', err)
    }
  }

  // if (loading) {
  //   return <div>Loading...</div>;
  // }

  return (
    <div className="container">
      <h1>PROJECT_NAME</h1>

      <form onSubmit={handleAddItem} className="add-form">
        <input
          type="text"
          placeholder="Name"
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
        />
        <input
          type="text"
          placeholder="Description (optional)"
          value={newDesc}
          onChange={(e) => setNewDesc(e.target.value)}
        />
        <button type="submit">Add Item</button>
      </form>

      <ul className="item-list">
        {items.map((item) => (
          <li key={item.id} className="item">
            <span>
              <strong>{item.name}</strong>
              {item.description && <span> - {item.description}</span>}
            </span>
            <button onClick={() => handleDelete(item.id)}>Delete</button>
          </li>
        ))}
      </ul>

      {items.length === 0 && <p>No items yet. Add one above!</p>}
    </div>
  )
}

export default App
