package teautil

import (
	"iter"
	"maps"
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/go-openapi/testify/v2/assert"
)

func TestNewRouter(t *testing.T) {
	r := NewRouter(map[int]dummyModel{2: {}})
	assert.Equal(t, 1, r.Len())

	m, ok := r.Get(0)
	assert.False(t, ok)

	m, ok = r.Get(1)
	assert.False(t, ok)

	m, ok = r.Get(2)
	assert.True(t, ok)
	assert.NotNil(t, m)

	m, ok = r.Get(3)
	assert.False(t, ok)
}

func TestRouterMapFuncs(t *testing.T) {
	r := NewRouter(map[int]dummyModel{
		1: {},
		3: {99},
	})

	assert.Equal(t, 2, r.Len())
	assert.Equal(t, []int{1, 3}, slices.Sorted(r.Keys()))

	saw := map[int]bool{}
	it := r.All()
	next, stop := iter.Pull2(it)

	k, _, ok := next()
	assert.True(t, ok)
	saw[k] = true

	k, _, ok = next()
	assert.True(t, ok)
	saw[k] = true

	_, _, ok = next()
	assert.False(t, ok)

	stop()

	assert.Equal(t, []int{1, 3}, slices.Sorted(maps.Keys(saw)))

	assert.Equal(t,
		[]dummyModel{{0}, {99}},
		slices.SortedFunc(r.Values(), func(a, b dummyModel) int {
			return a.counter - b.counter
		}),
	)
}

func TestRouterSet(t *testing.T) {
	r := NewRouter(map[int]dummyModel{2: {}})
	assert.Equal(t, 0, r.MustGet(2).counter)

	rr := r.Set(2, dummyModel{3})
	assert.NotEqual(t, r, rr)
	assert.Equal(t, 3, rr.MustGet(2).counter)

	assert.Equal(t, 0, r.MustGet(2).counter)
}

func TestRouterSingleUpdate(t *testing.T) {
	r := NewRouter(map[int]dummyModel{
		1: {},
		2: {},
	})

	child1 := r.MustGet(1)
	child2 := r.MustGet(2)
	assert.Equal(t, 0, child1.counter)
	assert.Equal(t, 0, child2.counter)

	rr, cmd := r.Update(bogusMsg{}, 1)

	child1 = rr.MustGet(1)
	child2 = rr.MustGet(2)
	assert.Equal(t, 1, child1.counter)
	assert.Equal(t, 0, child2.counter)

	wrappedMsg, ok := (cmd()).(WrappedMsg[int])
	assert.True(t, ok)
	assert.Equal(t, 1, wrappedMsg.Key)
}

func TestRouterRoutedUpdate(t *testing.T) {
	r := NewRouter(map[int]dummyModel{
		1: {},
		2: {},
	})

	child1 := r.MustGet(1)
	child2 := r.MustGet(2)
	assert.Equal(t, 0, child1.counter)
	assert.Equal(t, 0, child2.counter)

	rr, cmd := r.Update(WrappedMsg[int]{2, bogusMsg{}}, 1)

	child1 = rr.MustGet(1)
	child2 = rr.MustGet(2)
	assert.Equal(t, 0, child1.counter)
	assert.Equal(t, 1, child2.counter)

	wrappedMsg, ok := (cmd()).(WrappedMsg[int])
	assert.True(t, ok)
	assert.Equal(t, 2, wrappedMsg.Key)
}

func TestRouterUpdateBroadcast(t *testing.T) {
	r := NewRouter(map[int]dummyModel{
		1: {},
		2: {},
	}, WithBroadcastKey[int, dummyModel](-1))

	child1 := r.MustGet(1)
	child2 := r.MustGet(2)
	assert.Equal(t, 0, child1.counter)
	assert.Equal(t, 0, child2.counter)

	rr, cmd := r.Update(bogusMsg{}, -1)

	child1 = rr.MustGet(1)
	child2 = rr.MustGet(2)
	assert.Equal(t, 1, child1.counter)
	assert.Equal(t, 1, child2.counter)

	_, ok := (cmd()).(tea.BatchMsg)
	assert.True(t, ok)
}

func TestRouterUpdateAll(t *testing.T) {
	r := NewRouter(map[int]dummyModel{
		1: {},
		2: {},
	})

	child1 := r.MustGet(1)
	child2 := r.MustGet(2)
	assert.Equal(t, 0, child1.counter)
	assert.Equal(t, 0, child2.counter)

	rr, cmd := r.UpdateAll(bogusMsg{})

	child1 = rr.MustGet(1)
	child2 = rr.MustGet(2)
	assert.Equal(t, 1, child1.counter)
	assert.Equal(t, 1, child2.counter)

	_, ok := (cmd()).(tea.BatchMsg)
	assert.True(t, ok)
}

type dummyModel struct {
	counter int
}

// do we need to route Init command?
func (m dummyModel) Init() tea.Cmd {
	return nil
}

func (m dummyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return dummyModel{counter: m.counter + 1}, bogusCmd
}

func (m dummyModel) View() tea.View {
	return tea.NewView("")
}

// examples?

func ExampleRouter_UpdateAll() {
	// Inside your parent model's Update() implementation
	p := parentModel
	ParentModel_Update := func(msg tea.Msg) (tea.Model, tea.Cmd) {
		var cmds []tea.Cmd // all commands returned from the parent's Update
		var cmd tea.Cmd

		// . . . other message handling

		p.router, cmd = p.router.UpdateAll(msg)
		cmds = append(cmds, cmd)

		// . . . other message handling

		return p, tea.Batch(cmds...)
	}

	// bubbletea would call this with...
	var someMsg tea.Msg
	ParentModel_Update(someMsg)
}

// fake model for example
type M struct{}

func (m M) Init() tea.Cmd                           { return nil }
func (m M) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m M) View() tea.View                          { return tea.NewView("") }

type ParentModel struct {
	M
	router Router[int, ChildModel]
}
type ChildModel struct{ M }

func NewParentModel() tea.Model {
	return ParentModel{
		router: NewRouter(map[int]ChildModel{
			1: {},
			2: {},
			3: {},
		}),
	}
}

var (
	parentModel ParentModel
)
