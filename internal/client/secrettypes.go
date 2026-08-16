package client

type LoginPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Card struct {
	Number string `json:"number"`
	Holder string `json:"holder"`
	Expiry string `json:"expiry"`
	CVV    string `json:"cvv"`
}
