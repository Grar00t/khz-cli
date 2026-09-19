package main

import (
	"os"

	"github.com/Grar00t/khz-cli/internal/app"
)

func main() {
	os.Exit(app.RunIO(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
