package client

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/localstore"
)

func TestDisplaySecret_AllTypes(t *testing.T) {
	require.NoError(t, displaySecret(domain.SecretTypeText, []byte("hello"), "note"))

	lp, err := json.Marshal(LoginPassword{Login: "u", Password: "p"})
	require.NoError(t, err)
	require.NoError(t, displaySecret(domain.SecretTypeLoginPassword, lp, ""))

	card, err := json.Marshal(Card{Number: "4111", Holder: "A B", Expiry: "12/28", CVV: "123"})
	require.NoError(t, err)
	require.NoError(t, displaySecret(domain.SecretTypeCard, card, "visa"))

	require.NoError(t, displaySecret(domain.SecretTypeBinary, []byte{0x00, 0x01, 0xff}, "blob"))

	require.Error(t, displaySecret(domain.SecretTypeLoginPassword, []byte("not json"), ""))
}

func TestRunGet_DecryptsFromCache(t *testing.T) {
	cfg := testClientCfg(t, "") // salt is cached, so no server needed
	require.NoError(t, saveToken(cfg.TokenFile.Value, "tok"))

	store, err := localstore.Open(dbPath(cfg))
	require.NoError(t, err)

	salt := []byte("sixteen-byte-slt")
	require.NoError(t, store.SaveSalt(salt))
	key := DeriveKey("master-pw", salt)

	plaintext, err := json.Marshal(LoginPassword{Login: "alice", Password: "hunter2"})
	require.NoError(t, err)
	payload, err := Encrypt(key, plaintext)
	require.NoError(t, err)

	id := uuid.New()
	require.NoError(t, store.SaveSecret(localstore.Secret{
		ID: id.String(), Type: string(domain.SecretTypeLoginPassword), Payload: payload, Version: 1,
	}))
	require.NoError(t, store.Close())

	setStdin(t, "master-pw\n")
	require.NoError(t, runGet(cfg, []string{id.String()}))
}

func TestRunGet_NotCached(t *testing.T) {
	cfg := testClientCfg(t, "")
	require.NoError(t, saveToken(cfg.TokenFile.Value, "tok"))
	err := runGet(cfg, []string{uuid.New().String()})
	require.Error(t, err)
}

func TestDBPath(t *testing.T) {
	cfg := testClientCfg(t, "")
	require.Equal(t, filepath.Join(filepath.Dir(cfg.TokenFile.Value), "store.db"), dbPath(cfg))
}
