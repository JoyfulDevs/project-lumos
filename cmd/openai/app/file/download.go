package file

import (
	"errors"
	"io"
	"os"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"

	"github.com/joyfuldevs/project-lumos/cmd/openai/app/appctx"
)

var (
	downloadID   string
	downloadFile string
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a file",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		svc := openai.NewFileService(
			option.WithBaseURL(appctx.OpenAIAPIURLFrom(ctx)),
			option.WithAPIKey(appctx.OpenAIAPIKeyFrom(ctx)),
		)

		resp, err := svc.Content(ctx, downloadID)
		if err != nil {
			return errors.Join(errors.New("failed to request api"), err)
		}
		defer func() { _ = resp.Body.Close() }()

		if len(downloadFile) == 0 {
			_, _ = io.Copy(os.Stdout, resp.Body)
			return nil
		}

		w, err := os.Create(downloadFile)
		if err != nil {
			return errors.Join(errors.New("failed to create file"), err)
		}
		defer func() { _ = w.Close() }()

		if _, err := io.Copy(w, resp.Body); err != nil {
			return errors.Join(errors.New("failed to write file"), err)
		}

		return nil
	},
}

func init() {
	downloadCmd.Flags().StringVar(&downloadID, "id", "", "file id to download")
	downloadCmd.Flags().StringVarP(&downloadFile, "output", "o", "", "output file path")

	_ = downloadCmd.MarkFlagRequired("id")
}
