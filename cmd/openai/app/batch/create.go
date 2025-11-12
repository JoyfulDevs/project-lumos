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
	fileID string
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new batch",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc := openai.NewBatchService(
			option.WithBaseURL(appctx.OpenAIAPIURLFrom(ctx)),
			option.WithAPIKey(appctx.OpenAIAPIKeyFrom(ctx)),
		)

		resp, err := svc.New(ctx, openai.BatchNewParams{
			CompletionWindow: openai.BatchNewParamsCompletionWindow24h,
			Endpoint:         openai.BatchNewParamsEndpointV1Embeddings,
			InputFileID:      fileID,
		})
		if err != nil {
			return errors.Join(errors.New("failed to request api"), err)
		}

		fmt.Printf("[%s] %s\n", resp.Status, resp.ID)

		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&fileID, "id", "", "input file id for the batch process")

	_ = createCmd.MarkFlagRequired("id")
}
