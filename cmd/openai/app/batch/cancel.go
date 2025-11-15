package batch

import (
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"

	"github.com/joyfuldevs/project-lumos/cmd/openai/app/appctx"
)

var (
	batchID string
)

var cancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a batch",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc := openai.NewBatchService(
			option.WithBaseURL(appctx.OpenAIAPIURLFrom(ctx)),
			option.WithAPIKey(appctx.OpenAIAPIKeyFrom(ctx)),
		)

		resp, err := svc.Cancel(ctx, batchID)
		if err != nil {
			return errors.Join(errors.New("failed to request api"), err)
		}

		fmt.Printf("[%s] %s\n", resp.Status, resp.ID)

		return nil
	},
}

func init() {
	cancelCmd.Flags().StringVar(&batchID, "id", "", "batch id to cancel")

	_ = cancelCmd.MarkFlagRequired("id")
}
