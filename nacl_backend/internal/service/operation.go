package service

import (
	"context"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type OperationType string

var (
	TypeCreate  OperationType = "create"
	TypeUpdate  OperationType = "update"
	TypeDelete  OperationType = "delete"
	TypeDecrypt OperationType = "decrypt"
	TypeLogin   OperationType = "login"
)

func (ot OperationType) String() string {
	return string(ot)
}

func (svc *Service) SaveOperation(ctx context.Context, opType OperationType, service string, userID, credentialID uuid.UUID) error {
	operation := db.CreateOperationParams{
		UserID:  userID,
		OpType:  opType.String(),
		Service: service,
		CredentialID: pgtype.UUID{
			Bytes: [16]byte(credentialID),
			Valid: credentialID != uuid.Nil,
		},
	}

	err := svc.Queries.CreateOperation(ctx, operation)
	if err != nil {
		return err
	}

	return nil
}

func (svc *Service) ListOpsforUserID(ctx context.Context, userID uuid.UUID) ([]dto.OperationDataResponse, error) {
	ops, err := svc.Queries.GetOperationsForUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	parsedOps := make([]dto.OperationDataResponse, len(ops))
	for i, op := range ops {
		parsedSvcID, err := uuid.FromBytes(op.CredentialID.Bytes[:])
		if err != nil {
			parsedSvcID = uuid.Nil
		}
		operation := dto.OperationDataResponse{
			ID:           op.ID,
			OpType:       op.OpType,
			Service:      op.Service,
			CredentialID: parsedSvcID,
			CreatedAt:    op.CreatedAt.Time,
		}
		parsedOps[i] = operation
	}

	return parsedOps, nil
}
