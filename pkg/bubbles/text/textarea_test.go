package text

import (
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestTextArea(t *testing.T) {
	text := newTextArea("PLACEHOLDER")
	assert.NotNil(t, text)
	assert.Empty(t, text.Prompt)
	assert.Equal(t, "PLACEHOLDER", text.Placeholder)

	assert.Nil(t, text.Init())

	// HACKHACK: just include the raw ANSI escapes for now
	// fmt.Printf("ACTUAL: ====================\n%q\n=======================\n", text.View().Content)
	assert.Equal(t, "\x1b[38;5;244m\x1b[38;5;244m\x1b[38;5;244m\x1b[38;5;244m\x1b[m\x1b[m\x1b[38;5;244m\x1b[3;38;5;238mP\x1b[m\x1b[m\x1b[38;5;244m\x1b[3;38;5;238mLACEHOLDER\x1b[m\x1b[m\x1b[38;5;244m                       \x1b[m      \x1b[m\x1b[m\n\x1b[38;5;244m\x1b[38;5;244m\x1b[3;38;5;238m\x1b[38;5;244m\x1b[m\x1b[m\x1b[38;5;244m \x1b[m                                       \x1b[m\x1b[m\n\x1b[38;5;244m\x1b[38;5;244m\x1b[3;38;5;238m\x1b[38;5;244m\x1b[m\x1b[m\x1b[38;5;244m \x1b[m                                       \x1b[m\x1b[m\n\x1b[38;5;244m\x1b[38;5;244m\x1b[3;38;5;238m\x1b[38;5;244m\x1b[m\x1b[m\x1b[38;5;244m \x1b[m                                       \x1b[m\x1b[m\n\x1b[38;5;244m\x1b[38;5;244m\x1b[3;38;5;238m\x1b[38;5;244m\x1b[m\x1b[m\x1b[38;5;244m \x1b[m                                       \x1b[m\x1b[m\n\x1b[38;5;244m\x1b[38;5;244m\x1b[3;38;5;238m\x1b[38;5;244m\x1b[m\x1b[m\x1b[38;5;244m \x1b[m                                       \x1b[m\x1b[m\n                                        \n                                        \n                                        \n                                        ", text.View().Content)

	text2, cmd := text.Update(nil)
	assert.Equal(t, text, text2)
	assert.Nil(t, cmd)
}
