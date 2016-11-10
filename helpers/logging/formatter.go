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

package logging

import (
	"bytes"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

// Formatter is a compact output for logrus logs
type Formatter struct {
	DisableColors bool
	fancy         bool
	replacer      *strings.Replacer
	colors        map[string]func(string) string

	initter sync.Once
}

func (f *Formatter) init() {
	isColorTerminal := logrus.IsTerminal() && (runtime.GOOS != "windows")
	f.fancy = (!f.DisableColors && isColorTerminal)

	f.replacer = strings.NewReplacer(
		`"`, `\"`,
		`\n`, `\\n`,
	)

	// set up colors
	reset := "\x1b[0m"
	f.colors = map[string]func(string) string{
		"blue":   func(in string) string { return "\x1b[34m" + in + reset },
		"red":    func(in string) string { return "\x1b[31m" + in + reset },
		"white":  func(in string) string { return "\x1b[37;1m" + in + reset },
		"yellow": func(in string) string { return "\x1b[33m" + in + reset },
		"dim":    func(in string) string { return "\x1b[2m" + in + reset },
	}
}

// Format an entry for printing
func (f *Formatter) Format(entry *logrus.Entry) ([]byte, error) {
	f.initter.Do(f.init)

	var b bytes.Buffer

	// first, a timestamp
	b.WriteString(f.timestamp(entry.Time))
	b.WriteByte(' ')

	// then, our error level
	b.WriteString(f.level(entry.Level))
	b.WriteByte(' ')

	// next, the message
	b.WriteString(f.message(entry.Message))
	if f.fancy {
		b.WriteByte('\t')
	} else {
		b.WriteByte(' ')
	}

	// and our sorted keys
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	last := len(keys) - 1
	for i, k := range keys {
		b.WriteString(f.kv(k, entry.Data[k]))
		if i != last {
			b.WriteByte(' ')
		}
	}

	// end with a newline
	b.WriteByte('\n')

	return b.Bytes(), nil
}

func (f *Formatter) timestamp(t time.Time) string {
	formatted := t.Format(time.RFC3339)

	if f.fancy {
		return f.colors["dim"](formatted)
	}

	return f.kv("timestamp", formatted)
}

func (f *Formatter) level(lvl logrus.Level) string {
	var (
		level    string
		colorize func(string) string
	)

	switch lvl {
	case logrus.DebugLevel:
		level = "DEBUG"
		colorize = f.colors["white"]

	case logrus.InfoLevel:
		level = "INFO"
		colorize = f.colors["blue"]

	case logrus.WarnLevel:
		level = "WARN"
		colorize = f.colors["yellow"]

	case logrus.ErrorLevel:
		level = "ERROR"
		colorize = f.colors["red"]

	case logrus.FatalLevel:
		level = "FATAL"
		colorize = f.colors["red"]

	default:
		level = "UNKNOWN"
		colorize = f.colors["red"]
	}

	if !f.fancy {
		return f.kv("level", level)
	}

	return fmt.Sprintf("|%s|", colorize(level))
}

func (f *Formatter) message(msg string) string {
	if !f.fancy {
		return f.kv("msg", msg)
	}

	return f.colors["white"](msg)
}

func (f *Formatter) kv(key string, value interface{}) string {
	if !f.fancy {
		return fmt.Sprintf("%s=\"%v\"", key, f.replacer.Replace(fmt.Sprint(value)))
	}

	var color func(string) string
	if key == "error" {
		color = f.colors["red"]
	} else {
		color = f.colors["blue"]
	}

	return fmt.Sprintf("%s=%v", color(key), value)
}
