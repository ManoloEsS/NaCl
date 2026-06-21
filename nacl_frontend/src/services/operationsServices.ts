import { client } from '../api/client'
import {
  OperationData,
  OperationDataArraySchema
} from '../lib/responseValidation'

export const listOperations = async (): Promise<OperationData[]> => {
  const req = await client.get('/operations')
  const operations = OperationDataArraySchema.parse(req.data)
  return operations
}
