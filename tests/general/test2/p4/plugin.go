package p4

type Plugin struct {
}

func (p *Plugin) Init() error {
	println("initp4")
	return nil
}

func (p *Plugin) Serve() chan error {
	return make(chan error, 1)
}

func (p *Plugin) Stop() error {
	return nil
}

func (p *Plugin) Name() string {
	return "p4"
}

func (p *Plugin) Work() {
	println("wooooorking!!!")
}
