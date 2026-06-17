import { z } from 'zod'
import {
  username,
  service_username,
  encryption_algorithm,
  description,
  id,
  service,
  service_password
} from './primitives'

export const LoginResponseSchema = z
  .object({
    username,
    id,
    token: z.string()
  })
  .strict()

export const ServiceMetadataSchema = z
  .object({
    id,
    service,
    description,
    encryption_algorithm
  })
  .strict()

export const ServiceMetadataArraySchema = z.array(ServiceMetadataSchema)

export const ServiceCredentialsSchema = z
  .object({
    service,
    service_username,
    description,
    service_password,
    encryption_algorithm,
    created_at: z.string().pipe(z.coerce.date()),
    updated_at: z.string().pipe(z.coerce.date())
  })
  .strict()

export const ErrorSchema = z
  .object({
    error: z.string()
  })
  .strict()

export type UserData = z.infer<typeof LoginResponseSchema>
export type ServiceMetadata = z.infer<typeof ServiceMetadataSchema>
export type ServiceCredentials = z.infer<typeof ServiceCredentialsSchema>
export type ServiceArrayMetadata = z.infer<typeof ServiceMetadataArraySchema>
