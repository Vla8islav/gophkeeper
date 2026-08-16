package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestReadFlagsClient_Defaults(t *testing.T) {
	cfg, rest, err := ReadFlagsClient(nil, zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, "https://localhost:8080", cfg.ServerAddress.Value)
	require.Equal(t, "", cfg.CACertPath.Value)
	require.NotEmpty(t, cfg.TokenFile.Value)
	require.Empty(t, rest)
}

func TestReadFlagsClient_FlagsAndSubcommandRemainder(t *testing.T) {
	cfg, rest, err := ReadFlagsClient(
		[]string{"-base-url", "https://example.com", "-token", "/tmp/tok", "get", "some-id"},
		zap.NewNop(),
	)
	require.NoError(t, err)
	require.Equal(t, "https://example.com", cfg.ServerAddress.Value)
	require.True(t, cfg.ServerAddress.BeenSet)
	require.Equal(t, "/tmp/tok", cfg.TokenFile.Value)
	require.Equal(t, []string{"get", "some-id"}, rest)
}

func TestReadFlagsClient_EnvOverridesFlags(t *testing.T) {
	t.Setenv("SERVER_ADDRESS", "https://env.example.com")

	cfg, _, err := ReadFlagsClient([]string{"-base-url", "https://flag.example.com"}, zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, "https://env.example.com", cfg.ServerAddress.Value)
}

func TestReadFlagsClient_ConfigFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "client.json")
	require.NoError(t, os.WriteFile(path,
		[]byte(`{"server_address":"https://file.example.com","ca_cert":"/etc/ca.pem"}`), 0o600))

	cfg, _, err := ReadFlagsClient([]string{"-config", path}, zap.NewNop())
	require.NoError(t, err)
	require.Equal(t, "https://file.example.com", cfg.ServerAddress.Value)
	require.Equal(t, "/etc/ca.pem", cfg.CACertPath.Value)
}

func TestReadFlagsClient_InvalidConfigFile(t *testing.T) {
	_, _, err := ReadFlagsClient([]string{"-config", "/no/such/client.json"}, zap.NewNop())
	require.Error(t, err)
}
