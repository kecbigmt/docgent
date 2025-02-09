package cli

type CLI struct {
	Corpus struct {
		Create struct {
			DisplayName string `required:"" help:"Display name of the RAG corpus"`
			Description string `help:"Description of the RAG corpus"`
		} `cmd:"" help:"Create a new RAG corpus"`
		List struct {
			PageSize  int    `help:"Maximum number of corpora to return" default:"50"`
			PageToken string `help:"Page token for the next page of results"`
		} `cmd:"" help:"List RAG corpora"`
	} `cmd:"" help:"Manage RAG corpus"`

	File struct {
		Upload struct {
			File        string `arg:"" help:"Path to the file to upload" type:"path"`
			CorpusID    string `required:"" help:"ID of the RAG corpus"`
			Description string `help:"Description of the file"`
			ChunkSize   int    `help:"Size of each chunk" default:"1000"`
			Overlap     int    `help:"Overlap size between chunks" default:"100"`
		} `cmd:"" help:"Upload a file to the RAG corpus"`

		Delete struct {
			CorpusID string `required:"" help:"ID of the RAG corpus"`
			FileID   string `required:"" help:"ID of the RAG file to delete"`
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
