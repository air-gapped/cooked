package server

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m,
		// regexp2's fastclock goroutine (used by chroma for regex timeouts).
		goleak.IgnoreAnyFunction("github.com/dlclark/regexp2.runClock"),
	)
}
