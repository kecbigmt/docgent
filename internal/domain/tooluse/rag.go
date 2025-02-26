package tooluse

import (
	"encoding/xml"
	"fmt"
)

// QueryRAG is a tool to search for domain-specific related information from existing documents.
type QueryRAG struct {
	XMLName xml.Name `xml:"query_rag"`
	Query   string   `xml:"query"`
}

var QueryRAGUsage = Usage{
	Name:        "query_rag",
	Description: "Search for domain-specific information in APPROVED DOCUMENTS (secondary sources)",
	Parameters: []Parameter{
		{
			Name:        "query",
			Description: "The query to search for in the knowledge base of approved documents",
			Required:    true,
		},
	},
	Example: `<query_rag>
<query>What are the API endpoints for the user service?</query>
</query_rag>

IMPORTANT: This tool searches SECONDARY SOURCES (approved documents):
- Use for broad knowledge discovery across all approved documents
- Returns curated, organized information from multiple documents
- Complements find_source which accesses primary sources
- Best for general queries about documented knowledge
- Not as detailed as primary sources for specific conversations
- Use same language as the conversation history or the approved documents`,
}

func (t QueryRAG) Match(cases Cases) (string, bool, error) {
	if cases.QueryRAG == nil {
		return "", false, fmt.Errorf("query_rag handler is not implemented")
	}
	return cases.QueryRAG(t)
}
