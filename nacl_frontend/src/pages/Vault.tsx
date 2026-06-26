import { Layout } from '../components/Layout'
import type { CredentialMetadata } from '../lib/responseValidation'
import { listCredentials } from '../services/cryptoServices'
import { useEffect, useState } from 'react'
import { CredentialCard } from '../components/CredentialCard'

export const Vault = () => {
  const [credentials, setCredentials] = useState<CredentialMetadata[] | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchCredentials = async () => {
      setLoading(true)
      try {
        const credentials = await listCredentials()
        setCredentials(credentials)
      } finally {
        setLoading(false)
      }
    }
    fetchCredentials()
  }, [])

  const credentialList = () => {
    return (
      <div>
        {credentials!.map((c) => (
          <CredentialCard key={c.id} credential={c} />
        ))}
      </div>
    )
  }

  return (
    <Layout>
      {loading || !credentials ? <div>loading</div> : <div>{credentialList()}</div>}
    </Layout>
  )
}
