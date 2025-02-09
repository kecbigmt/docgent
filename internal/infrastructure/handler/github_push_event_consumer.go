package handler

import (
	"sort"

	"github.com/google/go-github/v68/github"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"docgent-backend/internal/application"
	"docgent-backend/internal/application/port"
	infragithub "docgent-backend/internal/infrastructure/github"
)

type GitHubPushEventConsumerParams struct {
	fx.In

	Logger                   *zap.Logger
	GitHubServiceProvider    *infragithub.ServiceProvider
	RAGService               port.RAGService
	ApplicationConfigService ApplicationConfigService
}

type GitHubPushEventConsumer struct {
	logger                   *zap.Logger
	githubServiceProvider    *infragithub.ServiceProvider
	ragService               port.RAGService
	applicationConfigService ApplicationConfigService
}

func NewGitHubPushEventConsumer(params GitHubPushEventConsumerParams) *GitHubPushEventConsumer {
	return &GitHubPushEventConsumer{
		logger:                   params.Logger,
		githubServiceProvider:    params.GitHubServiceProvider,
		ragService:               params.RAGService,
		applicationConfigService: params.ApplicationConfigService,
	}
}

func (c *GitHubPushEventConsumer) EventType() string {
	return "push"
}

func (c *GitHubPushEventConsumer) ConsumeEvent(event interface{}) {
	ev, ok := event.(*github.PushEvent)
	if !ok {
		c.logger.Error("Failed to convert event data to PushEvent")
		return
	}
	c.logger.Info("Processing issue comment event", zap.String("action", ev.GetAction()))

	installationID := ev.GetInstallation().GetID()
	workspace, err := c.applicationConfigService.GetWorkspaceByGitHubInstallationID(installationID)
	if err != nil {
		c.logger.Error("Failed to get workspace", zap.Error(err))
		return
	}

	defaultBranchRef := "refs/heads/" + workspace.GitHubDefaultBranch

	if ev.GetRef() != defaultBranchRef {
		c.logger.Info("Skipping non-default branch push", zap.String("ref", ev.GetRef()))
		return
	}

	newFiles, modifiedFiles, deletedFiles := classifyFilesBySimulation(ev.GetCommits())

	fileQueryService := c.githubServiceProvider.NewFileQueryService(installationID, workspace.GitHubOwner, workspace.GitHubRepo, workspace.GitHubDefaultBranch)
	ragCorpus := c.ragService.GetCorpus(workspace.VertexAICorpusID)
	ragFileSyncUsecase := application.NewRagFileSyncUsecase(ragCorpus, fileQueryService)

	c.logger.Info("Syncing RAG files...", zap.Strings("newFiles", newFiles), zap.Strings("modifiedFiles", modifiedFiles), zap.Strings("deletedFiles", deletedFiles))

	err = ragFileSyncUsecase.Execute(newFiles, modifiedFiles, deletedFiles)
	if err != nil {
		c.logger.Error("Failed to sync RAG files", zap.Error(err))
		return
	}

	c.logger.Info("Synced RAG files", zap.Strings("newFiles", newFiles), zap.Strings("modifiedFiles", modifiedFiles), zap.Strings("deletedFiles", deletedFiles))
}

type fileStatus int

const (
	fileStatusAdded fileStatus = iota
	fileStatusModified
	fileStatusRemoved
)

func classifyFilesBySimulation(commits []*github.HeadCommit) (newFiles, modifiedFiles, deletedFiles []string) {
	// Sort commits by timestamp
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Timestamp.Before(*commits[j].Timestamp.GetTime())
	})

	// Record file status
	fileStates := make(map[string]fileStatus)

	// Process each commit in chronological order
	for _, c := range commits {
		// (1) Process removed files first: files removed in the commit will no longer exist
		for _, file := range c.Removed {
			status, exists := fileStates[file]
			if exists && status == fileStatusAdded {
				delete(fileStates, file)
				continue
			}
			fileStates[file] = fileStatusRemoved
		}
		// (2) Process added files: files added in the commit will be new
		for _, file := range c.Added {
			status, exists := fileStates[file]
			if exists && status == fileStatusRemoved {
				fileStates[file] = fileStatusModified
				continue
			}
			fileStates[file] = fileStatusAdded
		}
		// (3) Process modified files: if there is a change, it is updated
		for _, file := range c.Modified {
			status, exists := fileStates[file]
			if exists && status == fileStatusAdded {
				continue
			}
			fileStates[file] = fileStatusModified
		}
	}

	// Classify files by status
	for file, fs := range fileStates {
		switch fs {
		case fileStatusAdded:
			newFiles = append(newFiles, file)
		case fileStatusModified:
			modifiedFiles = append(modifiedFiles, file)
		case fileStatusRemoved:
			deletedFiles = append(deletedFiles, file)
		}
	}
	return newFiles, modifiedFiles, deletedFiles
}
