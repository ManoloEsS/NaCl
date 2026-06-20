import { LoginResponseSchema, type UserData } from '../lib/responseValidation'
import {
  CreateUserSchema,
  LoginSchema,
  UpdatePasswordSchema,
  type CreateUserRequest,
  type LoginRequest,
  type UpdatePasswordRequest
} from '../lib/requestValidation'
import { client } from '../api/client'

export const loginService = async (login: LoginRequest): Promise<UserData> => {
  const validated = LoginSchema.parse(login)
  const res = await client.post('/login', validated)
  const userData = LoginResponseSchema.parse(res.data)
  return userData
}

export const registerUser = async (
  userData: CreateUserRequest
): Promise<void> => {
  const validated = CreateUserSchema.parse(userData)
  await client.post('/users', validated)
}

export const updatePassword = async (
  newPassData: UpdatePasswordRequest
): Promise<void> => {
  const validated = UpdatePasswordSchema.parse(newPassData)
  const { confirm_new_password: _confirm_new_password, ...newData } = validated
  await client.patch('/users', newData)
}
