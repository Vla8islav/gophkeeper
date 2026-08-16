package handler

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/Vla8islav/gophkeeper/internal/audit"
	"github.com/Vla8islav/gophkeeper/internal/domain"
	"github.com/Vla8islav/gophkeeper/internal/repository"
)

/*
Регистрация пользователя
Хендлер: POST /api/user/register

Регистрация производится по паре логин/пароль. Каждый логин должен быть уникальным.

После успешной регистрации должна происходить автоматическая аутентификация пользователя.

Для передачи аутентификационных данных используйте механизм cookies или HTTP-заголовок Authorization.

Формат запроса:

POST /api/user/register HTTP/1.1
Content-Type: application/json
...

{
    "login": "<login>",
    "password": "<password>"
}
Возможные коды ответа:

200 — пользователь успешно зарегистрирован и аутентифицирован;
400 — неверный формат запроса;
409 — логин уже занят;
500 — внутренняя ошибка сервера.
*/

func (h *Handler) UserRegisterHandler(w http.ResponseWriter, r *http.Request) {
	audit.SetOperation(r.Context(), "user.register")

	if r.Method != http.MethodPost {
		h.writeMethodNotAllowed(w, "only POST method is allowed")
		return
	}

	mimeType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if mimeType != "application/json" {
		h.writeBadRequest(w, "only application/json content type is supported")
		return
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		h.writeBadRequest(w, "failed to read request body: "+err.Error())
		return
	}

	var requestBodySerialised domain.UserRegisterRequest
	err = json.Unmarshal(requestBody, &requestBodySerialised)
	if err != nil {
		h.writeBadRequest(w, "couldn't parse requestBody:"+err.Error())
		return
	}

	if requestBodySerialised.Login == "" {
		h.writeBadRequest(w, "login cannot be empty")
		return
	}

	if requestBodySerialised.Password == "" {
		h.writeBadRequest(w, "password cannot be empty")
		return
	}

	authResult, err := h.service.CreateUser(r.Context(), requestBodySerialised)
	if errors.Is(err, repository.ErrUserAlreadyExists) {
		h.writeAlreadyExists(w, err.Error())
		return
	}

	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	loginResponse := domain.UserRegisterResponse{Token: authResult.Token}
	payload, err := json.Marshal(loginResponse)
	if err != nil {
		h.writeInternalServerError(w, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(payload)
	if err != nil {
		return
	}
}
