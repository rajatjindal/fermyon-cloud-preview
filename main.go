package main

import (
	"fmt"
	"os"

	"github.com/rajatjindal/fermyon-cloud-preview/cmd"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func execute() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
