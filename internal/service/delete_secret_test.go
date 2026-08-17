package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/mocks"
)

func TestGophkeeperService_DeleteSecret_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockGophkeeperRepository(ctrl)
	svc := gophkeeperService{repository: repository}

	const userID = int64(42)
	id := uuid.New()
	repository.EXPECT().DeleteSecret(gomock.Any(), userID, id).Return(nil)

	require.NoError(t, svc.DeleteSecret(context.Background(), userID, id))
}

func TestGophkeeperService_DeleteSecret_NilID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockGophkeeperRepository(ctrl)
	svc := gophkeeperService{repository: repository}

	// repo shouldn't be called when the id is invalid
	err := svc.DeleteSecret(context.Background(), 42, uuid.Nil)
	require.ErrorIs(t, err, domain.ErrInvalidSecretID)
}

func TestGophkeeperService_DeleteSecret_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockGophkeeperRepository(ctrl)
	svc := gophkeeperService{repository: repository}
	id := uuid.New()
	repository.EXPECT().DeleteSecret(gomock.Any(), int64(42), id).Return(domain.ErrSecretNotFound)

	require.ErrorIs(t, svc.DeleteSecret(context.Background(), 42, id), domain.ErrSecretNotFound)
}
