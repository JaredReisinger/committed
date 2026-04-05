package teautil

import (
	"reflect"

	tea "charm.land/bubbletea/v2"
)

// WrappedMsg represents a wrapper around the Msg returned from a Cmd. See
// [Wrap] for further details.
type WrappedMsg[I ~int] struct {
	Id  I
	Msg tea.Msg
}

// Wrap allows a parent component to accurately wrap the results of any Cmd,
// regardless of whether it's a direct Cmd or a collective one like [tea.Batch]
// or [tea.Sequence].  Note that the "wrapping" arg could also be an
// interface{}, or a generic type, or even a full `func(Cmd) Msg` (for the most
// flexibility), but using an int seems to satisfy the 80:20 rule of the
// simplest solution for the most-common case: enabling routing for
// sub-components.
func Wrap[I ~int](cmd tea.Cmd, id I) tea.Cmd {
	if cmd == nil {
		return cmd
	}

	// Since tea.Cmds are *functions*, we can't expand and wrap batched commands
	// immediately.  On the plus side, the batched command wrappers *do* simply
	// return []tea.Cmd when invoked.  This means we can implement our own
	// wrapper which calls the wrapped command and then inspect the resulting
	// tea.Msg to see if it's really a []tea.Cmd and if so, create a new
	// []tea.Cmd with each command individually wrapped.  What's worse, the "is
	// this really a []tea.Cmd-returning function" type is itself encapsulated
	// in the returned function. There's *no* information on the outer tea.Cmd
	// function for us to write a generic helper.
	return func() tea.Msg {
		msg := cmd()

		// We can detect tea.BatchMsg, but tea.sequenceMsg is *not* exported,
		// sadly, so we have to rely on reflection to detect/re-create the
		// original type.
		v := reflect.ValueOf(msg)
		switch v.Kind() {
		case reflect.Slice:
			// if the message is really a slice (of tea.Cmd), we wrap
			// the individual elements in-place
			if v.Type().ConvertibleTo(reflect.TypeFor[[]tea.Cmd]()) {
				for i := 0; i < v.Len(); i++ {
					el := v.Index(i)
					innerCmd := el.Interface().(tea.Cmd)
					el.Set(reflect.ValueOf(Wrap(innerCmd, id)))
				}

				return v.Interface()
			}

			fallthrough

		default:
			// for all other messages, we "simply" wrap the result
			return WrappedMsg[I]{Id: id, Msg: msg}
		}
	}
}
