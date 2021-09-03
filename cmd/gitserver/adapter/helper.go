package adapter

import (
	"os/exec"
	"strings"

	"github.com/cockroachdb/errors"
)

// wrapCmdError will wrap errors for cmd to include the arguments. If the error
// is an exec.ExitError and cmd was invoked with Output(), it will also include
// the captured stderr.
func wrapCmdError(cmd *exec.Cmd, err error) error {
	if err == nil {
		return nil
	}
	var e *exec.ExitError
	if errors.As(err, &e) {
		return errors.Wrapf(err, "%s %s failed with stderr: %s", cmd.Path, strings.Join(cmd.Args, " "), string(e.Stderr))
	}
	return errors.Wrapf(err, "%s %s failed", cmd.Path, strings.Join(cmd.Args, " "))
}
