import { useForm } from 'react-hook-form'
import { Layout } from '../components/Layout'
import {
  NewCredentialFormSchema,
  type NewCredentialFormRequest
} from '../lib/requestValidation'
import { zodResolver } from '@hookform/resolvers/zod'
import { createCredential } from '../services/cryptoServices'
import { useToast } from '../context/ToastContext'

const fieldInfo: Record<string, { title: string; body: string }> = {
  service: {
    title: 'Service Name',
    body: 'The name of the service or website you are creating credentials for (e.g., GitHub, AWS, Gmail).'
  },
  service_username: {
    title: 'Service Username',
    body: 'Your username or email used to log in to this service.'
  },
  service_password: {
    title: 'Service Password',
    body: 'The password for this service account.'
  },
  confirm_service_password: {
    title: 'Confirm Password',
    body: 'Re-enter the password to confirm it was typed correctly.'
  },
  description: {
    title: 'Description',
    body: 'A short note to help identify these credentials (optional).'
  },
  encryption_algorithm: {
    title: 'Encryption Algorithm',
    body: 'The cipher used to encrypt your credentials before storing them. AES-GCM is the recommended standard: it provides authenticated encryption, meaning any tampering with the ciphertext is detected on decryption.'
  },
  user_password: {
    title: 'User Password',
    body: 'Your master NaCl password. This is used to encrypt and decrypt all your stored credentials.'
  }
}

export const NewCredential = () => {
  const { showToast } = useToast()
  const {
    reset,
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting }
  } = useForm<NewCredentialFormRequest>({
    resolver: zodResolver(NewCredentialFormSchema)
  })

  const onSubmit = async (data: NewCredentialFormRequest) => {
    try {
      await createCredential(data)
      showToast('Credentials encrypted and saved!', 'success')
      reset()
    } catch {
      setError('root', { message: 'Could not encrypt and save credentials' })
    }
  }

  return (
    <Layout>
      <div className='form-with-info'>
        <div className='card card-narrow'>
          <form onSubmit={handleSubmit(onSubmit)}>
            <div className='form-group'>
              <div className='form-label-row'>
                <label htmlFor='service'>Service</label>
              </div>
              <input
                id='service'
                type='text'
                {...register('service')}
                autoFocus
              />
              {errors.service && (
                <span className='field-error'>{errors.service.message}</span>
              )}
            </div>
            <div className='form-group'>
              <div className='form-label-row'>
                <label htmlFor='service_username'>Service Username</label>
                <span
                  className='field-info'
                  data-tooltip={fieldInfo.service_username.body}
                >
                  ?
                </span>
              </div>
              <input
                id='service_username'
                type='text'
                {...register('service_username')}
              />
              {errors.service_username && (
                <span className='field-error'>
                  {errors.service_username.message}
                </span>
              )}
            </div>
            <div className='form-group'>
              <div className='form-label-row'>
                <label htmlFor='service_password'>Service Password</label>
                <span
                  className='field-info'
                  data-tooltip={fieldInfo.service_password.body}
                >
                  ?
                </span>
              </div>
              <input
                id='service_password'
                type='password'
                {...register('service_password')}
              />
              {errors.service_password && (
                <span className='field-error'>
                  {errors.service_password.message}
                </span>
              )}
            </div>
            <div className='form-group'>
              <div className='form-label-row'>
                <label htmlFor='confirm_service_password'>
                  Confirm Service Password
                </label>
                <span
                  className='field-info'
                  data-tooltip={fieldInfo.confirm_service_password.body}
                >
                  ?
                </span>
              </div>
              <input
                id='confirm_service_password'
                type='password'
                {...register('confirm_service_password')}
              />
              {errors.confirm_service_password && (
                <span className='field-error'>
                  {errors.confirm_service_password.message}
                </span>
              )}
            </div>
            <div className='form-group'>
              <div className='form-label-row'>
                <label htmlFor='description'>Description (optional)</label>
                <span
                  className='field-info'
                  data-tooltip={fieldInfo.description.body}
                >
                  ?
                </span>
              </div>
              <textarea
                id='description'
                rows={3}
                {...register('description')}
              />
              {errors.description && (
                <span className='field-error'>
                  {errors.description.message}
                </span>
              )}
            </div>
            <div className='form-group'>
              <div className='form-label-row'>
                <label htmlFor='encryption_algorithm'>
                  Encryption Algorithm
                </label>
                <span
                  className='field-info'
                  data-tooltip={fieldInfo.encryption_algorithm.body}
                >
                  ?
                </span>
              </div>
              <select
                id='encryption_algorithm'
                {...register('encryption_algorithm')}
              >
                <option value='aes-gcm'>AES-GCM</option>
              </select>
            </div>
            <div className='form-group'>
              <div className='form-label-row'>
                <label htmlFor='user_password'>User Password</label>
                <span
                  className='field-info'
                  data-tooltip={fieldInfo.user_password.body}
                >
                  ?
                </span>
              </div>
              <input
                id='user_password'
                type='password'
                {...register('user_password')}
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
                {isSubmitting ? 'Processing' : 'Encrypt and Save'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Layout>
  )
}
