package domain

type UserRegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserLoginRequest UserRegisterRequest

type CreateUserParams struct {
	Login        string
	PasswordHash string
}

type User struct {
	ID           int64
	Login        string
	PasswordHash string
}
