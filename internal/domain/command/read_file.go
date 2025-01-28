package command

type ReadFile struct {
	Path string `xml:"path"`
}

func (fc ReadFile) Match(cs Cases) { cs.ReadFile(fc) }
