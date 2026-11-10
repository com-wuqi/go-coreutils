package main

import (
	"os"

	"go-coreutils/internal/wc"
)

func main() {
	os.Exit(wc.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
