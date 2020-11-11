package plugin7

import "fmt"

type Plugin7 struct {
}

// No deps
func (s *Plugin7) Init() error {
	return nil
}

func (s *Plugin7) Name() string {
	return "Plugin7"
}

func (s *Plugin7) Boom() {
	fmt.Println("Boom")
}
