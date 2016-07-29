package prettyprinters

import (
	"bytes"
	"fmt"
)

// Renderable provides an interface for printable objects
type Renderable interface {
	fmt.Stringer
	Visible() bool
}

// StringRenderable provides a Renderable wrapper around strings.
type StringRenderable struct {
	Hidden   bool
	Contents string
}

// Visible returns the embedded Hidden field
func (r *StringRenderable) Visible() bool {
	return r.Hidden
}

// GoStringer returns the contents of the string; if the value is hidden then
// return an empty string (although this should not happen)
func (r *StringRenderable) String() string {
	if r.Hidden {
		return ""
	}
	return r.Contents
}

// VisibleString creates a new StringRenderable that is visible
func VisibleString(s string) *StringRenderable {
	return &StringRenderable{
		Hidden:   false,
		Contents: s,
	}
}

// HiddenString creates a non-renderable string
func HiddenString(s string) *StringRenderable {
	return &StringRenderable{
		Hidden:   true,
		Contents: s,
	}
}

// WrappedRenderable wraps a functor-like interface around Renderable, allowing
// you to nest string transformations over the final string value without having
// to immediately evaluate the string.  This will be specifically valueable
// because we can transform a hidden string that will be made visible later.
type WrappedRenderable struct {
	baseValue Renderable
	show      func(string) string
}

// Visible returns the visibility of the underlying Renderable
func (w *WrappedRenderable) Visible() bool {
	return w.baseValue.Visible()
}

// String returns the result of applying show to the underlying string value
func (w *WrappedRenderable) String() string {
	return w.show(w.baseValue.String())
}

// ApplyRenderable allows you to apply an arbitrary string transformation
func ApplyRenderable(r Renderable, f func(string) string) *WrappedRenderable {
	switch r.(type) {
	case *WrappedRenderable:
		return &WrappedRenderable{
			baseValue: r.(*WrappedRenderable).baseValue,
			show: func(s string) string {
				return f(r.(*WrappedRenderable).show(s))
			},
		}
	default:
		return &WrappedRenderable{
			baseValue: r,
			show: func(s string) string {
				return f(s)
			},
		}
	}
}

// Render is a shorthand for getting the string value of a Renderable along with
// it's visibility.  Returns a tuple of (str, visible).  The value of str is
// undefined if visible is false.
func Render(r Renderable) (string, bool) {
	return r.String(), r.Visible()
}

// writeRenderable acts like bytes.Buffer.WriteString() but appends the
// renderable string only if it's visible.
func writeRenderable(b bytes.Buffer, r Renderable) {
	if r.Visible() {
		return
	}
	_, _ = b.WriteString(r.String())
}
