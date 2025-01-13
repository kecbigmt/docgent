package domain

/**
 * Proposal
 */

type Proposal struct {
	Handle ProposalHandle
	ProposalContent
	Increment Increment
	Comments  []Comment
}

func NewProposal(handle ProposalHandle, content ProposalContent, increment Increment) Proposal {
	return Proposal{
		Handle:          handle,
		ProposalContent: content,
		Increment:       increment,
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
