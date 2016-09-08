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

package extensions_test

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
	"text/template"

	"github.com/asteris-llc/converge/render/extensions"
	"github.com/stretchr/testify/assert"
)

var keywords = map[string]struct{}{
	"env":      {},
	"platform": {},
	"param":    {},
	"split":    {},
	"lookup":   {},
}

var contextualFunctions = map[string]string{
	"param": "{{param `foo`}}",
}

func Test_MakeLanguage_MakesEntryForEachKnownKeyword(t *testing.T) {
	language := extensions.MakeLanguage()
	funcs := takeKeys(language.Funcs)
	assert.True(
		t,
		reflect.DeepEqual(keywords, takeKeys(language.Funcs)),
		fmt.Sprintf("Comparing funcs and keywords:\n%v\n%v\n", funcs, keywords),
	)
}

func Test_DefaultLanguage_MakesAnEntryForEachKnownKeyword(t *testing.T) {
	language := extensions.DefaultLanguage()
	funcs := takeKeys(language.Funcs)
	assert.True(
		t,
		reflect.DeepEqual(keywords, takeKeys(language.Funcs)),
		fmt.Sprintf("Comparing funcs and keywords:\n%v\n%v\n", funcs, keywords),
	)
}

func Test_DefaultLanguage_SetsUnimplementedForContextualFunctions(t *testing.T) {
	l := extensions.DefaultLanguage()
	for _, example := range contextualFunctions {
		_, err := renderTemplate(l, example)
		assert.Error(t, err)
	}
}

func Test_Validate_ReturnsEmptySlicesWhenValidDSL(t *testing.T) {
	l := extensions.MakeLanguage()
	_, _, ok := l.Validate()
	assert.True(t, ok)
}

func Test_Validate_ReturnsSlicesOfMissingWhenMissingL(t *testing.T) {
	expected := []string{"env", "param", "split", "lookup", "platform"}
	l := &extensions.LanguageExtension{}
	missing, _, ok := l.Validate()
	assert.False(t, ok)

	sort.Strings(expected)
	sort.Strings(missing)
	assert.Equal(t, expected, missing)
}

func Test_Validate_ReturnsSlicesOfExtraWhenExtra(t *testing.T) {
	expected := []string{"testkeyword"}
	l := extensions.DefaultLanguage()
	l.On("testkeyword", extensions.StubTemplateFunc)
	_, extra, ok := l.Validate()
	assert.False(t, ok)
	assert.Equal(t, expected, extra)
}

func Test_DefaultEnv_EnvExists(t *testing.T) {
	os.Setenv("FOO", "1")
	expected := "1"
	actual := extensions.DefaultEnv("FOO")
	assert.Equal(t, expected, actual)
}

func Test_DefaultEnv_EnvNotFound(t *testing.T) {
	expected := ""
	actual := extensions.DefaultEnv("fake_env_var")
	assert.Equal(t, expected, actual)
}

func Test_DefaultSplit_SplitsBasedOnFirstArg(t *testing.T) {
	expected := []string{"a", "test", "list!"}
	actual := extensions.DefaultSplit("#", "a#test#list!")
	assert.True(t, reflect.DeepEqual(expected, actual))
}

// strip the values out of a map so we can use reflect.DeepEqual for comparison
func takeKeys(m template.FuncMap) map[string]struct{} {
	out := make(map[string]struct{})
	for key := range m {
		out[key] = struct{}{}
	}
	return out
}

func renderTemplate(l *extensions.LanguageExtension, s string) (string, error) {
	var buffer bytes.Buffer
	useless := struct{}{}
	tmpl, err := template.New("unit test").Funcs(l.Funcs).Parse(s)
	if err != nil {
		return "", err
	}
	err = tmpl.Execute(&buffer, &useless)
	return buffer.String(), err
}
