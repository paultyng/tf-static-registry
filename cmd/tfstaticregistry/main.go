package main

import (
	"os"

	"github.com/mattn/go-colorable"

	"github.com/paultyng/terraform-static-registry/internal/cmd"
)

func main() {
	name := "tfstaticregistry"
	version := name + " Version " + version
	if commit != "" {
		version += " from commit " + commit
	}

	os.Exit(cmd.Run(
		name,
		version,
		os.Args[1:],
		os.Stdin,
		colorable.NewColorableStdout(),
		colorable.NewColorableStderr(),
	))
}
