import type { OperationData } from '../lib/responseValidation'

interface OperationProps {
  operation: OperationData
}

export const OperationCard = ({ operation }: OperationProps) => {
  return (
    <div>
      <div className='operation-card'>
        <div><strong>Service:</strong> {operation.service}</div>
        <div><strong>Operation:</strong> {operation.op_type}</div>
        <div><strong>Time:</strong> {operation.created_at.toString()}</div>
      </div>
    </div>
  )
}
