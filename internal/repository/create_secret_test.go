package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/helpers"
)

// helper: create a user to own the secrets (FK requires a real user_id).
func createTestUser(t *testing.T, ctx context.Context, storage *PostgresStorage, tag string) int64 {
	t.Helper()
	userID, err := storage.CreateUser(ctx, domain.CreateUserParams{
		Login:        helpers.UniqueLogin(tag),
		PasswordHash: "hashed-password",
	})
	require.NoError(t, err)
	require.Greater(t, userID, int64(0))
	return userID
}

func TestPostgresStorage_CreateSecret(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)
	ctx := context.Background()

	userID := createTestUser(t, ctx, storage, "create-secret-test")

	params := domain.CreateSecretParams{
		ID:      uuid.New(),
		UserID:  userID,
		Type:    domain.SecretTypeLoginPassword,
		Payload: []byte{0x00, 0x01, 0x02, 0xff}, // binary incl. a null byte
		Meta:    []byte("some-meta"),
	}

	err := storage.CreateSecret(ctx, params)
	require.NoError(t, err)

	var (
		gotUserID  int64
		gotType    domain.SecretType
		gotPayload []byte
		gotMeta    []byte
		gotVersion int64
		gotDeleted bool
	)
	err = storage.db.QueryRowContext(ctx,
		`SELECT user_id, type, payload, meta, version, deleted
                 FROM secrets
                 WHERE id = $1`,
		params.ID,
	).Scan(&gotUserID, &gotType, &gotPayload, &gotMeta, &gotVersion, &gotDeleted)
	require.NoError(t, err)

	require.Equal(t, params.UserID, gotUserID)
	require.Equal(t, params.Type, gotType)
	require.Equal(t, params.Payload, gotPayload) // byte-for-byte, null byte survives
	require.Equal(t, params.Meta, gotMeta)
	require.Equal(t, int64(1), gotVersion) // DB default
	require.False(t, gotDeleted)           // DB default
}

func TestPostgresStorage_CreateSecret_DuplicateID(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)
	ctx := context.Background()

	userID := createTestUser(t, ctx, storage, "duplicate-secret-test")

	params := domain.CreateSecretParams{
		ID:      uuid.New(),
		UserID:  userID,
		Type:    domain.SecretTypeText,
		Payload: []byte("first"),
	}

	err := storage.CreateSecret(ctx, params)
	require.NoError(t, err)

	// Same UUID again → ON CONFLICT DO NOTHING → no row returned → domain error.
	err = storage.CreateSecret(ctx, params)
	require.ErrorIs(t, err, domain.ErrSecretAlreadyExists)
}

func TestPostgresStorage_CreateSecret_NilMetaStoresNull(t *testing.T) {
	logger := zap.NewNop()
	cfg, _ := config.ReadFlagsServer(nil, logger)
	storage := InitTestPostgresStorage(t, cfg)
	ctx := context.Background()

	userID := createTestUser(t, ctx, storage, "nil-meta-secret-test")

	params := domain.CreateSecretParams{
		ID:      uuid.New(),
		UserID:  userID,
		Type:    domain.SecretTypeBinary,
		Payload: []byte{0xde, 0xad, 0xbe, 0xef},
		Meta:    nil, // no metadata
	}

	err := storage.CreateSecret(ctx, params)
	require.NoError(t, err)

	var gotMeta []byte
	err = storage.db.QueryRowContext(ctx,
		`SELECT meta FROM secrets WHERE id = $1`, params.ID,
	).Scan(&gotMeta)
	require.NoError(t, err)
	require.Nil(t, gotMeta) // SQL NULL scans into a nil []byte
}
