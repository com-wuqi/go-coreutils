package main

import (
	"os"

	"go-coreutils/internal/head"
)

func main() {
	os.Exit(head.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
