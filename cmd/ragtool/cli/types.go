package cli

type CLI struct {
	Corpus struct {
		Create struct {
			DisplayName                 string `required:"" help:"Display name of the RAG corpus"`
			Description                 string `help:"Description of the RAG corpus"`
			VectorDB                    string `help:"Vector database type. Available types: rag_managed_db, pinecone, vertex_vector_search" default:"rag_managed_db"`
			PineconeIndexName           string `help:"Pinecone index name"`
			PineconeAPIKeySecretVersion string `help:"Secret version for Pinecone API key"`
			VectorSearchIndex           string `help:"Vertex Vector Search index"`
			VectorSearchIndexEndpoint   string `help:"Vertex Vector Search index endpoint"`
			EmbeddingPredictionEndpoint string `help:"Vertex AI prediction endpoint for RAG embedding model"`
		} `cmd:"" help:"Create a new RAG corpus"`
		List struct {
			PageSize  int    `help:"Maximum number of corpora to return" default:"50"`
			PageToken string `help:"Page token for the next page of results"`
		} `cmd:"" help:"List RAG corpora"`
		Delete struct {
			CorpusID string `arg:"" required:"" help:"ID of the RAG corpus to delete"`
		} `cmd:"" help:"Delete a RAG corpus"`

		Retrieve struct {
			CorpusID                string  `arg:"" required:"" help:"ID of the RAG corpus to search"`
			Query                   string  `required:"" help:"Search query text"`
			TopK                    int32   `help:"Number of top results to return" default:"5"`
			VectorDistanceThreshold float64 `help:"Threshold for vector distance matching" default:"0.7"`
		} `cmd:"" help:"Search in a RAG corpus"`
	} `cmd:"" help:"Manage RAG corpus"`

	File struct {
		Upload struct {
			File        string `arg:"" help:"Path to the file to upload" type:"path"`
			CorpusID    string `required:"" help:"ID of the RAG corpus"`
			Description string `help:"Description of the file"`
			ChunkSize   int    `help:"Size of each chunk" default:"1000"`
			Overlap     int    `help:"Overlap between chunks" default:"100"`
		} `cmd:"" help:"Upload a file to the RAG corpus"`

		Delete struct {
			CorpusID string `required:"" help:"ID of the RAG corpus"`
			FileID   string `arg:"" required:"" help:"ID of the RAG file to delete"`
		} `cmd:"" help:"Delete a file from the RAG corpus"`

		List struct {
			CorpusID  string `required:"" help:"ID of the RAG corpus"`
			PageSize  int    `help:"Maximum number of files to return" default:"50"`
			PageToken string `help:"Page token for the next page of results"`
		} `cmd:"" help:"List files in the RAG corpus"`
	} `cmd:"" help:"Manage files in RAG corpus"`

	ProjectID string `required:"" help:"Google Cloud Project ID"`
	Location  string `help:"Google Cloud location" default:"us-central1"`
}
