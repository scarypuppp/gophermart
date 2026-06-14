package accrual_poller

import (
	"context"

	"github.com/scarypuppp/gophermart/internal/repository"
)

type AccrualPoller struct {
	Address string
	uow     repository.UnitOfWork
}

func RunCtx(ctx context.Context) {

}
