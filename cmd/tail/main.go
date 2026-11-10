package main

import (
	"os"

	"go-coreutils/internal/tail"
)

func main() {
	os.Exit(tail.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
