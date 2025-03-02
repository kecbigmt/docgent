package application

import (
	"context"
	"docgent/internal/application/port"
	"strings"
)

type RagFileSyncUsecase struct {
	ragCorpus        port.RAGCorpus
	fileQueryService port.FileQueryService
}

func NewRagFileSyncUsecase(ragCorpus port.RAGCorpus, fileQueryService port.FileQueryService) *RagFileSyncUsecase {
	return &RagFileSyncUsecase{
		ragCorpus:        ragCorpus,
		fileQueryService: fileQueryService,
	}
}

func (u *RagFileSyncUsecase) Execute(newFiles, modifiedFiles, deletedFiles []string) error {
	ctx := context.Background()

	ragFiles, err := u.ragCorpus.ListFiles(ctx)
	if err != nil {
		return err
	}

	ragFileMap := make(map[string]int64)
	for _, ragFile := range ragFiles {
		path, err := u.fileQueryService.GetFilePath(ragFile.URI)
		if err != nil {
			return err
		}
		ragFileMap[path] = ragFile.ID
	}

	for _, filePath := range newFiles {
		// If the file already exists, skip it
		_, exists := ragFileMap[filePath]
		if exists {
			continue
		}

		// If the file does not exist, upload it
		file, err := u.fileQueryService.FindFile(ctx, filePath)
		if err != nil {
			return err
		}

		// Get the URI (GitHub permalink) for the file
		uri, err := u.fileQueryService.GetURI(ctx, filePath)
		if err != nil {
			return err
		}

		reader := strings.NewReader(file.Content)
		// Use the URI as the displayName instead of the file path
		err = u.ragCorpus.UploadFile(ctx, reader, uri)
		if err != nil {
			return err
		}
	}

	for _, filePath := range modifiedFiles {
		file, err := u.fileQueryService.FindFile(ctx, filePath)
		if err != nil {
			return err
		}

		// Get the URI (GitHub permalink) for the file
		uri, err := u.fileQueryService.GetURI(ctx, filePath)
		if err != nil {
			return err
		}

		reader := strings.NewReader(file.Content)
		// Use the URI as the displayName instead of the file path
		err = u.ragCorpus.UploadFile(ctx, reader, uri)
		if err != nil {
			return err
		}

		// If the old file exists, delete it
		_, exists := ragFileMap[filePath]
		if exists {
			err = u.ragCorpus.DeleteFile(ctx, ragFileMap[filePath])
			if err != nil {
				return err
			}
		}
	}

	for _, filePath := range deletedFiles {
		// If the file exists, delete it
		_, exists := ragFileMap[filePath]
		if exists {
			err := u.ragCorpus.DeleteFile(ctx, ragFileMap[filePath])
			if err != nil {
				return err
			}
		}
	}

	return nil
}
