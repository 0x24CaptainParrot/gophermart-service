package service

type Order interface {
}

type Authorization interface {
	CreateUser()
	GenerateToken()
	ParseToken()
}

type Service struct {
	Order         Order
	Authorization Authorization
}


