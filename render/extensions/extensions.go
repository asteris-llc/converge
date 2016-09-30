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

package extensions

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"

	log "github.com/Sirupsen/logrus"

	"github.com/asteris-llc/converge/render/extensions/platform"
)

// RefFuncName is the name of the function to reference exported values from
// other nodes.  It is a const defined here to make it easily changeable to
// avoid bikeshedding
const RefFuncName string = "lookup"

// languageKeywords defines the known keywords that have been added to the
// templating language.  This is stored as a map for quick lookup and is used
// for DSL validation.
var languageKeywords = map[string]struct{}{
	"env":       {},
	"split":     {},
	"join":      {},
	RefFuncName: {},
	"platform":  {},
	"jsonify":   {},

	// functions for working with parameters
	"param":     {},
	"paramList": {},
	"paramMap":  {},
}

// LanguageExtension is a type wrapper around a template.FuncMap to allow us to
// encapsulate more context in the future.
type LanguageExtension struct {
	Funcs template.FuncMap

	innerLock *sync.RWMutex
}

// EmptyLanguage will create an empty language
func EmptyLanguage() *LanguageExtension {
	return &LanguageExtension{innerLock: new(sync.RWMutex)}
}

// MakeLanguage provides an empty language that implements a nil operation for
// all known keywords.
func MakeLanguage() *LanguageExtension {
	funcs := template.FuncMap{}
	for keyword := range languageKeywords {
		funcs[keyword] = StubTemplateFunc
	}
	return &LanguageExtension{Funcs: funcs, innerLock: new(sync.RWMutex)}
}

// MinimalLanguage provides a language extension where all known extensions are
// associated with NOP functions- as with MakeLanguage()- except that arity,
// input, and output types are respected and pure transformations are
// implemented.  It is less featureful than DefaultLanguage but will not
// introduce template errors that may be present when using an unmodified
// MakeLanguage.
func MinimalLanguage() *LanguageExtension {
	language := MakeLanguage()
	language.On("platform", newStub(&platform.Platform{}))
	language.On(RefFuncName, newStub(""))

	// params
	language.On("param", newStub(""))
	language.On("paramList", newStub([]interface{}{}))
	language.On("paramMap", newStub(map[string]interface{}{}))
	return language
}

// DefaultLanguage provides a default language extension.  It creates default
// implementations of context-free and non-dependency-generating functions
// (e.g. split) and provides a unimplemented function for functions that must be
// supplied with context or which may register dependencies.
func DefaultLanguage() *LanguageExtension {
	language := MakeLanguage()
	language.On("env", DefaultEnv)
	language.On("split", DefaultSplit)
	language.On("join", DefaultJoin)
	language.On("jsonify", DefaultJsonify)
	language.On("platform", platform.DefaultPlatform)
	language.On(RefFuncName, Unimplemented(RefFuncName))

	// params
	language.On("param", Unimplemented("param"))
	language.On("paramList", Unimplemented("paramList"))
	language.On("paramMap", Unimplemented("paramMap"))
	language.Validate()
	return language
}

// On provides a mechanism for defining an activity that will take place on
// encountering a keyword.  It inserts the key and value pair into the language
// and returns a reference to the language.  The language is mutated and the
// returned version is simply to allow method chaning, e.g.:
//   language = MakeLanguage().On("foo", foo).On("bar", bar).On("baz", baz)
func (l *LanguageExtension) On(keyword string, action interface{}) *LanguageExtension {
	l.innerLock.Lock()
	defer l.innerLock.Unlock()

	l.Funcs[keyword] = action
	return l
}

// Join adds the keywords from toAdd that do not exist in l and adds them
func (l *LanguageExtension) Join(toAdd *LanguageExtension) *LanguageExtension {
	l.innerLock.Lock()
	defer l.innerLock.Unlock()

	for keyword, f := range toAdd.Funcs {
		if _, found := l.Funcs[keyword]; !found {
			l.Funcs[keyword] = f
		}
	}
	return l
}

// Validate checks the defined language against the known keywords and returns
// the deltas, if any.  It returns true if the language exactly matches the
// known keyword list and false, with deltas, otherwise.
func (l *LanguageExtension) Validate() (missingKeywords []string, extraKeywords []string, valid bool) {
	l.innerLock.Lock()
	defer l.innerLock.Unlock()

	var missing []string
	var extra []string
	ok := true
	for key := range l.Funcs {
		if _, found := languageKeywords[key]; !found {
			extra = append(extra, key)
			ok = false
		}
	}
	for key := range languageKeywords {
		if _, found := l.Funcs[key]; !found {
			missing = append(missing, key)
			ok = false
		}
	}
	if !ok {
		log.Printf("[WARN] bad template DSL: extra keywords: %v, missing: %v\n",
			extra,
			missing,
		)
	}
	return missing, extra, ok
}

// Render provides a lightweight interface over template.New and
// template.Execute, it creates a new template given the name and input string,
// renders it with the currently defined language extensions, and writes the
// output into the provided io.Writer.  If any error is returned at any point it
// is passed on to the user.
func (l *LanguageExtension) Render(dotValues interface{}, name, toRender string) (bytes.Buffer, error) {
	l.innerLock.Lock()
	defer l.innerLock.Unlock()
	var output bytes.Buffer
	tmpl, err := template.New(name).Funcs(l.Funcs).Parse(toRender)
	if err != nil {
		return output, err
	}
	err = tmpl.Execute(&output, dotValues)
	return output, err
}

// StubTemplateFunc is the NOP function for template parsing
func StubTemplateFunc(...string) (string, error) {
	return "", nil
}

// newStub generates a stub function that always returns returnVal when called,
// and supports a variadic number of arguments.  It is used to generate stubs
// that need to return a specific value or real data type (e.g. stubs for
// `platform` which must return a valid `*platform.Platform` to prevent template
// execution errors).
func newStub(returnVal interface{}) func(...string) (interface{}, error) {
	return func(...string) (interface{}, error) {
		return returnVal, nil
	}
}

// RememberCalls is a utility function to instert calls into a list.
// RememberCalls takes a pointer to a list of strings, and an argument
// index.  It returns a variadic function that when called from gotemplate will
// take the indexed argument and append it to the provided list.
func RememberCalls(list *[]string, returnvalue interface{}) interface{} {
	return func(params ...string) (interface{}, error) {
		name := params[0]
		*list = append(*list, name)
		return returnvalue, nil
	}
}

// Unimplemented returns a function that will raise an error with the fact that
// the keyword is unimplemented.
func Unimplemented(name string) interface{} {
	return func(params ...string) (string, error) {
		return "", fmt.Errorf("%s is unimplemented in the current template language", name)
	}
}
