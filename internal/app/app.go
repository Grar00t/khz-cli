package app

import (
	"fmt"
	"io"
)

const Version = "0.1.0-dev"

func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printHelp(stdout)
		return 0
	}

	switch args[0] {
	case "version", "--version", "-v":
		fmt.Fprintf(stdout, "khz %s\n", Version)
		return 0
	case "help", "--help", "-h":
		printHelp(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "khz: unknown command %q\n", args[0])
		fmt.Fprintln(stderr, "Run 'khz help' for usage.")
		return 2
	}
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "KHZ — Human Decision CLI")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  khz version")
	fmt.Fprintln(w, "  khz help")
}
