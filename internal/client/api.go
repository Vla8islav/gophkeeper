package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/google/uuid"
)

type APIClient struct {
	baseURL string
	http    *http.Client
}

// NewAPIClient builds a client that verifies the server cert. caPath is optional
func NewAPIClient(baseURL, caPath string) (*APIClient, error) {
	transport := &http.Transport{}
	if caPath != "" {
		pem, err := os.ReadFile(caPath)
		if err != nil {
			return nil, fmt.Errorf("read CA: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("no certificates found in %s", caPath)
		}
		transport.TLSClientConfig = &tls.Config{RootCAs: pool}
	}
	return &APIClient{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second, Transport: transport},
	}, nil
}

func (c *APIClient) Register(login, password string) (string, error) {
	return c.authRequest("/api/user/register", login, password)
}

func (c *APIClient) Login(login, password string) (string, error) {
	return c.authRequest("/api/user/login", login, password)
}

func (c *APIClient) authRequest(path, login, password string) (string, error) {
	body, err := json.Marshal(domain.UserRegisterRequest{Login: login, Password: password})
	if err != nil {
		return "", err
	}
	resp, err := c.http.Post(c.baseURL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned %s", resp.Status)
	}

	var out domain.UserLoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	return out.Token, nil
}

func (c *APIClient) ListSecrets(token string) ([]domain.SecretSummaryResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/secret/list", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("unauthorized — try logging in again")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %s", resp.Status)
	}

	var out []domain.SecretSummaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *APIClient) CreateSecret(token string, req domain.CreateSecretRequest) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/secret/create", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned %s", resp.Status)
	}
	return nil
}

func (c *APIClient) GetUserSalt(token string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/user/salt", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("unauthorized — try logging in again")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %s", resp.Status)
	}
	var out domain.SaltResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Salt, nil
}

func (c *APIClient) DeleteSecret(token string, id uuid.UUID) error {
	req, err := http.NewRequest(http.MethodDelete,
		c.baseURL+"/api/secret/delete/"+id.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNoContent:
		return nil
	case http.StatusNotFound:
		return fmt.Errorf("secret %s not found", id)
	case http.StatusUnauthorized:
		return fmt.Errorf("unauthorized — try logging in again")
	default:
		return fmt.Errorf("server returned %s", resp.Status)
	}
}

func (c *APIClient) SyncSecrets(token string) ([]domain.SyncSecretResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/secret/sync", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("unauthorized — try logging in again")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned %s", resp.Status)
	}

	var out []domain.SyncSecretResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *APIClient) UpdateSecret(token string, id uuid.UUID, req domain.UpdateSecretRequest) (int64, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}
	httpReq, err := http.NewRequest(http.MethodPut,
		c.baseURL+"/api/secret/update/"+id.String(), bytes.NewReader(body))
	if err != nil {
		return 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var out domain.UpdateSecretResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return 0, err
		}
		return out.Version, nil
	case http.StatusConflict:
		return 0, fmt.Errorf("version conflict — run `sync` and try again")
	case http.StatusNotFound:
		return 0, fmt.Errorf("secret %s not found", id)
	case http.StatusUnauthorized:
		return 0, fmt.Errorf("unauthorized — try logging in again")
	default:
		return 0, fmt.Errorf("server returned %s", resp.Status)
	}
}
