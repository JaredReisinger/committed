package tui

import (
	"testing"

	"charm.land/bubbles/v2/key"
	"github.com/go-openapi/testify/v2/assert"
)

func TestKeyMap(t *testing.T) {
	dummyBindings := []key.Binding{
		key.NewBinding(),
	}

	keymap := buildHelpKeys(dummyBindings)
	assert.NotNil(t, keymap)

	// short help should combine default and dummy:
	assert.Equal(t,
		len(defaultKeyMap.ShortHelp())+len(dummyBindings),
		len(keymap.ShortHelp()),
	)
}
