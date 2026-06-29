import { Layout } from '../components/Layout'
import type { CredentialMetadata } from '../lib/responseValidation'
import { listCredentials } from '../services/cryptoServices'
import { useEffect, useState } from 'react'
import { CredentialCard } from '../components/CredentialCard'

export const Vault = () => {
  const [credentials, setCredentials] = useState<CredentialMetadata[] | null>(
    null
  )
  const [loading, setLoading] = useState(true)

  const fetchCredentials = async () => {
    try {
      const credentials = await listCredentials()
      setCredentials(credentials)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchCredentials()
  }, [])

  const credentialList = () => {
    return (
      <div>
        {credentials!.map((c) => (
          <CredentialCard
            key={c.id}
            credential={c}
            onDelete={fetchCredentials}
          />
        ))}
      </div>
    )
  }

  return (
    <Layout>
      <p className='credential-card'>
        <text>
          Use your master NaCl password to decrypt credentials.
          <br />
          For this demo you can use password: <strong>hashandeggs</strong>
        </text>
      </p>
      {loading || !credentials ? (
        <div>loading</div>
      ) : (
        <div>{credentialList()}</div>
      )}
    </Layout>
  )
}
