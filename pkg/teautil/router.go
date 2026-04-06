package teautil

import (
	"iter"
	"maps"

	tea "charm.land/bubbletea/v2"
)

// Router represents a model message router, for the very common case of a
// parent model that contains multiple child models that take focus or otherwise
// need messages passed to them, and command results re-routed back to them. I'm
// attempting to encapsulate this logic so that I don't have to re-implement it
// again and again.
//
// One tricky aspect is that parent models often care about the concrete type of
// their child models, but updates and message routing all follow the same
// pattern, *plus* a parent needs to re-assign child models from the model
// return value of Update().  The opinionated solution here is to strongly
// recommend using concrete [tea.Model]-implementing types, so that the parent
// does not need to perform any type-casting to use them. The corollary to this
// is that with Router thus a fully idempotent/immutable object, the caller must
// also be careful to perform mutating operations in "the correct way": call
// [Router.Set] for any change to a child model, save the returned [Router] into
// itself, *and return itself from that function so that callers can also
// acquire the new immutable object*.
type Router[K comparable, M tea.Model] struct {
	models          map[K]M
	useBroadcastKey bool
	broadcastKey    K
}

// NewRouter creates a new model message router.
func NewRouter[K comparable, M tea.Model](models map[K]M, options ...RouterOption[K, M]) Router[K, M] {
	r := Router[K, M]{
		models: maps.Clone(models),
	}

	for _, opt := range options {
		opt(&r)
	}

	return r
}

// RouterOption is the functional option type to adjust the configuration of a
// [Router] when it is created.
type RouterOption[K comparable, M tea.Model] func(r *Router[K, M])

// WithBroadcastKey defines a key value used to broadcast a message to *all*
// child models when passed as the key to [Router.Update]. If set, the broadcast
// key is considered invalid as the key for a specific child model. Even without
// a broadcast key explicitly set, you can use [Router.UpdateAll] to send a
// message to all child models.
func WithBroadcastKey[K comparable, M tea.Model](key K) RouterOption[K, M] {
	return func(r *Router[K, M]) {
		r.useBroadcastKey = true
		r.broadcastKey = key

		if _, ok := r.models[key]; ok {
			panic("broadcast key is already in use")
		}
	}
}

// Len returns the number of child models.
func (r Router[K, M]) Len() int {
	return len(r.models)
}

// All returns every key/model pair in a 2-value [iter.Seq2] in a random order.
func (r Router[K, M]) All() iter.Seq2[K, M] {
	return maps.All(r.models)
}

// Keys returns all of the keys as a [iter.Seq] in a random order.
func (r Router[K, M]) Keys() iter.Seq[K] {
	return maps.Keys(r.models)
}

// Keys returns all of the child models as a [iter.Seq] in a random order.
func (r Router[K, M]) Values() iter.Seq[M] {
	return maps.Values(r.models)
}

// Get returns the model with the given key, or (nil,false) if the key does not exist.
func (r Router[K, M]) Get(key K) (M, bool) {
	m, ok := r.models[key]
	return m, ok
}

// MustGet returns the model with the given key, panicking if it does not exist.
func (r Router[K, M]) MustGet(key K) M {
	m, ok := r.Get(key)
	if !ok {
		panic("index invalid")
	}

	return m
}

// immutability helper
func (r Router[K, M]) clone() Router[K, M] {
	return Router[K, M]{
		models:          maps.Clone(r.models),
		useBroadcastKey: r.useBroadcastKey,
		broadcastKey:    r.broadcastKey,
	}
}

// Set sets a new or changed child model into the router. Note that to support
// object immutability, a new [Router] object is returned.
func (r Router[K, M]) Set(key K, model M) Router[K, M] {
	rr := r.clone()
	rr.models[key] = model
	return rr
}

// Update handles routing messages to the appropriate child model. If the
// message is a [WrappedMsg], it is unwrapped and routed to that specific child
// (if if the broadcast key is passed as the key). Otherwise, if the key is the
// router's broadcast key, the message is sent to all child models. If the key
// is not the broadcast key, the message is either sent to the child model with
// the matching key or silently dropped if no such child model exists.
//
// Note that any [tea.Cmd] commands returned from individual child models are
// automatically wrapped to return a [WrappedMsg] so that it will be routed
// directly back to that child.
func (r Router[K, M]) Update(msg tea.Msg, key K) (Router[K, M], tea.Cmd) {
	// Should we add a fallback in case someone wraps using an int instead of a
	// known type?  Or at least an "if !ok but .(WrappedMsg[int])?"
	var keys iter.Seq[K]

	if wrappedMsg, ok := msg.(WrappedMsg[K]); ok {
		keys = singleSeq(wrappedMsg.Key)
		msg = wrappedMsg.Msg
	} else if r.useBroadcastKey && key == r.broadcastKey {
		keys = r.Keys()
	} else {
		keys = singleSeq(key)
	}

	return r.update(msg, keys)
}

// UpdateAll  routes messages to the every child model. The message is *not*
// checked to see if it is a [WrappedMsg], it is sent as-is.
//
// Note that any [tea.Cmd] commands returned from individual child models are
// automatically wrapped to return a [WrappedMsg] so that it will be routed
// directly back to that child.
func (r Router[K, M]) UpdateAll(msg tea.Msg) (Router[K, M], tea.Cmd) {
	return r.update(msg, r.Keys())
}

func (r Router[K, M]) update(msg tea.Msg, keys iter.Seq[K]) (Router[K, M], tea.Cmd) {
	var cmds []tea.Cmd

	for k := range keys {
		updatedChildModel, cmd := r.models[k].Update(msg)
		r.models[k] = updatedChildModel.(M)
		cmds = append(cmds, Wrap(cmd, k))
	}

	return r, tea.Batch(cmds...)
}

func singleSeq[K comparable](key K) iter.Seq[K] {
	return func(yield func(K) bool) {
		yield(key)
	}
}
