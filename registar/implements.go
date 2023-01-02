package registar

type implements struct {
	// plugin interface
	plugin any
	// method will be non-empty if we have Provided dep
	methods []string
}

func (i *implements) Plugin() any {
	return i.plugin
}

func (i *implements) Method() []string {
	return i.methods
}
