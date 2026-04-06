package cmd

import (
	"os"
	"os/exec"
	"testing"

	"github.com/go-openapi/testify/v2/assert"
)

func TestRootCmd(t *testing.T) {
	assert.NotEmpty(t, rootCmd.Version)
	assert.NotNil(t, rootCmd.RunE)
}

// this test is explicitly run in a separate process (with TEST_HELPER_PROCESS
// set) to catch exit codes.
func TestExecute(t *testing.T) {
	// In the helper process, just run Execute!
	if os.Getenv("TEST_HELPER_PROCESS") == "1" {
		Execute()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestExecute")
	cmd.Env = append(os.Environ(), "TEST_HELPER_PROCESS=1")

	err := cmd.Run()
	assert.Error(t, err) // need filename to succeed!
}
