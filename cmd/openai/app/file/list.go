package file

import (
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"

	"github.com/joyfuldevs/project-lumos/cmd/openai/app/appctx"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List uploaded files",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc := openai.NewFileService(
			option.WithBaseURL(appctx.OpenAIAPIURLFrom(ctx)),
			option.WithAPIKey(appctx.OpenAIAPIKeyFrom(ctx)),
		)

		resp, err := svc.List(ctx, openai.FileListParams{})
		if err != nil {
			return errors.Join(errors.New("failed to request api"), err)
		}

		items := resp.Data
		if len(items) == 0 {
			fmt.Println("no upload files.")
		}
		for _, item := range items {
			fmt.Printf("[%s] %s / %s\n", item.Purpose, item.ID, item.Filename)
		}

		return nil
	},
}
