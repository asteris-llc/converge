package extensions_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"text/template"

	"github.com/asteris-llc/converge/render/extensions"
	"github.com/stretchr/testify/assert"
)

var keywords = map[string]struct{}{
	"param": {},
	"split": {},
}

var contextualFunctions = map[string]string{
	"param": "{{param `foo`}}",
}

func Test_MakeLanguage_MakesEntryForEachKnownKeyword(t *testing.T) {
	language := extensions.MakeLanguage()
	funcs := takeKeys(language.Funcs)
	fmt.Println("Comparing funcs and keywords:")
	fmt.Println(funcs)
	fmt.Println(keywords)
	assert.True(t, reflect.DeepEqual(keywords, takeKeys(language.Funcs)))
}

func Test_DefaultLanguage_MakesAnEntryForEachKnownKeyword(t *testing.T) {
	language := extensions.DefaultLanguage()
	funcs := takeKeys(language.Funcs)
	fmt.Println("Comparing funcs and keywords:")
	fmt.Println(funcs)
	fmt.Println(keywords)
	assert.True(t, reflect.DeepEqual(keywords, takeKeys(language.Funcs)))
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
	expected := []string{"param", "split"}
	l := &extensions.LanguageExtension{}
	missing, _, ok := l.Validate()
	assert.False(t, ok)
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
