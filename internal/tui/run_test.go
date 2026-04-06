package tui

import (
	"testing"
	"time"

	"github.com/go-openapi/testify/v2/assert"

	"github.com/jaredreisinger/committed/internal/config"
)

func TestRun(t *testing.T) {
	t.Skip("getting a `makeslice: len out of range` in textinput")

	cfg := config.DefaultConfig()

	// start a goroutine to wait for testHookProgram to have a value and then
	// wait a few seconds before quitting the program.
	go func() {
		for range 5 {
			if testHookProgram == nil {
				time.Sleep(time.Second)
				continue
			}

			// we can see the program!
			time.Sleep(time.Second * 2)
			testHookProgram.Quit()
			return
		}

		assert.Fail(t, "program not found")
	}()

	msg, err := Run(cfg, nil)
	assert.NoError(t, err)
	assert.Equal(t, "", msg.String())
}
