package prettyprinters

import (
	"bytes"
	"fmt"
)

// Renderable provides an interface for printable objects
type Renderable interface {
	// The Renderable interface should provide an instance of String() to render
	// the object.  String() shall return the string-ified version of the object
	// regardless of the current visibility of the Renderable.
	fmt.Stringer

	// Visible returns true if the object should be rendered, and false
	// otherwise.  If a consumer chooses to ignore this value, the instance should
	// still provide a valid string value.
	Visible() bool

	Hide()
	Unhide()
}

// StringRenderable provides a Renderable wrapper around strings.
type StringRenderable struct {
	Hidden   bool
	Contents string
}

// Visible returns the embedded Hidden field
func (r *StringRenderable) Visible() bool {
	return !r.Hidden
}

// String returns the contents of the string
func (r *StringRenderable) String() string {
	return r.Contents
}

func (r *StringRenderable) Hide() {
	r.Hidden = true
}

func (r *StringRenderable) Unhide() {
	r.Hidden = false
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

func RenderableString(s string, visible bool) *StringRenderable {
	if visible {
		return VisibleString(s)
	}
	return HiddenString(s)
}

// WrappedRenderable wraps a functor-like interface around Renderable, allowing
// you to nest string transformations over the final string value without having
// to immediately evaluate the string.  This will be specifically valueable
// because we can transform a hidden string that will be made visible later.
type WrappedRenderable struct {
	baseValue Renderable
	show      func() string
}

// Visible returns the visibility of the underlying Renderable
func (w *WrappedRenderable) Visible() bool {
	return w.baseValue.Visible()
}

// String returns the result of applying show to the underlying string value
func (w *WrappedRenderable) String() string {
	return w.show()
}

func (w *WrappedRenderable) Hide() {
	w.baseValue.Hide()
}

func (w *WrappedRenderable) Unhide() {
	w.baseValue.Unhide()
}

// ApplyRenderable allows you to apply an arbitrary string transformation
func ApplyRenderable(r Renderable, f func(string) string) *WrappedRenderable {
	return &WrappedRenderable{
		baseValue: r,
		show: func() string {
			return f(r.String())
		},
	}
}

// Render is a shorthand for getting the string value of a Renderable along with
// it's visibility.  Returns a tuple of (str, visible).  The value of str is
// undefined if visible is false.
func Render(r Renderable) (string, bool) {
	return r.String(), r.Visible()
}

// SprintfRenderable is a utility function to reduce the overhead of writing
//    VisibleString(fmt.Sprintf(....))
func SprintfRenderable(visible bool, fmtStr string, args ...interface{}) *StringRenderable {
	contents := fmt.Sprintf(fmtStr, args)
	return &StringRenderable{
		Contents: contents,
		Hidden:   visible,
	}
}

// SprintfVisible is a shorthand for fmt.Sprintf and creating a Visible String
func SprintfVisible(fmtStr string, args ...interface{}) Renderable {
	return SprintfRenderable(true, fmtStr, args)
}

// SprintfHidden is a shorthand for fmt.Sprintf and creating a Hidden String
func SprintfHidden(fmtStr string, args ...interface{}) Renderable {
	return SprintfRenderable(false, fmtStr, args)
}

// writeRenderable acts like bytes.Buffer.WriteString() but appends the
// renderable string only if it's visible.
func writeRenderable(b *bytes.Buffer, r Renderable) {
	if !r.Visible() {
		return
	}
	_, _ = b.WriteString(r.String())
}

func SetVisibility(r Renderable, visible bool) Renderable {
	if visible {
		r.Unhide()
	} else {
		r.Hide()
	}
	return r
}
