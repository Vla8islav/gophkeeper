package client

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeriveKey_DeterministicAndSized(t *testing.T) {
	salt := []byte("sixteen-byte-slt")
	k1 := DeriveKey("master-pw", salt)
	k2 := DeriveKey("master-pw", salt)
	require.Len(t, k1, 32)                               // AES-256
	require.Equal(t, k1, k2)                             // same password + salt -> same key
	require.NotEqual(t, k1, DeriveKey("other-pw", salt)) // different password -> different key
}

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	key := DeriveKey("pw", []byte("sixteen-byte-slt"))
	plaintext := []byte("secret with a \x00 null byte")

	ct, err := Encrypt(key, plaintext)
	require.NoError(t, err)
	require.NotEqual(t, plaintext, ct)

	got, err := Decrypt(key, ct)
	require.NoError(t, err)
	require.Equal(t, plaintext, got)
}

func TestDecrypt_WrongKeyFails(t *testing.T) {
	salt := []byte("sixteen-byte-slt")
	ct, err := Encrypt(DeriveKey("right", salt), []byte("hello"))
	require.NoError(t, err)

	_, err = Decrypt(DeriveKey("wrong", salt), ct)
	require.Error(t, err) // GCM tag verification fails
}

func TestDecrypt_TooShort(t *testing.T) {
	key := DeriveKey("pw", []byte("sixteen-byte-slt"))
	_, err := Decrypt(key, []byte{0x01, 0x02})
	require.Error(t, err)
}

func TestEncrypt_UniqueNonces(t *testing.T) {
	key := DeriveKey("pw", []byte("sixteen-byte-slt"))
	a, err := Encrypt(key, []byte("same"))
	require.NoError(t, err)
	b, err := Encrypt(key, []byte("same"))
	require.NoError(t, err)
	require.NotEqual(t, a, b) // random nonce -> same plaintext encrypts differently
}
