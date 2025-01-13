package domain

/**
 * Proposal
 */

type Proposal struct {
	Handle ProposalHandle
	ProposalContent
	Increment Increment
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
