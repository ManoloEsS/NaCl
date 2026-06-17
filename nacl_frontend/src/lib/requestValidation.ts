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

export const NewServiceFormSchema = CreateServiceSchema.extend({
  confirm_service_password: z.string().min(1, 'Please confirm password')
}).refine((data) => data.service_password === data.confirm_service_password, {
  message: 'Passwords do not match',
  path: ['confirm_service_password']
})

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
export type NewServiceFormRequest = z.infer<typeof NewServiceFormSchema>
