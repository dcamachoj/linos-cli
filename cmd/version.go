package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const versionMaj = 0
const versionMin = 1
const versionRel = 3

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Linos CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(versionString())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func versionString() string {
	return fmt.Sprintf("v%d.%d.%d", versionMaj, versionMin, versionRel)
}
