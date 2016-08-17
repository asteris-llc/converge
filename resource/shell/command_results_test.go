package shell_test

import (
	"testing"

	"github.com/asteris-llc/converge/resource/shell"
)

func Test_Unlink_RemovesResult(t *testing.T) {
}

func newResults(op string, status uint32, stdout string) *shell.CommandResults {
	return &shell.CommandResults{
		ResultsContext: shell.ResultsContext{Operation: op},
		ExitStatus:     status,
		Stdout:         stdout,
	}
}
