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

export const CreateServiceSchema = z
  .object({
    service: z.string().min(1, 'Service required'),
    service_username,
    service_password,
    description,
    encryption_algorithm,
    user_password
  })
  .strict()

export const DecryptServiceSchema = z.object({
  user_password
})

export const UpdateServiceSchema = z.object({
  service_password,
  encryption_algorithm,
  user_password
})

export type CreateUserRequest = z.infer<typeof CreateUserSchema>
export type LoginRequest = CreateUserRequest
export type CreateServiceRequest = z.infer<typeof CreateServiceSchema>
export type DecryptServiceRequest = z.infer<typeof DecryptServiceSchema>
export type UpdateServiceRequest = z.infer<typeof UpdateServiceSchema>
