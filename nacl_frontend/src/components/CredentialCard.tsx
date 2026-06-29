import { useRef, useState, type ChangeEvent } from 'react'
import type {
  DecryptedCredentials,
  CredentialMetadata
} from '../lib/responseValidation'
import { decryptCredential, deleteCredential } from '../services/cryptoServices'
import { useToast } from '../context/ToastContext'

interface CredentialProps {
  credential: CredentialMetadata
  onDelete?: () => void
}

export const CredentialCard = ({ credential, onDelete }: CredentialProps) => {
  const { showToast } = useToast()
  const [show, setShow] = useState(false)
  const [decrypted, setDecrypted] = useState<DecryptedCredentials | null>(null)
  const [userPass, setUserPass] = useState('')

  const inputRef = useRef<HTMLInputElement>(null)
  const handleDecrypt = async () => {
    try {
      const decryptedSvc = await decryptCredential({
        user_password: userPass,
        credentialID: credential.id
      })
      setDecrypted(decryptedSvc)
      setShow(true)
    } catch (e: any) {
      if (e.response?.status === 403) {
        showToast('Password did not match', 'error')
      } else {
        showToast('Could not decrypt credentials', 'error')
      }
      setUserPass('')
      setDecrypted(null)
      setShow(false)
      inputRef.current?.focus()
    }
  }
  const handleHideDecrypt = () => {
    console.log('hidden')
    setShow(false)
    setDecrypted(null)
    setUserPass('')
  }

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => {
    setUserPass(event.target.value)
  }

  const copyToClipboard = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text)
      showToast(`${label} copied`, 'success')
    } catch {
      showToast('Failed to copy', 'error')
    }
  }

  const handleDelete = async () => {
    console.log('deleted')
    try {
      if (
        window.confirm(
          `This will delete the data for ${credential.service} permanently. Are you sure?`
        )
      ) {
        if (
          credential.description !== 'super secret password for secret account'
        ) {
          await deleteCredential({
            credentialID: credential.id,
            user_password: userPass
          })
          showToast(`${credential.service} deleted`, 'success')
        } else {
          showToast(
            'Deleting demo credential is disabled, try creating one yourself'
          )
        }
      }
    } catch (e) {
      onDelete?.()
      console.log(e)
      showToast('Failed to delete', 'error')
    }
  }

  return (
    <div>
      <div className='credential-card'>
        <div className='credential-service'>
          <strong>Service:</strong> {credential.service}
        </div>
        <div>
          <strong>Description:</strong>{' '}
          {credential.description ? credential.description : ''}
        </div>
        {show ? (
          <div>
            <div className='credential-card-line credential-copy-row'>
              <span>
                <strong>Username:</strong> {decrypted?.service_username}
              </span>
              <button
                className='btn-copy'
                onClick={() =>
                  copyToClipboard(decrypted?.service_username ?? '', 'Username')
                }
                title='Copy username'
              >
                Copy
              </button>
            </div>
            <div className='credential-card-line credential-copy-row'>
              <span>
                <strong>Password:</strong> {decrypted?.service_password}
              </span>
              <button
                className='btn-copy'
                onClick={() =>
                  copyToClipboard(decrypted?.service_password ?? '', 'Password')
                }
                title='Copy password'
              >
                Copy
              </button>
            </div>
            <div className='credential-card-line'>
              <strong>Encryption Algorithm:</strong>{' '}
              {credential.encryption_algorithm}
            </div>
            <div className='credential-card-line'>
              <strong>Created:</strong> {decrypted?.created_at.toString()}
            </div>
            <div className='credential-card-line'>
              <strong>Updated:</strong> {decrypted?.updated_at.toString()}
            </div>
            <div className='credential-btn-line'>
              <button onClick={handleHideDecrypt} className='btn-primary'>
                Hide credentials
              </button>
              <button onClick={handleDelete} className='btn-danger btn-primary'>
                Delete
              </button>
            </div>
          </div>
        ) : (
          <div className='credential-decrypt'>
            <input
              type='password'
              value={userPass}
              onChange={handleChange}
              placeholder='Enter your password'
              ref={inputRef}
            />
            <button onClick={handleDecrypt} className='btn-small btn-primary'>
              Show credentials
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
