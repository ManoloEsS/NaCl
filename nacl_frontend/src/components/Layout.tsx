import type { ReactNode } from 'react'
import { useAuth } from '../context/AuthContext'
import { NavLink } from 'react-router-dom'

interface Props {
  children: ReactNode
}

export const Layout = ({ children }: Props) => {
  const { user } = useAuth()

  return (
    <div className='layout'>
      <nav className='navbar'>
        <div className='nav-brand'>
          <NavLink to={'/dash'}>NaCl</NavLink>
        </div>
        <div className='nav-links'>
          <NavLink to='/vault'>Vault</NavLink>
          <NavLink to='/new'>New Credential</NavLink>
          <NavLink to='/account'>Account</NavLink>
        </div>
        <div className='nav-user'>
          <span>{user?.username || user?.id}</span>
        </div>
      </nav>
      <main className='main-content'>{children}</main>
    </div>
  )
}
