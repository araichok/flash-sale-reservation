package product

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(
	ctx context.Context,
	name string,
	stock int,
) (*Product, error) {
	// минимальная бизнес-проверка
	if stock < 0 {
		stock = 0
	}

	return s.repo.Create(ctx, name, stock)
}

func (s *Service) List(ctx context.Context) ([]Product, error) {
	return s.repo.List(ctx)
}
