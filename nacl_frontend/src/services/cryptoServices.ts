import z from 'zod'
import { client } from '../api/client'
import {
  CreateServiceSchema,
  DecryptServiceSchema,
  UpdateServiceSchema
} from '../lib/requestValidation'
import {
  ServiceCredentialsSchema,
  ServiceMetadataArraySchema,
  ServiceMetadataSchema,
  type ServiceCredentials,
  type ServiceMetadata
} from '../lib/responseValidation'

const ServiceID = z.uuid()

export const createService = async (newService: {
  service: string
  service_username: string
  service_password: string
  description: string
  encryption_algorithm: string
  user_password: string
}): Promise<ServiceMetadata> => {
  const validated = CreateServiceSchema.parse(newService)
  const req = await client.post('/services', validated)
  return ServiceMetadataSchema.parse(req.data)
}

export const listServices = async (): Promise<ServiceMetadata[]> => {
  const req = await client.get('/services')
  const services = ServiceMetadataArraySchema.parse(req.data)
  return services
}

export const decryptService = async (decryptRequest: {
  userPassword: string
  serviceID: string
}): Promise<ServiceCredentials> => {
  const { serviceID, ...decryptInput } = decryptRequest
  const validated = DecryptServiceSchema.parse(decryptInput)
  const validatedID = ServiceID.parse(serviceID)
  const req = await client.post(
    `/services/${validatedID}/credentials`,
    validated
  )
  return ServiceCredentialsSchema.parse(req.data)
}

export const updateService = async (updateRequest: {
  servicePassword: string
  encryptionAlgorithm: string
  userPassword: string
  serviceID: string
}): Promise<ServiceMetadata> => {
  const { serviceID, ...updateInput } = updateRequest
  const validated = UpdateServiceSchema.parse(updateInput)
  const validatedID = ServiceID.parse(serviceID)
  const req = await client.patch(`/services/${validatedID}`, validated)
  return ServiceMetadataSchema.parse(req.data)
}
