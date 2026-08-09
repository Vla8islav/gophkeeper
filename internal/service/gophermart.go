package service

import (
	"github.com/Vla8islav/gophkeeper/internal/domain"
)

type gophkeeperService struct {
	repository domain.GophkeeperRepository
	authSecret []byte
}

func NewMetricsService(
	repo domain.GophkeeperRepository,
	authSecret string,
) domain.GophkeeperService {
	return gophkeeperService{
		repository: repo,
		authSecret: []byte(authSecret),
	}
}
