package client

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/localstore"
)

func TestGet_OfflineRoundTrip(t *testing.T) {
	store, err := localstore.Open(filepath.Join(t.TempDir(), "store.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	salt := []byte("sixteen-byte-slt")
	key := DeriveKey("master-pw", salt)

	plaintext, err := json.Marshal(LoginPassword{Login: "alice", Password: "hunter2"})
	require.NoError(t, err)
	payloadCipher, err := Encrypt(key, plaintext)
	require.NoError(t, err)
	metaCipher, err := Encrypt(key, []byte("github"))
	require.NoError(t, err)

	const id = "11111111-1111-1111-1111-111111111111"
	require.NoError(t, store.SaveSecret(localstore.Secret{
		ID:      id,
		Type:    string(domain.SecretTypeLoginPassword),
		Payload: payloadCipher,
		Meta:    metaCipher,
		Version: 1,
	}))

	sec, err := store.GetSecret(id)
	require.NoError(t, err)
	require.Equal(t, id, sec.ID)

	gotPlain, err := Decrypt(key, sec.Payload)
	require.NoError(t, err)
	require.Equal(t, plaintext, gotPlain)

	var lp LoginPassword
	require.NoError(t, json.Unmarshal(gotPlain, &lp))
	require.Equal(t, "alice", lp.Login)
	require.Equal(t, "hunter2", lp.Password)

	gotMeta, err := Decrypt(key, sec.Meta)
	require.NoError(t, err)
	require.Equal(t, "github", string(gotMeta))

	_, err = Decrypt(DeriveKey("wrong-pw", salt), sec.Payload)
	require.Error(t, err)
}
