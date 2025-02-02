package dto

import "docgent-backend/internal/domain"

type Proposal struct {
	Handle   ProposalHandle `json:"handle"`
	Diffs    []Diff         `json:"diffs"`
	Title    string         `json:"title"`
	Body     string         `json:"body"`
	Comments []Comment      `json:"comments"`
}

type ProposalHandle struct {
	Source string `json:"source"`
	Value  string `json:"value"`
}

type Diff struct {
	OldName   string `json:"oldName"`
	NewName   string `json:"newName"`
	Body      string `json:"body"`
	IsNewFile bool   `json:"isNewFile"`
}

type Comment struct {
	Handle CommentHandle `json:"handle"`
	Author string        `json:"author"`
	Body   string        `json:"body"`
}

type CommentHandle struct {
	Source string `json:"source"`
	Value  string `json:"value"`
}

// 変換メソッド
func (p Proposal) ToDomain() domain.Proposal {
	return domain.NewProposal(
		domain.NewProposalHandle(p.Handle.Source, p.Handle.Value),
		convertDiffs(p.Diffs),
		domain.NewProposalContent(p.Title, p.Body),
		convertComments(p.Comments),
	)
}

func convertDiffs(diffs []Diff) domain.Diffs {
	result := make(domain.Diffs, 0, len(diffs))
	for _, d := range diffs {
		if d.IsNewFile {
			result = append(result, domain.NewCreateDiff(d.NewName, d.Body))
		} else {
			result = append(result, domain.NewUpdateDiff(d.OldName, d.NewName, d.Body))
		}
	}
	return result
}

func convertComments(comments []Comment) []domain.Comment {
	result := make([]domain.Comment, 0, len(comments))
	for _, c := range comments {
		result = append(result, domain.NewComment(
			domain.NewCommentHandle(c.Handle.Source, c.Handle.Value),
			c.Author,
			c.Body,
		))
	}
	return result
}
