package main

import (
	"os"

	"github.com/breca/vsixdler/internal/cmd"
)

func main() {
	if err := cmd.NewRoot().Execute(); err != nil {
		os.Exit(1)
	}
}
