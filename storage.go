package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	connStr := "user=postgres dbname=go_bank password=12345 sslmode=disable"
	db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }

    err = db.Ping()
    if err != nil {
        return nil, err
    }

    return &PostgresStorage{db: db}, nil
}


func (s *PostgresStorage) Init() error{
	return s.createAccountTable()
}

func (s *PostgresStorage) createAccountTable() error{
	query := `CREATE TABLE IF NOT EXISTS accounts (
        id SERIAL PRIMARY KEY,
        first_name VARCHAR(255) NOT NULL,
        last_name VARCHAR(255) NOT NULL,
        number BIGINT UNIQUE NOT NULL,
        balance BIGINT NOT NULL,
		created_at TIMESTAMP DEFAULT NOW()
    )`
    _, err := s.db.Exec(query)
    return err
}

// inserting records
func (s *PostgresStorage) CreateAccount(acc *Account) error {
    query := `INSERT INTO accounts (first_name, last_name, number, balance) VALUES ($1, $2, $3, $4)`
    _, err := s.db.Exec(query, acc.FirstName, acc.LastName, acc.Number, acc.Balance)
    return err
}

func (s *PostgresStorage) UpdateAccount(acc *Account) error {
    query := `INSERT INTO accounts (first_name, last_name, number, balance) VALUES ($1, $2, $3, $4)`
    res, err := s.db.Exec(query, acc.FirstName, acc.LastName, acc.Number, acc.Balance)
	if err != nil {
		return err
	}
    fmt.Printf("%+v\n", res)
	return nil
}

func (s *PostgresStorage) DeleteAccount(id int) error {
  return nil
}

func (s *PostgresStorage) GetAccountByID(id int) (*Account, error) {
  return nil, nil
}

func (s *PostgresStorage) GetAccounts() ([]*Account, error) {
  rows, err := s.db.Query(`SELECT* FROM accounts`)

  if err != nil {
    return nil, err
  }

  accounts := []*Account{}

  for rows.Next() {
    account := new(Account)
    err := rows.Scan(
        &account.ID, 
        &account.FirstName, 
        &account.LastName, 
        &account.Number, 
        &account.Balance, 
        &account.CreatedAt,
    )
    
    if err != nil {
        return nil, err
    }

    accounts = append(accounts, account)
}

  return  accounts, nil
}