import type { ChangeEvent, SubmitEvent, MouseEvent } from 'react'

export interface RegisterFormProps {
  handleRegisterUser: (_event: SubmitEvent<HTMLFormElement>) => void
  handleLoginUser: (_event: MouseEvent<HTMLButtonElement>) => void
  handleUsernameChange: (_event: ChangeEvent<HTMLInputElement>) => void
  handlePasswordChange: (_event: ChangeEvent<HTMLInputElement>) => void
  username: string
  password: string
}

export interface NotificationObj {
  message: string
  isError?: boolean
}

export interface NotificationProps {
  notification: NotificationObj | null
}
