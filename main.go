package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/paultyng/tf-static-registry/reggen"
)

func main() {
	args := os.Args
	if len(args) != 2 {
		log.Fatal("a path is required")
	}
	path, err := filepath.Abs(args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	modulepath := filepath.Join(path, "modules")
	outpath := filepath.Join(path, "public")

	err = reggen.Generate(modulepath, outpath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
