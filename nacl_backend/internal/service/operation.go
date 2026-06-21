package service

import (
	"context"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (svc *Service) SaveOperation(ctx context.Context, opType, service string, userID, serviceID uuid.UUID) error {
	operation := db.CreateOperationParams{
		UserID:  userID,
		OpType:  opType,
		Service: service,
		ServiceID: pgtype.UUID{
			Bytes: [16]byte(serviceID),
			Valid: serviceID != uuid.Nil,
		},
	}

	err := svc.Db.Queries().CreateOperation(ctx, operation)
	if err != nil {
		return err
	}

	return nil
}

func (svc *Service) ListOpsforUserID(ctx context.Context, userID uuid.UUID) ([]dto.OperationDataResponse, error) {
	ops, err := svc.Db.Queries().GetOperationsForUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	parsedOps := make([]dto.OperationDataResponse, len(ops))
	for i, op := range ops {
		parsedSvcID, err := uuid.FromBytes(op.ServiceID.Bytes[:])
		if err != nil {
			parsedSvcID = uuid.Nil
		}
		operation := dto.OperationDataResponse{
			ID:        op.ID,
			OpType:    op.OpType,
			Service:   op.Service,
			ServiceID: parsedSvcID,
			CreatedAt: op.CreatedAt.Time,
		}
		parsedOps[i] = operation
	}

	return parsedOps, nil
}
