import z from 'zod'
import { client } from '../api/client'
import {
  CreateCredentialSchema,
  DecryptCredentialSchema,
  UpdateCredentialSchema,
  type UpdateCredentialRequest,
  type NewCredentialFormRequest,
  type DecryptRequest
} from '../lib/requestValidation'
import {
  DecryptedCredentialSchema,
  CredentialMetadataArraySchema,
  CredentialMetadataSchema,
  type DecryptedCredentials,
  type CredentialMetadata
} from '../lib/responseValidation'

const CredentialID = z.uuid()

export const createCredential = async (
  newCredential: NewCredentialFormRequest
): Promise<CredentialMetadata> => {
  const {
    confirm_service_password: _confirm_service_password,
    ...credentialData
  } = newCredential
  const validated = CreateCredentialSchema.parse(credentialData)
  const req = await client.post('/credentials', validated)
  return CredentialMetadataSchema.parse(req.data)
}

export const listCredentials = async (): Promise<CredentialMetadata[]> => {
  const req = await client.get('/credentials')
  const credentials = CredentialMetadataArraySchema.parse(req.data)
  return credentials
}

export const decryptCredential = async (
  reqData: DecryptRequest
): Promise<DecryptedCredentials> => {
  const { credentialID, user_password } = reqData
  const validated = DecryptCredentialSchema.parse({ user_password })
  const validatedID = CredentialID.parse(credentialID)
  const req = await client.post(
    `/credentials/${validatedID}/decrypt`,
    validated
  )
  return DecryptedCredentialSchema.parse(req.data)
}

export const updateCredential = async ({
  credentialID,
  ...updateInput
}: UpdateCredentialRequest & { credentialID: string }): Promise<CredentialMetadata> => {
  const validated = UpdateCredentialSchema.parse(updateInput)
  const validatedID = CredentialID.parse(credentialID)
  const req = await client.patch(`/credentials/${validatedID}`, validated)
  return CredentialMetadataSchema.parse(req.data)
}
