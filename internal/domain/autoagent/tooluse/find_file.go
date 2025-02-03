package tooluse

var FindFileUsage = NewUsage("find_file", "Read a file", []Parameter{
	NewParameter("path", "The exact path to the file to read.", true),
}, "<find_file><path>/path/to/file.md</path></find_file>")

type FindFile struct {
	Path string `xml:"path"`
}

func (fc FindFile) Match(cs Cases) (string, bool, error) { return cs.FindFile(fc) }
