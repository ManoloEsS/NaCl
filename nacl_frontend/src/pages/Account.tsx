import { useAuth } from '../context/AuthContext'
import { Layout } from '../components/Layout'
import { useToast } from '../context/ToastContext'
import { useNavigate } from 'react-router-dom'
import {
  UpdatePasswordSchema,
  type UpdatePasswordRequest
} from '../lib/requestValidation'
import { useForm } from 'react-hook-form'
import { updatePassword } from '../services/authServices'
import { OperationData } from '../lib/responseValidation'
import { useEffect, useState } from 'react'
import { listOperations } from '../services/operationsServices'
import { OperationCard } from '../components/OperationCard'
import { zodResolver } from '@hookform/resolvers/zod'

export const Account = () => {
  const { user, logout } = useAuth()
  const navigate = useNavigate()
  const [operations, setOperations] = useState<OperationData[]>([])
  const [loading, setLoading] = useState(true)

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  useEffect(() => {
    const fetchOperations = async () => {
      try {
        const operations = await listOperations()
        setOperations(operations)
      } finally {
        setLoading(false)
      }
    }
    fetchOperations()
  }, [])

  const { showToast } = useToast()
  const {
    reset: resetPassUpdate,
    register: registerPassUpdate,
    handleSubmit: handlePassUpdateSubmit,
    setError: setPassUpdateError,
    formState: {
      errors: passUpdateErrors,
      isSubmitting: isSubmittingPassUpdate
    }
  } = useForm<UpdatePasswordRequest>({
    resolver: zodResolver(UpdatePasswordSchema)
  })

  const onPassUpdateSubmit = async (data: UpdatePasswordRequest) => {
    try {
      if (window.confirm('This will change your password. Are you sure?')) {
        await updatePassword(data)
        showToast('Password updated', 'success')
        resetPassUpdate()
      }
    } catch (error: any) {
      setPassUpdateError('root', { message: 'Could not update user password' })
      resetPassUpdate()
      if (error.response?.status === 403) {
        showToast('Current password does not match', 'error')
      } else {
        showToast('Something went wrong', 'error')
      }
    }
  }

  const operationList = () => {
    return (
      <div>
        {operations!
          .sort((a, b) => b.created_at.getTime() - a.created_at.getTime())
          .map((o) => (
            <OperationCard key={o.id} operation={o} />
          ))}
      </div>
    )
  }

  return (
    <Layout>
      <div className='form-with-info'>
        <div className='card card-narrow scrollable-list'>
          <h3 className='card-title'>Operation History</h3>
          {loading ? (
            <div style={{ color: '#9e9ea0' }}>loading...</div>
          ) : (
            <div>{operationList()}</div>
          )}
        </div>
        <div className='info-panel'>
          <div className='account-info-card'>
            <h3 className='info-panel-title'>Account</h3>
            <p className='account-info-text'>
              Logged in as <strong>{user?.username || user?.id}</strong>
            </p>
            <button onClick={handleLogout} className='btn-primary btn-full'>
              Logout
            </button>
          </div>
          <h3 className='info-panel-title info-panel-title--spaced'>Change Password</h3>
          <form onSubmit={handlePassUpdateSubmit(onPassUpdateSubmit)}>
            <div className='form-group'>
              <label htmlFor='user_password'>Current password</label>
              <input
                id='user_password'
                type='password'
                {...registerPassUpdate('user_password')}
                autoFocus
              />
              {passUpdateErrors.user_password && (
                <span className='field-error'>
                  {passUpdateErrors.user_password.message}
                </span>
              )}
            </div>
            <div className='form-group'>
              <label htmlFor='new_password'>New password</label>
              <input
                id='new_password'
                type='password'
                {...registerPassUpdate('new_password')}
              />
              {passUpdateErrors.new_password && (
                <span className='field-error'>
                  {passUpdateErrors.new_password.message}
                </span>
              )}
            </div>
            <div className='form-group'>
              <label htmlFor='confirm_new_password'>Confirm new password</label>
              <input
                id='confirm_new_password'
                type='password'
                {...registerPassUpdate('confirm_new_password')}
              />
              {passUpdateErrors.confirm_new_password && (
                <span className='field-error'>
                  {passUpdateErrors.confirm_new_password.message}
                </span>
              )}
            </div>
            {passUpdateErrors.root && (
              <p className='error'>{passUpdateErrors.root.message}</p>
            )}
            <div className='form-group'>
              <button
                type='submit'
                className='btn-primary btn-full'
                disabled={isSubmittingPassUpdate}
              >
                {isSubmittingPassUpdate ? 'Processing' : 'Update Password'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Layout>
  )
}
