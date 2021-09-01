package indexing

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

var schemeToExternalService = map[string]string{
	"semanticdb": extsvc.KindJVMPackages,
}

func NewDependencySyncScheduler(
	workerStore dbworkerstore.Store,
) *workerutil.Worker {
	rootContext := actor.WithActor(context.Background(), &actor.Actor{Internal: true})

	handler := &dependencySyncSchedulerHandler{
		workerStore: workerStore,
	}

	return dbworker.NewWorker(rootContext, workerStore, handler, workerutil.WorkerOptions{})
}

type dependencySyncSchedulerHandler struct {
	workerStore dbworkerstore.Store
}

func (h *dependencySyncSchedulerHandler) Handle(ctx context.Context, record workerutil.Record) error {
	return nil
}
