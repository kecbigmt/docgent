package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"

	"docgent-backend/cmd/cli/internal/dto"
	"docgent-backend/internal/domain"
	gh "docgent-backend/internal/infrastructure/github"
)

var CLI struct {
	PRIdentifier   string `arg:"" help:"PR identifier in format 'owner/repo/pulls/number'"`
	InstallationID int64  `flag:"" required:"" help:"GitHub App installation ID"`
	Output         string `flag:"" default:"proposal.json" help:"Output file path"`
}

func main() {
	ctx := kong.Parse(&CLI)

	appIDStr := os.Getenv("GITHUB_APP_ID")
	if appIDStr == "" {
		log.Fatal("GITHUB_APP_ID is not set")
	}
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		log.Fatalf("GITHUB_APP_ID is invalid: %v", err)
	}

	privateKey := os.Getenv("GITHUB_APP_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("GITHUB_APP_PRIVATE_KEY is not set")
	}

	// Parse PR identifier
	parts := strings.Split(CLI.PRIdentifier, "/")
	if len(parts) != 4 {
		ctx.FatalIfErrorf(fmt.Errorf("invalid PR identifier format. Expected 'owner/repo/pulls/number', got '%s'", CLI.PRIdentifier))
	}

	owner := parts[0]
	repo := parts[1]
	number := parts[3]

	api := gh.NewAPI(appID, []byte(privateKey))
	client := api.NewClient(CLI.InstallationID)

	// Create PullRequestAPI
	prAPI := gh.NewPullRequestAPI(client, owner, repo, "main")

	// Get proposal using domain handle
	handle := prAPI.NewProposalHandle(number)
	proposal, err := prAPI.GetProposal(handle)
	if err != nil {
		ctx.FatalIfErrorf(fmt.Errorf("failed to get proposal: %w", err))
	}

	// Convert domain.Proposal to dto.Proposal
	dtoProposal := convertToDTO(proposal)

	// Write to file
	file, err := os.Create(CLI.Output)
	if err != nil {
		ctx.FatalIfErrorf(fmt.Errorf("failed to create output file: %w", err))
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(dtoProposal); err != nil {
		ctx.FatalIfErrorf(fmt.Errorf("failed to encode proposal: %w", err))
	}
}

func convertToDTO(p domain.Proposal) dto.Proposal {
	diffs := make([]dto.Diff, len(p.Diffs))
	for i, d := range p.Diffs {
		diffs[i] = dto.Diff{
			OldName:   d.OldName,
			NewName:   d.NewName,
			Body:      d.Body,
			IsNewFile: d.IsNewFile,
		}
	}

	comments := make([]dto.Comment, len(p.Comments))
	for i, c := range p.Comments {
		comments[i] = dto.Comment{
			Handle: dto.CommentHandle{
				Source: c.Handle.Source,
				Value:  c.Handle.Value,
			},
			Author: c.Author,
			Body:   c.Body,
		}
	}

	return dto.Proposal{
		Handle: dto.ProposalHandle{
			Source: p.Handle.Source,
			Value:  p.Handle.Value,
		},
		Diffs:    diffs,
		Title:    p.Title,
		Body:     p.Body,
		Comments: comments,
	}
}
