package batch

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
	Short: "List all batches",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc := openai.NewBatchService(
			option.WithBaseURL(appctx.OpenAIAPIURLFrom(ctx)),
			option.WithAPIKey(appctx.OpenAIAPIKeyFrom(ctx)),
		)

		resp, err := svc.List(ctx, openai.BatchListParams{})
		if err != nil {
			return errors.Join(errors.New("failed to request api"), err)
		}

		items := resp.Data
		if len(items) == 0 {
			fmt.Println("no batch jobs.")
			return nil
		}
		for _, item := range items {
			fmt.Printf("[%s] %s\n", item.Status, item.ID)
			if len(item.OutputFileID) > 0 {
				fmt.Printf(" - output: %s\n", item.OutputFileID)
			}
			if len(item.ErrorFileID) > 0 {
				fmt.Printf(" - error: %s\n", item.ErrorFileID)
			}
		}

		return nil
	},
}
