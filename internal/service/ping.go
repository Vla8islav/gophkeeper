package service

import "context"

func (m gophkeeperService) Ping(ctx context.Context) error {
	return m.repository.Ping(ctx)
}
