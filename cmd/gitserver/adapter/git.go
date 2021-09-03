package adapter

import (
	"context"
	"os/exec"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
)

// A Git adapter allows callers to execute Git commands which can affect local
// and/or remote systems.
type Git struct{}

// GC will invoke `git-gc` to clean up any garbage in the repo. It will
// operate synchronously and be aggressive with its internal heuristics when
// deciding to act (meaning it will act now at lower thresholds).
func (git *Git) GC(ctx context.Context, dir server.GitDir) error {
	cmd := exec.CommandContext(ctx, "git", "-c", "gc.auto=1", "-c", "gc.autoDetach=false", "gc", "--auto")
	dir.Set(cmd)
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(wrapCmdError(cmd, err), "failed to git-gc")
	}
	return nil
}
