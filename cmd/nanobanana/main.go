package main

import (
	"os"

	"github.com/lyalindotcom/nano-banana-cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
