package batch

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "batch",
	Short: "Commands for managing batch operations",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(cancelCmd)
}
