package models

type User struct {
	Model
	Id int
	Username string
	Credits float64
}