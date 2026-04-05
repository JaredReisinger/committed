package teautil

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/go-openapi/testify/v2/assert"
)

func TestNewRouter(t *testing.T) {
	r := NewRouter(dummyIndex(2), dummyModel{})
	assert.Equal(t, 1, r.Len())

	m, ok := r.Index(0)
	assert.False(t, ok)
	assert.Nil(t, m)

	m, ok = r.Index(1)
	assert.False(t, ok)
	assert.Nil(t, m)

	m, ok = r.Index(2)
	assert.True(t, ok)
	assert.NotNil(t, m)

	m, ok = r.Index(3)
	assert.False(t, ok)
	assert.Nil(t, m)
}

func TestRouterSingleUpdate(t *testing.T) {
	r := NewRouter(1, dummyModel{}, dummyModel{})

	child1, _ := r.Index(1)
	child2, _ := r.Index(2)
	assert.Equal(t, 0, child1.counter)
	assert.Equal(t, 0, child2.counter)

	rr, cmd := r.Update(bogusMsg{}, 1)

	child1, _ = rr.Index(1)
	child2, _ = rr.Index(2)
	assert.Equal(t, 1, child1.counter)
	assert.Equal(t, 0, child2.counter)

	wrappedMsg, ok := (cmd()).(WrappedMsg[int])
	assert.True(t, ok)
	assert.Equal(t, 1, wrappedMsg.Id)
}

func TestRouterRoutedUpdate(t *testing.T) {
	r := NewRouter(1, dummyModel{}, dummyModel{})

	child1, _ := r.Index(1)
	child2, _ := r.Index(2)
	assert.Equal(t, 0, child1.counter)
	assert.Equal(t, 0, child2.counter)

	rr, cmd := r.Update(WrappedMsg[int]{2, bogusMsg{}}, 1)

	child1, _ = rr.Index(1)
	child2, _ = rr.Index(2)
	assert.Equal(t, 0, child1.counter)
	assert.Equal(t, 1, child2.counter)

	wrappedMsg, ok := (cmd()).(WrappedMsg[int])
	assert.True(t, ok)
	assert.Equal(t, 2, wrappedMsg.Id)
}

func TestRouterAllUpdate(t *testing.T) {
	r := NewRouter(1, dummyModel{}, dummyModel{})

	child1, _ := r.Index(1)
	child2, _ := r.Index(2)
	assert.Equal(t, 0, child1.counter)
	assert.Equal(t, 0, child2.counter)

	rr, cmd := r.Update(bogusMsg{}, -1)

	child1, _ = rr.Index(1)
	child2, _ = rr.Index(2)
	assert.Equal(t, 1, child1.counter)
	assert.Equal(t, 1, child2.counter)

	_, ok := (cmd()).(tea.BatchMsg)
	assert.True(t, ok)
}

type dummyIndex int

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
