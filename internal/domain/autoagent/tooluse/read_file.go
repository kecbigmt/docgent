package tooluse

var ReadFileUsage = NewUsage("read_file", "Read a file", []Parameter{
	NewParameter("path", "The exact path to the file to read", true),
}, "<read_file><path>/path/to/file.md</path></read_file>")

type ReadFile struct {
	Path string `xml:"path"`
}

func (fc ReadFile) Match(cs Cases) (string, bool, error) { return cs.ReadFile(fc) }
