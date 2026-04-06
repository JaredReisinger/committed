package teautil

import (
	"reflect"

	tea "charm.land/bubbletea/v2"
)

// WrappedMsg represents a wrapper around the Msg returned from a Cmd. See
// [Wrap] for further details.
type WrappedMsg[K comparable] struct {
	Key K
	Msg tea.Msg
}

// Wrap allows a parent component to accurately wrap the results of any
// [tea.Cmd], regardless of whether it's a direct command or a collective one
// like [tea.Batch] or [tea.Sequence].  This function is used by [Router] to
// automatically wrap commands returned from child models in [Router.Update] and
// [Router.UpdateAll].
func Wrap[K comparable](cmd tea.Cmd, key K) tea.Cmd {
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
					el.Set(reflect.ValueOf(Wrap(innerCmd, key)))
				}

				return v.Interface()
			}

			fallthrough

		default:
			// for all other messages, we "simply" wrap the result
			return WrappedMsg[K]{Key: key, Msg: msg}
		}
	}
}
