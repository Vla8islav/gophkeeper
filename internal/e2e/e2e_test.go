package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/Vla8islav/gophkeeper/internal/config"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/handler"
	"github.com/Vla8islav/gophkeeper/internal/repository"
	"github.com/Vla8islav/gophkeeper/internal/service"
)

// startPostgres spins up a throwaway Postgres and returns its DSN.
func startPostgres(t *testing.T) string {
	t.Helper()
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")
	ctx := context.Background()

	c, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("gophkeeper_e2e"),
		tcpostgres.WithUsername("default_user"),
		tcpostgres.WithPassword("default_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err)

	dsn, err := c.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	return dsn
}

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	cfg, err := config.ReadFlagsServer(nil, zap.NewNop())
	require.NoError(t, err)

	cfg.DatabaseURI.Value = startPostgres(t)
	cfg.DatabaseURI.BeenSet = true
	cfg.AuthTokenSecret.Value = "e2e-test-secret"
	cfg.AuthTokenSecret.BeenSet = true

	storage, err := repository.NewPostgresStorage(cfg, "../../migrations")
	require.NoError(t, err)

	svc := service.NewMetricsService(storage, cfg.AuthTokenSecret.Value)
	h := handler.NewHandler(svc, zap.NewNop())

	router := handler.NewRouter(h, cfg)

	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	return srv
}

// doJSON issues a JSON request, optionally with a Bearer token.
func doJSON(t *testing.T, srv *httptest.Server, method, path, token string, body []byte) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, srv.URL+path, bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	require.NoError(t, err)
	return resp
}

func TestE2E_RegisterLoginCreateSecret(t *testing.T) {
	srv := newTestServer(t)

	// 1. Register - 200 + token.
	resp := doJSON(t, srv, http.MethodPost, "/api/user/register", "",
		[]byte(`{"login":"e2e-user","password":"e2e-pass"}`))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var reg domain.UserRegisterResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&reg))
	resp.Body.Close()
	require.NotEmpty(t, reg.Token)

	// 2. Log in with the same creds - 200 + a fresh token.
	resp = doJSON(t, srv, http.MethodPost, "/api/user/login", "",
		[]byte(`{"login":"e2e-user","password":"e2e-pass"}`))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var login domain.UserLoginResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&login))
	resp.Body.Close()
	require.NotEmpty(t, login.Token)
	token := login.Token

	// 3. Create a secret WITH the token - 201. (Exercises WithAuth + context userID + insert.)
	secretID := uuid.New()
	createBody, err := json.Marshal(domain.CreateSecretRequest{
		ID:      secretID,
		Type:    domain.SecretTypeText,
		Payload: []byte("hello"),
	})
	require.NoError(t, err)

	resp = doJSON(t, srv, http.MethodPost, "/api/secret/create", token, createBody)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	// 4. Same request WITHOUT a token - 401. (This is the check that catches
	//    the "mounted outside the auth group" wiring bug.)
	resp = doJSON(t, srv, http.MethodPost, "/api/secret/create", "", createBody)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// 5. Re-create the SAME id - 409 (duplicate).
	resp = doJSON(t, srv, http.MethodPost, "/api/secret/create", token, createBody)
	require.Equal(t, http.StatusConflict, resp.StatusCode)
	resp.Body.Close()

	// 6. Garbage token - 401.
	resp = doJSON(t, srv, http.MethodPost, "/api/secret/create", "not-a-real-token", createBody)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// 7. Read the secret back - 200, payload round-trips.
	resp = doJSON(t, srv, http.MethodGet, "/api/secret/get/"+secretID.String(), token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var got domain.GetSecretResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	resp.Body.Close()
	require.Equal(t, secretID, got.ID)
	require.Equal(t, domain.SecretTypeText, got.Type)
	require.Equal(t, []byte("hello"), got.Payload) // base64 - []byte round-trip

	// 8. A second user must NOT see it - 404 (ownership isolation, full stack).
	resp = doJSON(t, srv, http.MethodPost, "/api/user/register", "",
		[]byte(`{"login":"e2e-intruder","password":"pw"}`))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var intruder domain.UserLoginResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&intruder))
	resp.Body.Close()

	resp = doJSON(t, srv, http.MethodGet, "/api/secret/get/"+secretID.String(), intruder.Token, nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode) // not 403 — non-leaky
	resp.Body.Close()

	// --- List path ---

	// L1. The owner's list contains the created secret, metadata only.
	resp = doJSON(t, srv, http.MethodGet, "/api/secret/list", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var list []domain.SecretSummaryResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	resp.Body.Close()
	require.Len(t, list, 1)
	require.Equal(t, secretID, list[0].ID)
	require.Equal(t, domain.SecretTypeText, list[0].Type)

	// L2. A brand-new user gets an empty list — and it must be [] , not null.
	resp = doJSON(t, srv, http.MethodPost, "/api/user/register", "",
		[]byte(`{"login":"e2e-list-other","password":"pw"}`))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var other domain.UserRegisterResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&other))
	resp.Body.Close()

	resp = doJSON(t, srv, http.MethodGet, "/api/secret/list", other.Token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	rawBody, err := io.ReadAll(resp.Body) // read raw, not decode
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, "[]", string(rawBody)) // proves nil-[] survives the full stack

	// L3. Requesting the list without a token - 401.
	resp = doJSON(t, srv, http.MethodGet, "/api/secret/list", "", nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// Updates
	upBody, err := json.Marshal(domain.UpdateSecretRequest{
		Payload: []byte("hello-v2"), Meta: []byte("label"), Version: 1,
	})
	require.NoError(t, err)

	resp = doJSON(t, srv, http.MethodPut, "/api/secret/update/"+secretID.String(), token, upBody)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var up domain.UpdateSecretResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&up))
	resp.Body.Close()
	require.Equal(t, int64(2), up.Version)

	// U2. Read back - new payload, version 2.
	resp = doJSON(t, srv, http.MethodGet, "/api/secret/get/"+secretID.String(), token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var afterUpdate domain.GetSecretResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&afterUpdate))
	resp.Body.Close()
	require.Equal(t, []byte("hello-v2"), afterUpdate.Payload)
	require.Equal(t, int64(2), afterUpdate.Version)

	// U3. A STALE update (still using version 1) - 409
	staleBody, err := json.Marshal(domain.UpdateSecretRequest{
		Payload: []byte("should-not-land"), Version: 1,
	})
	require.NoError(t, err)
	resp = doJSON(t, srv, http.MethodPut, "/api/secret/update/"+secretID.String(), token, staleBody)
	require.Equal(t, http.StatusConflict, resp.StatusCode)
	resp.Body.Close()

	// U4. The rejected write must NOT have landed — still v2, still "hello-v2"
	resp = doJSON(t, srv, http.MethodGet, "/api/secret/get/"+secretID.String(), token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var afterConflict domain.GetSecretResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&afterConflict))
	resp.Body.Close()
	require.Equal(t, []byte("hello-v2"), afterConflict.Payload) // unchanged
	require.Equal(t, int64(2), afterConflict.Version)

	// U5. Update at the CORRECT new version (2) - 200, bumps to 3.
	up2Body, err := json.Marshal(domain.UpdateSecretRequest{
		Payload: []byte("hello-v3"), Version: 2,
	})
	require.NoError(t, err)
	resp = doJSON(t, srv, http.MethodPut, "/api/secret/update/"+secretID.String(), token, up2Body)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&up))
	resp.Body.Close()
	require.Equal(t, int64(3), up.Version)

	// U6. Update a nonexistent secret - 404.
	missingBody, _ := json.Marshal(domain.UpdateSecretRequest{Payload: []byte("x"), Version: 1})
	resp = doJSON(t, srv, http.MethodPut, "/api/secret/update/"+uuid.New().String(), token, missingBody)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()

	// U7. Update without a token - 401.
	resp = doJSON(t, srv, http.MethodPut, "/api/secret/update/"+secretID.String(), "", upBody)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// --- Salt endpoint ---

	// S1. An authenticated user can fetch their KDF salt, and it's 16 bytes.
	resp = doJSON(t, srv, http.MethodGet, "/api/user/salt", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var salt1 domain.SaltResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&salt1))
	resp.Body.Close()
	require.Len(t, salt1.Salt, 16)

	// S2. The salt is STABLE across calls — same bytes every time.
	resp = doJSON(t, srv, http.MethodGet, "/api/user/salt", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var salt2 domain.SaltResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&salt2))
	resp.Body.Close()
	require.Equal(t, salt1.Salt, salt2.Salt)

	// S3. A different user gets a DIFFERENT salt (per-user randomness from crypto/rand).
	resp = doJSON(t, srv, http.MethodPost, "/api/user/register", "",
		[]byte(`{"login":"e2e-salt-other","password":"pw"}`))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var otherUser domain.UserRegisterResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&otherUser))
	resp.Body.Close()

	resp = doJSON(t, srv, http.MethodGet, "/api/user/salt", otherUser.Token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var otherSalt domain.SaltResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&otherSalt))
	resp.Body.Close()
	require.NotEqual(t, salt1.Salt, otherSalt.Salt)

	// S4. Without a token - 401.
	resp = doJSON(t, srv, http.MethodGet, "/api/user/salt", "", nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()

	// --- Delete path ---

	// D1. Delete the secret → 204.
	resp = doJSON(t, srv, http.MethodDelete, "/api/secret/delete/"+secretID.String(), token, nil)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	// D2. Gone from reads → 404.
	resp = doJSON(t, srv, http.MethodGet, "/api/secret/get/"+secretID.String(), token, nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()

	// D3. Gone from the list too.
	resp = doJSON(t, srv, http.MethodGet, "/api/secret/list", token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var listAfterDelete []domain.SecretSummaryResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&listAfterDelete))
	resp.Body.Close()
	for _, s := range listAfterDelete {
		require.NotEqual(t, secretID, s.ID) // the deleted one must not appear
	}

	// D4. Deleting again → 404 (already deleted; the AND deleted = FALSE at work).
	resp = doJSON(t, srv, http.MethodDelete, "/api/secret/delete/"+secretID.String(), token, nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()

	// D5. Delete without a token → 401.
	resp = doJSON(t, srv, http.MethodDelete, "/api/secret/delete/"+secretID.String(), "", nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}
