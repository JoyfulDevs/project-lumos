package file

import (
	"errors"
	"fmt"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"

	"github.com/joyfuldevs/project-lumos/cmd/openai/app/appctx"
)

var (
	uploadFile string
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a file for batch",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc := openai.NewFileService(
			option.WithBaseURL(appctx.OpenAIAPIURLFrom(ctx)),
			option.WithAPIKey(appctx.OpenAIAPIKeyFrom(ctx)),
		)

		f, err := os.Open(uploadFile)
		if err != nil {
			return errors.Join(errors.New("failed to open file"), err)
		}
		defer func() { _ = f.Close() }()

		resp, err := svc.New(ctx, openai.FileNewParams{
			File:    f,
			Purpose: openai.FilePurposeBatch,
		})
		if err != nil {
			return errors.Join(errors.New("failed to request api"), err)
		}

		fmt.Printf("%s: [%s] %s\n", resp.ID, resp.Purpose, resp.Filename)

		return nil
	},
}

func init() {
	uploadCmd.Flags().StringVarP(&uploadFile, "file", "f", "", "file path to upload")

	_ = uploadCmd.MarkFlagRequired("file")
}
