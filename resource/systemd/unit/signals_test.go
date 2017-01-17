// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package unit

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSignalByName(t *testing.T) {
	t.Parallel()
	t.Run("when-uppercase", func(t *testing.T) {
		t.Parallel()
		for k, v := range signalEnumToStringMap {
			parsed, err := ParseSignalByName(v)
			require.NoError(t, err)
			assert.Equal(t, k, parsed)
		}
	})
	t.Run("when-lowercase", func(t *testing.T) {
		t.Parallel()
		for k, v := range signalEnumToStringMap {
			parsed, err := ParseSignalByName(strings.ToLower(v))
			require.NoError(t, err)
			assert.Equal(t, k, parsed)
		}
	})
	t.Run("when-mixed-case", func(t *testing.T) {
		t.Parallel()
		for k, v := range signalEnumToStringMap {
			parsed, err := ParseSignalByName(randomizeCase(v))
			require.NoError(t, err)
			assert.Equal(t, k, parsed)
		}
	})
	t.Run("when-invalid", func(t *testing.T) {
		t.Parallel()
		_, err := ParseSignalByName("")
		assert.Error(t, err)
		_, err = ParseSignalByName("bad1")
		assert.Error(t, err)
	})
}

// TestSignalString runs a test
func TestSignalString(t *testing.T) {
	t.Parallel()

	for k, v := range signalEnumToStringMap {
		assert.Equal(t, v, k.String())
	}
}

// TestParseSignalByNumber runs a test
func TestParseSignalByNumber(t *testing.T) {
	t.Parallel()

	t.Run("when-valid", func(t *testing.T) {
		t.Parallel()

		for idx, val := range orderedSignals {
			parsed, err := ParseSignalByNumber(uint(idx + 1))
			require.NoError(t, err)
			assert.Equal(t, val, parsed)
		}
	})

	t.Run("when-not-valid", func(t *testing.T) {
		t.Parallel()

		_, err := ParseSignalByNumber(0)
		require.Error(t, err)

		_, err = ParseSignalByNumber(77)
		require.Error(t, err)
	})

}

var orderedSignals = []Signal{
	SIGHUP,
	SIGINT,
	SIGQUIT,
	SIGILL,
	SIGTRAP,
	SIGABRT,
	SIGEMT,
	SIGFPE,
	SIGKILL,
	SIGBUS,
	SIGSEGV,
	SIGSYS,
	SIGPIPE,
	SIGALRM,
	SIGTERM,
	SIGURG,
	SIGSTOP,
	SIGTSTP,
	SIGCONT,
	SIGCHLD,
	SIGTTIN,
	SIGTTOU,
	SIGIO,
	SIGXCPU,
	SIGXFSZ,
	SIGVTALRM,
	SIGPROF,
	SIGWINCH,
	SIGINFO,
	SIGUSR1,
	SIGUSR2,
}

var signalEnumToStringMap = map[Signal]string{
	SIGHUP:    "SIGHUP",
	SIGINT:    "SIGINT",
	SIGQUIT:   "SIGQUIT",
	SIGILL:    "SIGILL",
	SIGTRAP:   "SIGTRAP",
	SIGABRT:   "SIGABRT",
	SIGEMT:    "SIGEMT",
	SIGFPE:    "SIGFPE",
	SIGKILL:   "SIGKILL",
	SIGBUS:    "SIGBUS",
	SIGSEGV:   "SIGSEGV",
	SIGSYS:    "SIGSYS",
	SIGPIPE:   "SIGPIPE",
	SIGALRM:   "SIGALRM",
	SIGTERM:   "SIGTERM",
	SIGURG:    "SIGURG",
	SIGSTOP:   "SIGSTOP",
	SIGTSTP:   "SIGTSTP",
	SIGCONT:   "SIGCONT",
	SIGCHLD:   "SIGCHLD",
	SIGTTIN:   "SIGTTIN",
	SIGTTOU:   "SIGTTOU",
	SIGIO:     "SIGIO",
	SIGXCPU:   "SIGXCPU",
	SIGXFSZ:   "SIGXFSZ",
	SIGVTALRM: "SIGVTALRM",
	SIGPROF:   "SIGPROF",
	SIGWINCH:  "SIGWINCH",
	SIGINFO:   "SIGINFO",
	SIGUSR1:   "SIGUSR1",
	SIGUSR2:   "SIGUSR2",
}
