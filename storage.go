package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccounts() ([]*Account, error)
	GetAccountByID(int) (*Account, error)
  GetAccountByEmail(string) (*Account, error)
  CreateTransfer(*TransferRequest) (*TransferRequest, error)
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


func (s *PostgresStorage) Init() error {
    if err := s.createAccountTable(); err != nil {
        log.Println("Error creating accounts table:", err)
        return err
    }
    print("creating1")
    if err := s.createTransferTable(); err != nil {
        log.Println("Error creating transfers table:", err)
        return err
    }
        print("creating2")
    return nil
}


func (s *PostgresStorage) createAccountTable() error{
	query := `CREATE TABLE IF NOT EXISTS accounts (
        id SERIAL PRIMARY KEY,
        email VARCHAR(255) NOT NULL,
        first_name VARCHAR(255) NOT NULL,
        last_name VARCHAR(255) NOT NULL,
        number BIGINT UNIQUE NOT NULL,
        balance BIGINT NOT NULL,
        password VARCHAR(255) NOT NULL,
		    created_at TIMESTAMP DEFAULT NOW()
    )`
    _, err := s.db.Exec(query)
    return err
}

func (s *PostgresStorage) createTransferTable() error{
	query := `CREATE TABLE IF NOT EXISTS transfers (
        id SERIAL PRIMARY KEY,
        ToAccount BIGINT NOT NULL,
        FromAccount BIGINT NOT NULL,
        Amount BIGINT NOT NULL,
		    created_at TIMESTAMP DEFAULT NOW()
    )`
    _, err := s.db.Exec(query)
    return err
}

// inserting records
func (s *PostgresStorage) CreateAccount(acc *Account) error {
    query := `INSERT INTO accounts (email, first_name, last_name, number, balance, password) VALUES ($1, $2, $3, $4, $5, $6)`
    _, err := s.db.Exec(query, acc.Email, acc.FirstName, acc.LastName, acc.Number, acc.Balance, acc.Password)
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
   query := `DELETE FROM accounts WHERE id = $1`
   _, err := s.db.Exec(query, id)
   
  return err
}

func (s *PostgresStorage) GetAccountByID(id int) (*Account, error) {
   rows, err := s.db.Query(`SELECT* FROM accounts WHERE id = $1`, id)
   if err != nil {
     return nil, err
   }

   for rows.Next(){
    return scanIntoAccounts(rows)
   }
  return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgresStorage) GetAccountByEmail(email string) (*Account, error) {
  rows, err := s.db.Query(`SELECT* FROM accounts WHERE email = $1`, email)
   if err != nil {
     return nil, err
   }

  for rows.Next(){
    return scanIntoAccounts(rows)
  }
  return nil, fmt.Errorf("account with this email not found")
}

func (s *PostgresStorage) GetAccounts() ([]*Account, error) {
  rows, err := s.db.Query(`SELECT* FROM accounts`)

  if err != nil {
    return nil, err
  }

  accounts := []*Account{}

  for rows.Next() {
    account, err := scanIntoAccounts(rows)
    if err != nil {
        return nil, err
    }

    accounts = append(accounts, account)
}

  return  accounts, nil
}

func (s *PostgresStorage) CreateTransfer(trn *TransferRequest) (*TransferRequest, error){

  // Begin transaction
    tx, err := s.db.Begin()
    if err != nil {
        return nil, err
    }

    defer func() {
        if err != nil {
            tx.Rollback()
        } else {
            tx.Commit()
        }
    }()

  // Check if sender has enough balance
    var senderBalance int64
    checkBalanceQuery := `SELECT balance FROM accounts WHERE number = $1 FOR UPDATE`
    err = tx.QueryRow(checkBalanceQuery, trn.FromAccount).Scan(&senderBalance)
    if err != nil {
        return nil, err
    }

    if senderBalance < trn.Amount {
        return nil, fmt.Errorf("insufficient balance")
    }
  
   // Insert transfer record
    query := `INSERT INTO transfers (toaccount, fromaccount, amount) VALUES ($1, $2, $3)`
    _, err = s.db.Exec(query, trn.ToAccount, trn.FromAccount, trn.Amount)
    if err != nil {
        return nil, err
    }

  // Deduct amount from sender's balance
    updateFromBalance := `UPDATE accounts SET balance = balance - $1 WHERE number = $2`
    _, err = s.db.Exec(updateFromBalance, trn.Amount, trn.FromAccount)
    if err != nil {
        return nil, err
    }


  // Update recipient's balance
    updateToAccBalance := `UPDATE accounts SET balance = balance + $1 WHERE number = $2`
    _, err = s.db.Exec(updateToAccBalance, trn.Amount, trn.ToAccount)
    if err != nil {
        return nil, err
    }

  return trn, err
}

func scanIntoAccounts(rows *sql.Rows) (*Account, error) {
  account := new(Account)
  err := rows.Scan(
    &account.ID, 
    &account.Email, 
    &account.FirstName, 
    &account.LastName, 
    &account.Number, 
    &account.Balance, 
    &account.Password, 
    &account.CreatedAt,
    
  )

    return account, err
}