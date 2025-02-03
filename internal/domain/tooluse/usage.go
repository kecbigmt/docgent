package tooluse

type Usage struct {
	Name        string
	Description string
	Parameters  []Parameter
	Example     string
}

type Parameter struct {
	Name        string
	Description string
	Required    bool
}

func NewUsage(name string, description string, parameters []Parameter, example string) Usage {
	return Usage{
		Name:        name,
		Description: description,
		Parameters:  parameters,
		Example:     example,
	}
}

func NewParameter(name string, description string, required bool) Parameter {
	return Parameter{
		Name:        name,
		Description: description,
		Required:    required,
	}
}
