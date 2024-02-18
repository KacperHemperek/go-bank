package main

type AccountCreateRequest struct {
	FirstName string `json:"firstName" validate:"required,min=2,max=40"`
	LastName  string `json:"lastName" validate:"required,min=2,max=40"`
}
type Account struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Number    int64  `json:"number"`
	Balance   int64  `json:"balance"`
	CreateAt  string `json:"createAt"`
	UpdateAt  string `json:"updateAt"`
}

func NewAccount(firstName, lastName string) *Account {
	return &Account{
		FirstName: firstName,
		LastName:  lastName,
	}
}
