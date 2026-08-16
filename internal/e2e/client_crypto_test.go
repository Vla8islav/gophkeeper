package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Vla8islav/gophkeeper/internal/client"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestE2E_ClientEncryptionRoundTrip(t *testing.T) {
	srv := newTestServer(t)

	// plain; TLS only kicks in with a caPath.
	api, err := client.NewAPIClient(srv.URL, "")
	require.NoError(t, err)

	// 1. Register - token.
	token, err := api.Register("cli-crypto-user", "auth-pass")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	// 2. Fetch the KDF salt and derive the encryption key
	salt, err := api.GetUserSalt(token)
	require.NoError(t, err)
	require.Len(t, salt, 16)

	const masterPw = "correct-horse-battery-staple"
	key := client.DeriveKey(masterPw, salt)

	// 3. Encrypt a login+password secret and its label
	plaintext, err := json.Marshal(client.LoginPassword{Login: "me", Password: "hunter2"})
	require.NoError(t, err)
	payloadCipher, err := client.Encrypt(key, plaintext)
	require.NoError(t, err)
	metaCipher, err := client.Encrypt(key, []byte("my github"))
	require.NoError(t, err)

	// 4. Upload the CIPHERTEXT
	id := uuid.New()
	require.NoError(t, api.CreateSecret(token, domain.CreateSecretRequest{
		ID:      id,
		Type:    domain.SecretTypeLoginPassword,
		Payload: payloadCipher,
		Meta:    metaCipher,
	}))

	// 5. Read the raw stored secret back from the server
	resp := doJSON(t, srv, http.MethodGet, "/api/secret/get/"+id.String(), token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var stored domain.GetSecretResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&stored))
	resp.Body.Close()

	// Checks:

	// The server stored ciphertext instad of a plaintext
	require.NotEqual(t, plaintext, stored.Payload)
	require.NotContains(t, string(stored.Payload), "hunter2")
	require.NotContains(t, string(stored.Meta), "github")
	// check data
	require.Equal(t, payloadCipher, stored.Payload)
	require.Equal(t, metaCipher, stored.Meta)

	// test decryption
	gotPlain, err := client.Decrypt(key, stored.Payload)
	require.NoError(t, err)
	require.Equal(t, plaintext, gotPlain)

	var lp client.LoginPassword
	require.NoError(t, json.Unmarshal(gotPlain, &lp))
	require.Equal(t, "me", lp.Login)
	require.Equal(t, "hunter2", lp.Password)

	gotMeta, err := client.Decrypt(key, stored.Meta)
	require.NoError(t, err)
	require.Equal(t, "my github", string(gotMeta))

	// it's not decryptable with the wrong password
	wrongKey := client.DeriveKey("wrong-master-password", salt)
	_, err = client.Decrypt(wrongKey, stored.Payload)
	require.Error(t, err)
}
