package domain

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
)

// A RepositoryCleaner ensures the integrity of repositories on disk. If serious
// issues are discovered, it will flag a repository for removal.
type RepositoryCleaner struct {
	GarbageCollector interface {
		GC(ctx context.Context, dir server.GitDir) error
	}
}

// Clean executes janitoral tasks in the context of a single repository.
func (r *RepositoryCleaner) Clean(ctx context.Context, dir server.GitDir) error {
	var err error

	// TODO: add further cleanup tasks

	if err = r.GarbageCollector.GC(ctx, dir); err != nil {
		return err
	}

	return nil
}
