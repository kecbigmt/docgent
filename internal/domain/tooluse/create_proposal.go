package tooluse

import "encoding/xml"

type CreateProposal struct {
	XMLName     xml.Name `xml:"create_proposal"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
}

func (cp CreateProposal) Match(cs Cases) (string, bool, error) { return cs.CreateProposal(cp) }

func NewCreateProposal(title, description string) CreateProposal {
	return CreateProposal{
		XMLName:     xml.Name{Local: "create_proposal"},
		Title:       title,
		Description: description,
	}
}

var CreateProposalUsage = NewUsage("create_proposal", "Create a proposal", []Parameter{
	NewParameter("title", "The title of the proposal", true),
	NewParameter("description", "The description of the proposal", true),
}, "<create_proposal><title>Proposal Title</title><description>Proposal Description</description></create_proposal>")
