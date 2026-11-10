package main

import (
	"os"

	"go-coreutils/internal/echo"
)

func main() {
	os.Exit(echo.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
