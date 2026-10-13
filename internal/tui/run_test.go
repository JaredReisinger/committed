package tui

import (
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/assert"

	"github.com/jaredreisinger/committed/pkg/config"
)

func TestRun(t *testing.T) {
	// t.Skip("getting a `makeslice: len out of range` in textinput")

	cfg := config.DefaultConfig()

	// start a goroutine to wait for testHookProgram to have a value and then
	// wait a few seconds before quitting the program.
	go func() {
		for range 5 {
			if programTestHook == nil {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// We can see the program!
			time.Sleep(200 * time.Millisecond)

			// Fancy would be to navigate and create a specific message...
			programTestHook.Quit()
			return
		}

		assert.Fail(t, "program not found")
	}()

	msg, err := RunCommitForm(cfg, nil)
	assert.NoError(t, err)
	assert.Equal(t, "build: \n", msg.Format(cfg))
}
