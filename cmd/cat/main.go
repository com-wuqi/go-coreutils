package main

import (
	"os"

	"go-coreutils/internal/cat"
)

func main() {
	os.Exit(cat.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
