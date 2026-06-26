import { useState, type ChangeEvent } from 'react'
import type {
  DecryptedCredentials,
  CredentialMetadata
} from '../lib/responseValidation'
import { decryptCredential } from '../services/cryptoServices'
import { useToast } from '../context/ToastContext'

interface CredentialProps {
  credential: CredentialMetadata
}

export const CredentialCard = ({ credential }: CredentialProps) => {
  const { showToast } = useToast()
  const credentialStyle = {
    paddingTop: 10,
    paddingLeft: 2,
    border: 'solid',
    borderWidth: 1,
    marginBottom: 5
  }
  const [show, setShow] = useState(false)
  const [decrypted, setDecrypted] = useState<DecryptedCredentials | null>(null)
  const [userPass, setUserPass] = useState('')

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

  return (
    <div>
      <div style={credentialStyle}>
        <div>Service: {credential.service}</div>
        <div>Encryption Algorithm: {credential.encryption_algorithm}</div>
        <div>Description: {credential.description ? credential.description : ''}</div>
        {show ? (
          <div>
            <div>Username: {decrypted?.service_username}</div>
            <div>Password: {decrypted?.service_password}</div>
            <div>Created: {decrypted?.created_at.toString()}</div>
            <div>Updated: {decrypted?.updated_at.toString()}</div>

            <button onClick={handleHideDecrypt}>Hide credentials</button>
          </div>
        ) : (
          <div>
            <input type='password' value={userPass} onChange={handleChange} />
            <button onClick={handleDecrypt}>Show credentials</button>
          </div>
        )}
      </div>
    </div>
  )
}
