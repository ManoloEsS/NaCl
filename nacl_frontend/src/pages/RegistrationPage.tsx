import { useForm } from 'react-hook-form'
import { Link, useNavigate } from 'react-router-dom'
import {
  type CreateUserRequest,
  CreateUserSchema
} from '../lib/requestValidation'
import { zodResolver } from '@hookform/resolvers/zod'
import { LoginHero } from '../components/LoginHero'
import { useToast } from '../context/ToastContext'

export const RegistrationPage = () => {
  const { showToast } = useToast()
  const navigate = useNavigate()
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting }
  } = useForm<CreateUserRequest>({
    resolver: zodResolver(CreateUserSchema)
  })

  const onSubmit = async (_data: CreateUserRequest) => {
    try {
      // await registerUser(data)
      showToast(
        'Registering new users is disabled in demo, login with Test_user account',
        'success'
      )
      setTimeout(() => navigate('/login', { replace: true }), 1500)
    } catch {
      setError('root', { message: 'Could not register user' })
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
              autoComplete='new-password'
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
              {isSubmitting ? 'Registering...' : 'Register'}
            </button>
          </div>
          <div className='form-group'>
            <Link className='register-link' to='/login'>
              Already registered? Log in here
            </Link>
          </div>
        </form>
      </div>
    </div>
  )
}
