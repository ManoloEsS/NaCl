import { z } from 'zod'

export const username = z.string().min(1, 'Username is required')
export const service_username = z.string().min(1, 'Service username needed')
export const service_password = z.string().min(1, 'Service password required')
export const user_password = z.string().min(1, 'Password is required')
export const encryption_algorithm = z.enum(['aes-gcm'])
export const description = z.string().optional()
export const id = z.uuid()
export const service = z.string()
