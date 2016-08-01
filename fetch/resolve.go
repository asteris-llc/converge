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

package fetch

import (
	"log"
	"path"
	"strings"
)

// ResolveInContext resolves a path relative to another
func ResolveInContext(loc, ctx string) (string, error) {
	log.Printf("[TRACE] resolving URL %q in context of URL %q\n", loc, ctx)

	var (
		locScheme, locPath = parse(loc)
		ctxScheme, ctxPath = parse(ctx)
	)

	if ctx != "" && loc != ctx && !path.IsAbs(locPath) && (locScheme == "" || locScheme == ctxScheme) {
		locPath = path.Join(path.Dir(ctxPath), locPath)
		locScheme = ctxScheme
	}

	if locScheme == "" {
		locScheme = "file"
	}

	return locScheme + "://" + locPath, nil
}

func parse(loc string) (scheme, path string) {
	if strings.Contains(loc, "://") {
		parts := strings.SplitN(loc, "://", 2)
		return parts[0], parts[1]
	}

	return "", loc
}
