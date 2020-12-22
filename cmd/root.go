package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:          "linos-cli",
	Short:        "CLI for Linos Systems",
	Long:         "CLI for Linos Systems\n" + versionString(),
	SilenceUsage: true,
}

// Execute method
func Execute() {
	var err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
