import { Layout } from '../componets/Layout'
import type { ServiceMetadata } from '../lib/responseValidation'
import { listServices } from '../services/cryptoServices'
import { useEffect, useState } from 'react'
import { ServiceCard } from '../componets/ServiceCard'

export const Vault = () => {
  const [services, setServices] = useState<ServiceMetadata[] | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchServices = async () => {
      setLoading(true)
      try {
        const services = await listServices()
        setServices(services)
      } finally {
        setLoading(false)
      }
    }
    fetchServices()
  }, [])

  const serviceList = () => {
    return (
      <div>
        {services!.map((s) => (
          <ServiceCard key={s.id} service={s} />
        ))}
      </div>
    )
  }

  return (
    <Layout>
      {loading || !services ? <div>loading</div> : <div>{serviceList()}</div>}
    </Layout>
  )
}
