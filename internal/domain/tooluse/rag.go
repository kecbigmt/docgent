package tooluse

import (
	"encoding/xml"
	"fmt"
)

// Query RAG is a tool to search for related information from existing documents.
type QueryRAG struct {
	XMLName xml.Name `xml:"query_rag"`
	Query   string   `xml:"query"`
}

var QueryRAGUsage = Usage{
	Name:        "query_rag",
	Description: "Searches for related information from existing documents. Returns up to 10 results sorted by descending relevance.",
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
