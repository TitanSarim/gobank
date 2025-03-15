package main

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type TransferRequest struct{
	ToAccount int64    `json:"toAccount"`
	FromAccount int64    `json:"fromAccount"`
	Amount int64  `json:"amount"`
	CreatedAt time.Time `json:"createdAt`
}

type CreateAccountRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Password string `json:"password"`
}

type Account struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Number    int64  `json:"number"`
	Balance   int64  `json:"balance"`
    Password  string `json:"password"`
	CreatedAt time.Time `json:"createdAt`
}

func NewAccount(email, firstName, lastName, password string) (*Account, error) {

	encPw, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
        return nil, err
    }

	return &Account{
		Email: email,
		FirstName: firstName,
		LastName: lastName,
		Number:   int64(rand.Intn(100000)),
		Balance:  0,
		Password: string(encPw),
		CreatedAt: time.Now().UTC(),
	}, nil
}