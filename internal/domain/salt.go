package domain

type SaltResponse struct {
	Salt []byte `json:"salt"` // base64 in JSON, decodes to []byte client-side
}
