package collects_get_all_deps

type Plugin1 struct {
}

type SuperInterface interface {
	Super() string
}

type Super2Interface interface {
	Super2() string
}

func (f *Plugin1) Init() error {
	return nil
}

func (f *Plugin1) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (f *Plugin1) Stop() error {
	return nil
}

// Super and Super2 interface impl
func (f *Plugin1) Super() string {
	return "SUPER -> "
}

func (f *Plugin1) Super2() string {
	return "I'm also SUPER2 -> "
}
