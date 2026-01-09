package main

import (
	"os"

	"github.com/jslay88/zladxhd-installer/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
