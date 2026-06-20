import z from 'zod'
import { client } from '../api/client'
import {
  CreateServiceSchema,
  DecryptServiceSchema,
  UpdateServiceSchema,
  type UpdateServiceRequest,
  type NewServiceFormRequest,
  type DecryptRequest
} from '../lib/requestValidation'
import {
  ServiceCredentialsSchema,
  ServiceMetadataArraySchema,
  ServiceMetadataSchema,
  type ServiceCredentials,
  type ServiceMetadata
} from '../lib/responseValidation'

const ServiceID = z.uuid()

export const createService = async (
  newService: NewServiceFormRequest
): Promise<ServiceMetadata> => {
  const { confirm_service_password, ...serviceData } = newService
  const validated = CreateServiceSchema.parse(serviceData)
  const req = await client.post('/services', validated)
  return ServiceMetadataSchema.parse(req.data)
}

export const listServices = async (): Promise<ServiceMetadata[]> => {
  const req = await client.get('/services')
  const services = ServiceMetadataArraySchema.parse(req.data)
  return services
}

export const decryptService = async (
  reqData: DecryptRequest
): Promise<ServiceCredentials> => {
  const { serviceID, user_password } = reqData
  const validated = DecryptServiceSchema.parse({ user_password })
  const validatedID = ServiceID.parse(serviceID)
  const req = await client.post(
    `/services/${validatedID}/credentials`,
    validated
  )
  return ServiceCredentialsSchema.parse(req.data)
}

export const updateService = async ({
  serviceID,
  ...updateInput
}: UpdateServiceRequest & { serviceID: string }): Promise<ServiceMetadata> => {
  const validated = UpdateServiceSchema.parse(updateInput)
  const validatedID = ServiceID.parse(serviceID)
  const req = await client.patch(`/services/${validatedID}`, validated)
  return ServiceMetadataSchema.parse(req.data)
}
