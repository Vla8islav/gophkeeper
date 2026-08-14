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

func TestGophkeeperService_CreateSecret_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockGophkeeperRepository(ctrl)
	svc := gophkeeperService{repository: repository}

	const userID = int64(42)
	req := domain.CreateSecretRequest{
		ID:      uuid.New(),
		Type:    domain.SecretTypeLoginPassword,
		Payload: []byte("cipher"),
		Meta:    []byte("meta"),
	}

	repository.EXPECT().
		CreateSecret(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, params domain.CreateSecretParams) error {
			// UserID is taken from the argument, NOT from the client request. And the rest maps through.
			require.Equal(t, userID, params.UserID)
			require.Equal(t, req.ID, params.ID)
			require.Equal(t, req.Type, params.Type)
			require.Equal(t, req.Payload, params.Payload)
			require.Equal(t, req.Meta, params.Meta)
			return nil
		})

	err := svc.CreateSecret(context.Background(), userID, req)
	require.NoError(t, err)
}

func TestGophkeeperService_CreateSecret_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		req  domain.CreateSecretRequest
	}{
		{
			name: "missing id",
			req: domain.CreateSecretRequest{
				ID:      uuid.Nil,
				Type:    domain.SecretTypeText,
				Payload: []byte("x"),
			},
		},
		{
			name: "invalid type",
			req: domain.CreateSecretRequest{
				ID:      uuid.New(),
				Type:    domain.SecretType("bullshit"),
				Payload: []byte("x"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repository := mocks.NewMockGophkeeperRepository(ctrl)
			svc := gophkeeperService{repository: repository}

			// No EXPECT on CreateSecret - gomock fails the test if the repo is called
			err := svc.CreateSecret(context.Background(), 42, tt.req)
			require.ErrorIs(t, err, domain.ErrInvalidSecretType)
		})
	}
}

func TestGophkeeperService_CreateSecret_RepositoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	repository := mocks.NewMockGophkeeperRepository(ctrl)
	svc := gophkeeperService{repository: repository}

	req := domain.CreateSecretRequest{
		ID:      uuid.New(),
		Type:    domain.SecretTypeCard,
		Payload: []byte("cipher"),
	}

	repository.EXPECT().
		CreateSecret(gomock.Any(), gomock.Any()).
		Return(domain.ErrSecretAlreadyExists)

	err := svc.CreateSecret(context.Background(), 42, req)
	// The service wraps with %w, so ErrorIs still matches through the wrap.
	require.ErrorIs(t, err, domain.ErrSecretAlreadyExists)
}
