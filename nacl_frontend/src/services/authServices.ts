import { LoginResponseSchema, type UserData } from '../lib/responseValidation'
import { CreateUserSchema, LoginSchema } from '../lib/requestValidation'
import { client } from '../api/client'

export const loginService = async (
  username: string,
  user_password: string
): Promise<UserData> => {
  const validated = LoginSchema.parse({ username, user_password })
  const res = await client.post('/login', validated)
  const userData = LoginResponseSchema.parse(res.data)
  return userData
}

export const registerUser = async (
  username: string,
  user_password: string
): Promise<void> => {
  const validated = CreateUserSchema.parse({ username, user_password })
  await client.post('/users', validated)
}
