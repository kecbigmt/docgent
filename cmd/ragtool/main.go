package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"golang.org/x/oauth2/google"

	"docgent-backend/cmd/ragtool/cli"
	"docgent-backend/internal/infrastructure/google/vertexai/rag/lib"
)

var CLI cli.CLI

func main() {
	ctx := context.Background()

	kongCtx := kong.Parse(&CLI)

	credentials, err := google.FindDefaultCredentials(ctx)
	if err != nil {
		fmt.Printf("Failed to get Google Cloud credentials: %v\n", err)
		os.Exit(1)
	}

	client := lib.NewClientWithCredentials(credentials, CLI.ProjectID, CLI.Location)

	var cmdErr error
	switch kongCtx.Command() {
	case "file upload":
		cmdErr = cli.HandleFileUpload(ctx, &CLI, client)
	case "file delete":
		cmdErr = cli.HandleFileDelete(ctx, &CLI, client)
	case "corpus create":
		cmdErr = cli.HandleCorpusCreate(ctx, &CLI, client)
	case "corpus list":
		cmdErr = cli.HandleCorpusList(ctx, &CLI, client)
	}

	if cmdErr != nil {
		fmt.Println(cmdErr)
		os.Exit(1)
	}
}
