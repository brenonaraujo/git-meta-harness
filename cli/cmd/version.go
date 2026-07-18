package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// VersionCmd creates the `gmh version` command.
func VersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the gmh version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			cyan := color.New(color.FgCyan, color.Bold)
			yellow := color.New(color.FgYellow)
			fmt.Printf("%s %s\n", cyan.Sprint("gmh"), Version)
			fmt.Printf("  %s %s\n", yellow.Sprint("commit:"), Commit)
			fmt.Printf("  %s %s\n", yellow.Sprint("built:"), Date)
		},
	}
}
