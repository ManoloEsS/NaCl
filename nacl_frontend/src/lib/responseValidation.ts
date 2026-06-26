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

export const CredentialMetadataSchema = z
  .object({
    id,
    service,
    description,
    encryption_algorithm
  })
  .strict()

export const CredentialMetadataArraySchema = z.array(CredentialMetadataSchema)

export const DecryptedCredentialSchema = z
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

export const OperationDataSchema = z.object({
  id,
  op_type: z.string(),
  service,
  credential_id: id,
  created_at: z.string().pipe(z.coerce.date())
})

export const OperationDataArraySchema = z.array(OperationDataSchema)

export const ErrorSchema = z
  .object({
    error: z.string()
  })
  .strict()

export type UserData = z.infer<typeof LoginResponseSchema>
export type CredentialMetadata = z.infer<typeof CredentialMetadataSchema>
export type DecryptedCredentials = z.infer<typeof DecryptedCredentialSchema>
export type CredentialArrayMetadata = z.infer<typeof CredentialMetadataArraySchema>
export type OperationDataArray = z.infer<typeof OperationDataArraySchema>
export type OperationData = z.infer<typeof OperationDataSchema>
