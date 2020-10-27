package cmd

import (
	"io"

	"github.com/mitchellh/cli"
)

func Run(name, version string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	var ui cli.Ui = &cli.ColoredUi{
		ErrorColor: cli.UiColorRed,
		WarnColor:  cli.UiColorYellow,

		Ui: &cli.BasicUi{
			Reader:      stdin,
			Writer:      stdout,
			ErrorWriter: stderr,
		},
	}

	commands := initCommands(ui)

	cli := cli.CLI{
		Name:       name,
		Args:       args,
		Commands:   commands,
		HelpFunc:   cli.BasicHelpFunc(name),
		HelpWriter: stderr,
		Version:    version,
	}

	exitCode, err := cli.Run()
	if err != nil {
		return 1
	}
	return exitCode
}

func initCommands(ui cli.Ui) map[string]cli.CommandFactory {

	generateFactory := func() (cli.Command, error) {
		return &generateCmd{
			commonCmd: commonCmd{
				ui: ui,
			},
		}, nil
	}

	defaultFactory := func() (cli.Command, error) {
		return &defaultCmd{
			synopsis: "the generate command is run by default",
			Command: &generateCmd{
				commonCmd: commonCmd{
					ui: ui,
				},
			},
		}, nil
	}

	return map[string]cli.CommandFactory{
		"":         defaultFactory,
		"generate": generateFactory,
	}
}

type defaultCmd struct {
	cli.Command
	synopsis string
}

func (cmd *defaultCmd) Synopsis() string {
	return cmd.synopsis
}
