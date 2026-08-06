package repository

import (
	"context"
	"testing"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/stretchr/testify/require"
)

func TestPostgresStorage_Ping(t *testing.T) {
	cfg := config.ReadFlagsServer(nil)
	storage := InitTestPostgresStorage(t, cfg)

	ctx := context.Background()

	err := storage.Ping(ctx)
	require.NoError(t, err)
}
