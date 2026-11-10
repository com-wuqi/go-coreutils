package main

import (
	"os"

	"go-coreutils/internal/tee"
)

func main() {
	os.Exit(tee.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
