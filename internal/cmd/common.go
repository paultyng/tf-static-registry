package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
)

type commonCmd struct {
	ui cli.Ui
}

func (cmd *commonCmd) run(r func() error) int {
	err := r()
	if err != nil {
		// TODO: unwraps? check for special exit code error?
		cmd.ui.Error(fmt.Sprintf("Error executing command: %s\n", err))
		os.Exit(1)
	}
	return 0
}
