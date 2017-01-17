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
	"errors"
	"strings"
)

// Signal represents a Unix signal
type Signal uint

var (
	// ErrInvalidSignalName is returned when the signal name used is invalid
	ErrInvalidSignalName = errors.New("Invalid signal name")

	// ErrInvalidSignalNumber is returned when the signal number used is invalid
	ErrInvalidSignalNumber = errors.New("Invalid signal number")
)

var (
	SIGHUP    Signal = 1  // SIGHUP: terminal line hangup
	SIGINT    Signal = 2  // SIGINT: interrupt program
	SIGQUIT   Signal = 3  // SIGQUIT: quit program
	SIGILL    Signal = 4  // SIGILL: illegal instruction
	SIGTRAP   Signal = 5  // SIGTRAP: trace trap
	SIGABRT   Signal = 6  // SIGABRT: abort program (formerly SIGIOT)
	SIGEMT    Signal = 7  // SIGEMT: emulate instruction executed
	SIGFPE    Signal = 8  // SIGFPE: floating-point exception
	SIGKILL   Signal = 9  // SIGKILL: kill program
	SIGBUS    Signal = 10 // SIGBUS: bus error
	SIGSEGV   Signal = 11 // SIGSEGV: segmentation violation
	SIGSYS    Signal = 12 // SIGSYS: non-existent system call invoked
	SIGPIPE   Signal = 13 // SIGPIPE: write on a pipe with no reader
	SIGALRM   Signal = 14 // SIGALRM: real-time timer expired
	SIGTERM   Signal = 15 // SIGTERM: software termination signal
	SIGURG    Signal = 16 // SIGURG: urgent condition present on socket
	SIGSTOP   Signal = 17 // SIGSTOP: stop (cannot be caught or ignored)
	SIGTSTP   Signal = 18 // SIGTSTP: stop signal generated from keyboard
	SIGCONT   Signal = 19 // SIGCONT: continue after stop
	SIGCHLD   Signal = 20 // SIGCHLD: child status has changed
	SIGTTIN   Signal = 21 // SIGTTIN: background read attempted from control terminal
	SIGTTOU   Signal = 22 // SIGTTOU: background write attempted to control terminal
	SIGIO     Signal = 23 // SIGIO: I/O is possible on a descriptor (see fcntl(2))
	SIGXCPU   Signal = 24 // SIGXCPU: cpu time limit exceeded (see setrlimit(2))
	SIGXFSZ   Signal = 25 // SIGXFSZ: file size limit exceeded (see setrlimit(2))
	SIGVTALRM Signal = 26 // SIGVTALRM: virtual time alarm (see setitimer(2))
	SIGPROF   Signal = 27 // SIGPROF: profiling timer alarm (see setitimer(2))
	SIGWINCH  Signal = 28 // SIGWINCH: Window size change
	SIGINFO   Signal = 29 // SIGINFO: status request from keyboard
	SIGUSR1   Signal = 30 // SIGUSR1: User defined signal 1
	SIGUSR2   Signal = 31 // SIGUSR2: User defined signal 2
)

func (s Signal) String() string {
	switch s {
	case SIGHUP:
		return "SIGHUP"
	case SIGINT:
		return "SIGINT"
	case SIGQUIT:
		return "SIGQUIT"
	case SIGILL:
		return "SIGILL"
	case SIGTRAP:
		return "SIGTRAP"
	case SIGABRT:
		return "SIGABRT"
	case SIGEMT:
		return "SIGEMT"
	case SIGFPE:
		return "SIGFPE"
	case SIGKILL:
		return "SIGKILL"
	case SIGBUS:
		return "SIGBUS"
	case SIGSEGV:
		return "SIGSEGV"
	case SIGSYS:
		return "SIGSYS"
	case SIGPIPE:
		return "SIGPIPE"
	case SIGALRM:
		return "SIGALRM"
	case SIGTERM:
		return "SIGTERM"
	case SIGURG:
		return "SIGURG"
	case SIGSTOP:
		return "SIGSTOP"
	case SIGTSTP:
		return "SIGTSTP"
	case SIGCONT:
		return "SIGCONT"
	case SIGCHLD:
		return "SIGCHLD"
	case SIGTTIN:
		return "SIGTTIN"
	case SIGTTOU:
		return "SIGTTOU"
	case SIGIO:
		return "SIGIO"
	case SIGXCPU:
		return "SIGXCPU"
	case SIGXFSZ:
		return "SIGXFSZ"
	case SIGVTALRM:
		return "SIGVTALRM"
	case SIGPROF:
		return "SIGPROF"
	case SIGWINCH:
		return "SIGWINCH"
	case SIGINFO:
		return "SIGINFO"
	case SIGUSR1:
		return "SIGUSR1"
	case SIGUSR2:
		return "SIGUSR2"
	}
	return "Invalid signal"
}

// ParseSignalByName takes a string representing a signal name
func ParseSignalByName(s string) (Signal, error) {
	s = strings.ToUpper(s)
	if !strings.HasPrefix(s, "SIG") {
		s = "SIG" + s
	}
	switch s {
	case "SIGHUP":
		return SIGHUP, nil
	case "SIGINT":
		return SIGINT, nil
	case "SIGQUIT":
		return SIGQUIT, nil
	case "SIGILL":
		return SIGILL, nil
	case "SIGTRAP":
		return SIGTRAP, nil
	case "SIGABRT":
		return SIGABRT, nil
	case "SIGEMT":
		return SIGEMT, nil
	case "SIGFPE":
		return SIGFPE, nil
	case "SIGKILL":
		return SIGKILL, nil
	case "SIGBUS":
		return SIGBUS, nil
	case "SIGSEGV":
		return SIGSEGV, nil
	case "SIGSYS":
		return SIGSYS, nil
	case "SIGPIPE":
		return SIGPIPE, nil
	case "SIGALRM":
		return SIGALRM, nil
	case "SIGTERM":
		return SIGTERM, nil
	case "SIGURG":
		return SIGURG, nil
	case "SIGSTOP":
		return SIGSTOP, nil
	case "SIGTSTP":
		return SIGTSTP, nil
	case "SIGCONT":
		return SIGCONT, nil
	case "SIGCHLD":
		return SIGCHLD, nil
	case "SIGTTIN":
		return SIGTTIN, nil
	case "SIGTTOU":
		return SIGTTOU, nil
	case "SIGIO":
		return SIGIO, nil
	case "SIGXCPU":
		return SIGXCPU, nil
	case "SIGXFSZ":
		return SIGXFSZ, nil
	case "SIGVTALRM":
		return SIGVTALRM, nil
	case "SIGPROF":
		return SIGPROF, nil
	case "SIGWINCH":
		return SIGWINCH, nil
	case "SIGINFO":
		return SIGINFO, nil
	case "SIGUSR1":
		return SIGUSR1, nil
	case "SIGUSR2":
		return SIGUSR2, nil
	}
	return SIGUSR1, ErrInvalidSignalName
}

// ParseSignalByNumber takes a signal number and returns a Signal; it returns an
// error if the number is an invalid signal.
func ParseSignalByNumber(n uint) (Signal, error) {
	if n > 0 && n < 32 {
		return Signal(n), nil
	}
	return SIGUSR1, ErrInvalidSignalNumber
}
