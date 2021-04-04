package plugin10

import "fmt"

type Plugin10 struct{}

// No deps
func (s *Plugin10) Init() error {
	return nil
}

func (s *Plugin10) Serve() chan error {
	errCh := make(chan error)
	return errCh
}

func (s *Plugin10) Stop() error {
	return nil
}

func (s *Plugin10) Boo() {
	fmt.Println("Boo from plugin10")
}
