import { useForm } from 'react-hook-form'
import { Link, useNavigate } from 'react-router-dom'
import { type LoginRequest, LoginSchema } from '../lib/requestValidation'
import { zodResolver } from '@hookform/resolvers/zod'
import { useAuth } from '../context/AuthContext'
import { LoginHero } from '../components/LoginHero'
import { useEffect } from 'react'

export const LoginPage = () => {
  const { user, login } = useAuth()
  const navigate = useNavigate()
  const {
    register,
    handleSubmit,
    setError,
    setFocus,
    reset,
    formState: { errors, isSubmitting }
  } = useForm<LoginRequest>({
    resolver: zodResolver(LoginSchema)
  })

  useEffect(() => {
    if (user) navigate('/dash', { replace: true })
  }, [user, navigate])

  const onSubmit = async (data: LoginRequest) => {
    try {
      const userData = await login(data)
      navigate(userData ? '/dash' : '/login', { replace: true })
    } catch {
      setFocus('username')
      reset()
      setError('root', { message: 'Invalid email or password' })
    }
  }

  return (
    <div className='login-page'>
      <LoginHero />
      <div className='login-card'>
        <h1>NaCl</h1>
        <p className='login-subtitle'>Password Manager</p>
        <form onSubmit={handleSubmit(onSubmit)}>
          <div className='form-group'>
            <label htmlFor='username'>Username</label>
            <input
              id='username'
              type='text'
              {...register('username')}
              autoFocus
              autoComplete='username'
            />
            {errors.username && (
              <span className='field-error'>{errors.username.message}</span>
            )}
          </div>
          <div className='form-group'>
            <label htmlFor='password'>Password</label>
            <input
              id='password'
              type='password'
              {...register('user_password')}
              autoComplete='current-password'
            />
            {errors.user_password && (
              <span className='field-error'>
                {errors.user_password.message}
              </span>
            )}
          </div>
          {errors.root && <p className='error'>{errors.root.message}</p>}
          <div className='form-group'>
            <button
              type='submit'
              className='btn-primary btn-full'
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Logging in...' : 'Login'}
            </button>
          </div>
          <div className='form-group'>
            <Link className='register-link' to='/register'>
              Register new user
            </Link>
          </div>
        </form>
      </div>
    </div>
  )
}
