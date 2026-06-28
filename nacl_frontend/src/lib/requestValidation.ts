import { z } from 'zod'
import {
  username,
  service_password,
  service_username,
  encryption_algorithm,
  description,
  user_password
} from './primitives'

export const CreateUserSchema = z
  .object({
    username,
    user_password
  })
  .strict()

export const LoginSchema = CreateUserSchema

export const UpdatePasswordSchema = z
  .object({
    user_password,
    new_password: z.string().min(1, 'New password required'),
    confirm_new_password: z.string().min(1, 'Confirm new password')
  })
  .refine((data) => data.new_password === data.confirm_new_password, {
    message: 'Passwords do not match',
    path: ['confirm_new_password']
  })
  .strict()

export const CreateCredentialSchema = z
  .object({
    service: z.string().min(1, 'Service required'),
    service_username,
    service_password,
    description,
    encryption_algorithm,
    user_password
  })
  .strict()

export const NewCredentialFormSchema = CreateCredentialSchema.extend({
  confirm_service_password: z.string().min(1, 'Please confirm password')
})
  .refine((data) => data.service_password === data.confirm_service_password, {
    message: 'Passwords do not match',
    path: ['confirm_service_password']
  })
  .strict()

export const DecryptCredentialSchema = z
  .object({
    user_password
  })
  .strict()

export const DecryptRequestSchema = DecryptCredentialSchema.extend({
  credentialID: z.uuid()
}).strict()

export const DeleteCredentialSchema = z
  .object({
    user_password
  })
  .strict()

export const DeleteRequestSchema = DeleteCredentialSchema.extend({
  credentialID: z.uuid()
}).strict()

export const UpdateCredentialSchema = z
  .object({
    service_password,
    encryption_algorithm,
    user_password
  })
  .strict()

export type CreateUserRequest = z.infer<typeof CreateUserSchema>
export type LoginRequest = CreateUserRequest
export type CreateCredentialRequest = z.infer<typeof CreateCredentialSchema>
export type DecryptCredentialRequest = z.infer<typeof DecryptCredentialSchema>
export type UpdateCredentialRequest = z.infer<typeof UpdateCredentialSchema>
export type NewCredentialFormRequest = z.infer<typeof NewCredentialFormSchema>
export type UpdatePasswordRequest = z.infer<typeof UpdatePasswordSchema>
export type DecryptRequest = z.infer<typeof DecryptRequestSchema>
export type DeleteCredentialRequest = z.infer<typeof DeleteRequestSchema>
