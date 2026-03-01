package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/sota-io/sota-cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		red := color.New(color.FgRed, color.Bold)
		fmt.Fprintf(os.Stderr, "%s %v\n", red.Sprint("Error:"), err)
		os.Exit(1)
	}
}
