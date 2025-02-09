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
	Description: "Searches for domain-specific related information from existing documents. Returns up to 10 documents sorted by descending relevance, but they may not necessarily be directly related. You should not use this tool if the question is not related to the domain.",
	Parameters: []Parameter{
		{
			Name:        "query",
			Description: "The query to search for. It is recommended to incorporate as many domain-specific terms as possible.",
			Required:    true,
		},
	},
	Example: `<query_rag>
    <query>Search query here</query>
</query_rag>`,
}

func (t QueryRAG) Match(cases Cases) (string, bool, error) {
	if cases.QueryRAG == nil {
		return "", false, fmt.Errorf("query_rag handler is not implemented")
	}
	return cases.QueryRAG(t)
}
