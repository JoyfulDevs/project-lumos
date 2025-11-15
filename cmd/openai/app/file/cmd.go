package file

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:   "file",
	Short: "Commands for managing files",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(uploadCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(downloadCmd)
}
