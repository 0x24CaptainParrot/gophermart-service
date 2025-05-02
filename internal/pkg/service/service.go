package service

type Service struct {
	Authorization   Authorization
	Order           Order
	Balance         Balance
	OrderProcessing OrderProcessing
}

type Dependencies struct {
	Authorization   Authorization
	Order           Order
	Balance         Balance
	OrderProcessing OrderProcessing
}

func NewService(deps Dependencies) *Service {
	return &Service{
		Authorization:   deps.Authorization,
		Order:           deps.Order,
		Balance:         deps.Balance,
		OrderProcessing: deps.OrderProcessing,
	}
}
