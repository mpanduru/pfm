package main

import (
	"fmt"
	"os"

	"example.com/pfm/internal/app"
)

func main() {
	a := app.New()

	if err := a.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
