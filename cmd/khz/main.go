package main

import (
	"os"

	"github.com/Grar00t/khz-cli/internal/app"
)

func main() {
	os.Exit(app.Run(os.Args[1:], os.Stdout, os.Stderr))
}
