import { Layout } from '../components/Layout'
import { useToast } from '../context/ToastContext'
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
  const [operations, setOperations] = useState<OperationData[] | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchOperations = async () => {
      setLoading(true)
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
            <OperationCard key={o.credential_id} operation={o} />
          ))}
      </div>
    )
  }

  return (
    <Layout>
      <h2>Account</h2>
      <div>Hello from Account</div>
      <form onSubmit={handlePassUpdateSubmit(onPassUpdateSubmit)}>
        <div>
          <label htmlFor='user_password'>Current password </label>
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
        <div>
          <label htmlFor='new_password'>New password </label>
          <input
            id='new_password'
            type='password'
            {...registerPassUpdate('new_password')}
            autoFocus
          />
          {passUpdateErrors.new_password && (
            <span className='field-error'>
              {passUpdateErrors.new_password.message}
            </span>
          )}
        </div>
        <div>
          <label htmlFor='confirm_new_password'>Confirm new </label>
          <input
            id='confirm_new_password'
            type='password'
            {...registerPassUpdate('confirm_new_password')}
            autoFocus
          />
          {passUpdateErrors.confirm_new_password && (
            <span className='field-error'>
              {passUpdateErrors.confirm_new_password.message}
            </span>
          )}
        </div>
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
      <div>
        {loading ? <div>loading...</div> : <div>{operationList()}</div>}
      </div>
    </Layout>
  )
}
