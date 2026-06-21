import type { OperationData } from '../lib/responseValidation'

interface OperationProps {
  operation: OperationData
}

export const OperationCard = ({ operation }: OperationProps) => {
  const serviceStyle = {
    paddingTop: 10,
    paddingLeft: 2,
    border: 'solid',
    borderWidth: 1,
    marginBottom: 5
  }

  return (
    <div>
      <div style={serviceStyle}>
        <div>Service: {operation.service}</div>
        <div>Service ID: {operation.id}</div>
        <div>Operation type: {operation.op_type}</div>
        <div>Time: {operation.created_at.toString()}</div>
      </div>
    </div>
  )
}
