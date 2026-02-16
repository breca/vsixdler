package cmd

import (
	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "vsixdler",
		Short: "Download VS Code extensions from the marketplace",
	}

	root.AddCommand(newDownloadCmd())

	return root
}
