package teautil

import (
	"fmt"
	"reflect"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/go-openapi/testify/v2/assert"
)

type bogusMsg struct{}

func bogusCmd() tea.Msg {
	return bogusMsg{}
}

func nonBatchSliceCmd() tea.Msg {
	return []string{"test non-batch slice message"}
}

func TestWrap(t *testing.T) {
	cases := []struct {
		cmd               tea.Cmd
		id                int
		expectedMsgType   reflect.Type
		expectedCallCount int
	}{
		{nil, 0, nil, 0},
		{bogusCmd, 1, reflect.TypeFor[bogusMsg](), 1},
		{tea.Batch(bogusCmd, bogusCmd), 2, reflect.TypeFor[bogusMsg](), 2},
		{tea.Sequence(bogusCmd, bogusCmd), 3, reflect.TypeFor[bogusMsg](), 2},
		{nonBatchSliceCmd, 4, reflect.TypeFor[[]string](), 1},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			wrappedCmd := Wrap(c.cmd, c.id)
			actualCallCount := 0

			handler := func(msg tea.Msg) {
				actualCallCount++
				wrapper, ok := msg.(WrappedMsg[int])
				assert.True(t, ok)
				assert.Equal(t, c.id, wrapper.Key)
				assert.True(t, reflect.ValueOf(wrapper.Msg).Type().AssignableTo(c.expectedMsgType))
			}

			mockEval(t, wrappedCmd, handler)

			assert.Equal(t, c.expectedCallCount, actualCallCount)

		})
	}
}

type msgHandler func(msg tea.Msg)

func mockEval(t *testing.T, cmd tea.Cmd, handler msgHandler) {
	if cmd == nil {
		return
	}

	msg := cmd()

	// tea.BatchMsg and tea.sequenceMsg (not exported) should both be
	// convertible to []tea.Cmd... but the only way to manage this is with
	// reflection since tea.sequenceMsg isn't exported!
	v := reflect.ValueOf(msg)
	if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			innerCmd, ok := v.Index(i).Interface().(tea.Cmd)
			assert.True(t, ok)
			handler(innerCmd())
		}
		return
	}

	handler(msg)
}
