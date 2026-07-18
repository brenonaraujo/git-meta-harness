package source

import "os/exec"

// osExec is a thin wrapper over os/exec to centralize command
// execution. It returns combined output (so error messages include
// stderr). Returns an error if the command exits non-zero.
func osExec(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return out, err
}
