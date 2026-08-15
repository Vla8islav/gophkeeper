package domain

type SecretType string

const (
	SecretTypeLoginPassword SecretType = "login_password"
	SecretTypeText          SecretType = "text"
	SecretTypeBinary        SecretType = "binary"
	SecretTypeCard          SecretType = "card"
)

// Valid reports whether t is a known secret type.
func (t SecretType) Valid() bool {
	switch t {
	case SecretTypeLoginPassword, SecretTypeText, SecretTypeBinary, SecretTypeCard:
		return true
	default:
		return false
	}
}
