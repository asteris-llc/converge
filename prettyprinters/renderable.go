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
	return !r.Hidden
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

// ApplyRenderable allows you to apply an arbitrary string transformation
func ApplyRenderable(r Renderable, f func(string) string) *WrappedRenderable {
	return &WrappedRenderable{
		baseValue: r,
		show: func() string {
			return f(r.String())
		},
	}
}

// VisibilityWrapper allows you to easily toggle visibility of a Renderer on or
// off. Using this to enable visibility on a Renderable that is invisible may
// result in undefined behavior if String() is undefined on the underlying
// Renderable when it is not visible.
type VisibilityWrapper struct {
	baseValue        Renderable
	visibilityToggle *bool
}

// Visible will return True or False depending on the wrappers configured
// setting, or if none is provided, will default to the underlying Renderable's
// visibility settings.
func (v *VisibilityWrapper) Visible() bool {
	if v.visibilityToggle == nil {
		return v.baseValue.Visible()
	}
	return *v.visibilityToggle
}

// String will return the underlying Renderable's string value.  This may not be
// defined if the underlying Renderable string is not defined (e.g. when using
// the wrapper to force visibility)
func (v *VisibilityWrapper) String() string {
	return v.baseValue.String()
}

// ToggleVisible toggles the visibility settings for a Renderable.
func ToggleVisible(r Renderable, visible bool) *VisibilityWrapper {
	return &VisibilityWrapper{
		baseValue:        r,
		visibilityToggle: &visible,
	}
}

// Hide sets the Renderable to invisible
func Hide(r Renderable) *VisibilityWrapper {
	return ToggleVisible(r, false)
}

// Unhide sets the renderable to visible
func Unhide(r Renderable) *VisibilityWrapper {
	return ToggleVisible(r, true)
}

// Untoggle removes all layers of toggling and returns the underlying renderable.
func Untoggle(r Renderable) Renderable {
	switch r.(type) {
	case *VisibilityWrapper:
		return Untoggle(r.(*VisibilityWrapper).baseValue)
	default:
		return r
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
func writeRenderable(b bytes.Buffer, r Renderable) {
	if r.Visible() {
		return
	}
	_, _ = b.WriteString(r.String())
}
