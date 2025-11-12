package file

import (
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"

	"github.com/joyfuldevs/project-lumos/cmd/openai/app/appctx"
)

var (
	deleteID string
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete uploaded file",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc := openai.NewFileService(
			option.WithBaseURL(appctx.OpenAIAPIURLFrom(ctx)),
			option.WithAPIKey(appctx.OpenAIAPIKeyFrom(ctx)),
		)

		resp, err := svc.Delete(ctx, deleteID)
		if err != nil {
			return errors.Join(errors.New("failed to request api"), err)
		}

		fmt.Printf("%s: %v\n", resp.ID, resp.Deleted)

		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVar(&deleteID, "id", "", "file id to delete")

	_ = deleteCmd.MarkFlagRequired("id")
}
