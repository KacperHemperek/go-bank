package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"math/rand/v2"
)

type Storage interface {
	GetAccountByID(id int) (*Account, error)
	CreateAccount(account *Account) (*Account, error)
	GetAccounts() ([]*Account, error)
	DeleteAccount(id int) error
	Transfer(from, to int, amount int64) error
}

type PostgresStorage struct {
	db *sql.DB
}

func (s *PostgresStorage) Init() error {
	if err := s.createAccountTable(); err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) createAccountTable() error {
	_, err := s.db.Query(`
		CREATE TABLE IF NOT EXISTS accounts (
			id SERIAL PRIMARY KEY NOT NULL,
			first_name TEXT NOT NULL,
			last_name TEXT NOT NULL,
			number BIGINT NOT NULL,
			balance BIGINT NOT NULL,
			create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)

	return err
}

func (s *PostgresStorage) GetAccountByID(id int) (*Account, error) {
	query := `SELECT * FROM accounts WHERE id = $1;`

	rows, err := s.db.Query(query, id)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("account %d not found", id)
}

func (s *PostgresStorage) GetAccounts() ([]*Account, error) {
	query := `SELECT * FROM accounts`

	rows, err := s.db.Query(query)

	if err != nil {
		return nil, err
	}

	accounts := make([]*Account, 0)

	for rows.Next() {

		acc, err := scanIntoAccount(rows)

		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func (s *PostgresStorage) CreateAccount(account *Account) (*Account, error) {

	query := `
			INSERT INTO accounts (first_name, last_name, number, balance)
			VALUES ($1, $2, $3, 0)
			RETURNING *;
		`

	rows, err := s.db.Query(query, account.FirstName, account.LastName, int64(rand.IntN(1000000)))
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return scanIntoAccount(rows)
	}

	return nil, fmt.Errorf("could not create account")
}

func (s *PostgresStorage) DeleteAccount(id int) error {
	query := `DELETE FROM accounts WHERE id = $1;`

	_, err := s.db.Query(query, id)

	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) Transfer(from, to int, amount int64) error {

	return nil
}

func NewPostgresStorage() (*PostgresStorage, error) {
	conn := "user=postgres password=gobank dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", conn)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func scanIntoAccount(rows *sql.Rows) (*Account, error) {
	a := &Account{}

	if err := rows.Scan(&a.ID, &a.FirstName, &a.LastName, &a.Number, &a.Balance, &a.CreateAt, &a.UpdateAt); err != nil {
		return nil, err
	}
	return a, nil
}
