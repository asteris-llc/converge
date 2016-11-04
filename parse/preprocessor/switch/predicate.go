package control

import (
	"fmt"
	"strings"

	"github.com/asteris-llc/converge/render/extensions"
	"github.com/pkg/errors"
)

// Predicate represents a concrete or thunked predicate value
type Predicate interface {
	IsTrue() (bool, error)
}

// BasicPredicate defines a predicate type that is an already rendered string
type BasicPredicate struct {
	predicate string
}

// IsTrue provides an implementation of IsTrue for BasicPredicate
func (b *BasicPredicate) IsTrue() (bool, error) {
	lang := extensions.DefaultLanguage()
	if b.predicate == "" {
		return false, BadPredicate(b.predicate)
	}
	template := "{{ " + b.predicate + " }}"
	result, err := lang.Render(
		struct{}{},
		"predicate evaluation",
		template,
	)
	if err != nil {
		return false, errors.Wrap(err, "case evaluation failed")
	}

	truthiness := strings.TrimSpace(strings.ToLower(result.String()))

	switch truthiness {
	case "true", "t":
		return true, nil
	case "false", "f":
		return false, nil
	}
	return false, fmt.Errorf("%s: not a valid truth value; should be one of [f false t true]", truthiness)
}

// NewBasicPredicate creates a new basic predicate
func NewBasicPredicate(s string) *BasicPredicate {
	return &BasicPredicate{predicate: s}
}
