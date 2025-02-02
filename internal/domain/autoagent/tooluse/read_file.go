package tooluse

type ReadFile struct {
	Path string `xml:"path"`
}

func (fc ReadFile) Match(cs Cases) error { return cs.ReadFile(fc) }
