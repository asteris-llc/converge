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
	"io"
	"log"
	"text/template"
)

// languageKeywords defines the known keywords that have been added to the
// templating language.  This is stored as a map for quick lookup and is used
// for DSL validation.
var languageKeywords = map[string]struct{}{
	"param": {},
	"split": {},
}

// LanguageExtension is a type wrapper around a template.FuncMap to allow us to
// encapsulate more context in the future.
type LanguageExtension struct {
	Funcs template.FuncMap
}

// MakeLanguage provides an empty language that implements a nil operation for
// all known keywords.
func MakeLanguage() *LanguageExtension {
	funcs := template.FuncMap{}
	for keyword := range languageKeywords {
		funcs[keyword] = DoNothing()
	}
	return &LanguageExtension{Funcs: funcs}
}

// DefaultLanguage provides a default language extension.  It creates default
// implementations of context-free and non-dependency-generating functions
// (e.g. split) and provides a unimplemented function for functions that must be
// supplied with context or which may register dependencies.
func DefaultLanguage() *LanguageExtension {
	language := MakeLanguage()
	language.On("split", DefaultSplit)
	language.On("param", Unimplimented("param"))
	language.Validate()
	return language
}

// On provides a mechanism for defining an activity that will take place on
// encountering a keyword.  It inserts the key and value pair into the language
// and returns a reference to the language.  The language is mutated and the
// returned version is simply to allow method chaning, e.g.:
//   language = MakeLanguage().On("foo", foo).On("bar", bar).On("baz", baz)
func (l *LanguageExtension) On(keyword string, action interface{}) *LanguageExtension {
	l.Funcs[keyword] = action
	return l
}

// Validate checks the defined language against the known keywords and returns
// the deltas, if any.  It returns true if the language exactly matches the
// known keyword list and false, with deltas, otherwise.
func (l *LanguageExtension) Validate() (missingKeywords []string, extraKeywords []string, valid bool) {
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
func (l *LanguageExtension) Render(dotValues interface{}, output io.Writer, name string, toRender string) error {
	tmpl, err := template.New(name).Funcs(l.Funcs).Parse(toRender)
	if err != nil {
		return err
	}
	return tmpl.Execute(output, dotValues)
}

// SimpleRender provides a simple interface to render a template with language
// extensions assuming that there are no dotValues and provides an empty stub
// struct.
func (l *LanguageExtension) SimpleRender(output io.Writer, name, toRender string) error {
	stubObj := struct{}{}
	return l.Render(&stubObj, output, name, toRender)
}

// SimpleRenderStr provides the lightest weight wrapper around template
// rendering, assuming there are not struct values that need to be accessed and
// rendering into an ephemeral byte buffer that is returned as a string.
func (l *LanguageExtension) SimpleRenderStr(name, toRender string) (string, error) {
	var buffer bytes.Buffer
	err := l.SimpleRender(&buffer, name, toRender)
	return buffer.String(), err
}

// DoNothing returns a function that stubs a template function, returning an
// empty string and no error.
func DoNothing() interface{} {
	return func(params ...string) (string, error) {
		return "", nil
	}
}

// RegisterExistence is a utility function to instert calls into a list.
// RegisterExistence takes a pointer to a list of strings, and an argument
// index.  It returns a variadic function that when called from gotemplate will
// take the indexed argument and append it to the provided list.
func RegisterExistence(list *[]string, nameIndex int) interface{} {
	return func(params ...string) (string, error) {
		name := params[0]
		*list = append(*list, "param."+name)
		return name, nil
	}
}

// Unimplimented returns a function that will raise an error with the fact that
// the keyword is unimplemented.
func Unimplimented(name string) interface{} {
	return func(params ...string) (string, error) {
		return "", fmt.Errorf("%s is unimplimented in the current template language", name)
	}
}
