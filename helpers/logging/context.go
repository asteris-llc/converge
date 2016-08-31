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
	"context"

	"github.com/Sirupsen/logrus"
)

var loggerKey = struct{}{}

// WithLogger sets a logger on a context
func WithLogger(ctx context.Context, logger interface{}) context.Context {
	switch t := logger.(type) {
	case *logrus.Logger:
		return context.WithValue(ctx, loggerKey, logrus.NewEntry(t))

	case *logrus.Entry:
		return context.WithValue(ctx, loggerKey, t)

	default:
		return ctx
	}
}

// GetLogger retrieves a logging context context
func GetLogger(ctx context.Context) *logrus.Entry {
	val := ctx.Value(loggerKey)

	if val != nil {
		if entry, ok := val.(*logrus.Entry); ok {
			return entry
		}
	}

	return logrus.WithFields(logrus.Fields{})
}
