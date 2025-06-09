package main

import (
	"fmt"
	"os"

	"github.com/hexabase/hexabase-ai/cli/hks-func/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}