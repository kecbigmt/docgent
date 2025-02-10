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
		ragFileMap[ragFile.FileName] = ragFile.ID
	}

	for _, fileName := range newFiles {
		// If the file already exists, skip it
		_, exists := ragFileMap[fileName]
		if exists {
			continue
		}

		// If the file does not exist, upload it
		file, err := u.fileQueryService.FindFile(ctx, fileName)
		if err != nil {
			return err
		}
		reader := strings.NewReader(file.Content)
		err = u.ragCorpus.UploadFile(ctx, reader, fileName)
		if err != nil {
			return err
		}
	}

	for _, fileName := range modifiedFiles {
		file, err := u.fileQueryService.FindFile(ctx, fileName)
		if err != nil {
			return err
		}
		reader := strings.NewReader(file.Content)
		err = u.ragCorpus.UploadFile(ctx, reader, fileName)
		if err != nil {
			return err
		}

		// If the old file exists, delete it
		_, exists := ragFileMap[fileName]
		if exists {
			err = u.ragCorpus.DeleteFile(ctx, ragFileMap[fileName])
			if err != nil {
				return err
			}
		}
	}

	for _, fileName := range deletedFiles {
		// If the file exists, delete it
		_, exists := ragFileMap[fileName]
		if exists {
			err := u.ragCorpus.DeleteFile(ctx, ragFileMap[fileName])
			if err != nil {
				return err
			}
		}
	}

	return nil
}
