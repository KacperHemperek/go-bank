package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"math/rand/v2"
)

type Storage interface {
	GetAccountByID(id int) (*Account, error)
	CreateAccount(account *Account) (*Account, error)
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
	return nil, nil

}

func (s *PostgresStorage) CreateAccount(account *Account) (*Account, error) {

	query := `
			INSERT INTO accounts (first_name, last_name, number, balance)
			VALUES ($1, $2, $3, 0)
			RETURNING id, first_name, last_name, number, balance, create_at, updated_at;
		`

	rows, err := s.db.Query(query, account.FirstName, account.LastName, int64(rand.IntN(1000000)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := &Account{}

	for rows.Next() {
		if err := rows.Scan(&result.ID, &result.FirstName, &result.LastName, &result.Number, &result.Balance, &result.CreateAt, &result.UpdateAt); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (s *PostgresStorage) DeleteAccount(id int) error {
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
