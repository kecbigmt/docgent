package tooluse

var FindFileUsage = NewUsage("read_file", "Read a file", []Parameter{
	NewParameter("path", "The exact path to the file to read", true),
}, "<read_file><path>/path/to/file.md</path></read_file>")

type FindFile struct {
	Path string `xml:"path"`
}

func (fc FindFile) Match(cs Cases) (string, bool, error) { return cs.FindFile(fc) }
