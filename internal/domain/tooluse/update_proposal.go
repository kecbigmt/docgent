package tooluse

import "encoding/xml"

type UpdateProposal struct {
	XMLName     xml.Name `xml:"update_proposal"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
}

func (up UpdateProposal) Match(cs Cases) (string, bool, error) { return cs.UpdateProposal(up) }

func NewUpdateProposal(title, description string) UpdateProposal {
	return UpdateProposal{
		XMLName:     xml.Name{Local: "update_proposal"},
		Title:       title,
		Description: description,
	}
}

var UpdateProposalUsage = NewUsage("update_proposal", "Update a proposal", []Parameter{
	NewParameter("title", "The title of the proposal", true),
	NewParameter("description", "The description of the proposal", true),
}, "<update_proposal><title>Proposal Title</title><description>Proposal Description</description></update_proposal>")
