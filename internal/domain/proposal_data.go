package domain

/**
 * Proposal
 */

type Proposal struct {
	Handle ProposalHandle
	Diffs  Diffs
	ProposalContent
	Comments []Comment
}

func NewProposal(
	handle ProposalHandle,
	diffs Diffs,
	content ProposalContent,
	comments []Comment,
) Proposal {
	return Proposal{
		Handle:          handle,
		Diffs:           diffs,
		ProposalContent: content,
		Comments:        comments,
	}
}

/**
 * ProposalHandle
 */

type ProposalHandle struct {
	Source string
	Value  string
}

func NewProposalHandle(source, value string) ProposalHandle {
	return ProposalHandle{
		Source: source,
		Value:  value,
	}
}

/**
 * Diff
 */

type Diffs []Diff

func (ds Diffs) ToXMLString() string {
	str := "<diffs>"
	for _, d := range ds {
		str += d.ToXMLString()
	}
	str += "</diffs>"
	return str
}

type Diff struct {
	OldPath   string
	NewPath   string
	Body      string
	IsNewFile bool
}

func NewUpdateDiff(oldPath, newPath, body string) Diff {
	return Diff{
		OldPath:   oldPath,
		NewPath:   newPath,
		Body:      body,
		IsNewFile: false,
	}
}

func NewCreateDiff(newPath, body string) Diff {
	return Diff{
		OldPath:   "",
		NewPath:   newPath,
		Body:      body,
		IsNewFile: true,
	}
}

func (d Diff) ToXMLString() string {
	str := "<diff>"
	str += "<old_path>" + d.OldPath + "</old_path>"
	str += "<new_path>" + d.NewPath + "</new_path>"
	str += "<body>" + d.Body + "</body>"
	str += "</diff>"
	return str
}

/**
 * ProposalContent
 */

type ProposalContent struct {
	Title string
	Body  string
}

func NewProposalContent(title, body string) ProposalContent {
	return ProposalContent{
		Title: title,
		Body:  body,
	}
}

/**
 * Comment
 */

type Comment struct {
	Handle CommentHandle
	Author string
	Body   string
}

func NewComment(handle CommentHandle, author, body string) Comment {
	return Comment{
		Handle: handle,
		Author: author,
		Body:   body,
	}
}

type CommentHandle struct {
	Source string
	Value  string
}

func NewCommentHandle(source, value string) CommentHandle {
	return CommentHandle{
		Source: source,
		Value:  value,
	}
}
