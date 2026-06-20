import { useForm } from 'react-hook-form'
import { Layout } from '../componets/Layout'
import {
  NewServiceFormSchema,
  type NewServiceFormRequest
} from '../lib/requestValidation'
import { zodResolver } from '@hookform/resolvers/zod'
import { createService } from '../services/cryptoServices'
import { useToast } from '../context/ToastContext'

//TODO: styling and encryption information below form
export const NewService = () => {
  const { showToast } = useToast()
  const {
    reset,
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting }
  } = useForm<NewServiceFormRequest>({
    resolver: zodResolver(NewServiceFormSchema)
  })

  const onSubmit = async (data: NewServiceFormRequest) => {
    try {
      await createService(data)
      showToast('Credentials encrypted and saved!', 'success')
      reset()
    } catch {
      setError('root', { message: 'Could not encrypt and save credentials' })
    }
  }

  return (
    <Layout>
      <h2>New Service</h2>
      <div>Hello from NewService</div>
      <form onSubmit={handleSubmit(onSubmit)}>
        <div>
          <label htmlFor='service'>Service</label>
          <input id='service' type='text' {...register('service')} autoFocus />
          {errors.service && (
            <span className='field-error'>{errors.service.message}</span>
          )}
        </div>
        <div>
          <label htmlFor='service_username'>Service Username</label>
          <input
            id='service_username'
            type='text'
            {...register('service_username')}
            autoFocus
          />
          {errors.service_username && (
            <span className='field-error'>
              {errors.service_username.message}
            </span>
          )}
        </div>
        <div>
          <label htmlFor='service_password'>Service Password</label>
          <input
            id='service_password'
            type='password'
            {...register('service_password')}
            autoFocus
          />
          {errors.service_password && (
            <span className='field-error'>
              {errors.service_password.message}
            </span>
          )}
        </div>
        <div>
          <label htmlFor='confirm_service_password'>
            Confirm Service Password
          </label>
          <input
            id='confirm_service_password'
            type='password'
            {...register('confirm_service_password')}
            autoFocus
          />
          {errors.confirm_service_password && (
            <span className='field-error'>
              {errors.confirm_service_password.message}
            </span>
          )}
        </div>
        <div>
          <label htmlFor='description'>Description</label>
          <input id='description' type='text' {...register('description')} />
          {errors.description && (
            <span className='field-error'>{errors.description.message}</span>
          )}
        </div>
        <div>
          <label htmlFor='encryption_algorithm'>Encryption Algorithm</label>
          <select
            id='encryption_algorithm'
            {...register('encryption_algorithm')}
          >
            <option value='aes-gcm'>aes-gcm</option>
          </select>
        </div>
        <div>
          <label htmlFor='user_password'>User Password</label>
          <input
            id='user_password'
            type='password'
            {...register('user_password')}
            autoFocus
          />
          {errors.user_password && (
            <span className='field-error'>{errors.user_password.message}</span>
          )}
        </div>
        {errors.root && <p className='error'>{errors.root.message}</p>}
        <div className='form-group'>
          <button
            type='submit'
            className='btn-primary btn-full'
            disabled={isSubmitting}
          >
            {isSubmitting ? 'Processing' : 'Encrypt and Save'}
          </button>
        </div>
      </form>
    </Layout>
  )
}
