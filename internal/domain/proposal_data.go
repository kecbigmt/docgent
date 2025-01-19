package domain

import "strconv"

/**
 * Proposal
 */

type Proposal struct {
	Handle ProposalHandle
	Diffs  []Diff
	ProposalContent
	Comments []Comment
}

func NewProposal(
	handle ProposalHandle,
	diffs []Diff,
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
	OldFilename string
	NewFilename string
	Body        string
	IsNewFile   bool
}

func NewUpdateDiff(oldFilename, newFilename, body string) Diff {
	return Diff{
		OldFilename: oldFilename,
		NewFilename: newFilename,
		Body:        body,
		IsNewFile:   false,
	}
}

func NewCreateDiff(newFilename, body string) Diff {
	return Diff{
		OldFilename: "",
		NewFilename: newFilename,
		Body:        body,
		IsNewFile:   true,
	}
}

func (d Diff) ToXMLString() string {
	str := "<diff>"
	str += "<oldFilename>" + d.OldFilename + "</oldFilename>"
	str += "<newFilename>" + d.NewFilename + "</newFilename>"
	str += "<body>" + d.Body + "</body>"
	str += "<isNewFile>" + strconv.FormatBool(d.IsNewFile) + "</isNewFile>"
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
