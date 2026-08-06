package service

import (
	"github.com/Vla8islav/gophkeeper/internal/domain"
)

type gophkeeperService struct {
	repository      domain.GophkeeperRepository
	authSecret      []byte
	pollingInterval int
}

func NewMetricsService(
	repo domain.GophkeeperRepository,
	authSecret string,
	pollingInterval int,
) domain.GophkeeperService {
	return gophkeeperService{
		repository:      repo,
		authSecret:      []byte(authSecret),
		pollingInterval: pollingInterval,
	}
}
