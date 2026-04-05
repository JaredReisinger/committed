package teautil

import (
	"slices"

	tea "charm.land/bubbletea/v2"
)

// Type Router represents a model message router, for the very common case of a
// parent model that contains multiple child models that take focus or otherwise
// need messages passed to them, and command results re-routed back to them. I'm
// attempting to encapsulate this logic so that I don't have to re-implement it
// again and again.
//
// One tricky aspect is that parent models often care about the concrete type of
// their child models, but updates and message routing all follow the same
// pattern, *plus* a parent needs to re-assign child models from the model
// return value of Update().  The opinionated solution here is to always treat
// the child models as unknown [tea.Model] values, and force the parent to
// extract them from the Router and cast if it wants to do anything specific to
// the child model type.  (We do use a generic so that if all children are a
// more-specific type, that type is exposed.)
type Router[I ~int, T tea.Model] struct {
	firstId I
	models  []T
}

// NewRouter creates a new model message router.
func NewRouter[I ~int, T tea.Model](firstId I, models ...T) Router[I, T] {
	r := Router[I, T]{
		firstId: firstId,
		models:  slices.Clone(models),
	}

	return r
}

func (r Router[I, T]) Len() int {
	return len(r.models)
}

// should this return an error instead of ok? There's only one error condition.
// Should the returned model be a pointer or a value? We really should support
// value-based patterns.
func (r Router[I, T]) Index(i I) (*T, bool) {
	actualIndex := i - r.firstId
	if actualIndex < 0 || actualIndex >= I(len(r.models)) {
		return nil, false
	}
	return &r.models[actualIndex], true
}

func (r Router[I, T]) MustIndex(i I) *T {
	m, ok := r.Index(i)
	if !ok {
		panic("index invalid")
	}

	return m
}

// Update handles routing messages to the appropriate child model. If the
// message is a WrappedMsg, it is unwrapped and routed to that specific child.
// Otherwise, if id is >= 0, the message is passed to that specific child, and
// if id is < 0, it is sent to *all* of the children. Just as with normal
// models, the parent should assign/overwrite their Router member with the
// Router value returned from Update.
func (r Router[I, T]) Update(msg tea.Msg, id I) (Router[I, T], tea.Cmd) {
	// Should we add a fallback in case someone wraps using an int instead of a
	// known type?  Or at least an "if !ok but .(WrappedMsg[int])?"
	if wrappedMsg, ok := msg.(WrappedMsg[I]); ok {
		id = I(wrappedMsg.Id)
		msg = wrappedMsg.Msg
	}

	start := 0
	end := len(r.models)

	if id >= 0 {
		start = int(id - r.firstId)
		// REVIEW: we *can't* return an error if we want to follow the tea.Model
		// Update pattern. Perhaps we have an error member and return a
		// command/message that indicates the problem?  On the other hand, we're
		// passing `id` in, so this clearly isn't a model interface already.
		if start < 0 {
			panic("child model id out of range")
		}
		end = start + 1
	}

	var cmds []tea.Cmd

	for i := start; i < end; i++ {
		updatedChildModel, cmd := r.models[i].Update(msg)
		r.models[i] = updatedChildModel.(T)
		cmds = append(cmds, Wrap(cmd, i+int(r.firstId)))
	}

	return r, tea.Batch(cmds...)
}
