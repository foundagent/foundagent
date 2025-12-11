package main

import (
	"os"

	"github.com/foundagent/foundagent/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
