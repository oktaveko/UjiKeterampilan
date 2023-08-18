package models

type CreditCard struct {
	Id     int    `json:"id"`
	Number string `json:"number"`
	UserId int
}
