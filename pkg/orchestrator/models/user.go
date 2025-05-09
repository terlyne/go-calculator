package models

type User struct {
	Login    string
	Password string
}

var users = make(map[string]string) // Хранение пользователей (потом поменяю хранение на БД)
