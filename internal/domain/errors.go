package domain

import "fmt"

var ErrInvalidUserCredentials = fmt.Errorf("invalid user credentials")
var ErrOrderAlreadyUploadedByUser = fmt.Errorf("order already uploaded by user")
var ErrOrderAlreadyUploadedByAnotherUser = fmt.Errorf("order already uploaded by another user")
var ErrInvalidOrderNumber = fmt.Errorf("order didn't pass the luhn checksum check")
var ErrInvalidSum = fmt.Errorf("invalid order sum")
var ErrOrderEmpty = fmt.Errorf("empty order number")

var ErrNotEnoughMoney = fmt.Errorf("order not found")

var ErrSecretAlreadyExists = fmt.Errorf("secret already exists")
var ErrInvalidSecretType = fmt.Errorf("invalid secret type")
var ErrSecretNotFound = fmt.Errorf("secret not found")
var ErrInvalidSecretID = fmt.Errorf("invalid secret id")
var ErrVersionConflict = fmt.Errorf("secret version conflict")

var ErrSaltNotSet = fmt.Errorf("kdf salt not set for user")
