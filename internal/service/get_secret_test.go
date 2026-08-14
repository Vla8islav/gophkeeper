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

func TestGophkeeperService_GetSecret_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockGophkeeperRepository(ctrl)
	svc := gophkeeperService{repository: repository}

	const userID = int64(42)
	id := uuid.New()
	want := &domain.Secret{
		ID:      id,
		UserID:  userID,
		Type:    domain.SecretTypeText,
		Payload: []byte("cipher"),
		Version: 1,
	}

	// Pinning userID and id in the matcher proves the service forwards both
	// unchanged (userID from the token, id from the URL) — no swapping or mutation.
	repository.EXPECT().
		GetSecret(gomock.Any(), userID, id).
		Return(want, nil)

	got, err := svc.GetSecret(context.Background(), userID, id)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGophkeeperService_GetSecret_NilID(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockGophkeeperRepository(ctrl)
	svc := gophkeeperService{repository: repository}

	// No EXPECT → the repo must never be called when the id is invalid.
	got, err := svc.GetSecret(context.Background(), 42, uuid.Nil)
	require.ErrorIs(t, err, domain.ErrInvalidSecretID)
	require.Nil(t, got)
}

func TestGophkeeperService_GetSecret_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockGophkeeperRepository(ctrl)
	svc := gophkeeperService{repository: repository}

	id := uuid.New()
	repository.EXPECT().
		GetSecret(gomock.Any(), int64(42), id).
		Return(nil, domain.ErrSecretNotFound)

	got, err := svc.GetSecret(context.Background(), 42, id)
	require.ErrorIs(t, err, domain.ErrSecretNotFound) // still matches through the %w wrap
	require.Nil(t, got)
}
